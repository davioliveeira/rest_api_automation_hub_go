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

func TestTransformTask_IntegrationWithEngine(t *testing.T) {
	// Create registry and register Transform task
	registry := engine.NewRegistry()
	RegisterTransformTask(registry)

	// Create engine with registry
	eng := engine.NewEngine(registry)

	// Pre-populate context with data
	eng.GetContext().Set("user_data", map[string]interface{}{
		"firstName": "Alice",
		"lastName":  "Smith",
		"age":       30,
	})

	// Create workflow with Transform task
	workflow := engine.WorkflowDefinition{
		Name: "test-transform-workflow",
		Tasks: []engine.Task{
			{
				ID:   "reshape_user",
				Type: "transform",
				Config: map[string]interface{}{
					"template":      `{"fullName": "{{.firstName}} {{.lastName}}", "age": {{.age}}}`,
					"data_source":   "user_data",
					"output_format": "json",
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify transformed data stored in context
	result, exists := eng.GetContext().Get("reshape_user_result")
	assert.True(t, exists)
	assert.NotNil(t, result)

	// Verify transformation
	output := result.(map[string]interface{})
	assert.Equal(t, "Alice Smith", output["fullName"])
	assert.Equal(t, float64(30), output["age"])
}

func TestHTTPAndTransform_IntegrationWorkflow(t *testing.T) {
	// Create test API server that returns user data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice", "email": "alice@example.com"},
				map[string]interface{}{"id": 2, "name": "Bob", "email": "bob@example.com"},
				map[string]interface{}{"id": 3, "name": "Charlie", "email": "charlie@example.com"},
			},
		})
	}))
	defer server.Close()

	// Create registry and register both tasks
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)
	RegisterTransformTask(registry)

	// Create engine
	eng := engine.NewEngine(registry)

	// Workflow: HTTP fetch -> Transform to extract names
	workflow := engine.WorkflowDefinition{
		Name: "fetch-and-transform",
		Tasks: []engine.Task{
			{
				ID:   "fetch_users",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "GET",
					"url":    server.URL,
				},
			},
			{
				ID:   "extract_names",
				Type: "transform",
				Config: map[string]interface{}{
					"template":      `[{{range $i, $u := .body.users}}{{if $i}},{{end}}"{{$u.name}}"{{end}}]`,
					"data_source":   "fetch_users_result",
					"output_format": "json",
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify HTTP task result
	httpResult, exists := eng.GetContext().Get("fetch_users_result")
	assert.True(t, exists)
	assert.NotNil(t, httpResult)

	// Verify Transform task result
	transformResult, exists := eng.GetContext().Get("extract_names_result")
	assert.True(t, exists)
	assert.NotNil(t, transformResult)

	// Verify extracted names array
	names, ok := transformResult.([]interface{})
	assert.True(t, ok)
	assert.Len(t, names, 3)
	assert.Equal(t, "Alice", names[0])
	assert.Equal(t, "Bob", names[1])
	assert.Equal(t, "Charlie", names[2])
}

func TestHTTPAndTransform_ComplexDataReshaping(t *testing.T) {
	// Create test API server with nested response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"firstName": "John",
						"lastName":  "Doe",
						"age":       25,
					},
					"settings": map[string]interface{}{
						"theme": "dark",
					},
				},
			},
		})
	}))
	defer server.Close()

	// Create registry and register both tasks
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)
	RegisterTransformTask(registry)

	// Create engine
	eng := engine.NewEngine(registry)

	// Workflow: HTTP fetch -> Transform to flatten structure
	workflow := engine.WorkflowDefinition{
		Name: "fetch-and-flatten",
		Tasks: []engine.Task{
			{
				ID:   "fetch_user_profile",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "GET",
					"url":    server.URL,
				},
			},
			{
				ID:   "flatten_profile",
				Type: "transform",
				Config: map[string]interface{}{
					"template":      `{"name":"{{.body.data.user.profile.firstName}} {{.body.data.user.profile.lastName}}","age":{{.body.data.user.profile.age}},"theme":"{{.body.data.user.settings.theme}}"}`,
					"data_source":   "fetch_user_profile_result",
					"output_format": "json",
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify Transform result
	transformResult, exists := eng.GetContext().Get("flatten_profile_result")
	assert.True(t, exists)

	flattened := transformResult.(map[string]interface{})
	assert.Equal(t, "John Doe", flattened["name"])
	assert.Equal(t, float64(25), flattened["age"])
	assert.Equal(t, "dark", flattened["theme"])
}

func TestTransformTask_IntegrationWithMultipleContextSources(t *testing.T) {
	// Create registry and register Transform task
	registry := engine.NewRegistry()
	RegisterTransformTask(registry)

	// Create engine
	eng := engine.NewEngine(registry)

	// Populate context with multiple data sources
	eng.GetContext().Set("user_info", map[string]interface{}{
		"name": "Alice",
		"id":   123,
	})
	eng.GetContext().Set("order_info", map[string]interface{}{
		"total":    99.99,
		"currency": "USD",
	})

	// Create workflow that combines data from multiple sources
	workflow := engine.WorkflowDefinition{
		Name: "combine-context-data",
		Tasks: []engine.Task{
			{
				ID:   "combine_data",
				Type: "transform",
				Config: map[string]interface{}{
					// No data_source - access entire context
					"template":      `{"customer":"{{.user_info.name}}","customerId":{{.user_info.id}},"orderTotal":{{.order_info.total}},"currency":"{{.order_info.currency}}"}`,
					"output_format": "json",
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify combined result
	result, exists := eng.GetContext().Get("combine_data_result")
	assert.True(t, exists)

	combined := result.(map[string]interface{})
	assert.Equal(t, "Alice", combined["customer"])
	assert.Equal(t, float64(123), combined["customerId"])
	assert.Equal(t, float64(99.99), combined["orderTotal"])
	assert.Equal(t, "USD", combined["currency"])
}

func TestTransformTask_IntegrationWithCustomFunctions(t *testing.T) {
	// Create test API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []interface{}{"apple", "banana", "cherry"},
			"name":  "product list",
		})
	}))
	defer server.Close()

	// Create registry and register both tasks
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)
	RegisterTransformTask(registry)

	// Create engine
	eng := engine.NewEngine(registry)

	// Workflow: HTTP fetch -> Transform using custom functions
	workflow := engine.WorkflowDefinition{
		Name: "transform-with-functions",
		Tasks: []engine.Task{
			{
				ID:   "fetch_data",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "GET",
					"url":    server.URL,
				},
			},
			{
				ID:   "transform_data",
				Type: "transform",
				Config: map[string]interface{}{
					"template":      `{"title":"{{.body.name | toUpper}}","items":"{{.body.items | join ", "}}"}`,
					"data_source":   "fetch_data_result",
					"output_format": "json",
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify transform with custom functions
	result, exists := eng.GetContext().Get("transform_data_result")
	assert.True(t, exists)

	transformed := result.(map[string]interface{})
	assert.Equal(t, "PRODUCT LIST", transformed["title"])
	assert.Equal(t, "apple, banana, cherry", transformed["items"])
}
