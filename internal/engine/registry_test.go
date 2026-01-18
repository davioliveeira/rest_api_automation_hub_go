package engine

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.executors)
	assert.Equal(t, 0, len(registry.executors))
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	executor := &MockExecutor{}

	registry.Register("test_task", executor)

	retrieved, err := registry.Get("test_task")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Same(t, executor, retrieved)
}

func TestRegistry_RegisterOverwrite(t *testing.T) {
	registry := NewRegistry()

	executor1 := &MockExecutor{Output: "first"}
	executor2 := &MockExecutor{Output: "second"}

	registry.Register("test_task", executor1)
	registry.Register("test_task", executor2) // Overwrite

	retrieved, err := registry.Get("test_task")
	assert.NoError(t, err)
	assert.Same(t, executor2, retrieved)
}

func TestRegistry_GetNotFound(t *testing.T) {
	registry := NewRegistry()

	executor, err := registry.Get("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, executor)
	assert.Contains(t, err.Error(), "no executor registered")
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestRegistry_GetMultipleTypes(t *testing.T) {
	registry := NewRegistry()

	executor1 := &MockExecutor{Output: "http"}
	executor2 := &MockExecutor{Output: "db"}
	executor3 := &MockExecutor{Output: "parser"}

	registry.Register("http_request", executor1)
	registry.Register("database_query", executor2)
	registry.Register("html_parser", executor3)

	// Retrieve each
	http, err := registry.Get("http_request")
	assert.NoError(t, err)
	assert.Same(t, executor1, http)

	db, err := registry.Get("database_query")
	assert.NoError(t, err)
	assert.Same(t, executor2, db)

	parser, err := registry.Get("html_parser")
	assert.NoError(t, err)
	assert.Same(t, executor3, parser)
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	registry.Register("task1", &MockExecutor{})
	registry.Register("task2", &MockExecutor{})
	registry.Register("task3", &MockExecutor{})

	types := registry.List()

	assert.Len(t, types, 3)
	assert.Contains(t, types, "task1")
	assert.Contains(t, types, "task2")
	assert.Contains(t, types, "task3")
}

func TestRegistry_ListEmpty(t *testing.T) {
	registry := NewRegistry()

	types := registry.List()

	assert.NotNil(t, types)
	assert.Len(t, types, 0)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()
	var wg sync.WaitGroup

	// Multiple goroutines registering executors
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			executor := &MockExecutor{}
			registry.Register(fmt.Sprintf("task%d", id), executor)
		}(i)
	}

	// Multiple goroutines reading
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = registry.Get(fmt.Sprintf("task%d", id))
		}(i)
	}

	// Multiple goroutines listing
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = registry.List()
		}()
	}

	wg.Wait()

	// Verify all registrations completed
	types := registry.List()
	assert.Equal(t, 100, len(types))
}

func TestRegistry_ConcurrentRegisterAndGet(t *testing.T) {
	registry := NewRegistry()
	var wg sync.WaitGroup

	executor := &MockExecutor{Output: "shared"}
	registry.Register("shared_task", executor)

	// Many concurrent reads of the same executor
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			retrieved, err := registry.Get("shared_task")
			assert.NoError(t, err)
			assert.NotNil(t, retrieved)
		}()
	}

	// Some concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			registry.Register(fmt.Sprintf("dynamic%d", id), &MockExecutor{})
		}(i)
	}

	wg.Wait()

	// Should have original + 10 dynamic tasks
	types := registry.List()
	assert.GreaterOrEqual(t, len(types), 11)
}
