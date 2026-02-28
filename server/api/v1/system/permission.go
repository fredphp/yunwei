package system

import (
        "strconv"

        "yunwei/global"
        "yunwei/model/common/response"
        security "yunwei/service/security"

        "github.com/gin-gonic/gin"
)

// ==================== 权限管理 ====================

// GetPermissions 获取所有权限
func GetPermissions(c *gin.Context) {
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
func GetPermissionGroups(c *gin.Context) {
        groupSet := make(map[string]bool)
        for _, perm := range security.PredefinedPermissions {
                groupSet[perm.Group] = true
        }

        groupList := []string{}
        for group := range groupSet {
                groupList = append(groupList, group)
        }

        response.OkWithData(groupList, c)
}

// ==================== 角色管理 ====================

// GetRoles 获取所有角色
func GetRoles(c *gin.Context) {
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
func GetRole(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        rbac := security.NewRBACManager(global.DB)
        role, err := rbac.GetRoleByID(uint(id))
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
        Permissions []string `json:"permissions"`
}

// CreateRole 创建角色
func CreateRole(c *gin.Context) {
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

        rbac := security.NewRBACManager(global.DB)
        if err := rbac.CreateRole(role); err != nil {
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
func UpdateRole(c *gin.Context) {
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

        // 检查是否是超级管理员（RoleID == 1）
        isAdmin, _ := c.Get("isAdmin")
        isSuperAdmin := isAdmin == true

        // 只有超级管理员可以修改系统角色
        if role.IsSystem && !isSuperAdmin {
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
func DeleteRole(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        // 检查是否是超级管理员
        isAdmin, _ := c.Get("isAdmin")
        isSuperAdmin := isAdmin == true

        var role security.Role
        if err := global.DB.First(&role, id).Error; err != nil {
                response.FailWithMessage("角色不存在", c)
                return
        }

        // 只有超级管理员可以删除系统角色
        if role.IsSystem && !isSuperAdmin {
                response.FailWithMessage("系统角色不可删除", c)
                return
        }

        // 检查是否有用户使用此角色
        var userCount int64
        global.DB.Table("user_roles").Where("role_id = ?", id).Count(&userCount)
        if userCount > 0 {
                response.FailWithMessage("该角色下存在用户，无法删除", c)
                return
        }

        // 删除角色关联的权限
        global.DB.Exec("DELETE FROM role_permissions WHERE role_id = ?", id)
        
        // 删除角色
        if err := global.DB.Delete(&role).Error; err != nil {
                response.FailWithMessage("删除失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// ==================== 用户权限管理 ====================

// GetUserPermissions 获取当前用户权限
func GetUserPermissions(c *gin.Context) {
        // 从JWT获取用户ID
        userID, exists := c.Get("userID")
        if !exists {
                response.FailWithMessage("未登录", c)
                return
        }

        uid, _ := userID.(uint)

        rbac := security.NewRBACManager(global.DB)
        permissions, err := rbac.GetUserPermissions(uid)
        if err != nil {
                response.FailWithMessage("获取权限失败", c)
                return
        }

        roles, err := rbac.GetUserRoles(uid)
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

        // 角色代码列表
        roleCodes := make([]string, len(roles))
        for i, r := range roles {
                roleCodes[i] = r.Code
        }

        response.OkWithData(gin.H{
                "permissions": permissions,
                "codes":       codes,
                "permMap":     permMap,
                "roles":       roles,
                "roleCodes":   roleCodes,
                "isAdmin":     rbac.IsAdmin(uid),
        }, c)
}

// CheckPermissionRequest 检查权限请求
type CheckPermissionRequest struct {
        Permission string `json:"permission" binding:"required"`
}

// CheckPermission 检查用户权限
func CheckPermission(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
                response.FailWithMessage("未登录", c)
                return
        }

        var req CheckPermissionRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        uid, _ := userID.(uint)
        rbac := security.NewRBACManager(global.DB)
        hasPermission := rbac.CheckPermission(uid, req.Permission)

        response.OkWithData(gin.H{
                "hasPermission": hasPermission,
        }, c)
}

// CheckPermissionsRequest 批量检查权限请求
type CheckPermissionsRequest struct {
        Permissions []string `json:"permissions" binding:"required"`
}

// CheckPermissions 批量检查权限
func CheckPermissions(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
                response.FailWithMessage("未登录", c)
                return
        }

        var req CheckPermissionsRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        uid, _ := userID.(uint)
        rbac := security.NewRBACManager(global.DB)
        results := rbac.CheckPermissions(uid, req.Permissions)

        response.OkWithData(results, c)
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
        UserID uint `json:"userId" binding:"required"`
        RoleID uint `json:"roleId" binding:"required"`
}

// AssignRole 给用户分配角色
func AssignRole(c *gin.Context) {
        var req AssignRoleRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        rbac := security.NewRBACManager(global.DB)
        if err := rbac.AssignRole(req.UserID, req.RoleID); err != nil {
                response.FailWithMessage("分配角色失败", c)
                return
        }

        response.OkWithMessage("分配成功", c)
}

// RevokeRoleRequest 撤销角色请求
type RevokeRoleRequest struct {
        UserID uint `json:"userId" binding:"required"`
        RoleID uint `json:"roleId" binding:"required"`
}

// RevokeRole 撤销用户角色
func RevokeRole(c *gin.Context) {
        var req RevokeRoleRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        rbac := security.NewRBACManager(global.DB)
        if err := rbac.RevokeRole(req.UserID, req.RoleID); err != nil {
                response.FailWithMessage("撤销角色失败", c)
                return
        }

        response.OkWithMessage("撤销成功", c)
}

// GetUserRolePermissions 获取用户的角色和权限
func GetUserRolePermissions(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        rbac := security.NewRBACManager(global.DB)
        roles, err := rbac.GetUserRoles(uint(id))
        if err != nil {
                response.FailWithMessage("获取角色失败", c)
                return
        }

        permissions, err := rbac.GetUserPermissions(uint(id))
        if err != nil {
                response.FailWithMessage("获取权限失败", c)
                return
        }

        response.OkWithData(gin.H{
                "roles":       roles,
                "permissions": permissions,
        }, c)
}
