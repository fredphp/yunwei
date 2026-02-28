package security

import (
        "gorm.io/gorm"

        securityModel "yunwei/model/security"
)

// 类型别名 - 引用 model/security 中的类型
type Permission = securityModel.Permission
type Role = securityModel.Role
type User = securityModel.User
type OperationPermission = securityModel.OperationPermission
type PredefinedRole = securityModel.PredefinedRole

// 引用预定义数据
var PredefinedPermissions = securityModel.PredefinedPermissions
var PredefinedRoles = securityModel.PredefinedRoles

// RBACManager RBAC管理器
type RBACManager struct {
        db *gorm.DB
}

// NewRBACManager 创建RBAC管理器
func NewRBACManager(db *gorm.DB) *RBACManager {
        return &RBACManager{db: db}
}

// CheckPermission 检查用户权限
func (m *RBACManager) CheckPermission(userID uint, permissionCode string) bool {
        var count int64
        m.db.Table("users").
                Joins("JOIN user_roles ON users.id = user_roles.user_id").
                Joins("JOIN roles ON user_roles.role_id = roles.id").
                Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
                Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
                Where("users.id = ? AND permissions.code = ? AND users.status = 1", userID, permissionCode).
                Count(&count)

        return count > 0
}

// CheckPermissions 批量检查权限
func (m *RBACManager) CheckPermissions(userID uint, permissionCodes []string) map[string]bool {
        result := make(map[string]bool)
        for _, code := range permissionCodes {
                result[code] = m.CheckPermission(userID, code)
        }
        return result
}

// GetUserPermissions 获取用户所有权限
func (m *RBACManager) GetUserPermissions(userID uint) ([]Permission, error) {
        var permissions []Permission
        err := m.db.Table("permissions").
                Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
                Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
                Where("user_roles.user_id = ?", userID).
                Distinct().
                Find(&permissions).Error
        return permissions, err
}

// GetUserRoles 获取用户角色
func (m *RBACManager) GetUserRoles(userID uint) ([]Role, error) {
        var user User
        err := m.db.Preload("Roles").First(&user, userID).Error
        return user.Roles, err
}

// AssignRole 分配角色
func (m *RBACManager) AssignRole(userID uint, roleID uint) error {
        return m.db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).Error
}

// RevokeRole 撤销角色
func (m *RBACManager) RevokeRole(userID uint, roleID uint) error {
        return m.db.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID).Error
}

// HasRole 检查用户是否有指定角色
func (m *RBACManager) HasRole(userID uint, roleCode string) bool {
        var count int64
        m.db.Table("users").
                Joins("JOIN user_roles ON users.id = user_roles.user_id").
                Joins("JOIN roles ON user_roles.role_id = roles.id").
                Where("users.id = ? AND roles.code = ?", userID, roleCode).
                Count(&count)
        return count > 0
}

// IsAdmin 检查是否为管理员
func (m *RBACManager) IsAdmin(userID uint) bool {
        var user User
        m.db.First(&user, userID)
        return user.IsAdmin || m.HasRole(userID, "admin") || m.HasRole(userID, "super_admin")
}

// CanExecuteCommand 检查是否可以执行命令
func (m *RBACManager) CanExecuteCommand(userID uint, riskLevel int) bool {
        // 超级管理员和管理员可以执行所有命令
        if m.IsAdmin(userID) {
                return true
        }

        // 获取用户最高角色级别
        var maxLevel int
        m.db.Table("users").
                Joins("JOIN user_roles ON users.id = user_roles.user_id").
                Joins("JOIN roles ON user_roles.role_id = roles.id").
                Where("users.id = ?", userID).
                Select("MAX(roles.level)").
                Scan(&maxLevel)

        // 根据角色级别判断
        // Level 2(运维) 可以执行风险等级 <= 3 的命令
        // Level 1(只读) 不能执行命令
        if maxLevel >= 3 {
                return true
        } else if maxLevel >= 2 && riskLevel <= 3 {
                return true
        }
        return false
}

// CanApprove 检查是否可以审批
func (m *RBACManager) CanApprove(userID uint) bool {
        return m.CheckPermission(userID, "command:approve") || m.IsAdmin(userID)
}

// GetRoleByID 获取角色
func (m *RBACManager) GetRoleByID(roleID uint) (*Role, error) {
        var role Role
        err := m.db.Preload("Permissions").First(&role, roleID).Error
        return &role, err
}

// CreateRole 创建角色
func (m *RBACManager) CreateRole(role *Role) error {
        return m.db.Create(role).Error
}

// UpdateRole 更新角色
func (m *RBACManager) UpdateRole(role *Role) error {
        return m.db.Save(role).Error
}

// DeleteRole 删除角色
func (m *RBACManager) DeleteRole(roleID uint) error {
        var role Role
        m.db.First(&role, roleID)
        if role.IsSystem {
                return gorm.ErrRecordNotFound // 系统角色不能删除
        }
        return m.db.Delete(&role, roleID).Error
}
