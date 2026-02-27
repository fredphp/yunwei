package optimizer

import (
        "fmt"
        "strings"
        "time"

        "yunwei/model/server"
        "yunwei/service/detector"
)

// ActionType 操作类型
type ActionType string

const (
        ActionRestartProcess   ActionType = "restart_process"
        ActionCleanLogs        ActionType = "clean_logs"
        ActionClearCache       ActionType = "clear_cache"
        ActionScaleContainer   ActionType = "scale_container"
        ActionRestartService   ActionType = "restart_service"
        ActionKillProcess      ActionType = "kill_process"
        ActionCleanDocker      ActionType = "clean_docker"
        ActionFreeMemory       ActionType = "free_memory"
        ActionRestartNginx     ActionType = "restart_nginx"
        ActionRestartMySQL     ActionType = "restart_mysql"
        ActionRestartDocker    ActionType = "restart_docker"
)

// ActionStatus 操作状态
type ActionStatus string

const (
        ActionStatusPending   ActionStatus = "pending"
        ActionStatusRunning   ActionStatus = "running"
        ActionStatusSuccess   ActionStatus = "success"
        ActionStatusFailed    ActionStatus = "failed"
        ActionStatusCancelled ActionStatus = "cancelled"
)

// AutoAction 自动操作记录
type AutoAction struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`
        
        ServerID  uint           `json:"serverId" gorm:"index"`
        Server    *server.Server `json:"server" gorm:"foreignKey:ServerID"`
        AlertID   uint           `json:"alertId" gorm:"index"`
        
        // 操作信息
        Type       ActionType  `json:"type" gorm:"type:varchar(32)"`
        Status     ActionStatus `json:"status" gorm:"type:varchar(16);default:'pending'"`
        Command    string      `json:"command" gorm:"type:text"`
        Output     string      `json:"output" gorm:"type:text"`
        Error      string      `json:"error" gorm:"type:text"`
        
        // 执行信息
        ExecutedAt   *time.Time `json:"executedAt"`
        CompletedAt  *time.Time `json:"completedAt"`
        Duration     int64      `json:"duration"` // 毫秒
        
        // AI 决策
        AIDecision   bool       `json:"aiDecision" gorm:"default:false"`
        AIReason     string     `json:"aiReason" gorm:"type:text"`
        ManualConfirm bool      `json:"manualConfirm" gorm:"default:false"`
        ConfirmedBy  uint       `json:"confirmedBy"`
}

func (AutoAction) TableName() string {
        return "auto_actions"
}

// Optimizer 优化器
type Optimizer struct {
        actions map[detector.AlertType]ActionDefinition
}

// ActionDefinition 操作定义
type ActionDefinition struct {
        Type        ActionType
        Name        string
        Description string
        Command     string
        DangerLevel int // 1-5, 5最危险
        RequireConfirm bool
}

// NewOptimizer 创建优化器
func NewOptimizer() *Optimizer {
        return &Optimizer{
                actions: GetDefaultActions(),
        }
}

// GetDefaultActions 获取默认操作
func GetDefaultActions() map[detector.AlertType]ActionDefinition {
        return map[detector.AlertType]ActionDefinition{
                detector.AlertTypeCPUHigh: {
                        Type:        ActionClearCache,
                        Name:        "清理系统缓存",
                        Description: "释放系统缓存降低CPU压力",
                        Command:     "sync && echo 3 > /proc/sys/vm/drop_caches",
                        DangerLevel: 2,
                        RequireConfirm: false,
                },
                detector.AlertTypeMemoryLow: {
                        Type:        ActionFreeMemory,
                        Name:        "释放内存",
                        Description: "清理缓存并释放内存",
                        Command:     "sync && echo 3 > /proc/sys/vm/drop_caches && systemctl restart systemd-journald",
                        DangerLevel: 2,
                        RequireConfirm: false,
                },
                detector.AlertTypeDiskHigh: {
                        Type:        ActionCleanDocker,
                        Name:        "清理Docker和日志",
                        Description: "清理Docker无用资源和系统日志",
                        Command:     "docker system prune -af && journalctl --vacuum-time=3d && find /var/log -type f -name '*.log' -mtime +7 -delete",
                        DangerLevel: 3,
                        RequireConfirm: true,
                },
                detector.AlertTypeLoadHigh: {
                        Type:        ActionKillProcess,
                        Name:        "终止僵死进程",
                        Description: "查找并终止僵尸进程",
                        Command:     "ps aux | grep -w Z | awk '{print $2}' | xargs -r kill -9",
                        DangerLevel: 2,
                        RequireConfirm: false,
                },
                detector.AlertTypeNginxDown: {
                        Type:        ActionRestartNginx,
                        Name:        "重启Nginx",
                        Description: "重启Nginx服务",
                        Command:     "systemctl restart nginx && systemctl status nginx",
                        DangerLevel: 2,
                        RequireConfirm: false,
                },
                detector.AlertTypeMySQLSlow: {
                        Type:        ActionRestartMySQL,
                        Name:        "重启MySQL",
                        Description: "重启MySQL服务",
                        Command:     "systemctl restart mysqld",
                        DangerLevel: 4,
                        RequireConfirm: true,
                },
                detector.AlertTypeDockerDown: {
                        Type:        ActionRestartDocker,
                        Name:        "重启Docker容器",
                        Description: "重启异常的Docker容器",
                        Command:     "docker ps -a --filter 'status=exited' -q | xargs -r docker start",
                        DangerLevel: 2,
                        RequireConfirm: false,
                },
        }
}

// GetAction 获取操作
func (o *Optimizer) GetAction(alertType detector.AlertType) (ActionDefinition, bool) {
        action, ok := o.actions[alertType]
        return action, ok
}

// GenerateAction 生成操作
func (o *Optimizer) GenerateAction(alert detector.DetectionResult, server *server.Server) (*AutoAction, error) {
        actionDef, ok := o.actions[alert.Type]
        if !ok {
                return nil, fmt.Errorf("no action defined for alert type: %s", alert.Type)
        }

        action := &AutoAction{
                ServerID:      server.ID,
                Type:          actionDef.Type,
                Status:        ActionStatusPending,
                Command:       actionDef.Command,
                ManualConfirm: actionDef.RequireConfirm,
        }

        return action, nil
}

// ExecuteAction 执行操作
func (o *Optimizer) ExecuteAction(action *AutoAction, executor CommandExecutor) error {
        action.Status = ActionStatusRunning
        now := time.Now()
        action.ExecutedAt = &now

        output, err := executor.Execute(action.Command)

        completedAt := timePtr(time.Now())
        action.CompletedAt = completedAt
        action.Duration = action.CompletedAt.Sub(*action.ExecutedAt).Milliseconds()

        if err != nil {
                action.Status = ActionStatusFailed
                action.Error = err.Error()
                return err
        }

        action.Status = ActionStatusSuccess
        action.Output = output
        return nil
}

// CommandExecutor 命令执行器接口
type CommandExecutor interface {
        Execute(command string) (string, error)
}

// AIActionGenerator AI操作生成器
type AIActionGenerator struct {
        actions map[string]string
}

// NewAIActionGenerator 创建AI操作生成器
func NewAIActionGenerator() *AIActionGenerator {
        return &AIActionGenerator{
                actions: map[string]string{
                        // CPU 相关
                        "cpu_high_light":    "echo 1 > /proc/sys/vm/drop_caches",
                        "cpu_high_medium":   "echo 2 > /proc/sys/vm/drop_caches",
                        "cpu_high_heavy":    "echo 3 > /proc/sys/vm/drop_caches && pkill -f 'chrome|firefox'",
                        
                        // 内存相关
                        "memory_low_light":  "sync && echo 1 > /proc/sys/vm/drop_caches",
                        "memory_low_medium": "sync && echo 2 > /proc/sys/vm/drop_caches",
                        "memory_low_heavy":  "sync && echo 3 > /proc/sys/vm/drop_caches && systemctl restart rsyslog",
                        
                        // 磁盘相关
                        "disk_high_light":   "docker system prune -f",
                        "disk_high_medium":  "docker system prune -af && journalctl --vacuum-time=7d",
                        "disk_high_heavy":   "docker system prune -af --volumes && journalctl --vacuum-time=1d && find /var/log -type f -delete",
                        
                        // 服务重启
                        "restart_nginx":     "systemctl restart nginx",
                        "restart_mysql":     "systemctl restart mysqld",
                        "restart_redis":     "systemctl restart redis",
                        "restart_docker":    "systemctl restart docker",
                        "restart_php_fpm":   "systemctl restart php-fpm",
                        
                        // 日志清理
                        "clean_logs":        "journalctl --vacuum-time=3d && find /var/log -type f -name '*.gz' -delete",
                        "clean_old_logs":    "find /var/log -type f -mtime +7 -delete",
                        
                        // Docker 操作
                        "docker_prune":      "docker system prune -f",
                        "docker_deep_prune": "docker system prune -af --volumes",
                        "docker_restart_all":"docker restart $(docker ps -q)",
                        
                        // 进程管理
                        "kill_zombies":      "ps aux | grep -w Z | awk '{print $2}' | xargs -r kill -9",
                        "kill_defunct":      "pkill -f 'defunct'",
                },
        }
}

// GetAction 获取AI推荐的操作
func (g *AIActionGenerator) GetAction(key string) (string, bool) {
        cmd, ok := g.actions[key]
        return cmd, ok
}

// ParseAIResponse 解析AI响应，提取可执行命令
func (g *AIActionGenerator) ParseAIResponse(response string) []string {
        var commands []string
        lines := strings.Split(response, "\n")
        
        inCodeBlock := false
        for _, line := range lines {
                line = strings.TrimSpace(line)
                
                // 检测代码块
                if strings.HasPrefix(line, "```") {
                        inCodeBlock = !inCodeBlock
                        continue
                }
                
                // 提取代码块中的命令
                if inCodeBlock && line != "" {
                        // 跳过注释
                        if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "//") {
                                commands = append(commands, line)
                        }
                }
                
                // 提取独立命令行
                if !inCodeBlock {
                        // 匹配以 shell 命令开头的行
                        shellPrefixes := []string{"sudo ", "systemctl ", "docker ", "kill ", "pkill ", 
                                "rm ", "mv ", "cp ", "echo ", "sync ", "find ", "journalctl "}
                        for _, prefix := range shellPrefixes {
                                if strings.HasPrefix(line, prefix) {
                                        commands = append(commands, line)
                                        break
                                }
                        }
                }
        }
        
        return commands
}

// Helper function
func timePtr(t time.Time) *time.Time {
        return &t
}

// GetActionDescription 获取操作描述
func GetActionDescription(actionType ActionType) string {
        descriptions := map[ActionType]string{
                ActionRestartProcess:   "重启异常进程",
                ActionCleanLogs:        "自动清理日志文件",
                ActionClearCache:       "自动释放系统缓存",
                ActionScaleContainer:   "自动扩容Docker容器",
                ActionRestartService:   "自动重启服务",
                ActionKillProcess:      "终止异常进程",
                ActionCleanDocker:      "清理Docker无用资源",
                ActionFreeMemory:       "释放内存资源",
                ActionRestartNginx:     "重启Nginx服务",
                ActionRestartMySQL:     "重启MySQL服务",
                ActionRestartDocker:    "重启Docker服务",
        }
        
        if desc, ok := descriptions[actionType]; ok {
                return desc
        }
        return string(actionType)
}

// IsDangerousAction 判断是否为危险操作
func IsDangerousAction(actionType ActionType) bool {
        dangerousActions := map[ActionType]bool{
                ActionRestartMySQL:    true,
                ActionRestartDocker:   true,
                ActionKillProcess:     true,
                ActionScaleContainer:  true,
                ActionRestartService:  true,
        }
        return dangerousActions[actionType]
}

// RequiresConfirmation 判断是否需要确认
func RequiresConfirmation(actionType ActionType) bool {
        return IsDangerousAction(actionType)
}
