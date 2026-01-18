package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
	"github.com/davioliveira/rest_api_automation_hub_go/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockWorkflowRepositoryForHandlers is a mock implementation for handler tests
type mockWorkflowRepositoryForHandlers struct {
	workflows map[uuid.UUID]*repository.Workflow
	names     map[string]bool
}

func newMockWorkflowRepository() *mockWorkflowRepositoryForHandlers {
	return &mockWorkflowRepositoryForHandlers{
		workflows: make(map[uuid.UUID]*repository.Workflow),
		names:     make(map[string]bool),
	}
}

func (m *mockWorkflowRepositoryForHandlers) Create(workflow *repository.Workflow) error {
	if m.names[workflow.Name] {
		// Simulate duplicate key error
		return assert.AnError
	}
	if workflow.ID == uuid.Nil {
		workflow.ID = uuid.New()
	}
	m.workflows[workflow.ID] = workflow
	m.names[workflow.Name] = true
	return nil
}

func (m *mockWorkflowRepositoryForHandlers) GetByID(id uuid.UUID) (*repository.Workflow, error) {
	workflow, exists := m.workflows[id]
	if !exists {
		return nil, &repositoryError{message: "workflow not found: " + id.String()}
	}
	return workflow, nil
}

func (m *mockWorkflowRepositoryForHandlers) GetByName(name string) (*repository.Workflow, error) {
	for _, w := range m.workflows {
		if w.Name == name {
			return w, nil
		}
	}
	return nil, &repositoryError{message: "workflow not found: " + name}
}

func (m *mockWorkflowRepositoryForHandlers) GetAll() ([]*repository.Workflow, error) {
	workflows := make([]*repository.Workflow, 0, len(m.workflows))
	for _, w := range m.workflows {
		workflows = append(workflows, w)
	}
	return workflows, nil
}

func (m *mockWorkflowRepositoryForHandlers) Update(workflow *repository.Workflow) error {
	if _, exists := m.workflows[workflow.ID]; !exists {
		return &repositoryError{message: "workflow not found: " + workflow.ID.String()}
	}
	m.workflows[workflow.ID] = workflow
	return nil
}

func (m *mockWorkflowRepositoryForHandlers) Delete(id uuid.UUID) error {
	if _, exists := m.workflows[id]; !exists {
		return &repositoryError{message: "workflow not found: " + id.String()}
	}
	delete(m.workflows, id)
	return nil
}

// repositoryError is a simple error type for testing
type repositoryError struct {
	message string
}

func (e *repositoryError) Error() string {
	return e.message
}

// mockExecutionRepository for handlers tests
type mockExecutionRepository struct{}

func (m *mockExecutionRepository) Create(execution *repository.Execution) error {
	return nil
}

func (m *mockExecutionRepository) GetByID(id uuid.UUID) (*repository.Execution, error) {
	return nil, nil
}

func (m *mockExecutionRepository) GetByWorkflowID(workflowID uuid.UUID) ([]*repository.Execution, error) {
	return nil, nil
}

func (m *mockExecutionRepository) GetAll() ([]*repository.Execution, error) {
	return nil, nil
}

func (m *mockExecutionRepository) Update(execution *repository.Execution) error {
	return nil
}

// mockTaskLogRepository for handlers tests
type mockTaskLogRepository struct{}

func (m *mockTaskLogRepository) Create(taskLog *repository.TaskLog) error {
	return nil
}

func (m *mockTaskLogRepository) Update(taskLog *repository.TaskLog) error {
	return nil
}

func (m *mockTaskLogRepository) GetByExecutionID(executionID uuid.UUID) ([]*repository.TaskLog, error) {
	return nil, nil
}

func TestHandleCreateWorkflow(t *testing.T) {
	repo := newMockWorkflowRepository()
	mockExecRepo := &mockExecutionRepository{}
	mockTaskLogRepo := &mockTaskLogRepository{}
	mockEngine := engine.NewEngine(engine.NewRegistry())
	router := setupRouter(repo, mockExecRepo, mockTaskLogRepo, mockEngine)

	reqBody := CreateWorkflowRequest{
		Name: "test-workflow",
		Definition: map[string]interface{}{
			"name":  "test-workflow",
			"tasks": []interface{}{},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/workflows", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response repository.Workflow
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-workflow", response.Name)
	assert.NotEqual(t, uuid.Nil, response.ID)
}

func TestHandleCreateWorkflowInvalidJSON(t *testing.T) {
	repo := newMockWorkflowRepository()
	mockExecRepo := &mockExecutionRepository{}
	mockTaskLogRepo := &mockTaskLogRepository{}
	mockEngine := engine.NewEngine(engine.NewRegistry())
	router := setupRouter(repo, mockExecRepo, mockTaskLogRepo, mockEngine)

	req := httptest.NewRequest(http.MethodPost, "/workflows", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetWorkflow(t *testing.T) {
	repo := newMockWorkflowRepository()
	mockExecRepo := &mockExecutionRepository{}
	mockTaskLogRepo := &mockTaskLogRepository{}
	mockEngine := engine.NewEngine(engine.NewRegistry())
	router := setupRouter(repo, mockExecRepo, mockTaskLogRepo, mockEngine)

	// Create a workflow first
	workflow := &repository.Workflow{
		ID:   uuid.New(),
		Name: "test-workflow",
	}
	repo.Create(workflow)

	req := httptest.NewRequest(http.MethodGet, "/workflows/"+workflow.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response repository.Workflow
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, workflow.ID, response.ID)
	assert.Equal(t, "test-workflow", response.Name)
}

func TestHandleGetWorkflowNotFound(t *testing.T) {
	repo := newMockWorkflowRepository()
	mockExecRepo := &mockExecutionRepository{}
	mockTaskLogRepo := &mockTaskLogRepository{}
	mockEngine := engine.NewEngine(engine.NewRegistry())
	router := setupRouter(repo, mockExecRepo, mockTaskLogRepo, mockEngine)

	req := httptest.NewRequest(http.MethodGet, "/workflows/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleListWorkflows(t *testing.T) {
	repo := newMockWorkflowRepository()
	mockExecRepo := &mockExecutionRepository{}
	mockTaskLogRepo := &mockTaskLogRepository{}
	mockEngine := engine.NewEngine(engine.NewRegistry())
	router := setupRouter(repo, mockExecRepo, mockTaskLogRepo, mockEngine)

	// Create some workflows
	workflow1 := &repository.Workflow{ID: uuid.New(), Name: "workflow-1"}
	workflow2 := &repository.Workflow{ID: uuid.New(), Name: "workflow-2"}
	repo.Create(workflow1)
	repo.Create(workflow2)

	req := httptest.NewRequest(http.MethodGet, "/workflows", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []repository.Workflow
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.GreaterOrEqual(t, len(response), 2)
}

func TestHandleDeleteWorkflow(t *testing.T) {
	repo := newMockWorkflowRepository()
	mockExecRepo := &mockExecutionRepository{}
	mockTaskLogRepo := &mockTaskLogRepository{}
	mockEngine := engine.NewEngine(engine.NewRegistry())
	router := setupRouter(repo, mockExecRepo, mockTaskLogRepo, mockEngine)

	// Create a workflow first
	workflow := &repository.Workflow{ID: uuid.New(), Name: "test-workflow"}
	repo.Create(workflow)

	req := httptest.NewRequest(http.MethodDelete, "/workflows/"+workflow.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
