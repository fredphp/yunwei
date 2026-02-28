package security

import (
        "time"

        "gorm.io/gorm"
)

// Permission 权限
type Permission struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`

        Name        string `json:"name" gorm:"type:varchar(64);uniqueIndex;not null"`
        Code        string `json:"code" gorm:"type:varchar(64);uniqueIndex;not null"` // 权限代码
        Description string `json:"description" gorm:"type:varchar(255)"`
        Group       string `json:"group" gorm:"type:varchar(32)"` // 权限分组

        // 关联
        Roles []Role `json:"roles" gorm:"many2many:role_permissions;"`
}

func (Permission) TableName() string {
        return "permissions"
}

// Role 角色
type Role struct {
        ID          uint           `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time      `json:"createdAt"`
        UpdatedAt   time.Time      `json:"updatedAt"`
        DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

        Name        string `json:"name" gorm:"type:varchar(64);uniqueIndex;not null"`
        Code        string `json:"code" gorm:"type:varchar(64);uniqueIndex;not null"` // 角色代码
        Description string `json:"description" gorm:"type:varchar(255)"`
        IsDefault   bool   `json:"isDefault" gorm:"default:false"` // 是否默认角色
        IsSystem    bool   `json:"isSystem" gorm:"default:false"`  // 是否系统角色(不可删除)

        // 权限级别
        Level       int    `json:"level" gorm:"default:1"` // 1-普通用户 2-运维 3-管理员 4-超级管理员

        // 关联
        Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
        Users       []User       `json:"users" gorm:"many2many:user_roles;"`
}

func (Role) TableName() string {
        return "roles"
}

// User 用户
type User struct {
        ID          uint           `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time      `json:"createdAt"`
        UpdatedAt   time.Time      `json:"updatedAt"`
        DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

        Username    string `json:"username" gorm:"type:varchar(64);uniqueIndex;not null"`
        Password    string `json:"-" gorm:"type:varchar(255);not null"`
        Nickname    string `json:"nickname" gorm:"type:varchar(64)"`
        Email       string `json:"email" gorm:"type:varchar(128)"`
        Phone       string `json:"phone" gorm:"type:varchar(20)"`
        Avatar      string `json:"avatar" gorm:"type:varchar(255)"`

        Status      int    `json:"status" gorm:"default:1"` // 1-正常 2-禁用
        IsAdmin     bool   `json:"isAdmin" gorm:"default:false"`

        // 关联
        Roles       []Role `json:"roles" gorm:"many2many:user_roles;"`
}

func (User) TableName() string {
        return "users"
}

// OperationPermission 操作权限定义
type OperationPermission struct {
        Code        string
        Name        string
        Description string
        Group       string
        RiskLevel   int // 1-5, 5最高风险
}

// PredefinedPermissions 预定义权限
var PredefinedPermissions = []OperationPermission{
        // ==================== 服务器管理 ====================
        {Code: "server:view", Name: "查看服务器", Group: "服务器管理", RiskLevel: 1},
        {Code: "server:add", Name: "添加服务器", Group: "服务器管理", RiskLevel: 2},
        {Code: "server:edit", Name: "编辑服务器", Group: "服务器管理", RiskLevel: 2},
        {Code: "server:delete", Name: "删除服务器", Group: "服务器管理", RiskLevel: 4},
        {Code: "server:ssh", Name: "SSH连接", Group: "服务器管理", RiskLevel: 3},
        {Code: "server:execute", Name: "执行服务器命令", Group: "服务器管理", RiskLevel: 3},
        {Code: "server:analyze", Name: "服务器AI分析", Group: "服务器管理", RiskLevel: 2},

        // 服务器分组
        {Code: "server_group:view", Name: "查看服务器分组", Group: "服务器管理", RiskLevel: 1},
        {Code: "server_group:add", Name: "添加服务器分组", Group: "服务器管理", RiskLevel: 2},
        {Code: "server_group:edit", Name: "编辑服务器分组", Group: "服务器管理", RiskLevel: 2},
        {Code: "server_group:delete", Name: "删除服务器分组", Group: "服务器管理", RiskLevel: 3},

        // ==================== Kubernetes管理 ====================
        {Code: "k8s:view", Name: "查看K8s集群", Group: "Kubernetes管理", RiskLevel: 1},
        {Code: "k8s:add", Name: "添加K8s集群", Group: "Kubernetes管理", RiskLevel: 3},
        {Code: "k8s:edit", Name: "编辑K8s集群", Group: "Kubernetes管理", RiskLevel: 3},
        {Code: "k8s:delete", Name: "删除K8s集群", Group: "Kubernetes管理", RiskLevel: 4},
        {Code: "k8s:deploy", Name: "K8s部署操作", Group: "Kubernetes管理", RiskLevel: 3},
        {Code: "k8s:scale", Name: "K8s扩缩容", Group: "Kubernetes管理", RiskLevel: 3},

        // ==================== 灰度发布 ====================
        {Code: "canary:view", Name: "查看灰度发布", Group: "灰度发布", RiskLevel: 1},
        {Code: "canary:add", Name: "创建灰度发布", Group: "灰度发布", RiskLevel: 3},
        {Code: "canary:deploy", Name: "执行灰度发布", Group: "灰度发布", RiskLevel: 4},
        {Code: "canary:rollback", Name: "灰度回滚", Group: "灰度发布", RiskLevel: 4},
        {Code: "canary:config", Name: "灰度配置", Group: "灰度发布", RiskLevel: 3},

        // ==================== 负载均衡 ====================
        {Code: "lb:view", Name: "查看负载均衡", Group: "负载均衡", RiskLevel: 1},
        {Code: "lb:add", Name: "添加负载均衡", Group: "负载均衡", RiskLevel: 2},
        {Code: "lb:edit", Name: "编辑负载均衡", Group: "负载均衡", RiskLevel: 2},
        {Code: "lb:delete", Name: "删除负载均衡", Group: "负载均衡", RiskLevel: 4},
        {Code: "lb:operate", Name: "负载均衡操作", Group: "负载均衡", RiskLevel: 3},
        {Code: "lb:optimize", Name: "负载均衡优化", Group: "负载均衡", RiskLevel: 3},

        // ==================== 证书管理 ====================
        {Code: "cert:view", Name: "查看证书", Group: "证书管理", RiskLevel: 1},
        {Code: "cert:add", Name: "添加证书", Group: "证书管理", RiskLevel: 2},
        {Code: "cert:edit", Name: "编辑证书", Group: "证书管理", RiskLevel: 2},
        {Code: "cert:delete", Name: "删除证书", Group: "证书管理", RiskLevel: 4},
        {Code: "cert:renew", Name: "续签证书", Group: "证书管理", RiskLevel: 3},
        {Code: "cert:check", Name: "检查证书", Group: "证书管理", RiskLevel: 1},

        // ==================== CDN管理 ====================
        {Code: "cdn:view", Name: "查看CDN", Group: "CDN管理", RiskLevel: 1},
        {Code: "cdn:add", Name: "添加CDN域名", Group: "CDN管理", RiskLevel: 2},
        {Code: "cdn:edit", Name: "编辑CDN域名", Group: "CDN管理", RiskLevel: 2},
        {Code: "cdn:delete", Name: "删除CDN域名", Group: "CDN管理", RiskLevel: 4},
        {Code: "cdn:operate", Name: "CDN操作(刷新/预热)", Group: "CDN管理", RiskLevel: 3},
        {Code: "cdn:optimize", Name: "CDN优化", Group: "CDN管理", RiskLevel: 3},

        // ==================== 智能部署 ====================
        {Code: "deploy:view", Name: "查看部署方案", Group: "智能部署", RiskLevel: 1},
        {Code: "deploy:add", Name: "创建部署方案", Group: "智能部署", RiskLevel: 2},
        {Code: "deploy:execute", Name: "执行部署", Group: "智能部署", RiskLevel: 4},
        {Code: "deploy:rollback", Name: "部署回滚", Group: "智能部署", RiskLevel: 4},
        {Code: "deploy:analyze", Name: "部署分析", Group: "智能部署", RiskLevel: 2},

        // ==================== 任务调度 ====================
        {Code: "scheduler:view", Name: "查看调度任务", Group: "任务调度", RiskLevel: 1},
        {Code: "scheduler:add", Name: "创建调度任务", Group: "任务调度", RiskLevel: 2},
        {Code: "scheduler:operate", Name: "调度任务操作", Group: "任务调度", RiskLevel: 3},
        {Code: "scheduler:trigger", Name: "触发任务执行", Group: "任务调度", RiskLevel: 3},

        // ==================== Agent管理 ====================
        {Code: "agent:view", Name: "查看Agent", Group: "Agent管理", RiskLevel: 1},
        {Code: "agent:edit", Name: "编辑Agent", Group: "Agent管理", RiskLevel: 2},
        {Code: "agent:delete", Name: "删除Agent", Group: "Agent管理", RiskLevel: 4},
        {Code: "agent:operate", Name: "Agent操作", Group: "Agent管理", RiskLevel: 3},
        {Code: "agent:upgrade", Name: "Agent升级", Group: "Agent管理", RiskLevel: 3},

        // ==================== 高可用管理 ====================
        {Code: "ha:view", Name: "查看高可用状态", Group: "高可用管理", RiskLevel: 1},
        {Code: "ha:operate", Name: "高可用操作", Group: "高可用管理", RiskLevel: 4},
        {Code: "ha:failover", Name: "故障转移", Group: "高可用管理", RiskLevel: 5},
        {Code: "ha:config", Name: "高可用配置", Group: "高可用管理", RiskLevel: 4},

        // ==================== 备份管理 ====================
        {Code: "backup:view", Name: "查看备份", Group: "备份管理", RiskLevel: 1},
        {Code: "backup:add", Name: "创建备份", Group: "备份管理", RiskLevel: 2},
        {Code: "backup:execute", Name: "执行备份", Group: "备份管理", RiskLevel: 3},
        {Code: "backup:restore", Name: "恢复备份", Group: "备份管理", RiskLevel: 5},
        {Code: "backup:delete", Name: "删除备份", Group: "备份管理", RiskLevel: 4},

        // ==================== 成本控制 ====================
        {Code: "cost:view", Name: "查看成本数据", Group: "成本控制", RiskLevel: 1},
        {Code: "cost:analyze", Name: "成本分析", Group: "成本控制", RiskLevel: 2},
        {Code: "cost:optimize", Name: "成本优化", Group: "成本控制", RiskLevel: 3},
        {Code: "cost:config", Name: "成本配置", Group: "成本控制", RiskLevel: 3},

        // ==================== 多租户管理 ====================
        {Code: "tenant:view", Name: "查看租户", Group: "多租户管理", RiskLevel: 1},
        {Code: "tenant:add", Name: "添加租户", Group: "多租户管理", RiskLevel: 3},
        {Code: "tenant:edit", Name: "编辑租户", Group: "多租户管理", RiskLevel: 3},
        {Code: "tenant:delete", Name: "删除租户", Group: "多租户管理", RiskLevel: 5},

        // ==================== 命令执行 ====================
        {Code: "command:execute", Name: "执行命令", Group: "命令执行", RiskLevel: 3},
        {Code: "command:approve", Name: "审批命令", Group: "命令执行", RiskLevel: 4},
        {Code: "command:whitelist", Name: "管理白名单", Group: "命令执行", RiskLevel: 4},
        {Code: "command:blacklist", Name: "管理黑名单", Group: "命令执行", RiskLevel: 4},

        // ==================== AI运维 ====================
        {Code: "ai:analyze", Name: "AI分析", Group: "AI运维", RiskLevel: 2},
        {Code: "ai:execute", Name: "执行AI建议", Group: "AI运维", RiskLevel: 4},
        {Code: "ai:approve", Name: "审批AI决策", Group: "AI运维", RiskLevel: 4},
        {Code: "ai:config", Name: "AI配置", Group: "AI运维", RiskLevel: 4},

        // ==================== 告警管理 ====================
        {Code: "alert:view", Name: "查看告警", Group: "告警管理", RiskLevel: 1},
        {Code: "alert:handle", Name: "处理告警", Group: "告警管理", RiskLevel: 3},
        {Code: "alert:config", Name: "告警配置", Group: "告警管理", RiskLevel: 3},
        {Code: "alert_rule:view", Name: "查看告警规则", Group: "告警管理", RiskLevel: 1},
        {Code: "alert_rule:edit", Name: "编辑告警规则", Group: "告警管理", RiskLevel: 3},

        // ==================== 用户管理 ====================
        {Code: "user:view", Name: "查看用户", Group: "用户管理", RiskLevel: 2},
        {Code: "user:add", Name: "添加用户", Group: "用户管理", RiskLevel: 3},
        {Code: "user:edit", Name: "编辑用户", Group: "用户管理", RiskLevel: 3},
        {Code: "user:delete", Name: "删除用户", Group: "用户管理", RiskLevel: 4},

        // ==================== 角色管理 ====================
        {Code: "role:view", Name: "查看角色", Group: "角色管理", RiskLevel: 2},
        {Code: "role:add", Name: "添加角色", Group: "角色管理", RiskLevel: 4},
        {Code: "role:edit", Name: "编辑角色", Group: "角色管理", RiskLevel: 4},
        {Code: "role:delete", Name: "删除角色", Group: "角色管理", RiskLevel: 5},

        // ==================== 审计日志 ====================
        {Code: "audit:view", Name: "查看审计日志", Group: "审计", RiskLevel: 2},
        {Code: "audit:export", Name: "导出审计日志", Group: "审计", RiskLevel: 3},

        // ==================== 系统设置 ====================
        {Code: "system:config", Name: "系统配置", Group: "系统", RiskLevel: 4},
        {Code: "system:backup", Name: "系统备份", Group: "系统", RiskLevel: 3},
        {Code: "system:restore", Name: "系统恢复", Group: "系统", RiskLevel: 5},

        // ==================== 菜单管理 ====================
        {Code: "menu:view", Name: "查看菜单", Group: "菜单管理", RiskLevel: 2},
        {Code: "menu:add", Name: "添加菜单", Group: "菜单管理", RiskLevel: 4},
        {Code: "menu:edit", Name: "编辑菜单", Group: "菜单管理", RiskLevel: 4},
        {Code: "menu:delete", Name: "删除菜单", Group: "菜单管理", RiskLevel: 5},
}

// PredefinedRole 预定义角色定义
type PredefinedRole struct {
        Code        string
        Name        string
        Level       int
        Permissions []string
        IsSystem    bool
}

// PredefinedRoles 预定义角色
var PredefinedRoles = []PredefinedRole{
        {
                Code:        "super_admin",
                Name:        "超级管理员",
                Level:       4,
                Permissions: getAllPermissionCodes(),
                IsSystem:    true,
        },
        {
                Code:        "admin",
                Name:        "管理员",
                Level:       3,
                Permissions: []string{
                        // 服务器管理
                        "server:view", "server:add", "server:edit", "server:ssh", "server:execute", "server:analyze",
                        "server_group:view", "server_group:add", "server_group:edit", "server_group:delete",
                        // K8s管理
                        "k8s:view", "k8s:add", "k8s:edit", "k8s:deploy", "k8s:scale",
                        // 灰度发布
                        "canary:view", "canary:add", "canary:deploy", "canary:config",
                        // 负载均衡
                        "lb:view", "lb:add", "lb:edit", "lb:operate", "lb:optimize",
                        // 证书管理
                        "cert:view", "cert:add", "cert:edit", "cert:renew", "cert:check",
                        // CDN管理
                        "cdn:view", "cdn:add", "cdn:edit", "cdn:operate", "cdn:optimize",
                        // 智能部署
                        "deploy:view", "deploy:add", "deploy:execute", "deploy:analyze",
                        // 任务调度
                        "scheduler:view", "scheduler:add", "scheduler:operate", "scheduler:trigger",
                        // Agent管理
                        "agent:view", "agent:edit", "agent:operate", "agent:upgrade",
                        // 高可用
                        "ha:view", "ha:operate", "ha:config",
                        // 备份管理
                        "backup:view", "backup:add", "backup:execute",
                        // 成本控制
                        "cost:view", "cost:analyze", "cost:optimize", "cost:config",
                        // 命令执行
                        "command:execute", "command:approve",
                        // AI运维
                        "ai:analyze", "ai:execute", "ai:approve",
                        // 告警管理
                        "alert:view", "alert:handle", "alert:config",
                        "alert_rule:view", "alert_rule:edit",
                        // 用户管理
                        "user:view", "user:add", "user:edit",
                        // 角色管理
                        "role:view",
                        // 审计日志
                        "audit:view", "audit:export",
                },
                IsSystem: true,
        },
        {
                Code:        "operator",
                Name:        "运维人员",
                Level:       2,
                Permissions: []string{
                        // 服务器管理 - 只读和执行
                        "server:view", "server:ssh", "server:execute", "server:analyze",
                        "server_group:view",
                        // K8s管理 - 查看和部署
                        "k8s:view", "k8s:deploy", "k8s:scale",
                        // 灰度发布 - 查看
                        "canary:view",
                        // 负载均衡 - 查看和操作
                        "lb:view", "lb:operate",
                        // 证书管理 - 查看和检查
                        "cert:view", "cert:check",
                        // CDN管理 - 查看和操作
                        "cdn:view", "cdn:operate",
                        // 智能部署 - 查看和分析
                        "deploy:view", "deploy:analyze",
                        // 任务调度 - 查看和触发
                        "scheduler:view", "scheduler:trigger",
                        // Agent管理 - 查看和操作
                        "agent:view", "agent:operate",
                        // 高可用 - 查看
                        "ha:view",
                        // 备份管理 - 查看和执行
                        "backup:view", "backup:execute",
                        // 成本控制 - 查看和分析
                        "cost:view", "cost:analyze",
                        // 命令执行
                        "command:execute",
                        // AI运维 - 分析
                        "ai:analyze",
                        // 告警管理 - 查看和处理
                        "alert:view", "alert:handle",
                        "alert_rule:view",
                        // 审计日志
                        "audit:view",
                },
                IsSystem: true,
        },
        {
                Code:        "viewer",
                Name:        "只读用户",
                Level:       1,
                Permissions: []string{
                        // 所有资源的只读权限
                        "server:view", "server_group:view",
                        "k8s:view",
                        "canary:view",
                        "lb:view",
                        "cert:view",
                        "cdn:view",
                        "deploy:view",
                        "scheduler:view",
                        "agent:view",
                        "ha:view",
                        "backup:view",
                        "cost:view",
                        "alert:view", "alert_rule:view",
                        "audit:view",
                },
                IsSystem: true,
        },
}

func getAllPermissionCodes() []string {
        codes := make([]string, len(PredefinedPermissions))
        for i, p := range PredefinedPermissions {
                codes[i] = p.Code
        }
        return codes
}
