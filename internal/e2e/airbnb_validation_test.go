package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
	"github.com/davioliveira/rest_api_automation_hub_go/internal/repository"
	"github.com/davioliveira/rest_api_automation_hub_go/internal/tasks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// =============================================================================
// Mock Airbnb HTML Server
// =============================================================================

func createMockAirbnbServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
					<div class="listing-card" data-listing-id="101">
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
							<a href="/listings/101" class="view-link">View Details</a>
						</div>
					</div>
					<div class="listing-card" data-listing-id="102">
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
							<a href="/listings/102" class="view-link">View Details</a>
						</div>
					</div>
					<div class="listing-card" data-listing-id="103">
						<img src="/images/loft.jpg" class="listing-image" alt="City Loft">
						<div class="listing-details">
							<h3 class="listing-title">Modern City Loft</h3>
							<p class="listing-location">New York, NY</p>
							<div class="listing-price">
								<span class="price-amount">$450</span>
								<span class="price-period">/ night</span>
							</div>
							<div class="listing-rating">
								<span class="rating-score">4.7</span>
								<span class="rating-count">(203 reviews)</span>
							</div>
							<a href="/listings/103" class="view-link">View Details</a>
						</div>
					</div>
				</main>
			</body>
			</html>
		`))
	}))
}

// =============================================================================
// In-Memory Repositories for E2E Testing
// =============================================================================

type inMemoryWorkflowRepository struct {
	mu        sync.RWMutex
	workflows map[uuid.UUID]*repository.Workflow
	names     map[string]uuid.UUID
}

func newInMemoryWorkflowRepository() *inMemoryWorkflowRepository {
	return &inMemoryWorkflowRepository{
		workflows: make(map[uuid.UUID]*repository.Workflow),
		names:     make(map[string]uuid.UUID),
	}
}

func (r *inMemoryWorkflowRepository) Create(workflow *repository.Workflow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.names[workflow.Name]; exists {
		return &repoError{msg: "workflow already exists"}
	}
	if workflow.ID == uuid.Nil {
		workflow.ID = uuid.New()
	}
	workflow.CreatedAt = time.Now().UTC()
	workflow.UpdatedAt = workflow.CreatedAt
	r.workflows[workflow.ID] = workflow
	r.names[workflow.Name] = workflow.ID
	return nil
}

func (r *inMemoryWorkflowRepository) GetByID(id uuid.UUID) (*repository.Workflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	workflow, exists := r.workflows[id]
	if !exists {
		return nil, &repoError{msg: "workflow not found"}
	}
	return workflow, nil
}

func (r *inMemoryWorkflowRepository) GetByName(name string) (*repository.Workflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, exists := r.names[name]
	if !exists {
		return nil, &repoError{msg: "workflow not found"}
	}
	return r.workflows[id], nil
}

func (r *inMemoryWorkflowRepository) GetAll() ([]*repository.Workflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*repository.Workflow, 0, len(r.workflows))
	for _, w := range r.workflows {
		result = append(result, w)
	}
	return result, nil
}

func (r *inMemoryWorkflowRepository) Update(workflow *repository.Workflow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.workflows[workflow.ID]; !exists {
		return &repoError{msg: "workflow not found"}
	}
	workflow.UpdatedAt = time.Now().UTC()
	r.workflows[workflow.ID] = workflow
	return nil
}

func (r *inMemoryWorkflowRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	workflow, exists := r.workflows[id]
	if !exists {
		return &repoError{msg: "workflow not found"}
	}
	delete(r.names, workflow.Name)
	delete(r.workflows, id)
	return nil
}

type inMemoryExecutionRepository struct {
	mu         sync.RWMutex
	executions map[uuid.UUID]*repository.Execution
}

func newInMemoryExecutionRepository() *inMemoryExecutionRepository {
	return &inMemoryExecutionRepository{
		executions: make(map[uuid.UUID]*repository.Execution),
	}
}

func (r *inMemoryExecutionRepository) Create(execution *repository.Execution) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if execution.ID == uuid.Nil {
		execution.ID = uuid.New()
	}
	execution.CreatedAt = time.Now().UTC()
	execution.UpdatedAt = execution.CreatedAt
	execution.StartedAt = time.Now().UTC()
	r.executions[execution.ID] = execution
	return nil
}

func (r *inMemoryExecutionRepository) GetByID(id uuid.UUID) (*repository.Execution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	execution, exists := r.executions[id]
	if !exists {
		return nil, &repoError{msg: "execution not found"}
	}
	return execution, nil
}

func (r *inMemoryExecutionRepository) GetByWorkflowID(workflowID uuid.UUID) ([]*repository.Execution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*repository.Execution, 0)
	for _, e := range r.executions {
		if e.WorkflowID == workflowID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *inMemoryExecutionRepository) GetAll() ([]*repository.Execution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*repository.Execution, 0, len(r.executions))
	for _, e := range r.executions {
		result = append(result, e)
	}
	return result, nil
}

func (r *inMemoryExecutionRepository) Update(execution *repository.Execution) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.executions[execution.ID]; !exists {
		return &repoError{msg: "execution not found"}
	}
	execution.UpdatedAt = time.Now().UTC()
	r.executions[execution.ID] = execution
	return nil
}

type inMemoryTaskLogRepository struct {
	mu       sync.RWMutex
	taskLogs map[uuid.UUID]*repository.TaskLog
}

func newInMemoryTaskLogRepository() *inMemoryTaskLogRepository {
	return &inMemoryTaskLogRepository{
		taskLogs: make(map[uuid.UUID]*repository.TaskLog),
	}
}

func (r *inMemoryTaskLogRepository) Create(taskLog *repository.TaskLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if taskLog.ID == uuid.Nil {
		taskLog.ID = uuid.New()
	}
	taskLog.CreatedAt = time.Now().UTC()
	r.taskLogs[taskLog.ID] = taskLog
	return nil
}

func (r *inMemoryTaskLogRepository) Update(taskLog *repository.TaskLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.taskLogs[taskLog.ID]; !exists {
		return &repoError{msg: "task log not found"}
	}
	r.taskLogs[taskLog.ID] = taskLog
	return nil
}

func (r *inMemoryTaskLogRepository) GetByExecutionID(executionID uuid.UUID) ([]*repository.TaskLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*repository.TaskLog, 0)
	for _, tl := range r.taskLogs {
		if tl.ExecutionID == executionID {
			result = append(result, tl)
		}
	}
	return result, nil
}

type repoError struct {
	msg string
}

func (e *repoError) Error() string {
	return e.msg
}

// =============================================================================
// Test Router Setup
// =============================================================================

func setupTestRouter(
	workflowRepo repository.WorkflowRepository,
	execRepo repository.ExecutionRepository,
	taskLogRepo repository.TaskLogRepository,
	eng *engine.Engine,
) *gin.Engine {
	router := gin.New()

	// Workflow endpoints
	router.POST("/workflows", handleCreateWorkflow(workflowRepo))
	router.GET("/workflows/:id", handleGetWorkflow(workflowRepo))
	router.DELETE("/workflows/:id", handleDeleteWorkflow(workflowRepo))

	// Execution endpoints
	router.POST("/workflows/:id/run", handleRunWorkflow(workflowRepo, execRepo, taskLogRepo, eng))
	router.GET("/executions/:id", handleGetExecution(execRepo, taskLogRepo))

	return router
}

// Handler implementations for test (simplified versions)

func handleCreateWorkflow(repo repository.WorkflowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name       string         `json:"name" binding:"required"`
			Definition map[string]any `json:"definition" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		defBytes, _ := json.Marshal(req.Definition)
		workflow := &repository.Workflow{
			Name:       req.Name,
			Definition: datatypes.JSON(defBytes),
		}
		if err := repo.Create(workflow); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, workflow)
	}
}

func handleGetWorkflow(repo repository.WorkflowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		workflow, err := repo.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, workflow)
	}
}

func handleDeleteWorkflow(repo repository.WorkflowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := repo.Delete(id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}

func handleRunWorkflow(
	workflowRepo repository.WorkflowRepository,
	execRepo repository.ExecutionRepository,
	taskLogRepo repository.TaskLogRepository,
	eng *engine.Engine,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		workflow, err := workflowRepo.GetByID(workflowID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		workflowDef, err := workflow.ToWorkflowDefinition()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger := repository.NewExecutionLoggerAdapter(execRepo, taskLogRepo)

		execution := &repository.Execution{
			WorkflowID: workflowID,
			Status:     "pending",
		}
		if err := execRepo.Create(execution); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		go func(execID uuid.UUID) {
			eng.ExecuteWithLogging(*workflowDef, workflowID, logger, &execID)
		}(execution.ID)

		c.JSON(http.StatusAccepted, gin.H{
			"execution_id": execution.ID,
			"workflow_id":  workflowID,
			"status":       "pending",
		})
	}
}

func handleGetExecution(execRepo repository.ExecutionRepository, taskLogRepo repository.TaskLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		execution, err := execRepo.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		// Load task logs
		taskLogs, _ := taskLogRepo.GetByExecutionID(id)
		execution.TaskLogs = make([]repository.TaskLog, len(taskLogs))
		for i, tl := range taskLogs {
			execution.TaskLogs[i] = *tl
		}

		c.JSON(http.StatusOK, execution)
	}
}

// =============================================================================
// E2E Tests
// =============================================================================

func TestAirbnbWorkflowValidation_E2E(t *testing.T) {
	// 1. Setup mock Airbnb server
	mockServer := createMockAirbnbServer()
	defer mockServer.Close()

	// 2. Setup repositories
	workflowRepo := newInMemoryWorkflowRepository()
	execRepo := newInMemoryExecutionRepository()
	taskLogRepo := newInMemoryTaskLogRepository()

	// 3. Setup engine with all task executors
	registry := engine.NewRegistry()
	tasks.RegisterHTTPTask(registry)
	tasks.RegisterHTMLParserTask(registry)
	tasks.RegisterTransformTask(registry)
	eng := engine.NewEngine(registry)

	// 4. Setup test router
	router := setupTestRouter(workflowRepo, execRepo, taskLogRepo, eng)

	// 5. Create workflow via POST /workflows
	workflowDefinition := map[string]any{
		"name": "airbnb-price-monitor",
		"tasks": []any{
			map[string]any{
				"id":   "fetch_listings",
				"type": "http_request",
				"config": map[string]any{
					"method": "GET",
					"url":    mockServer.URL,
				},
			},
			map[string]any{
				"id":   "parse_listings",
				"type": "html_parser",
				"config": map[string]any{
					"html_source": "fetch_listings_result",
					"selectors": []any{
						map[string]any{"name": "titles", "selector": ".listing-title", "multiple": true},
						map[string]any{"name": "prices", "selector": ".price-amount", "multiple": true},
						map[string]any{"name": "locations", "selector": ".listing-location", "multiple": true},
						map[string]any{"name": "ratings", "selector": ".rating-score", "multiple": true},
					},
				},
			},
			map[string]any{
				"id":   "format_output",
				"type": "transform",
				"config": map[string]any{
					"data_source":   "parse_listings_result",
					"template":      `{"listing_count": {{len (index . 0).titles}}, "source": "airbnb-validation"}`,
					"output_format": "json",
				},
			},
		},
	}

	createReqBody, _ := json.Marshal(map[string]any{
		"name":       "airbnb-price-monitor",
		"definition": workflowDefinition,
	})

	createReq := httptest.NewRequest(http.MethodPost, "/workflows", bytes.NewReader(createReqBody))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)

	require.Equal(t, http.StatusCreated, createResp.Code, "Failed to create workflow")

	var createdWorkflow repository.Workflow
	err := json.Unmarshal(createResp.Body.Bytes(), &createdWorkflow)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, createdWorkflow.ID)

	t.Logf("Created workflow with ID: %s", createdWorkflow.ID)

	// 6. Execute workflow via POST /workflows/:id/run
	runReq := httptest.NewRequest(http.MethodPost, "/workflows/"+createdWorkflow.ID.String()+"/run", nil)
	runResp := httptest.NewRecorder()
	router.ServeHTTP(runResp, runReq)

	// AC2: Verify 202 Accepted response
	require.Equal(t, http.StatusAccepted, runResp.Code, "Expected 202 Accepted")

	var runResponse map[string]any
	err = json.Unmarshal(runResp.Body.Bytes(), &runResponse)
	require.NoError(t, err)

	executionID, err := uuid.Parse(runResponse["execution_id"].(string))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, executionID)

	t.Logf("Execution started with ID: %s", executionID)

	// 7. Poll GET /executions/:id until completed (with timeout)
	var execution *repository.Execution
	maxAttempts := 50
	for range maxAttempts {
		time.Sleep(100 * time.Millisecond)

		getExecReq := httptest.NewRequest(http.MethodGet, "/executions/"+executionID.String(), nil)
		getExecResp := httptest.NewRecorder()
		router.ServeHTTP(getExecResp, getExecReq)

		require.Equal(t, http.StatusOK, getExecResp.Code)

		execution = &repository.Execution{}
		err = json.Unmarshal(getExecResp.Body.Bytes(), execution)
		require.NoError(t, err)

		if execution.Status == "completed" || execution.Status == "failed" {
			break
		}
	}

	// AC3: Verify execution completed successfully
	require.NotNil(t, execution)
	assert.Equal(t, "completed", execution.Status, "Execution should complete successfully")

	t.Logf("Execution completed with status: %s", execution.Status)

	// AC3: Verify TaskLogs were created for each task
	assert.Len(t, execution.TaskLogs, 3, "Should have 3 task logs (fetch, parse, transform)")

	taskLogTypes := make(map[string]bool)
	for _, tl := range execution.TaskLogs {
		taskLogTypes[tl.TaskID] = true
		assert.Equal(t, "success", tl.Status, "Task %s should succeed", tl.TaskID)
	}

	assert.True(t, taskLogTypes["fetch_listings"], "fetch_listings task log should exist")
	assert.True(t, taskLogTypes["parse_listings"], "parse_listings task log should exist")
	assert.True(t, taskLogTypes["format_output"], "format_output task log should exist")

	// AC4: Verify context_snapshot contains extracted data
	require.NotNil(t, execution.ContextSnapshot, "Context snapshot should not be nil")

	var contextSnapshot map[string]any
	err = json.Unmarshal(execution.ContextSnapshot, &contextSnapshot)
	require.NoError(t, err)

	// Verify fetch_listings_result exists
	fetchResult, exists := contextSnapshot["fetch_listings_result"]
	assert.True(t, exists, "fetch_listings_result should exist in context")
	assert.NotNil(t, fetchResult)

	// Verify parse_listings_result exists and contains data
	parseResult, exists := contextSnapshot["parse_listings_result"]
	assert.True(t, exists, "parse_listings_result should exist in context")
	assert.NotNil(t, parseResult)

	// Verify parsed data structure
	parseResultSlice, ok := parseResult.([]any)
	if assert.True(t, ok, "parse_listings_result should be a slice") && len(parseResultSlice) > 0 {
		firstResult := parseResultSlice[0].(map[string]any)

		// Verify titles were extracted
		titles, exists := firstResult["titles"]
		assert.True(t, exists, "titles should exist")
		titlesSlice := titles.([]any)
		assert.Len(t, titlesSlice, 3, "Should have 3 listings")
		assert.Equal(t, "Beautiful Beach House", titlesSlice[0])
		assert.Equal(t, "Cozy Mountain Cabin", titlesSlice[1])
		assert.Equal(t, "Modern City Loft", titlesSlice[2])

		// Verify prices were extracted
		prices, exists := firstResult["prices"]
		assert.True(t, exists, "prices should exist")
		pricesSlice := prices.([]any)
		assert.Len(t, pricesSlice, 3)
		assert.Equal(t, "$350", pricesSlice[0])
		assert.Equal(t, "$275", pricesSlice[1])
		assert.Equal(t, "$450", pricesSlice[2])

		// Verify locations were extracted
		locations, exists := firstResult["locations"]
		assert.True(t, exists, "locations should exist")
		locationsSlice := locations.([]any)
		assert.Len(t, locationsSlice, 3)
		assert.Equal(t, "Malibu, CA", locationsSlice[0])

		// Verify ratings were extracted
		ratings, exists := firstResult["ratings"]
		assert.True(t, exists, "ratings should exist")
		ratingsSlice := ratings.([]any)
		assert.Len(t, ratingsSlice, 3)
		assert.Equal(t, "4.9", ratingsSlice[0])
	}

	// Verify format_output_result exists
	formatResult, exists := contextSnapshot["format_output_result"]
	assert.True(t, exists, "format_output_result should exist in context")
	assert.NotNil(t, formatResult)

	// Verify transform output structure
	formatResultMap, ok := formatResult.(map[string]any)
	if assert.True(t, ok, "format_output_result should be a map") {
		listingCount, exists := formatResultMap["listing_count"]
		assert.True(t, exists, "listing_count should exist")
		assert.Equal(t, float64(3), listingCount, "listing_count should be 3")

		source, exists := formatResultMap["source"]
		assert.True(t, exists, "source should exist")
		assert.Equal(t, "airbnb-validation", source)
	}

	t.Log("All E2E validations passed!")

	// 8. Cleanup - delete workflow
	deleteReq := httptest.NewRequest(http.MethodDelete, "/workflows/"+createdWorkflow.ID.String(), nil)
	deleteResp := httptest.NewRecorder()
	router.ServeHTTP(deleteResp, deleteReq)
	assert.Equal(t, http.StatusNoContent, deleteResp.Code)
}

func TestAirbnbWorkflowValidation_WorkflowNotFound(t *testing.T) {
	workflowRepo := newInMemoryWorkflowRepository()
	execRepo := newInMemoryExecutionRepository()
	taskLogRepo := newInMemoryTaskLogRepository()

	registry := engine.NewRegistry()
	eng := engine.NewEngine(registry)

	router := setupTestRouter(workflowRepo, execRepo, taskLogRepo, eng)

	// Try to run a non-existent workflow
	runReq := httptest.NewRequest(http.MethodPost, "/workflows/"+uuid.New().String()+"/run", nil)
	runResp := httptest.NewRecorder()
	router.ServeHTTP(runResp, runReq)

	assert.Equal(t, http.StatusNotFound, runResp.Code)
}

func TestAirbnbWorkflowValidation_InvalidWorkflowID(t *testing.T) {
	workflowRepo := newInMemoryWorkflowRepository()
	execRepo := newInMemoryExecutionRepository()
	taskLogRepo := newInMemoryTaskLogRepository()

	registry := engine.NewRegistry()
	eng := engine.NewEngine(registry)

	router := setupTestRouter(workflowRepo, execRepo, taskLogRepo, eng)

	runReq := httptest.NewRequest(http.MethodPost, "/workflows/invalid-uuid/run", nil)
	runResp := httptest.NewRecorder()
	router.ServeHTTP(runResp, runReq)

	assert.Equal(t, http.StatusBadRequest, runResp.Code)
}
