package tenant

import (
        "crypto/rand"
        "encoding/hex"
        "errors"
        "time"

        tenantModel "github.com/fredphp/yunwei/server/model/tenant"
        "gorm.io/gorm"
)

// TenantService 租户服务
type TenantService struct {
        db *gorm.DB
}

func NewTenantService(db *gorm.DB) *TenantService {
        return &TenantService{db: db}
}

// CreateTenant 创建租户
func (s *TenantService) CreateTenant(name, slug, plan, ownerEmail, ownerName string) (*tenantModel.Tenant, error) {
        // 检查slug是否已存在
        var count int64
        s.db.Model(&tenantModel.Tenant{}).Where("slug = ?", slug).Count(&count)
        if count > 0 {
                return nil, errors.New("租户标识已存在")
        }

        // 开始事务
        tx := s.db.Begin()
        defer func() {
                if r := recover(); r != nil {
                        tx.Rollback()
                }
        }()

        tenantID := generateID()
        tenant := &tenantModel.Tenant{
                ID:           tenantID,
                Name:         name,
                Slug:         slug,
                Status:       "active",
                Plan:         plan,
                BillingCycle: "monthly",
                ContactEmail: ownerEmail,
                ContactName:  ownerName,
                Settings:     tenantModel.JSON{},
                Features:     tenantModel.JSON{},
        }

        if err := tx.Create(tenant).Error; err != nil {
                tx.Rollback()
                return nil, err
        }

        // 创建配额
        quotaPreset, ok := tenantModel.PlanQuotas[plan]
        if !ok {
                quotaPreset = tenantModel.PlanQuotas["free"]
        }

        quota := &tenantModel.TenantQuota{
                ID:               generateID(),
                TenantID:         tenantID,
                MaxUsers:         quotaPreset.MaxUsers,
                MaxAdmins:        quotaPreset.MaxAdmins,
                MaxResources:     quotaPreset.MaxResources,
                MaxServers:       quotaPreset.MaxServers,
                MaxDatabases:     quotaPreset.MaxDatabases,
                MaxMonitors:      quotaPreset.MaxMonitors,
                MaxAlertRules:    quotaPreset.MaxAlertRules,
                MetricsRetention: quotaPreset.MetricsRetention,
                MaxCloudAccounts: quotaPreset.MaxCloudAccounts,
                BudgetLimit:      quotaPreset.BudgetLimit,
                MaxStorageGB:     quotaPreset.MaxStorageGB,
                MaxBackupGB:      quotaPreset.MaxBackupGB,
                MaxAPICalls:      quotaPreset.MaxAPICalls,
                MaxWebhooks:      quotaPreset.MaxWebhooks,
        }

        if err := tx.Create(quota).Error; err != nil {
                tx.Rollback()
                return nil, err
        }

        // 创建默认角色
        if err := s.createDefaultRoles(tx, tenantID); err != nil {
                tx.Rollback()
                return nil, err
        }

        // 获取owner角色
        var ownerRole tenantModel.TenantRole
        if err := tx.Where("tenant_id = ? AND slug = ?", tenantID, "owner").First(&ownerRole).Error; err != nil {
                tx.Rollback()
                return nil, err
        }

        // 创建所有者用户
        tenantUser := &tenantModel.TenantUser{
                ID:       generateID(),
                TenantID: tenantID,
                UserID:   generateID(), // 实际应该关联用户系统
                Email:    ownerEmail,
                Name:     ownerName,
                RoleID:   ownerRole.ID,
                RoleName: "owner",
                IsOwner:  true,
                IsAdmin:  true,
                Status:   "active",
                JoinedAt: time.Now(),
        }

        if err := tx.Create(tenantUser).Error; err != nil {
                tx.Rollback()
                return nil, err
        }

        return tenant, tx.Commit().Error
}

// createDefaultRoles 创建默认角色
func (s *TenantService) createDefaultRoles(tx *gorm.DB, tenantID string) error {
        for slug, roleData := range tenantModel.DefaultRoles {
                role := &tenantModel.TenantRole{
                        ID:          generateID(),
                        TenantID:    tenantID,
                        Name:        slug,
                        Slug:        slug,
                        Description: roleData["description"].(string),
                        IsSystem:    true,
                        Permissions: tenantModel.JSON{"permissions": roleData["permissions"]},
                        Scope:       "tenant",
                }

                if err := tx.Create(role).Error; err != nil {
                        return err
                }
        }
        return nil
}

// GetTenantByID 获取租户
func (s *TenantService) GetTenantByID(id string) (*tenantModel.Tenant, error) {
        var tenant tenantModel.Tenant
        err := s.db.Preload("Quota").First(&tenant, "id = ?", id).Error
        if err != nil {
                return nil, err
        }
        return &tenant, nil
}

// GetTenantBySlug 通过Slug获取租户
func (s *TenantService) GetTenantBySlug(slug string) (*tenantModel.Tenant, error) {
        var tenant tenantModel.Tenant
        err := s.db.Preload("Quota").First(&tenant, "slug = ?", slug).Error
        if err != nil {
                return nil, err
        }
        return &tenant, nil
}

// GetTenantByDomain 通过域名获取租户
func (s *TenantService) GetTenantByDomain(domain string) (*tenantModel.Tenant, error) {
        var tenant tenantModel.Tenant
        err := s.db.Preload("Quota").First(&tenant, "domain = ?", domain).Error
        if err != nil {
                return nil, err
        }
        return &tenant, nil
}

// UpdateTenant 更新租户
func (s *TenantService) UpdateTenant(id string, updates map[string]interface{}) error {
        return s.db.Model(&tenantModel.Tenant{}).Where("id = ?", id).Updates(updates).Error
}

// SuspendTenant 暂停租户
func (s *TenantService) SuspendTenant(id, reason string) error {
        return s.db.Model(&tenantModel.Tenant{}).Where("id = ?", id).Updates(map[string]interface{}{
                "status": "suspended",
        }).Error
}

// ActivateTenant 激活租户
func (s *TenantService) ActivateTenant(id string) error {
        return s.db.Model(&tenantModel.Tenant{}).Where("id = ?", id).Updates(map[string]interface{}{
                "status": "active",
        }).Error
}

// UpgradePlan 升级套餐
func (s *TenantService) UpgradePlan(tenantID, newPlan string) error {
        quotaPreset, ok := tenantModel.PlanQuotas[newPlan]
        if !ok {
                return errors.New("无效的套餐")
        }

        tx := s.db.Begin()
        defer func() {
                if r := recover(); r != nil {
                        tx.Rollback()
                }
        }()

        // 更新租户套餐
        if err := tx.Model(&tenantModel.Tenant{}).Where("id = ?", tenantID).Update("plan", newPlan).Error; err != nil {
                tx.Rollback()
                return err
        }

        // 更新配额
        updates := map[string]interface{}{
                "max_users":          quotaPreset.MaxUsers,
                "max_admins":         quotaPreset.MaxAdmins,
                "max_resources":      quotaPreset.MaxResources,
                "max_servers":        quotaPreset.MaxServers,
                "max_databases":      quotaPreset.MaxDatabases,
                "max_monitors":       quotaPreset.MaxMonitors,
                "max_alert_rules":    quotaPreset.MaxAlertRules,
                "metrics_retention":  quotaPreset.MetricsRetention,
                "max_cloud_accounts": quotaPreset.MaxCloudAccounts,
                "budget_limit":       quotaPreset.BudgetLimit,
                "max_storage_gb":     quotaPreset.MaxStorageGB,
                "max_backup_gb":      quotaPreset.MaxBackupGB,
                "max_api_calls":      quotaPreset.MaxAPICalls,
                "max_webhooks":       quotaPreset.MaxWebhooks,
        }

        if err := tx.Model(&tenantModel.TenantQuota{}).Where("tenant_id = ?", tenantID).Updates(updates).Error; err != nil {
                tx.Rollback()
                return err
        }

        return tx.Commit().Error
}

// ListTenants 列出租户（管理员用）
func (s *TenantService) ListTenants(page, pageSize int, status, plan string) ([]tenantModel.Tenant, int64, error) {
        var tenants []tenantModel.Tenant
        var total int64

        query := s.db.Model(&tenantModel.Tenant{})
        if status != "" {
                query = query.Where("status = ?", status)
        }
        if plan != "" {
                query = query.Where("plan = ?", plan)
        }

        query.Count(&total)
        err := query.Preload("Quota").
                Order("created_at DESC").
                Offset((page - 1) * pageSize).
                Limit(pageSize).
                Find(&tenants).Error

        return tenants, total, err
}

// DeleteTenant 删除租户（软删除）
func (s *TenantService) DeleteTenant(id string) error {
        return s.db.Delete(&tenantModel.Tenant{}, "id = ?", id).Error
}

// generateID 生成唯一ID
func generateID() string {
        bytes := make([]byte, 16)
        rand.Read(bytes)
        return hex.EncodeToString(bytes)
}
