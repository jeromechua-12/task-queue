package models

import "time"

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusError      TaskStatus = "error"
)

type Task struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Status    TaskStatus  `json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
	Options   TaskOptions `json:"options"`
}

type TaskOptions struct {
	Priority      int `json:"priority"` // priority from 1-5, lower = higher priority
	Delay         int `json:"delay"`    // delay before execution in seconds
	MaxRetries    int `json:"maxRetries"`
	TotalAttempts int `json:"totalAttempts"` // attempts made
}
