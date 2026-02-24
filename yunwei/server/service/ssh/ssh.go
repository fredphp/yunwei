package ssh

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// AuthType 认证类型
type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeKey      AuthType = "key"       // 私钥字符串
	AuthTypePEMFile  AuthType = "pem_file"  // PEM文件路径
	AuthTypeKeyFile  AuthType = "key_file"  // 私钥文件路径
	AuthTypeAgent    AuthType = "agent"     // SSH Agent
)

// SSHConfig SSH配置
type SSHConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	User         string        `json:"user"`
	AuthType     AuthType      `json:"authType"`
	Password     string        `json:"password,omitempty"`
	PrivateKey   string        `json:"privateKey,omitempty"`   // 私钥内容
	KeyFile      string        `json:"keyFile,omitempty"`      // 私钥文件路径
	KeyPassphrase string       `json:"keyPassphrase,omitempty"` // 私钥密码
	Timeout      time.Duration `json:"timeout"`
}

// SSHClient SSH客户端
type SSHClient struct {
	config *SSHConfig
	client *ssh.Client
}

// SSHResult 执行结果
type SSHResult struct {
	Success   bool     `json:"success"`
	Output    string   `json:"output"`
	Error     string   `json:"error"`
	ExitCode  int      `json:"exitCode"`
	Duration  int64    `json:"duration"` // 毫秒
	Timestamp time.Time `json:"timestamp"`
}

// NewSSHClient 创建SSH客户端
func NewSSHClient(config *SSHConfig) (*SSHClient, error) {
	if config.Port == 0 {
		config.Port = 22
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &SSHClient{config: config}, nil
}

// Connect 连接服务器
func (c *SSHClient) Connect() error {
	// 获取认证方法
	authMethod, err := c.getAuthMethod()
	if err != nil {
		return fmt.Errorf("获取认证方法失败: %w", err)
	}

	// SSH配置
	sshConfig := &ssh.ClientConfig{
		User:            c.config.User,
		Auth:            []ssh.AuthMethod{authMethod},
		Timeout:         c.config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应使用已知主机验证
	}

	// 连接
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	c.client = client
	return nil
}

// getAuthMethod 获取认证方法
func (c *SSHClient) getAuthMethod() (ssh.AuthMethod, error) {
	switch c.config.AuthType {
	case AuthTypePassword:
		if c.config.Password == "" {
			return nil, fmt.Errorf("密码未设置")
		}
		return ssh.Password(c.config.Password), nil

	case AuthTypeKey:
		// 直接使用私钥字符串
		if c.config.PrivateKey == "" {
			return nil, fmt.Errorf("私钥未设置")
		}
		return c.parsePrivateKey([]byte(c.config.PrivateKey), c.config.KeyPassphrase)

	case AuthTypePEMFile, AuthTypeKeyFile:
		// 从文件读取私钥
		keyPath := c.config.KeyFile
		if keyPath == "" {
			return nil, fmt.Errorf("私钥文件路径未设置")
		}
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("读取私钥文件失败: %w", err)
		}
		return c.parsePrivateKey(keyData, c.config.KeyPassphrase)

	case AuthTypeAgent:
		// 使用SSH Agent
		return c.sshAgent()

	default:
		return nil, fmt.Errorf("不支持的认证类型: %s", c.config.AuthType)
	}
}

// parsePrivateKey 解析私钥
func (c *SSHClient) parsePrivateKey(keyData []byte, passphrase string) (ssh.AuthMethod, error) {
	var signer ssh.Signer
	var err error

	// 尝试解析PEM块
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("无法解析PEM格式的私钥")
	}

	// 检查是否加密
	if strings.Contains(block.Headers["Proc-Type"], "ENCRYPTED") || x509IsEncryptedPEMBlock(block) {
		if passphrase == "" {
			return nil, fmt.Errorf("私钥已加密，需要提供密码")
		}
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyData)
	}

	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// x509IsEncryptedPEMBlock 检查PEM块是否加密
func x509IsEncryptedPEMBlock(block *pem.Block) bool {
	_, ok := block.Headers["DEK-Info"]
	return ok
}

// sshAgent 使用SSH Agent认证
func (c *SSHClient) sshAgent() (ssh.AuthMethod, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK环境变量未设置")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("连接SSH Agent失败: %w", err)
	}

	agent := ssh.NewClientConn(conn, "unix")
	signers, err := agent.Signers()
	if err != nil {
		return nil, fmt.Errorf("获取签名器失败: %w", err)
	}

	return ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
		return signers, nil
	}), nil
}

// Close 关闭连接
func (c *SSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Execute 执行命令
func (c *SSHClient) Execute(command string) (*SSHResult, error) {
	if c.client == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	startTime := time.Now()
	result := &SSHResult{
		Timestamp: startTime,
	}

	// 创建会话
	session, err := c.client.NewSession()
	if err != nil {
		result.Error = err.Error()
		result.Success = false
		return result, fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	// 设置终端模式
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	session.RequestPty("xterm", 80, 40, modes)

	// 执行命令
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	result.Duration = time.Since(startTime).Milliseconds()
	result.Output = stdout.String()
	result.Error = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
		}
		result.Success = false
	} else {
		result.ExitCode = 0
		result.Success = true
	}

	return result, nil
}

// ExecuteWithSudo 使用sudo执行命令
func (c *SSHClient) ExecuteWithSudo(command, sudoPassword string) (*SSHResult, error) {
	if sudoPassword != "" {
		command = fmt.Sprintf("echo '%s' | sudo -S %s", sudoPassword, command)
	} else {
		command = fmt.Sprintf("sudo %s", command)
	}
	return c.Execute(command)
}

// ExecuteScript 执行脚本
func (c *SSHClient) ExecuteScript(script string) (*SSHResult, error) {
	// 使用bash执行脚本
	command := fmt.Sprintf("bash -s << 'EOF'\n%s\nEOF", script)
	return c.Execute(command)
}

// UploadFile 上传文件
func (c *SSHClient) UploadFile(localPath, remotePath string) error {
	if c.client == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// 读取本地文件
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer localFile.Close()

	fileInfo, _ := localFile.Stat()
	fileSize := fileInfo.Size()

	// 创建SFTP会话
	sftp, err := c.newSFTP()
	if err != nil {
		return err
	}
	defer sftp.Close()

	// 创建远程文件
	remoteFile, err := sftp.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件失败: %w", err)
	}
	defer remoteFile.Close()

	// 复制文件内容
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	// 设置文件权限
	sftp.Chmod(remotePath, fileInfo.Mode())

	// 设置文件大小
	if err := remoteFile.Truncate(fileSize); err != nil {
		return fmt.Errorf("设置文件大小失败: %w", err)
	}

	return nil
}

// DownloadFile 下载文件
func (c *SSHClient) DownloadFile(remotePath, localPath string) error {
	if c.client == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// 创建SFTP会话
	sftp, err := c.newSFTP()
	if err != nil {
		return err
	}
	defer sftp.Close()

	// 打开远程文件
	remoteFile, err := sftp.Open(remotePath)
	if err != nil {
		return fmt.Errorf("打开远程文件失败: %w", err)
	}
	defer remoteFile.Close()

	// 创建本地文件
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer localFile.Close()

	// 复制文件内容
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}

	return nil
}

// SFTP接口（简化版，实际需要导入github.com/pkg/sftp）
type sftpClient interface {
	Create(path string) (io.WriteCloser, error)
	Open(path string) (io.ReadCloser, error)
	Chmod(path string, mode os.FileMode) error
	Close() error
}

func (c *SSHClient) newSFTP() (sftpClient, error) {
	// 这里需要实际的SFTP实现
	// 实际使用时需要导入 github.com/pkg/sftp
	return nil, fmt.Errorf("SFTP需要安装github.com/pkg/sftp")
}

// TestConnection 测试连接
func (c *SSHClient) TestConnection() error {
	result, err := c.Execute("echo 'connection test'")
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("连接测试失败: %s", result.Error)
	}
	return nil
}

// GetSystemInfo 获取系统信息
func (c *SSHClient) GetSystemInfo() (map[string]string, error) {
	commands := map[string]string{
		"hostname": "hostname",
		"os":       "cat /etc/os-release | grep PRETTY_NAME | cut -d'\"' -f2",
		"kernel":   "uname -r",
		"arch":     "uname -m",
		"cpu_cores": "nproc",
		"memory":   "free -m | awk '/Mem:/ {print $2}'",
		"disk":     "df -h / | tail -1 | awk '{print $2}'",
		"uptime":   "uptime -p",
	}

	info := make(map[string]string)
	for key, cmd := range commands {
		result, err := c.Execute(cmd)
		if err == nil && result.Success {
			info[key] = strings.TrimSpace(result.Output)
		} else {
			info[key] = "unknown"
		}
	}

	return info, nil
}

// Tunnel 创建SSH隧道
func (c *SSHClient) Tunnel(localPort, remoteHost, remotePort string) error {
	if c.client == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// 监听本地端口
	listener, err := net.Listen("tcp", "localhost:"+localPort)
	if err != nil {
		return fmt.Errorf("监听本地端口失败: %w", err)
	}

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				continue
			}

			go func() {
				// 连接远程端口
				remoteAddr := fmt.Sprintf("%s:%s", remoteHost, remotePort)
				remoteConn, err := c.client.Dial("tcp", remoteAddr)
				if err != nil {
					localConn.Close()
					return
				}

				// 双向转发
				go func() {
					io.Copy(localConn, remoteConn)
					localConn.Close()
					remoteConn.Close()
				}()
				go func() {
					io.Copy(remoteConn, localConn)
					localConn.Close()
					remoteConn.Close()
				}()
			}()
		}
	}()

	return nil
}

// DetectAuthType 自动检测认证类型
func DetectAuthType(keyData []byte) AuthType {
	// 尝试解析为PEM
	block, _ := pem.Decode(keyData)
	if block != nil {
		return AuthTypeKey
	}

	// 检查是否像密码（简单判断）
	if len(keyData) < 100 && !strings.Contains(string(keyData), "-----BEGIN") {
		return AuthTypePassword
	}

	return AuthTypeKey
}

// ParsePEMFile 解析PEM文件
func ParsePEMFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取PEM文件失败: %w", err)
	}

	// 验证PEM格式
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式")
	}

	return data, nil
}
