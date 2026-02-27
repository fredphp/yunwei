package global

import (
        "fmt"

        "yunwei/config"
        "yunwei/model/server"
        "yunwei/model/system"
        "yunwei/service/detector"
        "yunwei/service/executor"
        "yunwei/service/notify"
        "yunwei/service/patrol"
        "yunwei/service/prediction"
        "yunwei/service/scheduler"
        "yunwei/service/security"
        "yunwei/service/selfheal"
        "yunwei/service/workflow"

        "go.uber.org/zap"
        "gorm.io/driver/mysql"
        "gorm.io/gorm"
)

var (
        DB     *gorm.DB
        Logger *zap.Logger
)

// InitDB 初始化数据库
func InitDB() {
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
                config.CONFIG.Mysql.Username,
                config.CONFIG.Mysql.Password,
                config.CONFIG.Mysql.Host,
                config.CONFIG.Mysql.Port,
                config.CONFIG.Mysql.Database,
        )

        var err error
        DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
        if err != nil {
                panic("数据库连接失败: " + err.Error())
        }

        // 自动迁移数据库表
        AutoMigrate()
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() {
        err := DB.AutoMigrate(
                // 系统用户相关
                &system.SysUser{},
                &system.SysRole{},
                &system.SysMenu{},
                &system.SysApi{},

                // 服务器相关
                &server.Server{},
                &server.Group{},
                &server.ServerMetric{},
                &server.ServerLog{},
                &server.DockerContainer{},
                &server.PortInfo{},

                // 告警相关
                &detector.Alert{},
                &detector.DetectRule{},

                // 执行相关
                &executor.ExecutionRecord{},
                &executor.AutoAction{},

                // 安全相关
                &security.CommandWhitelist{},
                &security.CommandBlacklist{},
                &security.AuditLog{},
                &security.SecurityEvent{},
                &security.IPBlacklist{},
                &security.IPWhitelist{},
                &security.LoginRecord{},
                &security.SecurityRule{},

                // 自愈相关
                &selfheal.HealAction{},
                &selfheal.ServiceHealth{},
                &selfheal.HealRule{},

                // 预测相关
                &prediction.PredictionResult{},
                &prediction.AnomalyDetection{},
                &prediction.AutoScaleRecommendation{},

                // 巡检相关
                &patrol.PatrolRecord{},

                // 通知相关
                &notify.NotifyRecord{},

                // 工作流相关
                &workflow.WorkflowRecord{},

                // 调度相关
                &scheduler.Job{},
                &scheduler.JobLog{},
        )

        if err != nil {
                Logger.Error("数据库迁移失败: " + err.Error())
        } else {
                Logger.Info("数据库迁移完成")
        }
}

// InitLogger 初始化日志
func InitLogger() {
        var err error
        Logger, err = zap.NewProduction()
        if err != nil {
                panic("日志初始化失败: " + err.Error())
        }
}
