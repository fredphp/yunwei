package detector

import (
	"fmt"
	"time"

	"yunwei/model/server"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
	AlertLevelEmergency AlertLevel = "emergency"
)

// AlertType 告警类型
type AlertType string

const (
	AlertTypeCPUHigh       AlertType = "cpu_high"
	AlertTypeMemoryLow     AlertType = "memory_low"
	AlertTypeDiskHigh      AlertType = "disk_high"
	AlertTypeLoadHigh      AlertType = "load_high"
	AlertTypePortAttack    AlertType = "port_attack"
	AlertTypeNginxDown     AlertType = "nginx_down"
	AlertTypeMySQLSlow     AlertType = "mysql_slow"
	AlertTypeDockerDown    AlertType = "docker_down"
	AlertTypeProcessDown   AlertType = "process_down"
	AlertTypeNetworkAnomaly AlertType = "network_anomaly"
)

// Alert 告警
type Alert struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"createdAt"`
	ServerID    uint           `json:"serverId" gorm:"index"`
	Server      *server.Server `json:"server" gorm:"foreignKey:ServerID"`

	// 告警信息
	Type     AlertType  `json:"type" gorm:"type:varchar(32)"`
	Level    AlertLevel `json:"level" gorm:"type:varchar(16)"`
	Title    string     `json:"title" gorm:"type:varchar(128)"`
	Message  string     `json:"message" gorm:"type:text"`

	// 指标数据
	MetricValue float64 `json:"metricValue"`
	Threshold   float64 `json:"threshold"`

	// 状态
	Status         string     `json:"status" gorm:"type:varchar(16);default:'active'"` // active, acknowledged, resolved
	AcknowledgedBy uint       `json:"acknowledgedBy"`
	ResolvedAt     *time.Time `json:"resolvedAt"`
	ResolvedBy     uint       `json:"resolvedBy"`

	// 处理信息
	AutoResolved bool   `json:"autoResolved" gorm:"default:false"`
	ActionTaken  string `json:"actionTaken" gorm:"type:text"`
}

func (Alert) TableName() string {
	return "alerts"
}

// DetectRule 检测规则
type DetectRule struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 规则信息
	Name   string    `json:"name" gorm:"type:varchar(64);not null"`
	Type   AlertType `json:"type" gorm:"type:varchar(32)"`
	Enabled bool     `json:"enabled" gorm:"default:true"`

	// 阈值
	Threshold float64 `json:"threshold"`
	Duration  int     `json:"duration"` // 持续时间(秒)
	Count     int     `json:"count"`    // 连续次数

	// 告警级别
	Level AlertLevel `json:"level" gorm:"type:varchar(16)"`

	// 自动处理
	AutoAction    bool   `json:"autoAction" gorm:"default:false"`
	ActionCommand string `json:"actionCommand" gorm:"type:text"`

	// 通知
	NotifyEmail   bool   `json:"notifyEmail"`
	NotifySMS     bool   `json:"notifySMS"`
	NotifyWebhook bool   `json:"notifyWebhook"`
	WebhookURL    string `json:"webhookUrl" gorm:"type:varchar(255)"`

	// 描述
	Description string `json:"description" gorm:"type:varchar(255)"`
}

func (DetectRule) TableName() string {
	return "detect_rules"
}

// DetectionResult 检测结果
type DetectionResult struct {
	ServerID    uint
	Type        AlertType
	Level       AlertLevel
	Title       string
	Message     string
	MetricValue float64
	Threshold   float64
	Triggered   bool
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID     int
	Name    string
	CPU     float64
	Memory  float64
	Status  string
	Command string
}

// Detector 检测器
type Detector struct {
	rules []DetectRule
}

// NewDetector 创建检测器
func NewDetector() *Detector {
	return &Detector{
		rules: GetDefaultRules(),
	}
}

// GetDefaultRules 获取默认规则
func GetDefaultRules() []DetectRule {
	return []DetectRule{
		// CPU 告警规则
		{
			Name:          "CPU使用率过高-警告",
			Type:          AlertTypeCPUHigh,
			Enabled:       true,
			Threshold:     80,
			Duration:      60,
			Count:         3,
			Level:         AlertLevelWarning,
			AutoAction:    false,
			Description:   "CPU使用率超过80%持续1分钟",
		},
		{
			Name:          "CPU使用率过高-严重",
			Type:          AlertTypeCPUHigh,
			Enabled:       true,
			Threshold:     90,
			Duration:      30,
			Count:         2,
			Level:         AlertLevelCritical,
			AutoAction:    true,
			ActionCommand: "echo 3 > /proc/sys/vm/drop_caches",
			Description:   "CPU使用率超过90%持续30秒",
		},
		// 内存告警规则
		{
			Name:        "内存不足-警告",
			Type:        AlertTypeMemoryLow,
			Enabled:     true,
			Threshold:   80,
			Duration:    60,
			Count:       3,
			Level:       AlertLevelWarning,
			AutoAction:  false,
			Description: "内存使用率超过80%",
		},
		{
			Name:          "内存不足-严重",
			Type:          AlertTypeMemoryLow,
			Enabled:       true,
			Threshold:     90,
			Duration:      30,
			Count:         2,
			Level:         AlertLevelCritical,
			AutoAction:    true,
			ActionCommand: "sync && echo 3 > /proc/sys/vm/drop_caches",
			Description:   "内存使用率超过90%",
		},
		// 磁盘告警规则
		{
			Name:        "磁盘空间不足-警告",
			Type:        AlertTypeDiskHigh,
			Enabled:     true,
			Threshold:   80,
			Duration:    300,
			Count:       1,
			Level:       AlertLevelWarning,
			AutoAction:  false,
			Description: "磁盘使用率超过80%",
		},
		{
			Name:          "磁盘空间不足-严重",
			Type:          AlertTypeDiskHigh,
			Enabled:       true,
			Threshold:     90,
			Duration:      60,
			Count:         1,
			Level:         AlertLevelCritical,
			AutoAction:    true,
			ActionCommand: "docker system prune -f && journalctl --vacuum-time=3d",
			Description:   "磁盘使用率超过90%",
		},
		// 负载告警规则
		{
			Name:        "系统负载过高",
			Type:        AlertTypeLoadHigh,
			Enabled:     true,
			Threshold:   5,
			Duration:    60,
			Count:       3,
			Level:       AlertLevelWarning,
			AutoAction:  false,
			Description: "系统负载超过5持续1分钟",
		},
		{
			Name:          "系统负载极高",
			Type:          AlertTypeLoadHigh,
			Enabled:       true,
			Threshold:     10,
			Duration:      30,
			Count:         2,
			Level:         AlertLevelCritical,
			AutoAction:    true,
			ActionCommand: "pkill -f 'defunct'",
			Description:   "系统负载超过10持续30秒",
		},
		// Nginx 进程告警
		{
			Name:          "Nginx进程异常",
			Type:          AlertTypeNginxDown,
			Enabled:       true,
			Threshold:     1,
			Duration:      10,
			Count:         1,
			Level:         AlertLevelCritical,
			AutoAction:    true,
			ActionCommand: "systemctl restart nginx",
			Description:   "Nginx进程数量为0",
		},
		// MySQL 慢查询告警
		{
			Name:        "MySQL慢查询过多",
			Type:        AlertTypeMySQLSlow,
			Enabled:     true,
			Threshold:   100,
			Duration:    60,
			Count:       2,
			Level:       AlertLevelWarning,
			AutoAction:  false,
			Description: "每分钟慢查询超过100条",
		},
		// Docker 容器告警
		{
			Name:        "Docker容器异常",
			Type:        AlertTypeDockerDown,
			Enabled:     true,
			Threshold:   1,
			Duration:    30,
			Count:       1,
			Level:       AlertLevelWarning,
			AutoAction:  false,
			Description: "有容器处于非运行状态",
		},
		// 端口攻击告警
		{
			Name:        "端口安全告警",
			Type:        AlertTypePortAttack,
			Enabled:     true,
			Threshold:   3,
			Duration:    10,
			Count:       1,
			Level:       AlertLevelWarning,
			AutoAction:  false,
			Description: "检测到敏感端口暴露",
		},
	}
}

// Detect 执行检测
func (d *Detector) Detect(srv *server.Server, metric *server.ServerMetric, processes []ProcessInfo, containers []server.DockerContainer, ports []server.PortInfo) []DetectionResult {
	var results []DetectionResult

	for _, rule := range d.rules {
		if !rule.Enabled {
			continue
		}

		var result DetectionResult
		result.ServerID = srv.ID
		result.Type = rule.Type
		result.Threshold = rule.Threshold
		result.Level = rule.Level

		switch rule.Type {
		case AlertTypeCPUHigh:
			result = d.detectCPU(rule, srv, metric)
		case AlertTypeMemoryLow:
			result = d.detectMemory(rule, srv, metric)
		case AlertTypeDiskHigh:
			result = d.detectDisk(rule, srv, metric)
		case AlertTypeLoadHigh:
			result = d.detectLoad(rule, srv, metric)
		case AlertTypeNginxDown:
			result = d.detectNginx(rule, srv, processes)
		case AlertTypeDockerDown:
			result = d.detectDocker(rule, srv, containers)
		case AlertTypePortAttack:
			result = d.detectPortAttack(rule, srv, ports)
		}

		if result.Triggered {
			results = append(results, result)
		}
	}

	return results
}

// detectCPU 检测CPU
func (d *Detector) detectCPU(rule DetectRule, srv *server.Server, metric *server.ServerMetric) DetectionResult {
	result := DetectionResult{
		ServerID:    srv.ID,
		Type:        rule.Type,
		Threshold:   rule.Threshold,
		Level:       rule.Level,
		MetricValue: metric.CPUUsage,
	}

	if metric.CPUUsage > rule.Threshold {
		result.Triggered = true
		result.Title = "CPU使用率过高"
		result.Message = fmt.Sprintf("服务器 %s CPU使用率达到 %.2f%%，超过阈值 %.2f%%", srv.Name, metric.CPUUsage, rule.Threshold)
	}

	return result
}

// detectMemory 检测内存
func (d *Detector) detectMemory(rule DetectRule, srv *server.Server, metric *server.ServerMetric) DetectionResult {
	result := DetectionResult{
		ServerID:    srv.ID,
		Type:        rule.Type,
		Threshold:   rule.Threshold,
		Level:       rule.Level,
		MetricValue: metric.MemoryUsage,
	}

	if metric.MemoryUsage > rule.Threshold {
		result.Triggered = true
		result.Title = "内存使用率过高"
		result.Message = fmt.Sprintf("服务器 %s 内存使用率达到 %.2f%%，超过阈值 %.2f%%", srv.Name, metric.MemoryUsage, rule.Threshold)
	}

	return result
}

// detectDisk 检测磁盘
func (d *Detector) detectDisk(rule DetectRule, srv *server.Server, metric *server.ServerMetric) DetectionResult {
	result := DetectionResult{
		ServerID:    srv.ID,
		Type:        rule.Type,
		Threshold:   rule.Threshold,
		Level:       rule.Level,
		MetricValue: metric.DiskUsage,
	}

	if metric.DiskUsage > rule.Threshold {
		result.Triggered = true
		result.Title = "磁盘空间不足"
		result.Message = fmt.Sprintf("服务器 %s 磁盘使用率达到 %.2f%%，超过阈值 %.2f%%", srv.Name, metric.DiskUsage, rule.Threshold)
	}

	return result
}

// detectLoad 检测负载
func (d *Detector) detectLoad(rule DetectRule, srv *server.Server, metric *server.ServerMetric) DetectionResult {
	result := DetectionResult{
		ServerID:    srv.ID,
		Type:        rule.Type,
		Threshold:   rule.Threshold,
		Level:       rule.Level,
		MetricValue: metric.Load1,
	}

	if metric.Load1 > rule.Threshold {
		result.Triggered = true
		result.Title = "系统负载过高"
		result.Message = fmt.Sprintf("服务器 %s 系统负载达到 %.2f，超过阈值 %.2f", srv.Name, metric.Load1, rule.Threshold)
	}

	return result
}

// detectNginx 检测Nginx进程
func (d *Detector) detectNginx(rule DetectRule, srv *server.Server, processes []ProcessInfo) DetectionResult {
	result := DetectionResult{
		ServerID:  srv.ID,
		Type:      rule.Type,
		Threshold: rule.Threshold,
		Level:     rule.Level,
	}

	nginxCount := 0
	for _, p := range processes {
		if p.Name == "nginx" || p.Name == "nginx:" {
			nginxCount++
		}
	}

	result.MetricValue = float64(nginxCount)

	if nginxCount == 0 {
		result.Triggered = true
		result.Title = "Nginx进程异常"
		result.Message = fmt.Sprintf("服务器 %s 未检测到Nginx进程，服务可能已停止", srv.Name)
	}

	return result
}

// detectDocker 检测Docker容器
func (d *Detector) detectDocker(rule DetectRule, srv *server.Server, containers []server.DockerContainer) DetectionResult {
	result := DetectionResult{
		ServerID:  srv.ID,
		Type:      rule.Type,
		Threshold: rule.Threshold,
		Level:     rule.Level,
	}

	downCount := 0
	for _, c := range containers {
		if c.State != "running" {
			downCount++
		}
	}

	result.MetricValue = float64(downCount)

	if downCount > 0 {
		result.Triggered = true
		result.Title = "Docker容器异常"
		result.Message = fmt.Sprintf("服务器 %s 有 %d 个容器处于非运行状态", srv.Name, downCount)
	}

	return result
}

// detectPortAttack 检测端口攻击
func (d *Detector) detectPortAttack(rule DetectRule, srv *server.Server, ports []server.PortInfo) DetectionResult {
	result := DetectionResult{
		ServerID:  srv.ID,
		Type:      rule.Type,
		Threshold: rule.Threshold,
		Level:     rule.Level,
	}

	// 检测可疑端口
	suspiciousPorts := []int{22, 3389, 3306, 5432, 6379, 27017}
	suspiciousCount := 0

	for _, p := range ports {
		for _, sp := range suspiciousPorts {
			if p.Port == sp && p.State == "LISTEN" {
				suspiciousCount++
			}
		}
	}

	result.MetricValue = float64(suspiciousCount)

	if suspiciousCount > int(rule.Threshold) {
		result.Triggered = true
		result.Title = "端口安全告警"
		result.Message = fmt.Sprintf("服务器 %s 检测到 %d 个敏感端口暴露，可能存在安全风险", srv.Name, suspiciousCount)
	}

	return result
}
