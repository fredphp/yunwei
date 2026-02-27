package system

import (
	"time"

	"gorm.io/gorm"
)

// SysMenu 菜单表
type SysMenu struct {
	ID        uint           `json:"id" gorm:"primarykey;comment:主键ID"`
	CreatedAt time.Time      `json:"createdAt" gorm:"comment:创建时间"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:删除时间"`

	ParentID  uint   `json:"parentId" gorm:"default:0;comment:父菜单ID"`
	Title     string `json:"title" gorm:"type:varchar(64);not null;comment:菜单标题"`
	Name      string `json:"name" gorm:"type:varchar(64);not null;comment:路由名称"`
	Path      string `json:"path" gorm:"type:varchar(255);comment:路由路径"`
	Component string `json:"component" gorm:"type:varchar(255);comment:组件路径"`
	Icon      string `json:"icon" gorm:"type:varchar(64);comment:菜单图标"`
	Sort      int    `json:"sort" gorm:"default:0;comment:排序"`
	Status    int    `json:"status" gorm:"type:tinyint(1);default:1;comment:状态: 1启用, 0禁用"`
	Hidden    int    `json:"hidden" gorm:"type:tinyint(1);default:0;comment:是否隐藏: 1隐藏, 0显示"`
	
	Children  []SysMenu `json:"children" gorm:"-"`
}

// TableName 指定表名
func (SysMenu) TableName() string {
	return "sys_menus"
}
