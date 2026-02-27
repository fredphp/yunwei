package tenant

import (
        "errors"
        "fmt"
        "strings"

        tenantModel "yunwei/model/tenant"
        "github.com/gin-gonic/gin"
        "gorm.io/gorm"
)

// TenantContext 租户上下文
type TenantContext struct {
        TenantID   string
        TenantSlug string
        Tenant     *tenantModel.Tenant
        UserID     string
        UserRole   string
        IsOwner    bool
        IsAdmin    bool
        Permissions []string
}

// contextKey 上下文键
type contextKey string

const (
        TenantContextKey contextKey = "tenant_context"
)

// IsolationService 数据隔离服务
type IsolationService struct {
        db *gorm.DB
}

func NewIsolationService(db *gorm.DB) *IsolationService {
        return &IsolationService{db: db}
}

// GetTenantContext 从上下文获取租户信息
func GetTenantContext(c *gin.Context) *TenantContext {
        if tc, exists := c.Get(string(TenantContextKey)); exists {
                return tc.(*TenantContext)
        }
        return nil
}

// SetTenantContext 设置租户上下文
func SetTenantContext(c *gin.Context, tc *TenantContext) {
        c.Set(string(TenantContextKey), tc)
}

// TenantMiddleware 租户识别中间件
func (s *IsolationService) TenantMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                var tenant *tenantModel.Tenant

                // 1. 优先从请求头获取租户ID
                tenantID := c.GetHeader("X-Tenant-ID")
                if tenantID != "" {
                        tenant, _ = s.getTenantByID(tenantID)
                }

                // 2. 从子域名获取租户
                if tenant == nil {
                        host := c.Request.Host
                        parts := strings.Split(host, ".")
                        if len(parts) >= 1 {
                                subdomain := parts[0]
                                if subdomain != "www" && subdomain != "api" && subdomain != "app" {
                                        tenant, _ = s.getTenantBySlug(subdomain)
                                }
                        }
                }

                // 3. 从自定义域名获取租户
                if tenant == nil {
                        host := c.Request.Host
                        tenant, _ = s.getTenantByDomain(host)
                }

                // 4. 从查询参数获取（开发调试用）
                if tenant == nil {
                        tenantSlug := c.Query("tenant")
                        if tenantSlug != "" {
                                tenant, _ = s.getTenantBySlug(tenantSlug)
                        }
                }

                if tenant == nil {
                        c.JSON(401, gin.H{"error": "无法识别租户"})
                        c.Abort()
                        return
                }

                // 检查租户状态
                if tenant.Status == "suspended" {
                        c.JSON(403, gin.H{"error": "租户已被暂停"})
                        c.Abort()
                        return
                }

                if tenant.Status == "deleted" {
                        c.JSON(403, gin.H{"error": "租户已被删除"})
                        c.Abort()
                        return
                }

                // 设置租户上下文
                tc := &TenantContext{
                        TenantID:   tenant.ID,
                        TenantSlug: tenant.Slug,
                        Tenant:     tenant,
                }

                SetTenantContext(c, tc)
                c.Next()
        }
}

// AuthMiddleware 认证中间件（需要配合租户中间件使用）
func (s *IsolationService) AuthMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                tc := GetTenantContext(c)
                if tc == nil {
                        c.JSON(401, gin.H{"error": "租户上下文不存在"})
                        c.Abort()
                        return
                }

                // 从请求头获取用户信息（实际应该验证JWT Token）
                userID := c.GetHeader("X-User-ID")
                if userID == "" {
                        c.JSON(401, gin.H{"error": "未登录"})
                        c.Abort()
                        return
                }

                // 获取租户用户信息
                var tenantUser tenantModel.TenantUser
                if err := s.db.Where("tenant_id = ? AND user_id = ? AND status = ?",
                        tc.TenantID, userID, "active").First(&tenantUser).Error; err != nil {
                        c.JSON(403, gin.H{"error": "用户不属于该租户"})
                        c.Abort()
                        return
                }

                // 获取用户权限
                var role tenantModel.TenantRole
                permissions := []string{}
                if err := s.db.Where("id = ?", tenantUser.RoleID).First(&role).Error; err == nil {
                        if perms, ok := role.Permissions["permissions"].([]interface{}); ok {
                                for _, p := range perms {
                                        permissions = append(permissions, p.(string))
                                }
                        }
                }

                // 更新上下文
                tc.UserID = userID
                tc.UserRole = tenantUser.RoleName
                tc.IsOwner = tenantUser.IsOwner
                tc.IsAdmin = tenantUser.IsAdmin
                tc.Permissions = permissions

                SetTenantContext(c, tc)
                c.Next()
        }
}

// RequirePermission 权限检查中间件
func (s *IsolationService) RequirePermission(permission string) gin.HandlerFunc {
        return func(c *gin.Context) {
                tc := GetTenantContext(c)
                if tc == nil {
                        c.JSON(401, gin.H{"error": "未认证"})
                        c.Abort()
                        return
                }

                // Owner拥有所有权限
                if tc.IsOwner {
                        c.Next()
                        return
                }

                // 检查权限
                if !hasPermission(tc.Permissions, permission) {
                        c.JSON(403, gin.H{"error": "权限不足"})
                        c.Abort()
                        return
                }

                c.Next()
        }
}

// RequireAdmin 管理员权限中间件
func (s *IsolationService) RequireAdmin() gin.HandlerFunc {
        return func(c *gin.Context) {
                tc := GetTenantContext(c)
                if tc == nil || !tc.IsAdmin {
                        c.JSON(403, gin.H{"error": "需要管理员权限"})
                        c.Abort()
                        return
                }
                c.Next()
        }
}

// RequireOwner 所有者权限中间件
func (s *IsolationService) RequireOwner() gin.HandlerFunc {
        return func(c *gin.Context) {
                tc := GetTenantContext(c)
                if tc == nil || !tc.IsOwner {
                        c.JSON(403, gin.H{"error": "需要所有者权限"})
                        c.Abort()
                        return
                }
                c.Next()
        }
}

// ScopedDB 获取租户作用域的数据库连接
func (s *IsolationService) ScopedDB(c *gin.Context) *gorm.DB {
        tc := GetTenantContext(c)
        if tc == nil {
                return nil
        }
        return s.db.Where("tenant_id = ?", tc.TenantID)
}

// WithTenantScope 为查询添加租户隔离条件
func (s *IsolationService) WithTenantScope(db *gorm.DB, tenantID string) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
}

// CheckResourceAccess 检查资源访问权限
func (s *IsolationService) CheckResourceAccess(c *gin.Context, resourceType, resourceID string) (bool, error) {
        tc := GetTenantContext(c)
        if tc == nil {
                return false, errors.New("无租户上下文")
        }

        // Owner可以访问所有资源
        if tc.IsOwner {
                return true, nil
        }

        // 检查资源是否属于该租户
        var count int64
        result := s.db.Table(resourceType).
                Where("id = ? AND tenant_id = ?", resourceID, tc.TenantID).
                Count(&count)

        if result.Error != nil {
                return false, result.Error
        }

        return count > 0, nil
}

// CheckQuota 检查配额
func (s *IsolationService) CheckQuota(tenantID, quotaType string, delta int) (bool, error) {
        var quota tenantModel.TenantQuota
        if err := s.db.Where("tenant_id = ?", tenantID).First(&quota).Error; err != nil {
                return false, err
        }

        switch quotaType {
        case "users":
                return quota.CurrentUsers+delta <= quota.MaxUsers || quota.MaxUsers == -1, nil
        case "resources":
                return quota.CurrentResources+delta <= quota.MaxResources || quota.MaxResources == -1, nil
        case "storage":
                return quota.CurrentStorage+delta <= quota.MaxStorageGB || quota.MaxStorageGB == -1, nil
        case "api_calls":
                return quota.CurrentAPICalls+delta <= quota.MaxAPICalls || quota.MaxAPICalls == -1, nil
        }

        return true, nil
}

// IncrementUsage 增加使用量
func (s *IsolationService) IncrementUsage(tenantID, usageType string, delta int) error {
        quotaField := map[string]string{
                "users":     "current_users",
                "resources": "current_resources",
                "storage":   "current_storage_gb",
                "api_calls": "current_api_calls",
        }

        field, ok := quotaField[usageType]
        if !ok {
                return fmt.Errorf("未知的使用类型: %s", usageType)
        }

        return s.db.Model(&tenantModel.TenantQuota{}).
                Where("tenant_id = ?", tenantID).
                UpdateColumn(field, gorm.Expr(field+" + ?", delta)).Error
}

// DecrementUsage 减少使用量
func (s *IsolationService) DecrementUsage(tenantID, usageType string, delta int) error {
        return s.IncrementUsage(tenantID, usageType, -delta)
}

// 私有方法
func (s *IsolationService) getTenantByID(id string) (*tenantModel.Tenant, error) {
        var tenant tenantModel.Tenant
        err := s.db.First(&tenant, "id = ?", id).Error
        if err != nil {
                return nil, err
        }
        return &tenant, nil
}

func (s *IsolationService) getTenantBySlug(slug string) (*tenantModel.Tenant, error) {
        var tenant tenantModel.Tenant
        err := s.db.First(&tenant, "slug = ?", slug).Error
        if err != nil {
                return nil, err
        }
        return &tenant, nil
}

func (s *IsolationService) getTenantByDomain(domain string) (*tenantModel.Tenant, error) {
        var tenant tenantModel.Tenant
        err := s.db.First(&tenant, "domain = ?", domain).Error
        if err != nil {
                return nil, err
        }
        return &tenant, nil
}

// hasPermission 检查权限
func hasPermission(permissions []string, required string) bool {
        for _, p := range permissions {
                // 完全匹配
                if p == "*" || p == required {
                        return true
                }
                // 通配符匹配 (如 "users:*" 匹配 "users:create")
                if strings.HasSuffix(p, ":*") {
                        prefix := strings.TrimSuffix(p, "*")
                        if strings.HasPrefix(required, prefix) {
                                return true
                        }
                }
        }
        return false
}
