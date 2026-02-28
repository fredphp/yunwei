package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		
		// 简单的日志输出
		gin.DefaultWriter.Write([]byte(
			"[" + time.Now().Format("2006/01/02 - 15:04:05") + "] " +
			c.Request.Method + " " + path + " " +
			string(rune(c.Writer.Status())) + " " +
			latency.String() + "\n",
		))
	}
}
