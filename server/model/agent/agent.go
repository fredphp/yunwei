package agent

import (
        "time"

        "gorm.io/gorm"
)

// ==================== 版本管理 ====================

// AgentVersion Agent版本
type AgentVersion struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`
        UpdatedAt time.Time `json:"updatedAt"`

        // 版本信息
        Version     string `json:"version" gorm:"type:varchar(32);uniqueIndex:idx_version_platform_arch;not null;comment:版本号"`
        VersionCode int    `json:"versionCode" gorm:"not null;comment:版本代码(用于比较)"`

        // 构建信息
        BuildTime   time.Time `json:"buildTime" gorm:"comment:构建时间"`
        BuildCommit string    `json:"buildCommit" gorm:"type:varchar(64);comment:Git Commit"`
        BuildBranch string    `json:"buildBranch" gorm:"type:varchar(64);comment:Git Branch"`

        // 文件信息
        FileURL      string `json:"fileUrl" gorm:"type:varchar(512);comment:下载地址"`
        FileMD5      string `json:"fileMd5" gorm:"type:varchar(64);comment:文件MD5"`
        FileSHA256   string `json:"fileSha256" gorm:"type:varchar(128);comment:文件SHA256"`
        FileSize     int64  `json:"fileSize" gorm:"comment:文件大小(字节)"`
        SignatureURL string `json:"signatureUrl" gorm:"type:varchar(512);comment:签名文件地址"`

        // 平台支持
        Platform string `json:"platform" gorm:"type:varchar(32);uniqueIndex:idx_version_platform_arch;comment:平台(linux/windows/darwin)"`
        Arch     string `json:"arch" gorm:"type:varchar(32);uniqueIndex:idx_version_platform_arch;comment:架构(amd64/arm64/arm)"`

        // 兼容性
        MinVersion    string `json:"minVersion" gorm:"type:varchar(32);comment:最低可升级版本"`
        MinVersionCode int   `json:"minVersionCode" gorm:"comment:最低版本代码"`
        BreakingChanges bool `json:"breakingChanges" gorm:"default:false;comment:是否有破坏性变更"`

        // 更新内容
        Changelog    string `json:"changelog" gorm:"type:text;comment:更新日志"`
        ReleaseNotes string `json:"releaseNotes" gorm:"type:text;comment:发布说明"`
        ReleaseType  string `json:"releaseType" gorm:"type:varchar(16);default:'stable';comment:类型(stable/beta/alpha/nightly)"`

        // 升级策略
        ForceUpdate    bool `json:"forceUpdate" gorm:"default:false;comment:是否强制更新"`
        RollbackSupport bool `json:"rollbackSupport" gorm:"default:true;comment:是否支持回滚"`
        GracePeriod    int  `json:"gracePeriod" gorm:"default:0;comment:宽限期(小时),0=立即"`

        // 状态
        Enabled  bool `json:"enabled" gorm:"default:true;comment:是否启用"`
        IsLatest bool `json:"isLatest" gorm:"default:false;comment:是否最新版本"`

        // 统计
        DownloadCount int `json:"downloadCount" gorm:"default:0;comment:下载次数"`
        InstallCount  int `json:"installCount" gorm:"default:0;comment:安装次数"`
        SuccessCount  int `json:"successCount" gorm:"default:0;comment:成功安装数"`
        FailCount     int `json:"failCount" gorm:"default:0;comment:失败安装数"`
}

func (AgentVersion) TableName() string {
        return "agent_versions"
}

// ==================== Agent 实例 ====================

// AgentStatus Agent状态
type AgentStatus string

const (
        AgentStatusOnline     AgentStatus = "online"     // 在线
        AgentStatusOffline    AgentStatus = "offline"    // 离线
        AgentStatusUpgrading  AgentStatus = "upgrading"  // 升级中
        AgentStatusError      AgentStatus = "error"      // 异常
        AgentStatusPending    AgentStatus = "pending"    // 待安装
        AgentStatusDisabled   AgentStatus = "disabled"   // 已禁用
        AgentStatusRecovering AgentStatus = "recovering" // 恢复中
        AgentStatusRollback   AgentStatus = "rollback"   // 回滚中
)

// Agent Agent实例
type Agent struct {
        ID        uint           `json:"id" gorm:"primarykey"`
        CreatedAt time.Time      `json:"createdAt"`
        UpdatedAt time.Time      `json:"updatedAt"`
        DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

        // 关联服务器
        ServerID   uint   `json:"serverId" gorm:"uniqueIndex;comment:服务器ID"`
        ServerName string `json:"serverName" gorm:"type:varchar(128);comment:服务器名称"`
        ServerIP   string `json:"serverIp" gorm:"type:varchar(64);comment:服务器IP"`

        // Agent标识
        AgentID     string `json:"agentId" gorm:"type:varchar(64);uniqueIndex;comment:Agent唯一标识"`
        AgentSecret string `json:"-" gorm:"type:varchar(128);comment:Agent密钥"`

        // 版本信息
        Version       string `json:"version" gorm:"type:varchar(32);index;comment:当前版本"`
        VersionCode   int    `json:"versionCode" gorm:"default:0;comment:版本代码"`
        TargetVersion string `json:"targetVersion" gorm:"type:varchar(32);comment:目标版本(升级用)"`

        // 平台信息
        Platform string `json:"platform" gorm:"type:varchar(32);index;comment:平台"`
        Arch     string `json:"arch" gorm:"type:varchar(32);index;comment:架构"`
        OS       string `json:"os" gorm:"type:varchar(64);comment:操作系统"`
        Kernel   string `json:"kernel" gorm:"type:varchar(128);comment:内核版本"`

        // 状态
        Status        AgentStatus `json:"status" gorm:"type:varchar(16);index;default:'pending';comment:状态"`
        StatusMessage string      `json:"statusMessage" gorm:"type:varchar(512);comment:状态消息"`
        StatusChangedAt *time.Time `json:"statusChangedAt" gorm:"comment:状态变更时间"`

        // 心跳信息
        LastHeartbeat *time.Time `json:"lastHeartbeat" gorm:"index;comment:最后心跳时间"`
        HeartbeatIP   string     `json:"heartbeatIp" gorm:"type:varchar(64);comment:心跳来源IP"`
        HeartbeatPort int        `json:"heartbeatPort" gorm:"comment:心跳来源端口"`
        HeartbeatLatency int     `json:"heartbeatLatency" gorm:"comment:心跳延迟(ms)"`

        // 离线监控
        OfflineCount     int        `json:"offlineCount" gorm:"default:0;comment:离线次数累计"`
        LastOfflineAt    *time.Time `json:"lastOfflineAt" gorm:"comment:最后离线时间"`
        LastOnlineAt     *time.Time `json:"lastOnlineAt" gorm:"comment:最后在线时间"`
        TotalOfflineTime int64      `json:"totalOfflineTime" gorm:"default:0;comment:累计离线时间(秒)"`

        // 错误监控
        ErrorCount   int        `json:"errorCount" gorm:"default:0;comment:错误次数"`
        LastErrorAt  *time.Time `json:"lastErrorAt" gorm:"comment:最后错误时间"`
        LastErrorMsg string     `json:"lastErrorMsg" gorm:"type:text;comment:最后错误消息"`

        // 自动恢复
        AutoRecover   bool       `json:"autoRecover" gorm:"default:true;comment:是否自动恢复"`
        RecoverCount  int        `json:"recoverCount" gorm:"default:0;comment:恢复次数"`
        LastRecoverAt *time.Time `json:"lastRecoverAt" gorm:"comment:最后恢复时间"`

        // 升级配置
        AutoUpgrade    bool   `json:"autoUpgrade" gorm:"default:true;comment:是否自动升级"`
        UpgradeChannel string `json:"upgradeChannel" gorm:"type:varchar(16);default:'stable';comment:升级通道(stable/beta/alpha)"`
        UpgradeWindow  string `json:"upgradeWindow" gorm:"type:varchar(64);comment:升级时间窗口(如: 02:00-06:00)"`

        // 灰度发布
        GrayGroup  string `json:"grayGroup" gorm:"type:varchar(32);index;comment:灰度分组"`
        GrayWeight int    `json:"grayWeight" gorm:"default:0;comment:灰度权重(0-100)"`

        // 配置同步
        ConfigHash  string     `json:"configHash" gorm:"type:varchar(64);comment:配置Hash"`
        ConfigAt    *time.Time `json:"configAt" gorm:"comment:配置同步时间"`
        ConfigError string     `json:"configError" gorm:"type:varchar(512);comment:配置错误"`

        // 资源统计
        UptimeSeconds  int64 `json:"uptimeSeconds" gorm:"comment:运行时长(秒)"`
        TaskCount      int64 `json:"taskCount" gorm:"default:0;comment:执行任务总数"`
        TaskSuccessCount int64 `json:"taskSuccessCount" gorm:"default:0;comment:成功任务数"`
        TaskFailCount  int64 `json:"taskFailCount" gorm:"default:0;comment:失败任务数"`

        // 标签
        Tags string `json:"tags" gorm:"type:text;comment:标签(JSON数组)"`

        // 备注
        Description string `json:"description" gorm:"type:varchar(512);comment:描述"`
}

func (Agent) TableName() string {
        return "agents"
}

// ==================== 升级任务 ====================

// UpgradeTaskStatus 升级任务状态
type UpgradeTaskStatus string

const (
        UpgradeStatusPending    UpgradeTaskStatus = "pending"    // 待执行
        UpgradeStatusScheduled  UpgradeTaskStatus = "scheduled"  // 已调度
        UpgradeStatusDownloading UpgradeTaskStatus = "downloading" // 下载中
        UpgradeStatusInstalling UpgradeTaskStatus = "installing" // 安装中
        UpgradeStatusVerifying  UpgradeTaskStatus = "verifying"  // 验证中
        UpgradeStatusSuccess    UpgradeTaskStatus = "success"    // 成功
        UpgradeStatusFailed     UpgradeTaskStatus = "failed"     // 失败
        UpgradeStatusCanceled   UpgradeTaskStatus = "canceled"   // 已取消
        UpgradeStatusRolledback UpgradeTaskStatus = "rolledback" // 已回滚
)

// AgentUpgradeTask 升级任务
type AgentUpgradeTask struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`
        UpdatedAt time.Time `json:"updatedAt"`

        // Agent信息
        AgentID    uint   `json:"agentId" gorm:"index;not null;comment:Agent ID"`
        AgentUUID  string `json:"agentUuid" gorm:"type:varchar(64);index;comment:Agent UUID"`
        ServerID   uint   `json:"serverId" gorm:"index;comment:服务器ID"`
        ServerName string `json:"serverName" gorm:"type:varchar(128);comment:服务器名称"`
        ServerIP   string `json:"serverIp" gorm:"type:varchar(64);comment:服务器IP"`

        // 版本信息
        FromVersion     string `json:"fromVersion" gorm:"type:varchar(32);comment:原版本"`
        FromVersionCode int    `json:"fromVersionCode" gorm:"comment:原版本代码"`
        ToVersion       string `json:"toVersion" gorm:"type:varchar(32);comment:目标版本"`
        ToVersionCode   int    `json:"toVersionCode" gorm:"comment:目标版本代码"`
        VersionID       uint   `json:"versionId" gorm:"comment:版本记录ID"`

        // 任务信息
        TaskType     string     `json:"taskType" gorm:"type:varchar(16);default:'manual';comment:任务类型(manual/auto/gray/scheduled)"`
        Priority     int        `json:"priority" gorm:"default:5;comment:优先级(1-10)"`
        ScheduledAt  *time.Time `json:"scheduledAt" gorm:"comment:计划执行时间"`
        StrategyID   uint       `json:"strategyId" gorm:"comment:灰度策略ID"`

        // 状态
        Status     string `json:"status" gorm:"type:varchar(16);default:'pending';index;comment:状态"`
        StatusMsg  string `json:"statusMsg" gorm:"type:varchar(512);comment:状态消息"`
        Progress   int    `json:"progress" gorm:"default:0;comment:进度(0-100)"`
        ProgressDetail string `json:"progressDetail" gorm:"type:text;comment:进度详情(JSON)"`

        // 执行信息
        StartedAt   *time.Time `json:"startedAt" gorm:"comment:开始时间"`
        CompletedAt *time.Time `json:"completedAt" gorm:"comment:完成时间"`
        Duration    int64      `json:"duration" gorm:"comment:耗时(毫秒)"`

        // 下载信息
        DownloadURL    string `json:"downloadUrl" gorm:"type:varchar(512);comment:下载地址"`
        DownloadSize   int64  `json:"downloadSize" gorm:"comment:文件大小"`
        DownloadMD5    string `json:"downloadMd5" gorm:"type:varchar(64);comment:文件MD5"`
        DownloadedSize int64  `json:"downloadedSize" gorm:"comment:已下载大小"`
        DownloadSpeed  int64  `json:"downloadSpeed" gorm:"comment:下载速度(KB/s)"`

        // 结果
        Result      string `json:"result" gorm:"type:text;comment:结果详情"`
        Output      string `json:"output" gorm:"type:text;comment:输出日志"`
        Error       string `json:"error" gorm:"type:text;comment:错误信息"`

        // 回滚
        RollbackEnabled bool       `json:"rollbackEnabled" gorm:"default:true;comment:允许回滚"`
        RollbackAt      *time.Time `json:"rollbackAt" gorm:"comment:回滚时间"`
        RollbackError   string     `json:"rollbackError" gorm:"type:text;comment:回滚错误"`

        // 重试
        RetryCount int `json:"retryCount" gorm:"default:0;comment:重试次数"`
        MaxRetry   int `json:"maxRetry" gorm:"default:3;comment:最大重试次数"`
        NextRetryAt *time.Time `json:"nextRetryAt" gorm:"comment:下次重试时间"`

        // 创建者
        CreatedBy     uint   `json:"createdBy" gorm:"comment:创建者ID"`
        CreatedByName string `json:"createdByName" gorm:"type:varchar(64);comment:创建者名称"`
}

func (AgentUpgradeTask) TableName() string {
        return "agent_upgrade_tasks"
}

// ==================== 心跳记录 ====================

// AgentHeartbeatRecord 心跳记录
type AgentHeartbeatRecord struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt" gorm:"index"`

        AgentID   uint   `json:"agentId" gorm:"index;comment:Agent ID"`
        AgentUUID string `json:"agentUuid" gorm:"type:varchar(64);index;comment:Agent UUID"`
        ServerID  uint   `json:"serverId" gorm:"index;comment:服务器ID"`

        // 心跳信息
        IP      string `json:"ip" gorm:"type:varchar(64);comment:来源IP"`
        Port    int    `json:"port" gorm:"comment:来源端口"`
        Version string `json:"version" gorm:"type:varchar(32);comment:版本"`
        Status  string `json:"status" gorm:"type:varchar(16);comment:状态"`

        // 运行信息
        UptimeSeconds  int64   `json:"uptimeSeconds" gorm:"comment:运行时长"`
        CPUUsage       float64 `json:"cpuUsage" gorm:"comment:CPU使用率"`
        MemoryUsage    float64 `json:"memoryUsage" gorm:"comment:内存使用率"`
        GoroutineCount int     `json:"goroutineCount" gorm:"comment:协程数"`

        // 任务信息
        PendingTasks   int `json:"pendingTasks" gorm:"comment:待处理任务数"`
        RunningTasks   int `json:"runningTasks" gorm:"comment:运行中任务数"`
        CompletedTasks int `json:"completedTasks" gorm:"comment:已完成任务数"`
        FailedTasks    int `json:"failedTasks" gorm:"comment:失败任务数"`

        // 网络
        NetInBytes  uint64 `json:"netInBytes" gorm:"comment:网络入流量"`
        NetOutBytes uint64 `json:"netOutBytes" gorm:"comment:网络出流量"`

        // 延迟
        LatencyMs int `json:"latencyMs" gorm:"comment:心跳延迟(ms)"`
}

func (AgentHeartbeatRecord) TableName() string {
        return "agent_heartbeat_records"
}

// ==================== 恢复记录 ====================

// RecoverTriggerType 恢复触发类型
type RecoverTriggerType string

const (
        RecoverTriggerAuto     RecoverTriggerType = "auto"     // 自动触发
        RecoverTriggerManual   RecoverTriggerType = "manual"   // 手动触发
        RecoverTriggerScheduled RecoverTriggerType = "scheduled" // 定时触发
)

// AgentRecoverRecord 恢复记录
type AgentRecoverRecord struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`

        AgentID    uint   `json:"agentId" gorm:"index;comment:Agent ID"`
        AgentUUID  string `json:"agentUuid" gorm:"type:varchar(64);comment:Agent UUID"`
        ServerID   uint   `json:"serverId" gorm:"index;comment:服务器ID"`
        ServerName string `json:"serverName" gorm:"type:varchar(128);comment:服务器名称"`
        ServerIP   string `json:"serverIp" gorm:"type:varchar(64);comment:服务器IP"`

        // 恢复信息
        TriggerType  RecoverTriggerType `json:"triggerType" gorm:"type:varchar(16);comment:触发类型"`
        TriggerCause string             `json:"triggerCause" gorm:"type:varchar(64);comment:触发原因"`
        TriggerMsg   string             `json:"triggerMsg" gorm:"type:text;comment:触发消息"`

        // 执行信息
        Action      string     `json:"action" gorm:"type:varchar(32);comment:恢复动作"`
        Command     string     `json:"command" gorm:"type:text;comment:执行命令"`
        Status      string     `json:"status" gorm:"type:varchar(16);comment:状态"`

        StartedAt   *time.Time `json:"startedAt" gorm:"comment:开始时间"`
        CompletedAt *time.Time `json:"completedAt" gorm:"comment:完成时间"`
        Duration    int64      `json:"duration" gorm:"comment:耗时(毫秒)"`

        Output      string `json:"output" gorm:"type:text;comment:输出"`
        Error       string `json:"error" gorm:"type:text;comment:错误"`

        // 结果
        Success    bool `json:"success" gorm:"comment:是否成功"`
        RetryCount int  `json:"retryCount" gorm:"comment:重试次数"`
}

func (AgentRecoverRecord) TableName() string {
        return "agent_recover_records"
}

// ==================== 灰度发布策略 ====================

// GrayStrategyStatus 灰度策略状态
type GrayStrategyStatus string

const (
        GrayStatusPending   GrayStrategyStatus = "pending"   // 待执行
        GrayStatusRunning   GrayStrategyStatus = "running"   // 执行中
        GrayStatusPaused    GrayStrategyStatus = "paused"    // 已暂停
        GrayStatusCompleted GrayStrategyStatus = "completed" // 已完成
        GrayStatusCanceled  GrayStrategyStatus = "canceled"  // 已取消
        GrayStatusRolledback GrayStrategyStatus = "rolledback" // 已回滚
)

// GrayReleaseStrategy 灰度发布策略
type GrayReleaseStrategy struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`
        UpdatedAt time.Time `json:"updatedAt"`

        Name        string `json:"name" gorm:"type:varchar(64);not null;comment:策略名称"`
        Description string `json:"description" gorm:"type:varchar(512);comment:描述"`

        // 目标版本
        VersionID   uint   `json:"versionId" gorm:"comment:版本ID"`
        VersionName string `json:"versionName" gorm:"type:varchar(32);comment:版本号"`

        // 灰度规则
        StrategyType string `json:"strategyType" gorm:"type:varchar(16);comment:策略类型(weight/group/label/canary)"`

        // 权重策略
        InitialWeight int `json:"initialWeight" gorm:"default:5;comment:初始权重(%)"`
        TargetWeight  int `json:"targetWeight"  gorm:"default:100;comment:目标权重(%)"`
        WeightPercent int `json:"weightPercent" gorm:"default:5;comment:当前灰度权重(%)"`
        StepSize      int `json:"stepSize" gorm:"default:10;comment:步进大小(%)"`
        StepInterval  int `json:"stepInterval" gorm:"default:30;comment:步进间隔(分钟)"`

        // 分组策略
        GroupList string `json:"groupList" gorm:"type:text;comment:灰度分组列表(JSON)"`

        // 标签策略
        LabelSelector string `json:"labelSelector" gorm:"type:text;comment:标签选择器(JSON)"`

        // 服务器范围
        ServerIDs string `json:"serverIds" gorm:"type:text;comment:指定服务器ID列表(JSON)"`

        // 状态
        Status      string `json:"status" gorm:"type:varchar(16);default:'pending';index;comment:状态"`
        CurrentStep int    `json:"currentStep" gorm:"default:0;comment:当前步数"`
        CurrentWeight int  `json:"currentWeight" gorm:"default:0;comment:当前权重(%)"`

        // 时间
        ScheduledAt *time.Time `json:"scheduledAt" gorm:"comment:计划开始时间"`
        StartedAt   *time.Time `json:"startedAt" gorm:"comment:实际开始时间"`
        CompletedAt *time.Time `json:"completedAt" gorm:"comment:完成时间"`
        PausedAt    *time.Time `json:"pausedAt" gorm:"comment:暂停时间"`

        // 统计
        TotalAgents    int `json:"totalAgents" gorm:"comment:总Agent数"`
        UpgradedAgents int `json:"upgradedAgents" gorm:"comment:已升级数"`
        SuccessAgents  int `json:"successAgents" gorm:"comment:成功数"`
        FailedAgents   int `json:"failedAgents" gorm:"comment:失败数"`

        // 监控配置
        PauseOnFailure   bool    `json:"pauseOnFailure" gorm:"default:true;comment:失败时暂停"`
        FailureThreshold float64 `json:"failureThreshold" gorm:"default:10;comment:失败阈值(%)"`
        AutoRollback     bool    `json:"autoRollback" gorm:"default:true;comment:自动回滚"`
        RollbackThreshold float64 `json:"rollbackThreshold" gorm:"default:30;comment:回滚阈值(%)"`

        // 健康检查
        HealthCheckEnabled bool   `json:"healthCheckEnabled" gorm:"default:true;comment:启用健康检查"`
        HealthCheckURL     string `json:"healthCheckUrl" gorm:"type:varchar(256);comment:健康检查URL"`
        HealthCheckTimeout int    `json:"healthCheckTimeout" gorm:"default:30;comment:健康检查超时(秒)"`

        Enabled bool `json:"enabled" gorm:"default:true;comment:是否启用"`

        // 创建者
        CreatedBy     uint   `json:"createdBy" gorm:"comment:创建者ID"`
        CreatedByName string `json:"createdByName" gorm:"type:varchar(64);comment:创建者名称"`
}

func (GrayReleaseStrategy) TableName() string {
        return "gray_release_strategies"
}

// ==================== Agent 配置 ====================

// AgentConfig Agent配置
type AgentConfig struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`
        UpdatedAt time.Time `json:"updatedAt"`

        // 配置信息
        Name        string `json:"name" gorm:"type:varchar(64);not null;comment:配置名称"`
        Description string `json:"description" gorm:"type:varchar(512);comment:描述"`

        // 配置内容
        ConfigJSON string `json:"configJson" gorm:"type:longtext;comment:配置内容(JSON)"`
        ConfigHash string `json:"configHash" gorm:"type:varchar(64);index;comment:配置Hash"`

        // 应用范围
        Scope      string `json:"scope" gorm:"type:varchar(16);comment:范围(all/group/agent)"`
        ScopeValue string `json:"scopeValue" gorm:"type:text;comment:范围值(JSON)"`

        // 版本控制
        Version     int  `json:"version" gorm:"default:1;comment:配置版本"`
        IsDefault   bool `json:"isDefault" gorm:"default:false;comment:是否默认配置"`

        // 状态
        Enabled bool `json:"enabled" gorm:"default:true;comment:是否启用"`

        // 统计
        AppliedCount int `json:"appliedCount" gorm:"default:0;comment:应用次数"`
}

func (AgentConfig) TableName() string {
        return "agent_configs"
}

// ==================== Agent 指标 ====================

// AgentMetric Agent指标
type AgentMetric struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt" gorm:"index"`

        AgentID  uint `json:"agentId" gorm:"index;comment:Agent ID"`
        ServerID uint `json:"serverId" gorm:"index;comment:服务器ID"`

        // 资源使用
        CPUUsage    float64 `json:"cpuUsage" gorm:"comment:CPU使用率"`
        MemoryUsage float64 `json:"memoryUsage" gorm:"comment:内存使用率"`
        MemoryUsed  uint64  `json:"memoryUsed" gorm:"comment:已用内存(MB)"`
        MemoryTotal uint64  `json:"memoryTotal" gorm:"comment:总内存(MB)"`

        // 运行状态
        GoroutineCount int `json:"goroutineCount" gorm:"comment:协程数"`
        ThreadCount    int `json:"threadCount" gorm:"comment:线程数"`
        HandleCount    int `json:"handleCount" gorm:"comment:句柄数"`

        // 网络
        NetInBytes  uint64 `json:"netInBytes" gorm:"comment:网络入流量"`
        NetOutBytes uint64 `json:"netOutBytes" gorm:"comment:网络出流量"`

        // 任务
        PendingTasks   int `json:"pendingTasks" gorm:"comment:待处理任务"`
        RunningTasks   int `json:"runningTasks" gorm:"comment:运行中任务"`
        CompletedTasks int `json:"completedTasks" gorm:"comment:已完成任务"`
        FailedTasks    int `json:"failedTasks" gorm:"comment:失败任务"`

        // 延迟
        TaskAvgLatency float64 `json:"taskAvgLatency" gorm:"comment:任务平均延迟(ms)"`
        TaskMaxLatency float64 `json:"taskMaxLatency" gorm:"comment:任务最大延迟(ms)"`
}

func (AgentMetric) TableName() string {
        return "agent_metrics"
}
