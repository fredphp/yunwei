package system

import (
	"time"

	"gorm.io/gorm"
)

// SysApi API表
type SysApi struct {
	ID          uint           `json:"id" gorm:"primarykey;comment:主键ID"`
	CreatedAt   time.Time      `json:"createdAt" gorm:"comment:创建时间"`
	UpdatedAt   time.Time      `json:"updatedAt" gorm:"comment:更新时间"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;comment:删除时间"`

	Path        string `json:"path" gorm:"type:varchar(255);not null;comment:API路径"`
	Method      string `json:"method" gorm:"type:varchar(16);not null;comment:请求方法"`
	Group       string `json:"group" gorm:"type:varchar(64);comment:API分组"`
	Description string `json:"description" gorm:"type:varchar(255);comment:API描述"`
}

// TableName 指定表名
func (SysApi) TableName() string {
	return "sys_apis"
}
