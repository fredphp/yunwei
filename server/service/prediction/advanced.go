package prediction

import (
        "encoding/json"
        "fmt"
        "math"
        "time"

        "yunwei/global"
        "yunwei/model/server"
        "yunwei/service/ai/llm"
)

// AnomalyType 异常类型
type AnomalyType string

const (
        AnomalyTypeCPU      AnomalyType = "cpu_anomaly"
        AnomalyTypeMemory   AnomalyType = "memory_anomaly"
        AnomalyTypeDisk     AnomalyType = "disk_anomaly"
        AnomalyTypeNetwork  AnomalyType = "network_anomaly"
        AnomalyTypeProcess  AnomalyType = "process_anomaly"
        AnomalyTypeCombined AnomalyType = "combined_anomaly"
)

// AnomalyDetection 异常检测结果
type AnomalyDetection struct {
        ID           uint           `json:"id" gorm:"primarykey"`
        CreatedAt    time.Time      `json:"createdAt"`
        ServerID     uint           `json:"serverId" gorm:"index"`
        Server       *server.Server `json:"server" gorm:"foreignKey:ServerID"`

        Type         AnomalyType    `json:"type" gorm:"type:varchar(32)"`
        Severity     string         `json:"severity" gorm:"type:varchar(16)"` // low, medium, high, critical
        
        Score        float64        `json:"score"` // 异常分数 0-100
        Confidence   float64        `json:"confidence"`
        
        Description  string         `json:"description" gorm:"type:text"`
        Indicators   string         `json:"indicators" gorm:"type:text"` // JSON格式的指标列表
        
        PredictedAt  time.Time      `json:"predictedAt"`
        OccurredAt   *time.Time     `json:"occurredAt"`
        
        Status       string         `json:"status" gorm:"type:varchar(16);default:'predicted'"` // predicted, occurred, false_positive
        
        AIAnalysis   string         `json:"aiAnalysis" gorm:"type:text"`
}

func (AnomalyDetection) TableName() string {
        return "anomaly_detections"
}

// AutoScaleRecommendation 自动扩容建议
type AutoScaleRecommendation struct {
        ID          uint           `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time      `json:"createdAt"`
        ServerID    uint           `json:"serverId" gorm:"index"`
        Server      *server.Server `json:"server" gorm:"foreignKey:ServerID"`

        // 当前状态
        CurrentCPU    float64 `json:"currentCpu"`
        CurrentMemory float64 `json:"currentMemory"`
        CurrentLoad   float64 `json:"currentLoad"`

        // 建议配置
        RecommendedCPU    int     `json:"recommendedCpu"`    // 推荐CPU核心数
        RecommendedMemory int     `json:"recommendedMemory"` // 推荐内存(GB)
        
        // Docker/K8s 扩容
        ScaleType       string `json:"scaleType"`        // docker, k8s, vm
        ContainerName   string `json:"containerName"`
        Replicas        int    `json:"replicas"`         // 当前副本数
        RecommendedReplicas int `json:"recommendedReplicas"` // 推荐副本数
        
        // 理由
        Reason         string   `json:"reason" gorm:"type:text"`
        AIRecommendation string `json:"aiRecommendation" gorm:"type:text"`
        
        // 状态
        Status         string `json:"status" gorm:"type:varchar(16);default:'pending'"` // pending, applied, dismissed
        AppliedAt      *time.Time `json:"appliedAt"`
}

func (AutoScaleRecommendation) TableName() string {
        return "auto_scale_recommendations"
}

// AdvancedPredictor 高级预测器
type AdvancedPredictor struct {
        llmClient *llm.GLM5Client
}

// NewAdvancedPredictor 创建高级预测器
func NewAdvancedPredictor(llmClient *llm.GLM5Client) *AdvancedPredictor {
        return &AdvancedPredictor{
                llmClient: llmClient,
        }
}

// DetectAnomaly 检测异常
func (p *AdvancedPredictor) DetectAnomaly(serverID uint, history []server.ServerMetric) (*AnomalyDetection, error) {
        if len(history) < 20 {
                return nil, fmt.Errorf("历史数据不足")
        }

        // 计算统计指标
        cpuStats := calculateStats(history, func(m server.ServerMetric) float64 { return m.CPUUsage })
        memStats := calculateStats(history, func(m server.ServerMetric) float64 { return m.MemoryUsage })
        loadStats := calculateStats(history, func(m server.ServerMetric) float64 { return m.Load1 })

        // 异常检测
        var anomalies []string
        var maxScore float64

        // CPU 异常检测
        if cpuStats.Last > cpuStats.Mean+2*cpuStats.StdDev {
                anomalies = append(anomalies, "CPU使用率异常偏高")
                maxScore = math.Max(maxScore, (cpuStats.Last-cpuStats.Mean)/cpuStats.StdDev*20)
        }

        // 内存异常检测
        if memStats.Last > memStats.Mean+2*memStats.StdDev {
                anomalies = append(anomalies, "内存使用率异常偏高")
                maxScore = math.Max(maxScore, (memStats.Last-memStats.Mean)/memStats.StdDev*20)
        }

        // 负载异常检测
        if loadStats.Last > loadStats.Mean+2*loadStats.StdDev {
                anomalies = append(anomalies, "系统负载异常偏高")
                maxScore = math.Max(maxScore, (loadStats.Last-loadStats.Mean)/loadStats.StdDev*20)
        }

        // 突变检测
        if cpuStats.ChangeRate > 50 {
                anomalies = append(anomalies, fmt.Sprintf("CPU使用率突变 %.1f%%", cpuStats.ChangeRate))
                maxScore = math.Max(maxScore, cpuStats.ChangeRate)
        }

        if memStats.ChangeRate > 30 {
                anomalies = append(anomalies, fmt.Sprintf("内存使用率突变 %.1f%%", memStats.ChangeRate))
                maxScore = math.Max(maxScore, memStats.ChangeRate)
        }

        if len(anomalies) == 0 {
                return nil, nil // 无异常
        }

        // 确定严重程度
        severity := "low"
        if maxScore > 70 {
                severity = "critical"
        } else if maxScore > 50 {
                severity = "high"
        } else if maxScore > 30 {
                severity = "medium"
        }

        // AI 分析
        aiAnalysis := ""
        if p.llmClient != nil {
                aiAnalysis = p.aiAnalyzeAnomaly(serverID, history, anomalies)
        }

        indicatorsJSON, _ := json.Marshal(map[string]interface{}{
                "cpu":  cpuStats,
                "mem":  memStats,
                "load": loadStats,
        })

        detection := &AnomalyDetection{
                ServerID:     serverID,
                Type:         AnomalyTypeCombined,
                Severity:     severity,
                Score:        math.Min(maxScore, 100),
                Confidence:   calculateConfidence(len(history)),
                Description:  fmt.Sprintf("检测到%d个异常指标", len(anomalies)),
                Indicators:   string(indicatorsJSON),
                PredictedAt:  time.Now(),
                Status:       "predicted",
                AIAnalysis:   aiAnalysis,
        }

        // 保存
        global.DB.Create(detection)

        return detection, nil
}

// PredictDiskFull 预测磁盘满
func (p *AdvancedPredictor) PredictDiskFull(serverID uint, history []server.ServerMetric) (*PredictionResult, error) {
        if len(history) < 10 {
                return nil, fmt.Errorf("历史数据不足")
        }

        // 提取磁盘数据
        var data []HistoryData
        for _, m := range history {
                data = append(data, HistoryData{
                        Timestamp: m.CreatedAt,
                        Value:     m.DiskUsage,
                })
        }

        // 计算增长率
        stats := calculateStats(history, func(m server.ServerMetric) float64 { return m.DiskUsage })
        
        // 预测磁盘满的时间
        daysToFull := -1.0
        if stats.ChangeRate > 0 {
                remaining := 100 - stats.Last
                daysToFull = remaining / stats.ChangeRate / 24 // 转换为天
        }

        level := PredictionLevelNormal
        if daysToFull > 0 && daysToFull < 7 {
                level = PredictionLevelCritical
        } else if daysToFull > 0 && daysToFull < 14 {
                level = PredictionLevelWarning
        }

        var suggestions []string
        if level == PredictionLevelCritical {
                suggestions = []string{
                        fmt.Sprintf("磁盘将在 %.1f 天后满", daysToFull),
                        "立即执行磁盘清理",
                        "docker system prune -af",
                        "journalctl --vacuum-time=1d",
                        "考虑扩容磁盘",
                }
        } else if level == PredictionLevelWarning {
                suggestions = []string{
                        fmt.Sprintf("磁盘将在 %.1f 天后满", daysToFull),
                        "规划清理任务",
                        "检查大文件和日志",
                }
        } else {
                suggestions = []string{
                        "磁盘空间充足",
                }
        }

        confidence := 0.6
        if stats.StdDev < 5 {
                confidence = 0.9
        }

        trendDir := "stable"
        if stats.ChangeRate > 0 {
                trendDir = "up"
        }

        result := &PredictionResult{
                ServerID:       serverID,
                Type:           PredictionDisk,
                Level:          level,
                CurrentValue:   stats.Last,
                PredictedValue: stats.Last + stats.ChangeRate*24,
                PredictedAt:    time.Now().Add(24 * time.Hour),
                Confidence:     confidence,
                Trend:          trendDir,
                TrendRate:      stats.ChangeRate,
                Summary:        fmt.Sprintf("磁盘使用率 %.1f%%，预计 %.1f 天后满", stats.Last, daysToFull),
                Suggestions:    suggestions,
        }

        return result, nil
}

// PredictTrafficPeak 预测流量峰值
func (p *AdvancedPredictor) PredictTrafficPeak(serverID uint, history []server.ServerMetric) (*PredictionResult, error) {
        if len(history) < 24 {
                return nil, fmt.Errorf("历史数据不足")
        }

        // 分析流量模式
        var hourlyTraffic [24]float64
        countPerHour := make(map[int]int)

        for _, m := range history {
                hour := m.CreatedAt.Hour()
                hourlyTraffic[hour] += float64(m.NetIn + m.NetOut)
                countPerHour[hour]++
        }

        // 计算每小时平均流量
        for i := 0; i < 24; i++ {
                if countPerHour[i] > 0 {
                        hourlyTraffic[i] /= float64(countPerHour[i])
                }
        }

        // 找到峰值时段
        maxHour := 0
        maxTraffic := hourlyTraffic[0]
        for i := 1; i < 24; i++ {
                if hourlyTraffic[i] > maxTraffic {
                        maxHour = i
                        maxTraffic = hourlyTraffic[i]
                }
        }

        // 当前小时
        currentHour := time.Now().Hour()
        currentTraffic := hourlyTraffic[currentHour]

        // 预测
        level := PredictionLevelNormal
        summary := fmt.Sprintf("流量峰值时段在 %d:00，当前流量正常", maxHour)
        
        if currentHour == maxHour || (currentHour+1)%24 == maxHour {
                level = PredictionLevelWarning
                summary = fmt.Sprintf("即将进入流量峰值时段 (%d:00)，请做好准备", maxHour)
        }

        suggestions := []string{
                fmt.Sprintf("历史峰值时段: %d:00 - %d:00", maxHour, (maxHour+1)%24),
                "建议在峰值前检查服务状态",
                "准备好扩容方案",
        }

        if level == PredictionLevelWarning {
                suggestions = append(suggestions,
                        "当前接近峰值时段",
                        "监控服务响应时间",
                        "准备备用资源",
                )
        }

        result := &PredictionResult{
                ServerID:       serverID,
                Type:           PredictionNetwork,
                Level:          level,
                CurrentValue:   currentTraffic,
                PredictedValue: maxTraffic,
                PredictedAt:    time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), maxHour, 0, 0, 0, time.Local),
                Confidence:     0.7,
                Trend:          currentTraffic < maxTraffic ? "up" : "stable",
                Summary:        summary,
                Suggestions:    suggestions,
        }

        return result, nil
}

// RecommendAutoScale 自动扩容建议
func (p *AdvancedPredictor) RecommendAutoScale(srv *server.Server, history []server.ServerMetric) (*AutoScaleRecommendation, error) {
        if len(history) < 10 {
                return nil, fmt.Errorf("历史数据不足")
        }

        // 计算当前状态
        cpuStats := calculateStats(history, func(m server.ServerMetric) float64 { return m.CPUUsage })
        memStats := calculateStats(history, func(m server.ServerMetric) float64 { return m.MemoryUsage })

        // 判断是否需要扩容
        needScale := cpuStats.Last > 70 || memStats.Last > 80 || cpuStats.Trend == "up"

        if !needScale {
                return nil, nil
        }

        // 计算推荐配置
        recommendation := &AutoScaleRecommendation{
                ServerID:        srv.ID,
                CurrentCPU:      cpuStats.Last,
                CurrentMemory:   memStats.Last,
                CurrentLoad:     history[len(history)-1].Load1,
                RecommendedCPU:  srv.CPUCores,
                RecommendedMemory: int(srv.MemoryTotal / 1024),
                Status:          "pending",
        }

        // CPU 扩容建议
        if cpuStats.Last > 80 {
                recommendation.RecommendedCPU = srv.CPUCores + 2
                recommendation.Reason = fmt.Sprintf("CPU使用率持续 %.1f%%，建议增加2个核心", cpuStats.Last)
        } else if cpuStats.Last > 70 {
                recommendation.RecommendedCPU = srv.CPUCores + 1
                recommendation.Reason = fmt.Sprintf("CPU使用率 %.1f%%，建议增加1个核心", cpuStats.Last)
        }

        // 内存扩容建议
        if memStats.Last > 85 {
                recommendation.RecommendedMemory = int(float64(srv.MemoryTotal) * 1.5 / 1024)
                recommendation.Reason += fmt.Sprintf("; 内存使用率 %.1f%%，建议扩容50%%", memStats.Last)
        } else if memStats.Last > 75 {
                recommendation.RecommendedMemory = int(float64(srv.MemoryTotal) * 1.25 / 1024)
                recommendation.Reason += fmt.Sprintf("; 内存使用率 %.1f%%，建议扩容25%%", memStats.Last)
        }

        // AI 分析
        if p.llmClient != nil {
                recommendation.AIRecommendation = p.aiAnalyzeScale(srv, history)
        }

        // 保存
        global.DB.Create(recommendation)

        return recommendation, nil
}

// aiAnalyzeAnomaly AI分析异常
func (p *AdvancedPredictor) aiAnalyzeAnomaly(serverID uint, history []server.ServerMetric, anomalies []string) string {
        if p.llmClient == nil {
                return ""
        }

        var historyStr string
        for i := len(history) - 1; i >= 0 && i >= len(history)-10; i-- {
                m := history[i]
                historyStr += fmt.Sprintf("[%s] CPU:%.1f%% Mem:%.1f%% Load:%.2f\n",
                        m.CreatedAt.Format("15:04"), m.CPUUsage, m.MemoryUsage, m.Load1)
        }

        prompt := fmt.Sprintf(`分析服务器异常：

检测到的异常:
%s

最近历史数据:
%s

请分析:
1. 异常可能的原因
2. 可能的影响
3. 推荐的处理步骤
4. 预防措施

简要回复(200字以内)。`, anomalies, historyStr)

        response, err := p.llmClient.QuickChat(prompt)
        if err != nil {
                return ""
        }

        return response
}

// aiAnalyzeScale AI分析扩容需求
func (p *AdvancedPredictor) aiAnalyzeScale(srv *server.Server, history []server.ServerMetric) string {
        if p.llmClient == nil {
                return ""
        }

        prompt := fmt.Sprintf(`服务器资源分析:

服务器: %s
CPU核心: %d
内存: %d MB
磁盘: %d GB

当前状态:
- CPU使用率: %.1f%%
- 内存使用率: %.1f%%
- 系统负载: %.2f

请给出扩容建议，包括:
1. 是否需要扩容
2. 推荐的资源配置
3. 扩容时机建议
4. 成本优化建议

简要回复(150字以内)。`,
                srv.Name, srv.CPUCores, srv.MemoryTotal, srv.DiskTotal,
                history[len(history)-1].CPUUsage,
                history[len(history)-1].MemoryUsage,
                history[len(history)-1].Load1,
        )

        response, err := p.llmClient.QuickChat(prompt)
        if err != nil {
                return ""
        }

        return response
}

// Stats 统计结果
type Stats struct {
        Mean       float64
        StdDev     float64
        Min        float64
        Max        float64
        Last       float64
        Trend      string
        ChangeRate float64 // 变化率(百分比/小时)
}

// calculateStats 计算统计指标
func calculateStats(history []server.ServerMetric, getValue func(server.ServerMetric) float64) Stats {
        if len(history) == 0 {
                return Stats{}
        }

        var sum float64
        values := make([]float64, len(history))
        
        for i, m := range history {
                v := getValue(m)
                values[i] = v
                sum += v
        }

        mean := sum / float64(len(values))

        // 标准差
        var variance float64
        for _, v := range values {
                variance += (v - mean) * (v - mean)
        }
        variance /= float64(len(values))
        stdDev := math.Sqrt(variance)

        // 最小最大值
        min, max := values[0], values[0]
        for _, v := range values {
                if v < min {
                        min = v
                }
                if v > max {
                        max = v
                }
        }

        // 趋势
        trend := "stable"
        changeRate := 0.0
        if len(values) >= 2 {
                changeRate = (values[len(values)-1] - values[0]) / float64(len(values))
                if changeRate > 0.1 {
                        trend = "up"
                } else if changeRate < -0.1 {
                        trend = "down"
                }
        }

        return Stats{
                Mean:       mean,
                StdDev:     stdDev,
                Min:        min,
                Max:        max,
                Last:       values[len(values)-1],
                Trend:      trend,
                ChangeRate: changeRate,
        }
}

// calculateConfidence 计算置信度
func calculateConfidence(dataPoints int) float64 {
        if dataPoints >= 100 {
                return 0.95
        } else if dataPoints >= 50 {
                return 0.85
        } else if dataPoints >= 20 {
                return 0.75
        } else if dataPoints >= 10 {
                return 0.6
        }
        return 0.4
}
