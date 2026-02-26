package system

import (
	"time"

	"gorm.io/gorm"
)

// SysRole 角色表
type SysRole struct {
	ID          uint           `json:"id" gorm:"primarykey;comment:主键ID"`
	CreatedAt   time.Time      `json:"createdAt" gorm:"comment:创建时间"`
	UpdatedAt   time.Time      `json:"updatedAt" gorm:"comment:更新时间"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;comment:删除时间"`

	Name        string `json:"name" gorm:"type:varchar(64);not null;comment:角色名称"`
	Keyword     string `json:"keyword" gorm:"uniqueIndex;type:varchar(64);not null;comment:角色关键字"`
	Description string `json:"description" gorm:"type:varchar(255);comment:角色描述"`
	Status      int    `json:"status" gorm:"type:tinyint(1);default:1;comment:状态: 1启用, 0禁用"`
}

// TableName 指定表名
func (SysRole) TableName() string {
	return "sys_roles"
}

// SysRoleApi 角色-API关联表
type SysRoleApi struct {
	ID     uint `json:"id" gorm:"primarykey"`
	RoleID uint `json:"roleId" gorm:"index;not null;comment:角色ID"`
	ApiID  uint `json:"apiId" gorm:"index;not null;comment:API ID"`
}

// TableName 指定表名
func (SysRoleApi) TableName() string {
	return "sys_role_apis"
}

// SysRoleMenu 角色-菜单关联表
type SysRoleMenu struct {
	ID     uint `json:"id" gorm:"primarykey"`
	RoleID uint `json:"roleId" gorm:"index;not null;comment:角色ID"`
	MenuID uint `json:"menuId" gorm:"index;not null;comment:菜单ID"`
}

// TableName 指定表名
func (SysRoleMenu) TableName() string {
	return "sys_role_menus"
}
