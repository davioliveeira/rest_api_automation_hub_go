package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
	"github.com/davioliveira/rest_api_automation_hub_go/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// CreateWorkflowRequest represents the request body for creating a workflow
type CreateWorkflowRequest struct {
	Name       string                 `json:"name" binding:"required"`
	Definition map[string]interface{} `json:"definition" binding:"required"`
}

// handleCreateWorkflow handles POST /workflows
func handleCreateWorkflow(repo repository.WorkflowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateWorkflowRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// Convert definition to JSONB
		defJSONBytes, err := json.Marshal(req.Definition)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow definition"})
			return
		}
		defJSON := datatypes.JSON(defJSONBytes)

		workflow := &repository.Workflow{
			Name:       req.Name,
			Definition: defJSON,
		}

		if err := repo.Create(workflow); err != nil {
			// Check for duplicate name
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				c.JSON(http.StatusConflict, gin.H{"error": "Workflow with this name already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, workflow)
	}
}

// handleGetWorkflow handles GET /workflows/:id
func handleGetWorkflow(repo repository.WorkflowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
			return
		}

		workflow, err := repo.GetByID(id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve workflow"})
			return
		}

		c.JSON(http.StatusOK, workflow)
	}
}

// handleListWorkflows handles GET /workflows
func handleListWorkflows(repo repository.WorkflowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		workflows, err := repo.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve workflows"})
			return
		}

		c.JSON(http.StatusOK, workflows)
	}
}

// handleUpdateWorkflow handles PUT /workflows/:id
func handleUpdateWorkflow(repo repository.WorkflowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
			return
		}

		var req CreateWorkflowRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// Get existing workflow
		workflow, err := repo.GetByID(id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve workflow"})
			return
		}

		// Update fields
		workflow.Name = req.Name
		defJSONBytes, err := json.Marshal(req.Definition)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow definition"})
			return
		}
		workflow.Definition = datatypes.JSON(defJSONBytes)

		if err := repo.Update(workflow); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				c.JSON(http.StatusConflict, gin.H{"error": "Workflow with this name already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow"})
			return
		}

		c.JSON(http.StatusOK, workflow)
	}
}

// handleDeleteWorkflow handles DELETE /workflows/:id
func handleDeleteWorkflow(repo repository.WorkflowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
			return
		}

		if err := repo.Delete(id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow"})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

// handleRunWorkflow handles POST /workflows/:id/run
func handleRunWorkflow(
	workflowRepo repository.WorkflowRepository,
	execRepo repository.ExecutionRepository,
	taskLogRepo repository.TaskLogRepository,
	executionEngine *engine.Engine,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		workflowID, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
			return
		}

		// Load workflow from database
		workflow, err := workflowRepo.GetByID(workflowID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve workflow"})
			return
		}

		// Convert to engine.WorkflowDefinition
		workflowDef, err := workflow.ToWorkflowDefinition()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow definition", "details": err.Error()})
			return
		}

		// Create execution logger adapter
		logger := repository.NewExecutionLoggerAdapter(execRepo, taskLogRepo)

		// Create execution record first to get the ID
		execution := &repository.Execution{
			WorkflowID: workflowID,
			Status:     "pending",
		}
		if err := execRepo.Create(execution); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create execution record"})
			return
		}

		// Execute workflow asynchronously
		go func(execID uuid.UUID) {
			// Execute with logging using the execution ID we created
			execRecord, execErr := executionEngine.ExecuteWithLogging(*workflowDef, workflowID, logger, &execID)
			if execErr != nil {
				// Error is already logged in ExecuteWithLogging
				return
			}
			_ = execRecord // Execution is already updated by ExecuteWithLogging
		}(execution.ID)

		c.JSON(http.StatusAccepted, gin.H{
			"execution_id": execution.ID,
			"workflow_id":  workflowID,
			"status":       "pending",
			"message":      "Workflow execution started",
		})
	}
}

// handleGetExecution handles GET /executions/:id
func handleGetExecution(execRepo repository.ExecutionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		executionID, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
			return
		}

		execution, err := execRepo.GetByID(executionID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve execution"})
			return
		}

		c.JSON(http.StatusOK, execution)
	}
}

// handleListExecutions handles GET /executions
func handleListExecutions(execRepo repository.ExecutionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		executions, err := execRepo.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve executions"})
			return
		}

		c.JSON(http.StatusOK, executions)
	}
}
