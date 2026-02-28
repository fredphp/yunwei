package system

import (
	"time"
)

// SysMigration 数据迁移记录表
type SysMigration struct {
	ID          uint      `json:"id" gorm:"primarykey;comment:主键ID"`
	CreatedAt   time.Time `json:"createdAt" gorm:"comment:执行时间"`
	Name        string    `json:"name" gorm:"type:varchar(255);uniqueIndex;not null;comment:迁移文件名"`
	Checksum    string    `json:"checksum" gorm:"type:varchar(64);comment:文件校验和"`
	ExecutionMs int64     `json:"executionMs" gorm:"comment:执行耗时(毫秒)"`
	Status      string    `json:"status" gorm:"type:varchar(16);default:'success';comment:状态: success/failed"`
	ErrorMsg    string    `json:"errorMsg" gorm:"type:text;comment:错误信息"`
}

// TableName 指定表名
func (SysMigration) TableName() string {
	return "sys_migrations"
}
