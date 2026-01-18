package engine

import (
	"fmt"
	"sync"
)

// Registry maintains a mapping of task types to their executor implementations.
// It provides thread-safe registration and lookup of task executors.
type Registry struct {
	mu        sync.RWMutex
	executors map[string]TaskExecutor
}

// NewRegistry creates and initializes a new task registry.
func NewRegistry() *Registry {
	return &Registry{
		executors: make(map[string]TaskExecutor),
	}
}

// Register adds a task executor for the given task type.
// If a task type is already registered, it will be overwritten.
// This operation is thread-safe.
func (r *Registry) Register(taskType string, executor TaskExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executors[taskType] = executor
}

// Get retrieves the task executor for the given task type.
// Returns an error if the task type is not registered.
// This operation is thread-safe and uses a read lock for optimal performance.
func (r *Registry) Get(taskType string) (TaskExecutor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	executor, exists := r.executors[taskType]
	if !exists {
		return nil, fmt.Errorf("no executor registered for task type: %s", taskType)
	}
	return executor, nil
}

// List returns all registered task types.
// Useful for debugging and diagnostics.
// This operation is thread-safe and returns a snapshot of registered types.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.executors))
	for taskType := range r.executors {
		types = append(types, taskType)
	}
	return types
}
