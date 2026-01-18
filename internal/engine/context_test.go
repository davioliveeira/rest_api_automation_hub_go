package engine

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExecutionContext(t *testing.T) {
	ctx := NewExecutionContext()
	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.data)
	assert.Equal(t, 0, len(ctx.data))
}

func TestExecutionContext_SetAndGet(t *testing.T) {
	ctx := NewExecutionContext()

	// Test setting and getting a value
	ctx.Set("key1", "value1")
	val, exists := ctx.Get("key1")

	assert.True(t, exists)
	assert.Equal(t, "value1", val)
}

func TestExecutionContext_GetNonExistent(t *testing.T) {
	ctx := NewExecutionContext()

	val, exists := ctx.Get("nonexistent")

	assert.False(t, exists)
	assert.Nil(t, val)
}

func TestExecutionContext_SetMultipleTypes(t *testing.T) {
	ctx := NewExecutionContext()

	ctx.Set("string", "value")
	ctx.Set("int", 42)
	ctx.Set("bool", true)
	ctx.Set("map", map[string]interface{}{"nested": "data"})

	strVal, _ := ctx.Get("string")
	intVal, _ := ctx.Get("int")
	boolVal, _ := ctx.Get("bool")
	mapVal, _ := ctx.Get("map")

	assert.Equal(t, "value", strVal)
	assert.Equal(t, 42, intVal)
	assert.Equal(t, true, boolVal)
	assert.Equal(t, map[string]interface{}{"nested": "data"}, mapVal)
}

func TestExecutionContext_GetAll(t *testing.T) {
	ctx := NewExecutionContext()

	ctx.Set("key1", "value1")
	ctx.Set("key2", 123)
	ctx.Set("key3", true)

	snapshot := ctx.GetAll()

	assert.Equal(t, 3, len(snapshot))
	assert.Equal(t, "value1", snapshot["key1"])
	assert.Equal(t, 123, snapshot["key2"])
	assert.Equal(t, true, snapshot["key3"])
}

func TestExecutionContext_GetAllIsSnapshot(t *testing.T) {
	ctx := NewExecutionContext()
	ctx.Set("original", "value")

	snapshot := ctx.GetAll()

	// Modify the snapshot
	snapshot["modified"] = "newValue"

	// Original context should not be affected
	_, exists := ctx.Get("modified")
	assert.False(t, exists)
}

func TestExecutionContext_Clear(t *testing.T) {
	ctx := NewExecutionContext()

	ctx.Set("key1", "value1")
	ctx.Set("key2", "value2")
	ctx.Set("key3", "value3")

	assert.Equal(t, 3, len(ctx.GetAll()))

	ctx.Clear()

	assert.Equal(t, 0, len(ctx.GetAll()))
	_, exists := ctx.Get("key1")
	assert.False(t, exists)
}

func TestExecutionContext_ConcurrentAccess(t *testing.T) {
	ctx := NewExecutionContext()
	var wg sync.WaitGroup

	// Launch 100 goroutines writing to context
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx.Set(fmt.Sprintf("key%d", id), id)
		}(i)
	}

	// Launch 100 goroutines reading from context
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = ctx.Get(fmt.Sprintf("key%d", id))
		}(i)
	}

	// Launch goroutines calling GetAll
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ctx.GetAll()
		}()
	}

	wg.Wait()

	// Verify all writes completed successfully
	snapshot := ctx.GetAll()
	assert.Equal(t, 100, len(snapshot))
}

func TestExecutionContext_ConcurrentSetAndClear(t *testing.T) {
	ctx := NewExecutionContext()
	var wg sync.WaitGroup

	// Continuously write data
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ctx.Set(fmt.Sprintf("key%d_%d", id, j), id*10+j)
			}
		}(i)
	}

	// Occasionally clear
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx.Clear()
		}()
	}

	wg.Wait()

	// Should complete without race conditions
	// Final state may vary due to concurrent clears
	snapshot := ctx.GetAll()
	assert.NotNil(t, snapshot)
}

func TestExecutionContext_OverwriteValue(t *testing.T) {
	ctx := NewExecutionContext()

	ctx.Set("key", "value1")
	val1, _ := ctx.Get("key")
	assert.Equal(t, "value1", val1)

	ctx.Set("key", "value2")
	val2, _ := ctx.Get("key")
	assert.Equal(t, "value2", val2)
}
