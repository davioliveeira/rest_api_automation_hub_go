package repository

import (
	"encoding/json"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
)

// ToWorkflowDefinition converts a database Workflow model to engine.WorkflowDefinition
func (w *Workflow) ToWorkflowDefinition() (*engine.WorkflowDefinition, error) {
	var def engine.WorkflowDefinition

	// Unmarshal the JSONB definition
	if err := json.Unmarshal(w.Definition, &def); err != nil {
		return nil, err
	}

	// Ensure the name matches
	def.Name = w.Name

	return &def, nil
}

// FromWorkflowDefinition creates a database Workflow model from engine.WorkflowDefinition
func FromWorkflowDefinition(name string, def *engine.WorkflowDefinition) (*Workflow, error) {
	// Marshal the definition to JSON
	defJSON, err := json.Marshal(def)
	if err != nil {
		return nil, err
	}

	return &Workflow{
		Name:       name,
		Definition: defJSON,
	}, nil
}
