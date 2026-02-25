package tenant

import (
        tenantModel "yunwei/model/tenant"
        "github.com/gin-gonic/gin"
)

// TenantContext 租户上下文
type TenantContext struct {
        TenantID    string
        TenantSlug  string
        Tenant      *tenantModel.Tenant
        UserID      string
        UserRole    string
        IsOwner     bool
        IsAdmin     bool
        Permissions []string
}

// contextKey 上下文键类型
type contextKey string

const (
        // TenantContextKey 租户上下文键
        TenantContextKey contextKey = "tenant_context"
)

// GetTenantContext 从Gin上下文获取租户信息
func GetTenantContext(c *gin.Context) *TenantContext {
        if tc, exists := c.Get(string(TenantContextKey)); exists {
                return tc.(*TenantContext)
        }
        return nil
}

// SetTenantContext 设置租户上下文到Gin上下文
func SetTenantContext(c *gin.Context, tc *TenantContext) {
        c.Set(string(TenantContextKey), tc)
}

// MustGetTenantContext 获取租户上下文，如果不存在则panic
func MustGetTenantContext(c *gin.Context) *TenantContext {
        tc := GetTenantContext(c)
        if tc == nil {
                panic("tenant context not found")
        }
        return tc
}

// GetTenantID 获取当前租户ID
func GetTenantID(c *gin.Context) string {
        tc := GetTenantContext(c)
        if tc == nil {
                return ""
        }
        return tc.TenantID
}

// GetUserID 获取当前用户ID
func GetUserID(c *gin.Context) string {
        tc := GetTenantContext(c)
        if tc == nil {
                return ""
        }
        return tc.UserID
}

// IsOwner 检查当前用户是否为所有者
func IsOwner(c *gin.Context) bool {
        tc := GetTenantContext(c)
        return tc != nil && tc.IsOwner
}

// IsAdmin 检查当前用户是否为管理员
func IsAdmin(c *gin.Context) bool {
        tc := GetTenantContext(c)
        return tc != nil && tc.IsAdmin
}

// HasPermission 检查当前用户是否有指定权限
func HasPermission(c *gin.Context, permission string) bool {
        tc := GetTenantContext(c)
        if tc == nil {
                return false
        }

        // Owner拥有所有权限
        if tc.IsOwner {
                return true
        }

        // 检查权限列表
        for _, p := range tc.Permissions {
                if p == "*" || p == permission {
                        return true
                }
        }

        return false
}

// RequireTenant 租户识别中间件（简化版）
func RequireTenant() gin.HandlerFunc {
        return func(c *gin.Context) {
                // 从请求头获取租户ID
                tenantID := c.GetHeader("X-Tenant-ID")
                if tenantID == "" {
                        c.JSON(401, gin.H{"error": "需要提供租户标识"})
                        c.Abort()
                        return
                }

                // 创建基础租户上下文
                tc := &TenantContext{
                        TenantID: tenantID,
                }
                SetTenantContext(c, tc)
                c.Next()
        }
}

// RequireAuth 认证中间件（简化版）
func RequireAuth() gin.HandlerFunc {
        return func(c *gin.Context) {
                tc := GetTenantContext(c)
                if tc == nil {
                        c.JSON(401, gin.H{"error": "租户上下文不存在"})
                        c.Abort()
                        return
                }

                // 从请求头获取用户ID
                userID := c.GetHeader("X-User-ID")
                if userID == "" {
                        c.JSON(401, gin.H{"error": "未登录"})
                        c.Abort()
                        return
                }

                tc.UserID = userID
                SetTenantContext(c, tc)
                c.Next()
        }
}

// RequireAdmin 管理员权限中间件
func RequireAdmin() gin.HandlerFunc {
        return func(c *gin.Context) {
                if !IsAdmin(c) {
                        c.JSON(403, gin.H{"error": "需要管理员权限"})
                        c.Abort()
                        return
                }
                c.Next()
        }
}

// RequireOwner 所有者权限中间件
func RequireOwner() gin.HandlerFunc {
        return func(c *gin.Context) {
                if !IsOwner(c) {
                        c.JSON(403, gin.H{"error": "需要所有者权限"})
                        c.Abort()
                        return
                }
                c.Next()
        }
}

// RequirePermission 权限检查中间件
func RequirePermission(permission string) gin.HandlerFunc {
        return func(c *gin.Context) {
                if !HasPermission(c, permission) {
                        c.JSON(403, gin.H{"error": "权限不足"})
                        c.Abort()
                        return
                }
                c.Next()
        }
}
