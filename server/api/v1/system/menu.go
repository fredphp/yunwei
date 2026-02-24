package system

import (
	"gin-vue-admin/global"
	"gin-vue-admin/model/common/response"
	"gin-vue-admin/model/request"
	"gin-vue-admin/model/system"

	"github.com/gin-gonic/gin"
)

type MenuApi struct{}

// GetMenuList 获取菜单列表
// @Summary 获取菜单列表
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/menus [get]
func (m *MenuApi) GetMenuList(c *gin.Context) {
	var menus []system.SysMenu
	global.DB.Where("status = ?", 1).Order("sort").Find(&menus)

	// 构建树形结构
	menuTree := buildMenuTree(menus, 0)
	response.OkWithData(menuTree, c)
}

// buildMenuTree 构建菜单树
func buildMenuTree(menus []system.SysMenu, parentId uint) []system.SysMenu {
	var tree []system.SysMenu
	for _, menu := range menus {
		if menu.ParentID == parentId {
			menu.Children = buildMenuTree(menus, menu.ID)
			tree = append(tree, menu)
		}
	}
	return tree
}

// GetMenu 获取菜单详情
// @Summary 获取菜单详情
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path int true "菜单ID"
// @Success 200 {object} response.Response
// @Router /api/v1/menus/{id} [get]
func (m *MenuApi) GetMenu(c *gin.Context) {
	var req struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	var menu system.SysMenu
	if err := global.DB.First(&menu, req.ID).Error; err != nil {
		response.FailWithMessage("菜单不存在", c)
		return
	}

	response.OkWithData(menu, c)
}

// CreateMenu 创建菜单
// @Summary 创建菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param data body request.MenuCreate true "菜单信息"
// @Success 200 {object} response.Response
// @Router /api/v1/menus [post]
func (m *MenuApi) CreateMenu(c *gin.Context) {
	var req request.MenuCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	menu := system.SysMenu{
		ParentID:  req.ParentID,
		Title:     req.Title,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		Sort:      req.Sort,
		Hidden:    req.Hidden,
		Status:    1,
	}

	if err := global.DB.Create(&menu).Error; err != nil {
		response.FailWithMessage("创建失败", c)
		return
	}

	response.OkWithMessage("创建成功", c)
}

// UpdateMenu 更新菜单
// @Summary 更新菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param data body request.MenuUpdate true "菜单信息"
// @Success 200 {object} response.Response
// @Router /api/v1/menus/{id} [put]
func (m *MenuApi) UpdateMenu(c *gin.Context) {
	var req request.MenuUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	updates := map[string]interface{}{
		"parent_id":  req.ParentID,
		"title":      req.Title,
		"name":       req.Name,
		"path":       req.Path,
		"component":  req.Component,
		"icon":       req.Icon,
		"sort":       req.Sort,
		"hidden":     req.Hidden,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := global.DB.Model(&system.SysMenu{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		response.FailWithMessage("更新失败", c)
		return
	}

	response.OkWithMessage("更新成功", c)
}

// DeleteMenu 删除菜单
// @Summary 删除菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path int true "菜单ID"
// @Success 200 {object} response.Response
// @Router /api/v1/menus/{id} [delete]
func (m *MenuApi) DeleteMenu(c *gin.Context) {
	var req struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	// 检查是否有子菜单
	var count int64
	global.DB.Model(&system.SysMenu{}).Where("parent_id = ?", req.ID).Count(&count)
	if count > 0 {
		response.FailWithMessage("存在子菜单，无法删除", c)
		return
	}

	if err := global.DB.Delete(&system.SysMenu{}, req.ID).Error; err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.OkWithMessage("删除成功", c)
}

// GetUserMenus 获取用户菜单
// @Summary 获取用户菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/user/menus [get]
func (m *MenuApi) GetUserMenus(c *gin.Context) {
	// TODO: 根据用户角色获取对应的菜单
	var menus []system.SysMenu
	global.DB.Where("status = ?", 1).Order("sort").Find(&menus)

	menuTree := buildMenuTree(menus, 0)
	response.OkWithData(menuTree, c)
}
