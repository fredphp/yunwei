package ssh

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient SSH 客户端
type SSHClient struct {
	Host       string
	Port       int
	User       string
	Password   string
	PrivateKey string
	client     *ssh.Client
}

// SystemInfo 系统信息
type SystemInfo struct {
	Hostname     string
	OS           string
	Arch         string
	Kernel       string
	CPUCores     int
	MemoryTotal  uint64
	DiskTotal    uint64
	CPUUsage     float64
	MemoryUsage  float64
	DiskUsage    float64
	Load1        float64
	Load5        float64
	Load15       float64
	Uptime       string
}

// Connect 连接服务器
func (c *SSHClient) Connect() error {
	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: 10 * time.Second,
	}

	// 密码认证
	if c.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(c.Password))
	}

	// 密钥认证
	if c.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(c.PrivateKey))
		if err != nil {
			return fmt.Errorf("解析私钥失败: %w", err)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	if len(config.Auth) == 0 {
		return fmt.Errorf("请提供密码或私钥")
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	c.client = client
	return nil
}

// Close 关闭连接
func (c *SSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ExecuteCommand 执行命令
func (c *SSHClient) ExecuteCommand(cmd string, timeout int) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("未连接")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	// 设置超时
	if timeout > 0 {
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			session.Close()
		}()
	}

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

// GetSystemInfo 获取系统信息
func (c *SSHClient) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// 获取主机名
	hostname, _ := c.ExecuteCommand("hostname", 5)
	info.Hostname = strings.TrimSpace(hostname)

	// 获取系统信息
	osInfo, _ := c.ExecuteCommand("cat /etc/os-release | grep PRETTY_NAME | cut -d'\"' -f2", 5)
	info.OS = strings.TrimSpace(osInfo)

	// 获取架构
	arch, _ := c.ExecuteCommand("uname -m", 5)
	info.Arch = strings.TrimSpace(arch)

	// 获取内核版本
	kernel, _ := c.ExecuteCommand("uname -r", 5)
	info.Kernel = strings.TrimSpace(kernel)

	// 获取CPU核心数
	cpuCores, _ := c.ExecuteCommand("nproc", 5)
	fmt.Sscanf(strings.TrimSpace(cpuCores), "%d", &info.CPUCores)

	// 获取内存总量
	memInfo, _ := c.ExecuteCommand("cat /proc/meminfo | grep MemTotal | awk '{print $2}'", 5)
	fmt.Sscanf(strings.TrimSpace(memInfo), "%d", &info.MemoryTotal)
	info.MemoryTotal = info.MemoryTotal / 1024 // 转换为 MB

	// 获取磁盘总量
	diskInfo, _ := c.ExecuteCommand("df -BG / | tail -1 | awk '{print $2}'", 5)
	fmt.Sscanf(strings.TrimSpace(diskInfo), "%d", &info.DiskTotal)

	// 获取CPU使用率
	cpuUsage, _ := c.ExecuteCommand("top -bn1 | grep 'Cpu(s)' | awk '{print $2}'", 5)
	fmt.Sscanf(strings.TrimSpace(cpuUsage), "%f", &info.CPUUsage)

	// 获取内存使用率
	memUsage, _ := c.ExecuteCommand("free | grep Mem | awk '{print ($3/$2)*100}'", 5)
	fmt.Sscanf(strings.TrimSpace(memUsage), "%f", &info.MemoryUsage)

	// 获取磁盘使用率
	diskUsage, _ := c.ExecuteCommand("df -h / | tail -1 | awk '{print $5}' | tr -d '%'", 5)
	fmt.Sscanf(strings.TrimSpace(diskUsage), "%f", &info.DiskUsage)

	// 获取负载
	loadInfo, _ := c.ExecuteCommand("cat /proc/loadavg | awk '{print $1, $2, $3}'", 5)
	fmt.Sscanf(strings.TrimSpace(loadInfo), "%f %f %f", &info.Load1, &info.Load5, &info.Load15)

	return info, nil
}

// GetCPUUsage 获取CPU使用率
func (c *SSHClient) GetCPUUsage() (float64, error) {
	output, err := c.ExecuteCommand("top -bn1 | grep 'Cpu(s)' | awk '{print $2}'", 5)
	if err != nil {
		return 0, err
	}
	var usage float64
	fmt.Sscanf(strings.TrimSpace(output), "%f", &usage)
	return usage, nil
}

// GetMemoryUsage 获取内存使用情况
func (c *SSHClient) GetMemoryUsage() (used, free, total uint64, usage float64, err error) {
	output, err := c.ExecuteCommand("free -m | grep Mem", 5)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	fields := strings.Fields(output)
	if len(fields) >= 4 {
		fmt.Sscanf(fields[1], "%d", &total)
		fmt.Sscanf(fields[2], "%d", &used)
		fmt.Sscanf(fields[3], "%d", &free)
		if total > 0 {
			usage = float64(used) / float64(total) * 100
		}
	}
	return
}

// GetDiskUsage 获取磁盘使用情况
func (c *SSHClient) GetDiskUsage() (used, free, total uint64, usage float64, err error) {
	output, err := c.ExecuteCommand("df -BG / | tail -1", 5)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	fields := strings.Fields(output)
	if len(fields) >= 4 {
		fmt.Sscanf(fields[1], "%d", &total)
		fmt.Sscanf(fields[2], "%d", &used)
		fmt.Sscanf(fields[3], "%d", &free)
		usageStr := strings.TrimSuffix(fields[4], "%")
		fmt.Sscanf(usageStr, "%f", &usage)
	}
	return
}

// GetDockerContainers 获取 Docker 容器列表
func (c *SSHClient) GetDockerContainers() ([]map[string]string, error) {
	output, err := c.ExecuteCommand("docker ps -a --format '{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.State}}'", 10)
	if err != nil {
		return nil, err
	}

	var containers []map[string]string
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			containers = append(containers, map[string]string{
				"id":     parts[0],
				"name":   parts[1],
				"image":  parts[2],
				"status": parts[3],
				"state":  parts[4],
			})
		}
	}
	return containers, nil
}

// GetDockerStats 获取 Docker 容器资源使用
func (c *SSHClient) GetDockerStats(containerID string) (map[string]string, error) {
	output, err := c.ExecuteCommand(
		fmt.Sprintf("docker stats %s --no-stream --format '{{.CPUPerc}}|{{.MemPerc}}|{{.NetIO}}|{{.BlockIO}}'", containerID), 
		10)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(output), "|")
	if len(parts) >= 4 {
		return map[string]string{
			"cpu":    strings.TrimSuffix(parts[0], "%"),
			"memory": strings.TrimSuffix(parts[1], "%"),
			"netIO":  parts[2],
			"diskIO": parts[3],
		}, nil
	}
	return nil, fmt.Errorf("解析失败")
}

// GetPorts 获取端口占用
func (c *SSHClient) GetPorts() ([]map[string]interface{}, error) {
	output, err := c.ExecuteCommand("netstat -tulnp 2>/dev/null || ss -tulnp", 10)
	if err != nil {
		return nil, err
	}

	var ports []map[string]interface{}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "LISTEN") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				// 解析端口号
				localAddr := fields[3]
				var port string
				if strings.Contains(localAddr, ":") {
					parts := strings.Split(localAddr, ":")
					port = parts[len(parts)-1]
				}

				// 解析进程
				var pid, process string
				if len(fields) >= 6 {
					progInfo := fields[5]
					if strings.Contains(progInfo, "/") {
						progParts := strings.Split(progInfo, "/")
						pid = strings.TrimSuffix(progParts[0], "/")
						if len(progParts) > 1 {
							process = progParts[1]
						}
					}
				}

				protocol := "tcp"
				if strings.HasPrefix(fields[0], "udp") {
					protocol = "udp"
				}

				ports = append(ports, map[string]interface{}{
					"port":     port,
					"protocol": protocol,
					"pid":      pid,
					"process":  process,
					"state":    "LISTEN",
				})
			}
		}
	}
	return ports, nil
}

// GetProcesses 获取进程列表
func (c *SSHClient) GetProcesses(topN int) ([]map[string]interface{}, error) {
	if topN == 0 {
		topN = 10
	}

	output, err := c.ExecuteCommand(
		fmt.Sprintf("ps aux --sort=-%%cpu | head -%d | tail -n +2", topN+1), 10)
	if err != nil {
		return nil, err
	}

	var processes []map[string]interface{}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 11 {
			processes = append(processes, map[string]interface{}{
				"user":    fields[0],
				"pid":     fields[1],
				"cpu":     fields[2],
				"memory":  fields[3],
				"vsz":     fields[4],
				"rss":     fields[5],
				"stat":    fields[7],
				"start":   fields[8],
				"time":    fields[9],
				"command": strings.Join(fields[10:], " "),
			})
		}
	}
	return processes, nil
}

// Ping 检测连接
func Ping(host string, port int, timeout time.Duration) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
