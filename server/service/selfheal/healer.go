package selfheal

import (
        "encoding/json"
        "fmt"
        "strings"
        "time"

        "yunwei/global"
        "yunwei/model/notify"
        "yunwei/model/server"
)

// HealStatus 自愈状态
type HealStatus string

const (
        HealStatusPending   HealStatus = "pending"
        HealStatusRunning   HealStatus = "running"
        HealStatusSuccess   HealStatus = "success"
        HealStatusFailed    HealStatus = "failed"
        HealStatusTimeout   HealStatus = "timeout"
)

// ServiceType 服务类型
type ServiceType string

const (
        ServiceNginx   ServiceType = "nginx"
        ServiceDocker  ServiceType = "docker"
        ServiceRedis   ServiceType = "redis"
        ServiceMySQL   ServiceType = "mysql"
        ServicePHP     ServiceType = "php-fpm"
        ServiceCustom  ServiceType = "custom"
)

// HealAction 自愈动作
type HealAction struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`

        ServerID    uint           `json:"serverId" gorm:"index"`
        Server      *server.Server `json:"server" gorm:"foreignKey:ServerID"`
        
        // 服务信息
        ServiceType ServiceType `json:"serviceType" gorm:"type:varchar(32)"`
        ServiceName string      `json:"serviceName" gorm:"type:varchar(64)"`
        
        // 问题信息
        ProblemType  string `json:"problemType" gorm:"type:varchar(64)"` // crashed, timeout, memory_leak, etc.
        ProblemDesc  string `json:"problemDesc" gorm:"type:text"`
        DetectedAt   time.Time `json:"detectedAt"`
        
        // 自愈信息
        Status       HealStatus `json:"status" gorm:"type:varchar(16)"`
        ActionType   string     `json:"actionType" gorm:"type:varchar(32)"` // restart, reload, scale, clean
        Command      string     `json:"command" gorm:"type:text"`
        Output       string     `json:"output" gorm:"type:text"`
        
        // 结果
        Success      bool   `json:"success"`
        RetryCount   int    `json:"retryCount"`
        MaxRetry     int    `json:"maxRetry"`
        ErrorMessage string `json:"errorMessage" gorm:"type:text"`
        
        // 时间
        StartedAt   *time.Time `json:"startedAt"`
        CompletedAt *time.Time `json:"completedAt"`
        Duration    int64      `json:"duration"` // 毫秒
}

func (HealAction) TableName() string {
        return "heal_actions"
}

// ServiceHealth 服务健康状态
type ServiceHealth struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        
        ServerID    uint           `json:"serverId" gorm:"index"`
        Server      *server.Server `json:"server" gorm:"foreignKey:ServerID"`
        
        ServiceType ServiceType `json:"serviceType" gorm:"type:varchar(32)"`
        ServiceName string      `json:"serviceName" gorm:"type:varchar(64)"`
        
        // 状态
        Status      string `json:"status" gorm:"type:varchar(16)"` // running, stopped, crashed, unknown
        IsHealthy   bool   `json:"isHealthy"`
        
        // 进程信息
        PID         int    `json:"pid"`
        CPUUsage    float64 `json:"cpuUsage"`
        MemoryUsage float64 `json:"memoryUsage"`
        Uptime      int64   `json:"uptime"` // 秒
        
        // 检测
        LastCheck   time.Time `json:"lastCheck"`
        CheckCount  int       `json:"checkCount"`
        FailCount   int       `json:"failCount"` // 连续失败次数
}

func (ServiceHealth) TableName() string {
        return "service_health"
}

// HealRule 自愈规则
type HealRule struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`
        
        Name        string      `json:"name" gorm:"type:varchar(64)"`
        ServiceType ServiceType `json:"serviceType" gorm:"type:varchar(32)"`
        Enabled     bool        `json:"enabled" gorm:"default:true"`
        
        // 触发条件
        ProblemTypes []string `json:"problemTypes" gorm:"type:text"` // JSON数组
        FailThreshold int      `json:"failThreshold"` // 连续失败多少次触发
        CheckInterval int      `json:"checkInterval"` // 检查间隔(秒)
        
        // 自愈动作
        ActionType   string `json:"actionType" gorm:"type:varchar(32)"`
        Commands     []string `json:"commands" gorm:"type:text"` // JSON数组
        PreCheck     string `json:"preCheck" gorm:"type:text"` // 执行前检查命令
        PostCheck    string `json:"postCheck" gorm:"type:text"` // 执行后检查命令
        
        // 重试配置
        MaxRetry     int  `json:"maxRetry"`
        RetryDelay   int  `json:"retryDelay"` // 重试延迟(秒)
        CooldownTime int  `json:"cooldownTime"` // 冷却时间(秒)，防止频繁自愈
        
        // 通知
        NotifyOnHeal bool `json:"notifyOnHeal"`
        NotifyOnFail bool `json:"notifyOnFail"`
}

func (HealRule) TableName() string {
        return "heal_rules"
}

// SelfHealer 自愈系统
type SelfHealer struct {
        rules    []HealRule
        notifier notify.Notifier
        executor CommandExecutor
}

// CommandExecutor 命令执行器接口
type CommandExecutor interface {
        Execute(serverID uint, command string) (string, error)
}

// NewSelfHealer 创建自愈系统
func NewSelfHealer() *SelfHealer {
        return &SelfHealer{
                rules: GetDefaultHealRules(),
        }
}

// SetNotifier 设置通知器
func (h *SelfHealer) SetNotifier(notifier notify.Notifier) {
        h.notifier = notifier
}

// SetExecutor 设置执行器
func (h *SelfHealer) SetExecutor(executor CommandExecutor) {
        h.executor = executor
}

// GetDefaultHealRules 获取默认自愈规则
func GetDefaultHealRules() []HealRule {
        return []HealRule{
                // Nginx 自愈规则
                {
                        Name:         "Nginx进程异常自动重启",
                        ServiceType:  ServiceNginx,
                        Enabled:      true,
                        ProblemTypes: []string{"crashed", "stopped", "timeout"},
                        FailThreshold: 1,
                        CheckInterval: 10,
                        ActionType:   "restart",
                        Commands:     []string{"systemctl restart nginx"},
                        PreCheck:     "systemctl is-active nginx",
                        PostCheck:    "systemctl is-active nginx",
                        MaxRetry:     3,
                        RetryDelay:   5,
                        CooldownTime: 60,
                        NotifyOnHeal: true,
                        NotifyOnFail: true,
                },
                // Docker 自愈规则
                {
                        Name:         "Docker服务异常自动重启",
                        ServiceType:  ServiceDocker,
                        Enabled:      true,
                        ProblemTypes: []string{"crashed", "stopped"},
                        FailThreshold: 2,
                        CheckInterval: 30,
                        ActionType:   "restart",
                        Commands:     []string{"systemctl restart docker"},
                        PreCheck:     "systemctl is-active docker",
                        PostCheck:    "docker ps",
                        MaxRetry:     2,
                        RetryDelay:   10,
                        CooldownTime: 300,
                        NotifyOnHeal: true,
                        NotifyOnFail: true,
                },
                // Docker 容器自愈规则
                {
                        Name:         "Docker容器异常自动重启",
                        ServiceType:  ServiceDocker,
                        Enabled:      true,
                        ProblemTypes: []string{"container_crashed", "container_exited"},
                        FailThreshold: 1,
                        CheckInterval: 15,
                        ActionType:   "restart",
                        Commands:     []string{"docker ps -a --filter 'status=exited' -q | xargs -r docker start"},
                        PostCheck:    "docker ps",
                        MaxRetry:     3,
                        RetryDelay:   5,
                        CooldownTime: 60,
                        NotifyOnHeal: true,
                },
                // Redis 自愈规则
                {
                        Name:         "Redis服务异常自动重启",
                        ServiceType:  ServiceRedis,
                        Enabled:      true,
                        ProblemTypes: []string{"crashed", "stopped", "timeout"},
                        FailThreshold: 2,
                        CheckInterval: 15,
                        ActionType:   "restart",
                        Commands:     []string{"systemctl restart redis"},
                        PreCheck:     "systemctl is-active redis",
                        PostCheck:    "redis-cli ping",
                        MaxRetry:     3,
                        RetryDelay:   5,
                        CooldownTime: 120,
                        NotifyOnHeal: true,
                        NotifyOnFail: true,
                },
                // MySQL 自愈规则
                {
                        Name:         "MySQL服务异常自动重启",
                        ServiceType:  ServiceMySQL,
                        Enabled:      true,
                        ProblemTypes: []string{"crashed", "stopped"},
                        FailThreshold: 2,
                        CheckInterval: 30,
                        ActionType:   "restart",
                        Commands:     []string{"systemctl restart mysqld"},
                        PreCheck:     "systemctl is-active mysqld",
                        PostCheck:    "mysqladmin ping",
                        MaxRetry:     2,
                        RetryDelay:   10,
                        CooldownTime: 300,
                        NotifyOnHeal: true,
                        NotifyOnFail: true,
                },
                // PHP-FPM 自愈规则
                {
                        Name:         "PHP-FPM服务异常自动重启",
                        ServiceType:  ServicePHP,
                        Enabled:      true,
                        ProblemTypes: []string{"crashed", "stopped", "timeout"},
                        FailThreshold: 2,
                        CheckInterval: 15,
                        ActionType:   "restart",
                        Commands:     []string{"systemctl restart php-fpm"},
                        PreCheck:     "systemctl is-active php-fpm",
                        PostCheck:    "systemctl is-active php-fpm",
                        MaxRetry:     3,
                        RetryDelay:   5,
                        CooldownTime: 60,
                        NotifyOnHeal: true,
                },
        }
}

// CheckServiceHealth 检查服务健康状态
func (h *SelfHealer) CheckServiceHealth(srv *server.Server, serviceType ServiceType) (*ServiceHealth, error) {
        health := &ServiceHealth{
                ServerID:    srv.ID,
                ServiceType: serviceType,
                ServiceName: string(serviceType),
                LastCheck:   time.Now(),
        }

        if h.executor == nil {
                health.Status = "unknown"
                health.IsHealthy = false
                return health, fmt.Errorf("执行器未配置")
        }

        // 获取服务状态命令
        var checkCmd string
        switch serviceType {
        case ServiceNginx:
                checkCmd = "systemctl is-active nginx && pgrep nginx | wc -l"
        case ServiceDocker:
                checkCmd = "systemctl is-active docker && docker info > /dev/null 2>&1 && echo ok"
        case ServiceRedis:
                checkCmd = "systemctl is-active redis && redis-cli ping"
        case ServiceMySQL:
                checkCmd = "systemctl is-active mysqld && mysqladmin ping"
        case ServicePHP:
                checkCmd = "systemctl is-active php-fpm && pgrep php-fpm | wc -l"
        default:
                checkCmd = fmt.Sprintf("systemctl is-active %s", serviceType)
        }

        output, err := h.executor.Execute(srv.ID, checkCmd)
        if err != nil {
                health.Status = "crashed"
                health.IsHealthy = false
                health.FailCount++
                return health, err
        }

        // 解析输出
        output = strings.TrimSpace(output)
        if strings.Contains(output, "active") || strings.Contains(output, "ok") || strings.Contains(output, "PONG") {
                health.Status = "running"
                health.IsHealthy = true
                health.FailCount = 0
        } else {
                health.Status = "stopped"
                health.IsHealthy = false
                health.FailCount++
        }

        // 保存健康状态
        global.DB.Create(health)

        return health, nil
}

// TriggerHeal 触发自愈
func (h *SelfHealer) TriggerHeal(srv *server.Server, serviceType ServiceType, problemType, problemDesc string) (*HealAction, error) {
        // 查找匹配的规则
        var matchedRule *HealRule
        for _, rule := range h.rules {
                if rule.ServiceType == serviceType && rule.Enabled {
                        for _, pt := range rule.ProblemTypes {
                                if pt == problemType {
                                        matchedRule = &rule
                                        break
                                }
                        }
                }
        }

        if matchedRule == nil {
                return nil, fmt.Errorf("未找到匹配的自愈规则")
        }

        // 检查冷却时间
        var lastHeal HealAction
        result := global.DB.Where("server_id = ? AND service_type = ? AND status = ?",
                srv.ID, serviceType, HealStatusSuccess).
                Order("created_at DESC").
                First(&lastHeal)

        if result.Error == nil {
                cooldownEnd := lastHeal.CreatedAt.Add(time.Duration(matchedRule.CooldownTime) * time.Second)
                if time.Now().Before(cooldownEnd) {
                        return nil, fmt.Errorf("自愈冷却中，请等待至 %s", cooldownEnd.Format("15:04:05"))
                }
        }

        // 创建自愈动作
        action := &HealAction{
                ServerID:     srv.ID,
                ServiceType:  serviceType,
                ServiceName:  string(serviceType),
                ProblemType:  problemType,
                ProblemDesc:  problemDesc,
                DetectedAt:   time.Now(),
                Status:       HealStatusPending,
                ActionType:   matchedRule.ActionType,
                MaxRetry:     matchedRule.MaxRetry,
        }

        // 序列化命令
        commandsJSON, _ := json.Marshal(matchedRule.Commands)
        action.Command = string(commandsJSON)

        global.DB.Create(action)

        // 执行自愈
        h.executeHeal(action, matchedRule)

        return action, nil
}

// executeHeal 执行自愈
func (h *SelfHealer) executeHeal(action *HealAction, rule *HealRule) {
        action.Status = HealStatusRunning
        now := time.Now()
        action.StartedAt = &now
        global.DB.Save(action)

        var commands []string
        json.Unmarshal([]byte(action.Command), &commands)

        var allOutput []string
        success := true

        for retry := 0; retry <= action.MaxRetry; retry++ {
                action.RetryCount = retry
                allOutput = append(allOutput, fmt.Sprintf("=== 第%d次尝试 ===", retry+1))

                // 执行前检查
                if rule.PreCheck != "" && h.executor != nil {
                        output, err := h.executor.Execute(action.ServerID, rule.PreCheck)
                        allOutput = append(allOutput, fmt.Sprintf("[PreCheck] %s", output))
                        if err == nil && (strings.Contains(output, "active") || strings.Contains(output, "ok")) {
                                allOutput = append(allOutput, "服务已正常运行，跳过自愈")
                                success = true
                                break
                        }
                }

                // 执行命令
                for _, cmd := range commands {
                        if h.executor == nil {
                                allOutput = append(allOutput, fmt.Sprintf("[ERROR] 执行器未配置: %s", cmd))
                                success = false
                                continue
                        }

                        output, err := h.executor.Execute(action.ServerID, cmd)
                        if err != nil {
                                allOutput = append(allOutput, fmt.Sprintf("[ERROR] %s: %s", cmd, err.Error()))
                                success = false
                        } else {
                                allOutput = append(allOutput, fmt.Sprintf("[OK] %s: %s", cmd, output))
                        }
                }

                // 执行后检查
                if rule.PostCheck != "" && h.executor != nil {
                        time.Sleep(time.Duration(rule.RetryDelay) * time.Second)
                        output, err := h.executor.Execute(action.ServerID, rule.PostCheck)
                        allOutput = append(allOutput, fmt.Sprintf("[PostCheck] %s", output))

                        if err == nil && (strings.Contains(output, "active") || strings.Contains(output, "ok") || strings.Contains(output, "PONG")) {
                                allOutput = append(allOutput, "✅ 自愈成功")
                                success = true
                                break
                        }
                }

                // 重试延迟
                if retry < action.MaxRetry {
                        time.Sleep(time.Duration(rule.RetryDelay) * time.Second)
                }
        }

        // 更新状态
        completedAt := time.Now()
        action.CompletedAt = &completedAt
        action.Duration = completedAt.Sub(*action.StartedAt).Milliseconds()
        action.Output = strings.Join(allOutput, "\n")
        action.Success = success

        if success {
                action.Status = HealStatusSuccess
        } else {
                action.Status = HealStatusFailed
                action.ErrorMessage = "自愈失败，已达到最大重试次数"
        }

        global.DB.Save(action)

        // 发送通知
        if h.notifier != nil {
                if success && rule.NotifyOnHeal {
                        h.notifier.SendMessage("自愈成功", fmt.Sprintf("服务器 %s 的 %s 服务已自动恢复", action.Server.Name, action.ServiceName))
                } else if !success && rule.NotifyOnFail {
                        h.notifier.SendMessage("自愈失败", fmt.Sprintf("服务器 %s 的 %s 服务自愈失败，请人工介入", action.Server.Name, action.ServiceName))
                }
        }
}

// MonitorAndHeal 监控并自愈
func (h *SelfHealer) MonitorAndHeal() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
                // 获取所有在线服务器
                var servers []server.Server
                global.DB.Where("agent_online = ?", true).Find(&servers)

                for _, srv := range servers {
                        // 检查各服务
                        serviceTypes := []ServiceType{ServiceNginx, ServiceDocker, ServiceRedis, ServiceMySQL, ServicePHP}

                        for _, st := range serviceTypes {
                                health, err := h.CheckServiceHealth(&srv, st)
                                if err != nil || !health.IsHealthy {
                                        // 检查是否需要触发自愈
                                        h.checkAndTriggerHeal(&srv, st, health)
                                }
                        }
                }
        }
}

// checkAndTriggerHeal 检查并触发自愈
func (h *SelfHealer) checkAndTriggerHeal(srv *server.Server, serviceType ServiceType, health *ServiceHealth) {
        // 查找匹配规则
        for _, rule := range h.rules {
                if rule.ServiceType != serviceType || !rule.Enabled {
                        continue
                }

                // 检查失败次数是否达到阈值
                if health.FailCount >= rule.FailThreshold {
                        problemType := "crashed"
                        if health.Status == "stopped" {
                                problemType = "stopped"
                        } else if health.Status == "timeout" {
                                problemType = "timeout"
                        }

                        problemDesc := fmt.Sprintf("服务 %s %s，连续失败 %d 次", serviceType, health.Status, health.FailCount)
                        h.TriggerHeal(srv, serviceType, problemType, problemDesc)
                        break
                }
        }
}

// GetHealHistory 获取自愈历史
func (h *SelfHealer) GetHealHistory(serverID uint, limit int) ([]HealAction, error) {
        var actions []HealAction
        query := global.DB.Model(&HealAction{}).Order("created_at DESC")
        if serverID > 0 {
                query = query.Where("server_id = ?", serverID)
        }
        if limit > 0 {
                query = query.Limit(limit)
        }
        err := query.Find(&actions).Error
        return actions, err
}

// GetServiceHealthHistory 获取服务健康历史
func (h *SelfHealer) GetServiceHealthHistory(serverID uint, limit int) ([]ServiceHealth, error) {
        var healths []ServiceHealth
        query := global.DB.Model(&ServiceHealth{}).Order("created_at DESC")
        if serverID > 0 {
                query = query.Where("server_id = ?", serverID)
        }
        if limit > 0 {
                query = query.Limit(limit)
        }
        err := query.Find(&healths).Error
        return healths, err
}

// ForceHeal 强制自愈
func (h *SelfHealer) ForceHeal(serverID uint, serviceType ServiceType) (*HealAction, error) {
        var srv server.Server
        if err := global.DB.First(&srv, serverID).Error; err != nil {
                return nil, fmt.Errorf("服务器不存在")
        }

        return h.TriggerHeal(&srv, serviceType, "manual", "手动触发自愈")
}
