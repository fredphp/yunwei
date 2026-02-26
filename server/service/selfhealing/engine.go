package selfhealing

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"yunwei/global"
	"yunwei/model/server"
	"yunwei/service/notifier"
)

// ServiceType 服务类型
type ServiceType string

const (
	ServiceNginx   ServiceType = "nginx"
	ServiceDocker  ServiceType = "docker"
	ServiceRedis   ServiceType = "redis"
	ServiceMySQL   ServiceType = "mysql"
	ServicePHP     ServiceType = "php-fpm"
	ServiceJava    ServiceType = "java"
	ServiceNode    ServiceType = "node"
	ServiceCustom  ServiceType = "custom"
)

// HealAction 自愈动作
type HealAction string

const (
	HealActionRestart    HealAction = "restart"
	HealActionReload     HealAction = "reload"
	HealActionKill       HealAction = "kill"
	HealActionClearCache HealAction = "clear_cache"
	HealActionCleanLog   HealAction = "clean_log"
	HealActionScaleUp    HealAction = "scale_up"
	HealActionScaleDown  HealAction = "scale_down"
)

// HealStatus 自愈状态
type HealStatus string

const (
	HealStatusPending   HealStatus = "pending"
	HealStatusRunning   HealStatus = "running"
	HealStatusSuccess   HealStatus = "success"
	HealStatusFailed    HealStatus = "failed"
	HealStatusSkipped   HealStatus = "skipped"
)

// ServiceRule 服务规则
type ServiceRule struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string      `json:"name" gorm:"type:varchar(64)"`
	ServiceType ServiceType `json:"serviceType" gorm:"type:varchar(32)"`
	Enabled     bool        `json:"enabled" gorm:"default:true"`

	// 检测规则
	CheckCommand   string `json:"checkCommand" gorm:"type:text"`   // 检测命令
	CheckInterval  int    `json:"checkInterval"`  // 检测间隔(秒)
	MaxRetries     int    `json:"maxRetries"`     // 最大重试次数
	RetryInterval  int    `json:"retryInterval"`  // 重试间隔(秒)

	// 自愈规则
	AutoHeal       bool       `json:"autoHeal"`
	HealAction     HealAction `json:"healAction" gorm:"type:varchar(32)"`
	HealCommand    string     `json:"healCommand" gorm:"type:text"`
	HealTimeout    int        `json:"healTimeout"` // 超时时间(秒)

	// 通知
	NotifyOnHeal   bool `json:"notifyOnHeal"`
	NotifyOnFail   bool `json:"notifyOnFail"`

	// 限制
	MaxHealPerHour int `json:"maxHealPerHour"` // 每小时最大自愈次数
}

func (ServiceRule) TableName() string {
	return "service_rules"
}

// HealRecord 自愈记录
type HealRecord struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	ServerID    uint           `json:"serverId" gorm:"index"`
	Server      *server.Server `json:"server" gorm:"foreignKey:ServerID"`
	RuleID      uint           `json:"ruleId" gorm:"index"`

	ServiceType ServiceType `json:"serviceType" gorm:"type:varchar(32)"`
	ServiceName string      `json:"serviceName" gorm:"type:varchar(64)"`

	// 问题信息
	IssueType   string `json:"issueType" gorm:"type:varchar(32)"`
	IssueDetail string `json:"issueDetail" gorm:"type:text"`

	// 自愈信息
	Action      HealAction `json:"action" gorm:"type:varchar(32)"`
	Command     string     `json:"command" gorm:"type:text"`
	Status      HealStatus `json:"status" gorm:"type:varchar(16)"`

	// 执行结果
	Output      string     `json:"output" gorm:"type:text"`
	Error       string     `json:"error" gorm:"type:text"`
	Duration    int64      `json:"duration"` // 毫秒
	RetryCount  int        `json:"retryCount"`

	// 通知
	Notified    bool       `json:"notified"`
}

func (HealRecord) TableName() string {
	return "heal_records"
}

// SelfHealingEngine 自愈引擎
type SelfHealingEngine struct {
	rules      []ServiceRule
	notifier   *notifier.NotifierService
	executor   CommandExecutor
	healCounts map[uint]map[time.Time]int // 每小时自愈计数
}

// CommandExecutor 命令执行器接口
type CommandExecutor interface {
	Execute(serverID uint, command string) (string, error)
}

// NewSelfHealingEngine 创建自愈引擎
func NewSelfHealingEngine() *SelfHealingEngine {
	return &SelfHealingEngine{
		rules:      getDefaultRules(),
		notifier:   notifier.NewNotifierService(),
		healCounts: make(map[uint]map[time.Time]int),
	}
}

// SetExecutor 设置命令执行器
func (e *SelfHealingEngine) SetExecutor(executor CommandExecutor) {
	e.executor = executor
}

// SetNotifier 设置通知服务
func (e *SelfHealingEngine) SetNotifier(n *notifier.NotifierService) {
	e.notifier = n
}

// getDefaultRules 获取默认规则
func getDefaultRules() []ServiceRule {
	return []ServiceRule{
		{
			Name:          "Nginx自动重启",
			ServiceType:   ServiceNginx,
			Enabled:       true,
			CheckCommand:  "systemctl is-active nginx",
			CheckInterval: 30,
			MaxRetries:    3,
			RetryInterval: 10,
			AutoHeal:      true,
			HealAction:    HealActionRestart,
			HealCommand:   "systemctl restart nginx",
			HealTimeout:   30,
			NotifyOnHeal:  true,
			NotifyOnFail:  true,
			MaxHealPerHour: 5,
		},
		{
			Name:          "Docker自动重启",
			ServiceType:   ServiceDocker,
			Enabled:       true,
			CheckCommand:  "systemctl is-active docker",
			CheckInterval: 30,
			MaxRetries:    3,
			RetryInterval: 10,
			AutoHeal:      true,
			HealAction:    HealActionRestart,
			HealCommand:   "systemctl restart docker",
			HealTimeout:   60,
			NotifyOnHeal:  true,
			NotifyOnFail:  true,
			MaxHealPerHour: 3,
		},
		{
			Name:          "Redis自动重启",
			ServiceType:   ServiceRedis,
			Enabled:       true,
			CheckCommand:  "redis-cli ping",
			CheckInterval: 30,
			MaxRetries:    3,
			RetryInterval: 5,
			AutoHeal:      true,
			HealAction:    HealActionRestart,
			HealCommand:   "systemctl restart redis",
			HealTimeout:   30,
			NotifyOnHeal:  true,
			NotifyOnFail:  true,
			MaxHealPerHour: 5,
		},
		{
			Name:          "MySQL自动重启",
			ServiceType:   ServiceMySQL,
			Enabled:       true,
			CheckCommand:  "systemctl is-active mysqld",
			CheckInterval: 60,
			MaxRetries:    2,
			RetryInterval: 30,
			AutoHeal:      true,
			HealAction:    HealActionRestart,
			HealCommand:   "systemctl restart mysqld",
			HealTimeout:   120,
			NotifyOnHeal:  true,
			NotifyOnFail:  true,
			MaxHealPerHour: 2,
		},
		{
			Name:          "PHP-FPM自动重启",
			ServiceType:   ServicePHP,
			Enabled:       true,
			CheckCommand:  "systemctl is-active php-fpm",
			CheckInterval: 30,
			MaxRetries:    3,
			RetryInterval: 10,
			AutoHeal:      true,
			HealAction:    HealActionRestart,
			HealCommand:   "systemctl restart php-fpm",
			HealTimeout:   30,
			NotifyOnHeal:  true,
			NotifyOnFail:  true,
			MaxHealPerHour: 5,
		},
	}
}

// CheckService 检查服务状态
func (e *SelfHealingEngine) CheckService(serverID uint, rule ServiceRule) (bool, string, error) {
	if e.executor == nil {
		return false, "", fmt.Errorf("命令执行器未设置")
	}

	output, err := e.executor.Execute(serverID, rule.CheckCommand)
	if err != nil {
		return false, output, err
	}

	// 判断服务是否正常
	output = strings.TrimSpace(output)
	switch rule.ServiceType {
	case ServiceNginx, ServiceDocker, ServiceMySQL, ServicePHP:
		return output == "active", output, nil
	case ServiceRedis:
		return strings.Contains(output, "PONG") || output == "active", output, nil
	default:
		return output != "", output, nil
	}
}

// HealService 自愈服务
func (e *SelfHealingEngine) HealService(serverID uint, rule ServiceRule, issueDetail string) (*HealRecord, error) {
	record := &HealRecord{
		ServerID:    serverID,
		RuleID:      rule.ID,
		ServiceType: rule.ServiceType,
		ServiceName: rule.Name,
		IssueDetail: issueDetail,
		Action:      rule.HealAction,
		Command:     rule.HealCommand,
		Status:      HealStatusPending,
		RetryCount:  0,
	}

	// 检查自愈次数限制
	if !e.canHeal(serverID, rule) {
		record.Status = HealStatusSkipped
		record.Error = "超过每小时最大自愈次数"
		global.DB.Create(record)
		return record, fmt.Errorf("超过每小时最大自愈次数")
	}

	// 检查是否自动自愈
	if !rule.AutoHeal {
		record.Status = HealStatusSkipped
		record.Error = "未启用自动自愈"
		global.DB.Create(record)
		return record, fmt.Errorf("未启用自动自愈")
	}

	// 执行自愈
	record.Status = HealStatusRunning
	global.DB.Create(record)

	startTime := time.Now()
	output, err := e.executeHeal(serverID, rule)
	record.Duration = time.Since(startTime).Milliseconds()
	record.Output = output

	if err != nil {
		record.Status = HealStatusFailed
		record.Error = err.Error()

		// 重试
		for i := 0; i < rule.MaxRetries; i++ {
			record.RetryCount++
			time.Sleep(time.Duration(rule.RetryInterval) * time.Second)

			output, err = e.executeHeal(serverID, rule)
			record.Duration = time.Since(startTime).Milliseconds()
			record.Output = output

			if err == nil {
				record.Status = HealStatusSuccess
				break
			}
		}

		if record.Status != HealStatusSuccess {
			record.Error = err.Error()
		}
	} else {
		record.Status = HealStatusSuccess
	}

	// 更新自愈计数
	e.incrementHealCount(serverID)

	// 保存记录
	global.DB.Save(record)

	// 发送通知
	go e.sendNotification(record, rule)

	return record, nil
}

// executeHeal 执行自愈命令
func (e *SelfHealingEngine) executeHeal(serverID uint, rule ServiceRule) (string, error) {
	if e.executor == nil {
		return "", fmt.Errorf("命令执行器未设置")
	}

	return e.executor.Execute(serverID, rule.HealCommand)
}

// canHeal 检查是否可以自愈
func (e *SelfHealingEngine) canHeal(serverID uint, rule ServiceRule) bool {
	if rule.MaxHealPerHour == 0 {
		return true
	}

	hour := time.Now().Truncate(time.Hour)
	if counts, ok := e.healCounts[serverID]; ok {
		if count, ok := counts[hour]; ok {
			return count < rule.MaxHealPerHour
		}
	}
	return true
}

// incrementHealCount 增加自愈计数
func (e *SelfHealingEngine) incrementHealCount(serverID uint) {
	hour := time.Now().Truncate(time.Hour)
	if e.healCounts[serverID] == nil {
		e.healCounts[serverID] = make(map[time.Time]int)
	}
	e.healCounts[serverID][hour]++
}

// sendNotification 发送通知
func (e *SelfHealingEngine) sendNotification(record *HealRecord, rule ServiceRule) {
	if e.notifier == nil {
		return
	}

	var serverName string
	var srv server.Server
	if global.DB.First(&srv, record.ServerID).Error == nil {
		serverName = srv.Name
	}

	if record.Status == HealStatusSuccess && rule.NotifyOnHeal {
		title := fmt.Sprintf("✅ 服务自愈成功 - %s", record.ServiceName)
		content := fmt.Sprintf(
			"服务器: %s\n服务: %s\n动作: %s\n问题: %s\n耗时: %dms",
			serverName, record.ServiceName, record.Action, record.IssueDetail, record.Duration,
		)
		e.notifier.Broadcast(title, content)
		record.Notified = true
	} else if record.Status == HealStatusFailed && rule.NotifyOnFail {
		title := fmt.Sprintf("❌ 服务自愈失败 - %s", record.ServiceName)
		content := fmt.Sprintf(
			"服务器: %s\n服务: %s\n动作: %s\n问题: %s\n错误: %s\n重试次数: %d",
			serverName, record.ServiceName, record.Action, record.IssueDetail, record.Error, record.RetryCount,
		)
		e.notifier.Broadcast(title, content)
		record.Notified = true
	}

	global.DB.Save(record)
}

// RunCheck 执行服务检查
func (e *SelfHealingEngine) RunCheck(serverID uint) ([]HealRecord, error) {
	var records []HealRecord

	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		healthy, output, err := e.CheckService(serverID, rule)
		if !healthy {
			issueDetail := fmt.Sprintf("服务异常: %s (输出: %s)", err, output)
			record, healErr := e.HealService(serverID, rule, issueDetail)
			if healErr != nil {
				record = &HealRecord{
					ServerID:    serverID,
					ServiceType: rule.ServiceType,
					ServiceName: rule.Name,
					IssueDetail: issueDetail,
					Status:      HealStatusFailed,
					Error:       healErr.Error(),
				}
				global.DB.Create(record)
			}
			records = append(records, *record)
		}
	}

	return records, nil
}

// RunAllChecks 检查所有服务器
func (e *SelfHealingEngine) RunAllChecks() (map[uint][]HealRecord, error) {
	var servers []server.Server
	global.DB.Where("agent_online = ?", true).Find(&servers)

	results := make(map[uint][]HealRecord)
	for _, srv := range servers {
		records, err := e.RunCheck(srv.ID)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("服务器 %d 检查失败: %v", srv.ID, err))
		}
		if len(records) > 0 {
			results[srv.ID] = records
		}
	}

	return results, nil
}

// StartScheduler 启动定时检查
func (e *SelfHealingEngine) StartScheduler() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			e.RunAllChecks()
		}
	}()
}

// GetRecords 获取自愈记录
func (e *SelfHealingEngine) GetRecords(serverID uint, limit int) ([]HealRecord, error) {
	var records []HealRecord
	query := global.DB.Order("created_at DESC")
	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&records).Error
	return records, err
}

// GetRules 获取规则
func (e *SelfHealingEngine) GetRules() []ServiceRule {
	return e.rules
}

// AddRule 添加规则
func (e *SelfHealingEngine) AddRule(rule ServiceRule) {
	e.rules = append(e.rules, rule)
}

// DockerContainerHeal Docker容器自愈
func (e *SelfHealingEngine) DockerContainerHeal(serverID uint, containerName string) (*HealRecord, error) {
	rule := ServiceRule{
		ServiceType:  ServiceDocker,
		HealAction:   HealActionRestart,
		HealCommand:  fmt.Sprintf("docker restart %s", containerName),
		HealTimeout:  60,
		NotifyOnHeal: true,
		NotifyOnFail: true,
	}

	record := &HealRecord{
		ServerID:    serverID,
		ServiceType: ServiceDocker,
		ServiceName: fmt.Sprintf("Docker容器: %s", containerName),
		Action:      HealActionRestart,
		Command:     rule.HealCommand,
		Status:      HealStatusPending,
	}

	if e.executor == nil {
		record.Status = HealStatusFailed
		record.Error = "命令执行器未设置"
		global.DB.Create(record)
		return record, fmt.Errorf("命令执行器未设置")
	}

	record.Status = HealStatusRunning
	global.DB.Create(record)

	startTime := time.Now()
	output, err := e.executor.Execute(serverID, rule.HealCommand)
	record.Duration = time.Since(startTime).Milliseconds()
	record.Output = output

	if err != nil {
		record.Status = HealStatusFailed
		record.Error = err.Error()
	} else {
		record.Status = HealStatusSuccess
	}

	global.DB.Save(record)
	go e.sendNotification(record, rule)

	return record, err
}

// GetStatistics 获取自愈统计
func (e *SelfHealingEngine) GetStatistics(serverID uint, days int) map[string]interface{} {
	startTime := time.Now().AddDate(0, 0, -days)
	stats := make(map[string]interface{})

	// 总自愈次数
	var total int64
	query := global.DB.Model(&HealRecord{}).Where("created_at > ?", startTime)
	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}
	query.Count(&total)
	stats["total"] = total

	// 成功次数
	var success int64
	query = global.DB.Model(&HealRecord{}).Where("status = ? AND created_at > ?", HealStatusSuccess, startTime)
	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}
	query.Count(&success)
	stats["success"] = success

	// 失败次数
	var failed int64
	query = global.DB.Model(&HealRecord{}).Where("status = ? AND created_at > ?", HealStatusFailed, startTime)
	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}
	query.Count(&failed)
	stats["failed"] = failed

	// 成功率
	if total > 0 {
		stats["successRate"] = float64(success) / float64(total) * 100
	} else {
		stats["successRate"] = 0
	}

	// 各服务类型统计
	var byService []struct {
		ServiceType ServiceType `json:"serviceType"`
		Count       int         `json:"count"`
	}
	global.DB.Model(&HealRecord{}).
		Select("service_type, count(*) as count").
		Where("created_at > ?", startTime).
		Group("service_type").
		Scan(&byService)
	stats["byService"] = byService

	return stats
}

// ManualHeal 手动自愈
func (e *SelfHealingEngine) ManualHeal(serverID uint, serviceType ServiceType, command string) (*HealRecord, error) {
	rule := ServiceRule{
		ServiceType:  serviceType,
		HealCommand:  command,
		HealTimeout:  60,
		NotifyOnHeal: true,
		NotifyOnFail: true,
	}

	issueDetail := "手动触发自愈"

	return e.HealService(serverID, rule, issueDetail)
}

// ExportRecords 导出自愈记录
func (e *SelfHealingEngine) ExportRecords(serverID uint, startTime, endTime time.Time) ([]HealRecord, error) {
	var records []HealRecord
	query := global.DB.Where("created_at >= ? AND created_at <= ?", startTime, endTime)
	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}
	err := query.Order("created_at ASC").Find(&records).Error
	return records, err
}

// ToJSON 转换为JSON
func (r *HealRecord) ToJSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}
