package patrol

import "time"

// PatrolStatus 巡检状态
type PatrolStatus string

const (
	PatrolStatusRunning   PatrolStatus = "running"
	PatrolStatusCompleted PatrolStatus = "completed"
	PatrolStatusFailed    PatrolStatus = "failed"
)

// PatrolType 巡检类型
type PatrolType string

const (
	PatrolTypeDaily   PatrolType = "daily"
	PatrolTypeHourly  PatrolType = "hourly"
	PatrolTypeManual  PatrolType = "manual"
	PatrolTypeTrigger PatrolType = "trigger" // 告警触发
)

// PatrolRecord 巡检记录
type PatrolRecord struct {
	ID        uint         `json:"id" gorm:"primarykey"`
	CreatedAt time.Time    `json:"createdAt"`
	Type      PatrolType   `json:"type" gorm:"type:varchar(16)"`
	Status    PatrolStatus `json:"status" gorm:"type:varchar(16)"`

	// 统计信息
	TotalServers   int `json:"totalServers"`
	OnlineServers  int `json:"onlineServers"`
	OfflineServers int `json:"offlineServers"`
	AlertCount     int `json:"alertCount"`
	CriticalCount  int `json:"criticalCount"`
	WarningCount   int `json:"warningCount"`

	// 服务器详情
	HealthyServers    string `json:"healthyServers" gorm:"type:text"`
	WarningServers    string `json:"warningServers" gorm:"type:text"`
	CriticalServers   string `json:"criticalServers" gorm:"type:text"`
	OfflineServerList string `json:"offlineServerList" gorm:"type:text"`

	// 报告
	Summary     string `json:"summary" gorm:"type:text"`
	Suggestions string `json:"suggestions" gorm:"type:text"`
	ReportURL   string `json:"reportUrl" gorm:"type:varchar(255)"`

	// 时间信息
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int64      `json:"duration"` // 毫秒
}

func (PatrolRecord) TableName() string {
	return "patrol_records"
}

// ServerCheckResult 服务器检查结果
type ServerCheckResult struct {
	ServerID    uint   `json:"serverId"`
	ServerName  string `json:"serverName"`
	Status      string `json:"status"` // healthy, warning, critical, offline
	Checks      []CheckItem
	Suggestions []string
}

// CheckItem 检查项
type CheckItem struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // pass, warning, fail
	Value   string `json:"value"`
	Message string `json:"message"`
}

// DailyReport 日报
type DailyReport struct {
	Date         string    `json:"date"`
	GeneratedAt  time.Time `json:"generatedAt"`

	// 概览
	TotalServers   int     `json:"totalServers"`
	OnlineRate     float64 `json:"onlineRate"`
	AvgCPUUsage    float64 `json:"avgCpuUsage"`
	AvgMemoryUsage float64 `json:"avgMemoryUsage"`
	AvgDiskUsage   float64 `json:"avgDiskUsage"`

	// 告警统计
	TotalAlerts    int `json:"totalAlerts"`
	CriticalAlerts int `json:"criticalAlerts"`
	WarningAlerts  int `json:"warningAlerts"`
	ResolvedAlerts int `json:"resolvedAlerts"`

	// 趋势
	Trends TrendAnalysis

	// 建议
	Recommendations []string
}

// ServerUsage 服务器使用率
type ServerUsage struct {
	ServerID   uint    `json:"serverId"`
	ServerName string  `json:"serverName"`
	Usage      float64 `json:"usage"`
}

// TrendAnalysis 趋势分析
type TrendAnalysis struct {
	CPUTrend    string `json:"cpuTrend"`    // up, down, stable
	MemoryTrend string `json:"memoryTrend"`
	DiskTrend   string `json:"diskTrend"`
	AlertTrend  string `json:"alertTrend"`
}
