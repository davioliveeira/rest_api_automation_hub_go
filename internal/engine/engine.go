package engine

import (
	"fmt"
	"log/slog"
)

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
