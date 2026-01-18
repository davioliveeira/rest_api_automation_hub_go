package engine

import "sync"

// ExecutionContext provides thread-safe key-value storage for sharing data between tasks
// within a single workflow execution. It uses RWMutex for optimal read performance.
type ExecutionContext struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewExecutionContext creates a new ExecutionContext with an initialized data map
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in the context using the provided key.
// This operation is thread-safe and uses a write lock.
func (ctx *ExecutionContext) Set(key string, value interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.data[key] = value
}

// Get retrieves a value from the context by key.
// Returns the value and true if found, nil and false otherwise.
// This operation is thread-safe and uses a read lock.
func (ctx *ExecutionContext) Get(key string) (interface{}, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	val, exists := ctx.data[key]
	return val, exists
}

// GetAll returns a snapshot of all key-value pairs in the context.
// The returned map is a copy to prevent external modifications.
// This operation is thread-safe and uses a read lock.
func (ctx *ExecutionContext) GetAll() map[string]interface{} {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	snapshot := make(map[string]interface{}, len(ctx.data))
	for k, v := range ctx.data {
		snapshot[k] = v
	}
	return snapshot
}

// Clear removes all key-value pairs from the context.
// This operation is thread-safe and uses a write lock.
func (ctx *ExecutionContext) Clear() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.data = make(map[string]interface{})
}
