package security

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SecurityLevel 安全级别
type SecurityLevel string

const (
	SecurityLevelSafe      SecurityLevel = "safe"      // 安全
	SecurityLevelWarning   SecurityLevel = "warning"   // 警告
	SecurityLevelDangerous SecurityLevel = "dangerous" // 危险
	SecurityLevelForbidden SecurityLevel = "forbidden" // 禁止
)

// ValidationResult 验证结果
type ValidationResult struct {
	Allowed          bool           `json:"allowed"`
	SecurityLevel    SecurityLevel  `json:"securityLevel"`
	Message          string         `json:"message"`
	SafeCommands     []string       `json:"safeCommands"`     // 处理后的安全命令
	ForbiddenCommands []string      `json:"forbiddenCommands"` // 被禁止的命令
	Warnings         []string       `json:"warnings"`         // 警告信息
	RequiresApproval bool           `json:"requiresApproval"` // 是否需要审批
}

// CommandWhitelist 命令白名单
type CommandWhitelist struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string `json:"name" gorm:"type:varchar(64)"`
	Pattern     string `json:"pattern" gorm:"type:varchar(255)"` // 正则表达式
	Description string `json:"description" gorm:"type:varchar(255)"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`
	RiskLevel   int    `json:"riskLevel"` // 1-5, 1最安全
}

func (CommandWhitelist) TableName() string {
	return "command_whitelist"
}

// CommandBlacklist 命令黑名单
type CommandBlacklist struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string `json:"name" gorm:"type:varchar(64)"`
	Pattern     string `json:"pattern" gorm:"type:varchar(255)"` // 正则表达式
	Reason      string `json:"reason" gorm:"type:varchar(255)"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`
}

func (CommandBlacklist) TableName() string {
	return "command_blacklist"
}

// AuditLog 审计日志
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	UserID     uint   `json:"userId" gorm:"index"`
	Username   string `json:"username" gorm:"type:varchar(64)"`
	ServerID   uint   `json:"serverId" gorm:"index"`
	ServerName string `json:"serverName" gorm:"type:varchar(64)"`

	Action     string `json:"action" gorm:"type:varchar(32)"` // execute, approve, reject, cancel
	Resource   string `json:"resource" gorm:"type:varchar(64)"`
	Command    string `json:"command" gorm:"type:text"`
	Result     string `json:"result" gorm:"type:text"` // success, failed, forbidden
	IP         string `json:"ip" gorm:"type:varchar(45)"`
	UserAgent  string `json:"userAgent" gorm:"type:varchar(255)"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// SecurityChecker 安全检查器
type SecurityChecker struct {
	whitelist []CommandWhitelist
	blacklist []CommandBlacklist
}

// NewSecurityChecker 创建安全检查器
func NewSecurityChecker() *SecurityChecker {
	return &SecurityChecker{
		whitelist: getDefaultWhitelist(),
		blacklist: getDefaultBlacklist(),
	}
}

// getDefaultWhitelist 获取默认白名单
func getDefaultWhitelist() []CommandWhitelist {
	return []CommandWhitelist{
		// 系统状态查看
		{Name: "查看进程", Pattern: `^ps\s+`, RiskLevel: 1, Enabled: true},
		{Name: "查看内存", Pattern: `^free\s+`, RiskLevel: 1, Enabled: true},
		{Name: "查看磁盘", Pattern: `^df\s+`, RiskLevel: 1, Enabled: true},
		{Name: "查看负载", Pattern: `^(uptime|w)\s*$`, RiskLevel: 1, Enabled: true},
		{Name: "查看网络", Pattern: `^(netstat|ss|ip)\s+`, RiskLevel: 1, Enabled: true},
		{Name: "查看系统信息", Pattern: `^(uname|hostname|date|whoami|id)\s*$`, RiskLevel: 1, Enabled: true},

		// 日志查看
		{Name: "查看日志", Pattern: `^(tail|head|cat|less|more)\s+.*\.log`, RiskLevel: 1, Enabled: true},
		{Name: "查看Journal", Pattern: `^journalctl\s+`, RiskLevel: 1, Enabled: true},

		// 服务管理
		{Name: "查看服务状态", Pattern: `^systemctl\s+status\s+`, RiskLevel: 1, Enabled: true},
		{Name: "重启服务", Pattern: `^systemctl\s+restart\s+\w+$`, RiskLevel: 2, Enabled: true},
		{Name: "启动服务", Pattern: `^systemctl\s+start\s+\w+$`, RiskLevel: 2, Enabled: true},
		{Name: "停止服务", Pattern: `^systemctl\s+stop\s+\w+$`, RiskLevel: 3, Enabled: true},

		// Docker 操作
		{Name: "查看容器", Pattern: `^docker\s+(ps|stats|logs|inspect)\s+`, RiskLevel: 1, Enabled: true},
		{Name: "重启容器", Pattern: `^docker\s+restart\s+\S+$`, RiskLevel: 2, Enabled: true},
		{Name: "启动容器", Pattern: `^docker\s+start\s+\S+$`, RiskLevel: 2, Enabled: true},
		{Name: "停止容器", Pattern: `^docker\s+stop\s+\S+$`, RiskLevel: 3, Enabled: true},
		{Name: "清理Docker", Pattern: `^docker\s+system\s+prune\s+`, RiskLevel: 3, Enabled: true},

		// 缓存清理
		{Name: "清理缓存", Pattern: `^echo\s+[123]\s*>\s*/proc/sys/vm/drop_caches$`, RiskLevel: 2, Enabled: true},
		{Name: "同步缓存", Pattern: `^sync\s*$`, RiskLevel: 1, Enabled: true},

		// 进程管理
		{Name: "终止进程", Pattern: `^kill\s+-?\d*\s*\d+$`, RiskLevel: 3, Enabled: true},
		{Name: "终止进程组", Pattern: `^pkill\s+`, RiskLevel: 4, Enabled: true},

		// 日志清理
		{Name: "清理日志文件", Pattern: `^find\s+/var/log\s+.*-delete$`, RiskLevel: 3, Enabled: true},
		{Name: "清理Journal", Pattern: `^journalctl\s+--vacuum-`, RiskLevel: 3, Enabled: true},
	}
}

// getDefaultBlacklist 获取默认黑名单
func getDefaultBlacklist() []CommandBlacklist {
	return []CommandBlacklist{
		// 系统破坏
		{Name: "删除根目录", Pattern: `rm\s+(-[rf]+\s+)*/*\s*$`, Reason: "禁止删除根目录", Enabled: true},
		{Name: "格式化磁盘", Pattern: `mkfs\.\w+\s+/dev/`, Reason: "禁止格式化磁盘", Enabled: true},
		{Name: "DD写入磁盘", Pattern: `dd\s+.*of=/dev/`, Reason: "禁止直接写入磁盘", Enabled: true},
		{Name: "Fork炸弹", Pattern: `:?\(\)\s*\{\s*:\|:&\s*\}\s*;:`, Reason: "禁止Fork炸弹", Enabled: true},
		{Name: "关机", Pattern: `(shutdown|poweroff|halt|init\s+0)\s`, Reason: "禁止远程关机", Enabled: true},
		{Name: "重启", Pattern: `(reboot|init\s+6)\s`, Reason: "禁止远程重启(需审批)", Enabled: true},

		// 权限修改
		{Name: "修改根权限", Pattern: `chmod\s+(-R\s+)?777\s+/$`, Reason: "禁止修改根目录权限", Enabled: true},
		{Name: "修改所有者", Pattern: `chown\s+(-R\s+)?\S+\s+/$`, Reason: "禁止修改根目录所有者", Enabled: true},

		// 用户操作
		{Name: "添加用户", Pattern: `(useradd|adduser)\s+`, Reason: "禁止添加用户", Enabled: true},
		{Name: "删除用户", Pattern: `userdel\s+`, Reason: "禁止删除用户", Enabled: true},
		{Name: "修改密码", Pattern: `passwd\s+`, Reason: "禁止修改密码", Enabled: true},
		{Name: "修改sudoers", Pattern: `visudo|/etc/sudoers`, Reason: "禁止修改sudo配置", Enabled: true},

		// 网络危险操作
		{Name: "清空防火墙", Pattern: `iptables\s+-F`, Reason: "禁止清空防火墙规则", Enabled: true},
		{Name: "关闭防火墙", Pattern: `systemctl\s+(stop|disable)\s+firewalld`, Reason: "禁止关闭防火墙", Enabled: true},

		// 数据库危险操作
		{Name: "删除数据库", Pattern: `DROP\s+(DATABASE|SCHEMA)`, Reason: "禁止删除数据库", Enabled: true},
		{Name: "删除表", Pattern: `DROP\s+TABLE`, Reason: "禁止删除表", Enabled: true},
		{Name: "清空表", Pattern: `TRUNCATE\s+TABLE?`, Reason: "禁止清空表", Enabled: true},
		{Name: "删除数据", Pattern: `DELETE\s+FROM\s+\S+\s*;?\s*$`, Reason: "禁止无条件删除数据", Enabled: true},

		// 其他危险操作
		{Name: "下载执行", Pattern: `(curl|wget).*\|\s*(bash|sh)`, Reason: "禁止下载并执行脚本", Enabled: true},
		{Name: "远程执行", Pattern: `ssh\s+.*<<\s*EOF`, Reason: "禁止SSH远程脚本执行", Enabled: true},
	}
}

// ValidateCommands 验证命令安全性
func (s *SecurityChecker) ValidateCommands(commands []string) *ValidationResult {
	result := &ValidationResult{
		SafeCommands:      []string{},
		ForbiddenCommands: []string{},
		Warnings:          []string{},
		Allowed:           true,
		RequiresApproval:  false,
	}

	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		// 检查黑名单
		if s.isBlacklisted(cmd) {
			result.ForbiddenCommands = append(result.ForbiddenCommands, cmd)
			result.Allowed = false
			result.SecurityLevel = SecurityLevelForbidden
			result.Message = "包含禁止执行的命令"
			continue
		}

		// 检查危险等级
		level, warnings := s.checkDangerLevel(cmd)
		result.Warnings = append(result.Warnings, warnings...)

		// 检查是否需要审批
		if level == SecurityLevelDangerous {
			result.RequiresApproval = true
		}

		// 设置安全级别
		if result.SecurityLevel == "" || s.compareLevel(level, result.SecurityLevel) > 0 {
			result.SecurityLevel = level
		}

		result.SafeCommands = append(result.SafeCommands, cmd)
	}

	if len(result.ForbiddenCommands) > 0 {
		result.Message = fmt.Sprintf("发现 %d 条禁止执行的命令", len(result.ForbiddenCommands))
	}

	if result.Allowed && len(result.Warnings) > 0 {
		result.Message = strings.Join(result.Warnings, "; ")
	}

	if result.Allowed && len(result.SafeCommands) > 0 {
		result.Message = "命令安全检查通过"
	}

	return result
}

// isBlacklisted 检查是否在黑名单中
func (s *SecurityChecker) isBlacklisted(cmd string) bool {
	for _, item := range s.blacklist {
		if !item.Enabled {
			continue
		}

		matched, err := regexp.MatchString(item.Pattern, cmd)
		if err == nil && matched {
			return true
		}

		// 简单字符串匹配
		if strings.Contains(cmd, item.Pattern) {
			return true
		}
	}

	// 额外检查：rm -rf /
	if strings.Contains(cmd, "rm") && strings.Contains(cmd, "-rf") {
		if strings.Contains(cmd, "/*") || strings.HasSuffix(cmd, "/") {
			return true
		}
	}

	return false
}

// checkDangerLevel 检查危险等级
func (s *SecurityChecker) checkDangerLevel(cmd string) (SecurityLevel, []string) {
	var warnings []string
	level := SecurityLevelSafe

	// 危险关键词检查
	dangerousPatterns := map[SecurityLevel][]string{
		SecurityLevelWarning: {
			"restart", "reload", "kill", "pkill",
		},
		SecurityLevelDangerous: {
			"shutdown", "reboot", "halt", "init 0", "init 6",
			"rm -rf", "mkfs", "dd if=",
			"DROP", "TRUNCATE", "DELETE FROM",
		},
	}

	for lvl, patterns := range dangerousPatterns {
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(cmd), strings.ToLower(pattern)) {
				if s.compareLevel(lvl, level) > 0 {
					level = lvl
				}
				warnings = append(warnings, fmt.Sprintf("命令包含危险操作: %s", pattern))
			}
		}
	}

	// 检查是否在白名单中
	inWhitelist := false
	for _, item := range s.whitelist {
		if !item.Enabled {
			continue
		}

		matched, err := regexp.MatchString(item.Pattern, cmd)
		if err == nil && matched {
			inWhitelist = true
			// 根据白名单的风险等级调整
			if item.RiskLevel <= 1 {
				level = SecurityLevelSafe
			} else if item.RiskLevel <= 2 {
				if s.compareLevel(level, SecurityLevelWarning) > 0 {
					level = SecurityLevelWarning
				}
			} else if item.RiskLevel <= 3 {
				if s.compareLevel(level, SecurityLevelDangerous) > 0 {
					level = SecurityLevelDangerous
				}
			}
			break
		}
	}

	if !inWhitelist && level == SecurityLevelSafe {
		// 不在白名单中，标记为警告
		level = SecurityLevelWarning
		warnings = append(warnings, "命令不在白名单中")
	}

	return level, warnings
}

// compareLevel 比较安全级别
func (s *SecurityChecker) compareLevel(a, b SecurityLevel) int {
	levels := map[SecurityLevel]int{
		SecurityLevelSafe:      1,
		SecurityLevelWarning:   2,
		SecurityLevelDangerous: 3,
		SecurityLevelForbidden: 4,
	}
	return levels[a] - levels[b]
}

// IsCommandSafe 快速检查命令是否安全
func (s *SecurityChecker) IsCommandSafe(cmd string) bool {
	result := s.ValidateCommands([]string{cmd})
	return result.Allowed && result.SecurityLevel == SecurityLevelSafe
}

// SanitizeCommand 清理命令
func (s *SecurityChecker) SanitizeCommand(cmd string) string {
	// 移除危险字符
	cmd = strings.TrimSpace(cmd)

	// 移除可能的注入
	cmd = strings.ReplaceAll(cmd, "&&", ";")
	cmd = strings.ReplaceAll(cmd, "||", ";")
	cmd = strings.ReplaceAll(cmd, "`", "'")
	cmd = strings.ReplaceAll(cmd, "$(", "(")

	return cmd
}

// AddToWhitelist 添加到白名单
func (s *SecurityChecker) AddToWhitelist(item CommandWhitelist) {
	s.whitelist = append(s.whitelist, item)
}

// AddToBlacklist 添加到黑名单
func (s *SecurityChecker) AddToBlacklist(item CommandBlacklist) {
	s.blacklist = append(s.blacklist, item)
}

// GetWhitelist 获取白名单
func (s *SecurityChecker) GetWhitelist() []CommandWhitelist {
	return s.whitelist
}

// GetBlacklist 获取黑名单
func (s *SecurityChecker) GetBlacklist() []CommandBlacklist {
	return s.blacklist
}
