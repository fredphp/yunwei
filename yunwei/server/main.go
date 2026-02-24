package main

import (
	"fmt"

	"yunwei/config"
	"yunwei/global"
	"yunwei/grpc"
	"yunwei/router"
	"yunwei/service/ai/llm"
	"yunwei/service/ai/decision"
	"yunwei/service/notify"
	"yunwei/service/patrol"
	"yunwei/service/prediction"
	"yunwei/service/scheduler"
	"yunwei/service/selfheal"
	"yunwei/service/workflow"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	config.Init()

	// 初始化数据库
	global.InitDB()

	// 初始化日志
	global.InitLogger()

	// 设置 Gin 模式
	if config.CONFIG.System.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化核心服务
	initServices()

	// 初始化 Gin
	r := gin.New()
	r.Use(gin.Recovery())

	// 初始化路由
	router.InitRouter(r)

	// 启动 gRPC 服务
	grpcServer := grpc.NewAgentGRPCServer(config.CONFIG.System.GrpcPort)
	if err := grpcServer.Start(); err != nil {
		global.Logger.Error("gRPC服务启动失败: " + err.Error())
	}

	// 启动服务
	fmt.Printf(`
	╔═══════════════════════════════════════════════════════════════╗
	║                                                               ║
	║     🤖 AI 自动化运维管理系统 启动成功!                        ║
	║                                                               ║
	║     HTTP:   http://localhost:%s                               ║
	║     gRPC:   localhost:%s                                      ║
	║     WebSocket: ws://localhost:%s/ws                           ║
	║                                                               ║
	║     功能模块:                                                 ║
	║     ✅ 服务器管理      ✅ Agent监控                           ║
	║     ✅ AI智能分析      ✅ 自动修复                            ║
	║     ✅ 异常预测        ✅ 安全防护                            ║
	║     ✅ 自动巡检        ✅ 自愈系统                            ║
	║                                                               ║
	╚═══════════════════════════════════════════════════════════════╝
	`, config.CONFIG.System.Port, config.CONFIG.System.GrpcPort, config.CONFIG.System.Port)

	r.Run(":" + config.CONFIG.System.Port)
}

// initServices 初始化服务
func initServices() {
	// 初始化 LLM 客户端
	var llmClient *llm.GLM5Client
	if config.CONFIG.AI.Enabled && config.CONFIG.AI.APIKey != "" {
		llmClient = llm.NewGLM5Client(llm.GLM5Config{
			APIKey:      config.CONFIG.AI.APIKey,
			BaseURL:     config.CONFIG.AI.BaseURL,
			Model:       config.CONFIG.AI.Model,
			MaxTokens:   config.CONFIG.AI.MaxTokens,
			Temperature: config.CONFIG.AI.Temperature,
		})
	}

	// 初始化通知器
	notifier := notify.NewMultiNotifier(notify.NotifyConfig{
		TelegramEnabled:  config.CONFIG.Notify.Telegram.Enabled,
		TelegramToken:    config.CONFIG.Notify.Telegram.Token,
		TelegramChatID:   config.CONFIG.Notify.Telegram.ChatID,
		WeChatEnabled:    config.CONFIG.Notify.WeChat.Enabled,
		WeChatWebhook:    config.CONFIG.Notify.WeChat.Webhook,
		DingTalkEnabled:  config.CONFIG.Notify.DingTalk.Enabled,
		DingTalkWebhook:  config.CONFIG.Notify.DingTalk.Webhook,
	})

	// 初始化巡检机器人
	patrolRobot := patrol.NewPatrolRobot()
	patrolRobot.SetNotifier(notifier)

	// 初始化自愈系统
	healer := selfheal.NewSelfHealer()
	healer.SetNotifier(notifier)

	// 初始化预测器
	var predictor *prediction.Predictor
	if llmClient != nil {
		predictor = prediction.NewPredictor(llmClient)
	}

	// 初始化工作流引擎
	workflowEngine := workflow.NewWorkflowEngine()
	if llmClient != nil {
		workflowEngine.SetLLMClient(llmClient)
	}
	workflowEngine.SetNotifier(notifier)

	// 初始化调度器
	sched := scheduler.NewScheduler()
	sched.SetPatrolRobot(patrolRobot)
	sched.SetHealer(healer)
	sched.SetPredictor(predictor)

	// 启动调度器
	if config.CONFIG.Scheduler.Enabled {
		if err := sched.Start(); err != nil {
			global.Logger.Error("调度器启动失败: " + err.Error())
		}
	}

	// 启动自愈监控
	if config.CONFIG.SelfHeal.Enabled {
		go healer.MonitorAndHeal()
	}

	global.Logger.Info("所有服务初始化完成")
}
