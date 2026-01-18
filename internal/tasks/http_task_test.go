package tasks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestHTTPTask_Execute_GET_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
	assert.NotNil(t, result.Output)

	output := result.Output.(map[string]interface{})
	assert.Equal(t, 200, output["status_code"])
	assert.NotNil(t, output["body"])

	body := output["body"].(map[string]interface{})
	assert.Equal(t, "success", body["message"])
}

func TestHTTPTask_Execute_POST_WithBody(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "test", reqBody["key"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"received": "ok"})
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "POST",
		"url":    server.URL,
		"body":   `{"key":"test"}`,
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
}

func TestHTTPTask_Execute_WithCustomHeaders(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "CustomValue", r.Header.Get("X-Custom-Header"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
		"headers": map[string]interface{}{
			"Authorization":   "Bearer token123",
			"X-Custom-Header": "CustomValue",
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
}

func TestHTTPTask_Execute_BodyInterpolation(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		assert.Equal(t, "user123", reqBody["user_id"])
		assert.Equal(t, "test@example.com", reqBody["email"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	// Set context values
	ctx.Set("userId", "user123")
	ctx.Set("userEmail", "test@example.com")

	config := map[string]interface{}{
		"method": "POST",
		"url":    server.URL,
		"body":   `{"user_id":"{{.context.userId}}","email":"{{.context.userEmail}}"}`,
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
}

func TestHTTPTask_Execute_Timeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method":  "GET",
		"url":     server.URL,
		"timeout": 1, // 1 second timeout
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "request execution failed")
}

func TestHTTPTask_Execute_HTTPError_404(t *testing.T) {
	// Create test server returning 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "HTTP 404")
}

func TestHTTPTask_Execute_HTTPError_500(t *testing.T) {
	// Create test server returning 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "HTTP 500")
}

func TestHTTPTask_Execute_MissingMethod(t *testing.T) {
	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"url": "http://example.com",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "missing or invalid 'method'")
}

func TestHTTPTask_Execute_MissingURL(t *testing.T) {
	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "GET",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "missing or invalid 'url'")
}

func TestHTTPTask_Execute_InvalidURL(t *testing.T) {
	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "GET",
		"url":    "://invalid-url",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "failed to create request")
}

func TestHTTPTask_Execute_NonJSONResponse(t *testing.T) {
	// Create test server returning plain text
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Plain text response"))
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	output := result.Output.(map[string]interface{})
	assert.Equal(t, "Plain text response", output["body"])
}

func TestHTTPTask_Execute_PUT_Method(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "PUT",
		"url":    server.URL,
		"body":   `{"update":"data"}`,
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
}

func TestHTTPTask_Execute_DELETE_Method(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"method": "DELETE",
		"url":    server.URL,
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
}

func TestHTTPTask_InterpolateBody_Success(t *testing.T) {
	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()
	ctx.Set("name", "John")
	ctx.Set("age", 30)

	template := `{"name":"{{.context.name}}","age":{{.context.age}}}`

	result, err := task.interpolateBody(template, ctx)

	assert.NoError(t, err)
	assert.Contains(t, result, `"name":"John"`)
	assert.Contains(t, result, `"age":30`)
}

func TestHTTPTask_InterpolateBody_InvalidTemplate(t *testing.T) {
	task := &HTTPTask{}
	ctx := engine.NewExecutionContext()

	template := `{{.context.invalid`

	_, err := task.interpolateBody(template, ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

func TestHTTPTask_RegisterHTTPTask(t *testing.T) {
	registry := engine.NewRegistry()

	RegisterHTTPTask(registry)

	executor, err := registry.Get("http_request")
	assert.NoError(t, err)
	assert.NotNil(t, executor)
	assert.IsType(t, &HTTPTask{}, executor)
}
