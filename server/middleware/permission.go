package middleware

import (
	"strconv"
	"strings"

	"yunwei/global"
	"yunwei/model/common/response"
	security "yunwei/service/security"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限中间件
func PermissionMiddleware() gin.HandlerFunc {
	rbacManager := security.NewRBACManager(global.DB)

	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			response.FailWithMessage("未登录", c)
			c.Abort()
			return
		}

		// 超级管理员跳过权限检查
		if isAdmin, _ := c.Get("isAdmin"); isAdmin == true {
			c.Next()
			return
		}

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 映射到权限代码
		permission := mapToPermission(path, method)
		if permission == "" {
			// 没有对应权限要求，放行
			c.Next()
			return
		}

		// 检查权限
		uid, _ := userID.(uint)
		if !rbacManager.CheckPermission(uid, permission) {
			response.FailWithMessage("权限不足: "+permission, c)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission 要求特定权限
func RequirePermission(permission string) gin.HandlerFunc {
	rbacManager := security.NewRBACManager(global.DB)

	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			response.FailWithMessage("未登录", c)
			c.Abort()
			return
		}

		// 超级管理员跳过权限检查
		if isAdmin, _ := c.Get("isAdmin"); isAdmin == true {
			c.Next()
			return
		}

		uid, _ := userID.(uint)
		if !rbacManager.CheckPermission(uid, permission) {
			response.FailWithMessage("权限不足", c)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 要求任意一个权限
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	rbacManager := security.NewRBACManager(global.DB)

	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			response.FailWithMessage("未登录", c)
			c.Abort()
			return
		}

		// 超级管理员跳过权限检查
		if isAdmin, _ := c.Get("isAdmin"); isAdmin == true {
			c.Next()
			return
		}

		uid, _ := userID.(uint)
		for _, perm := range permissions {
			if rbacManager.CheckPermission(uid, perm) {
				c.Next()
				return
			}
		}

		response.FailWithMessage("权限不足", c)
		c.Abort()
	}
}

// RequireAllPermissions 要求所有权限
func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	rbacManager := security.NewRBACManager(global.DB)

	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			response.FailWithMessage("未登录", c)
			c.Abort()
			return
		}

		// 超级管理员跳过权限检查
		if isAdmin, _ := c.Get("isAdmin"); isAdmin == true {
			c.Next()
			return
		}

		uid, _ := userID.(uint)
		for _, perm := range permissions {
			if !rbacManager.CheckPermission(uid, perm) {
				response.FailWithMessage("权限不足: "+perm, c)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireRole 要求特定角色
func RequireRole(roleCodes ...string) gin.HandlerFunc {
	rbacManager := security.NewRBACManager(global.DB)

	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			response.FailWithMessage("未登录", c)
			c.Abort()
			return
		}

		uid, _ := userID.(uint)
		for _, code := range roleCodes {
			if rbacManager.HasRole(uid, code) {
				c.Next()
				return
			}
		}

		response.FailWithMessage("需要特定角色权限", c)
		c.Abort()
	}
}

// RequireAdmin 要求管理员权限
func RequireAdmin() gin.HandlerFunc {
	rbacManager := security.NewRBACManager(global.DB)

	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			response.FailWithMessage("未登录", c)
			c.Abort()
			return
		}

		uid, _ := userID.(uint)
		if !rbacManager.IsAdmin(uid) {
			response.FailWithMessage("需要管理员权限", c)
			c.Abort()
			return
		}

		c.Next()
	}
}

// mapToPermission 将请求路径映射到权限代码
func mapToPermission(path, method string) string {
	// 移除API前缀
	path = strings.TrimPrefix(path, "/api/v1")

	// 解析路径
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}

	resource := parts[0]
	action := ""

	// 根据方法和路径确定操作
	switch method {
	case "GET":
		action = "view"
		// 特殊处理
		if len(parts) > 1 {
			if parts[len(parts)-1] == "export" {
				action = "export"
			} else if parts[len(parts)-1] == "analyze" {
				action = "analyze"
			}
		}
	case "POST":
		action = "add"
		// 检查是否为特殊操作
		if len(parts) > 2 {
			id := parts[len(parts)-2]
			if _, err := strconv.ParseUint(id, 10, 64); err == nil {
				// 带ID的操作
				op := parts[len(parts)-1]
				switch op {
				case "execute", "command":
					action = "execute"
				case "approve":
					action = "approve"
				case "renew":
					action = "renew"
				case "check":
					action = "check"
				case "refresh":
					action = "view" // 刷新状态视为查看
				case "purge", "preheat":
					action = "operate"
				case "promote", "complete", "rollback", "pause", "abort":
					action = "deploy"
				case "disable":
					action = "disable"
				case "enable":
					action = "enable"
				case "cancel", "retry":
					action = "operate"
				case "trigger":
					action = "trigger"
				}
			}
		}
	case "PUT", "PATCH":
		action = "edit"
	case "DELETE":
		action = "delete"
	}

	// 资源映射
	resourceMap := map[string]string{
		"servers":      "server",
		"groups":       "server_group",
		"alerts":       "alert",
		"rules":        "alert_rule",
		"actions":      "auto_action",
		"decisions":    "ai_decision",
		"clusters":     "k8s_cluster",
		"kubernetes":   "k8s",
		"canary":       "canary",
		"loadbalancer": "lb",
		"certificates": "cert",
		"cdn":          "cdn",
		"deploy":       "deploy",
		"scheduler":    "scheduler",
		"agents":       "agent",
		"versions":     "agent_version",
		"upgrades":     "agent_upgrade",
		"gray":         "gray_strategy",
		"ha":           "ha",
		"backup":       "backup",
		"cost":         "cost",
		"tenants":      "tenant",
		"users":        "user",
		"roles":        "role",
		"menus":        "menu",
	}

	if mapped, ok := resourceMap[resource]; ok {
		resource = mapped
	}

	// 构建权限代码
	if action != "" && resource != "" {
		return resource + ":" + action
	}

	return ""
}

// GetPermissionContext 获取权限上下文
func GetPermissionContext(c *gin.Context) *PermissionContext {
	userID, _ := c.Get("userID")
	isAdmin, _ := c.Get("isAdmin")

	uid, _ := userID.(uint)

	return &PermissionContext{
		UserID:   uid,
		IsAdmin:  isAdmin == true,
		Roles:    getUserRoles(c),
		Permissions: getUserPermissions(c),
	}
}

// PermissionContext 权限上下文
type PermissionContext struct {
	UserID      uint
	IsAdmin     bool
	Roles       []string
	Permissions map[string]bool
}

// HasPermission 检查是否有权限
func (pc *PermissionContext) HasPermission(permission string) bool {
	if pc.IsAdmin {
		return true
	}
	return pc.Permissions[permission]
}

// HasAnyPermission 检查是否有任意权限
func (pc *PermissionContext) HasAnyPermission(permissions ...string) bool {
	if pc.IsAdmin {
		return true
	}
	for _, perm := range permissions {
		if pc.Permissions[perm] {
			return true
		}
	}
	return false
}

// HasRole 检查是否有角色
func (pc *PermissionContext) HasRole(roleCode string) bool {
	for _, role := range pc.Roles {
		if role == roleCode {
			return true
		}
	}
	return false
}

// getUserRoles 从上下文获取用户角色
func getUserRoles(c *gin.Context) []string {
	roles, exists := c.Get("userRoles")
	if !exists {
		return []string{}
	}
	if roleList, ok := roles.([]string); ok {
		return roleList
	}
	return []string{}
}

// getUserPermissions 从上下文获取用户权限
func getUserPermissions(c *gin.Context) map[string]bool {
	perms, exists := c.Get("userPermissions")
	if !exists {
		return map[string]bool{}
	}
	if permMap, ok := perms.(map[string]bool); ok {
		return permMap
	}
	return map[string]bool{}
}
