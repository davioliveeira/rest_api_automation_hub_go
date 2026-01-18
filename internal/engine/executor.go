package engine

// TaskExecutor defines the contract for all task implementations.
// Each task type (http_request, html_parser, transform, etc.) must implement this interface.
// The Execute method receives the current ExecutionContext and task-specific configuration,
// and returns a TaskResult indicating success or failure.
type TaskExecutor interface {
	// Execute runs the task with the given context and configuration.
	// The ExecutionContext allows reading data from previous tasks and storing output.
	// The config map contains task-specific configuration from the workflow JSON.
	// Returns a TaskResult indicating success/failure, output data, and any error message.
	Execute(ctx *ExecutionContext, config map[string]interface{}) TaskResult
}
