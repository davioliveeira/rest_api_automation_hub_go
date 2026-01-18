package repository

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskLogRepository interface defines task log data operations
type TaskLogRepository interface {
	Create(taskLog *TaskLog) error
	Update(taskLog *TaskLog) error
	GetByExecutionID(executionID uuid.UUID) ([]*TaskLog, error)
}

// GormTaskLogRepository implements TaskLogRepository using GORM
type GormTaskLogRepository struct {
	db *gorm.DB
}

// NewTaskLogRepository creates a new task log repository
func NewTaskLogRepository(db *gorm.DB) TaskLogRepository {
	return &GormTaskLogRepository{db: db}
}

// Create inserts a new task log
func (r *GormTaskLogRepository) Create(taskLog *TaskLog) error {
	slog.Info("Creating task log", "execution_id", taskLog.ExecutionID, "task_id", taskLog.TaskID)

	if taskLog.StartedAt.IsZero() {
		taskLog.StartedAt = time.Now().UTC()
	}
	if taskLog.CompletedAt.IsZero() {
		taskLog.CompletedAt = time.Now().UTC()
	}

	if err := r.db.Create(taskLog).Error; err != nil {
		slog.Error("Failed to create task log", "error", err, "execution_id", taskLog.ExecutionID, "task_id", taskLog.TaskID)
		return fmt.Errorf("failed to create task log: %w", err)
	}

	slog.Info("Task log created successfully", "id", taskLog.ID, "task_id", taskLog.TaskID)
	return nil
}

// Update updates an existing task log
func (r *GormTaskLogRepository) Update(taskLog *TaskLog) error {
	slog.Info("Updating task log", "id", taskLog.ID, "task_id", taskLog.TaskID, "status", taskLog.Status)

	if err := r.db.Save(taskLog).Error; err != nil {
		slog.Error("Failed to update task log", "error", err, "id", taskLog.ID)
		return fmt.Errorf("failed to update task log: %w", err)
	}

	slog.Info("Task log updated successfully", "id", taskLog.ID)
	return nil
}

// GetByExecutionID retrieves all task logs for an execution
func (r *GormTaskLogRepository) GetByExecutionID(executionID uuid.UUID) ([]*TaskLog, error) {
	slog.Info("Retrieving task logs by execution ID", "execution_id", executionID)

	var taskLogs []*TaskLog
	if err := r.db.Where("execution_id = ?", executionID).Order("started_at ASC").Find(&taskLogs).Error; err != nil {
		slog.Error("Failed to retrieve task logs", "error", err, "execution_id", executionID)
		return nil, fmt.Errorf("failed to retrieve task logs: %w", err)
	}

	slog.Info("Task logs retrieved successfully", "execution_id", executionID, "count", len(taskLogs))
	return taskLogs, nil
}
