package monitor

import (
	"context"
	"fmt"
	"time"

	"yunwei/global"
	"yunwei/model/server"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	URL         string        `json:"url"`
	Timeout     time.Duration `json:"timeout"`
	EnableCache bool          `json:"enableCache"`
	CacheTTL    time.Duration `json:"cacheTtl"`
}

// PrometheusClient Prometheus客户端
type PrometheusClient struct {
	client api.Client
	api    v1.API
	config PrometheusConfig
}

// MetricType 指标类型
type MetricType string

const (
	MetricCPU       MetricType = "cpu"
	MetricMemory    MetricType = "memory"
	MetricDisk      MetricType = "disk"
	MetricNetwork   MetricType = "network"
	MetricLoad      MetricType = "load"
	MetricProcess   MetricType = "process"
	MetricContainer MetricType = "container"
	MetricCustom    MetricType = "custom"
)

// MetricData 指标数据
type MetricData struct {
	Timestamp   time.Time   `json:"timestamp"`
	Name        string      `json:"name"`
	Type        MetricType  `json:"type"`
	Value       float64     `json:"value"`
	Labels      map[string]string `json:"labels"`
	ServerID    uint        `json:"serverId"`
	ServerName  string      `json:"serverName"`
}

// MetricRange 指标范围数据
type MetricRange struct {
	Name   string      `json:"name"`
	Type   MetricType  `json:"type"`
	Values []MetricPoint `json:"values"`
}

// MetricPoint 指标点
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// AlertState 告警状态
type AlertState struct {
	Name        string            `json:"name"`
	State       string            `json:"state"` // firing, pending, inactive
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Value       float64           `json:"value"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      *time.Time        `json:"endsAt"`
}

// NewPrometheusClient 创建Prometheus客户端
func NewPrometheusClient(config PrometheusConfig) (*PrometheusClient, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	client, err := api.NewClient(api.Config{
		Address: config.URL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建Prometheus客户端失败: %w", err)
	}

	return &PrometheusClient{
		client: client,
		api:    v1.NewAPI(client),
		config: config,
	}, nil
}

// Query 执行即时查询
func (p *PrometheusClient) Query(ctx context.Context, query string) ([]MetricData, error) {
	result, _, err := p.api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	return p.parseResult(result), nil
}

// QueryRange 执行范围查询
func (p *PrometheusClient) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]MetricRange, error) {
	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	result, _, err := p.api.QueryRange(ctx, query, r)
	if err != nil {
		return nil, fmt.Errorf("范围查询失败: %w", err)
	}

	return p.parseRangeResult(result), nil
}

// GetCPUUsage 获取CPU使用率
func (p *PrometheusClient) GetCPUUsage(ctx context.Context, serverID uint, duration time.Duration) (*MetricRange, error) {
	query := fmt.Sprintf(`100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[%s])) * 100)`, duration)
	
	start := time.Now().Add(-duration)
	end := time.Now()
	
	ranges, err := p.QueryRange(ctx, query, start, end, 15*time.Second)
	if err != nil {
		return nil, err
	}
	
	if len(ranges) > 0 {
		return &ranges[0], nil
	}
	return nil, fmt.Errorf("无数据")
}

// GetMemoryUsage 获取内存使用率
func (p *PrometheusClient) GetMemoryUsage(ctx context.Context, serverID uint, duration time.Duration) (*MetricRange, error) {
	query := `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
	
	start := time.Now().Add(-duration)
	end := time.Now()
	
	ranges, err := p.QueryRange(ctx, query, start, end, 15*time.Second)
	if err != nil {
		return nil, err
	}
	
	if len(ranges) > 0 {
		return &ranges[0], nil
	}
	return nil, fmt.Errorf("无数据")
}

// GetDiskUsage 获取磁盘使用率
func (p *PrometheusClient) GetDiskUsage(ctx context.Context, serverID uint) ([]MetricData, error) {
	query := `(1 - (node_filesystem_avail_bytes{fstype!="tmpfs"} / node_filesystem_size_bytes{fstype!="tmpfs"})) * 100`
	
	return p.Query(ctx, query)
}

// GetNetworkTraffic 获取网络流量
func (p *PrometheusClient) GetNetworkTraffic(ctx context.Context, serverID uint, duration time.Duration) (inbound, outbound *MetricRange, err error) {
	// 入站流量
	inQuery := `rate(node_network_receive_bytes_total{device!="lo"}[5m]) * 8`
	start := time.Now().Add(-duration)
	end := time.Now()
	
	inRanges, err := p.QueryRange(ctx, inQuery, start, end, 15*time.Second)
	if err != nil {
		return nil, nil, err
	}
	
	// 出站流量
	outQuery := `rate(node_network_transmit_bytes_total{device!="lo"}[5m]) * 8`
	outRanges, err := p.QueryRange(ctx, outQuery, start, end, 15*time.Second)
	if err != nil {
		return nil, nil, err
	}
	
	if len(inRanges) > 0 {
		inbound = &inRanges[0]
	}
	if len(outRanges) > 0 {
		outbound = &outRanges[0]
	}
	
	return inbound, outbound, nil
}

// GetLoadAverage 获取系统负载
func (p *PrometheusClient) GetLoadAverage(ctx context.Context, serverID uint) (load1, load5, load15 float64, err error) {
	// Load 1
	result1, err := p.Query(ctx, `node_load1`)
	if err != nil {
		return 0, 0, 0, err
	}
	
	// Load 5
	result5, err := p.Query(ctx, `node_load5`)
	if err != nil {
		return 0, 0, 0, err
	}
	
	// Load 15
	result15, err := p.Query(ctx, `node_load15`)
	if err != nil {
		return 0, 0, 0, err
	}
	
	if len(result1) > 0 {
		load1 = result1[0].Value
	}
	if len(result5) > 0 {
		load5 = result5[0].Value
	}
	if len(result15) > 0 {
		load15 = result15[0].Value
	}
	
	return load1, load5, load15, nil
}

// GetAlerts 获取告警
func (p *PrometheusClient) GetAlerts(ctx context.Context) ([]AlertState, error) {
	result, err := p.api.Alerts(ctx)
	if err != nil {
		return nil, err
	}
	
	var alerts []AlertState
	for _, alert := range result.Alerts {
		state := AlertState{
			Name:     alert.Name(),
			State:    string(alert.State()),
			Severity: string(alert.Labels["severity"]),
			Labels:   make(map[string]string),
			Annotations: make(map[string]string),
			StartsAt: alert.ActiveAt,
		}
		
		for k, v := range alert.Labels {
			state.Labels[string(k)] = string(v)
		}
		for k, v := range alert.Annotations {
			state.Annotations[string(k)] = string(v)
		}
		
		alerts = append(alerts, state)
	}
	
	return alerts, nil
}

// CollectAndSave 收集并保存指标
func (p *PrometheusClient) CollectAndSave(ctx context.Context, srv *server.Server) error {
	// 收集CPU
	cpu, err := p.Query(ctx, fmt.Sprintf(`100 - (avg by(instance) (irate(node_cpu_seconds_total{instance="%s:9100",mode="idle"}[5m])) * 100)`, srv.Host))
	if err == nil && len(cpu) > 0 {
		srv.CPUUsage = cpu[0].Value
	}
	
	// 收集内存
	mem, err := p.Query(ctx, fmt.Sprintf(`(1 - (node_memory_MemAvailable_bytes{instance="%s:9100"} / node_memory_MemTotal_bytes{instance="%s:9100"})) * 100`, srv.Host, srv.Host))
	if err == nil && len(mem) > 0 {
		srv.MemoryUsage = mem[0].Value
	}
	
	// 收集磁盘
	disk, err := p.Query(ctx, fmt.Sprintf(`(1 - (node_filesystem_avail_bytes{instance="%s:9100",fstype!="tmpfs"} / node_filesystem_size_bytes{instance="%s:9100",fstype!="tmpfs"})) * 100`, srv.Host, srv.Host))
	if err == nil && len(disk) > 0 {
		maxDisk := 0.0
		for _, d := range disk {
			if d.Value > maxDisk {
				maxDisk = d.Value
			}
		}
		srv.DiskUsage = maxDisk
	}
	
	// 收集负载
	load1, load5, load15, _ := p.GetLoadAverage(ctx, srv.ID)
	srv.Load1 = load1
	srv.Load5 = load5
	srv.Load15 = load15
	
	// 更新心跳
	now := time.Now()
	srv.LastHeartbeat = &now
	srv.AgentOnline = true
	
	// 保存指标
	metric := server.ServerMetric{
		ServerID:    srv.ID,
		CPUUsage:    srv.CPUUsage,
		MemoryUsage: srv.MemoryUsage,
		DiskUsage:   srv.DiskUsage,
		Load1:       srv.Load1,
		Load5:       srv.Load5,
		Load15:      srv.Load15,
	}
	
	global.DB.Create(&metric)
	global.DB.Save(srv)
	
	return nil
}

// parseResult 解析查询结果
func (p *PrometheusClient) parseResult(result model.Value) []MetricData {
	var metrics []MetricData
	
	switch v := result.(type) {
	case model.Vector:
		for _, sample := range v {
			metric := MetricData{
				Timestamp: sample.Timestamp.Time(),
				Value:     float64(sample.Value),
				Labels:    make(map[string]string),
			}
			for k, val := range sample.Metric {
				metric.Labels[string(k)] = string(val)
			}
			metrics = append(metrics, metric)
		}
	case *model.Scalar:
		metrics = append(metrics, MetricData{
			Timestamp: v.Timestamp.Time(),
			Value:     float64(v.Value),
		})
	}
	
	return metrics
}

// parseRangeResult 解析范围查询结果
func (p *PrometheusClient) parseRangeResult(result model.Value) []MetricRange {
	var ranges []MetricRange
	
	switch v := result.(type) {
	case model.Matrix:
		for _, stream := range v {
			r := MetricRange{
				Labels: make(map[string]string),
			}
			for k, val := range stream.Metric {
				if k == "__name__" {
					r.Name = string(val)
				} else {
					r.Labels[string(k)] = string(val)
				}
			}
			
			for _, point := range stream.Values {
				r.Values = append(r.Values, MetricPoint{
					Timestamp: point.Timestamp.Time(),
					Value:     float64(point.Value),
				})
			}
			
			ranges = append(ranges, r)
		}
	}
	
	return ranges
}

// GetDashboardData 获取仪表盘数据
func (p *PrometheusClient) GetDashboardData(ctx context.Context, serverID uint) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	
	// 获取最近1小时数据
	duration := time.Hour
	
	// CPU
	cpu, _ := p.GetCPUUsage(ctx, serverID, duration)
	data["cpu"] = cpu
	
	// 内存
	mem, _ := p.GetMemoryUsage(ctx, serverID, duration)
	data["memory"] = mem
	
	// 网络
	inNet, outNet, _ := p.GetNetworkTraffic(ctx, serverID, duration)
	data["networkIn"] = inNet
	data["networkOut"] = outNet
	
	// 负载
	load1, load5, load15, _ := p.GetLoadAverage(ctx, serverID)
	data["load"] = map[string]float64{
		"load1":  load1,
		"load5":  load5,
		"load15": load15,
	}
	
	// 磁盘
	disk, _ := p.GetDiskUsage(ctx, serverID)
	data["disk"] = disk
	
	return data, nil
}
