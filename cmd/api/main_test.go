package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

// mockWorkflowRepository is a mock implementation for testing
type mockWorkflowRepository struct{}

func (m *mockWorkflowRepository) Create(workflow *repository.Workflow) error {
	return nil
}

func (m *mockWorkflowRepository) GetByID(id uuid.UUID) (*repository.Workflow, error) {
	return nil, nil
}

func (m *mockWorkflowRepository) GetByName(name string) (*repository.Workflow, error) {
	return nil, nil
}

func (m *mockWorkflowRepository) GetAll() ([]*repository.Workflow, error) {
	return nil, nil
}

func (m *mockWorkflowRepository) Update(workflow *repository.Workflow) error {
	return nil
}

func (m *mockWorkflowRepository) Delete(id uuid.UUID) error {
	return nil
}

// Helper function to create test router
// Note: mockExecutionRepository and mockTaskLogRepository are defined in handlers_test.go
func createTestRouter() *gin.Engine {
	mockWorkflowRepo := &mockWorkflowRepository{}
	mockExecRepo := &mockExecutionRepository{}
	mockTaskLogRepo := &mockTaskLogRepository{}
	mockEngine := engine.NewEngine(engine.NewRegistry())
	return setupRouter(mockWorkflowRepo, mockExecRepo, mockTaskLogRepo, mockEngine)
}

func TestHealthEndpoint(t *testing.T) {
	// For health tests, we don't need a real repository since health check
	// will fail without DB, but we can still test the endpoint structure
	// We'll skip DB-dependent health checks in unit tests
	router := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	// Health endpoint will return 500 if DB is not connected, which is expected in tests
	// We just verify it returns JSON
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "status")
}

func TestHealthEndpointReturnsJSON(t *testing.T) {
	router := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestHealthEndpointStructure(t *testing.T) {
	router := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	_, exists := response["status"]
	assert.True(t, exists, "Response should contain 'status' field")
}

func TestSetupRouter(t *testing.T) {
	router := createTestRouter()
	assert.NotNil(t, router, "Router should not be nil")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	// Router should respond (even if DB is not connected)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestGetPortDefault(t *testing.T) {
	os.Unsetenv("PORT")
	port := getPort()
	assert.Equal(t, "8080", port, "Default port should be 8080")
}

func TestGetPortFromEnv(t *testing.T) {
	os.Setenv("PORT", "3000")
	defer os.Unsetenv("PORT")

	port := getPort()
	assert.Equal(t, "3000", port, "Port should be read from environment variable")
}
