package engine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ExecutionLogger interface for logging executions (to avoid import cycle)
type ExecutionLogger interface {
	CreateExecution(execution *ExecutionRecord) error
	UpdateExecution(execution *ExecutionRecord) error
	CreateTaskLog(taskLog *TaskLogRecord) error
	UpdateTaskLog(taskLog *TaskLogRecord) error
}

// ExecutionRecord represents an execution record for logging
type ExecutionRecord struct {
	ID              uuid.UUID
	WorkflowID      uuid.UUID
	Status          string
	ContextSnapshot datatypes.JSON
	StartedAt       time.Time
	CompletedAt     *time.Time
}

// TaskLogRecord represents a task log record for logging
type TaskLogRecord struct {
	ID          uuid.UUID
	ExecutionID uuid.UUID
	TaskID      string
	TaskType    string
	Status      string
	Input       datatypes.JSON
	Output      datatypes.JSON
	Error       string
	StartedAt   time.Time
	CompletedAt time.Time
}

// Engine orchestrates workflow execution with sequential task processing
type Engine struct {
	context  *ExecutionContext
	registry *Registry
}

// NewEngine creates a new Engine instance with an initialized ExecutionContext and Registry.
// If registry is nil, a new empty registry will be created.
func NewEngine(registry *Registry) *Engine {
	slog.Info("Initializing new execution engine")
	if registry == nil {
		registry = NewRegistry()
	}
	return &Engine{
		context:  NewExecutionContext(),
		registry: registry,
	}
}

// Execute processes a workflow by iterating through its tasks sequentially.
// Each task is looked up in the registry, executed with the current context,
// and its result is stored for subsequent tasks to access.
func (e *Engine) Execute(workflow WorkflowDefinition) error {
	slog.Info("Starting workflow execution", "workflow", workflow.Name, "task_count", len(workflow.Tasks))

	for i, task := range workflow.Tasks {
		slog.Info("Processing task",
			"index", i,
			"id", task.ID,
			"type", task.Type,
		)

		// Lookup executor from registry
		executor, err := e.registry.Get(task.Type)
		if err != nil {
			slog.Error("Task executor not found", "type", task.Type, "error", err)
			return fmt.Errorf("task executor not found for type '%s': %w", task.Type, err)
		}

		// Execute task with current context and config
		result := executor.Execute(e.context, task.Config)

		// Handle result
		if result.Status == "success" {
			slog.Info("Task completed successfully",
				"id", task.ID,
				"type", task.Type,
			)
			// Store result in context for subsequent tasks
			e.context.Set(task.ID+"_result", result.Output)
		} else {
			slog.Error("Task failed",
				"id", task.ID,
				"type", task.Type,
				"error", result.Error,
			)
			return fmt.Errorf("task %s failed: %s", task.ID, result.Error)
		}
	}

	slog.Info("Workflow execution completed successfully",
		"workflow", workflow.Name,
		"total_tasks", len(workflow.Tasks),
	)

	return nil
}

// GetContext returns the engine's ExecutionContext for inspection or testing
func (e *Engine) GetContext() *ExecutionContext {
	return e.context
}

// ExecuteWithLogging processes a workflow with full logging to database.
// It creates an Execution record (or updates existing if executionID is provided),
// logs each task execution, and updates the execution status.
// If executionID is nil, a new execution will be created.
func (e *Engine) ExecuteWithLogging(
	workflow WorkflowDefinition,
	workflowID uuid.UUID,
	logger ExecutionLogger,
	executionID *uuid.UUID,
) (*ExecutionRecord, error) {
	// Create new context for this execution
	e.context = NewExecutionContext()

	// Create or update Execution record
	var execution *ExecutionRecord
	if executionID != nil {
		// Use existing execution ID
		execution = &ExecutionRecord{
			ID:         *executionID,
			WorkflowID: workflowID,
			Status:     "running",
			StartedAt:  time.Now().UTC(),
		}
		// Update existing execution
		if err := logger.UpdateExecution(execution); err != nil {
			// If update fails, try to create (execution might not exist yet)
			if err := logger.CreateExecution(execution); err != nil {
				return nil, fmt.Errorf("failed to create/update execution record: %w", err)
			}
		}
	} else {
		// Create new execution
		execution = &ExecutionRecord{
			ID:         uuid.New(),
			WorkflowID: workflowID,
			Status:     "running",
			StartedAt:  time.Now().UTC(),
		}
		if err := logger.CreateExecution(execution); err != nil {
			return nil, fmt.Errorf("failed to create execution record: %w", err)
		}
	}

	slog.Info("Starting workflow execution with logging",
		"execution_id", execution.ID,
		"workflow", workflow.Name,
		"task_count", len(workflow.Tasks),
	)

	var executionError error

	// Execute each task with logging
	for i, task := range workflow.Tasks {
		taskStartTime := time.Now().UTC()

		// Create TaskLog record
		taskLog := &TaskLogRecord{
			ID:          uuid.New(),
			ExecutionID: execution.ID,
			TaskID:      task.ID,
			TaskType:    task.Type,
			Status:      "running",
			StartedAt:   taskStartTime,
		}

		// Serialize task config as input
		if configJSON, err := json.Marshal(task.Config); err == nil {
			taskLog.Input = datatypes.JSON(configJSON)
		}

		// Log task start
		if err := logger.CreateTaskLog(taskLog); err != nil {
			slog.Error("Failed to create task log", "error", err, "task_id", task.ID)
			// Continue execution even if logging fails
		}

		slog.Info("Processing task",
			"execution_id", execution.ID,
			"index", i,
			"id", task.ID,
			"type", task.Type,
		)

		// Lookup executor from registry
		executor, err := e.registry.Get(task.Type)
		if err != nil {
			slog.Error("Task executor not found", "type", task.Type, "error", err)
			taskLog.Status = "failed"
			taskLog.Error = fmt.Sprintf("task executor not found for type '%s': %v", task.Type, err)
			taskLog.CompletedAt = time.Now().UTC()
			logger.UpdateTaskLog(taskLog) // Update task log
			executionError = fmt.Errorf("task executor not found for type '%s': %w", task.Type, err)
			break
		}

		// Execute task with current context and config
		result := executor.Execute(e.context, task.Config)

		// Update TaskLog with result
		taskLog.CompletedAt = time.Now().UTC()
		if result.Status == "success" {
			taskLog.Status = "success"
			slog.Info("Task completed successfully",
				"execution_id", execution.ID,
				"id", task.ID,
				"type", task.Type,
			)
			// Store result in context for subsequent tasks
			e.context.Set(task.ID+"_result", result.Output)

			// Serialize output
			if outputJSON, err := json.Marshal(result.Output); err == nil {
				taskLog.Output = datatypes.JSON(outputJSON)
			}
		} else {
			taskLog.Status = "failed"
			taskLog.Error = result.Error
			slog.Error("Task failed",
				"execution_id", execution.ID,
				"id", task.ID,
				"type", task.Type,
				"error", result.Error,
			)
			executionError = fmt.Errorf("task %s failed: %s", task.ID, result.Error)
		}

		// Update task log
		if err := logger.UpdateTaskLog(taskLog); err != nil {
			slog.Error("Failed to update task log", "error", err, "task_id", task.ID)
		}

		// If task failed, stop execution
		if result.Status != "success" {
			break
		}
	}

	// Save context snapshot
	contextSnapshot := e.context.GetAll()
	if snapshotJSON, err := json.Marshal(contextSnapshot); err == nil {
		execution.ContextSnapshot = datatypes.JSON(snapshotJSON)
	}

	// Update Execution status
	completedAt := time.Now().UTC()
	execution.CompletedAt = &completedAt
	if executionError != nil {
		execution.Status = "failed"
	} else {
		execution.Status = "completed"
		slog.Info("Workflow execution completed successfully",
			"execution_id", execution.ID,
			"workflow", workflow.Name,
			"total_tasks", len(workflow.Tasks),
		)
	}

	if err := logger.UpdateExecution(execution); err != nil {
		slog.Error("Failed to update execution", "error", err, "execution_id", execution.ID)
		return execution, fmt.Errorf("failed to update execution: %w", err)
	}

	return execution, executionError
}
