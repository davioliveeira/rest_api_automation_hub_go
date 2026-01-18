package repository

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowRepository interface defines workflow data operations
type WorkflowRepository interface {
	Create(workflow *Workflow) error
	GetByID(id uuid.UUID) (*Workflow, error)
	GetByName(name string) (*Workflow, error)
	GetAll() ([]*Workflow, error)
	Update(workflow *Workflow) error
	Delete(id uuid.UUID) error
}

// GormWorkflowRepository implements WorkflowRepository using GORM
type GormWorkflowRepository struct {
	db *gorm.DB
}

// NewWorkflowRepository creates a new workflow repository
func NewWorkflowRepository(db *gorm.DB) WorkflowRepository {
	return &GormWorkflowRepository{db: db}
}

// Create inserts a new workflow
func (r *GormWorkflowRepository) Create(workflow *Workflow) error {
	slog.Info("Creating workflow", "name", workflow.Name)

	if err := r.db.Create(workflow).Error; err != nil {
		slog.Error("Failed to create workflow", "error", err, "name", workflow.Name)
		// Check for duplicate key error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("workflow with name '%s' already exists", workflow.Name)
		}
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	slog.Info("Workflow created successfully", "id", workflow.ID, "name", workflow.Name)
	return nil
}

// GetByID retrieves a workflow by ID
func (r *GormWorkflowRepository) GetByID(id uuid.UUID) (*Workflow, error) {
	slog.Info("Retrieving workflow by ID", "id", id)

	var workflow Workflow
	if err := r.db.Where("id = ?", id).First(&workflow).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Warn("Workflow not found", "id", id)
			return nil, fmt.Errorf("workflow not found: %s", id)
		}
		slog.Error("Failed to retrieve workflow", "error", err, "id", id)
		return nil, fmt.Errorf("failed to retrieve workflow: %w", err)
	}

	slog.Info("Workflow retrieved successfully", "id", id, "name", workflow.Name)
	return &workflow, nil
}

// GetByName retrieves a workflow by name
func (r *GormWorkflowRepository) GetByName(name string) (*Workflow, error) {
	slog.Info("Retrieving workflow by name", "name", name)

	var workflow Workflow
	if err := r.db.Where("name = ?", name).First(&workflow).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Warn("Workflow not found", "name", name)
			return nil, fmt.Errorf("workflow not found: %s", name)
		}
		slog.Error("Failed to retrieve workflow", "error", err, "name", name)
		return nil, fmt.Errorf("failed to retrieve workflow: %w", err)
	}

	slog.Info("Workflow retrieved successfully", "name", name, "id", workflow.ID)
	return &workflow, nil
}

// GetAll retrieves all workflows
func (r *GormWorkflowRepository) GetAll() ([]*Workflow, error) {
	slog.Info("Retrieving all workflows")

	var workflows []*Workflow
	if err := r.db.Order("created_at DESC").Find(&workflows).Error; err != nil {
		slog.Error("Failed to retrieve workflows", "error", err)
		return nil, fmt.Errorf("failed to retrieve workflows: %w", err)
	}

	slog.Info("Workflows retrieved successfully", "count", len(workflows))
	return workflows, nil
}

// Update updates an existing workflow
func (r *GormWorkflowRepository) Update(workflow *Workflow) error {
	slog.Info("Updating workflow", "id", workflow.ID, "name", workflow.Name)

	if err := r.db.Save(workflow).Error; err != nil {
		slog.Error("Failed to update workflow", "error", err, "id", workflow.ID)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("workflow with name '%s' already exists", workflow.Name)
		}
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	slog.Info("Workflow updated successfully", "id", workflow.ID)
	return nil
}

// Delete deletes a workflow by ID
func (r *GormWorkflowRepository) Delete(id uuid.UUID) error {
	slog.Info("Deleting workflow", "id", id)

	result := r.db.Delete(&Workflow{}, id)
	if result.Error != nil {
		slog.Error("Failed to delete workflow", "error", result.Error, "id", id)
		return fmt.Errorf("failed to delete workflow: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		slog.Warn("Workflow not found for deletion", "id", id)
		return fmt.Errorf("workflow not found: %s", id)
	}

	slog.Info("Workflow deleted successfully", "id", id)
	return nil
}
