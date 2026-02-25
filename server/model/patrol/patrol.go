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
	PatrolTypeTrigger PatrolType = "trigger"
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
	Duration    int64      `json:"duration"`
}

func (PatrolRecord) TableName() string {
	return "patrol_records"
}

// DailyReport 日报
type DailyReport struct {
	Date          string      `json:"date"`
	GeneratedAt   time.Time   `json:"generatedAt"`
	TotalServers  int         `json:"totalServers"`
	OnlineRate    float64     `json:"onlineRate"`
	AvgCPUUsage   float64     `json:"avgCpuUsage"`
	AvgMemoryUsage float64    `json:"avgMemoryUsage"`
	AvgDiskUsage  float64     `json:"avgDiskUsage"`
	TotalAlerts   int         `json:"totalAlerts"`
	CriticalAlerts int        `json:"criticalAlerts"`
	WarningAlerts int         `json:"warningAlerts"`
	ResolvedAlerts int        `json:"resolvedAlerts"`
	Recommendations []string  `json:"recommendations"`
}
