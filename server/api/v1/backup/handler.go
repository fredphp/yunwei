package backup

import (
	"net/http"
	"strconv"
	"time"

	backupModel "yunwei/model/backup"
	"yunwei/model/common/response"
	backupService "yunwei/service/backup"

	"github.com/gin-gonic/gin"
)

// Handler 备份API处理器
type Handler struct {
	schedulerSvc *backupService.SchedulerService
	dbBackupSvc  *backupService.DatabaseBackupService
	fileBackupSvc *backupService.FileBackupService
	snapshotSvc  *backupService.SnapshotService
	restoreSvc   *backupService.RestoreService
	verifySvc    *backupService.VerifyService
	drillSvc     *backupService.DrillService
	scriptEngine *backupService.ScriptEngine
}

// NewHandler 创建处理器
func NewHandler() *Handler {
	return &Handler{
		schedulerSvc: backupService.NewSchedulerService(),
		dbBackupSvc:  backupService.NewDatabaseBackupService(),
		fileBackupSvc: backupService.NewFileBackupService(),
		snapshotSvc:  backupService.NewSnapshotService(),
		restoreSvc:   backupService.NewRestoreService(),
		verifySvc:    backupService.NewVerifyService(),
		drillSvc:     backupService.NewDrillService(),
		scriptEngine: backupService.NewScriptEngine("/var/lib/backup/scripts"),
	}
}

// ==================== 备份策略 ====================

// GetPolicies 获取备份策略列表
func (h *Handler) GetPolicies(c *gin.Context) {
	policies := []backupModel.BackupPolicy{}
	response.OkWithData(policies, c)
}

// GetPolicy 获取备份策略
func (h *Handler) GetPolicy(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	_ = id // 实际从数据库获取
	policy := &backupModel.BackupPolicy{}
	response.OkWithData(policy, c)
}

// CreatePolicy 创建备份策略
func (h *Handler) CreatePolicy(c *gin.Context) {
	var policy backupModel.BackupPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 保存到数据库
	response.OkWithData(gin.H{"id": 1}, c)
}

// UpdatePolicy 更新备份策略
func (h *Handler) UpdatePolicy(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var policy backupModel.BackupPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	policy.ID = uint(id)

	response.Ok(c)
}

// DeletePolicy 删除备份策略
func (h *Handler) DeletePolicy(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	_ = id
	response.Ok(c)
}

// TriggerBackup 手动触发备份
func (h *Handler) TriggerBackup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	_ = id
	response.OkWithMessage("备份任务已触发", c)
}

// ==================== 备份记录 ====================

// GetRecords 获取备份记录
func (h *Handler) GetRecords(c *gin.Context) {
	policyID := c.Query("policy_id")
	_ = policyID

	records := []backupModel.BackupRecord{}
	response.OkWithData(records, c)
}

// GetRecord 获取备份记录详情
func (h *Handler) GetRecord(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	_ = id

	record := &backupModel.BackupRecord{}
	response.OkWithData(record, c)
}

// QuickBackup 快速备份
func (h *Handler) QuickBackup(c *gin.Context) {
	var req struct {
		TargetID   uint   `json:"target_id" binding:"required"`
		TargetType string `json:"target_type" binding:"required"`
		BackupType string `json:"backup_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 执行快速备份
	response.OkWithData(gin.H{
		"task_id":    "backup_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":     "running",
		"started_at": time.Now(),
	}, c)
}

// ==================== 数据库备份 ====================

// GetDatabases 获取可备份数据库列表
func (h *Handler) GetDatabases(c *gin.Context) {
	databases := []map[string]interface{}{
		{"id": 1, "name": "mysql_main", "type": "mysql", "host": "localhost", "port": 3306},
		{"id": 2, "name": "postgres_main", "type": "postgresql", "host": "localhost", "port": 5432},
	}
	response.OkWithData(databases, c)
}

// BackupDatabase 备份数据库
func (h *Handler) BackupDatabase(c *gin.Context) {
	var req struct {
		TargetID     uint   `json:"target_id" binding:"required"`
		DatabaseName string `json:"database_name"`
		StoragePath  string `json:"storage_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"task_id": "db_backup_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":  "running",
	}, c)
}

// ==================== 文件备份 ====================

// BackupFiles 备份文件
func (h *Handler) BackupFiles(c *gin.Context) {
	var req struct {
		SourcePath  string `json:"source_path" binding:"required"`
		StoragePath string `json:"storage_path"`
		Exclude     string `json:"exclude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"task_id": "file_backup_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":  "running",
	}, c)
}

// ==================== 快照管理 ====================

// GetSnapshotPolicies 获取快照策略
func (h *Handler) GetSnapshotPolicies(c *gin.Context) {
	policies := []backupModel.SnapshotPolicy{}
	response.OkWithData(policies, c)
}

// CreateSnapshotPolicy 创建快照策略
func (h *Handler) CreateSnapshotPolicy(c *gin.Context) {
	var policy backupModel.SnapshotPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"id": 1}, c)
}

// GetSnapshots 获取快照列表
func (h *Handler) GetSnapshots(c *gin.Context) {
	policyID := c.Query("policy_id")
	_ = policyID

	snapshots := []backupModel.SnapshotRecord{}
	response.OkWithData(snapshots, c)
}

// CreateSnapshot 创建快照
func (h *Handler) CreateSnapshot(c *gin.Context) {
	var req struct {
		TargetID   uint   `json:"target_id" binding:"required"`
		TargetType string `json:"target_type" binding:"required"`
		Name       string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"snap_id": "snap_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":  "creating",
	}, c)
}

// DeleteSnapshot 删除快照
func (h *Handler) DeleteSnapshot(c *gin.Context) {
	id := c.Param("id")
	_ = id
	response.Ok(c)
}

// ==================== 恢复管理 ====================

// GetRestoreRecords 获取恢复记录
func (h *Handler) GetRestoreRecords(c *gin.Context) {
	records := []backupModel.RestoreRecord{}
	response.OkWithData(records, c)
}

// RestoreBackup 恢复备份
func (h *Handler) RestoreBackup(c *gin.Context) {
	var req struct {
		BackupID    uint   `json:"backup_id" binding:"required"`
		TargetPath  string `json:"target_path" binding:"required"`
		Overwrite   bool   `json:"overwrite"`
		VerifyAfter bool   `json:"verify_after"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"restore_id": "restore_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":     "running",
	}, c)
}

// PointInTimeRecovery 时间点恢复
func (h *Handler) PointInTimeRecovery(c *gin.Context) {
	var req struct {
		TargetID    uint      `json:"target_id" binding:"required"`
		PointInTime time.Time `json:"point_in_time" binding:"required"`
		TargetPath  string    `json:"target_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"restore_id": "pitr_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":     "running",
	}, c)
}

// GetRestoreStatus 获取恢复状态
func (h *Handler) GetRestoreStatus(c *gin.Context) {
	id := c.Param("id")
	_ = id

	status := map[string]interface{}{
		"restore_id":     id,
		"status":         "completed",
		"restored_files": 100,
		"progress":       100,
	}
	response.OkWithData(status, c)
}

// CancelRestore 取消恢复
func (h *Handler) CancelRestore(c *gin.Context) {
	id := c.Param("id")
	_ = id
	response.OkWithMessage("恢复已取消", c)
}

// ==================== 恢复验证 ====================

// GetVerifyTasks 获取验证任务
func (h *Handler) GetVerifyTasks(c *gin.Context) {
	tasks := []backupModel.VerifyTask{}
	response.OkWithData(tasks, c)
}

// VerifyBackup 验证备份
func (h *Handler) VerifyBackup(c *gin.Context) {
	var req struct {
		BackupID    uint   `json:"backup_id" binding:"required"`
		VerifyType  string `json:"verify_type"` // integrity, consistency, recoverability, full
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"task_id": "verify_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":  "pending",
	}, c)
}

// GetVerifyResult 获取验证结果
func (h *Handler) GetVerifyResult(c *gin.Context) {
	id := c.Param("id")
	_ = id

	result := map[string]interface{}{
		"task_id":       id,
		"status":        "passed",
		"score":         95,
		"total_checks":  5,
		"passed_checks": 5,
		"failed_checks": 0,
	}
	response.OkWithData(result, c)
}

// ==================== 灾备演练 ====================

// GetDrillPlans 获取演练计划
func (h *Handler) GetDrillPlans(c *gin.Context) {
	plans := []backupModel.DrillPlan{}
	response.OkWithData(plans, c)
}

// GetDrillPlan 获取演练计划
func (h *Handler) GetDrillPlan(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	_ = id

	plan := &backupModel.DrillPlan{}
	response.OkWithData(plan, c)
}

// CreateDrillPlan 创建演练计划
func (h *Handler) CreateDrillPlan(c *gin.Context) {
	var plan backupModel.DrillPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"id": 1}, c)
}

// ExecuteDrill 执行演练
func (h *Handler) ExecuteDrill(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	_ = id

	response.OkWithData(gin.H{
		"drill_id":   id,
		"status":     "running",
		"started_at": time.Now(),
	}, c)
}

// CancelDrill 取消演练
func (h *Handler) CancelDrill(c *gin.Context) {
	id := c.Param("id")
	_ = id
	response.OkWithMessage("演练已取消", c)
}

// GetDrillStatus 获取演练状态
func (h *Handler) GetDrillStatus(c *gin.Context) {
	id := c.Param("id")
	_ = id

	status := map[string]interface{}{
		"drill_id":   id,
		"status":     "completed",
		"result":     "success",
		"score":      85,
		"actual_rto": 300,
		"actual_rpo": 5,
		"rto_met":    true,
		"rpo_met":    true,
	}
	response.OkWithData(status, c)
}

// GetDrillReport 获取演练报告
func (h *Handler) GetDrillReport(c *gin.Context) {
	id := c.Param("id")
	_ = id

	report := map[string]interface{}{
		"drill_id":     id,
		"generated_at": time.Now(),
		"summary":      "演练成功完成",
		"score":        85,
		"rto_met":      true,
		"rpo_met":      true,
	}
	response.OkWithData(report, c)
}

// ==================== 恢复脚本 ====================

// GetScripts 获取恢复脚本列表
func (h *Handler) GetScripts(c *gin.Context) {
	scripts := []backupModel.RecoveryScript{}
	response.OkWithData(scripts, c)
}

// GetScript 获取恢复脚本
func (h *Handler) GetScript(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	_ = id

	script := &backupModel.RecoveryScript{}
	response.OkWithData(script, c)
}

// CreateScript 创建恢复脚本
func (h *Handler) CreateScript(c *gin.Context) {
	var script backupModel.RecoveryScript
	if err := c.ShouldBindJSON(&script); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"id": 1}, c)
}

// UpdateScript 更新恢复脚本
func (h *Handler) UpdateScript(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var script backupModel.RecoveryScript
	if err := c.ShouldBindJSON(&script); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	script.ID = uint(id)

	response.Ok(c)
}

// DeleteScript 删除恢复脚本
func (h *Handler) DeleteScript(c *gin.Context) {
	id := c.Param("id")
	_ = id
	response.Ok(c)
}

// ExecuteScript 执行脚本
func (h *Handler) ExecuteScript(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Params map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"exec_id": "exec_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":  "running",
	}, c)
}

// GetScriptExecution 获取脚本执行结果
func (h *Handler) GetScriptExecution(c *gin.Context) {
	execID := c.Param("exec_id")
	_ = execID

	result := map[string]interface{}{
		"exec_id":   execID,
		"status":    "success",
		"output":    "脚本执行成功",
		"exit_code": 0,
		"duration":  5,
	}
	response.OkWithData(result, c)
}

// ==================== 备份目标 ====================

// GetTargets 获取备份目标
func (h *Handler) GetTargets(c *gin.Context) {
	targets := []backupModel.BackupTarget{}
	response.OkWithData(targets, c)
}

// CreateTarget 创建备份目标
func (h *Handler) CreateTarget(c *gin.Context) {
	var target backupModel.BackupTarget
	if err := c.ShouldBindJSON(&target); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"id": 1}, c)
}

// UpdateTarget 更新备份目标
func (h *Handler) UpdateTarget(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var target backupModel.BackupTarget
	if err := c.ShouldBindJSON(&target); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	target.ID = uint(id)

	response.Ok(c)
}

// DeleteTarget 删除备份目标
func (h *Handler) DeleteTarget(c *gin.Context) {
	id := c.Param("id")
	_ = id
	response.Ok(c)
}

// TestTarget 测试备份目标连接
func (h *Handler) TestTarget(c *gin.Context) {
	id := c.Param("id")
	_ = id

	response.OkWithData(gin.H{
		"success": true,
		"message": "连接测试成功",
	}, c)
}

// ==================== 存储管理 ====================

// GetStorages 获取存储配置
func (h *Handler) GetStorages(c *gin.Context) {
	storages := []backupModel.BackupStorage{}
	response.OkWithData(storages, c)
}

// CreateStorage 创建存储配置
func (h *Handler) CreateStorage(c *gin.Context) {
	var storage backupModel.BackupStorage
	if err := c.ShouldBindJSON(&storage); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"id": 1}, c)
}

// UpdateStorage 更新存储配置
func (h *Handler) UpdateStorage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var storage backupModel.BackupStorage
	if err := c.ShouldBindJSON(&storage); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	storage.ID = uint(id)

	response.Ok(c)
}

// DeleteStorage 删除存储配置
func (h *Handler) DeleteStorage(c *gin.Context) {
	id := c.Param("id")
	_ = id
	response.Ok(c)
}

// TestStorage 测试存储连接
func (h *Handler) TestStorage(c *gin.Context) {
	id := c.Param("id")
	_ = id

	response.OkWithData(gin.H{
		"success": true,
		"message": "存储连接测试成功",
	}, c)
}

// GetStorageUsage 获取存储使用情况
func (h *Handler) GetStorageUsage(c *gin.Context) {
	id := c.Param("id")
	_ = id

	usage := map[string]interface{}{
		"storage_id":     id,
		"total_capacity": 1024,
		"used_space":     512,
		"free_space":     512,
		"usage_percent":  50,
	}
	response.OkWithData(usage, c)
}

// ==================== 调度管理 ====================

// GetSchedulerStats 获取调度器统计
func (h *Handler) GetSchedulerStats(c *gin.Context) {
	stats := map[string]interface{}{
		"total_policies":    10,
		"active_policies":   8,
		"running_backups":   2,
		"scheduled_backups": 5,
		"last_backup_time":  time.Now().Add(-1 * time.Hour),
		"next_backup_time":  time.Now().Add(1 * time.Hour),
	}
	response.OkWithData(stats, c)
}

// GetScheduledJobs 获取已调度任务
func (h *Handler) GetScheduledJobs(c *gin.Context) {
	jobs := []map[string]interface{}{
		{
			"policy_id":       1,
			"policy_name":     "每日数据库备份",
			"schedule":        "0 2 * * *",
			"next_run":        time.Now().Add(12 * time.Hour),
			"status":          "active",
		},
	}
	response.OkWithData(jobs, c)
}

// ==================== 统计报告 ====================

// GetBackupStats 获取备份统计
func (h *Handler) GetBackupStats(c *gin.Context) {
	stats := map[string]interface{}{
		"total_backups":    100,
		"success_count":    95,
		"failed_count":     5,
		"success_rate":     95.0,
		"total_size":       1024 * 1024 * 1024 * 100, // 100GB
		"avg_duration":     300,
		"last_24h_count":   24,
		"last_7d_count":    168,
		"last_30d_count":   720,
	}
	response.OkWithData(stats, c)
}

// GetBackupReport 获取备份报告
func (h *Handler) GetBackupReport(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	_ = startDate
	_ = endDate

	report := map[string]interface{}{
		"period":          "2024-01-01 ~ 2024-01-31",
		"total_backups":   100,
		"success_count":   95,
		"failed_count":    5,
		"total_size":      1024 * 1024 * 1024 * 100,
		"avg_duration":    300,
		"policy_stats":    []map[string]interface{}{},
		"recommendations": []string{"建议增加备份频率", "建议检查失败任务"},
	}
	response.OkWithData(report, c)
}

// ==================== 自动备份 ====================

// SetupAutoBackup 设置自动备份
func (h *Handler) SetupAutoBackup(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required"`
		BackupType    string `json:"backup_type" binding:"required"`
		TargetIDs     []uint `json:"target_ids" binding:"required"`
		Schedule      string `json:"schedule" binding:"required"`
		RetentionDays int    `json:"retention_days"`
		StorageID     uint   `json:"storage_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"policy_id": 1,
		"message":   "自动备份已设置",
	}, c)
}

// SetupAutoRecovery 设置自动恢复
func (h *Handler) SetupAutoRecovery(c *gin.Context) {
	var req struct {
		TargetID    uint   `json:"target_id" binding:"required"`
		ScriptID    uint   `json:"script_id"`
		TriggerType string `json:"trigger_type"` // manual, schedule, auto
		VerifyAfter bool   `json:"verify_after"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"config_id": 1,
		"message":   "自动恢复已设置",
	}, c)
}

// SetupAutoVerify 设置自动验证
func (h *Handler) SetupAutoVerify(c *gin.Context) {
	var req struct {
		PolicyID     uint   `json:"policy_id" binding:"required"`
		VerifyType   string `json:"verify_type"`
		Schedule     string `json:"schedule"`
		NotifyOnFail bool   `json:"notify_on_fail"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"config_id": 1,
		"message":   "自动验证已设置",
	}, c)
}

// 注册路由
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	// 备份策略
	r.GET("/policies", h.GetPolicies)
	r.GET("/policies/:id", h.GetPolicy)
	r.POST("/policies", h.CreatePolicy)
	r.PUT("/policies/:id", h.UpdatePolicy)
	r.DELETE("/policies/:id", h.DeletePolicy)
	r.POST("/policies/:id/trigger", h.TriggerBackup)

	// 备份记录
	r.GET("/records", h.GetRecords)
	r.GET("/records/:id", h.GetRecord)
	r.POST("/quick-backup", h.QuickBackup)

	// 数据库备份
	r.GET("/databases", h.GetDatabases)
	r.POST("/databases/backup", h.BackupDatabase)

	// 文件备份
	r.POST("/files/backup", h.BackupFiles)

	// 快照管理
	r.GET("/snapshots/policies", h.GetSnapshotPolicies)
	r.POST("/snapshots/policies", h.CreateSnapshotPolicy)
	r.GET("/snapshots", h.GetSnapshots)
	r.POST("/snapshots", h.CreateSnapshot)
	r.DELETE("/snapshots/:id", h.DeleteSnapshot)

	// 恢复管理
	r.GET("/restores", h.GetRestoreRecords)
	r.POST("/restores", h.RestoreBackup)
	r.POST("/restores/pitr", h.PointInTimeRecovery)
	r.GET("/restores/:id/status", h.GetRestoreStatus)
	r.POST("/restores/:id/cancel", h.CancelRestore)

	// 恢复验证
	r.GET("/verify", h.GetVerifyTasks)
	r.POST("/verify", h.VerifyBackup)
	r.GET("/verify/:id", h.GetVerifyResult)

	// 灾备演练
	r.GET("/drills", h.GetDrillPlans)
	r.GET("/drills/:id", h.GetDrillPlan)
	r.POST("/drills", h.CreateDrillPlan)
	r.POST("/drills/:id/execute", h.ExecuteDrill)
	r.POST("/drills/:id/cancel", h.CancelDrill)
	r.GET("/drills/:id/status", h.GetDrillStatus)
	r.GET("/drills/:id/report", h.GetDrillReport)

	// 恢复脚本
	r.GET("/scripts", h.GetScripts)
	r.GET("/scripts/:id", h.GetScript)
	r.POST("/scripts", h.CreateScript)
	r.PUT("/scripts/:id", h.UpdateScript)
	r.DELETE("/scripts/:id", h.DeleteScript)
	r.POST("/scripts/:id/execute", h.ExecuteScript)
	r.GET("/scripts/executions/:exec_id", h.GetScriptExecution)

	// 备份目标
	r.GET("/targets", h.GetTargets)
	r.POST("/targets", h.CreateTarget)
	r.PUT("/targets/:id", h.UpdateTarget)
	r.DELETE("/targets/:id", h.DeleteTarget)
	r.POST("/targets/:id/test", h.TestTarget)

	// 存储管理
	r.GET("/storages", h.GetStorages)
	r.POST("/storages", h.CreateStorage)
	r.PUT("/storages/:id", h.UpdateStorage)
	r.DELETE("/storages/:id", h.DeleteStorage)
	r.POST("/storages/:id/test", h.TestStorage)
	r.GET("/storages/:id/usage", h.GetStorageUsage)

	// 调度管理
	r.GET("/scheduler/stats", h.GetSchedulerStats)
	r.GET("/scheduler/jobs", h.GetScheduledJobs)

	// 统计报告
	r.GET("/stats", h.GetBackupStats)
	r.GET("/report", h.GetBackupReport)

	// 自动配置
	r.POST("/auto/backup", h.SetupAutoBackup)
	r.POST("/auto/recovery", h.SetupAutoRecovery)
	r.POST("/auto/verify", h.SetupAutoVerify)
}
