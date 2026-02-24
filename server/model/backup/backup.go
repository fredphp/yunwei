package backup

import (
	"time"
)

// ==================== 备份策略 ====================

// BackupPolicy 备份策略
type BackupPolicy struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Type        string    `gorm:"size:20;not null" json:"type"` // database, file, snapshot, full
	Description string    `gorm:"size:500" json:"description"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`

	// 备份目标
	TargetID    uint   `gorm:"not null" json:"target_id"`     // 目标ID(服务器/数据库ID)
	TargetType  string `gorm:"size:50;not null" json:"target_type"` // server, database, application
	TargetName  string `gorm:"size:100" json:"target_name"`

	// 备份源配置
	SourceType   string `gorm:"size:50;not null" json:"source_type"`   // mysql, postgresql, mongodb, redis, file, directory
	SourceConfig string `gorm:"type:text" json:"source_config"`        // JSON配置
	SourcePath   string `gorm:"size:500" json:"source_path"`           // 文件路径(文件备份)
	ExcludePaths string `gorm:"type:text" json:"exclude_paths"`        // 排除路径(多个用逗号分隔)

	// 备份计划
	ScheduleType   string `gorm:"size:20;not null" json:"schedule_type"`   // cron, interval, manual
	ScheduleExpr   string `gorm:"size:100" json:"schedule_expr"`           // cron表达式
	IntervalMinute int    `json:"interval_minute"`                         // 间隔分钟数

	// 备份保留
	RetentionDays   int  `gorm:"default:30" json:"retention_days"`     // 保留天数
	RetentionCount  int  `gorm:"default:10" json:"retention_count"`    // 保留数量
	Compress        bool `gorm:"default:true" json:"compress"`         // 是否压缩
	CompressType    string `gorm:"size:20;default:'gzip'" json:"compress_type"` // gzip, zstd, none
	Encrypt         bool `gorm:"default:false" json:"encrypt"`        // 是否加密
	EncryptKey      string `gorm:"size:100" json:"encrypt_key"`          // 加密密钥(加密存储)

	// 存储配置
	StorageType   string `gorm:"size:20;not null" json:"storage_type"`   // local, s3, oss, nfs, ftp
	StorageConfig string `gorm:"type:text" json:"storage_config"`        // JSON配置
	StoragePath   string `gorm:"size:500" json:"storage_path"`           // 存储路径

	// 高级选项
	PreScript    string `gorm:"type:text" json:"pre_script"`    // 备份前脚本
	PostScript   string `gorm:"type:text" json:"post_script"`   // 备份后脚本
	MaxFileSize  int64  `json:"max_file_size"`                   // 最大文件大小(MB), 0表示不限制
	SplitSize    int64  `json:"split_size"`                      // 分卷大小(MB), 0表示不分卷

	// 并发控制
	MaxConcurrency int  `gorm:"default:1" json:"max_concurrency"` // 最大并发数
	Timeout        int  `gorm:"default:3600" json:"timeout"`      // 超时时间(秒)
	RetryCount     int  `gorm:"default:3" json:"retry_count"`     // 重试次数
	RetryInterval  int  `gorm:"default:60" json:"retry_interval"` // 重试间隔(秒)

	// 告警配置
	NotifyOnSuccess bool   `gorm:"default:false" json:"notify_on_success"`
	NotifyOnFail    bool   `gorm:"default:true" json:"notify_on_fail"`
	NotifyChannels  string `gorm:"size:200" json:"notify_channels"` // 通知渠道(逗号分隔)

	// 统计
	LastBackupTime   *time.Time `json:"last_backup_time"`
	LastBackupSize   int64      `json:"last_backup_size"`
	LastBackupStatus string     `gorm:"size:20" json:"last_backup_status"`
	TotalBackupCount int        `gorm:"default:0" json:"total_backup_count"`
	TotalBackupSize  int64      `gorm:"default:0" json:"total_backup_size"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Records []BackupRecord `gorm:"foreignKey:PolicyID" json:"records,omitempty"`
}

// BackupRecord 备份记录
type BackupRecord struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	PolicyID uint   `gorm:"not null;index" json:"policy_id"`
	TaskID   string `gorm:"size:50;index" json:"task_id"` // 任务ID

	// 备份信息
	BackupType   string    `gorm:"size:20;not null" json:"backup_type"` // full, incremental, differential
	TriggerType  string    `gorm:"size:20;not null" json:"trigger_type"` // manual, schedule, auto
	TriggerBy    string    `gorm:"size:100" json:"trigger_by"`          // 触发人/系统

	// 执行状态
	Status      string    `gorm:"size:20;not null" json:"status"` // pending, running, success, failed, cancelled
	StartTime   time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int        `json:"duration"` // 持续时间(秒)

	// 备份文件
	FileName     string `gorm:"size:200" json:"file_name"`
	FilePath     string `gorm:"size:500" json:"file_path"`
	FileSize     int64  `json:"file_size"`      // 原始大小(字节)
	CompressSize int64  `json:"compress_size"`  // 压缩后大小(字节)
	Checksum     string `gorm:"size:64" json:"checksum"` // SHA256校验

	// 存储信息
	StorageType   string `gorm:"size:20" json:"storage_type"`
	StoragePath   string `gorm:"size:500" json:"storage_path"`
	StorageNode   string `gorm:"size:100" json:"storage_node"` // 存储节点

	// 增量备份
	BaseBackupID  *uint  `json:"base_backup_id"`   // 基础备份ID(增量备份)
	ChangedFiles  int    `json:"changed_files"`    // 变更文件数
	ChangedSize   int64  `json:"changed_size"`     // 变更大小

	// 元数据
	Metadata string `gorm:"type:text" json:"metadata"` // JSON格式元数据

	// 执行日志
	Log string `gorm:"type:text" json:"log"`

	// 错误信息
	ErrorCode string `gorm:"size:50" json:"error_code"`
	ErrorMsg  string `gorm:"type:text" json:"error_msg"`

	CreatedAt time.Time `json:"created_at"`

	// 关联
	Policy    *BackupPolicy `gorm:"foreignKey:PolicyID" json:"policy,omitempty"`
	Restores  []RestoreRecord `gorm:"foreignKey:BackupID" json:"restores,omitempty"`
}

// ==================== 恢复管理 ====================

// RestoreRecord 恢复记录
type RestoreRecord struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	BackupID  uint   `gorm:"not null;index" json:"backup_id"`
	TaskID    string `gorm:"size:50;index" json:"task_id"`

	// 恢复目标
	TargetID    uint   `gorm:"not null" json:"target_id"`
	TargetType  string `gorm:"size:50" json:"target_type"`
	TargetName  string `gorm:"size:100" json:"target_name"`

	// 恢复配置
	RestoreType   string `gorm:"size:20;not null" json:"restore_type"` // full, partial, point_in_time
	RestorePath   string `gorm:"size:500" json:"restore_path"`         // 恢复路径
	PointInTime   *time.Time `json:"point_in_time"`                    // 时间点恢复
	Overwrite     bool   `gorm:"default:false" json:"overwrite"`       // 是否覆盖

	// 执行状态
	Status      string     `gorm:"size:20;not null" json:"status"` // pending, running, verifying, success, failed, cancelled
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int        `json:"duration"`

	// 恢复进度
	TotalFiles   int   `json:"total_files"`
	RestoredFiles int  `json:"restored_files"`
	TotalSize    int64 `json:"total_size"`
	RestoredSize int64 `json:"restored_size"`
	Progress     int   `json:"progress"` // 进度百分比

	// 恢复脚本
	PreScript  string `gorm:"type:text" json:"pre_script"`  // 恢复前脚本
	PostScript string `gorm:"type:text" json:"post_script"` // 恢复后脚本

	// 验证信息
	VerifyEnabled  bool   `gorm:"default:true" json:"verify_enabled"`
	VerifyStatus   string `gorm:"size:20" json:"verify_status"` // pending, running, passed, failed
	VerifyResult   string `gorm:"type:text" json:"verify_result"`
	VerifyTime     *time.Time `json:"verify_time"`

	// 执行日志
	Log string `gorm:"type:text" json:"log"`

	// 错误信息
	ErrorCode string `gorm:"size:50" json:"error_code"`
	ErrorMsg  string `gorm:"type:text" json:"error_msg"`

	// 触发信息
	TriggerType string `gorm:"size:20" json:"trigger_type"` // manual, auto, drill
	TriggerBy   string `gorm:"size:100" json:"trigger_by"`

	CreatedAt time.Time `json:"created_at"`

	// 关联
	Backup *BackupRecord `gorm:"foreignKey:BackupID" json:"backup,omitempty"`
}

// ==================== 灾备演练 ====================

// DrillPlan 灾备演练计划
type DrillPlan struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Type        string    `gorm:"size:20;not null" json:"type"` // table_top, partial, full
	Scenario    string    `gorm:"size:50;not null" json:"scenario"` // server_failure, db_failure, data_corruption, ransomware, natural_disaster

	// 演练目标
	Objectives string `gorm:"type:text" json:"objectives"` // JSON数组
	Scope      string `gorm:"type:text" json:"scope"`      // 演练范围

	// 演练配置
	TargetSystems  string `gorm:"type:text" json:"target_systems"`  // JSON数组,目标系统
	BackupPolicyIDs string `gorm:"type:text" json:"backup_policy_ids"` // JSON数组,关联的备份策略

	// 时间配置
	ScheduledTime time.Time  `json:"scheduled_time"`
	EstimatedDuration int     `gorm:"default:60" json:"estimated_duration"` // 预计时长(分钟)
	ActualStartTime  *time.Time `json:"actual_start_time"`
	ActualEndTime    *time.Time `json:"actual_end_time"`

	// 演练团队
	Commander    string `gorm:"size:100" json:"commander"`     // 指挥官
	Participants string `gorm:"type:text" json:"participants"` // JSON数组,参与者

	// 演练步骤
	Steps string `gorm:"type:text" json:"steps"` // JSON数组,演练步骤

	// 状态
	Status string `gorm:"size:20;default:'planned'" json:"status"` // planned, in_progress, completed, cancelled, failed

	// 结果
	Result        string `gorm:"size:20" json:"result"` // success, partial, failed
	Score         int    `json:"score"`                 // 评分(0-100)
	Findings      string `gorm:"type:text" json:"findings"`      // 发现的问题
	Improvements  string `gorm:"type:text" json:"improvements"` // 改进建议
	Lessons       string `gorm:"type:text" json:"lessons"`      // 经验教训

	// RTO/RPO 验证
	TargetRTO    int  `gorm:"default:60" json:"target_rto"`    // 目标RTO(分钟)
	TargetRPO    int  `gorm:"default:5" json:"target_rpo"`     // 目标RPO(分钟)
	ActualRTO    int  `json:"actual_rto"`                      // 实际RTO
	ActualRPO    int  `json:"actual_rpo"`                      // 实际RPO
	RTOMet       bool `json:"rto_met"`                         // 是否满足RTO
	RPOMet       bool `json:"rpo_met"`                         // 是否满足RPO

	// 报告
	ReportURL string `gorm:"size:500" json:"report_url"` // 演练报告URL

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Executions []DrillExecution `gorm:"foreignKey:DrillID" json:"executions,omitempty"`
}

// DrillExecution 演练执行记录
type DrillExecution struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	DrillID  uint   `gorm:"not null;index" json:"drill_id"`
	StepID   int    `gorm:"not null" json:"step_id"` // 步骤序号
	StepName string `gorm:"size:100" json:"step_name"`

	// 执行信息
	Action      string `gorm:"type:text" json:"action"`        // 执行动作
	Expected    string `gorm:"type:text" json:"expected"`      // 预期结果
	Actual      string `gorm:"type:text" json:"actual"`        // 实际结果

	// 状态
	Status      string     `gorm:"size:20" json:"status"` // pending, running, success, failed, skipped
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int        `json:"duration"`

	// 结果
	Passed    bool   `json:"passed"`
	Issues    string `gorm:"type:text" json:"issues"`     // 发现的问题
	Solutions string `gorm:"type:text" json:"solutions"`  // 解决方案

	// 日志
	Log string `gorm:"type:text" json:"log"`

	CreatedAt time.Time `json:"created_at"`

	// 关联
	Drill *DrillPlan `gorm:"foreignKey:DrillID" json:"drill,omitempty"`
}

// ==================== 快照管理 ====================

// SnapshotPolicy 快照策略
type SnapshotPolicy struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`

	// 目标配置
	TargetID    uint   `gorm:"not null" json:"target_id"`
	TargetType  string `gorm:"size:50;not null" json:"target_type"` // vm, volume, filesystem, database
	TargetName  string `gorm:"size:100" json:"target_name"`

	// 快照类型
	SnapshotType string `gorm:"size:20;not null" json:"snapshot_type"` // full, incremental

	// 计划
	ScheduleType   string `gorm:"size:20;not null" json:"schedule_type"`
	ScheduleExpr   string `gorm:"size:100" json:"schedule_expr"`
	IntervalMinute int    `json:"interval_minute"`

	// 保留策略
	RetentionDays  int `gorm:"default:7" json:"retention_days"`
	RetentionCount int `gorm:"default:10" json:"retention_count"`
	MaxStorageSize int64 `json:"max_storage_size"` // 最大存储大小(GB)

	// 快照配置
	Quiesce       bool `gorm:"default:true" json:"quiesce"`       // 是否静默
	Consistent    bool `gorm:"default:true" json:"consistent"`    // 是否一致性快照
	Compress      bool `gorm:"default:true" json:"compress"`      // 是否压缩

	// 告警配置
	NotifyOnSuccess bool   `gorm:"default:false" json:"notify_on_success"`
	NotifyOnFail    bool   `gorm:"default:true" json:"notify_on_fail"`
	NotifyChannels  string `gorm:"size:200" json:"notify_channels"`

	// 统计
	LastSnapshotTime   *time.Time `json:"last_snapshot_time"`
	LastSnapshotStatus string     `gorm:"size:20" json:"last_snapshot_status"`
	TotalSnapshotCount int        `gorm:"default:0" json:"total_snapshot_count"`
	TotalStorageSize   int64      `gorm:"default:0" json:"total_storage_size"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Snapshots []SnapshotRecord `gorm:"foreignKey:PolicyID" json:"snapshots,omitempty"`
}

// SnapshotRecord 快照记录
type SnapshotRecord struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	PolicyID uint   `gorm:"not null;index" json:"policy_id"`
	SnapID   string `gorm:"size:100;uniqueIndex" json:"snap_id"` // 快照ID

	// 快照信息
	Name        string    `gorm:"size:100" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	SnapType    string    `gorm:"size:20" json:"snap_type"` // full, incremental

	// 执行状态
	Status      string     `gorm:"size:20" json:"status"` // creating, available, deleting, error
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int        `json:"duration"`

	// 大小信息
	VolumeSize  int64 `json:"volume_size"`   // 卷大小(字节)
	SnapSize    int64 `json:"snap_size"`     // 快照大小(字节)
	StorageSize int64 `json:"storage_size"`  // 实际存储大小(字节)

	// 增量快照
	BaseSnapID *uint  `json:"base_snap_id"`  // 基础快照ID
	ChangedSize int64 `json:"changed_size"`  // 变更大小

	// 存储位置
	StorageType string `gorm:"size:20" json:"storage_type"`
	StorageID   string `gorm:"size:100" json:"storage_id"`

	// 元数据
	Metadata string `gorm:"type:text" json:"metadata"`

	// 标签
	Tags string `gorm:"type:text" json:"tags"` // JSON

	// 过期时间
	ExpireTime *time.Time `json:"expire_time"`

	// 触发信息
	TriggerType string `gorm:"size:20" json:"trigger_type"` // manual, schedule, auto
	TriggerBy   string `gorm:"size:100" json:"trigger_by"`

	// 错误信息
	ErrorCode string `gorm:"size:50" json:"error_code"`
	ErrorMsg  string `gorm:"type:text" json:"error_msg"`

	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Policy *SnapshotPolicy `gorm:"foreignKey:PolicyID" json:"policy,omitempty"`
}

// ==================== 恢复验证 ====================

// VerifyTask 验证任务
type VerifyTask struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	TaskID   string `gorm:"size:50;uniqueIndex" json:"task_id"`
	RecordID uint   `gorm:"not null;index" json:"record_id"` // 备份记录或恢复记录ID
	RecordType string `gorm:"size:20;not null" json:"record_type"` // backup, restore, snapshot

	// 验证类型
	VerifyType string `gorm:"size:20;not null" json:"verify_type"` // integrity, consistency, recoverability, full

	// 验证配置
	Checksum    bool `gorm:"default:true" json:"checksum"`    // 校验和验证
	FileCount   bool `gorm:"default:true" json:"file_count"`  // 文件数量验证
	FileSize    bool `gorm:"default:true" json:"file_size"`   // 文件大小验证
	Structure   bool `gorm:"default:true" json:"structure"`   // 目录结构验证
	Content     bool `gorm:"default:false" json:"content"`    // 内容验证(抽样)
	MountTest   bool `gorm:"default:true" json:"mount_test"`  // 挂载测试
	RestoreTest bool `gorm:"default:false" json:"restore_test"` // 恢复测试

	// 执行状态
	Status      string     `gorm:"size:20" json:"status"` // pending, running, passed, failed, warning
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int        `json:"duration"`

	// 验证结果
	TotalChecks   int `json:"total_checks"`
	PassedChecks  int `json:"passed_checks"`
	FailedChecks  int `json:"failed_checks"`
	WarningChecks int `json:"warning_checks"`

	// 详细结果
	Results string `gorm:"type:text" json:"results"` // JSON数组

	// 评分
	Score int `json:"score"` // 0-100

	// 错误信息
	ErrorCode string `gorm:"size:50" json:"error_code"`
	ErrorMsg  string `gorm:"type:text" json:"error_msg"`

	CreatedAt time.Time `json:"created_at"`

	// 关联
	Checks []VerifyCheck `gorm:"foreignKey:TaskID" json:"checks,omitempty"`
}

// VerifyCheck 验证检查项
type VerifyCheck struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	TaskID uint   `gorm:"not null;index" json:"task_id"`

	// 检查项
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"size:500" json:"description"`
	CheckType   string `gorm:"size:50" json:"check_type"` // checksum, size, count, structure, content

	// 状态
	Status   string `gorm:"size:20" json:"status"` // pending, passed, failed, warning, skipped
	Duration int    `json:"duration"`

	// 预期值
	Expected string `gorm:"size:500" json:"expected"`

	// 实际值
	Actual string `gorm:"size:500" json:"actual"`

	// 详细信息
	Detail string `gorm:"type:text" json:"detail"`

	// 错误信息
	ErrorMsg string `gorm:"type:text" json:"error_msg"`

	CreatedAt time.Time `json:"created_at"`

	// 关联
	Task *VerifyTask `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

// ==================== 恢复脚本 ====================

// RecoveryScript 恢复脚本
type RecoveryScript struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Type        string    `gorm:"size:20;not null" json:"type"` // backup, restore, verify, drill

	// 脚本内容
	Language    string `gorm:"size:20;default:'bash'" json:"language"` // bash, python, go
	Script      string `gorm:"type:longtext" json:"script"`
	Timeout     int    `gorm:"default:300" json:"timeout"` // 超时时间(秒)

	// 参数
	Parameters string `gorm:"type:text" json:"parameters"` // JSON数组,参数定义

	// 执行配置
	PreScript  string `gorm:"type:text" json:"pre_script"`  // 前置脚本
	PostScript string `gorm:"type:text" json:"post_script"` // 后置脚本
	IgnoreError bool  `gorm:"default:false" json:"ignore_error"` // 是否忽略错误

	// 适用范围
	TargetTypes string `gorm:"type:text" json:"target_types"` // JSON数组,适用的目标类型

	// 验证
	VerifyScript string `gorm:"type:text" json:"verify_script"` // 验证脚本

	// 统计
	ExecCount    int  `gorm:"default:0" json:"exec_count"`
	SuccessCount int  `gorm:"default:0" json:"success_count"`
	FailedCount  int  `gorm:"default:0" json:"failed_count"`
	AvgDuration  int  `json:"avg_duration"` // 平均执行时间(秒)

	Enabled      bool  `gorm:"default:true" json:"enabled"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Executions []ScriptExecution `gorm:"foreignKey:ScriptID" json:"executions,omitempty"`
}

// ScriptExecution 脚本执行记录
type ScriptExecution struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	ScriptID uint   `gorm:"not null;index" json:"script_id"`
	ExecID   string `gorm:"size:50;uniqueIndex" json:"exec_id"`

	// 触发信息
	TriggerType string `gorm:"size:20" json:"trigger_type"` // manual, auto, schedule, drill
	TriggerBy   string `gorm:"size:100" json:"trigger_by"`

	// 目标信息
	TargetID   uint   `json:"target_id"`
	TargetType string `gorm:"size:50" json:"target_type"`

	// 执行参数
	Parameters string `gorm:"type:text" json:"parameters"` // JSON

	// 执行状态
	Status      string     `gorm:"size:20" json:"status"` // pending, running, success, failed, timeout
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int        `json:"duration"`

	// 执行结果
	ExitCode int    `json:"exit_code"`
	Output   string `gorm:"type:text" json:"output"`
	Error    string `gorm:"type:text" json:"error"`

	// 验证结果
	VerifyStatus string `gorm:"size:20" json:"verify_status"` // pending, passed, failed
	VerifyOutput string `gorm:"type:text" json:"verify_output"`

	CreatedAt time.Time `json:"created_at"`

	// 关联
	Script *RecoveryScript `gorm:"foreignKey:ScriptID" json:"script,omitempty"`
}

// ==================== 备份目标 ====================

// BackupTarget 备份目标
type BackupTarget struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Type        string    `gorm:"size:20;not null" json:"type"` // database, filesystem, application, vm
	Description string    `gorm:"size:500" json:"description"`

	// 连接配置
	Host        string `gorm:"size:100" json:"host"`
	Port        int    `json:"port"`
	Username    string `gorm:"size:100" json:"username"`
	Password    string `gorm:"size:200" json:"password"` // 加密存储
	PrivateKey  string `gorm:"type:text" json:"private_key"` // 加密存储

	// 数据库特定配置
	DbType     string `gorm:"size:20" json:"db_type"`     // mysql, postgresql, mongodb, redis
	DbName     string `gorm:"size:100" json:"db_name"`
	DbConfig   string `gorm:"type:text" json:"db_config"` // JSON

	// 文件系统配置
	RootPath    string `gorm:"size:500" json:"root_path"`
	ExcludePaths string `gorm:"type:text" json:"exclude_paths"`

	// 应用配置
	AppType     string `gorm:"size:50" json:"app_type"` // web, api, microservice
	AppConfig   string `gorm:"type:text" json:"app_config"`

	// 状态
	Status      string `gorm:"size:20;default:'active'" json:"status"` // active, inactive, error
	LastCheck   *time.Time `json:"last_check"`
	CheckResult string `gorm:"type:text" json:"check_result"`

	// 业务信息
	Owner       string `gorm:"size:100" json:"owner"`
	Department  string `gorm:"size:100" json:"department"`
	Priority    string `gorm:"size:20;default:'medium'" json:"priority"` // high, medium, low

	// RTO/RPO 要求
	TargetRTO   int `gorm:"default:60" json:"target_rto"`  // 目标RTO(分钟)
	TargetRPO   int `gorm:"default:5" json:"target_rpo"`   // 目标RPO(分钟)

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Policies []BackupPolicy `gorm:"foreignKey:TargetID" json:"policies,omitempty"`
}

// ==================== 存储配置 ====================

// BackupStorage 备份存储
type BackupStorage struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Type        string    `gorm:"size:20;not null" json:"type"` // local, s3, oss, nfs, ftp, sftp
	Description string    `gorm:"size:500" json:"description"`

	// 存储配置
	Config      string `gorm:"type:text" json:"config"` // JSON配置(端点、认证等)
	Path        string `gorm:"size:500" json:"path"`    // 存储路径

	// 本地存储
	LocalPath   string `gorm:"size:500" json:"local_path"`

	// S3/OSS配置
	Endpoint    string `gorm:"size:200" json:"endpoint"`
	Region      string `gorm:"size:50" json:"region"`
	Bucket      string `gorm:"size:100" json:"bucket"`
	AccessKey   string `gorm:"size:100" json:"access_key"`
	SecretKey   string `gorm:"size:100" json:"secret_key"` // 加密存储

	// NFS配置
	NfsServer   string `gorm:"size:100" json:"nfs_server"`
	NfsPath     string `gorm:"size:500" json:"nfs_path"`
	MountPoint  string `gorm:"size:500" json:"mount_point"`

	// FTP/SFTP配置
	FtpHost     string `gorm:"size:100" json:"ftp_host"`
	FtpPort     int    `json:"ftp_port"`
	FtpUser     string `gorm:"size:100" json:"ftp_user"`
	FtpPassword string `gorm:"size:200" json:"ftp_password"` // 加密存储

	// 容量管理
	MaxCapacity int64 `json:"max_capacity"` // 最大容量(GB), 0表示不限制
	UsedSpace   int64 `json:"used_space"`   // 已用空间(GB)
	AlertThreshold int `gorm:"default:80" json:"alert_threshold"` // 告警阈值(%)

	// 状态
	Status      string `gorm:"size:20;default:'active'" json:"status"` // active, inactive, error
	LastCheck   *time.Time `json:"last_check"`
	CheckResult string `gorm:"type:text" json:"check_result"`

	Enabled      bool `gorm:"default:true" json:"enabled"`
	IsDefault    bool `gorm:"default:false" json:"is_default"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
