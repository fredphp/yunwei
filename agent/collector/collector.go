package collector

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config 采集器配置
type Config struct {
	EnableDocker bool
	EnablePorts  bool
}

// Metric 指标数据
type Metric struct {
	Name      string
	Value     float64
	Type      string // gauge, counter
	Timestamp time.Time
	Tags      map[string]string
}

// Collector 指标采集器
type Collector struct {
	config     *Config
	hostname   string
	os         string
	arch       string
}

// NewCollector 创建采集器
func NewCollector(config *Config) *Collector {
	hostname, _ := os.Hostname()
	
	return &Collector{
		config:   config,
		hostname: hostname,
		os:       runtime.GOOS,
		arch:     runtime.GOARCH,
	}
}

// Collect 采集所有指标
func (c *Collector) Collect() *CollectResult {
	now := time.Now()
	result := &CollectResult{
		Timestamp: now,
		Hostname:  c.hostname,
		OS:        c.os,
		Arch:      c.arch,
	}

	// CPU 指标
	c.collectCPU(result)

	// 内存指标
	c.collectMemory(result)

	// 磁盘指标
	c.collectDisk(result)

	// 网络指标
	c.collectNetwork(result)

	// 负载指标
	c.collectLoad(result)

	// 进程指标
	c.collectProcesses(result)

	// Docker 容器
	if c.config.EnableDocker {
		c.collectDocker(result)
	}

	// 端口占用
	if c.config.EnablePorts {
		c.collectPorts(result)
	}

	return result
}

// CollectResult 采集结果
type CollectResult struct {
	Timestamp   time.Time          `json:"timestamp"`
	Hostname    string             `json:"hostname"`
	OS          string             `json:"os"`
	Arch        string             `json:"arch"`
	
	// CPU
	CPUUsage    float64            `json:"cpuUsage"`
	CPUUser     float64            `json:"cpuUser"`
	CPUSystem   float64            `json:"cpuSystem"`
	CPUIdle     float64            `json:"cpuIdle"`
	CPUIowait   float64            `json:"cpuIowait"`
	CPUCores    int                `json:"cpuCores"`
	
	// 内存
	MemoryTotal uint64             `json:"memoryTotal"`
	MemoryUsed  uint64             `json:"memoryUsed"`
	MemoryFree  uint64             `json:"memoryFree"`
	MemoryCache uint64             `json:"memoryCache"`
	MemoryUsage float64            `json:"memoryUsage"`
	SwapTotal   uint64             `json:"swapTotal"`
	SwapUsed    uint64             `json:"swapUsed"`
	SwapUsage   float64            `json:"swapUsage"`
	
	// 磁盘
	DiskTotal   uint64             `json:"diskTotal"`
	DiskUsed    uint64             `json:"diskUsed"`
	DiskFree    uint64             `json:"diskFree"`
	DiskUsage   float64            `json:"diskUsage"`
	DiskRead    uint64             `json:"diskRead"`
	DiskWrite   uint64             `json:"diskWrite"`
	
	// 网络
	NetIn       uint64             `json:"netIn"`
	NetOut      uint64             `json:"netOut"`
	NetInPps    uint64             `json:"netInPps"`
	NetOutPps   uint64             `json:"netOutPps"`
	
	// 负载
	Load1       float64            `json:"load1"`
	Load5       float64            `json:"load5"`
	Load15      float64            `json:"load15"`
	
	// 进程
	ProcessCount int               `json:"processCount"`
	Processes   []ProcessInfo     `json:"processes"`
	
	// Docker
	Containers  []ContainerInfo   `json:"containers"`
	
	// 端口
	Ports       []PortInfo        `json:"ports"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID     int     `json:"pid"`
	Name    string  `json:"name"`
	User    string  `json:"user"`
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	State   string  `json:"state"`
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	Status      string  `json:"status"`
	State       string  `json:"state"`
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
}

// PortInfo 端口信息
type PortInfo struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	PID      int    `json:"pid"`
	Process  string `json:"process"`
	State    string `json:"state"`
}

// collectCPU 采集CPU指标
func (c *Collector) collectCPU(result *CollectResult) {
	// 读取 /proc/stat
	file, err := os.Open("/proc/stat")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) >= 8 {
				user, _ := strconv.ParseFloat(fields[1], 64)
				nice, _ := strconv.ParseFloat(fields[2], 64)
				system, _ := strconv.ParseFloat(fields[3], 64)
				idle, _ := strconv.ParseFloat(fields[4], 64)
				iowait, _ := strconv.ParseFloat(fields[5], 64)
				
				total := user + nice + system + idle + iowait
				if total > 0 {
					result.CPUUsage = (user + nice + system) / total * 100
					result.CPUUser = user / total * 100
					result.CPUSystem = system / total * 100
					result.CPUIdle = idle / total * 100
					result.CPUIowait = iowait / total * 100
				}
			}
		}
	}

	// CPU 核心数
	result.CPUCores = runtime.NumCPU()
}

// collectMemory 采集内存指标
func (c *Collector) collectMemory(result *CollectResult) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			value, _ := strconv.ParseUint(fields[1], 10, 64)
			
			switch strings.TrimSuffix(fields[0], ":") {
			case "MemTotal":
				result.MemoryTotal = value / 1024 // KB -> MB
			case "MemFree":
				result.MemoryFree = value / 1024
			case "MemAvailable":
				// 可用内存
			case "Buffers", "Cached":
				result.MemoryCache += value / 1024
			case "SwapTotal":
				result.SwapTotal = value / 1024
			case "SwapFree":
				// 交换分区空闲
			}
		}
	}

	// 计算使用率和已用内存
	if result.MemoryTotal > 0 {
		result.MemoryUsed = result.MemoryTotal - result.MemoryFree
		result.MemoryUsage = float64(result.MemoryUsed) / float64(result.MemoryTotal) * 100
	}
	if result.SwapTotal > 0 {
		result.SwapUsed = result.SwapTotal
		result.SwapUsage = float64(result.SwapUsed) / float64(result.SwapTotal) * 100
	}
}

// collectDisk 采集磁盘指标
func (c *Collector) collectDisk(result *CollectResult) {
	// 使用 df 命令
	output, err := exec.Command("df", "-BG", "/").Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 4 {
			result.DiskTotal, _ = strconv.ParseUint(strings.TrimSuffix(fields[1], "G"), 10, 64)
			result.DiskUsed, _ = strconv.ParseUint(strings.TrimSuffix(fields[2], "G"), 10, 64)
			result.DiskFree, _ = strconv.ParseUint(strings.TrimSuffix(fields[3], "G"), 10, 64)
			usageStr := strings.TrimSuffix(fields[4], "%")
			result.DiskUsage, _ = strconv.ParseFloat(usageStr, 64)
		}
	}

	// 磁盘IO
	ioOutput, err := exec.Command("cat", "/proc/diskstats").Output()
	if err == nil {
		_ = string(ioOutput) // 解析磁盘IO统计
	}
}

// collectNetwork 采集网络指标
func (c *Collector) collectNetwork(result *CollectResult) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				iface := strings.TrimSpace(parts[0])
				// 跳过 lo 接口
				if iface == "lo" {
					continue
				}
				fields := strings.Fields(strings.TrimSpace(parts[1]))
				if len(fields) >= 10 {
					recv, _ := strconv.ParseUint(fields[0], 10, 64)
					send, _ := strconv.ParseUint(fields[8], 10, 64)
					result.NetIn += recv
					result.NetOut += send
				}
			}
		}
	}
}

// collectLoad 采集负载指标
func (c *Collector) collectLoad(result *CollectResult) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			result.Load1, _ = strconv.ParseFloat(fields[0], 64)
			result.Load5, _ = strconv.ParseFloat(fields[1], 64)
			result.Load15, _ = strconv.ParseFloat(fields[2], 64)
		}
	}
}

// collectProcesses 采集进程指标
func (c *Collector) collectProcesses(result *CollectResult) {
	// 进程数量
	entries, err := os.ReadDir("/proc")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				if _, err := strconv.Atoi(entry.Name()); err == nil {
					result.ProcessCount++
				}
			}
		}
	}

	// TOP 10 进程
	output, err := exec.Command("sh", "-c", "ps aux --sort=-%cpu | head -11 | tail -10").Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 11 {
			proc := ProcessInfo{
				User:   fields[0],
				State:  fields[7],
			}
			proc.PID, _ = strconv.Atoi(fields[1])
			proc.CPU, _ = strconv.ParseFloat(fields[2], 64)
			proc.Memory, _ = strconv.ParseFloat(fields[3], 64)
			proc.Name = fields[10]
			result.Processes = append(result.Processes, proc)
		}
	}
}

// collectDocker 采集 Docker 容器指标
func (c *Collector) collectDocker(result *CollectResult) {
	// 检查 docker 命令是否存在
	if _, err := exec.LookPath("docker"); err != nil {
		return
	}

	// 获取容器列表
	output, err := exec.Command("docker", "ps", "-a", "--format", 
		"{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.State}}").Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			container := ContainerInfo{
				ID:     parts[0],
				Name:   parts[1],
				Image:  parts[2],
				Status: parts[3],
				State:  parts[4],
			}

			// 获取容器资源使用
			statsOutput, err := exec.Command("docker", "stats", "--no-stream", 
				"--format", "{{.CPUPerc}}|{{.MemPerc}}", parts[0]).Output()
			if err == nil {
				statsParts := strings.Split(string(statsOutput), "|")
				if len(statsParts) >= 2 {
					container.CPUUsage, _ = strconv.ParseFloat(strings.TrimSuffix(statsParts[0], "%"), 64)
					container.MemoryUsage, _ = strconv.ParseFloat(strings.TrimSuffix(statsParts[1], "%"), 64)
				}
			}

			result.Containers = append(result.Containers, container)
		}
	}
}

// collectPorts 采集端口占用
func (c *Collector) collectPorts(result *CollectResult) {
	// 使用 ss 或 netstat 命令
	var output []byte
	var err error

	if _, err = exec.LookPath("ss"); err == nil {
		output, err = exec.Command("ss", "-tulnp").Output()
	} else if _, err = exec.LookPath("netstat"); err == nil {
		output, err = exec.Command("netstat", "-tulnp").Output()
	}

	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "LISTEN") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				var protocol, port, pid, process string

				if strings.HasPrefix(fields[0], "tcp") {
					protocol = "tcp"
				} else if strings.HasPrefix(fields[0], "udp") {
					protocol = "udp"
				}

				// 解析端口
				localAddr := fields[3]
				if strings.Contains(localAddr, ":") {
					portParts := strings.Split(localAddr, ":")
					port = portParts[len(portParts)-1]
				}

				// 解析进程
				if len(fields) >= 6 {
					procInfo := fields[5]
					if strings.Contains(procInfo, ",") {
						procParts := strings.Split(procInfo, ",")
						pid = procParts[0]
						if len(procParts) > 1 {
							process = procParts[1]
						}
					}
				}

				portInt, _ := strconv.Atoi(port)
				pidInt, _ := strconv.Atoi(pid)

				result.Ports = append(result.Ports, PortInfo{
					Port:     portInt,
					Protocol: protocol,
					PID:      pidInt,
					Process:  process,
					State:    "LISTEN",
				})
			}
		}
	}
}

// GetHostname 获取主机名
func (c *Collector) GetHostname() string {
	return c.hostname
}

// GetOS 获取操作系统
func (c *Collector) GetOS() string {
	return c.os
}

// GetArch 获取架构
func (c *Collector) GetArch() string {
	return c.arch
}

// Helper function
func parseUint(s string) uint64 {
	val, _ := strconv.ParseUint(s, 10, 64)
	return val
}

func parseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

// PrintResult 打印采集结果
func PrintResult(result *CollectResult) {
	fmt.Println("========================================")
	fmt.Printf("时间: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("主机: %s (%s/%s)\n", result.Hostname, result.OS, result.Arch)
	fmt.Println("----------------------------------------")
	fmt.Printf("CPU: %.1f%% (用户: %.1f%%, 系统: %.1f%%)\n", 
		result.CPUUsage, result.CPUUser, result.CPUSystem)
	fmt.Printf("内存: %.1f%% (%d/%d MB)\n", 
		result.MemoryUsage, result.MemoryUsed, result.MemoryTotal)
	fmt.Printf("磁盘: %.1f%% (%d/%d GB)\n", 
		result.DiskUsage, result.DiskUsed, result.DiskTotal)
	fmt.Printf("负载: %.2f, %.2f, %.2f\n", 
		result.Load1, result.Load5, result.Load15)
	fmt.Printf("进程数: %d\n", result.ProcessCount)
	fmt.Printf("容器数: %d\n", len(result.Containers))
	fmt.Printf("监听端口: %d\n", len(result.Ports))
	fmt.Println("========================================")
}
