package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.context)
	assert.Equal(t, 0, len(engine.context.GetAll()))
}

func TestEngine_ExecuteEmptyWorkflow(t *testing.T) {
	engine := NewEngine()

	workflow := WorkflowDefinition{
		Name:  "empty-workflow",
		Tasks: []Task{},
	}

	err := engine.Execute(workflow)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(engine.context.GetAll()))
}

func TestEngine_ExecuteSequentialTasks(t *testing.T) {
	engine := NewEngine()

	workflow := WorkflowDefinition{
		Name: "simple-workflow",
		Tasks: []Task{
			{ID: "task1", Type: "mock", Config: map[string]interface{}{}},
			{ID: "task2", Type: "mock", Config: map[string]interface{}{}},
		},
	}

	err := engine.Execute(workflow)
	assert.NoError(t, err)

	result1, exists1 := engine.context.Get("task1_result")
	result2, exists2 := engine.context.Get("task2_result")

	assert.True(t, exists1)
	assert.True(t, exists2)
	assert.Equal(t, "placeholder", result1)
	assert.Equal(t, "placeholder", result2)
}

func TestEngine_GetContext(t *testing.T) {
	engine := NewEngine()

	ctx := engine.GetContext()

	assert.NotNil(t, ctx)
	assert.Same(t, engine.context, ctx)
}
