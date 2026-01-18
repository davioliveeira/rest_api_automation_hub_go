package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
	"github.com/davioliveira/rest_api_automation_hub_go/internal/tasks"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Initialize task registry
	registry := engine.NewRegistry()

	// Register task executors
	tasks.RegisterHTTPTask(registry) // Story 2.1

	// Create engine with registry
	_ = engine.NewEngine(registry) // Will be used in Epic 3 for workflow execution

	router := setupRouter()
	port := getPort()

	slog.Info("Starting GoAutomation Hub API Server", "port", port)
	if err := router.Run(":" + port); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/health", healthHandler)
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
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
