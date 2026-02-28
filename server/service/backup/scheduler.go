package backup

import (
        "context"
        "encoding/json"
        "fmt"
        "sync"
        "time"

        "yunwei/model/backup"

        "github.com/robfig/cron/v3"
)

// SchedulerService 备份调度服务
type SchedulerService struct {
        cron           *cron.Cron
        mu             sync.RWMutex
        jobs           map[uint]cron.EntryID
        activeBackups  map[uint]*BackupContext
        dbBackupSvc    *DatabaseBackupService
        fileBackupSvc  *FileBackupService
        snapshotSvc    *SnapshotService
        notifySvc      *NotifyService
}

// BackupContext 备份上下文
type BackupContext struct {
        PolicyID  uint
        Status    string
        StartTime time.Time
        Progress  int
        Cancelled bool
}

// NewSchedulerService 创建调度服务
func NewSchedulerService() *SchedulerService {
        return &SchedulerService{
                cron:          cron.New(cron.WithSeconds()),
                jobs:          make(map[uint]cron.EntryID),
                activeBackups: make(map[uint]*BackupContext),
                dbBackupSvc:   NewDatabaseBackupService(),
                fileBackupSvc: NewFileBackupService(),
                snapshotSvc:   NewSnapshotService(),
                notifySvc:     NewNotifyService(),
        }
}

// Start 启动调度器
func (s *SchedulerService) Start() {
        s.cron.Start()
}

// Stop 停止调度器
func (s *SchedulerService) Stop() {
        s.cron.Stop()
}

// SchedulePolicy 调度备份策略
func (s *SchedulerService) SchedulePolicy(policy *backup.BackupPolicy) error {
        s.mu.Lock()
        defer s.mu.Unlock()

        // 如果已有调度，先移除
        if entryID, exists := s.jobs[policy.ID]; exists {
                s.cron.Remove(entryID)
                delete(s.jobs, policy.ID)
        }

        if !policy.Enabled {
                return nil
        }

        var scheduleExpr string
        switch policy.ScheduleType {
        case "cron":
                scheduleExpr = policy.ScheduleExpr
        case "interval":
                // 转换间隔分钟为cron表达式
                scheduleExpr = fmt.Sprintf("*/%d * * * *", policy.IntervalMinute)
        default:
                return fmt.Errorf("不支持的调度类型: %s", policy.ScheduleType)
        }

        // 添加定时任务
        entryID, err := s.cron.AddFunc(scheduleExpr, func() {
                s.executeBackup(policy)
        })
        if err != nil {
                return fmt.Errorf("添加定时任务失败: %v", err)
        }

        s.jobs[policy.ID] = entryID
        return nil
}

// UnschedulePolicy 取消调度
func (s *SchedulerService) UnschedulePolicy(policyID uint) {
        s.mu.Lock()
        defer s.mu.Unlock()

        if entryID, exists := s.jobs[policyID]; exists {
                s.cron.Remove(entryID)
                delete(s.jobs, policyID)
        }
}

// executeBackup 执行备份
func (s *SchedulerService) executeBackup(policy *backup.BackupPolicy) {
        ctx := context.Background()

        s.mu.Lock()
        s.activeBackups[policy.ID] = &BackupContext{
                PolicyID:  policy.ID,
                Status:    "running",
                StartTime: time.Now(),
        }
        s.mu.Unlock()

        defer func() {
                s.mu.Lock()
                delete(s.activeBackups, policy.ID)
                s.mu.Unlock()
        }()

        // 创建备份目标
        target := &backup.BackupTarget{
                ID:       policy.TargetID,
                Name:     policy.TargetName,
                Type:     policy.TargetType,
                DbConfig: policy.SourceConfig,
                RootPath: policy.SourcePath,
        }

        var result interface{}
        var err error

        switch policy.Type {
        case "database":
                dbResult, dbErr := s.dbBackupSvc.Execute(ctx, policy, target)
                result = dbResult
                err = dbErr
        case "file", "filesystem":
                fileResult, fileErr := s.fileBackupSvc.Execute(ctx, policy, target)
                result = fileResult
                err = fileErr
        case "snapshot":
                snapResult, snapErr := s.snapshotSvc.CreateSnapshot(ctx, &backup.SnapshotPolicy{
                        ID:           policy.ID,
                        Name:         policy.Name,
                        SnapshotType: policy.SourceType,
                }, target)
                result = snapResult
                err = snapErr
        default:
                err = fmt.Errorf("不支持的备份类型: %s", policy.Type)
        }

        // 发送通知
        if err != nil {
                if policy.NotifyOnFail {
                        s.notifySvc.NotifyBackupFailed(policy, err)
                }
        } else {
                if policy.NotifyOnSuccess {
                        if backupResult, ok := result.(*BackupResult); ok {
                                s.notifySvc.NotifyBackupSuccess(policy, backupResult)
                        }
                }
        }
}

// TriggerBackup 手动触发备份
func (s *SchedulerService) TriggerBackup(policyID uint) error {
        s.mu.RLock()
        if _, exists := s.activeBackups[policyID]; exists {
                s.mu.RUnlock()
                return fmt.Errorf("备份任务已在运行中")
        }
        s.mu.RUnlock()

        // 实际应该从数据库获取策略
        // 这里简化处理
        go s.executeBackup(&backup.BackupPolicy{
                ID:         policyID,
                Type:       "file",
                TargetID:   1,
                TargetName: "default",
                TargetType: "filesystem",
        })

        return nil
}

// CancelBackup 取消备份
func (s *SchedulerService) CancelBackup(policyID uint) error {
        s.mu.Lock()
        defer s.mu.Unlock()

        if ctx, exists := s.activeBackups[policyID]; exists {
                ctx.Cancelled = true
                return nil
        }

        return fmt.Errorf("备份任务不存在或已完成")
}

// GetBackupStatus 获取备份状态
func (s *SchedulerService) GetBackupStatus(policyID uint) (*BackupContext, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        if ctx, exists := s.activeBackups[policyID]; exists {
                return ctx, nil
        }

        return nil, fmt.Errorf("备份任务不存在")
}

// ListScheduledPolicies 列出已调度的策略
func (s *SchedulerService) ListScheduledPolicies() []uint {
        s.mu.RLock()
        defer s.mu.RUnlock()

        ids := make([]uint, 0, len(s.jobs))
        for id := range s.jobs {
                ids = append(ids, id)
        }

        return ids
}

// ListActiveBackups 列出活动备份
func (s *SchedulerService) ListActiveBackups() []*BackupContext {
        s.mu.RLock()
        defer s.mu.RUnlock()

        contexts := make([]*BackupContext, 0, len(s.activeBackups))
        for _, ctx := range s.activeBackups {
                contexts = append(contexts, ctx)
        }

        return contexts
}

// ==================== 自动备份策略管理 ====================

// AutoBackupPolicy 自动备份策略配置
type AutoBackupPolicy struct {
        ID               uint              `json:"id"`
        Name             string            `json:"name"`
        Enabled          bool              `json:"enabled"`
        BackupType       string            `json:"backup_type"` // database, file, full
        Schedule         ScheduleConfig    `json:"schedule"`
        Retention        RetentionConfig   `json:"retention"`
        Targets          []BackupTargetRef `json:"targets"`
        Storage          StorageRef        `json:"storage"`
        Notifications    NotificationConfig `json:"notifications"`
        PreBackupScript  string            `json:"pre_backup_script"`
        PostBackupScript string            `json:"post_backup_script"`
}

// ScheduleConfig 调度配置
type ScheduleConfig struct {
        Type     string `json:"type"`     // daily, weekly, monthly, interval
        Time     string `json:"time"`     // HH:MM
        DayOfWeek int    `json:"day_of_week"` // 0-6 (周日=0)
        DayOfMonth int   `json:"day_of_month"` // 1-31
        IntervalMinutes int `json:"interval_minutes"`
}

// RetentionConfig 保留配置
type RetentionConfig struct {
        Days       int  `json:"days"`       // 保留天数
        MaxCount   int  `json:"max_count"`  // 最大保留数量
        MinCount   int  `json:"min_count"`  // 最小保留数量
        KeepFirst  bool `json:"keep_first"` // 保留第一个备份
        KeepLast   bool `json:"keep_last"`  // 保留最后一个备份
}

// BackupTargetRef 备份目标引用
type BackupTargetRef struct {
        ID       uint   `json:"id"`
        Type     string `json:"type"`     // database, filesystem
        Name     string `json:"name"`
        Priority int    `json:"priority"` // 备份优先级
}

// StorageRef 存储引用
type StorageRef struct {
        ID       uint   `json:"id"`
        Type     string `json:"type"` // local, s3, oss
        Path     string `json:"path"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
        OnSuccess bool     `json:"on_success"`
        OnFailure bool     `json:"on_failure"`
        Channels  []string `json:"channels"` // email, sms, webhook
}

// CreateAutoBackupPolicy 创建自动备份策略
func (s *SchedulerService) CreateAutoBackupPolicy(config AutoBackupPolicy) error {
        // 验证配置
        if err := s.validateAutoBackupPolicy(config); err != nil {
                return err
        }

        // 创建备份策略
        policy := &backup.BackupPolicy{
                Name:          config.Name,
                Type:          config.BackupType,
                Enabled:       config.Enabled,
                ScheduleType:  config.Schedule.Type,
                StorageType:   config.Storage.Type,
                StoragePath:   config.Storage.Path,
                NotifyOnSuccess: config.Notifications.OnSuccess,
                NotifyOnFail:   config.Notifications.OnFailure,
                PreScript:     config.PreBackupScript,
                PostScript:    config.PostBackupScript,
        }

        // 根据调度类型设置cron表达式
        switch config.Schedule.Type {
        case "daily":
                policy.ScheduleExpr = s.buildDailyCron(config.Schedule.Time)
        case "weekly":
                policy.ScheduleExpr = s.buildWeeklyCron(config.Schedule.Time, config.Schedule.DayOfWeek)
        case "monthly":
                policy.ScheduleExpr = s.buildMonthlyCron(config.Schedule.Time, config.Schedule.DayOfMonth)
        case "interval":
                policy.IntervalMinute = config.Schedule.IntervalMinutes
                policy.ScheduleType = "interval"
        }

        // 设置保留策略
        policy.RetentionDays = config.Retention.Days
        policy.RetentionCount = config.Retention.MaxCount

        // 添加到调度器
        return s.SchedulePolicy(policy)
}

// validateAutoBackupPolicy 验证自动备份策略
func (s *SchedulerService) validateAutoBackupPolicy(config AutoBackupPolicy) error {
        if config.Name == "" {
                return fmt.Errorf("策略名称不能为空")
        }

        if len(config.Targets) == 0 {
                return fmt.Errorf("必须指定至少一个备份目标")
        }

        switch config.Schedule.Type {
        case "daily", "weekly", "monthly", "interval":
                // valid
        default:
                return fmt.Errorf("不支持的调度类型: %s", config.Schedule.Type)
        }

        return nil
}

// buildDailyCron 构建每日cron表达式
func (s *SchedulerService) buildDailyCron(timeStr string) string {
        // timeStr format: HH:MM
        var hour, minute int
        fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
        return fmt.Sprintf("%d %d * * *", minute, hour)
}

// buildWeeklyCron 构建每周cron表达式
func (s *SchedulerService) buildWeeklyCron(timeStr string, dayOfWeek int) string {
        var hour, minute int
        fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
        return fmt.Sprintf("%d %d * * %d", minute, hour, dayOfWeek)
}

// buildMonthlyCron 构建每月cron表达式
func (s *SchedulerService) buildMonthlyCron(timeStr string, dayOfMonth int) string {
        var hour, minute int
        fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
        return fmt.Sprintf("%d %d %d * *", minute, hour, dayOfMonth)
}

// ==================== 备份清理 ====================

// CleanupOldBackups 清理过期备份
func (s *SchedulerService) CleanupOldBackups(policyID uint, retentionDays, retentionCount int) error {
        // 获取策略的所有备份记录
        // 按时间排序
        // 删除超过保留天数或数量的备份

        // 简化实现
        return nil
}

// ==================== 备份状态监控 ====================

// BackupStats 备份统计
type BackupStats struct {
        TotalPolicies    int   `json:"total_policies"`
        ActivePolicies   int   `json:"active_policies"`
        TotalBackups     int64 `json:"total_backups"`
        TotalSize        int64 `json:"total_size"`
        SuccessRate      float64 `json:"success_rate"`
        LastBackupTime   *time.Time `json:"last_backup_time"`
        FailedBackups    int   `json:"failed_backups"`
        ScheduledBackups int   `json:"scheduled_backups"`
}

// GetBackupStats 获取备份统计
func (s *SchedulerService) GetBackupStats() *BackupStats {
        s.mu.RLock()
        defer s.mu.RUnlock()

        return &BackupStats{
                TotalPolicies:  len(s.jobs),
                ActivePolicies: len(s.activeBackups),
        }
}

// ==================== 备份队列管理 ====================

// BackupQueue 备份队列
type BackupQueue struct {
        mu       sync.Mutex
        items    []*BackupQueueItem
        cond     *sync.Cond
        running  bool
        workers  int
}

// BackupQueueItem 备份队列项
type BackupQueueItem struct {
        ID        uint
        PolicyID  uint
        Priority  int
        CreatedAt time.Time
        Status    string
}

// NewBackupQueue 创建备份队列
func NewBackupQueue(workers int) *BackupQueue {
        q := &BackupQueue{
                items:   make([]*BackupQueueItem, 0),
                workers: workers,
        }
        q.cond = sync.NewCond(&q.mu)
        return q
}

// Enqueue 入队
func (q *BackupQueue) Enqueue(item *BackupQueueItem) {
        q.mu.Lock()
        defer q.mu.Unlock()

        // 按优先级插入
        inserted := false
        for i, existing := range q.items {
                if item.Priority > existing.Priority {
                        q.items = append(q.items[:i], append([]*BackupQueueItem{item}, q.items[i:]...)...)
                        inserted = true
                        break
                }
        }

        if !inserted {
                q.items = append(q.items, item)
        }

        q.cond.Signal()
}

// Dequeue 出队
func (q *BackupQueue) Dequeue() *BackupQueueItem {
        q.mu.Lock()
        defer q.mu.Unlock()

        for len(q.items) == 0 && q.running {
                q.cond.Wait()
        }

        if len(q.items) == 0 {
                return nil
        }

        item := q.items[0]
        q.items = q.items[1:]
        return item
}

// Start 启动队列处理
func (q *BackupQueue) Start(scheduler *SchedulerService) {
        q.mu.Lock()
        q.running = true
        q.mu.Unlock()

        for i := 0; i < q.workers; i++ {
                go func() {
                        for {
                                item := q.Dequeue()
                                if item == nil {
                                        return
                                }

                                // 执行备份
                                scheduler.TriggerBackup(item.PolicyID)
                        }
                }()
        }
}

// Stop 停止队列处理
func (q *BackupQueue) Stop() {
        q.mu.Lock()
        q.running = false
        q.mu.Unlock()
        q.cond.Broadcast()
}

// ==================== 备份报告 ====================

// BackupReport 备份报告
type BackupReport struct {
        Period       string        `json:"period"`
        TotalBackups int           `json:"total_backups"`
        SuccessCount int           `json:"success_count"`
        FailedCount  int           `json:"failed_count"`
        TotalSize    int64         `json:"total_size"`
        Duration     int           `json:"duration"`
        PolicyStats  []PolicyStat  `json:"policy_stats"`
        Recommendations []string   `json:"recommendations"`
}

// PolicyStat 策略统计
type PolicyStat struct {
        PolicyID    uint   `json:"policy_id"`
        PolicyName  string `json:"policy_name"`
        BackupCount int    `json:"backup_count"`
        SuccessRate float64 `json:"success_rate"`
        AvgDuration int    `json:"avg_duration"`
        TotalSize   int64  `json:"total_size"`
}

// GenerateBackupReport 生成备份报告
func (s *SchedulerService) GenerateBackupReport(startTime, endTime time.Time) (*BackupReport, error) {
        report := &BackupReport{
                Period:        fmt.Sprintf("%s ~ %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
                Recommendations: make([]string, 0),
        }

        // 实际应该从数据库查询统计数据
        // 简化实现

        // 添加建议
        if report.FailedCount > 0 {
                report.Recommendations = append(report.Recommendations, "建议检查失败的备份任务并修复问题")
        }

        return report, nil
}

// GetNextBackupTime 获取下次备份时间
func (s *SchedulerService) GetNextBackupTime(policyID uint) (*time.Time, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        entryID, exists := s.jobs[policyID]
        if !exists {
                return nil, fmt.Errorf("策略未调度")
        }

        entry := s.cron.Entry(entryID)
        if !entry.Valid() {
                return nil, fmt.Errorf("无效的调度条目")
        }

        nextTime := entry.Next
        return &nextTime, nil
}

// ScheduleSnapshotPolicy 调度快照策略
func (s *SchedulerService) ScheduleSnapshotPolicy(policy *backup.SnapshotPolicy) error {
        s.mu.Lock()
        defer s.mu.Unlock()

        if !policy.Enabled {
                return nil
        }

        var scheduleExpr string
        switch policy.ScheduleType {
        case "cron":
                scheduleExpr = policy.ScheduleExpr
        case "interval":
                scheduleExpr = fmt.Sprintf("*/%d * * * *", policy.IntervalMinute)
        default:
                return fmt.Errorf("不支持的调度类型: %s", policy.ScheduleType)
        }

        entryID, err := s.cron.AddFunc(scheduleExpr, func() {
                s.executeSnapshot(policy)
        })
        if err != nil {
                return fmt.Errorf("添加定时任务失败: %v", err)
        }

        s.jobs[policy.ID] = entryID
        return nil
}

// executeSnapshot 执行快照
func (s *SchedulerService) executeSnapshot(policy *backup.SnapshotPolicy) {
        ctx := context.Background()

        target := &backup.BackupTarget{
                ID:   policy.TargetID,
                Name: policy.TargetName,
                Type: policy.TargetType,
        }

        result, err := s.snapshotSvc.CreateSnapshot(ctx, policy, target)
        if err != nil {
                if policy.NotifyOnFail {
                        // 发送失败通知
                }
                return
        }

        if policy.NotifyOnSuccess {
                // 发送成功通知
                _ = result
        }
}

// ExportScheduleConfig 导出调度配置
func (s *SchedulerService) ExportScheduleConfig(policyID uint) (string, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        if _, exists := s.jobs[policyID]; !exists {
                return "", fmt.Errorf("策略未调度")
        }

        config := map[string]interface{}{
                "policy_id": policyID,
                "exported_at": time.Now(),
        }

        data, err := json.MarshalIndent(config, "", "  ")
        if err != nil {
                return "", err
        }

        return string(data), nil
}
