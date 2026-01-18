package tasks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestHTTPTask_IntegrationWithEngine(t *testing.T) {
	// Create test API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": 123,
			"name":    "John Doe",
			"email":   "john@example.com",
		})
	}))
	defer server.Close()

	// Create registry and register HTTP task
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)

	// Create engine with registry
	eng := engine.NewEngine(registry)

	// Create workflow with HTTP request task
	workflow := engine.WorkflowDefinition{
		Name: "test-http-workflow",
		Tasks: []engine.Task{
			{
				ID:   "fetch_user",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "GET",
					"url":    server.URL,
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify response was stored in context
	result, exists := eng.GetContext().Get("fetch_user_result")
	assert.True(t, exists)
	assert.NotNil(t, result)

	// Verify response structure
	output := result.(map[string]interface{})
	assert.Equal(t, 200, output["status_code"])
	assert.NotNil(t, output["body"])

	body := output["body"].(map[string]interface{})
	assert.Equal(t, float64(123), body["user_id"])
	assert.Equal(t, "John Doe", body["name"])
}

func TestHTTPTask_IntegrationWithContextSharing(t *testing.T) {
	// Create test API server that expects data from context
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Echo back the request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"received_user": reqBody["user_id"],
			"received_name": reqBody["name"],
		})
	}))
	defer server.Close()

	// Create registry and register HTTP task
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)

	// Create mock executor to set initial context
	mockTask := &engine.MockExecutor{
		Output: map[string]interface{}{
			"user_id": "user-456",
			"name":    "Jane Smith",
		},
	}
	registry.Register("mock", mockTask)

	// Create engine with registry
	eng := engine.NewEngine(registry)

	// Create workflow with multiple tasks
	workflow := engine.WorkflowDefinition{
		Name: "test-context-sharing-workflow",
		Tasks: []engine.Task{
			{
				ID:     "prepare_data",
				Type:   "mock",
				Config: map[string]interface{}{},
			},
			{
				ID:   "send_request",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "POST",
					"url":    server.URL,
					"body":   `{"user_id":"{{.context.prepare_data_result.user_id}}","name":"{{.context.prepare_data_result.name}}"}`,
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify first task result
	_, exists := eng.GetContext().Get("prepare_data_result")
	assert.True(t, exists)

	// Verify HTTP task result
	httpResult, exists := eng.GetContext().Get("send_request_result")
	assert.True(t, exists)
	assert.NotNil(t, httpResult)

	// Verify the HTTP task received and processed the context data
	output := httpResult.(map[string]interface{})
	body := output["body"].(map[string]interface{})
	assert.Equal(t, "user-456", body["received_user"])
	assert.Equal(t, "Jane Smith", body["received_name"])
}

func TestHTTPTask_IntegrationWithSubsequentTasks(t *testing.T) {
	// Create test API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"api_data": "important value",
		})
	}))
	defer server.Close()

	// Create registry and register tasks
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)

	mockProcessor := &engine.MockExecutor{
		Output: "processed",
	}
	registry.Register("processor", mockProcessor)

	// Create engine
	eng := engine.NewEngine(registry)

	// Workflow: HTTP fetch -> process result
	workflow := engine.WorkflowDefinition{
		Name: "test-sequential-workflow",
		Tasks: []engine.Task{
			{
				ID:   "fetch_api",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "GET",
					"url":    server.URL,
				},
			},
			{
				ID:     "process_data",
				Type:   "processor",
				Config: map[string]interface{}{},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Both tasks should have results in context
	apiResult, exists1 := eng.GetContext().Get("fetch_api_result")
	processResult, exists2 := eng.GetContext().Get("process_data_result")

	assert.True(t, exists1)
	assert.True(t, exists2)
	assert.NotNil(t, apiResult)
	assert.NotNil(t, processResult)
}

func TestHTTPTask_IntegrationWithFailure(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server error"))
	}))
	defer server.Close()

	// Create registry and register tasks
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)

	mockTask := &engine.MockExecutor{Output: "should not execute"}
	registry.Register("mock", mockTask)

	// Create engine
	eng := engine.NewEngine(registry)

	// Workflow: failing HTTP -> subsequent task
	workflow := engine.WorkflowDefinition{
		Name: "test-failure-workflow",
		Tasks: []engine.Task{
			{
				ID:   "failing_request",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "GET",
					"url":    server.URL,
				},
			},
			{
				ID:     "should_not_run",
				Type:   "mock",
				Config: map[string]interface{}{},
			},
		},
	}

	// Execute workflow - should fail
	err := eng.Execute(workflow)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 500")

	// Verify second task did not execute
	_, exists := eng.GetContext().Get("should_not_run_result")
	assert.False(t, exists)
}
