package system

import (
	"yunwei/global"
	"yunwei/model/common/response"
	"yunwei/model/request"
	"yunwei/model/system"
	"yunwei/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserApi struct{}

// Login 用户登录
// @Summary 用户登录
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body request.Login true "登录信息"
// @Success 200 {object} response.Response
// @Router /api/v1/login [post]
func (u *UserApi) Login(c *gin.Context) {
	var req request.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 查询用户
	var user system.SysUser
	if err := global.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		response.FailWithMessage("用户不存在", c)
		return
	}

	// 验证密码
	if user.Password != utils.MD5(req.Password) {
		response.FailWithMessage("密码错误", c)
		return
	}

	// 检查状态
	if user.Status != 1 {
		response.FailWithMessage("用户已被禁用", c)
		return
	}

	// 生成 Token
	token, err := utils.GenerateToken(user.ID, user.Username, user.NickName, user.RoleID)
	if err != nil {
		response.FailWithMessage("生成Token失败", c)
		return
	}

	response.OkWithData(gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"nickName": user.NickName,
			"avatar":   user.Avatar,
			"roleId":   user.RoleID,
		},
	}, c)
}

// Register 用户注册
// @Summary 用户注册
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body request.Register true "注册信息"
// @Success 200 {object} response.Response
// @Router /api/v1/register [post]
func (u *UserApi) Register(c *gin.Context) {
	var req request.Register
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 检查用户是否存在
	var count int64
	global.DB.Model(&system.SysUser{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		response.FailWithMessage("用户名已存在", c)
		return
	}

	// 创建用户
	user := system.SysUser{
		Username: req.Username,
		Password: utils.MD5(req.Password),
		NickName: req.NickName,
		Status:   1,
	}

	if err := global.DB.Create(&user).Error; err != nil {
		response.FailWithMessage("注册失败", c)
		return
	}

	response.OkWithMessage("注册成功", c)
}

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Param username query string false "用户名"
// @Success 200 {object} response.Response
// @Router /api/v1/users [get]
func (u *UserApi) GetUserList(c *gin.Context) {
	var req request.UserList
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	var users []system.SysUser
	var total int64

	db := global.DB.Model(&system.SysUser{}).Preload("Role")

	if req.Username != "" {
		db = db.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}

	db.Count(&total)
	db.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&users)

	response.OkWithPage(users, total, req.Page, req.PageSize, c)
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (u *UserApi) GetUser(c *gin.Context) {
	var req request.UserInfo
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	var user system.SysUser
	if err := global.DB.Preload("Role").First(&user, req.ID).Error; err != nil {
		response.FailWithMessage("用户不存在", c)
		return
	}

	response.OkWithData(user, c)
}

// CreateUser 创建用户
// @Summary 创建用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body request.UserCreate true "用户信息"
// @Success 200 {object} response.Response
// @Router /api/v1/users [post]
func (u *UserApi) CreateUser(c *gin.Context) {
	var req request.UserCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 检查用户是否存在
	var count int64
	global.DB.Model(&system.SysUser{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		response.FailWithMessage("用户名已存在", c)
		return
	}

	user := system.SysUser{
		Username: req.Username,
		Password: utils.MD5(req.Password),
		NickName: req.NickName,
		Email:    req.Email,
		Phone:    req.Phone,
		RoleID:   req.RoleId,
		Status:   1,
	}

	if err := global.DB.Create(&user).Error; err != nil {
		response.FailWithMessage("创建失败", c)
		return
	}

	response.OkWithMessage("创建成功", c)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body request.UserUpdate true "用户信息"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [put]
func (u *UserApi) UpdateUser(c *gin.Context) {
	var req request.UserUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	updates := map[string]interface{}{
		"nick_name": req.NickName,
		"email":     req.Email,
		"phone":     req.Phone,
		"role_id":   req.RoleId,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := global.DB.Model(&system.SysUser{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		response.FailWithMessage("更新失败", c)
		return
	}

	response.OkWithMessage("更新成功", c)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [delete]
func (u *UserApi) DeleteUser(c *gin.Context) {
	var req request.UserInfo
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	if err := global.DB.Delete(&system.SysUser{}, req.ID).Error; err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.OkWithMessage("删除成功", c)
}

// GetUserinfo 获取当前用户信息
// @Summary 获取当前用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/user/info [get]
func (u *UserApi) GetUserinfo(c *gin.Context) {
	claims, _ := c.Get("claims")
	userClaims := claims.(*utils.CustomClaims)

	var user system.SysUser
	if err := global.DB.Preload("Role").First(&user, userClaims.ID).Error; err != nil {
		response.FailWithMessage("获取用户信息失败", c)
		return
	}

	response.OkWithData(user, c)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body request.SetUserPassword true "密码信息"
// @Success 200 {object} response.Response
// @Router /api/v1/user/password [put]
func (u *UserApi) ChangePassword(c *gin.Context) {
	var req request.SetUserPassword
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	claims, _ := c.Get("claims")
	userClaims := claims.(*utils.CustomClaims)

	var user system.SysUser
	if err := global.DB.First(&user, userClaims.ID).Error; err != nil {
		response.FailWithMessage("用户不存在", c)
		return
	}

	// 验证旧密码
	if user.Password != utils.MD5(req.Password) {
		response.FailWithMessage("旧密码错误", c)
		return
	}

	// 更新密码
	if err := global.DB.Model(&user).Update("password", utils.MD5(req.NewPassword)).Error; err != nil {
		response.FailWithMessage("修改失败", c)
		return
	}

	response.OkWithMessage("修改成功", c)
}
