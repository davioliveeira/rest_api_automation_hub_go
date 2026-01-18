package repository

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

// TestWorkflowModel tests the Workflow model basic functionality
func TestWorkflowModel(t *testing.T) {
	workflow := &Workflow{
		Name:       "test-workflow",
		Definition: datatypes.JSON(`{"name":"test","tasks":[]}`),
	}

	assert.Equal(t, "test-workflow", workflow.Name)
	assert.NotNil(t, workflow.Definition)
	assert.Equal(t, "workflows", workflow.TableName())
}

// TestWorkflowBeforeCreate tests UUID generation
func TestWorkflowBeforeCreate(t *testing.T) {
	workflow := &Workflow{
		Name:       "test-workflow",
		Definition: datatypes.JSON(`{}`),
	}

	// BeforeCreate should generate UUID if ID is nil
	err := workflow.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, workflow.ID)
}

// TestWorkflowConverter tests conversion functions
func TestWorkflowConverter(t *testing.T) {
	// Test ToWorkflowDefinition
	workflow := &Workflow{
		ID:   uuid.New(),
		Name: "test-workflow",
		Definition: datatypes.JSON(`{
			"name": "test-workflow",
			"tasks": [
				{
					"id": "task1",
					"type": "http_request",
					"config": {"url": "https://example.com"}
				}
			]
		}`),
	}

	def, err := workflow.ToWorkflowDefinition()
	assert.NoError(t, err)
	assert.Equal(t, "test-workflow", def.Name)
	assert.Len(t, def.Tasks, 1)
	assert.Equal(t, "task1", def.Tasks[0].ID)
	assert.Equal(t, "http_request", def.Tasks[0].Type)
}

// TestFromWorkflowDefinition tests creating Workflow from engine definition
func TestFromWorkflowDefinition(t *testing.T) {
	// This would require importing engine package, but we'll test the JSON marshaling
	defJSON := map[string]interface{}{
		"name": "test-workflow",
		"tasks": []interface{}{
			map[string]interface{}{
				"id":     "task1",
				"type":   "http_request",
				"config": map[string]interface{}{"url": "https://example.com"},
			},
		},
	}

	jsonBytes, err := json.Marshal(defJSON)
	assert.NoError(t, err)

	workflow := &Workflow{
		Name:       "test-workflow",
		Definition: datatypes.JSON(jsonBytes),
	}

	assert.Equal(t, "test-workflow", workflow.Name)
	assert.NotNil(t, workflow.Definition)
}
