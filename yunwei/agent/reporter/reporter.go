package reporter

import (
	"agent/collector"
	"agent/executor"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Reporter 上报器
type Reporter struct {
	serverAddr string
	agentID    string
	collector  *collector.Collector
	executor   *executor.Executor
	conn       *grpc.ClientConn
}

// NewReporter 创建上报器
func NewReporter(serverAddr, agentID string, coll *collector.Collector, exec *executor.Executor) *Reporter {
	return &Reporter{
		serverAddr: serverAddr,
		agentID:    agentID,
		collector:  coll,
		executor:   exec,
	}
}

// Connect 连接服务器
func (r *Reporter) Connect() error {
	conn, err := grpc.NewClient(r.serverAddr, 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	r.conn = conn
	return nil
}

// Close 关闭连接
func (r *Reporter) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// SendHeartbeat 发送心跳
func (r *Reporter) SendHeartbeat(ctx context.Context) error {
	if r.conn == nil {
		return fmt.Errorf("未连接")
	}

	// 采集当前状态
	result := r.collector.Collect()

	// 构建心跳请求
	req := &HeartbeatRequest{
		AgentId:   r.agentID,
		Timestamp: time.Now().Unix(),
		Status: &AgentStatus{
			CpuUsage:     result.CPUUsage,
			MemoryUsage:  result.MemoryUsage,
			DiskUsage:    result.DiskUsage,
			ProcessCount: int32(result.ProcessCount),
		},
	}

	// 发送心跳（这里简化处理，实际应使用 gRPC）
	log.Printf("心跳: CPU=%.1f%%, 内存=%.1f%%, 磁盘=%.1f%%", 
		result.CPUUsage, result.MemoryUsage, result.DiskUsage)

	// TODO: 实际 gRPC 调用
	// client := pb.NewAgentServiceClient(r.conn)
	// _, err := client.Heartbeat(ctx, req)

	return nil
}

// ReportMetrics 上报指标
func (r *Reporter) ReportMetrics(ctx context.Context, result *collector.CollectResult) error {
	if r.conn == nil {
		return fmt.Errorf("未连接")
	}

	// 构建指标请求
	req := &MetricsRequest{
		AgentId:   r.agentID,
		Timestamp: time.Now().Unix(),
		Metrics:   r.buildMetrics(result),
	}

	// 发送指标（简化处理）
	log.Printf("上报指标: %d 项", len(req.Metrics))

	// TODO: 实际 gRPC 调用
	// client := pb.NewAgentServiceClient(r.conn)
	// _, err := client.ReportMetrics(ctx, req)

	return nil
}

// buildMetrics 构建指标列表
func (r *Reporter) buildMetrics(result *collector.CollectResult) []*Metric {
	var metrics []*Metric

	// CPU 指标
	metrics = append(metrics, &Metric{
		Name: "cpu_usage", Value: result.CPUUsage, Type: "gauge",
	})
	metrics = append(metrics, &Metric{
		Name: "cpu_user", Value: result.CPUUser, Type: "gauge",
	})
	metrics = append(metrics, &Metric{
		Name: "cpu_system", Value: result.CPUSystem, Type: "gauge",
	})

	// 内存指标
	metrics = append(metrics, &Metric{
		Name: "memory_usage", Value: result.MemoryUsage, Type: "gauge",
	})
	metrics = append(metrics, &Metric{
		Name: "memory_used", Value: float64(result.MemoryUsed), Type: "gauge",
	})
	metrics = append(metrics, &Metric{
		Name: "memory_free", Value: float64(result.MemoryFree), Type: "gauge",
	})

	// 磁盘指标
	metrics = append(metrics, &Metric{
		Name: "disk_usage", Value: result.DiskUsage, Type: "gauge",
	})
	metrics = append(metrics, &Metric{
		Name: "disk_used", Value: float64(result.DiskUsed), Type: "gauge",
	})
	metrics = append(metrics, &Metric{
		Name: "disk_free", Value: float64(result.DiskFree), Type: "gauge",
	})

	// 网络指标
	metrics = append(metrics, &Metric{
		Name: "net_in", Value: float64(result.NetIn), Type: "counter",
	})
	metrics = append(metrics, &Metric{
		Name: "net_out", Value: float64(result.NetOut), Type: "counter",
	})

	// 负载指标
	metrics = append(metrics, &Metric{
		Name: "load1", Value: result.Load1, Type: "gauge",
	})
	metrics = append(metrics, &Metric{
		Name: "load5", Value: result.Load5, Type: "gauge",
	})
	metrics = append(metrics, &Metric{
		Name: "load15", Value: result.Load15, Type: "gauge",
	})

	// 进程数
	metrics = append(metrics, &Metric{
		Name: "process_count", Value: float64(result.ProcessCount), Type: "gauge",
	})

	// Docker 容器数
	metrics = append(metrics, &Metric{
		Name: "container_count", Value: float64(len(result.Containers)), Type: "gauge",
	})

	return metrics
}

// ReceiveTasks 接收任务
func (r *Reporter) ReceiveTasks(ctx context.Context) ([]*Task, error) {
	if r.conn == nil {
		return nil, fmt.Errorf("未连接")
	}

	// TODO: 实际 gRPC 调用
	// client := pb.NewAgentServiceClient(r.conn)
	// resp, err := client.ReceiveTasks(ctx, &TaskRequest{AgentId: r.agentID})

	return nil, nil
}

// ExecuteTask 执行任务
func (r *Reporter) ExecuteTask(ctx context.Context, task *Task) error {
	log.Printf("执行任务 [%d]: %s", task.Id, task.Action)

	startTime := time.Now()
	result := r.executor.Execute(ctx, task.Action, int(task.Timeout))

	// 上报任务结果
	taskResult := &TaskResult{
		TaskId:      task.Id,
		AgentId:     r.agentID,
		Success:     result.Success,
		Output:      result.Output,
		Error:       result.Error,
		Duration:    result.Duration,
		CompletedAt: time.Now().Unix(),
	}

	// TODO: 上报结果
	log.Printf("任务完成 [%d]: success=%v, duration=%dms", 
		task.Id, result.Success, time.Since(startTime).Milliseconds())

	// 如果任务执行失败，记录错误
	if !result.Success {
		log.Printf("任务执行失败 [%d]: %s", task.Id, result.Error)
	}

	return nil
}

// SendLog 发送日志
func (r *Reporter) SendLog(ctx context.Context, level, source, message string) error {
	if r.conn == nil {
		return fmt.Errorf("未连接")
	}

	// 构建日志请求
	req := &LogEntry{
		Timestamp: time.Now().Unix(),
		Level:     level,
		Source:    source,
		Message:   message,
	}

	// 发送日志
	log.Printf("[%s] %s: %s", level, source, message)

	// TODO: 实际 gRPC 调用
	_ = req

	return nil
}

// ReportDocker 报告 Docker 状态
func (r *Reporter) ReportDocker(ctx context.Context, containers []collector.ContainerInfo) error {
	if r.conn == nil {
		return fmt.Errorf("未连接")
	}

	// 序列化容器信息
	data, _ := json.Marshal(containers)
	
	log.Printf("报告 Docker 状态: %d 个容器", len(containers))
	_ = data

	return nil
}

// ReportPorts 报告端口占用
func (r *Reporter) ReportPorts(ctx context.Context, ports []collector.PortInfo) error {
	if r.conn == nil {
		return fmt.Errorf("未连接")
	}

	// 序列化端口信息
	data, _ := json.Marshal(ports)
	
	log.Printf("报告端口状态: %d 个端口", len(ports))
	_ = data

	return nil
}

// Proto 结构定义（简化版本）

type HeartbeatRequest struct {
	AgentId   string        `json:"agentId"`
	Timestamp int64         `json:"timestamp"`
	Status    *AgentStatus  `json:"status"`
}

type AgentStatus struct {
	CpuUsage     float64 `json:"cpuUsage"`
	MemoryUsage  float64 `json:"memoryUsage"`
	DiskUsage    float64 `json:"diskUsage"`
	ProcessCount int32   `json:"processCount"`
}

type MetricsRequest struct {
	AgentId   string    `json:"agentId"`
	Timestamp int64     `json:"timestamp"`
	Metrics   []*Metric `json:"metrics"`
}

type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

type Task struct {
	Id        int64  `json:"id"`
	Type      string `json:"type"`
	Action    string `json:"action"`
	Timeout   int32  `json:"timeout"`
	CreatedAt int64  `json:"createdAt"`
}

type TaskResult struct {
	TaskId      int64  `json:"taskId"`
	AgentId     string `json:"agentId"`
	Success     bool   `json:"success"`
	Output      string `json:"output"`
	Error       string `json:"error"`
	Duration    int64  `json:"duration"`
	CompletedAt int64  `json:"completedAt"`
}

type LogEntry struct {
	Timestamp int64  `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}
