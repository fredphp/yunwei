package ha

import (
	"context"
	"fmt"
	"sync"
	"time"

	"yunwei/global"
	"yunwei/model/ha"
)

// TaskHAManager 任务 HA 管理器
type TaskHAManager struct {
	mu            sync.RWMutex
	lockService   *DistributedLockService
	clusterMgr    *ClusterManager
	nodeID        string
	taskLocks     map[uint]string // taskID -> lockValue
	runningTasks  map[uint]bool
}

// NewTaskHAManager 创建任务 HA 管理器
func NewTaskHAManager(nodeID string, lockService *DistributedLockService, clusterMgr *ClusterManager) *TaskHAManager {
	return &TaskHAManager{
		nodeID:       nodeID,
		lockService:  lockService,
		clusterMgr:   clusterMgr,
		taskLocks:    make(map[uint]string),
		runningTasks: make(map[uint]bool),
	}
}

// ==================== 任务锁 ====================

// AcquireTaskLock 获取任务锁
func (m *TaskHAManager) AcquireTaskLock(ctx context.Context, taskID uint, ttl time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("task:%d", taskID)
	
	result, err := m.lockService.Acquire(ctx, lockKey, ttl, 
		WithWaitTimeout(5*time.Second),
		WithRetryInterval(100*time.Millisecond),
	)
	if err != nil {
		return false, err
	}

	if result.Acquired {
		m.mu.Lock()
		m.taskLocks[taskID] = result.LockValue
		m.mu.Unlock()
	}

	return result.Acquired, nil
}

// ReleaseTaskLock 释放任务锁
func (m *TaskHAManager) ReleaseTaskLock(ctx context.Context, taskID uint) error {
	m.mu.Lock()
	lockValue, exists := m.taskLocks[taskID]
	if !exists {
		m.mu.Unlock()
		return nil
	}
	delete(m.taskLocks, taskID)
	m.mu.Unlock()

	lockKey := fmt.Sprintf("task:%d", taskID)
	return m.lockService.Release(ctx, lockKey, lockValue)
}

// IsTaskLocked 检查任务是否被锁定
func (m *TaskHAManager) IsTaskLocked(ctx context.Context, taskID uint) (bool, string, error) {
	lockKey := fmt.Sprintf("task:%d", taskID)
	return m.lockService.IsHeld(ctx, lockKey)
}

// ==================== 任务漂移 ====================

// HandleNodeFailure 处理节点故障
func (m *TaskHAManager) HandleNodeFailure(failedNodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取故障节点上运行的任务
	var tasks []TaskInfo
	// TODO: 从任务调度中心获取

	for _, task := range tasks {
		// 尝试获取任务锁
		if acquired, _ := m.AcquireTaskLock(context.Background(), task.ID, 30*time.Second); acquired {
			// 重新调度任务
			go m.rescheduleTask(task)
		}
	}

	return nil
}

// rescheduleTask 重新调度任务
func (m *TaskHAManager) rescheduleTask(task TaskInfo) {
	// 选择新节点
	node, err := m.clusterMgr.SelectNode("round-robin")
	if err != nil {
		return
	}

	// 记录漂移事件
	event := &ha.ClusterEvent{
		EventType: "task_migration",
		NodeID:    node.NodeID,
		Title:     fmt.Sprintf("Task %d migrated to node %s", task.ID, node.NodeName),
		Level:     "info",
		Source:    "task_ha_manager",
	}
	global.DB.Create(event)

	// TODO: 将任务发送到新节点执行
}

// TaskInfo 任务信息
type TaskInfo struct {
	ID       uint
	Name     string
	Type     string
	Command  string
	Priority int
}

// ==================== 任务状态同步 ====================

// SyncTaskState 同步任务状态
func (m *TaskHAManager) SyncTaskState(taskID uint, status string, progress int, output string) error {
	// 记录任务状态到数据库（用于故障恢复）
	state := &TaskState{
		TaskID:    taskID,
		NodeID:    m.nodeID,
		Status:    status,
		Progress:  progress,
		Output:    output,
		UpdatedAt: time.Now(),
	}

	// TODO: 保存到共享存储（如 Redis）

	_ = state
	return nil
}

// GetTaskState 获取任务状态
func (m *TaskHAManager) GetTaskState(taskID uint) (*TaskState, error) {
	// TODO: 从共享存储获取
	return nil, nil
}

// TaskState 任务状态
type TaskState struct {
	TaskID    uint
	NodeID    string
	Status    string
	Progress  int
	Output    string
	UpdatedAt time.Time
}

// ==================== 任务重试 ====================

// RetryTask 重试任务
func (m *TaskHAManager) RetryTask(ctx context.Context, taskID uint, maxRetries int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取任务锁
	acquired, err := m.AcquireTaskLock(ctx, taskID, 30*time.Second)
	if err != nil {
		return err
	}
	if !acquired {
		return fmt.Errorf("task %d is locked by another node", taskID)
	}
	defer m.ReleaseTaskLock(ctx, taskID)

	// TODO: 执行任务重试逻辑

	return nil
}

// ==================== 任务超时监控 ====================

// StartTaskMonitor 启动任务监控
func (m *TaskHAManager) StartTaskMonitor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.checkTaskTimeouts()
			}
		}
	}()
}

// checkTaskTimeouts 检查任务超时
func (m *TaskHAManager) checkTaskTimeouts() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for taskID := range m.runningTasks {
		// 检查任务状态
		state, err := m.GetTaskState(taskID)
		if err != nil || state == nil {
			continue
		}

		// 如果任务超时，标记为失败并释放锁
		if time.Since(state.UpdatedAt) > 10*time.Minute {
			m.ReleaseTaskLock(context.Background(), taskID)
			delete(m.runningTasks, taskID)
		}
	}
}

// ==================== 标记任务状态 ====================

// MarkTaskRunning 标记任务运行中
func (m *TaskHAManager) MarkTaskRunning(taskID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runningTasks[taskID] = true
}

// MarkTaskCompleted 标记任务完成
func (m *TaskHAManager) MarkTaskCompleted(taskID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.runningTasks, taskID)
	m.ReleaseTaskLock(context.Background(), taskID)
}

// GetRunningTasks 获取运行中的任务
func (m *TaskHAManager) GetRunningTasks() []uint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tasks []uint
	for taskID := range m.runningTasks {
		tasks = append(tasks, taskID)
	}
	return tasks
}
