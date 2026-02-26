package agent

import (
        "fmt"
        "sync"
        "time"

        "yunwei/global"
        "yunwei/model/agent"
        "yunwei/service/notifier"
)

// UpgradeEngine 升级引擎
type UpgradeEngine struct {
        versionManager *VersionManager
        notifier       *notifier.NotifierService
        mu             sync.RWMutex
        pendingTasks   map[uint]*agent.AgentUpgradeTask // 正在进行的升级任务
}

// NewUpgradeEngine 创建升级引擎
func NewUpgradeEngine() *UpgradeEngine {
        return &UpgradeEngine{
                versionManager: NewVersionManager(""),
                notifier:       notifier.NewNotifierService(),
                pendingTasks:   make(map[uint]*agent.AgentUpgradeTask),
        }
}

// SetNotifier 设置通知服务
func (e *UpgradeEngine) SetNotifier(n *notifier.NotifierService) {
        e.notifier = n
}

// ==================== 升级任务创建 ====================

// CreateUpgradeTask 创建升级任务
func (e *UpgradeEngine) CreateUpgradeTask(req *CreateUpgradeRequest) (*agent.AgentUpgradeTask, error) {
        // 获取 Agent 信息
        var ag agent.Agent
        if err := global.DB.First(&ag, req.AgentID).Error; err != nil {
                return nil, fmt.Errorf("Agent 不存在: %w", err)
        }

        // 获取目标版本
        targetVersion, err := e.versionManager.GetVersionByNumber(req.TargetVersion, ag.Platform, ag.Arch)
        if err != nil {
                return nil, fmt.Errorf("目标版本不存在: %w", err)
        }

        // 检查是否需要升级
        needUpgrade, err := e.versionManager.NeedUpgrade(ag.Version, req.TargetVersion)
        if err != nil {
                return nil, err
        }
        if !needUpgrade {
                return nil, fmt.Errorf("当前版本 %s 已经是最新或更高版本", ag.Version)
        }

        // 检查是否已有进行中的任务
        var existingTask agent.AgentUpgradeTask
        err = global.DB.Where("agent_id = ? AND status IN ?", req.AgentID, 
                []string{"pending", "downloading", "installing"}).First(&existingTask).Error
        if err == nil {
                return nil, fmt.Errorf("已有升级任务正在进行中")
        }

        // 创建任务
        task := &agent.AgentUpgradeTask{
                AgentID:        req.AgentID,
                AgentUUID:      ag.AgentID,
                ServerID:       ag.ServerID,
                ServerName:     ag.ServerName,
                FromVersion:    ag.Version,
                ToVersion:      targetVersion.Version,
                ToVersionCode:  targetVersion.VersionCode,
                TaskType:       req.TaskType,
                Priority:       req.Priority,
                ScheduledAt:    req.ScheduledAt,
                Status:         "pending",
                RollbackEnabled: req.RollbackEnabled,
                MaxRetry:       req.MaxRetry,
                CreatedBy:      req.CreatedBy,
                CreatedByName:  req.CreatedByName,
        }

        if task.MaxRetry == 0 {
                task.MaxRetry = 3
        }

        if err := global.DB.Create(task).Error; err != nil {
                return nil, err
        }

        // 更新 Agent 目标版本
        ag.TargetVersion = req.TargetVersion
        global.DB.Save(&ag)

        return task, nil
}

// CreateUpgradeRequest 创建升级请求
type CreateUpgradeRequest struct {
        AgentID        uint       `json:"agentId" binding:"required"`
        TargetVersion  string     `json:"targetVersion" binding:"required"`
        TaskType       string     `json:"taskType"`       // manual/auto/gray
        Priority       int        `json:"priority"`       // 1-10
        ScheduledAt    *time.Time `json:"scheduledAt"`    // 计划执行时间
        RollbackEnabled bool       `json:"rollbackEnabled"`
        MaxRetry       int        `json:"maxRetry"`
        CreatedBy      uint       `json:"createdBy"`
        CreatedByName  string     `json:"createdByName"`
}

// CreateBatchUpgradeTask 批量创建升级任务
func (e *UpgradeEngine) CreateBatchUpgradeTask(req *BatchUpgradeRequest) (*BatchUpgradeResult, error) {
        result := &BatchUpgradeResult{
                CreatedTasks: make([]agent.AgentUpgradeTask, 0),
                FailedAgents: make([]FailedAgent, 0),
        }

        // 获取目标版本
        var version agent.AgentVersion
        if req.VersionID > 0 {
                if err := global.DB.First(&version, req.VersionID).Error; err != nil {
                        return nil, fmt.Errorf("版本不存在")
                }
        } else {
                v, err := e.versionManager.GetLatestVersion(req.Platform, req.Arch, req.ReleaseType)
                if err != nil {
                        return nil, fmt.Errorf("获取最新版本失败: %w", err)
                }
                version = *v
        }

        // 获取需要升级的 Agent 列表
        query := global.DB.Model(&agent.Agent{}).Where("status = ? AND platform = ? AND arch = ?",
                agent.AgentStatusOnline, version.Platform, version.Arch)

        if len(req.AgentIDs) > 0 {
                query = query.Where("id IN ?", req.AgentIDs)
        }
        if len(req.GroupIDs) > 0 {
                // 按分组过滤
        }

        var agents []agent.Agent
        query.Find(&agents)

        // 批量创建任务
        for _, ag := range agents {
                taskReq := &CreateUpgradeRequest{
                        AgentID:        ag.ID,
                        TargetVersion:  version.Version,
                        TaskType:       "batch",
                        Priority:       req.Priority,
                        RollbackEnabled: true,
                        MaxRetry:       3,
                        CreatedBy:      req.CreatedBy,
                        CreatedByName:  req.CreatedByName,
                }

                task, err := e.CreateUpgradeTask(taskReq)
                if err != nil {
                        result.FailedAgents = append(result.FailedAgents, FailedAgent{
                                AgentID: ag.ID,
                                Name:    ag.ServerName,
                                Reason:  err.Error(),
                        })
                        continue
                }

                result.CreatedTasks = append(result.CreatedTasks, *task)
        }

        result.TotalAgents = len(agents)
        result.SuccessCount = len(result.CreatedTasks)
        result.FailedCount = len(result.FailedAgents)

        return result, nil
}

// BatchUpgradeRequest 批量升级请求
type BatchUpgradeRequest struct {
        VersionID    uint   `json:"versionId"`      // 指定版本ID
        Platform     string `json:"platform"`       // 平台
        Arch         string `json:"arch"`           // 架构
        ReleaseType  string `json:"releaseType"`    // 发布类型
        AgentIDs     []uint `json:"agentIds"`       // 指定 Agent
        GroupIDs     []uint `json:"groupIds"`       // 指定分组
        Priority     int    `json:"priority"`       // 优先级
        CreatedBy    uint   `json:"createdBy"`
        CreatedByName string `json:"createdByName"`
}

// BatchUpgradeResult 批量升级结果
type BatchUpgradeResult struct {
        TotalAgents  int                      `json:"totalAgents"`
        SuccessCount int                      `json:"successCount"`
        FailedCount  int                      `json:"failedCount"`
        CreatedTasks []agent.AgentUpgradeTask `json:"createdTasks"`
        FailedAgents []FailedAgent            `json:"failedAgents"`
}

// FailedAgent 失败的 Agent
type FailedAgent struct {
        AgentID uint   `json:"agentId"`
        Name    string `json:"name"`
        Reason  string `json:"reason"`
}

// ==================== 升级执行 ====================

// ExecuteUpgrade 执行升级
func (e *UpgradeEngine) ExecuteUpgrade(taskID uint) error {
        var task agent.AgentUpgradeTask
        if err := global.DB.First(&task, taskID).Error; err != nil {
                return err
        }

        // 检查状态
        if task.Status != "pending" {
                return fmt.Errorf("任务状态不是待执行: %s", task.Status)
        }

        // 获取 Agent 信息
        var ag agent.Agent
        if err := global.DB.First(&ag, task.AgentID).Error; err != nil {
                return fmt.Errorf("Agent 不存在")
        }

        // 检查 Agent 是否在线
        if ag.Status != agent.AgentStatusOnline {
                return fmt.Errorf("Agent 不在线")
        }

        // 更新任务状态
        now := time.Now()
        task.Status = "downloading"
        task.StartedAt = &now
        task.StatusMsg = "正在下发升级指令"
        global.DB.Save(&task)

        // 更新 Agent 状态
        ag.Status = agent.AgentStatusUpgrading
        ag.StatusMessage = fmt.Sprintf("正在升级到 %s", task.ToVersion)
        global.DB.Save(&ag)

        // 添加到待处理列表
        e.mu.Lock()
        e.pendingTasks[taskID] = &task
        e.mu.Unlock()

        // 发送升级指令给 Agent (通过 gRPC 或 WebSocket)
        go e.sendUpgradeCommand(&ag, &task)

        return nil
}

// sendUpgradeCommand 发送升级指令
func (e *UpgradeEngine) sendUpgradeCommand(ag *agent.Agent, task *agent.AgentUpgradeTask) {
        // 获取版本信息
        versionInfo, err := e.versionManager.GetVersionByNumber(task.ToVersion, ag.Platform, ag.Arch)
        if err != nil {
                e.handleUpgradeFailed(task, fmt.Sprintf("获取版本信息失败: %v", err))
                return
        }

        // 构建升级指令
        upgradeCmd := &UpgradeCommand{
                TaskID:       task.ID,
                Version:      task.ToVersion,
                VersionCode:  task.ToVersionCode,
                DownloadURL:  versionInfo.FileURL,
                FileMD5:      versionInfo.FileMD5,
                FileSize:     versionInfo.FileSize,
                ForceUpdate:  versionInfo.ForceUpdate,
                RollbackEnabled: task.RollbackEnabled,
                Timeout:      600, // 10分钟超时
        }

        // TODO: 通过 gRPC 或 WebSocket 发送给 Agent
        // 这里是模拟发送成功
        _ = upgradeCmd

        // 记录事件
        e.recordTaskEvent(task.ID, "command_sent", map[string]interface{}{
                "version": task.ToVersion,
        }, "system", "升级指令已发送")

        // 启动超时监控
        go e.monitorUpgradeTimeout(task.ID)
}

// UpgradeCommand 升级指令
type UpgradeCommand struct {
        TaskID          uint   `json:"taskId"`
        Version         string `json:"version"`
        VersionCode     int    `json:"versionCode"`
        DownloadURL     string `json:"downloadUrl"`
        FileMD5         string `json:"fileMd5"`
        FileSize        int64  `json:"fileSize"`
        ForceUpdate     bool   `json:"forceUpdate"`
        RollbackEnabled bool   `json:"rollbackEnabled"`
        Timeout         int    `json:"timeout"` // 秒
}

// monitorUpgradeTimeout 监控升级超时
func (e *UpgradeEngine) monitorUpgradeTimeout(taskID uint) {
        time.Sleep(10 * time.Minute) // 10分钟超时

        e.mu.RLock()
        task, exists := e.pendingTasks[taskID]
        e.mu.RUnlock()

        if !exists {
                return
        }

        // 检查是否还在执行
        if task.Status == "downloading" || task.Status == "installing" {
                e.handleUpgradeFailed(task, "升级超时")
        }
}

// ==================== 升级结果处理 ====================

// HandleUpgradeProgress 处理升级进度
func (e *UpgradeEngine) HandleUpgradeProgress(taskID uint, progress int, status string, message string) error {
        var task agent.AgentUpgradeTask
        if err := global.DB.First(&task, taskID).Error; err != nil {
                return err
        }

        task.Progress = progress
        task.Status = status
        task.StatusMsg = message
        return global.DB.Save(&task).Error
}

// HandleUpgradeSuccess 处理升级成功
func (e *UpgradeEngine) HandleUpgradeSuccess(taskID uint, output string) error {
        var task agent.AgentUpgradeTask
        if err := global.DB.First(&task, taskID).Error; err != nil {
                return err
        }

        now := time.Now()
        task.Status = "success"
        task.Progress = 100
        task.CompletedAt = &now
        if task.StartedAt != nil {
                task.Duration = now.Sub(*task.StartedAt).Milliseconds()
        }
        task.Output = output
        global.DB.Save(&task)

        // 更新 Agent 版本
        var ag agent.Agent
        if err := global.DB.First(&ag, task.AgentID).Error; err == nil {
                ag.Version = task.ToVersion
                ag.VersionCode = task.ToVersionCode
                ag.TargetVersion = ""
                ag.Status = agent.AgentStatusOnline
                ag.StatusMessage = ""
                global.DB.Save(&ag)
        }

        // 增加安装计数
        e.versionManager.IncrementInstallCount(uint(task.ToVersionCode))

        // 从待处理列表移除
        e.mu.Lock()
        delete(e.pendingTasks, taskID)
        e.mu.Unlock()

        // 记录事件
        e.recordTaskEvent(taskID, "success", nil, "agent", "升级成功")

        // 发送通知
        if e.notifier != nil {
                e.notifier.Broadcast(
                        fmt.Sprintf("✅ Agent 升级成功 - %s", task.ServerName),
                        fmt.Sprintf("版本: %s -> %s\n耗时: %dms", task.FromVersion, task.ToVersion, task.Duration),
                )
        }

        return nil
}

// handleUpgradeFailed 处理升级失败
func (e *UpgradeEngine) handleUpgradeFailed(task *agent.AgentUpgradeTask, errMsg string) {
        now := time.Now()
        task.Status = "failed"
        task.Error = errMsg
        task.CompletedAt = &now
        if task.StartedAt != nil {
                task.Duration = now.Sub(*task.StartedAt).Milliseconds()
        }
        global.DB.Save(task)

        // 更新 Agent 状态
        var ag agent.Agent
        if err := global.DB.First(&ag, task.AgentID).Error; err == nil {
                ag.Status = agent.AgentStatusError
                ag.StatusMessage = fmt.Sprintf("升级失败: %s", errMsg)
                ag.ErrorCount++
                ag.LastErrorAt = &now
                ag.LastErrorMsg = errMsg
                global.DB.Save(&ag)
        }

        // 从待处理列表移除
        e.mu.Lock()
        delete(e.pendingTasks, task.ID)
        e.mu.Unlock()

        // 记录事件
        e.recordTaskEvent(task.ID, "failed", map[string]interface{}{
                "error": errMsg,
        }, "system", "升级失败")

        // 发送通知
        if e.notifier != nil {
                e.notifier.Broadcast(
                        fmt.Sprintf("❌ Agent 升级失败 - %s", task.ServerName),
                        fmt.Sprintf("版本: %s -> %s\n错误: %s", task.FromVersion, task.ToVersion, errMsg),
                )
        }

        // 检查是否需要重试
        if task.RetryCount < task.MaxRetry {
                go e.retryUpgrade(task)
        }
}

// retryUpgrade 重试升级
func (e *UpgradeEngine) retryUpgrade(task *agent.AgentUpgradeTask) {
        task.RetryCount++
        task.Status = "pending"
        task.Progress = 0
        task.Error = ""
        task.StartedAt = nil
        task.CompletedAt = nil
        global.DB.Save(task)

        // 延迟重试
        time.Sleep(time.Duration(task.RetryCount*30) * time.Second)

        e.ExecuteUpgrade(task.ID)
}

// recordTaskEvent 记录任务事件
func (e *UpgradeEngine) recordTaskEvent(taskID uint, eventType string, data map[string]interface{}, operator, message string) {
        // TODO: 实现事件记录
}

// ==================== 升级回滚 ====================

// RollbackUpgrade 回滚升级
func (e *UpgradeEngine) RollbackUpgrade(taskID uint) error {
        var task agent.AgentUpgradeTask
        if err := global.DB.First(&task, taskID).Error; err != nil {
                return err
        }

        if !task.RollbackEnabled {
                return fmt.Errorf("该任务未启用回滚")
        }

        if task.Status != "success" {
                return fmt.Errorf("只有成功的任务才能回滚")
        }

        // 创建回滚任务
        rollbackTask := &agent.AgentUpgradeTask{
                AgentID:        task.AgentID,
                AgentUUID:      task.AgentUUID,
                ServerID:       task.ServerID,
                ServerName:     task.ServerName,
                FromVersion:    task.ToVersion,
                ToVersion:      task.FromVersion,
                TaskType:       "rollback",
                Priority:       10, // 高优先级
                Status:         "pending",
                RollbackEnabled: false,
                MaxRetry:       1,
        }

        if err := global.DB.Create(rollbackTask).Error; err != nil {
                return err
        }

        // 更新原任务
        now := time.Now()
        task.RollbackAt = &now
        task.Status = "rolledback"
        global.DB.Save(&task)

        // 执行回滚
        return e.ExecuteUpgrade(rollbackTask.ID)
}

// ==================== 任务查询 ====================

// GetUpgradeTask 获取升级任务
func (e *UpgradeEngine) GetUpgradeTask(id uint) (*agent.AgentUpgradeTask, error) {
        var task agent.AgentUpgradeTask
        err := global.DB.First(&task, id).Error
        return &task, err
}

// ListUpgradeTasks 列出升级任务
func (e *UpgradeEngine) ListUpgradeTasks(filter *UpgradeTaskFilter) ([]agent.AgentUpgradeTask, int64, error) {
        query := global.DB.Model(&agent.AgentUpgradeTask{})

        if filter != nil {
                if filter.AgentID > 0 {
                        query = query.Where("agent_id = ?", filter.AgentID)
                }
                if filter.Status != "" {
                        query = query.Where("status = ?", filter.Status)
                }
                if filter.TaskType != "" {
                        query = query.Where("task_type = ?", filter.TaskType)
                }
        }

        var total int64
        query.Count(&total)

        var tasks []agent.AgentUpgradeTask
        err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&tasks).Error
        return tasks, total, err
}

// UpgradeTaskFilter 升级任务过滤器
type UpgradeTaskFilter struct {
        AgentID  uint   `json:"agentId"`
        Status   string `json:"status"`
        TaskType string `json:"taskType"`
        Limit    int    `json:"limit"`
        Offset   int    `json:"offset"`
}

// CancelUpgradeTask 取消升级任务
func (e *UpgradeEngine) CancelUpgradeTask(taskID uint) error {
        var task agent.AgentUpgradeTask
        if err := global.DB.First(&task, taskID).Error; err != nil {
                return err
        }

        if task.Status == "success" || task.Status == "failed" {
                return fmt.Errorf("任务已完成，无法取消")
        }

        if task.Status == "installing" {
                return fmt.Errorf("任务正在安装中，无法取消")
        }

        task.Status = "canceled"
        task.StatusMsg = "用户取消"
        now := time.Now()
        task.CompletedAt = &now
        global.DB.Save(&task)

        // 更新 Agent 状态
        var ag agent.Agent
        if err := global.DB.First(&ag, task.AgentID).Error; err == nil {
                ag.Status = agent.AgentStatusOnline
                ag.StatusMessage = ""
                ag.TargetVersion = ""
                global.DB.Save(&ag)
        }

        return nil
}

// ==================== 统计 ====================

// GetUpgradeStats 获取升级统计
func (e *UpgradeEngine) GetUpgradeStats() (*UpgradeStats, error) {
        stats := &UpgradeStats{}

        // 今日升级数
        today := time.Now().Truncate(24 * time.Hour)
        global.DB.Model(&agent.AgentUpgradeTask{}).
                Where("created_at >= ?", today).
                Count(&stats.TodayTotal)

        global.DB.Model(&agent.AgentUpgradeTask{}).
                Where("created_at >= ? AND status = ?", today, "success").
                Count(&stats.TodaySuccess)

        global.DB.Model(&agent.AgentUpgradeTask{}).
                Where("created_at >= ? AND status = ?", today, "failed").
                Count(&stats.TodayFailed)

        // 进行中的任务
        global.DB.Model(&agent.AgentUpgradeTask{}).
                Where("status IN ?", []string{"pending", "downloading", "installing"}).
                Count(&stats.Running)

        // 总计
        global.DB.Model(&agent.AgentUpgradeTask{}).Count(&stats.Total)
        global.DB.Model(&agent.AgentUpgradeTask{}).
                Where("status = ?", "success").Count(&stats.TotalSuccess)
        global.DB.Model(&agent.AgentUpgradeTask{}).
                Where("status = ?", "failed").Count(&stats.TotalFailed)

        // 成功率
        if stats.Total > 0 {
                stats.SuccessRate = float64(stats.TotalSuccess) / float64(stats.Total) * 100
        }

        return stats, nil
}

// UpgradeStats 升级统计
type UpgradeStats struct {
        TodayTotal   int64   `json:"todayTotal"`
        TodaySuccess int64   `json:"todaySuccess"`
        TodayFailed  int64   `json:"todayFailed"`
        Running      int64   `json:"running"`
        Total        int64   `json:"total"`
        TotalSuccess int64   `json:"totalSuccess"`
        TotalFailed  int64   `json:"totalFailed"`
        SuccessRate  float64 `json:"successRate"`
}
