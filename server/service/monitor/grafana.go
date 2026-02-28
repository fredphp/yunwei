package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GrafanaConfig Grafana配置
type GrafanaConfig struct {
	URL      string `json:"url"`
	APIKey   string `json:"apiKey"`
	Username string `json:"username"`
	Password string `json:"password"`
	Timeout  time.Duration `json:"timeout"`
}

// GrafanaClient Grafana客户端
type GrafanaClient struct {
	config     GrafanaConfig
	httpClient *http.Client
}

// Dashboard Grafana仪表盘
type Dashboard struct {
	ID          uint                   `json:"id"`
	UID         string                 `json:"uid"`
	Title       string                 `json:"title"`
	Tags        []string               `json:"tags"`
	Editable    bool                   `json:"editable"`
	Panels      []Panel                `json:"panels"`
	Dashboard   map[string]interface{} `json:"dashboard,omitempty"`
}

// Panel 仪表盘面板
type Panel struct {
	ID      uint                   `json:"id"`
	Title   string                 `json:"title"`
	Type    string                 `json:"type"`
	GridPos GridPos                `json:"gridPos"`
	Targets []Target               `json:"targets"`
	Options map[string]interface{} `json:"options"`
}

// GridPos 网格位置
type GridPos struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// Target 查询目标
type Target struct {
	RefID      string `json:"refId"`
	Datasource string `json:"datasource"`
	Expr       string `json:"expr"`
	Legend     string `json:"legendFormat"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID           uint                   `json:"id"`
	UID          string                 `json:"uid"`
	Title        string                 `json:"title"`
	FolderUID    string                 `json:"folderUid"`
	RuleGroup    string                 `json:"ruleGroup"`
	NoDataState  string                 `json:"noDataState"`
	ExecErrState string                 `json:"execErrState"`
	For          string                 `json:"for"`
	Annotations  map[string]string      `json:"annotations"`
	Labels       map[string]string      `json:"labels"`
	Data         []AlertCondition       `json:"data"`
}

// AlertCondition 告警条件
type AlertCondition struct {
	RefID     string                 `json:"refId"`
	QueryType string                 `json:"queryType"`
	RelativeTimeRange RelativeTimeRange `json:"relativeTimeRange"`
	DatasourceUID string              `json:"datasourceUid"`
	Model     map[string]interface{} `json:"model"`
}

// RelativeTimeRange 相对时间范围
type RelativeTimeRange struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// NewGrafanaClient 创建Grafana客户端
func NewGrafanaClient(config GrafanaConfig) *GrafanaClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &GrafanaClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// doRequest 执行HTTP请求
func (g *GrafanaClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, g.config.URL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// 认证
	if g.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.config.APIKey)
	} else if g.config.Username != "" {
		req.SetBasicAuth(g.config.Username, g.config.Password)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Grafana API错误: %s", string(respBody))
	}

	return respBody, nil
}

// GetDashboards 获取所有仪表盘
func (g *GrafanaClient) GetDashboards() ([]Dashboard, error) {
	resp, err := g.doRequest("GET", "/api/search?type=dash-db", nil)
	if err != nil {
		return nil, err
	}

	var dashboards []Dashboard
	if err := json.Unmarshal(resp, &dashboards); err != nil {
		return nil, err
	}

	return dashboards, nil
}

// GetDashboard 获取仪表盘详情
func (g *GrafanaClient) GetDashboard(uid string) (*Dashboard, error) {
	resp, err := g.doRequest("GET", "/api/dashboards/uid/"+uid, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Dashboard Dashboard `json:"dashboard"`
		Meta      struct {
			IsStarred bool   `json:"isStarred"`
			Version   int    `json:"version"`
			FolderID  int    `json:"folderId"`
		} `json:"meta"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result.Dashboard, nil
}

// CreateDashboard 创建仪表盘
func (g *GrafanaClient) CreateDashboard(dashboard Dashboard) error {
	body := map[string]interface{}{
		"dashboard": dashboard,
		"overwrite": false,
	}

	_, err := g.doRequest("POST", "/api/dashboards/db", body)
	return err
}

// CreateServerDashboard 为服务器创建监控仪表盘
func (g *GrafanaClient) CreateServerDashboard(serverID uint, serverName, host string) (*Dashboard, error) {
	dashboard := Dashboard{
		UID:      fmt.Sprintf("server-%d", serverID),
		Title:    fmt.Sprintf("服务器监控 - %s", serverName),
		Tags:     []string{"server", "monitoring"},
		Editable: true,
		Panels: []Panel{
			// CPU 使用率
			{
				ID:    1,
				Title: "CPU使用率",
				Type:  "gauge",
				GridPos: GridPos{X: 0, Y: 0, W: 8, H: 6},
				Targets: []Target{
					{
						RefID:      "A",
						Datasource: "Prometheus",
						Expr:       fmt.Sprintf(`100 - (avg by(instance) (irate(node_cpu_seconds_total{instance="%s:9100",mode="idle"}[5m])) * 100)`, host),
						Legend:     "CPU使用率",
					},
				},
			},
			// 内存使用率
			{
				ID:    2,
				Title: "内存使用率",
				Type:  "gauge",
				GridPos: GridPos{X: 8, Y: 0, W: 8, H: 6},
				Targets: []Target{
					{
						RefID:      "A",
						Datasource: "Prometheus",
						Expr:       fmt.Sprintf(`(1 - (node_memory_MemAvailable_bytes{instance="%s:9100"} / node_memory_MemTotal_bytes{instance="%s:9100"})) * 100`, host, host),
						Legend:     "内存使用率",
					},
				},
			},
			// 磁盘使用率
			{
				ID:    3,
				Title: "磁盘使用率",
				Type:  "gauge",
				GridPos: GridPos{X: 16, Y: 0, W: 8, H: 6},
				Targets: []Target{
					{
						RefID:      "A",
						Datasource: "Prometheus",
						Expr:       fmt.Sprintf(`(1 - (node_filesystem_avail_bytes{instance="%s:9100",fstype!="tmpfs"} / node_filesystem_size_bytes{instance="%s:9100",fstype!="tmpfs"})) * 100`, host, host),
						Legend:     "磁盘使用率",
					},
				},
			},
			// 系统负载
			{
				ID:    4,
				Title: "系统负载",
				Type:  "timeseries",
				GridPos: GridPos{X: 0, Y: 6, W: 12, H: 8},
				Targets: []Target{
					{
						RefID:      "A",
						Datasource: "Prometheus",
						Expr:       fmt.Sprintf(`node_load1{instance="%s:9100"}`, host),
						Legend:     "1分钟负载",
					},
					{
						RefID:      "B",
						Datasource: "Prometheus",
						Expr:       fmt.Sprintf(`node_load5{instance="%s:9100"}`, host),
						Legend:     "5分钟负载",
					},
					{
						RefID:      "C",
						Datasource: "Prometheus",
						Expr:       fmt.Sprintf(`node_load15{instance="%s:9100"}`, host),
						Legend:     "15分钟负载",
					},
				},
			},
			// 网络流量
			{
				ID:    5,
				Title: "网络流量",
				Type:  "timeseries",
				GridPos: GridPos{X: 12, Y: 6, W: 12, H: 8},
				Targets: []Target{
					{
						RefID:      "A",
						Datasource: "Prometheus",
						Expr:       fmt.Sprintf(`rate(node_network_receive_bytes_total{instance="%s:9100",device!="lo"}[5m]) * 8`, host),
						Legend:     "入站",
					},
					{
						RefID:      "B",
						Datasource: "Prometheus",
						Expr:       fmt.Sprintf(`rate(node_network_transmit_bytes_total{instance="%s:9100",device!="lo"}[5m]) * 8`, host),
						Legend:     "出站",
					},
				},
			},
			// CPU历史趋势
			{
				ID:    6,
				Title: "CPU历史趋势",
				Type:  "timeseries",
				GridPos: GridPos{X: 0, Y: 14, W: 24, H: 8},
				Targets: []Target{
					{
						RefID:      "A",
						Datasource: "Prometheus",
						Expr:       fmt.Sprintf(`100 - (avg by(instance) (irate(node_cpu_seconds_total{instance="%s:9100",mode="idle"}[5m])) * 100)`, host),
						Legend:     "CPU使用率",
					},
				},
			},
		},
	}

	err := g.CreateDashboard(dashboard)
	if err != nil {
		return nil, err
	}

	return &dashboard, nil
}

// GetAlertRules 获取告警规则
func (g *GrafanaClient) GetAlertRules() ([]AlertRule, error) {
	resp, err := g.doRequest("GET", "/api/v1/provisioning/alert-rules", nil)
	if err != nil {
		return nil, err
	}

	var rules []AlertRule
	if err := json.Unmarshal(resp, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

// CreateAlertRule 创建告警规则
func (g *GrafanaClient) CreateAlertRule(rule AlertRule) error {
	_, err := g.doRequest("POST", "/api/v1/provisioning/alert-rules", rule)
	return err
}

// CreateServerAlertRules 为服务器创建告警规则
func (g *GrafanaClient) CreateServerAlertRules(serverID uint, serverName, host string) error {
	// CPU高告警
	cpuAlert := AlertRule{
		UID:          fmt.Sprintf("cpu-high-%d", serverID),
		Title:        fmt.Sprintf("%s - CPU使用率过高", serverName),
		FolderUID:    "alerts",
		RuleGroup:    "server-alerts",
		NoDataState:  "NoData",
		ExecErrState: "Alerting",
		For:          "5m",
		Labels: map[string]string{
			"severity": "warning",
			"server":   serverName,
		},
		Annotations: map[string]string{
			"description": fmt.Sprintf("服务器 %s CPU使用率超过90%%", serverName),
		},
		Data: []AlertCondition{
			{
				RefID:     "A",
				QueryType: "",
				RelativeTimeRange: RelativeTimeRange{From: 300, To: 0},
				DatasourceUID: "prometheus",
				Model: map[string]interface{}{
					"expr": fmt.Sprintf(`100 - (avg by(instance) (irate(node_cpu_seconds_total{instance="%s:9100",mode="idle"}[5m])) * 100) > 90`, host),
					"refId": "A",
				},
			},
		},
	}

	// 内存高告警
	memAlert := AlertRule{
		UID:          fmt.Sprintf("memory-high-%d", serverID),
		Title:        fmt.Sprintf("%s - 内存使用率过高", serverName),
		FolderUID:    "alerts",
		RuleGroup:    "server-alerts",
		NoDataState:  "NoData",
		ExecErrState: "Alerting",
		For:          "5m",
		Labels: map[string]string{
			"severity": "warning",
			"server":   serverName,
		},
		Annotations: map[string]string{
			"description": fmt.Sprintf("服务器 %s 内存使用率超过90%%", serverName),
		},
		Data: []AlertCondition{
			{
				RefID:     "A",
				QueryType: "",
				RelativeTimeRange: RelativeTimeRange{From: 300, To: 0},
				DatasourceUID: "prometheus",
				Model: map[string]interface{}{
					"expr": fmt.Sprintf(`(1 - (node_memory_MemAvailable_bytes{instance="%s:9100"} / node_memory_MemTotal_bytes{instance="%s:9100"})) * 100 > 90`, host, host),
					"refId": "A",
				},
			},
		},
	}

	// 磁盘高告警
	diskAlert := AlertRule{
		UID:          fmt.Sprintf("disk-high-%d", serverID),
		Title:        fmt.Sprintf("%s - 磁盘空间不足", serverName),
		FolderUID:    "alerts",
		RuleGroup:    "server-alerts",
		NoDataState:  "NoData",
		ExecErrState: "Alerting",
		For:          "5m",
		Labels: map[string]string{
			"severity": "critical",
			"server":   serverName,
		},
		Annotations: map[string]string{
			"description": fmt.Sprintf("服务器 %s 磁盘使用率超过85%%", serverName),
		},
		Data: []AlertCondition{
			{
				RefID:     "A",
				QueryType: "",
				RelativeTimeRange: RelativeTimeRange{From: 300, To: 0},
				DatasourceUID: "prometheus",
				Model: map[string]interface{}{
					"expr": fmt.Sprintf(`(1 - (node_filesystem_avail_bytes{instance="%s:9100",fstype!="tmpfs"} / node_filesystem_size_bytes{instance="%s:9100",fstype!="tmpfs"})) * 100 > 85`, host, host),
					"refId": "A",
				},
			},
		},
	}

	// 创建告警规则
	if err := g.CreateAlertRule(cpuAlert); err != nil {
		return fmt.Errorf("创建CPU告警失败: %w", err)
	}
	if err := g.CreateAlertRule(memAlert); err != nil {
		return fmt.Errorf("创建内存告警失败: %w", err)
	}
	if err := g.CreateAlertRule(diskAlert); err != nil {
		return fmt.Errorf("创建磁盘告警失败: %w", err)
	}

	return nil
}

// GetDashboardSnapshot 获取仪表盘快照URL
func (g *GrafanaClient) GetDashboardSnapshot(uid string) (string, error) {
	body := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"uid": uid,
		},
		"expires": 86400, // 24小时有效
	}

	resp, err := g.doRequest("POST", "/api/snapshots", body)
	if err != nil {
		return "", err
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	return result.URL, nil
}
