package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Executor 命令执行器
type Executor struct {
	timeout time.Duration
}

// NewExecutor 创建执行器
func NewExecutor() *Executor {
	return &Executor{
		timeout: 300 * time.Second,
	}
}

// ExecuteResult 执行结果
type ExecuteResult struct {
	Success   bool     `json:"success"`
	Output    string   `json:"output"`
	Error     string   `json:"error"`
	Duration  int64    `json:"duration"` // 毫秒
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// Execute 执行命令
func (e *Executor) Execute(ctx context.Context, command string, timeout int) *ExecuteResult {
	result := &ExecuteResult{
		StartTime: time.Now(),
	}

	if timeout <= 0 {
		timeout = 300
	}

	// 创建带超时的上下文
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 执行命令
	cmd := exec.CommandContext(execCtx, "sh", "-c", command)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Output = stderr.String()
	} else {
		result.Success = true
		result.Output = stdout.String()
	}

	// 检查是否超时
	if execCtx.Err() == context.DeadlineExceeded {
		result.Success = false
		result.Error = fmt.Sprintf("命令执行超时 (%d秒)", timeout)
	}

	return result
}

// ExecuteScript 执行脚本
func (e *Executor) ExecuteScript(ctx context.Context, script string, timeout int) *ExecuteResult {
	// 脚本可能包含多行，需要特殊处理
	return e.Execute(ctx, script, timeout)
}

// ExecuteWithEnv 执行带环境变量的命令
func (e *Executor) ExecuteWithEnv(ctx context.Context, command string, env map[string]string, timeout int) *ExecuteResult {
	result := &ExecuteResult{
		StartTime: time.Now(),
	}

	if timeout <= 0 {
		timeout = 300
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "sh", "-c", command)
	
	// 设置环境变量
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Output = stderr.String()
	} else {
		result.Success = true
		result.Output = stdout.String()
	}

	return result
}

// CheckCommand 检查命令是否存在
func (e *Executor) CheckCommand(command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}
	
	_, err := exec.LookPath(parts[0])
	return err == nil
}

// GetCommandPath 获取命令路径
func (e *Executor) GetCommandPath(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}
	
	path, err := exec.LookPath(parts[0])
	if err != nil {
		return ""
	}
	return path
}

// SafeExecute 安全执行（带白名单检查）
func (e *Executor) SafeExecute(ctx context.Context, command string, allowedCommands []string, timeout int) *ExecuteResult {
	// 检查命令是否在白名单中
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return &ExecuteResult{
			Success: false,
			Error:   "空命令",
		}
	}

	allowed := false
	for _, cmd := range allowedCommands {
		if parts[0] == cmd {
			allowed = true
			break
		}
	}

	if !allowed {
		return &ExecuteResult{
			Success: false,
			Error:   fmt.Sprintf("命令 %s 不在白名单中", parts[0]),
		}
	}

	return e.Execute(ctx, command, timeout)
}

// ExecuteDocker 执行 Docker 命令
func (e *Executor) ExecuteDocker(ctx context.Context, action, container string, timeout int) *ExecuteResult {
	var command string
	
	switch action {
	case "start":
		command = fmt.Sprintf("docker start %s", container)
	case "stop":
		command = fmt.Sprintf("docker stop %s", container)
	case "restart":
		command = fmt.Sprintf("docker restart %s", container)
	case "logs":
		command = fmt.Sprintf("docker logs --tail 100 %s", container)
	case "inspect":
		command = fmt.Sprintf("docker inspect %s", container)
	default:
		return &ExecuteResult{
			Success: false,
			Error:   fmt.Sprintf("未知的 Docker 操作: %s", action),
		}
	}

	return e.Execute(ctx, command, timeout)
}

// ExecuteSystemctl 执行 systemctl 命令
func (e *Executor) ExecuteSystemctl(ctx context.Context, action, service string, timeout int) *ExecuteResult {
	var command string
	
	switch action {
	case "start":
		command = fmt.Sprintf("systemctl start %s", service)
	case "stop":
		command = fmt.Sprintf("systemctl stop %s", service)
	case "restart":
		command = fmt.Sprintf("systemctl restart %s", service)
	case "status":
		command = fmt.Sprintf("systemctl status %s", service)
	case "enable":
		command = fmt.Sprintf("systemctl enable %s", service)
	case "disable":
		command = fmt.Sprintf("systemctl disable %s", service)
	default:
		return &ExecuteResult{
			Success: false,
			Error:   fmt.Sprintf("未知的 systemctl 操作: %s", action),
		}
	}

	return e.Execute(ctx, command, timeout)
}

// CleanupFiles 清理文件
func (e *Executor) CleanupFiles(ctx context.Context, path string, days int) *ExecuteResult {
	command := fmt.Sprintf("find %s -type f -mtime +%d -delete", path, days)
	return e.Execute(ctx, command, 600) // 10分钟超时
}

// GetDiskUsage 获取磁盘使用情况
func (e *Executor) GetDiskUsage(ctx context.Context, path string) *ExecuteResult {
	command := fmt.Sprintf("du -sh %s", path)
	return e.Execute(ctx, command, 60)
}

// KillProcess 杀死进程
func (e *Executor) KillProcess(ctx context.Context, pid int, signal string) *ExecuteResult {
	if signal == "" {
		signal = "TERM"
	}
	command := fmt.Sprintf("kill -%s %d", signal, pid)
	return e.Execute(ctx, command, 10)
}
