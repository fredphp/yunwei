package tenant

import (
        "errors"
        "strings"
        "time"

        tenantModel "github.com/fredphp/yunwei/server/model/tenant"
        "gorm.io/gorm"
)

// RBACService 角色权限服务
type RBACService struct {
        db *gorm.DB
}

func NewRBACService(db *gorm.DB) *RBACService {
        return &RBACService{db: db}
}

// CreateRole 创建角色
func (s *RBACService) CreateRole(tenantID, name, slug, description string, permissions []string) (*tenantModel.TenantRole, error) {
        // 检查slug是否已存在
        var count int64
        s.db.Model(&tenantModel.TenantRole{}).Where("tenant_id = ? AND slug = ?", tenantID, slug).Count(&count)
        if count > 0 {
                return nil, errors.New("角色标识已存在")
        }

        role := &tenantModel.TenantRole{
                ID:          generateID(),
                TenantID:    tenantID,
                Name:        name,
                Slug:        slug,
                Description: description,
                IsSystem:    false,
                Permissions: tenantModel.JSON{"permissions": permissions},
                Scope:       "tenant",
        }

        if err := s.db.Create(role).Error; err != nil {
                return nil, err
        }

        return role, nil
}

// UpdateRole 更新角色
func (s *RBACService) UpdateRole(roleID, tenantID string, updates map[string]interface{}) error {
        // 检查是否为系统角色
        var role tenantModel.TenantRole
        if err := s.db.Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error; err != nil {
                return err
        }

        if role.IsSystem {
                return errors.New("系统角色不可修改")
        }

        return s.db.Model(&tenantModel.TenantRole{}).
                Where("id = ? AND tenant_id = ?", roleID, tenantID).
                Updates(updates).Error
}

// DeleteRole 删除角色
func (s *RBACService) DeleteRole(roleID, tenantID string) error {
        // 检查是否为系统角色
        var role tenantModel.TenantRole
        if err := s.db.Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error; err != nil {
                return err
        }

        if role.IsSystem {
                return errors.New("系统角色不可删除")
        }

        // 检查是否有用户使用此角色
        var userCount int64
        s.db.Model(&tenantModel.TenantUser{}).Where("role_id = ?", roleID).Count(&userCount)
        if userCount > 0 {
                return errors.New("该角色下存在用户，无法删除")
        }

        return s.db.Delete(&tenantModel.TenantRole{}, "id = ?", roleID).Error
}

// GetRole 获取角色
func (s *RBACService) GetRole(roleID, tenantID string) (*tenantModel.TenantRole, error) {
        var role tenantModel.TenantRole
        err := s.db.Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error
        if err != nil {
                return nil, err
        }
        return &role, nil
}

// ListRoles 列出角色
func (s *RBACService) ListRoles(tenantID string) ([]tenantModel.TenantRole, error) {
        var roles []tenantModel.TenantRole
        err := s.db.Where("tenant_id = ?", tenantID).
                Order("is_system DESC, created_at ASC").
                Find(&roles).Error
        return roles, err
}

// AssignRole 分配角色给用户
func (s *RBACService) AssignRole(tenantID, userID, roleID string) error {
        // 检查角色是否存在
        var role tenantModel.TenantRole
        if err := s.db.Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error; err != nil {
                return errors.New("角色不存在")
        }

        // 检查用户是否存在
        var tenantUser tenantModel.TenantUser
        if err := s.db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&tenantUser).Error; err != nil {
                return errors.New("用户不存在")
        }

        // 更新用户角色
        return s.db.Model(&tenantUser).Updates(map[string]interface{}{
                "role_id":   roleID,
                "role_name": role.Name,
                "is_admin":  role.Slug == "admin" || role.Slug == "owner",
        }).Error
}

// AddUserToTenant 添加用户到租户
func (s *RBACService) AddUserToTenant(tenantID, email, name, roleID, invitedBy string) (*tenantModel.TenantUser, error) {
        // 检查用户是否已存在
        var count int64
        s.db.Model(&tenantModel.TenantUser{}).Where("tenant_id = ? AND email = ?", tenantID, email).Count(&count)
        if count > 0 {
                return nil, errors.New("用户已存在于该租户")
        }

        // 检查配额
        var quota tenantModel.TenantQuota
        if err := s.db.Where("tenant_id = ?", tenantID).First(&quota).Error; err != nil {
                return nil, err
        }

        if quota.MaxUsers != -1 && quota.CurrentUsers >= quota.MaxUsers {
                return nil, errors.New("用户数量已达上限")
        }

        // 获取角色
        var role tenantModel.TenantRole
        if err := s.db.Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error; err != nil {
                return nil, errors.New("角色不存在")
        }

        // 开始事务
        tx := s.db.Begin()
        defer func() {
                if r := recover(); r != nil {
                        tx.Rollback()
                }
        }()

        tenantUser := &tenantModel.TenantUser{
                ID:        generateID(),
                TenantID:  tenantID,
                UserID:    generateID(), // 实际应该关联用户系统
                Email:     email,
                Name:      name,
                RoleID:    roleID,
                RoleName:  role.Name,
                IsAdmin:   role.Slug == "admin" || role.Slug == "owner",
                Status:    "pending",
                InvitedBy: invitedBy,
                JoinedAt:  time.Now(),
        }

        if err := tx.Create(tenantUser).Error; err != nil {
                tx.Rollback()
                return nil, err
        }

        // 更新配额使用量
        if err := tx.Model(&tenantModel.TenantQuota{}).
                Where("tenant_id = ?", tenantID).
                UpdateColumn("current_users", gorm.Expr("current_users + 1")).Error; err != nil {
                tx.Rollback()
                return nil, err
        }

        return tenantUser, tx.Commit().Error
}

// RemoveUserFromTenant 从租户移除用户
func (s *RBACService) RemoveUserFromTenant(tenantID, userID string) error {
        // 获取用户信息
        var tenantUser tenantModel.TenantUser
        if err := s.db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&tenantUser).Error; err != nil {
                return errors.New("用户不存在")
        }

        // 不能移除owner
        if tenantUser.IsOwner {
                return errors.New("不能移除租户所有者")
        }

        // 开始事务
        tx := s.db.Begin()
        defer func() {
                if r := recover(); r != nil {
                        tx.Rollback()
                }
        }()

        if err := tx.Delete(&tenantUser).Error; err != nil {
                tx.Rollback()
                return err
        }

        // 更新配额使用量
        if err := tx.Model(&tenantModel.TenantQuota{}).
                Where("tenant_id = ?", tenantID).
                UpdateColumn("current_users", gorm.Expr("current_users - 1")).Error; err != nil {
                tx.Rollback()
                return err
        }

        return tx.Commit().Error
}

// ListUsers 列出租户用户
func (s *RBACService) ListUsers(tenantID string, page, pageSize int) ([]tenantModel.TenantUser, int64, error) {
        var users []tenantModel.TenantUser
        var total int64

        query := s.db.Model(&tenantModel.TenantUser{}).Where("tenant_id = ?", tenantID)
        query.Count(&total)

        err := query.Order("joined_at DESC").
                Offset((page - 1) * pageSize).
                Limit(pageSize).
                Find(&users).Error

        return users, total, err
}

// UpdateUserStatus 更新用户状态
func (s *RBACService) UpdateUserStatus(tenantID, userID, status string) error {
        return s.db.Model(&tenantModel.TenantUser{}).
                Where("id = ? AND tenant_id = ?", userID, tenantID).
                Update("status", status).Error
}

// CheckPermission 检查用户权限
func (s *RBACService) CheckPermission(tenantID, userID, permission string) (bool, error) {
        // 获取用户信息
        var tenantUser tenantModel.TenantUser
        if err := s.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).
                First(&tenantUser).Error; err != nil {
                return false, err
        }

        // Owner拥有所有权限
        if tenantUser.IsOwner {
                return true, nil
        }

        // 获取角色权限
        var role tenantModel.TenantRole
        if err := s.db.Where("id = ?", tenantUser.RoleID).First(&role).Error; err != nil {
                return false, err
        }

        // 解析权限
        permissions := []string{}
        if perms, ok := role.Permissions["permissions"].([]interface{}); ok {
                for _, p := range perms {
                        permissions = append(permissions, p.(string))
                }
        }

        // 检查权限
        for _, p := range permissions {
                if p == "*" || p == permission {
                        return true, nil
                }
                // 通配符匹配
                if strings.HasSuffix(p, ":*") {
                        prefix := strings.TrimSuffix(p, "*")
                        if strings.HasPrefix(permission, prefix) {
                                return true, nil
                        }
                }
        }

        return false, nil
}

// GetUserPermissions 获取用户所有权限
func (s *RBACService) GetUserPermissions(tenantID, userID string) ([]string, error) {
        // 获取用户信息
        var tenantUser tenantModel.TenantUser
        if err := s.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).
                First(&tenantUser).Error; err != nil {
                return nil, err
        }

        // Owner拥有所有权限
        if tenantUser.IsOwner {
                return []string{"*"}, nil
        }

        // 获取角色权限
        var role tenantModel.TenantRole
        if err := s.db.Where("id = ?", tenantUser.RoleID).First(&role).Error; err != nil {
                return nil, err
        }

        permissions := []string{}
        if perms, ok := role.Permissions["permissions"].([]interface{}); ok {
                for _, p := range perms {
                        permissions = append(permissions, p.(string))
                }
        }

        return permissions, nil
}
