package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine(nil)

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.context)
	assert.NotNil(t, engine.registry)
	assert.Equal(t, 0, len(engine.context.GetAll()))
}

func TestNewEngine_WithRegistry(t *testing.T) {
	registry := NewRegistry()
	registry.Register("test", &MockExecutor{})

	engine := NewEngine(registry)

	assert.NotNil(t, engine)
	assert.Same(t, registry, engine.registry)

	// Verify registry is accessible
	_, err := engine.registry.Get("test")
	assert.NoError(t, err)
}

func TestEngine_Execute_EmptyWorkflow(t *testing.T) {
	engine := NewEngine(nil)

	workflow := WorkflowDefinition{
		Name:  "empty-workflow",
		Tasks: []Task{},
	}

	err := engine.Execute(workflow)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(engine.context.GetAll()))
}

func TestEngine_GetContext(t *testing.T) {
	engine := NewEngine(nil)

	ctx := engine.GetContext()

	assert.NotNil(t, ctx)
	assert.Same(t, engine.context, ctx)
}

func TestEngine_MultipleExecutions(t *testing.T) {
	// Test that multiple engine instances don't interfere
	registry := NewRegistry()
	registry.Register("test", &MockExecutor{Output: "result"})

	engine1 := NewEngine(registry)
	engine2 := NewEngine(registry)

	workflow1 := WorkflowDefinition{
		Name:  "workflow1",
		Tasks: []Task{{ID: "task1", Type: "test", Config: map[string]interface{}{}}},
	}

	workflow2 := WorkflowDefinition{
		Name:  "workflow2",
		Tasks: []Task{{ID: "task2", Type: "test", Config: map[string]interface{}{}}},
	}

	err1 := engine1.Execute(workflow1)
	err2 := engine2.Execute(workflow2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Verify isolation
	_, exists := engine1.context.Get("task1_result")
	assert.True(t, exists)
	_, exists = engine1.context.Get("task2_result")
	assert.False(t, exists)

	_, exists = engine2.context.Get("task2_result")
	assert.True(t, exists)
	_, exists = engine2.context.Get("task1_result")
	assert.False(t, exists)
}

// New tests for registry integration

func TestEngine_WithMockExecutor(t *testing.T) {
	registry := NewRegistry()
	mockExecutor := &MockExecutor{
		ShouldFail: false,
		Output:     map[string]interface{}{"result": "success"},
	}
	registry.Register("mock_task", mockExecutor)

	engine := NewEngine(registry)

	workflow := WorkflowDefinition{
		Name: "test-workflow",
		Tasks: []Task{
			{ID: "task1", Type: "mock_task", Config: map[string]interface{}{}},
		},
	}

	err := engine.Execute(workflow)
	assert.NoError(t, err)

	// Verify result was stored in context
	result, exists := engine.context.Get("task1_result")
	assert.True(t, exists)
	assert.Equal(t, map[string]interface{}{"result": "success"}, result)
}

func TestEngine_UnregisteredTaskType(t *testing.T) {
	registry := NewRegistry()
	engine := NewEngine(registry)

	workflow := WorkflowDefinition{
		Name: "test-workflow",
		Tasks: []Task{
			{ID: "task1", Type: "unknown_type", Config: map[string]interface{}{}},
		},
	}

	err := engine.Execute(workflow)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task executor not found")
	assert.Contains(t, err.Error(), "unknown_type")
}

func TestEngine_TaskFailure(t *testing.T) {
	registry := NewRegistry()
	mockExecutor := &MockExecutor{
		ShouldFail: true,
		ErrorMsg:   "intentional failure",
	}
	registry.Register("failing_task", mockExecutor)

	engine := NewEngine(registry)

	workflow := WorkflowDefinition{
		Name: "test-workflow",
		Tasks: []Task{
			{ID: "task1", Type: "failing_task", Config: map[string]interface{}{}},
		},
	}

	err := engine.Execute(workflow)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task1 failed")
	assert.Contains(t, err.Error(), "intentional failure")
}

func TestEngine_TaskFailureStopsExecution(t *testing.T) {
	registry := NewRegistry()

	failingExecutor := &MockExecutor{
		ShouldFail: true,
		ErrorMsg:   "first task failed",
	}
	successExecutor := &MockExecutor{
		ShouldFail: false,
		Output:     "should not execute",
	}

	registry.Register("failing", failingExecutor)
	registry.Register("success", successExecutor)

	engine := NewEngine(registry)

	workflow := WorkflowDefinition{
		Name: "test-workflow",
		Tasks: []Task{
			{ID: "task1", Type: "failing", Config: map[string]interface{}{}},
			{ID: "task2", Type: "success", Config: map[string]interface{}{}},
		},
	}

	err := engine.Execute(workflow)
	assert.Error(t, err)

	// Task1 should have no result (failed)
	_, exists := engine.context.Get("task1_result")
	assert.False(t, exists)

	// Task2 should not have executed
	_, exists = engine.context.Get("task2_result")
	assert.False(t, exists)
}

func TestEngine_MultipleTasksContextSharing(t *testing.T) {
	registry := NewRegistry()

	// First executor sets data in context
	executor1 := &MockExecutor{Output: map[string]interface{}{"step1": "data"}}
	registry.Register("task_type_1", executor1)

	// Second executor can read from context
	executor2 := &MockExecutor{Output: map[string]interface{}{"step2": "more data"}}
	registry.Register("task_type_2", executor2)

	engine := NewEngine(registry)

	workflow := WorkflowDefinition{
		Name: "multi-task-workflow",
		Tasks: []Task{
			{ID: "task1", Type: "task_type_1", Config: map[string]interface{}{}},
			{ID: "task2", Type: "task_type_2", Config: map[string]interface{}{}},
		},
	}

	err := engine.Execute(workflow)
	assert.NoError(t, err)

	// Verify both results are in context
	result1, exists1 := engine.context.Get("task1_result")
	assert.True(t, exists1)
	assert.Equal(t, map[string]interface{}{"step1": "data"}, result1)

	result2, exists2 := engine.context.Get("task2_result")
	assert.True(t, exists2)
	assert.Equal(t, map[string]interface{}{"step2": "more data"}, result2)
}

func TestEngine_TaskWithConfig(t *testing.T) {
	registry := NewRegistry()

	executor := &MockExecutor{
		Output: map[string]interface{}{"processed": true},
	}
	registry.Register("configurable_task", executor)

	engine := NewEngine(registry)

	workflow := WorkflowDefinition{
		Name: "config-workflow",
		Tasks: []Task{
			{
				ID:   "task1",
				Type: "configurable_task",
				Config: map[string]interface{}{
					"param1": "value1",
					"param2": 123,
					"nested": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
	}

	err := engine.Execute(workflow)
	assert.NoError(t, err)

	result, exists := engine.context.Get("task1_result")
	assert.True(t, exists)
	assert.Equal(t, map[string]interface{}{"processed": true}, result)
}
