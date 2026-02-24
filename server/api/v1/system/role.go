package system

import (
	"gin-vue-admin/global"
	"gin-vue-admin/model/common/response"
	"gin-vue-admin/model/request"
	"gin-vue-admin/model/system"

	"github.com/gin-gonic/gin"
)

type RoleApi struct{}

// GetRoleList 获取角色列表
// @Summary 获取角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Success 200 {object} response.Response
// @Router /api/v1/roles [get]
func (r *RoleApi) GetRoleList(c *gin.Context) {
	var req request.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	var roles []system.SysRole
	var total int64

	global.DB.Model(&system.SysRole{}).Count(&total)
	global.DB.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&roles)

	response.OkWithPage(roles, total, req.Page, req.PageSize, c)
}

// GetRole 获取角色详情
// @Summary 获取角色详情
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id} [get]
func (r *RoleApi) GetRole(c *gin.Context) {
	var req struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	var role system.SysRole
	if err := global.DB.First(&role, req.ID).Error; err != nil {
		response.FailWithMessage("角色不存在", c)
		return
	}

	// 获取角色的菜单ID
	var menuIds []uint
	global.DB.Model(&system.SysRoleMenu{}).Where("role_id = ?", req.ID).Pluck("menu_id", &menuIds)

	// 获取角色的API ID
	var apiIds []uint
	global.DB.Model(&system.SysRoleApi{}).Where("role_id = ?", req.ID).Pluck("api_id", &apiIds)

	response.OkWithData(gin.H{
		"role":    role,
		"menuIds": menuIds,
		"apiIds":  apiIds,
	}, c)
}

// CreateRole 创建角色
// @Summary 创建角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param data body request.RoleCreate true "角色信息"
// @Success 200 {object} response.Response
// @Router /api/v1/roles [post]
func (r *RoleApi) CreateRole(c *gin.Context) {
	var req request.RoleCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 检查关键字是否存在
	var count int64
	global.DB.Model(&system.SysRole{}).Where("keyword = ?", req.Keyword).Count(&count)
	if count > 0 {
		response.FailWithMessage("角色关键字已存在", c)
		return
	}

	// 创建角色
	role := system.SysRole{
		Name:        req.Name,
		Keyword:     req.Keyword,
		Description: req.Description,
		Status:      1,
	}

	tx := global.DB.Begin()
	if err := tx.Create(&role).Error; err != nil {
		tx.Rollback()
		response.FailWithMessage("创建失败", c)
		return
	}

	// 关联菜单
	for _, menuId := range req.MenuIds {
		roleMenu := system.SysRoleMenu{RoleID: role.ID, MenuID: menuId}
		tx.Create(&roleMenu)
	}

	// 关联API
	for _, apiId := range req.ApiIds {
		roleApi := system.SysRoleApi{RoleID: role.ID, ApiID: apiId}
		tx.Create(&roleApi)
	}

	tx.Commit()
	response.OkWithMessage("创建成功", c)
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param data body request.RoleUpdate true "角色信息"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id} [put]
func (r *RoleApi) UpdateRole(c *gin.Context) {
	var req request.RoleUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	tx := global.DB.Begin()

	// 更新角色基本信息
	updates := map[string]interface{}{
		"name":        req.Name,
		"keyword":     req.Keyword,
		"description": req.Description,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := tx.Model(&system.SysRole{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		tx.Rollback()
		response.FailWithMessage("更新失败", c)
		return
	}

	// 更新菜单关联
	tx.Where("role_id = ?", req.ID).Delete(&system.SysRoleMenu{})
	for _, menuId := range req.MenuIds {
		roleMenu := system.SysRoleMenu{RoleID: req.ID, MenuID: menuId}
		tx.Create(&roleMenu)
	}

	// 更新API关联
	tx.Where("role_id = ?", req.ID).Delete(&system.SysRoleApi{})
	for _, apiId := range req.ApiIds {
		roleApi := system.SysRoleApi{RoleID: req.ID, ApiID: apiId}
		tx.Create(&roleApi)
	}

	tx.Commit()
	response.OkWithMessage("更新成功", c)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id} [delete]
func (r *RoleApi) DeleteRole(c *gin.Context) {
	var req struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	// 检查是否有用户使用该角色
	var count int64
	global.DB.Model(&system.SysUser{}).Where("role_id = ?", req.ID).Count(&count)
	if count > 0 {
		response.FailWithMessage("该角色下存在用户，无法删除", c)
		return
	}

	tx := global.DB.Begin()
	
	// 删除角色关联的菜单
	tx.Where("role_id = ?", req.ID).Delete(&system.SysRoleMenu{})
	// 删除角色关联的API
	tx.Where("role_id = ?", req.ID).Delete(&system.SysRoleApi{})
	// 删除角色
	if err := tx.Delete(&system.SysRole{}, req.ID).Error; err != nil {
		tx.Rollback()
		response.FailWithMessage("删除失败", c)
		return
	}

	tx.Commit()
	response.OkWithMessage("删除成功", c)
}
