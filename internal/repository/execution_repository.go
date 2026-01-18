package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExecutionRepository interface defines execution data operations
type ExecutionRepository interface {
	Create(execution *Execution) error
	GetByID(id uuid.UUID) (*Execution, error)
	GetByWorkflowID(workflowID uuid.UUID) ([]*Execution, error)
	GetAll() ([]*Execution, error)
	Update(execution *Execution) error
}

// GormExecutionRepository implements ExecutionRepository using GORM
type GormExecutionRepository struct {
	db *gorm.DB
}

// NewExecutionRepository creates a new execution repository
func NewExecutionRepository(db *gorm.DB) ExecutionRepository {
	return &GormExecutionRepository{db: db}
}

// Create inserts a new execution
func (r *GormExecutionRepository) Create(execution *Execution) error {
	slog.Info("Creating execution", "workflow_id", execution.WorkflowID)

	if execution.Status == "" {
		execution.Status = "pending"
	}
	if execution.StartedAt.IsZero() {
		execution.StartedAt = time.Now().UTC()
	}

	if err := r.db.Create(execution).Error; err != nil {
		slog.Error("Failed to create execution", "error", err, "workflow_id", execution.WorkflowID)
		return fmt.Errorf("failed to create execution: %w", err)
	}

	slog.Info("Execution created successfully", "id", execution.ID, "workflow_id", execution.WorkflowID)
	return nil
}

// GetByID retrieves an execution by ID with task logs
func (r *GormExecutionRepository) GetByID(id uuid.UUID) (*Execution, error) {
	slog.Info("Retrieving execution by ID", "id", id)

	var execution Execution
	if err := r.db.Preload("TaskLogs").Preload("Workflow").Where("id = ?", id).First(&execution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Warn("Execution not found", "id", id)
			return nil, fmt.Errorf("execution not found: %s", id)
		}
		slog.Error("Failed to retrieve execution", "error", err, "id", id)
		return nil, fmt.Errorf("failed to retrieve execution: %w", err)
	}

	slog.Info("Execution retrieved successfully", "id", id, "status", execution.Status)
	return &execution, nil
}

// GetByWorkflowID retrieves all executions for a workflow
func (r *GormExecutionRepository) GetByWorkflowID(workflowID uuid.UUID) ([]*Execution, error) {
	slog.Info("Retrieving executions by workflow ID", "workflow_id", workflowID)

	var executions []*Execution
	if err := r.db.Where("workflow_id = ?", workflowID).Order("started_at DESC").Find(&executions).Error; err != nil {
		slog.Error("Failed to retrieve executions", "error", err, "workflow_id", workflowID)
		return nil, fmt.Errorf("failed to retrieve executions: %w", err)
	}

	slog.Info("Executions retrieved successfully", "workflow_id", workflowID, "count", len(executions))
	return executions, nil
}

// GetAll retrieves all executions
func (r *GormExecutionRepository) GetAll() ([]*Execution, error) {
	slog.Info("Retrieving all executions")

	var executions []*Execution
	if err := r.db.Order("started_at DESC").Find(&executions).Error; err != nil {
		slog.Error("Failed to retrieve executions", "error", err)
		return nil, fmt.Errorf("failed to retrieve executions: %w", err)
	}

	slog.Info("Executions retrieved successfully", "count", len(executions))
	return executions, nil
}

// Update updates an existing execution
func (r *GormExecutionRepository) Update(execution *Execution) error {
	slog.Info("Updating execution", "id", execution.ID, "status", execution.Status)

	if err := r.db.Save(execution).Error; err != nil {
		slog.Error("Failed to update execution", "error", err, "id", execution.ID)
		return fmt.Errorf("failed to update execution: %w", err)
	}

	slog.Info("Execution updated successfully", "id", execution.ID, "status", execution.Status)
	return nil
}
