package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Workflow represents a workflow definition in the database
type Workflow struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name       string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Definition datatypes.JSON `gorm:"type:jsonb;not null" json:"definition"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate GORM hook to generate UUID
func (w *Workflow) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Workflow) TableName() string {
	return "workflows"
}

// Execution represents a workflow execution in the database
type Execution struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	WorkflowID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"workflow_id"`
	Workflow        Workflow       `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
	Status          string         `gorm:"type:varchar(50);not null;default:'pending'" json:"status"` // pending, running, completed, failed
	ContextSnapshot datatypes.JSON `gorm:"type:jsonb" json:"context_snapshot,omitempty"`
	StartedAt       time.Time      `gorm:"not null" json:"started_at"`
	CompletedAt     *time.Time     `gorm:"default:null" json:"completed_at,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	TaskLogs        []TaskLog      `gorm:"foreignKey:ExecutionID" json:"task_logs,omitempty"`
}

// BeforeCreate GORM hook to generate UUID
func (e *Execution) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Execution) TableName() string {
	return "executions"
}

// TaskLog represents a task execution log in the database
type TaskLog struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	ExecutionID uuid.UUID      `gorm:"type:uuid;not null;index" json:"execution_id"`
	Execution   Execution      `gorm:"foreignKey:ExecutionID" json:"execution,omitempty"`
	TaskID      string         `gorm:"type:varchar(255);not null" json:"task_id"`
	TaskType    string         `gorm:"type:varchar(100)" json:"task_type"`
	Status      string         `gorm:"type:varchar(50);not null" json:"status"` // success, failed
	Input       datatypes.JSON `gorm:"type:jsonb" json:"input,omitempty"`
	Output      datatypes.JSON `gorm:"type:jsonb" json:"output,omitempty"`
	Error       string         `gorm:"type:text" json:"error,omitempty"`
	StartedAt   time.Time      `gorm:"not null" json:"started_at"`
	CompletedAt time.Time      `gorm:"not null" json:"completed_at"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

// BeforeCreate GORM hook to generate UUID
func (t *TaskLog) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (TaskLog) TableName() string {
	return "task_logs"
}
