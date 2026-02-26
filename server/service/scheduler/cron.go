package scheduler

import (
        "fmt"
        "sync"
        "time"

        "github.com/robfig/cron/v3"
)

// CronScheduler Cron 调度器
type CronScheduler struct {
        cron    *cron.Cron
        jobs    map[uint]cron.EntryID
        mu      sync.RWMutex
        running bool
}

// NewCronScheduler 创建 Cron 调度器
func NewCronScheduler() *CronScheduler {
        return &CronScheduler{
                cron: cron.New(cron.WithSeconds(), cron.WithLocation(time.Local)),
                jobs: make(map[uint]cron.EntryID),
        }
}

// Start 启动调度器
func (s *CronScheduler) Start() {
        s.mu.Lock()
        defer s.mu.Unlock()
        
        if !s.running {
                s.cron.Start()
                s.running = true
        }
}

// Stop 停止调度器
func (s *CronScheduler) Stop() {
        s.mu.Lock()
        defer s.mu.Unlock()
        
        if s.running {
                ctx := s.cron.Stop()
                <-ctx.Done()
                s.running = false
        }
}

// AddFunc 添加定时任务
func (s *CronScheduler) AddFunc(spec string, cmd func()) error {
        s.mu.Lock()
        defer s.mu.Unlock()
        
        entryID, err := s.cron.AddFunc(spec, cmd)
        if err != nil {
                return fmt.Errorf("添加定时任务失败: %w", err)
        }
        
        // 存储任务 ID 映射（这里使用负数表示未关联任务 ID 的临时任务）
        _ = entryID
        
        return nil
}

// AddTask 添加任务
func (s *CronScheduler) AddTask(taskID uint, spec string, cmd func()) error {
        s.mu.Lock()
        defer s.mu.Unlock()
        
        // 如果已存在，先移除
        if entryID, exists := s.jobs[taskID]; exists {
                s.cron.Remove(entryID)
        }
        
        entryID, err := s.cron.AddFunc(spec, cmd)
        if err != nil {
                return fmt.Errorf("添加定时任务失败: %w", err)
        }
        
        s.jobs[taskID] = entryID
        
        return nil
}

// RemoveTask 移除任务
func (s *CronScheduler) RemoveTask(taskID uint) {
        s.mu.Lock()
        defer s.mu.Unlock()
        
        if entryID, exists := s.jobs[taskID]; exists {
                s.cron.Remove(entryID)
                delete(s.jobs, taskID)
        }
}

// GetNextRunTime 获取下次执行时间
func (s *CronScheduler) GetNextRunTime(spec string) *time.Time {
        schedule, err := cron.ParseStandard(spec)
        if err != nil {
                return nil
        }
        
        next := schedule.Next(time.Now())
        return &next
}

// GetEntry 获取任务条目
func (s *CronScheduler) GetEntry(taskID uint) *cron.Entry {
        s.mu.RLock()
        defer s.mu.RUnlock()
        
        if entryID, exists := s.jobs[taskID]; exists {
                entry := s.cron.Entry(entryID)
                return &entry
        }
        
        return nil
}

// GetEntries 获取所有任务条目
func (s *CronScheduler) GetEntries() []cron.Entry {
        return s.cron.Entries()
}

// ValidateCronExpr 验证 Cron 表达式
func ValidateCronExpr(spec string) error {
        _, err := cron.ParseStandard(spec)
        if err != nil {
                return fmt.Errorf("无效的 Cron 表达式: %w", err)
        }
        return nil
}

// ParseCronExpr 解析 Cron 表达式
func ParseCronExpr(spec string) (string, error) {
        schedule, err := cron.ParseStandard(spec)
        if err != nil {
                return "", err
        }
        
        next := schedule.Next(time.Now())
        return next.Format("2006-01-02 15:04:05"), nil
}

// GetCronDescription 获取 Cron 表达式描述
func GetCronDescription(spec string) string {
        // 简单的描述生成
        // 实际可以使用更复杂的解析库
        return fmt.Sprintf("Cron: %s", spec)
}

// CronJobInfo Cron 任务信息
type CronJobInfo struct {
        TaskID       uint      `json:"taskId"`
        TaskName     string    `json:"taskName"`
        CronExpr     string    `json:"cronExpr"`
        NextRunTime  time.Time `json:"nextRunTime"`
        PrevRunTime  time.Time `json:"prevRunTime"`
        IsRunning    bool      `json:"isRunning"`
}

// GetCronJobs 获取所有 Cron 任务
func (s *CronScheduler) GetCronJobs() []*CronJobInfo {
        s.mu.RLock()
        defer s.mu.RUnlock()
        
        var jobs []*CronJobInfo
        
        for taskID, entryID := range s.jobs {
                entry := s.cron.Entry(entryID)
                
                job := &CronJobInfo{
                        TaskID:      taskID,
                        NextRunTime: entry.Next,
                        PrevRunTime: entry.Prev,
                        IsRunning:   true,
                }
                
                // 获取任务名称
                task, err := GetTask(taskID)
                if err == nil {
                        job.TaskName = task.Name
                }
                
                jobs = append(jobs, job)
        }
        
        return jobs
}

// DistributedCronScheduler 分布式 Cron 调度器
type DistributedCronScheduler struct {
        *CronScheduler
        nodeID      string
        leaderNode  string
        leaderMu    sync.RWMutex
        electionTTL time.Duration
}

// NewDistributedCronScheduler 创建分布式 Cron 调度器
func NewDistributedCronScheduler(nodeID string) *DistributedCronScheduler {
        return &DistributedCronScheduler{
                CronScheduler: NewCronScheduler(),
                nodeID:        nodeID,
                electionTTL:   30 * time.Second,
        }
}

// IsLeader 检查是否是 Leader
func (s *DistributedCronScheduler) IsLeader() bool {
        s.leaderMu.RLock()
        defer s.leaderMu.RUnlock()
        return s.leaderNode == s.nodeID
}

// ElectLeader 选举 Leader
func (s *DistributedCronScheduler) ElectLeader() error {
        // TODO: 实现基于 Redis 或数据库的 Leader 选举
        // 这里简化为当前节点成为 Leader
        s.leaderMu.Lock()
        s.leaderNode = s.nodeID
        s.leaderMu.Unlock()
        return nil
}

// AddDistributedTask 添加分布式任务
func (s *DistributedCronScheduler) AddDistributedTask(taskID uint, spec string, cmd func()) error {
        // 只有 Leader 才执行
        wrappedCmd := func() {
                if s.IsLeader() {
                        cmd()
                }
        }
        
        return s.AddTask(taskID, spec, wrappedCmd)
}

// Heartbeat 心跳
func (s *DistributedCronScheduler) Heartbeat() error {
        // TODO: 更新心跳时间到 Redis 或数据库
        return nil
}

// CheckLeader 检查 Leader 状态
func (s *DistributedCronScheduler) CheckLeader() {
        // TODO: 检查 Leader 是否存活，如果不存活则重新选举
}
