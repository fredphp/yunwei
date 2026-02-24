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

// PredictionType 预测类型
type PredictionType string

const (
        PredictionCPU     PredictionType = "cpu"
        PredictionMemory  PredictionType = "memory"
        PredictionDisk    PredictionType = "disk"
        PredictionNetwork PredictionType = "network"
        PredictionLoad    PredictionType = "load"
)

// PredictionLevel 预测级别
type PredictionLevel string

const (
        PredictionLevelNormal   PredictionLevel = "normal"
        PredictionLevelWarning  PredictionLevel = "warning"
        PredictionLevelCritical PredictionLevel = "critical"
)

// PredictionResult 预测结果
type PredictionResult struct {
        ID          uint           `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time      `json:"createdAt"`
        ServerID    uint           `json:"serverId" gorm:"index"`
        Server      *server.Server `json:"server" gorm:"foreignKey:ServerID"`

        // 预测信息
        Type        PredictionType  `json:"type" gorm:"type:varchar(32)"`
        Level       PredictionLevel `json:"level" gorm:"type:varchar(16)"`
        
        // 预测值
        CurrentValue   float64   `json:"currentValue"`
        PredictedValue float64   `json:"predictedValue"`
        PredictedAt    time.Time `json:"predictedAt"` // 预测时间点
        Confidence     float64   `json:"confidence"`  // 置信度 0-1

        // 分析信息
        Trend       string  `json:"trend"`       // up, down, stable
        TrendRate   float64 `json:"trendRate"`   // 变化率
        TimeToAlert int     `json:"timeToAlert"` // 预计到达告警阈值的时间(分钟)

        // 建议
        Summary     string   `json:"summary"`
        Suggestions []string `json:"suggestions"`
        Actions     []string `json:"actions"` // 推荐操作

        // AI 分析
        AIAnalysis string `json:"aiAnalysis" gorm:"type:text"`
}

func (PredictionResult) TableName() string {
        return "prediction_results"
}

// HistoryData 历史数据点
type HistoryData struct {
        Timestamp time.Time
        Value     float64
}

// Predictor 预测器
type Predictor struct {
        llmClient *llm.GLM5Client
}

// NewPredictor 创建预测器
func NewPredictor(llmClient *llm.GLM5Client) *Predictor {
        return &Predictor{
                llmClient: llmClient,
        }
}

// PredictCPU 预测CPU使用率
func (p *Predictor) PredictCPU(serverID uint, history []server.ServerMetric) (*PredictionResult, error) {
        if len(history) < 10 {
                return nil, fmt.Errorf("历史数据不足")
        }

        // 提取CPU数据
        var data []HistoryData
        for _, m := range history {
                data = append(data, HistoryData{
                        Timestamp: m.CreatedAt,
                        Value:     m.CPUUsage,
                })
        }

        // 计算趋势
        trend, rate := p.calculateTrend(data)

        // 线性预测
        predictedValue := p.linearPredict(data, 60*time.Minute)

        // 计算置信度
        confidence := p.calculateConfidence(data)

        // 确定级别
        level := PredictionLevelNormal
        if predictedValue > 90 {
                level = PredictionLevelCritical
        } else if predictedValue > 80 {
                level = PredictionLevelWarning
        }

        // 计算到达告警阈值的时间
        timeToAlert := p.calculateTimeToAlert(data, 90)

        result := &PredictionResult{
                ServerID:       serverID,
                Type:           PredictionCPU,
                Level:          level,
                CurrentValue:   data[len(data)-1].Value,
                PredictedValue: predictedValue,
                PredictedAt:    time.Now().Add(time.Hour),
                Confidence:     confidence,
                Trend:          trend,
                TrendRate:      rate,
                TimeToAlert:    timeToAlert,
                Summary:        fmt.Sprintf("CPU使用率%s趋势，当前%.2f%%，预测1小时后%.2f%%", trend, data[len(data)-1].Value, predictedValue),
                Suggestions:    p.getCPUSuggestions(level, trend),
        }

        return result, nil
}

// PredictMemory 预测内存使用率
func (p *Predictor) PredictMemory(serverID uint, history []server.ServerMetric) (*PredictionResult, error) {
        if len(history) < 10 {
                return nil, fmt.Errorf("历史数据不足")
        }

        var data []HistoryData
        for _, m := range history {
                data = append(data, HistoryData{
                        Timestamp: m.CreatedAt,
                        Value:     m.MemoryUsage,
                })
        }

        trend, rate := p.calculateTrend(data)
        predictedValue := p.linearPredict(data, 60*time.Minute)
        confidence := p.calculateConfidence(data)

        level := PredictionLevelNormal
        if predictedValue > 90 {
                level = PredictionLevelCritical
        } else if predictedValue > 80 {
                level = PredictionLevelWarning
        }

        timeToAlert := p.calculateTimeToAlert(data, 90)

        result := &PredictionResult{
                ServerID:       serverID,
                Type:           PredictionMemory,
                Level:          level,
                CurrentValue:   data[len(data)-1].Value,
                PredictedValue: predictedValue,
                PredictedAt:    time.Now().Add(time.Hour),
                Confidence:     confidence,
                Trend:          trend,
                TrendRate:      rate,
                TimeToAlert:    timeToAlert,
                Summary:        fmt.Sprintf("内存使用率%s趋势，当前%.2f%%，预测1小时后%.2f%%", trend, data[len(data)-1].Value, predictedValue),
                Suggestions:    p.getMemorySuggestions(level, trend),
        }

        return result, nil
}

// PredictDisk 预测磁盘使用率
func (p *Predictor) PredictDisk(serverID uint, history []server.ServerMetric) (*PredictionResult, error) {
        if len(history) < 10 {
                return nil, fmt.Errorf("历史数据不足")
        }

        var data []HistoryData
        for _, m := range history {
                data = append(data, HistoryData{
                        Timestamp: m.CreatedAt,
                        Value:     m.DiskUsage,
                })
        }

        trend, rate := p.calculateTrend(data)
        
        // 磁盘预测更长时间
        predictedValue := p.linearPredict(data, 24*time.Hour)
        confidence := p.calculateConfidence(data)

        level := PredictionLevelNormal
        if predictedValue > 95 {
                level = PredictionLevelCritical
        } else if predictedValue > 85 {
                level = PredictionLevelWarning
        }

        timeToAlert := p.calculateTimeToAlert(data, 90)

        // 计算磁盘满的时间
        diskFullTime := p.calculateDiskFullTime(data)

        result := &PredictionResult{
                ServerID:       serverID,
                Type:           PredictionDisk,
                Level:          level,
                CurrentValue:   data[len(data)-1].Value,
                PredictedValue: predictedValue,
                PredictedAt:    time.Now().Add(24 * time.Hour),
                Confidence:     confidence,
                Trend:          trend,
                TrendRate:      rate,
                TimeToAlert:    timeToAlert,
                Summary:        fmt.Sprintf("磁盘使用率%s趋势，当前%.2f%%，预测24小时后%.2f%%", trend, data[len(data)-1].Value, predictedValue),
                Suggestions:    p.getDiskSuggestions(level, diskFullTime),
        }

        return result, nil
}

// PredictNetwork 预测网络流量峰值
func (p *Predictor) PredictNetwork(serverID uint, history []server.ServerMetric) (*PredictionResult, error) {
        if len(history) < 10 {
                return nil, fmt.Errorf("历史数据不足")
        }

        var data []HistoryData
        for _, m := range history {
                data = append(data, HistoryData{
                        Timestamp: m.CreatedAt,
                        Value:     float64(m.NetIn + m.NetOut),
                })
        }

        trend, rate := p.calculateTrend(data)
        predictedValue := p.linearPredict(data, 60*time.Minute)
        confidence := p.calculateConfidence(data)

        // 获取历史峰值
        maxValue := 0.0
        for _, d := range data {
                if d.Value > maxValue {
                        maxValue = d.Value
                }
        }

        level := PredictionLevelNormal
        if predictedValue > maxValue*1.5 {
                level = PredictionLevelWarning
        }

        result := &PredictionResult{
                ServerID:       serverID,
                Type:           PredictionNetwork,
                Level:          level,
                CurrentValue:   data[len(data)-1].Value,
                PredictedValue: predictedValue,
                PredictedAt:    time.Now().Add(time.Hour),
                Confidence:     confidence,
                Trend:          trend,
                TrendRate:      rate,
                Summary:        fmt.Sprintf("网络流量%s趋势，当前%.2f Mbps，预测1小时后%.2f Mbps", trend, data[len(data)-1].Value/1000000, predictedValue/1000000),
                Suggestions:    p.getNetworkSuggestions(level),
        }

        return result, nil
}

// AIPredict AI预测分析
func (p *Predictor) AIPredict(serverID uint, history []server.ServerMetric) (*PredictionResult, error) {
        if p.llmClient == nil {
                return nil, fmt.Errorf("AI客户端未配置")
        }

        // 构建历史数据摘要
        var dataStr string
        for i, m := range history {
                if i >= 20 {
                        break
                }
                dataStr += fmt.Sprintf("[%s] CPU:%.1f%% Mem:%.1f%% Disk:%.1f%% Load:%.2f\n",
                        m.CreatedAt.Format("15:04"), m.CPUUsage, m.MemoryUsage, m.DiskUsage, m.Load1)
        }

        prompt := fmt.Sprintf(`作为运维专家，分析以下服务器历史数据并预测未来趋势：

服务器历史数据（最近20条）：
%s

请分析：
1. 各指标趋势（上升/下降/稳定）
2. 预测1小时后的资源使用情况
3. 可能出现的异常风险
4. 推荐的预防措施

请按JSON格式返回：
{
  "trends": {"cpu": "up/down/stable", "memory": "...", "disk": "..."},
  "predictions": {"cpu": 数值, "memory": 数值, "disk": 数值},
  "risks": ["风险1", "风险2"],
  "suggestions": ["建议1", "建议2"],
  "actions": ["操作1", "操作2"]
}`, dataStr)

        response, err := p.llmClient.QuickChat(prompt)
        if err != nil {
                return nil, fmt.Errorf("AI分析失败: %w", err)
        }

        // 解析AI响应
        result := &PredictionResult{
                ServerID:    serverID,
                Type:        PredictionCPU, // 综合
                Level:       PredictionLevelNormal,
                AIAnalysis:  response,
                CreatedAt:   time.Now(),
        }

        // 尝试解析JSON
        jsonStart := indexOf(response, "{")
        jsonEnd := lastIndexOf(response, "}")
        if jsonStart >= 0 && jsonEnd > jsonStart {
                jsonStr := response[jsonStart : jsonEnd+1]
                var aiResult struct {
                        Trends      map[string]string `json:"trends"`
                        Predictions map[string]float64 `json:"predictions"`
                        Risks       []string `json:"risks"`
                        Suggestions []string `json:"suggestions"`
                        Actions     []string `json:"actions"`
                }
                if err := json.Unmarshal([]byte(jsonStr), &aiResult); err == nil {
                        result.Suggestions = aiResult.Suggestions
                        result.Actions = aiResult.Actions
                        if len(aiResult.Risks) > 0 {
                                result.Level = PredictionLevelWarning
                        }
                }
        }

        return result, nil
}

// calculateTrend 计算趋势
func (p *Predictor) calculateTrend(data []HistoryData) (string, float64) {
        if len(data) < 2 {
                return "stable", 0
        }

        // 线性回归
        n := float64(len(data))
        sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

        for i, d := range data {
                x := float64(i)
                y := d.Value
                sumX += x
                sumY += y
                sumXY += x * y
                sumX2 += x * x
        }

        slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
        
        if slope > 0.1 {
                return "up", slope
        } else if slope < -0.1 {
                return "down", slope
        }
        return "stable", slope
}

// linearPredict 线性预测
func (p *Predictor) linearPredict(data []HistoryData, duration time.Duration) float64 {
        if len(data) < 2 {
                return data[len(data)-1].Value
        }

        // 简单线性外推
        n := len(data)
        first := data[0]
        last := data[n-1]

        timeDiff := last.Timestamp.Sub(first.Timestamp).Minutes()
        if timeDiff == 0 {
                return last.Value
        }

        valueDiff := last.Value - first.Value
        rate := valueDiff / timeDiff // 每分钟变化量

        // 预测
        futureMinutes := duration.Minutes()
        predicted := last.Value + rate*futureMinutes

        // 限制范围
        if predicted < 0 {
                predicted = 0
        } else if predicted > 100 {
                predicted = 100
        }

        return predicted
}

// calculateConfidence 计算置信度
func (p *Predictor) calculateConfidence(data []HistoryData) float64 {
        if len(data) < 5 {
                return 0.3
        }

        // 计算标准差
        var sum, mean, variance float64
        for _, d := range data {
                sum += d.Value
        }
        mean = sum / float64(len(data))

        for _, d := range data {
                variance += math.Pow(d.Value-mean, 2)
        }
        variance /= float64(len(data))
        stdDev := math.Sqrt(variance)

        // 标准差越小，置信度越高
        if stdDev < 5 {
                return 0.9
        } else if stdDev < 10 {
                return 0.7
        } else if stdDev < 20 {
                return 0.5
        }
        return 0.3
}

// calculateTimeToAlert 计算到达告警阈值的时间
func (p *Predictor) calculateTimeToAlert(data []HistoryData, threshold float64) int {
        if len(data) < 2 {
                return -1
        }

        last := data[len(data)-1]
        if last.Value >= threshold {
                return 0
        }

        trend, rate := p.calculateTrend(data)
        if trend != "up" || rate <= 0 {
                return -1 // 不会到达
        }

        // 估算时间
        timeToAlert := (threshold - last.Value) / rate
        if timeToAlert < 0 {
                return -1
        }

        return int(timeToAlert)
}

// calculateDiskFullTime 计算磁盘满的时间
func (p *Predictor) calculateDiskFullTime(data []HistoryData) int {
        if len(data) < 2 {
                return -1
        }

        trend, rate := p.calculateTrend(data)
        if trend != "up" || rate <= 0 {
                return -1
        }

        last := data[len(data)-1]
        timeToFull := (100 - last.Value) / rate

        return int(timeToFull)
}

// getCPUSuggestions 获取CPU建议
func (p *Predictor) getCPUSuggestions(level PredictionLevel, trend string) []string {
        if level == PredictionLevelCritical {
                return []string{
                        "立即检查CPU密集型进程",
                        "考虑重启异常服务",
                        "清理系统缓存",
                        "评估是否需要扩容",
                }
        } else if level == PredictionLevelWarning {
                return []string{
                        "监控CPU使用趋势",
                        "优化高CPU进程",
                        "准备扩容方案",
                }
        }
        return []string{
                "CPU使用率正常",
                "继续监控趋势",
        }
}

// getMemorySuggestions 获取内存建议
func (p *Predictor) getMemorySuggestions(level PredictionLevel, trend string) []string {
        if level == PredictionLevelCritical {
                return []string{
                        "立即释放内存缓存",
                        "检查内存泄漏",
                        "重启占用内存过高的服务",
                }
        } else if level == PredictionLevelWarning {
                return []string{
                        "监控内存使用趋势",
                        "检查内存密集型进程",
                }
        }
        return []string{
                "内存使用率正常",
        }
}

// getDiskSuggestions 获取磁盘建议
func (p *Predictor) getDiskSuggestions(level PredictionLevel, diskFullTime int) []string {
        if level == PredictionLevelCritical {
                return []string{
                        "立即清理磁盘空间",
                        "清理Docker无用镜像和容器",
                        "清理系统日志",
                        "考虑扩容磁盘",
                }
        } else if level == PredictionLevelWarning {
                return []string{
                        "规划磁盘清理任务",
                        "检查大文件和日志",
                        "预测磁盘满时间：" + fmt.Sprintf("%d分钟", diskFullTime),
                }
        }
        return []string{
                "磁盘空间充足",
        }
}

// getNetworkSuggestions 获取网络建议
func (p *Predictor) getNetworkSuggestions(level PredictionLevel) []string {
        if level == PredictionLevelWarning {
                return []string{
                        "监控网络流量",
                        "检查异常流量来源",
                        "考虑带宽扩容",
                }
        }
        return []string{
                "网络流量正常",
        }
}

// Helper functions
func indexOf(s string, substr string) int {
        for i := 0; i <= len(s)-len(substr); i++ {
                if s[i:i+len(substr)] == substr {
                        return i
                }
        }
        return -1
}

func lastIndexOf(s string, substr string) int {
        for i := len(s) - len(substr); i >= 0; i-- {
                if s[i:i+len(substr)] == substr {
                        return i + len(substr) - 1
                }
        }
        return -1
}

// Save 保存预测结果
func (p *Predictor) Save(result *PredictionResult) error {
        return global.DB.Create(result).Error
}

// GetPredictions 获取预测历史
func (p *Predictor) GetPredictions(serverID uint, limit int) ([]PredictionResult, error) {
        var results []PredictionResult
        query := global.DB.Model(&PredictionResult{}).Order("created_at DESC")
        if serverID > 0 {
                query = query.Where("server_id = ?", serverID)
        }
        if limit > 0 {
                query = query.Limit(limit)
        }
        err := query.Find(&results).Error
        return results, err
}
