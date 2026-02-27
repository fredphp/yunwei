package main

import (
        "yunwei/config"
        "yunwei/global"
        "yunwei/router"
        "yunwei/api/v1/scheduler"
        "fmt"

        "github.com/gin-gonic/gin"
)

func main() {
        // 初始化配置
        config.Init()

        // 初始化数据库
        global.InitDB()

        // 初始化日志
        global.InitLogger()

        // 初始化任务调度中心
        scheduler.InitJobCenter()
        scheduler.GetJobCenter().Start()

        // 设置 Gin 模式
        if config.CONFIG.System.Env == "production" {
                gin.SetMode(gin.ReleaseMode)
        }

        // 初始化 Gin
        r := gin.New()
        r.Use(gin.Recovery())

        // 初始化路由
        router.InitRouter(r)

        // 启动服务
        fmt.Printf(`
        ╔═══════════════════════════════════════════════════════════╗
        ║                                                           ║
        ║     AI 自动化运维管理系统 启动成功!                       ║
        ║                                                           ║
        ║     HTTP:  http://localhost:%s                            ║
        ║     gRPC:  localhost:%s                                   ║
        ║                                                           ║
        ╚═══════════════════════════════════════════════════════════╝
        `, config.CONFIG.System.Port, config.CONFIG.System.GrpcPort)

        r.Run(":" + config.CONFIG.System.Port)
}
