package auth

import (
	"yunwei/global"
	"yunwei/model/common/response"
	"yunwei/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	NickName string `json:"nickName"`
	Role     string `json:"role"`
}

func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	// 查询用户
	var user struct {
		ID       uint
		Username string
		Password string
		NickName string
		Role     string
		Status   int
	}

	result := global.DB.Table("users").
		Where("username = ?", req.Username).
		First(&user)

	if result.Error != nil {
		response.FailWithMessage("用户不存在", c)
		return
	}

	// 验证密码 (简化版本，实际应该使用加密)
	if user.Password != utils.MD5(req.Password) {
		response.FailWithMessage("密码错误", c)
		return
	}

	if user.Status != 1 {
		response.FailWithMessage("用户已被禁用", c)
		return
	}

	// 生成 Token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, 0)
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
			"role":     user.Role,
		},
	}, c)
}

func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		NickName string `json:"nickName"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	// 检查用户是否存在
	var count int64
	global.DB.Table("users").Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		response.FailWithMessage("用户名已存在", c)
		return
	}

	// 创建用户
	user := map[string]interface{}{
		"username":   req.Username,
		"password":   utils.MD5(req.Password),
		"nick_name":  req.NickName,
		"role":       "user",
		"status":     1,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	if err := global.DB.Table("users").Create(&user).Error; err != nil {
		response.FailWithMessage("注册失败", c)
		return
	}

	response.OkWithMessage("注册成功", c)
}

func GetUserInfo(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.FailWithMessage("未登录", c)
		return
	}

	userClaims := claims.(*utils.CustomClaims)

	var user User
	global.DB.Table("users").
		Select("id, username, nick_name, role").
		Where("id = ?", userClaims.ID).
		First(&user)

	response.OkWithData(user, c)
}
