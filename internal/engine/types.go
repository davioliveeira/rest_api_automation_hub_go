package engine

// WorkflowDefinition represents a complete workflow with its tasks
type WorkflowDefinition struct {
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"`
}

// Task represents a single executable task within a workflow
type Task struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// TaskResult represents the outcome of a task execution
type TaskResult struct {
	Status string      `json:"status"` // "success" or "failed"
	Output interface{} `json:"output"`
	Error  string      `json:"error,omitempty"`
}
