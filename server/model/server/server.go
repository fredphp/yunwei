package server

import (
        "time"

        "gorm.io/gorm"
)

// Server 服务器
type Server struct {
        ID          uint           `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time      `json:"createdAt"`
        UpdatedAt   time.Time      `json:"updatedAt"`
        DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

        // 基本信息
        Name        string `json:"name" gorm:"type:varchar(64);not null;comment:服务器名称"`
        Hostname    string `json:"hostname" gorm:"type:varchar(64);comment:主机名"`
        Host        string `json:"host" gorm:"type:varchar(64);not null;comment:IP地址"`
        Port        int    `json:"port" gorm:"default:22;comment:SSH端口"`
        User        string `json:"user" gorm:"type:varchar(32);comment:SSH用户"`
        
        // 认证方式
        AuthType    string `json:"authType" gorm:"type:varchar(16);default:'password';comment:认证方式(password/sshKey)"`
        Password    string `json:"-" gorm:"type:varchar(255);comment:SSH密码(加密)"`
        PrivateKey  string `json:"-" gorm:"type:text;comment:SSH私钥(加密)"`
        SshKeyID    *uint  `json:"sshKeyId" gorm:"comment:SSH密钥ID"`
        SshKey      *SshKey `json:"sshKey" gorm:"foreignKey:SshKeyID"`
        
        // 分组
        GroupID     uint   `json:"groupId" gorm:"comment:分组ID"`
        Group       *Group `json:"group" gorm:"foreignKey:GroupID"`
        
        // 系统信息
        OS          string `json:"os" gorm:"type:varchar(64);comment:操作系统"`
        Arch        string `json:"arch" gorm:"type:varchar(32);comment:架构"`
        Kernel      string `json:"kernel" gorm:"type:varchar(64);comment:内核版本"`
        CPUCores    int    `json:"cpuCores" gorm:"comment:CPU核心数"`
        MemoryTotal uint64 `json:"memoryTotal" gorm:"comment:内存总量(MB)"`
        DiskTotal   uint64 `json:"diskTotal" gorm:"comment:磁盘总量(GB)"`
        
        // 状态
        Status      string `json:"status" gorm:"type:varchar(16);default:'pending';comment:状态"`
        SSHStatus   string `json:"sshStatus" gorm:"type:varchar(16);comment:SSH状态"`
        SSHError    string `json:"sshError" gorm:"type:varchar(255);comment:SSH错误"`
        
        // 实时指标
        CPUUsage    float64 `json:"cpuUsage" gorm:"comment:CPU使用率"`
        MemoryUsage float64 `json:"memoryUsage" gorm:"comment:内存使用率"`
        DiskUsage   float64 `json:"diskUsage" gorm:"comment:磁盘使用率"`
        Load1       float64 `json:"load1" gorm:"comment:1分钟负载"`
        Load5       float64 `json:"load5" gorm:"comment:5分钟负载"`
        Load15      float64 `json:"load15" gorm:"comment:15分钟负载"`
        
        // Agent
        AgentID     string `json:"agentId" gorm:"type:varchar(64);comment:Agent ID"`
        AgentOnline bool   `json:"agentOnline" gorm:"default:false;comment:Agent在线"`
        LastCheck   *time.Time `json:"lastCheck" gorm:"comment:最后检测时间"`
        LastHeartbeat *time.Time `json:"lastHeartbeat" gorm:"comment:最后心跳时间"`
        
        // 其他
        Description string `json:"description" gorm:"type:varchar(255);comment:描述"`
}

func (Server) TableName() string {
        return "servers"
}

// Group 服务器分组
type Group struct {
        ID          uint           `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time      `json:"createdAt"`
        UpdatedAt   time.Time      `json:"updatedAt"`
        DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

        Name        string `json:"name" gorm:"type:varchar(64);not null;comment:分组名称"`
        Description string `json:"description" gorm:"type:varchar(255);comment:描述"`
        ParentID    uint   `json:"parentId" gorm:"default:0;comment:父分组ID"`
}

func (Group) TableName() string {
        return "server_groups"
}

// ServerMetric 服务器指标
type ServerMetric struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt" gorm:"index"`
        ServerID  uint      `json:"serverId" gorm:"index;not null"`
        
        // CPU
        CPUUsage  float64 `json:"cpuUsage"`
        CPUUser   float64 `json:"cpuUser"`
        CPUSystem float64 `json:"cpuSystem"`
        CPUIdle   float64 `json:"cpuIdle"`
        
        // 内存
        MemoryUsage float64 `json:"memoryUsage"`
        MemoryUsed  uint64  `json:"memoryUsed"`
        MemoryFree  uint64  `json:"memoryFree"`
        MemoryCache uint64  `json:"memoryCache"`
        
        // 磁盘
        DiskUsage  float64 `json:"diskUsage"`
        DiskUsed   uint64  `json:"diskUsed"`
        DiskFree   uint64  `json:"diskFree"`
        DiskIORead  uint64 `json:"diskIORead"`
        DiskIOWrite uint64 `json:"diskIOWrite"`
        
        // 网络
        NetIn  uint64 `json:"netIn"`
        NetOut uint64 `json:"netOut"`
        
        // 负载
        Load1  float64 `json:"load1"`
        Load5  float64 `json:"load5"`
        Load15 float64 `json:"load15"`
        
        // 进程
        ProcessCount int `json:"processCount"`
}

func (ServerMetric) TableName() string {
        return "server_metrics"
}

// ServerLog 服务器日志
type ServerLog struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`
        
        ServerID  uint   `json:"serverId" gorm:"index"`
        Type      string `json:"type" gorm:"type:varchar(32);comment:类型"`
        Content   string `json:"content" gorm:"type:text;comment:内容"`
        Output    string `json:"output" gorm:"type:text;comment:输出"`
        Error     string `json:"error" gorm:"type:text;comment:错误"`
        Duration  int64  `json:"duration" gorm:"comment:耗时(ms)"`
}

func (ServerLog) TableName() string {
        return "server_logs"
}

// DockerContainer Docker 容器
type DockerContainer struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        ServerID    uint      `json:"serverId" gorm:"index"`
        
        ContainerID string `json:"containerId" gorm:"type:varchar(64)"`
        Name        string `json:"name" gorm:"type:varchar(128)"`
        Image       string `json:"image" gorm:"type:varchar(255)"`
        Status      string `json:"status" gorm:"type:varchar(32)"`
        State       string `json:"state" gorm:"type:varchar(32)"`
        
        CPUUsage    float64 `json:"cpuUsage"`
        MemoryUsage float64 `json:"memoryUsage"`
        NetIO       string  `json:"netIO"`
        BlockIO     string  `json:"blockIO"`
}

func (DockerContainer) TableName() string {
        return "docker_containers"
}

// PortInfo 端口信息
type PortInfo struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`
        ServerID  uint      `json:"serverId" gorm:"index"`
        
        Port      int    `json:"port"`
        Protocol  string `json:"protocol" gorm:"type:varchar(16)"`
        Service   string `json:"service" gorm:"type:varchar(64)"`
        PID       int    `json:"pid"`
        Process   string `json:"process" gorm:"type:varchar(128)"`
        State     string `json:"state" gorm:"type:varchar(32)"`
}

func (PortInfo) TableName() string {
        return "port_infos"
}
