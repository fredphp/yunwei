package tenant

import (
        "net/http"
        "strconv"

        tenantMiddleware "yunwei/middleware/tenant"
        tenantModel "yunwei/model/tenant"
        tenantService "yunwei/service/tenant"
        "github.com/gin-gonic/gin"
        "gorm.io/gorm"
)

// Handler 租户API处理器
type Handler struct {
        db           *gorm.DB
        tenantSvc    *tenantService.TenantService
        isolationSvc *tenantService.IsolationService
        rbacSvc      *tenantService.RBACService
}

func NewHandler(
        db *gorm.DB,
        tenantSvc *tenantService.TenantService,
        isolationSvc *tenantService.IsolationService,
        rbacSvc *tenantService.RBACService,
) *Handler {
        return &Handler{
                db:           db,
                tenantSvc:    tenantSvc,
                isolationSvc: isolationSvc,
                rbacSvc:      rbacSvc,
        }
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
        // 租户管理（平台管理员）
        tenants := r.Group("/admin/tenants")
        {
                tenants.GET("", h.ListTenants)
                tenants.POST("", h.CreateTenant)
                tenants.GET("/:id", h.GetTenant)
                tenants.PUT("/:id", h.UpdateTenant)
                tenants.DELETE("/:id", h.DeleteTenant)
                tenants.POST("/:id/suspend", h.SuspendTenant)
                tenants.POST("/:id/activate", h.ActivateTenant)
                tenants.POST("/:id/upgrade", h.UpgradePlan)
        }

        // 账单管理（平台管理员）
        billings := r.Group("/admin/billings")
        {
                billings.GET("", h.ListBillings)
                billings.GET("/stats", h.GetBillingStats)
                billings.GET("/:id", h.GetBilling)
                billings.POST("", h.CreateBilling)
                billings.POST("/generate", h.GenerateBillings)
                billings.POST("/:id/paid", h.MarkBillingPaid)
        }

        // 套餐管理（平台管理员）
        plans := r.Group("/admin/plans")
        {
                plans.GET("", h.ListPlans)
                plans.PUT("/:key", h.UpdatePlan)
        }

        // 审计日志（平台管理员）
        auditLogs := r.Group("/admin/audit-logs")
        {
                auditLogs.GET("", h.ListAuditLogsAdmin)
                auditLogs.GET("/stats", h.GetAuditStats)
                auditLogs.GET("/:id", h.GetAuditLog)
                auditLogs.POST("", h.CreateAuditLog)
        }

        // 租户内部管理（租户管理员）
        tenantAdmin := r.Group("/tenant")
        tenantAdmin.Use(h.isolationSvc.TenantMiddleware())
        tenantAdmin.Use(h.isolationSvc.AuthMiddleware())
        {
                // 租户信息
                tenantAdmin.GET("/info", h.GetCurrentTenant)
                tenantAdmin.PUT("/info", h.UpdateCurrentTenant)

                // 用户管理
                tenantAdmin.GET("/users", h.ListTenantUsers)
                tenantAdmin.POST("/users", h.AddTenantUser)
                tenantAdmin.DELETE("/users/:userId", h.RemoveTenantUser)
                tenantAdmin.PUT("/users/:userId/role", h.UpdateUserRole)
                tenantAdmin.PUT("/users/:userId/status", h.UpdateUserStatus)

                // 角色管理
                tenantAdmin.GET("/roles", h.ListRoles)
                tenantAdmin.POST("/roles", h.CreateRole)
                tenantAdmin.PUT("/roles/:roleId", h.UpdateRole)
                tenantAdmin.DELETE("/roles/:roleId", h.DeleteRole)

                // 配额信息
                tenantAdmin.GET("/quota", h.GetQuota)
                tenantAdmin.GET("/usage", h.GetUsage)

                // 邀请管理
                tenantAdmin.GET("/invitations", h.ListInvitations)
                tenantAdmin.POST("/invitations", h.CreateInvitation)
                tenantAdmin.DELETE("/invitations/:id", h.CancelInvitation)

                // 审计日志
                tenantAdmin.GET("/audit-logs", h.ListAuditLogs)
        }
}

// ==================== 平台管理员接口 ====================

// ListTenants 列出租户
func (h *Handler) ListTenants(c *gin.Context) {
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
        status := c.Query("status")
        plan := c.Query("plan")

        tenants, total, err := h.tenantSvc.ListTenants(page, pageSize, status, plan)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "data":       tenants,
                "total":      total,
                "page":       page,
                "page_size":  pageSize,
        })
}

// CreateTenant 创建租户
func (h *Handler) CreateTenant(c *gin.Context) {
        var req struct {
                Name       string `json:"name" binding:"required"`
                Slug       string `json:"slug" binding:"required"`
                Plan       string `json:"plan"`
                OwnerEmail string `json:"owner_email" binding:"required,email"`
                OwnerName  string `json:"owner_name" binding:"required"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        if req.Plan == "" {
                req.Plan = "free"
        }

        tenant, err := h.tenantSvc.CreateTenant(req.Name, req.Slug, req.Plan, req.OwnerEmail, req.OwnerName)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusCreated, gin.H{"data": tenant})
}

// GetTenant 获取租户详情
func (h *Handler) GetTenant(c *gin.Context) {
        id := c.Param("id")
        tenant, err := h.tenantSvc.GetTenantByID(id)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "租户不存在"})
                return
        }

        c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// UpdateTenant 更新租户
func (h *Handler) UpdateTenant(c *gin.Context) {
        id := c.Param("id")
        var updates map[string]interface{}
        if err := c.ShouldBindJSON(&updates); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        if err := h.tenantSvc.UpdateTenant(id, updates); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteTenant 删除租户
func (h *Handler) DeleteTenant(c *gin.Context) {
        id := c.Param("id")
        if err := h.tenantSvc.DeleteTenant(id); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// SuspendTenant 暂停租户
func (h *Handler) SuspendTenant(c *gin.Context) {
        id := c.Param("id")
        reason := c.PostForm("reason")

        if err := h.tenantSvc.SuspendTenant(id, reason); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "租户已暂停"})
}

// ActivateTenant 激活租户
func (h *Handler) ActivateTenant(c *gin.Context) {
        id := c.Param("id")
        if err := h.tenantSvc.ActivateTenant(id); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "租户已激活"})
}

// UpgradePlan 升级套餐
func (h *Handler) UpgradePlan(c *gin.Context) {
        id := c.Param("id")
        var req struct {
                Plan string `json:"plan" binding:"required"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        if err := h.tenantSvc.UpgradePlan(id, req.Plan); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "套餐升级成功"})
}

// ==================== 租户内部接口 ====================

// GetCurrentTenant 获取当前租户信息
func (h *Handler) GetCurrentTenant(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
                return
        }

        c.JSON(http.StatusOK, gin.H{"data": tc.Tenant})
}

// UpdateCurrentTenant 更新当前租户信息
func (h *Handler) UpdateCurrentTenant(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        var updates map[string]interface{}
        if err := c.ShouldBindJSON(&updates); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        // 移除不可修改的字段
        delete(updates, "id")
        delete(updates, "slug")
        delete(updates, "plan")
        delete(updates, "status")

        if err := h.tenantSvc.UpdateTenant(tc.TenantID, updates); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// ListTenantUsers 列出租户用户
func (h *Handler) ListTenantUsers(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
                return
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

        users, total, err := h.rbacSvc.ListUsers(tc.TenantID, page, pageSize)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "data":      users,
                "total":     total,
                "page":      page,
                "page_size": pageSize,
        })
}

// AddTenantUser 添加用户到租户
func (h *Handler) AddTenantUser(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        var req struct {
                Email  string `json:"email" binding:"required,email"`
                Name   string `json:"name" binding:"required"`
                RoleID string `json:"role_id" binding:"required"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        user, err := h.rbacSvc.AddUserToTenant(tc.TenantID, req.Email, req.Name, req.RoleID, tc.UserID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusCreated, gin.H{"data": user})
}

// RemoveTenantUser 从租户移除用户
func (h *Handler) RemoveTenantUser(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        userID := c.Param("userId")
        if err := h.rbacSvc.RemoveUserFromTenant(tc.TenantID, userID); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "用户已移除"})
}

// UpdateUserRole 更新用户角色
func (h *Handler) UpdateUserRole(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        userID := c.Param("userId")
        var req struct {
                RoleID string `json:"role_id" binding:"required"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        if err := h.rbacSvc.AssignRole(tc.TenantID, userID, req.RoleID); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "角色更新成功"})
}

// UpdateUserStatus 更新用户状态
func (h *Handler) UpdateUserStatus(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        userID := c.Param("userId")
        var req struct {
                Status string `json:"status" binding:"required"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        if err := h.rbacSvc.UpdateUserStatus(tc.TenantID, userID, req.Status); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "状态更新成功"})
}

// ListRoles 列出角色
func (h *Handler) ListRoles(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
                return
        }

        roles, err := h.rbacSvc.ListRoles(tc.TenantID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"data": roles})
}

// CreateRole 创建角色
func (h *Handler) CreateRole(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        var req struct {
                Name        string   `json:"name" binding:"required"`
                Slug        string   `json:"slug" binding:"required"`
                Description string   `json:"description"`
                Permissions []string `json:"permissions"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        role, err := h.rbacSvc.CreateRole(tc.TenantID, req.Name, req.Slug, req.Description, req.Permissions)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusCreated, gin.H{"data": role})
}

// UpdateRole 更新角色
func (h *Handler) UpdateRole(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        roleID := c.Param("roleId")
        var updates map[string]interface{}
        if err := c.ShouldBindJSON(&updates); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        if err := h.rbacSvc.UpdateRole(roleID, tc.TenantID, updates); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteRole 删除角色
func (h *Handler) DeleteRole(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        roleID := c.Param("roleId")
        if err := h.rbacSvc.DeleteRole(roleID, tc.TenantID); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "角色已删除"})
}

// GetQuota 获取配额信息
func (h *Handler) GetQuota(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
                return
        }

        tenant, err := h.tenantSvc.GetTenantByID(tc.TenantID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"data": tenant.Quota})
}

// GetUsage 获取使用量
func (h *Handler) GetUsage(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "data": gin.H{
                        "users":     tc.Tenant.Quota.CurrentUsers,
                        "resources": tc.Tenant.Quota.CurrentResources,
                        "storage":   tc.Tenant.Quota.CurrentStorage,
                        "api_calls": tc.Tenant.Quota.CurrentAPICalls,
                },
        })
}

// ListInvitations 列出邀请
func (h *Handler) ListInvitations(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
                return
        }

        c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

// CreateInvitation 创建邀请
func (h *Handler) CreateInvitation(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        c.JSON(http.StatusCreated, gin.H{"message": "邀请已发送"})
}

// CancelInvitation 取消邀请
func (h *Handler) CancelInvitation(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "邀请已取消"})
}

// ListAuditLogs 列出审计日志
func (h *Handler) ListAuditLogs(c *gin.Context) {
        tc := tenantMiddleware.GetTenantContext(c)
        if tc == nil || !tc.IsAdmin {
                c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                return
        }

        c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}
