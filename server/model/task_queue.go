package model

import "time"

// TaskQueue 任务队列模型
type TaskQueue struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name       string `json:"name" gorm:"type:varchar(64);uniqueIndex"`
	MaxWorkers int    `json:"maxWorkers"`
	MaxPending int    `json:"maxPending"`
	Priority   int    `json:"priority"`
	Timeout    int    `json:"timeout"`
	MaxRetry   int    `json:"maxRetry"`
	Enabled    bool   `json:"enabled" gorm:"default:true"`
}

func (TaskQueue) TableName() string {
	return "scheduler_task_queues"
}
