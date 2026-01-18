package repository

import (
	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
)

// ExecutionLoggerAdapter adapts ExecutionRepository and TaskLogRepository to engine.ExecutionLogger
type ExecutionLoggerAdapter struct {
	execRepo    ExecutionRepository
	taskLogRepo TaskLogRepository
}

// NewExecutionLoggerAdapter creates a new adapter
func NewExecutionLoggerAdapter(execRepo ExecutionRepository, taskLogRepo TaskLogRepository) engine.ExecutionLogger {
	return &ExecutionLoggerAdapter{
		execRepo:    execRepo,
		taskLogRepo: taskLogRepo,
	}
}

// CreateExecution creates an execution record
func (a *ExecutionLoggerAdapter) CreateExecution(exec *engine.ExecutionRecord) error {
	execution := &Execution{
		ID:              exec.ID,
		WorkflowID:      exec.WorkflowID,
		Status:          exec.Status,
		ContextSnapshot: exec.ContextSnapshot,
		StartedAt:       exec.StartedAt,
		CompletedAt:     exec.CompletedAt,
	}
	return a.execRepo.Create(execution)
}

// UpdateExecution updates an execution record
func (a *ExecutionLoggerAdapter) UpdateExecution(exec *engine.ExecutionRecord) error {
	execution := &Execution{
		ID:              exec.ID,
		WorkflowID:      exec.WorkflowID,
		Status:          exec.Status,
		ContextSnapshot: exec.ContextSnapshot,
		StartedAt:      exec.StartedAt,
		CompletedAt:    exec.CompletedAt,
	}
	return a.execRepo.Update(execution)
}

// CreateTaskLog creates a task log record
func (a *ExecutionLoggerAdapter) CreateTaskLog(taskLog *engine.TaskLogRecord) error {
	tl := &TaskLog{
		ID:          taskLog.ID,
		ExecutionID: taskLog.ExecutionID,
		TaskID:      taskLog.TaskID,
		TaskType:    taskLog.TaskType,
		Status:      taskLog.Status,
		Input:       taskLog.Input,
		Output:      taskLog.Output,
		Error:       taskLog.Error,
		StartedAt:   taskLog.StartedAt,
		CompletedAt: taskLog.CompletedAt,
	}
	return a.taskLogRepo.Create(tl)
}

// UpdateTaskLog updates a task log record
func (a *ExecutionLoggerAdapter) UpdateTaskLog(taskLog *engine.TaskLogRecord) error {
	tl := &TaskLog{
		ID:          taskLog.ID,
		ExecutionID: taskLog.ExecutionID,
		TaskID:      taskLog.TaskID,
		TaskType:    taskLog.TaskType,
		Status:      taskLog.Status,
		Input:       taskLog.Input,
		Output:      taskLog.Output,
		Error:       taskLog.Error,
		StartedAt:   taskLog.StartedAt,
		CompletedAt: taskLog.CompletedAt,
	}
	return a.taskLogRepo.Update(tl)
}
