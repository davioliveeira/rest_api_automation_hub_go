package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
	"github.com/davioliveira/rest_api_automation_hub_go/internal/repository"
	"github.com/davioliveira/rest_api_automation_hub_go/internal/tasks"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Initialize database
	if err := repository.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repository.CloseDB()

	// Run migrations
	if err := repository.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	workflowRepo := repository.NewWorkflowRepository(repository.DB)
	execRepo := repository.NewExecutionRepository(repository.DB)
	taskLogRepo := repository.NewTaskLogRepository(repository.DB)

	// Initialize task registry
	registry := engine.NewRegistry()

	// Register task executors
	tasks.RegisterHTTPTask(registry)      // Story 2.1
	tasks.RegisterTransformTask(registry) // Story 2.2
	tasks.RegisterHTMLParserTask(registry) // Story 2.3

	// Create engine with registry
	executionEngine := engine.NewEngine(registry)

	router := setupRouter(workflowRepo, execRepo, taskLogRepo, executionEngine)
	port := getPort()

	slog.Info("Starting GoAutomation Hub API Server", "port", port)
	if err := router.Run(":" + port); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func setupRouter(
	workflowRepo repository.WorkflowRepository,
	execRepo repository.ExecutionRepository,
	taskLogRepo repository.TaskLogRepository,
	executionEngine *engine.Engine,
) *gin.Engine {
	router := gin.Default()

	// Health endpoint (includes database check)
	router.GET("/health", healthHandler)

	// Workflow endpoints
	router.POST("/workflows", handleCreateWorkflow(workflowRepo))
	router.GET("/workflows/:id", handleGetWorkflow(workflowRepo))
	router.GET("/workflows", handleListWorkflows(workflowRepo))
	router.PUT("/workflows/:id", handleUpdateWorkflow(workflowRepo))
	router.DELETE("/workflows/:id", handleDeleteWorkflow(workflowRepo))

	// Execution endpoints (Story 3.2)
	router.POST("/workflows/:id/run", handleRunWorkflow(workflowRepo, execRepo, taskLogRepo, executionEngine))
	router.GET("/executions", handleListExecutions(execRepo))
	router.GET("/executions/:id", handleGetExecution(execRepo))

	return router
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8080"
	}
	return port
}

func healthHandler(c *gin.Context) {
	slog.Info("Health check requested")
	dbHealthy := repository.HealthCheck() == nil
	if !dbHealthy {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":   "unhealthy",
			"database": "disconnected",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": "connected",
	})
}
