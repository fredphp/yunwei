package middleware

import (
	"strings"

	"yunwei/model/common/response"
	"yunwei/utils"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			token = c.Request.Header.Get("x-token")
		}

		if token == "" {
			response.FailWithMessage("未登录或登录已过期", c)
			c.Abort()
			return
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			response.FailWithMessage("Token验证失败", c)
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("claims", claims)
		c.Set("userID", claims.ID)
		
		// 检查是否是管理员角色 (RoleID = 1 为管理员)
		c.Set("isAdmin", claims.RoleID == 1)

		c.Next()
	}
}
