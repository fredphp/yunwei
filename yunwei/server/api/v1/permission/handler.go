package permission

import (
	"strconv"

	"yunwei/global"
	"yunwei/model/common/response"
	security "yunwei/service/security"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler 权限处理器
type Handler struct {
	rbac *security.RBACManager
}

// NewHandler 创建处理器
func NewHandler() *Handler {
	return &Handler{
		rbac: security.NewRBACManager(global.DB),
	}
}

// ==================== 权限管理 ====================

// GetPermissions 获取所有权限
// @Summary 获取所有权限
// @Tags 权限管理
// @Success 200 {object} response.Response
// @Router /api/v1/permissions [get]
func (h *Handler) GetPermissions(c *gin.Context) {
	// 按分组组织权限
	groups := make(map[string][]security.OperationPermission)
	for _, perm := range security.PredefinedPermissions {
		groups[perm.Group] = append(groups[perm.Group], perm)
	}

	response.OkWithData(gin.H{
		"permissions": security.PredefinedPermissions,
		"groups":      groups,
	}, c)
}

// GetPermissionGroups 获取权限分组
// @Summary 获取权限分组
// @Tags 权限管理
// @Success 200 {object} response.Response
// @Router /api/v1/permissions/groups [get]
func (h *Handler) GetPermissionGroups(c *gin.Context) {
	groups := make(map[string]string)
	for _, perm := range security.PredefinedPermissions {
		if _, exists := groups[perm.Group]; !exists {
			groups[perm.Group] = perm.Group
		}
	}

	groupList := []string{}
	for group := range groups {
		groupList = append(groupList, group)
	}

	response.OkWithData(groupList, c)
}

// ==================== 角色管理 ====================

// GetRoles 获取所有角色
// @Summary 获取所有角色
// @Tags 角色管理
// @Success 200 {object} response.Response
// @Router /api/v1/roles [get]
func (h *Handler) GetRoles(c *gin.Context) {
	var roles []security.Role

	query := global.DB.Model(&security.Role{})

	// 搜索
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	var total int64

	query.Count(&total)
	query.Preload("Permissions").Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles)

	response.OkWithData(gin.H{
		"list":     roles,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}, c)
}

// GetRole 获取角色详情
// @Summary 获取角色详情
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id} [get]
func (h *Handler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	role, err := h.rbac.GetRoleByID(uint(id))
	if err != nil {
		response.FailWithMessage("角色不存在", c)
		return
	}

	response.OkWithData(role, c)
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Code        string   `json:"code" binding:"required"`
	Description string   `json:"description"`
	Level       int      `json:"level"`
	Permissions []string `json:"permissions"` // 权限代码列表
}

// CreateRole 创建角色
// @Summary 创建角色
// @Tags 角色管理
// @Param body body CreateRoleRequest true "请求参数"
// @Success 200 {object} response.Response
// @Router /api/v1/roles [post]
func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 检查代码是否已存在
	var count int64
	global.DB.Model(&security.Role{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		response.FailWithMessage("角色代码已存在", c)
		return
	}

	// 创建角色
	role := &security.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Level:       req.Level,
		IsSystem:    false,
	}

	// 获取权限
	if len(req.Permissions) > 0 {
		var permissions []security.Permission
		global.DB.Where("code IN ?", req.Permissions).Find(&permissions)
		role.Permissions = permissions
	}

	if err := h.rbac.CreateRole(role); err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}

	response.OkWithData(role, c)
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Level       int      `json:"level"`
	Permissions []string `json:"permissions"`
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Param body body UpdateRoleRequest true "请求参数"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id} [put]
func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var role security.Role
	if err := global.DB.First(&role, id).Error; err != nil {
		response.FailWithMessage("角色不存在", c)
		return
	}

	if role.IsSystem {
		response.FailWithMessage("系统角色不可修改", c)
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	// 更新基本信息
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Level > 0 {
		updates["level"] = req.Level
	}

	global.DB.Model(&role).Updates(updates)

	// 更新权限
	if req.Permissions != nil {
		var permissions []security.Permission
		global.DB.Where("code IN ?", req.Permissions).Find(&permissions)
		global.DB.Model(&role).Association("Permissions").Replace(permissions)
	}

	response.OkWithData(role, c)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id} [delete]
func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	if err := h.rbac.DeleteRole(uint(id)); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}

	response.Ok(c)
}

// GetRolePermissions 获取角色权限
// @Summary 获取角色权限
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id}/permissions [get]
func (h *Handler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var role security.Role
	if err := global.DB.Preload("Permissions").First(&role, id).Error; err != nil {
		response.FailWithMessage("角色不存在", c)
		return
	}

	response.OkWithData(role.Permissions, c)
}

// UpdateRolePermissions 更新角色权限
// @Summary 更新角色权限
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Param body body UpdatePermissionsRequest true "请求参数"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id}/permissions [put]
func (h *Handler) UpdateRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var role security.Role
	if err := global.DB.First(&role, id).Error; err != nil {
		response.FailWithMessage("角色不存在", c)
		return
	}

	if role.IsSystem {
		response.FailWithMessage("系统角色权限不可修改", c)
		return
	}

	var req UpdatePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	var permissions []security.Permission
	global.DB.Where("code IN ?", req.PermissionCodes).Find(&permissions)

	if err := global.DB.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		response.FailWithMessage("更新权限失败", c)
		return
	}

	response.Ok(c)
}

// UpdatePermissionsRequest 更新权限请求
type UpdatePermissionsRequest struct {
	PermissionCodes []string `json:"permissionCodes"`
}

// ==================== 用户权限管理 ====================

// GetUserPermissions 获取用户权限
// @Summary 获取用户权限
// @Tags 用户权限
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id}/permissions [get]
func (h *Handler) GetUserPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	permissions, err := h.rbac.GetUserPermissions(uint(id))
	if err != nil {
		response.FailWithMessage("获取权限失败", c)
		return
	}

	roles, err := h.rbac.GetUserRoles(uint(id))
	if err != nil {
		response.FailWithMessage("获取角色失败", c)
		return
	}

	// 权限代码列表
	codes := make([]string, len(permissions))
	for i, p := range permissions {
		codes[i] = p.Code
	}

	// 权限映射
	permMap := make(map[string]bool)
	for _, p := range permissions {
		permMap[p.Code] = true
	}

	response.OkWithData(gin.H{
		"permissions": permissions,
		"codes":       codes,
		"permMap":     permMap,
		"roles":       roles,
	}, c)
}

// GetUserRoles 获取用户角色
// @Summary 获取用户角色
// @Tags 用户权限
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id}/roles [get]
func (h *Handler) GetUserRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	roles, err := h.rbac.GetUserRoles(uint(id))
	if err != nil {
		response.FailWithMessage("获取角色失败", c)
		return
	}

	response.OkWithData(roles, c)
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	RoleID uint `json:"roleId" binding:"required"`
}

// AssignRole 给用户分配角色
// @Summary 给用户分配角色
// @Tags 用户权限
// @Param id path int true "用户ID"
// @Param body body AssignRoleRequest true "请求参数"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id}/roles [post]
func (h *Handler) AssignRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	if err := h.rbac.AssignRole(uint(id), req.RoleID); err != nil {
		response.FailWithMessage("分配角色失败", c)
		return
	}

	response.Ok(c)
}

// RevokeRoleRequest 撤销角色请求
type RevokeRoleRequest struct {
	RoleID uint `json:"roleId" binding:"required"`
}

// RevokeRole 撤销用户角色
// @Summary 撤销用户角色
// @Tags 用户权限
// @Param id path int true "用户ID"
// @Param body body RevokeRoleRequest true "请求参数"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id}/roles [delete]
func (h *Handler) RevokeRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var req RevokeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	if err := h.rbac.RevokeRole(uint(id), req.RoleID); err != nil {
		response.FailWithMessage("撤销角色失败", c)
		return
	}

	response.Ok(c)
}

// BatchAssignRolesRequest 批量分配角色请求
type BatchAssignRolesRequest struct {
	UserIDs []uint `json:"userIds" binding:"required"`
	RoleID  uint   `json:"roleId" binding:"required"`
}

// BatchAssignRoles 批量分配角色
// @Summary 批量分配角色
// @Tags 用户权限
// @Param body body BatchAssignRolesRequest true "请求参数"
// @Success 200 {object} response.Response
// @Router /api/v1/users/batch-roles [post]
func (h *Handler) BatchAssignRoles(c *gin.Context) {
	var req BatchAssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	success := 0
	failed := 0
	for _, userID := range req.UserIDs {
		if err := h.rbac.AssignRole(userID, req.RoleID); err != nil {
			failed++
		} else {
			success++
		}
	}

	response.OkWithData(gin.H{
		"success": success,
		"failed":  failed,
	}, c)
}

// ==================== 权限检查 ====================

// CheckPermissionRequest 检查权限请求
type CheckPermissionRequest struct {
	UserID     uint   `json:"userId" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

// CheckPermission 检查用户权限
// @Summary 检查用户权限
// @Tags 权限检查
// @Param body body CheckPermissionRequest true "请求参数"
// @Success 200 {object} response.Response
// @Router /api/v1/permissions/check [post]
func (h *Handler) CheckPermission(c *gin.Context) {
	var req CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	hasPermission := h.rbac.CheckPermission(req.UserID, req.Permission)

	response.OkWithData(gin.H{
		"hasPermission": hasPermission,
	}, c)
}

// CheckPermissionsRequest 批量检查权限请求
type CheckPermissionsRequest struct {
	UserID      uint     `json:"userId" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

// CheckPermissions 批量检查权限
// @Summary 批量检查权限
// @Tags 权限检查
// @Param body body CheckPermissionsRequest true "请求参数"
// @Success 200 {object} response.Response
// @Router /api/v1/permissions/check-batch [post]
func (h *Handler) CheckPermissions(c *gin.Context) {
	var req CheckPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	results := h.rbac.CheckPermissions(req.UserID, req.Permissions)

	response.OkWithData(results, c)
}

// ==================== 初始化数据 ====================

// InitPermissions 初始化权限数据
func InitPermissions(db *gorm.DB) error {
	// 创建权限表
	if err := db.AutoMigrate(&security.Permission{}, &security.Role{}, &security.User{}); err != nil {
		return err
	}

	// 初始化权限
	for _, perm := range security.PredefinedPermissions {
		var existing security.Permission
		if err := db.Where("code = ?", perm.Code).First(&existing).Error; err != nil {
			// 不存在则创建
			db.Create(&security.Permission{
				Name:        perm.Name,
				Code:        perm.Code,
				Description: perm.Description,
				Group:       perm.Group,
			})
		}
	}

	// 初始化角色
	for _, roleDef := range security.PredefinedRoles {
		var existing security.Role
		if err := db.Where("code = ?", roleDef.Code).First(&existing).Error; err != nil {
			// 不存在则创建
			role := security.Role{
				Name:        roleDef.Name,
				Code:        roleDef.Code,
				Level:       roleDef.Level,
				IsSystem:    roleDef.IsSystem,
				Description: roleDef.Name,
			}

			// 获取权限
			var permissions []security.Permission
			db.Where("code IN ?", roleDef.Permissions).Find(&permissions)
			role.Permissions = permissions

			db.Create(&role)
		}
	}

	return nil
}

// RegisterRoutes 注册路由
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	// 权限管理
	permissions := r.Group("/permissions")
	{
		permissions.GET("", h.GetPermissions)
		permissions.GET("/groups", h.GetPermissionGroups)
		permissions.POST("/check", h.CheckPermission)
		permissions.POST("/check-batch", h.CheckPermissions)
	}

	// 角色管理
	roles := r.Group("/roles")
	{
		roles.GET("", h.GetRoles)
		roles.GET("/:id", h.GetRole)
		roles.POST("", h.CreateRole)
		roles.PUT("/:id", h.UpdateRole)
		roles.DELETE("/:id", h.DeleteRole)
		roles.GET("/:id/permissions", h.GetRolePermissions)
		roles.PUT("/:id/permissions", h.UpdateRolePermissions)
	}

	// 用户权限
	users := r.Group("/users")
	{
		users.GET("/:id/permissions", h.GetUserPermissions)
		users.GET("/:id/roles", h.GetUserRoles)
		users.POST("/:id/roles", h.AssignRole)
		users.DELETE("/:id/roles", h.RevokeRole)
		users.POST("/batch-roles", h.BatchAssignRoles)
	}
}
