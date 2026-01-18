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

func TestHTMLParserTask_IntegrationWithEngine(t *testing.T) {
	// Create registry and register HTML Parser task
	registry := engine.NewRegistry()
	RegisterHTMLParserTask(registry)

	// Create engine
	eng := engine.NewEngine(registry)

	// Pre-populate context with HTML content
	htmlContent := `
	<html>
		<body>
			<h1 class="title">Test Page</h1>
			<div class="products">
				<div class="product">Product 1</div>
				<div class="product">Product 2</div>
				<div class="product">Product 3</div>
			</div>
		</body>
	</html>
	`
	eng.GetContext().Set("html_data", htmlContent)

	// Create workflow with HTML Parser task
	workflow := engine.WorkflowDefinition{
		Name: "test-html-parser-workflow",
		Tasks: []engine.Task{
			{
				ID:   "parse_html",
				Type: "html_parser",
				Config: map[string]interface{}{
					"html_source": "html_data",
					"selectors": []interface{}{
						map[string]interface{}{
							"name":     "title",
							"selector": ".title",
						},
						map[string]interface{}{
							"name":     "products",
							"selector": ".product",
							"multiple": true,
						},
					},
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify parsed data stored in context
	result, exists := eng.GetContext().Get("parse_html_result")
	assert.True(t, exists)
	assert.NotNil(t, result)

	// Verify extraction
	results := result.([]map[string]any)
	assert.Len(t, results, 1)
	assert.Equal(t, "Test Page", results[0]["title"])

	products := results[0]["products"].([]string)
	assert.Len(t, products, 3)
	assert.Equal(t, "Product 1", products[0])
	assert.Equal(t, "Product 2", products[1])
	assert.Equal(t, "Product 3", products[2])
}

func TestHTTPAndHTMLParser_IntegrationWorkflow(t *testing.T) {
	// Create test API server that returns HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
				<head><title>Product Listing</title></head>
				<body>
					<div class="listing">
						<div class="item">
							<h2 class="name">Item A</h2>
							<span class="price">$10.00</span>
							<a href="/item-a" class="link">View</a>
						</div>
						<div class="item">
							<h2 class="name">Item B</h2>
							<span class="price">$20.00</span>
							<a href="/item-b" class="link">View</a>
						</div>
						<div class="item">
							<h2 class="name">Item C</h2>
							<span class="price">$30.00</span>
							<a href="/item-c" class="link">View</a>
						</div>
					</div>
				</body>
			</html>
		`))
	}))
	defer server.Close()

	// Create registry and register both tasks
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)
	RegisterHTMLParserTask(registry)

	// Create engine
	eng := engine.NewEngine(registry)

	// Workflow: HTTP fetch HTML -> Parse with CSS selectors
	workflow := engine.WorkflowDefinition{
		Name: "fetch-and-parse-html",
		Tasks: []engine.Task{
			{
				ID:   "fetch_page",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "GET",
					"url":    server.URL,
				},
			},
			{
				ID:   "parse_items",
				Type: "html_parser",
				Config: map[string]interface{}{
					"html_source": "fetch_page_result",
					"selectors": []interface{}{
						map[string]interface{}{
							"name":     "names",
							"selector": ".item .name",
							"multiple": true,
						},
						map[string]interface{}{
							"name":     "prices",
							"selector": ".item .price",
							"multiple": true,
						},
						map[string]interface{}{
							"name":      "links",
							"selector":  ".item .link",
							"attribute": "href",
							"multiple":  true,
						},
					},
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify HTTP task result
	httpResult, exists := eng.GetContext().Get("fetch_page_result")
	assert.True(t, exists)
	assert.NotNil(t, httpResult)

	// Verify HTML Parser task result
	parseResult, exists := eng.GetContext().Get("parse_items_result")
	assert.True(t, exists)
	assert.NotNil(t, parseResult)

	// Verify extracted data
	results := parseResult.([]map[string]any)
	assert.Len(t, results, 1)

	names := results[0]["names"].([]string)
	assert.Len(t, names, 3)
	assert.Equal(t, "Item A", names[0])
	assert.Equal(t, "Item B", names[1])
	assert.Equal(t, "Item C", names[2])

	prices := results[0]["prices"].([]string)
	assert.Len(t, prices, 3)
	assert.Equal(t, "$10.00", prices[0])
	assert.Equal(t, "$20.00", prices[1])
	assert.Equal(t, "$30.00", prices[2])

	links := results[0]["links"].([]string)
	assert.Len(t, links, 3)
	assert.Equal(t, "/item-a", links[0])
	assert.Equal(t, "/item-b", links[1])
	assert.Equal(t, "/item-c", links[2])
}

func TestHTTPHTMLParserTransform_CompleteDataPipeline(t *testing.T) {
	// Create test API server with product listings HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
				<body>
					<div class="products">
						<article class="product" data-id="1">
							<h3 class="title">Laptop</h3>
							<span class="price">$999</span>
						</article>
						<article class="product" data-id="2">
							<h3 class="title">Mouse</h3>
							<span class="price">$29</span>
						</article>
					</div>
				</body>
			</html>
		`))
	}))
	defer server.Close()

	// Create registry and register all three tasks
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)
	RegisterHTMLParserTask(registry)
	RegisterTransformTask(registry)

	// Create engine
	eng := engine.NewEngine(registry)

	// Complete data pipeline: HTTP -> HTML Parser -> Transform
	workflow := engine.WorkflowDefinition{
		Name: "complete-scraping-pipeline",
		Tasks: []engine.Task{
			{
				ID:   "fetch_html",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "GET",
					"url":    server.URL,
				},
			},
			{
				ID:   "extract_products",
				Type: "html_parser",
				Config: map[string]interface{}{
					"html_source": "fetch_html_result",
					"selectors": []interface{}{
						map[string]interface{}{
							"name":     "titles",
							"selector": ".product .title",
							"multiple": true,
						},
						map[string]interface{}{
							"name":     "prices",
							"selector": ".product .price",
							"multiple": true,
						},
					},
				},
			},
			{
				ID:   "format_output",
				Type: "transform",
				Config: map[string]interface{}{
					"template":      `Product Count: {{len (index . 0).titles}}`,
					"data_source":   "extract_products_result",
					"output_format": "string",
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify all task results exist
	httpResult, exists := eng.GetContext().Get("fetch_html_result")
	assert.True(t, exists)
	assert.NotNil(t, httpResult)

	parseResult, exists := eng.GetContext().Get("extract_products_result")
	assert.True(t, exists)
	assert.NotNil(t, parseResult)

	transformResult, exists := eng.GetContext().Get("format_output_result")
	assert.True(t, exists)
	assert.NotNil(t, transformResult)

	// Verify parsed products
	parsedData := parseResult.([]map[string]any)
	assert.Len(t, parsedData, 1)

	titles := parsedData[0]["titles"].([]string)
	assert.Len(t, titles, 2)
	assert.Equal(t, "Laptop", titles[0])
	assert.Equal(t, "Mouse", titles[1])

	prices := parsedData[0]["prices"].([]string)
	assert.Len(t, prices, 2)
	assert.Equal(t, "$999", prices[0])
	assert.Equal(t, "$29", prices[1])

	// Note: Transform operates on the []map[string]any structure from HTML parser
	// The template accesses the first element implicitly
	assert.Contains(t, transformResult.(string), "Product Count:")
}

func TestHTMLParserTask_IntegrationWithRealWorldHTML(t *testing.T) {
	// Simulate more realistic HTML structure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>Airbnb Clone - Listings</title>
			</head>
			<body>
				<main class="listings-container">
					<div class="listing-card" data-listing-id="123">
						<img src="/images/beach-house.jpg" class="listing-image" alt="Beach House">
						<div class="listing-details">
							<h3 class="listing-title">Beautiful Beach House</h3>
							<p class="listing-location">Malibu, CA</p>
							<div class="listing-price">
								<span class="price-amount">$350</span>
								<span class="price-period">/ night</span>
							</div>
							<div class="listing-rating">
								<span class="rating-score">4.9</span>
								<span class="rating-count">(127 reviews)</span>
							</div>
							<a href="/listings/123" class="view-link">View Details</a>
						</div>
					</div>
					<div class="listing-card" data-listing-id="456">
						<img src="/images/cabin.jpg" class="listing-image" alt="Mountain Cabin">
						<div class="listing-details">
							<h3 class="listing-title">Cozy Mountain Cabin</h3>
							<p class="listing-location">Aspen, CO</p>
							<div class="listing-price">
								<span class="price-amount">$275</span>
								<span class="price-period">/ night</span>
							</div>
							<div class="listing-rating">
								<span class="rating-score">4.8</span>
								<span class="rating-count">(89 reviews)</span>
							</div>
							<a href="/listings/456" class="view-link">View Details</a>
						</div>
					</div>
				</main>
			</body>
			</html>
		`))
	}))
	defer server.Close()

	// Create registry and register tasks
	registry := engine.NewRegistry()
	RegisterHTTPTask(registry)
	RegisterHTMLParserTask(registry)

	// Create engine
	eng := engine.NewEngine(registry)

	// Workflow simulating Airbnb scraping scenario
	workflow := engine.WorkflowDefinition{
		Name: "airbnb-scraping-simulation",
		Tasks: []engine.Task{
			{
				ID:   "fetch_listings_page",
				Type: "http_request",
				Config: map[string]interface{}{
					"method": "GET",
					"url":    server.URL,
				},
			},
			{
				ID:   "extract_listing_data",
				Type: "html_parser",
				Config: map[string]interface{}{
					"html_source": "fetch_listings_page_result",
					"selectors": []interface{}{
						map[string]interface{}{
							"name":     "titles",
							"selector": ".listing-title",
							"multiple": true,
						},
						map[string]interface{}{
							"name":     "locations",
							"selector": ".listing-location",
							"multiple": true,
						},
						map[string]interface{}{
							"name":     "prices",
							"selector": ".price-amount",
							"multiple": true,
						},
						map[string]interface{}{
							"name":     "ratings",
							"selector": ".rating-score",
							"multiple": true,
						},
						map[string]interface{}{
							"name":      "images",
							"selector":  ".listing-image",
							"attribute": "src",
							"multiple":  true,
						},
						map[string]interface{}{
							"name":      "listing_urls",
							"selector":  ".view-link",
							"attribute": "href",
							"multiple":  true,
						},
					},
				},
			},
		},
	}

	// Execute workflow
	err := eng.Execute(workflow)
	assert.NoError(t, err)

	// Verify extracted listing data
	listingData, exists := eng.GetContext().Get("extract_listing_data_result")
	assert.True(t, exists)

	results := listingData.([]map[string]any)
	assert.Len(t, results, 1)

	data := results[0]

	// Verify titles
	titles := data["titles"].([]string)
	assert.Len(t, titles, 2)
	assert.Equal(t, "Beautiful Beach House", titles[0])
	assert.Equal(t, "Cozy Mountain Cabin", titles[1])

	// Verify locations
	locations := data["locations"].([]string)
	assert.Len(t, locations, 2)
	assert.Equal(t, "Malibu, CA", locations[0])
	assert.Equal(t, "Aspen, CO", locations[1])

	// Verify prices
	prices := data["prices"].([]string)
	assert.Len(t, prices, 2)
	assert.Equal(t, "$350", prices[0])
	assert.Equal(t, "$275", prices[1])

	// Verify ratings
	ratings := data["ratings"].([]string)
	assert.Len(t, ratings, 2)
	assert.Equal(t, "4.9", ratings[0])
	assert.Equal(t, "4.8", ratings[1])

	// Verify images
	images := data["images"].([]string)
	assert.Len(t, images, 2)
	assert.Equal(t, "/images/beach-house.jpg", images[0])
	assert.Equal(t, "/images/cabin.jpg", images[1])

	// Verify listing URLs
	urls := data["listing_urls"].([]string)
	assert.Len(t, urls, 2)
	assert.Equal(t, "/listings/123", urls[0])
	assert.Equal(t, "/listings/456", urls[1])
}
