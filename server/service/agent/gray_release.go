package agent

import (
        "encoding/json"
        "fmt"
        "math/rand"
        "sync"
        "time"

        "yunwei/global"
        "yunwei/model/agent"
)

// GrayReleaseEngine 灰度发布引擎
type GrayReleaseEngine struct {
        upgradeEngine *UpgradeEngine
        mu            sync.RWMutex
        strategies    map[uint]*GrayReleaseContext // 正在执行的策略
}

// GrayReleaseContext 灰度发布上下文
type GrayReleaseContext struct {
        Strategy     *agent.GrayReleaseStrategy
        TargetAgents []agent.Agent
        CurrentBatch int
        StartTime    time.Time
        Canceled     bool
}

// NewGrayReleaseEngine 创建灰度发布引擎
func NewGrayReleaseEngine() *GrayReleaseEngine {
        return &GrayReleaseEngine{
                upgradeEngine: NewUpgradeEngine(),
                strategies:    make(map[uint]*GrayReleaseContext),
        }
}

// ==================== 策略管理 ====================

// CreateStrategy 创建灰度策略
func (e *GrayReleaseEngine) CreateStrategy(req *CreateStrategyRequest) (*agent.GrayReleaseStrategy, error) {
        // 获取版本信息
        var version agent.AgentVersion
        if err := global.DB.First(&version, req.VersionID).Error; err != nil {
                return nil, fmt.Errorf("版本不存在")
        }

        strategy := &agent.GrayReleaseStrategy{
                Name:             req.Name,
                Description:      req.Description,
                VersionID:        req.VersionID,
                VersionName:      version.Version,
                StrategyType:     req.StrategyType,
                WeightPercent:    req.WeightPercent,
                StepSize:         req.StepSize,
                StepInterval:     req.StepInterval,
                GroupList:        req.GroupList,
                LabelSelector:    req.LabelSelector,
                Status:           "pending",
                PauseOnFailure:   req.PauseOnFailure,
                FailureThreshold: req.FailureThreshold,
                AutoRollback:     req.AutoRollback,
                Enabled:          true,
        }

        if strategy.StepSize == 0 {
                strategy.StepSize = 10 // 默认步进 10%
        }
        if strategy.StepInterval == 0 {
                strategy.StepInterval = 30 // 默认 30 分钟
        }
        if strategy.FailureThreshold == 0 {
                strategy.FailureThreshold = 10 // 默认 10% 失败阈值
        }

        if err := global.DB.Create(strategy).Error; err != nil {
                return nil, err
        }

        return strategy, nil
}

// CreateStrategyRequest 创建策略请求
type CreateStrategyRequest struct {
        Name             string  `json:"name" binding:"required"`
        Description      string  `json:"description"`
        VersionID        uint    `json:"versionId" binding:"required"`
        StrategyType     string  `json:"strategyType"`     // weight/group/label
        WeightPercent    int     `json:"weightPercent"`    // 初始灰度百分比
        StepSize         int     `json:"stepSize"`         // 步进大小
        StepInterval     int     `json:"stepInterval"`     // 步进间隔(分钟)
        GroupList        string  `json:"groupList"`        // 分组列表 JSON
        LabelSelector    string  `json:"labelSelector"`    // 标签选择器
        PauseOnFailure   bool    `json:"pauseOnFailure"`   // 失败时暂停
        FailureThreshold float64 `json:"failureThreshold"` // 失败阈值
        AutoRollback     bool    `json:"autoRollback"`     // 自动回滚
}

// UpdateStrategy 更新策略
func (e *GrayReleaseEngine) UpdateStrategy(strategy *agent.GrayReleaseStrategy) error {
        return global.DB.Save(strategy).Error
}

// DeleteStrategy 删除策略
func (e *GrayReleaseEngine) DeleteStrategy(id uint) error {
        // 检查是否正在执行
        e.mu.RLock()
        _, running := e.strategies[id]
        e.mu.RUnlock()

        if running {
                return fmt.Errorf("策略正在执行中，无法删除")
        }

        return global.DB.Delete(&agent.GrayReleaseStrategy{}, id).Error
}

// GetStrategy 获取策略
func (e *GrayReleaseEngine) GetStrategy(id uint) (*agent.GrayReleaseStrategy, error) {
        var strategy agent.GrayReleaseStrategy
        err := global.DB.First(&strategy, id).Error
        return &strategy, err
}

// ListStrategies 列出策略
func (e *GrayReleaseEngine) ListStrategies(status string) ([]agent.GrayReleaseStrategy, error) {
        var strategies []agent.GrayReleaseStrategy
        query := global.DB.Model(&agent.GrayReleaseStrategy{})
        if status != "" {
                query = query.Where("status = ?", status)
        }
        err := query.Order("created_at DESC").Find(&strategies).Error
        return strategies, err
}

// ==================== 策略执行 ====================

// StartStrategy 启动灰度发布
func (e *GrayReleaseEngine) StartStrategy(id uint) error {
        strategy, err := e.GetStrategy(id)
        if err != nil {
                return err
        }

        if strategy.Status != "pending" {
                return fmt.Errorf("策略状态不正确: %s", strategy.Status)
        }

        // 获取目标 Agent 列表
        targetAgents, err := e.getTargetAgents(strategy)
        if err != nil {
                return err
        }

        if len(targetAgents) == 0 {
                return fmt.Errorf("没有符合条件的 Agent")
        }

        // 更新策略状态
        now := time.Now()
        strategy.Status = "running"
        strategy.StartedAt = &now
        strategy.TotalAgents = len(targetAgents)
        global.DB.Save(strategy)

        // 创建执行上下文
        ctx := &GrayReleaseContext{
                Strategy:     strategy,
                TargetAgents: targetAgents,
                StartTime:    now,
        }

        e.mu.Lock()
        e.strategies[id] = ctx
        e.mu.Unlock()

        // 开始执行
        go e.executeStrategy(ctx)

        return nil
}

// getTargetAgents 获取目标 Agent 列表
func (e *GrayReleaseEngine) getTargetAgents(strategy *agent.GrayReleaseStrategy) ([]agent.Agent, error) {
        var agents []agent.Agent

        query := global.DB.Model(&agent.Agent{}).Where("status = ?", agent.AgentStatusOnline)

        // 根据策略类型过滤
        switch strategy.StrategyType {
        case "group":
                // 按分组过滤
                var groupIDs []uint
                if err := json.Unmarshal([]byte(strategy.GroupList), &groupIDs); err == nil {
                        query = query.Where("server_id IN (SELECT id FROM servers WHERE group_id IN ?)", groupIDs)
                }
        case "label":
                // 按标签选择器过滤
                // TODO: 实现标签选择器逻辑
        case "weight":
                // 权重策略，获取所有符合条件的 Agent
        default:
                // 获取所有在线 Agent
        }

        // 获取版本信息
        var version agent.AgentVersion
        if err := global.DB.First(&version, strategy.VersionID).Error; err != nil {
                return nil, err
        }

        // 过滤平台和架构
        query = query.Where("platform = ? AND arch = ?", version.Platform, version.Arch)

        err := query.Find(&agents).Error
        return agents, err
}

// executeStrategy 执行灰度策略
func (e *GrayReleaseEngine) executeStrategy(ctx *GrayReleaseContext) {
        strategy := ctx.Strategy

        switch strategy.StrategyType {
        case "weight":
                e.executeWeightStrategy(ctx)
        case "group":
                e.executeGroupStrategy(ctx)
        case "label":
                e.executeLabelStrategy(ctx)
        default:
                e.executeWeightStrategy(ctx)
        }
}

// executeWeightStrategy 执行权重策略
func (e *GrayReleaseEngine) executeWeightStrategy(ctx *GrayReleaseContext) {
        strategy := ctx.Strategy
        agents := ctx.TargetAgents
        totalAgents := len(agents)

        // 计算初始灰度数量
        initialCount := totalAgents * strategy.WeightPercent / 100
        if initialCount == 0 {
                initialCount = 1
        }

        // 随机选择初始 Agent
        rand.Shuffle(len(agents), func(i, j int) {
                agents[i], agents[j] = agents[j], agents[i]
        })

        // 执行初始灰度
        batch := agents[:initialCount]
        e.upgradeBatch(ctx, batch)

        // 步进升级
        for {
                // 检查是否取消
                if ctx.Canceled {
                        return
                }

                // 等待步进间隔
                time.Sleep(time.Duration(strategy.StepInterval) * time.Minute)

                // 检查失败率
                if !e.checkFailureRate(ctx) {
                        if strategy.PauseOnFailure {
                                e.pauseStrategy(strategy.ID)
                                return
                        }
                        if strategy.AutoRollback {
                                e.autoRollback(ctx)
                                return
                        }
                }

                // 检查是否完成
                if strategy.UpgradedAgents >= totalAgents {
                        e.completeStrategy(strategy.ID)
                        return
                }

                // 计算下一批
                currentPercent := strategy.WeightPercent + strategy.StepSize*(strategy.CurrentStep+1)
                if currentPercent > 100 {
                        currentPercent = 100
                }
                nextCount := totalAgents * currentPercent / 100

                if nextCount <= strategy.UpgradedAgents {
                        // 已经升级完成
                        e.completeStrategy(strategy.ID)
                        return
                }

                // 获取下一批 Agent
                nextBatch := agents[strategy.UpgradedAgents:nextCount]
                if len(nextBatch) == 0 {
                        e.completeStrategy(strategy.ID)
                        return
                }

                // 执行升级
                e.upgradeBatch(ctx, nextBatch)

                // 更新步数
                strategy.CurrentStep++
                global.DB.Save(strategy)
        }
}

// executeGroupStrategy 执行分组策略
func (e *GrayReleaseEngine) executeGroupStrategy(ctx *GrayReleaseContext) {
        // 按分组顺序执行
        agents := ctx.TargetAgents
        e.upgradeBatch(ctx, agents)
        e.completeStrategy(ctx.Strategy.ID)
}

// executeLabelStrategy 执行标签策略
func (e *GrayReleaseEngine) executeLabelStrategy(ctx *GrayReleaseContext) {
        agents := ctx.TargetAgents
        e.upgradeBatch(ctx, agents)
        e.completeStrategy(ctx.Strategy.ID)
}

// upgradeBatch 批量升级
func (e *GrayReleaseEngine) upgradeBatch(ctx *GrayReleaseContext, agents []agent.Agent) {
        strategy := ctx.Strategy

        for _, ag := range agents {
                // 检查是否取消
                if ctx.Canceled {
                        return
                }

                // 创建升级任务
                req := &CreateUpgradeRequest{
                        AgentID:        ag.ID,
                        TargetVersion:  strategy.VersionName,
                        TaskType:       "gray",
                        Priority:       7,
                        RollbackEnabled: true,
                        MaxRetry:       2,
                }

                task, err := e.upgradeEngine.CreateUpgradeTask(req)
                if err != nil {
                        strategy.FailedAgents++
                        continue
                }

                // 执行升级
                if err := e.upgradeEngine.ExecuteUpgrade(task.ID); err != nil {
                        strategy.FailedAgents++
                        continue
                }

                strategy.UpgradedAgents++
        }

        global.DB.Save(strategy)
}

// checkFailureRate 检查失败率
func (e *GrayReleaseEngine) checkFailureRate(ctx *GrayReleaseContext) bool {
        strategy := ctx.Strategy

        if strategy.UpgradedAgents == 0 {
                return true
        }

        failureRate := float64(strategy.FailedAgents) / float64(strategy.UpgradedAgents) * 100
        return failureRate <= strategy.FailureThreshold
}

// autoRollback 自动回滚
func (e *GrayReleaseEngine) autoRollback(ctx *GrayReleaseContext) {
        strategy := ctx.Strategy
        strategy.Status = "rollback"
        global.DB.Save(strategy)

        // TODO: 执行回滚逻辑
}

// ==================== 策略控制 ====================

// PauseStrategy 暂停策略
func (e *GrayReleaseEngine) PauseStrategy(id uint) error {
        return e.pauseStrategy(id)
}

func (e *GrayReleaseEngine) pauseStrategy(id uint) error {
        strategy, err := e.GetStrategy(id)
        if err != nil {
                return err
        }

        if strategy.Status != "running" {
                return fmt.Errorf("策略不在运行状态")
        }

        strategy.Status = "paused"
        global.DB.Save(strategy)

        return nil
}

// ResumeStrategy 恢复策略
func (e *GrayReleaseEngine) ResumeStrategy(id uint) error {
        strategy, err := e.GetStrategy(id)
        if err != nil {
                return err
        }

        if strategy.Status != "paused" {
                return fmt.Errorf("策略不在暂停状态")
        }

        strategy.Status = "running"
        global.DB.Save(strategy)

        // 恢复执行
        e.mu.RLock()
        ctx, exists := e.strategies[id]
        e.mu.RUnlock()

        if exists {
                go e.executeStrategy(ctx)
        }

        return nil
}

// CancelStrategy 取消策略
func (e *GrayReleaseEngine) CancelStrategy(id uint) error {
        e.mu.Lock()
        ctx, exists := e.strategies[id]
        if exists {
                ctx.Canceled = true
                delete(e.strategies, id)
        }
        e.mu.Unlock()

        strategy, err := e.GetStrategy(id)
        if err != nil {
                return err
        }

        strategy.Status = "canceled"
        now := time.Now()
        strategy.CompletedAt = &now
        global.DB.Save(strategy)

        return nil
}

// completeStrategy 完成策略
func (e *GrayReleaseEngine) completeStrategy(id uint) {
        e.mu.Lock()
        delete(e.strategies, id)
        e.mu.Unlock()

        strategy, err := e.GetStrategy(id)
        if err != nil {
                return
        }

        strategy.Status = "completed"
        now := time.Now()
        strategy.CompletedAt = &now
        global.DB.Save(strategy)
}

// ==================== 策略监控 ====================

// GetStrategyProgress 获取策略进度
func (e *GrayReleaseEngine) GetStrategyProgress(id uint) (*StrategyProgress, error) {
        strategy, err := e.GetStrategy(id)
        if err != nil {
                return nil, err
        }

        progress := &StrategyProgress{
                StrategyID:      strategy.ID,
                StrategyName:    strategy.Name,
                Status:          strategy.Status,
                TotalAgents:     strategy.TotalAgents,
                UpgradedAgents:  strategy.UpgradedAgents,
                SuccessAgents:   strategy.SuccessAgents,
                FailedAgents:    strategy.FailedAgents,
                CurrentStep:     strategy.CurrentStep,
                CurrentPercent:  0,
                EstimatedTime:   0,
        }

        // 计算当前百分比
        if strategy.TotalAgents > 0 {
                progress.CurrentPercent = float64(strategy.UpgradedAgents) / float64(strategy.TotalAgents) * 100
        }

        // 获取进行中的任务数
        var runningTasks int64
        global.DB.Model(&agent.AgentUpgradeTask{}).
                Where("task_type = ? AND status IN ?", "gray", []string{"pending", "downloading", "installing"}).
                Count(&runningTasks)
        progress.RunningTasks = int(runningTasks)

        // 预估剩余时间
        if strategy.StrategyType == "weight" && strategy.Status == "running" {
                _ = strategy.TotalAgents - strategy.UpgradedAgents // remainingAgents calculation
                remainingSteps := (100 - int(progress.CurrentPercent)) / strategy.StepSize
                progress.EstimatedTime = remainingSteps * strategy.StepInterval
        }

        return progress, nil
}

// StrategyProgress 策略进度
type StrategyProgress struct {
        StrategyID     uint    `json:"strategyId"`
        StrategyName   string  `json:"strategyName"`
        Status         string  `json:"status"`
        TotalAgents    int     `json:"totalAgents"`
        UpgradedAgents int     `json:"upgradedAgents"`
        SuccessAgents  int     `json:"successAgents"`
        FailedAgents   int     `json:"failedAgents"`
        RunningTasks   int     `json:"runningTasks"`
        CurrentStep    int     `json:"currentStep"`
        CurrentPercent float64 `json:"currentPercent"`
        EstimatedTime  int     `json:"estimatedTime"` // 预估剩余时间(分钟)
}

// GetStrategyStats 获取策略统计
func (e *GrayReleaseEngine) GetStrategyStats() (*StrategyStats, error) {
        stats := &StrategyStats{}

        global.DB.Model(&agent.GrayReleaseStrategy{}).Count(&stats.Total)
        global.DB.Model(&agent.GrayReleaseStrategy{}).Where("status = ?", "running").Count(&stats.Running)
        global.DB.Model(&agent.GrayReleaseStrategy{}).Where("status = ?", "paused").Count(&stats.Paused)
        global.DB.Model(&agent.GrayReleaseStrategy{}).Where("status = ?", "completed").Count(&stats.Completed)
        global.DB.Model(&agent.GrayReleaseStrategy{}).Where("status = ?", "canceled").Count(&stats.Canceled)

        return stats, nil
}

// StrategyStats 策略统计
type StrategyStats struct {
        Total     int64 `json:"total"`
        Running   int64 `json:"running"`
        Paused    int64 `json:"paused"`
        Completed int64 `json:"completed"`
        Canceled  int64 `json:"canceled"`
}
