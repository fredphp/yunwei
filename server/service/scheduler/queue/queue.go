package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"yunwei/global"
	"yunwei/service/scheduler"
)

// QueueBackend 队列后端接口
type QueueBackend interface {
	Enqueue(queueName string, task *scheduler.Task) error
	Dequeue(queueName string, timeout time.Duration) (*scheduler.Task, error)
	Ack(taskID uint) error
	Nack(taskID uint, reason string) error
	Size(queueName string) (int64, error)
	Clear(queueName string) error
}

// MemoryQueueBackend 内存队列后端
type MemoryQueueBackend struct {
	queues map[string]*PriorityQueue
	mu     sync.RWMutex
}

// PriorityQueue 优先级队列
type PriorityQueue struct {
	items []*QueueItem
	mu    sync.Mutex
	cond  *sync.Cond
}

// QueueItem 队列项
type QueueItem struct {
	Task     *scheduler.Task
	Priority int
	Index    int
}

// Len 实现 heap.Interface
func (pq *PriorityQueue) Len() int { return len(pq.items) }

// Less 实现 heap.Interface
func (pq *PriorityQueue) Less(i, j int) bool {
	// 优先级高的排前面，相同优先级按创建时间排序
	if pq.items[i].Priority != pq.items[j].Priority {
		return pq.items[i].Priority > pq.items[j].Priority
	}
	return pq.items[i].Task.CreatedAt.Before(pq.items[j].Task.CreatedAt)
}

// Swap 实现 heap.Interface
func (pq *PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].Index = i
	pq.items[j].Index = j
}

// Push 实现 heap.Interface
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*QueueItem)
	item.Index = n
	pq.items = append(pq.items, item)
}

// Pop 实现 heap.Interface
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	pq.items = old[0 : n-1]
	return item
}

// NewMemoryQueueBackend 创建内存队列后端
func NewMemoryQueueBackend() *MemoryQueueBackend {
	return &MemoryQueueBackend{
		queues: make(map[string]*PriorityQueue),
	}
}

// getOrCreateQueue 获取或创建队列
func (b *MemoryQueueBackend) getOrCreateQueue(queueName string) *PriorityQueue {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if _, exists := b.queues[queueName]; !exists {
		pq := &PriorityQueue{
			items: make([]*QueueItem, 0),
		}
		pq.cond = sync.NewCond(&pq.mu)
		b.queues[queueName] = pq
	}
	return b.queues[queueName]
}

// Enqueue 入队
func (b *MemoryQueueBackend) Enqueue(queueName string, task *scheduler.Task) error {
	pq := b.getOrCreateQueue(queueName)
	
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	item := &QueueItem{
		Task:     task,
		Priority: int(task.Priority),
	}
	
	pq.items = append(pq.items, item)
	pq.cond.Signal()
	
	// 更新数据库状态
	global.DB.Model(&scheduler.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":   scheduler.TaskStatusQueued,
		"queue_at": time.Now(),
	})
	
	return nil
}

// Dequeue 出队
func (b *MemoryQueueBackend) Dequeue(queueName string, timeout time.Duration) (*scheduler.Task, error) {
	pq := b.getOrCreateQueue(queueName)
	
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	// 等待任务
	if len(pq.items) == 0 {
		done := make(chan struct{})
		go func() {
			pq.cond.Wait()
			close(done)
		}()
		
		select {
		case <-done:
		case <-time.After(timeout):
			return nil, fmt.Errorf("timeout")
		}
	}
	
	if len(pq.items) == 0 {
		return nil, fmt.Errorf("queue empty")
	}
	
	// 取出最高优先级任务
	var highestIdx int
	highestPriority := -1
	for i, item := range pq.items {
		if item.Priority > highestPriority {
			highestPriority = item.Priority
			highestIdx = i
		}
	}
	
	item := pq.items[highestIdx]
	pq.items = append(pq.items[:highestIdx], pq.items[highestIdx+1:]...)
	
	return item.Task, nil
}

// Ack 确认任务
func (b *MemoryQueueBackend) Ack(taskID uint) error {
	return nil
}

// Nack 拒绝任务
func (b *MemoryQueueBackend) Nack(taskID uint, reason string) error {
	return nil
}

// Size 队列大小
func (b *MemoryQueueBackend) Size(queueName string) (int64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	if pq, exists := b.queues[queueName]; exists {
		return int64(len(pq.items)), nil
	}
	return 0, nil
}

// Clear 清空队列
func (b *MemoryQueueBackend) Clear(queueName string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	delete(b.queues, queueName)
	return nil
}

// TaskQueue 任务队列管理器
type TaskQueue struct {
	backend   QueueBackend
	queues    map[string]*QueueConfig
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// QueueConfig 队列配置
type QueueConfig struct {
	Name       string
	MaxWorkers int
	MaxPending int
	Priority   int
	Timeout    int
	MaxRetry   int
}

// NewTaskQueue 创建任务队列
func NewTaskQueue(backend QueueBackend) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskQueue{
		backend: backend,
		queues:  make(map[string]*QueueConfig),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterQueue 注册队列
func (tq *TaskQueue) RegisterQueue(config *QueueConfig) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	tq.queues[config.Name] = config
	
	// 保存到数据库
	var existing scheduler.TaskQueue
	result := global.DB.Where("name = ?", config.Name).First(&existing)
	if result.Error != nil {
		queue := &scheduler.TaskQueue{
			Name:       config.Name,
			MaxWorkers: config.MaxWorkers,
			MaxPending: config.MaxPending,
			Priority:   config.Priority,
			Timeout:    config.Timeout,
			MaxRetry:   config.MaxRetry,
			Enabled:    true,
		}
		global.DB.Create(queue)
	}
	
	return nil
}

// EnqueueTask 任务入队
func (tq *TaskQueue) EnqueueTask(task *scheduler.Task) error {
	// 检查幂等性
	if task.IdempotentKey != "" {
		existing, err := scheduler.GetTaskByIdempotentKey(task.IdempotentKey)
		if err == nil && existing != nil {
			// 检查去重窗口
			if time.Since(existing.CreatedAt).Seconds() < float64(task.DedupWindow) {
				return fmt.Errorf("duplicate task: %s", task.IdempotentKey)
			}
		}
	}
	
	// 检查队列容量
	tq.mu.RLock()
	config, exists := tq.queues[task.QueueName]
	tq.mu.RUnlock()
	
	if !exists {
		task.QueueName = "default"
		config = tq.queues["default"]
	}
	
	size, _ := tq.backend.Size(task.QueueName)
	if config != nil && size >= int64(config.MaxPending) {
		return fmt.Errorf("queue %s is full", task.QueueName)
	}
	
	// 创建任务
	if err := scheduler.CreateTask(task); err != nil {
		return err
	}
	
	// 入队
	if err := tq.backend.Enqueue(task.QueueName, task); err != nil {
		return err
	}
	
	// 记录事件
	scheduler.RecordTaskEvent(task.ID, "enqueued", map[string]interface{}{
		"queue": task.QueueName,
	}, "system", "任务已入队")
	
	return nil
}

// DequeueTask 任务出队
func (tq *TaskQueue) DequeueTask(queueName string, timeout time.Duration) (*scheduler.Task, error) {
	return tq.backend.Dequeue(queueName, timeout)
}

// AckTask 确认任务完成
func (tq *TaskQueue) AckTask(taskID uint) error {
	return tq.backend.Ack(taskID)
}

// NackTask 任务处理失败
func (tq *TaskQueue) NackTask(taskID uint, reason string) error {
	task, err := scheduler.GetTask(taskID)
	if err != nil {
		return err
	}
	
	// 更新重试次数
	task.RetryCount++
	
	// 检查是否需要重试
	if task.RetryCount < task.MaxRetry {
		task.Status = scheduler.TaskStatusRetrying
		
		// 计算重试延迟
		delay := time.Duration(task.RetryDelay) * time.Second
		if task.RetryBackoff == "exponential" {
			delay = delay * time.Duration(1<<uint(task.RetryCount-1))
		}
		
		// 延迟入队
		time.AfterFunc(delay, func() {
			tq.backend.Enqueue(task.QueueName, task)
		})
		
		// 记录事件
		scheduler.RecordTaskEvent(taskID, "retry_scheduled", map[string]interface{}{
			"attempt": task.RetryCount,
			"delay":   delay.String(),
			"reason":  reason,
		}, "system", "任务已安排重试")
	} else {
		task.Status = scheduler.TaskStatusFailed
		task.Error = reason
		
		// 记录事件
		scheduler.RecordTaskEvent(taskID, "failed", map[string]interface{}{
			"attempts": task.RetryCount,
			"reason":   reason,
		}, "system", "任务执行失败")
	}
	
	global.DB.Save(task)
	
	return tq.backend.Nack(taskID, reason)
}

// GetQueueStats 获取队列统计
func (tq *TaskQueue) GetQueueStats(queueName string) (*QueueStats, error) {
	stats := &QueueStats{}
	
	// 队列大小
	size, err := tq.backend.Size(queueName)
	if err != nil {
		return nil, err
	}
	stats.Pending = size
	
	// 运行中任务数
	var running int64
	global.DB.Model(&scheduler.Task{}).Where("queue_name = ? AND status = ?", queueName, scheduler.TaskStatusRunning).Count(&running)
	stats.Running = running
	
	// 今日统计
	today := time.Now().Format("2006-01-02")
	var completed, failed int64
	global.DB.Model(&scheduler.Task{}).Where("queue_name = ? AND status = ? AND DATE(updated_at) = ?", queueName, scheduler.TaskStatusSuccess, today).Count(&completed)
	global.DB.Model(&scheduler.Task{}).Where("queue_name = ? AND status = ? AND DATE(updated_at) = ?", queueName, scheduler.TaskStatusFailed, today).Count(&failed)
	
	stats.CompletedToday = completed
	stats.FailedToday = failed
	
	return stats, nil
}

// QueueStats 队列统计
type QueueStats struct {
	Pending       int64 `json:"pending"`
	Running       int64 `json:"running"`
	CompletedToday int64 `json:"completedToday"`
	FailedToday   int64 `json:"failedToday"`
}

// GetAllQueueStats 获取所有队列统计
func (tq *TaskQueue) GetAllQueueStats() (map[string]*QueueStats, error) {
	stats := make(map[string]*QueueStats)
	
	tq.mu.RLock()
	for name := range tq.queues {
		tq.mu.RUnlock()
		qs, err := tq.GetQueueStats(name)
		if err == nil {
			stats[name] = qs
		}
		tq.mu.RLock()
	}
	tq.mu.RUnlock()
	
	return stats, nil
}

// Stop 停止队列
func (tq *TaskQueue) Stop() {
	tq.cancel()
}

// BatchEnqueue 批量入队
func (tq *TaskQueue) BatchEnqueue(tasks []*scheduler.Task) ([]uint, []error) {
	var taskIDs []uint
	var errors []error
	
	for _, task := range tasks {
		if err := tq.EnqueueTask(task); err != nil {
			errors = append(errors, err)
		} else {
			taskIDs = append(taskIDs, task.ID)
		}
	}
	
	return taskIDs, errors
}

// DelayedEnqueue 延迟入队
func (tq *TaskQueue) DelayedEnqueue(task *scheduler.Task, delay time.Duration) error {
	// 先创建任务
	if err := scheduler.CreateTask(task); err != nil {
		return err
	}
	
	// 设置调度类型
	task.ScheduleType = "delayed"
	task.ScheduleTime = new(time.Time)
	*task.ScheduleTime = time.Now().Add(delay)
	global.DB.Save(task)
	
	// 延迟入队
	time.AfterFunc(delay, func() {
		tq.backend.Enqueue(task.QueueName, task)
	})
	
	return nil
}

// PriorityEnqueue 优先入队
func (tq *TaskQueue) PriorityEnqueue(task *scheduler.Task) error {
	task.Priority = scheduler.PriorityCritical
	return tq.EnqueueTask(task)
}

// DedupEnqueue 去重入队
func (tq *TaskQueue) DedupEnqueue(task *scheduler.Task, windowSeconds int) error {
	task.DedupWindow = windowSeconds
	return tq.EnqueueTask(task)
}

// SerializeTask 序列化任务
func SerializeTask(task *scheduler.Task) ([]byte, error) {
	return json.Marshal(task)
}

// DeserializeTask 反序列化任务
func DeserializeTask(data []byte) (*scheduler.Task, error) {
	var task scheduler.Task
	err := json.Unmarshal(data, &task)
	return &task, err
}
