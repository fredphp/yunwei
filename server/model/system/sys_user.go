package system

import (
	"time"

	"gorm.io/gorm"
)

// SysUser 用户表
type SysUser struct {
	ID        uint           `json:"id" gorm:"primarykey;comment:主键ID"`
	CreatedAt time.Time      `json:"createdAt" gorm:"comment:创建时间"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:删除时间"`

	Username  string  `json:"username" gorm:"uniqueIndex;type:varchar(64);not null;comment:用户名"`
	Password  string  `json:"-" gorm:"type:varchar(128);not null;comment:密码"`
	NickName  string  `json:"nickName" gorm:"type:varchar(64);comment:昵称"`
	Avatar    string  `json:"avatar" gorm:"type:varchar(255);comment:头像"`
	Email     string  `json:"email" gorm:"type:varchar(128);comment:邮箱"`
	Phone     string  `json:"phone" gorm:"type:varchar(20);comment:手机号"`
	Status    int     `json:"status" gorm:"type:tinyint(1);default:1;comment:状态: 1启用, 0禁用"`
	RoleID    uint    `json:"roleId" gorm:"comment:角色ID"`
	Role      SysRole `json:"role" gorm:"foreignKey:RoleID"`
}

// TableName 指定表名
func (SysUser) TableName() string {
	return "sys_users"
}
