package engine

// MockExecutor is a test implementation of TaskExecutor.
// It can be configured to return success or failure for testing purposes.
// This executor is intended for testing only and should not be used in production.
type MockExecutor struct {
	// ShouldFail determines if the executor should simulate a failure
	ShouldFail bool
	// Output is the data returned on successful execution
	Output interface{}
	// ErrorMsg is the error message returned on failure
	ErrorMsg string
}

// Execute implements the TaskExecutor interface for testing.
// Returns a successful TaskResult with the configured Output, or a failed
// TaskResult with the configured ErrorMsg if ShouldFail is true.
func (m *MockExecutor) Execute(ctx *ExecutionContext, config map[string]interface{}) TaskResult {
	if m.ShouldFail {
		return TaskResult{
			Status: "failed",
			Output: nil,
			Error:  m.ErrorMsg,
		}
	}

	// Simulate successful task execution
	output := m.Output
	if output == nil {
		output = map[string]interface{}{"mock": "result"}
	}

	return TaskResult{
		Status: "success",
		Output: output,
		Error:  "",
	}
}
