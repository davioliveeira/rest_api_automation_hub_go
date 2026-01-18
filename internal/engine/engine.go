package engine

import "log/slog"

// Engine orchestrates workflow execution with sequential task processing
type Engine struct {
	context *ExecutionContext
}

// NewEngine creates a new Engine instance with an initialized ExecutionContext.
func NewEngine() *Engine {
	slog.Info("Initializing new execution engine")
	return &Engine{
		context: NewExecutionContext(),
	}
}

// Execute processes a workflow by iterating through its tasks sequentially.
// Each task is looked up in the registry, executed with the current context,
// and its result is stored for subsequent tasks to access.
func (e *Engine) Execute(workflow WorkflowDefinition) error {
	slog.Info("Starting workflow execution", "workflow", workflow.Name)

	for i, task := range workflow.Tasks {
		slog.Info("Processing task", "index", i, "id", task.ID, "type", task.Type)

		// Placeholder: Story 1.3 will add actual task execution
		// For now, just demonstrate sequential processing
		e.context.Set(task.ID+"_result", "placeholder")
	}

	slog.Info("Workflow execution completed", "workflow", workflow.Name)
	return nil
}

// GetContext returns the engine's ExecutionContext for inspection or testing
func (e *Engine) GetContext() *ExecutionContext {
	return e.context
}
