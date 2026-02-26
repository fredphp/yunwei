package ha

import (
	"time"

	"gorm.io/gorm"
)

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusOnline    NodeStatus = "online"
	NodeStatusOffline   NodeStatus = "offline"
	NodeStatusStarting  NodeStatus = "starting"
	NodeStatusStopping  NodeStatus = "stopping"
	NodeStatusLeader    NodeStatus = "leader"
	NodeStatusFollower  NodeStatus = "follower"
)

// ClusterNode 集群节点
type ClusterNode struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 节点信息
	NodeID      string `json:"nodeId" gorm:"type:varchar(64);uniqueIndex;not null;comment:节点ID"`
	NodeName    string `json:"nodeName" gorm:"type:varchar(128);comment:节点名称"`
	Hostname    string `json:"hostname" gorm:"type:varchar(128);comment:主机名"`
	
	// 网络信息
	InternalIP string `json:"internalIp" gorm:"type:varchar(64);comment:内网IP"`
	ExternalIP string `json:"externalIp" gorm:"type:varchar(64);comment:外网IP"`
	APIPort    int    `json:"apiPort" gorm:"comment:API端口"`
	GRPCPort   int    `json:"grpcPort" gorm:"comment:gRPC端口"`
	
	// 状态
	Status       NodeStatus `json:"status" gorm:"type:varchar(16);default:'offline';comment:状态"`
	Role         string     `json:"role" gorm:"type:varchar(16);default:'follower';comment:角色(leader/follower)"`
	IsLeader     bool       `json:"isLeader" gorm:"default:false;comment:是否Leader"`
	
	// 心跳
	LastHeartbeat  *time.Time `json:"lastHeartbeat" gorm:"comment:最后心跳时间"`
	HeartbeatIP    string     `json:"heartbeatIp" gorm:"type:varchar(64);comment:心跳来源IP"`
	HeartbeatCount int64      `json:"heartbeatCount" gorm:"default:0;comment:心跳次数"`
	
	// 资源信息
	CPUUsage    float64 `json:"cpuUsage" gorm:"comment:CPU使用率"`
	MemoryUsage float64 `json:"memoryUsage" gorm:"comment:内存使用率"`
	DiskUsage   float64 `json:"diskUsage" gorm:"comment:磁盘使用率"`
	
	// 负载信息
	GoroutineCount int   `json:"goroutineCount" gorm:"comment:协程数"`
	RequestCount   int64 `json:"requestCount" gorm:"default:0;comment:请求数"`
	ConnectionCount int  `json:"connectionCount" gorm:"comment:连接数"`
	
	// 版本信息
	Version   string `json:"version" gorm:"type:varchar(32);comment:版本"`
	GoVersion string `json:"goVersion" gorm:"type:varchar(32);comment:Go版本"`
	
	// 配置
	Weight      int    `json:"weight" gorm:"default:100;comment:权重"`
	DataCenter  string `json:"dataCenter" gorm:"type:varchar(64);comment:数据中心"`
	Zone        string `json:"zone" gorm:"type:varchar(64);comment:可用区"`
	Rack        string `json:"rack" gorm:"type:varchar(64);comment:机架"`
	
	// 标签
	Labels string `json:"labels" gorm:"type:text;comment:标签(JSON)"`
	
	// 启用状态
	Enabled bool `json:"enabled" gorm:"default:true;comment:是否启用"`
}

func (ClusterNode) TableName() string {
	return "cluster_nodes"
}

// DistributedLock 分布式锁记录
type DistributedLock struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 锁信息
	LockKey   string `json:"lockKey" gorm:"type:varchar(255);uniqueIndex;not null;comment:锁键"`
	LockValue string `json:"lockValue" gorm:"type:varchar(128);comment:锁值"`
	
	// 持有者
	HolderNodeID string `json:"holderNodeId" gorm:"type:varchar(64);index;comment:持有者节点ID"`
	HolderIP     string `json:"holderIp" gorm:"type:varchar(64);comment:持有者IP"`
	
	// 时间
	AcquiredAt   *time.Time `json:"acquiredAt" gorm:"comment:获取时间"`
	ExpiresAt    *time.Time `json:"expiresAt" gorm:"comment:过期时间"`
	ReleasedAt   *time.Time `json:"releasedAt" gorm:"comment:释放时间"`
	
	// 状态
	Status       string `json:"status" gorm:"type:varchar(16);comment:状态(acquired/released/expired)"`
	RenewCount   int    `json:"renewCount" gorm:"default:0;comment:续期次数"`
	WaitCount    int64  `json:"waitCount" gorm:"default:0;comment:等待次数"`
	
	// 配置
	TTLSeconds int `json:"ttlSeconds" gorm:"comment:TTL(秒)"`
	
	// 来源
	ResourceType string `json:"resourceType" gorm:"type:varchar(64);comment:资源类型"`
	ResourceID   string `json:"resourceId" gorm:"type:varchar(128);comment:资源ID"`
}

func (DistributedLock) TableName() string {
	return "distributed_locks"
}

// LeaderElection Leader选举记录
type LeaderElection struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 选举信息
	ElectionKey string `json:"electionKey" gorm:"type:varchar(128);uniqueIndex;not null;comment:选举键"`
	
	// Leader信息
	LeaderNodeID string `json:"leaderNodeId" gorm:"type:varchar(64);index;comment:Leader节点ID"`
	LeaderIP     string `json:"leaderIp" gorm:"type:varchar(64);comment:Leader IP"`
	
	// 任期
	Term        int64      `json:"term" gorm:"comment:任期号"`
	AcquiredAt  *time.Time `json:"acquiredAt" gorm:"comment:获取时间"`
	ExpiresAt   *time.Time `json:"expiresAt" gorm:"comment:过期时间"`
	
	// 统计
	RenewCount   int   `json:"renewCount" gorm:"default:0;comment:续期次数"`
	LeaderCount  int64 `json:"leaderCount" gorm:"default:0;comment:担任Leader次数"`
	
	// 候选人
	Candidates string `json:"candidates" gorm:"type:text;comment:候选人列表(JSON)"`
	
	// 状态
	Status string `json:"status" gorm:"type:varchar(16);comment:状态"`
}

func (LeaderElection) TableName() string {
	return "leader_elections"
}

// HAClusterConfig HA集群配置
type HAClusterConfig struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 配置名称
	Name        string `json:"name" gorm:"type:varchar(64);not null;comment:配置名称"`
	Description string `json:"description" gorm:"type:varchar(255);comment:描述"`
	
	// 集群配置
	ClusterName     string `json:"clusterName" gorm:"type:varchar(64);comment:集群名称"`
	ClusterMode     string `json:"clusterMode" gorm:"type:varchar(32);default:'active-active';comment:集群模式(active-active/active-passive)"`
	
	// 节点配置
	MinNodes       int  `json:"minNodes" gorm:"default:1;comment:最小节点数"`
	MaxNodes       int  `json:"maxNodes" gorm:"default:10;comment:最大节点数"`
	AutoDiscovery  bool `json:"autoDiscovery" gorm:"default:true;comment:自动发现节点"`
	
	// 心跳配置
	HeartbeatInterval   int `json:"heartbeatInterval" gorm:"default:10;comment:心跳间隔(秒)"`
	HeartbeatTimeout    int `json:"heartbeatTimeout" gorm:"default:30;comment:心跳超时(秒)"`
	
	// 选举配置
	ElectionTimeout     int `json:"electionTimeout" gorm:"default:30;comment:选举超时(秒)"`
	LeaderLeaseSeconds  int `json:"leaderLeaseSeconds" gorm:"default:15;comment:Leader租约(秒)"`
	
	// 故障转移
	FailoverEnabled     bool `json:"failoverEnabled" gorm:"default:true;comment:启用故障转移"`
	FailoverTimeout     int  `json:"failoverTimeout" gorm:"default:60;comment:故障转移超时(秒)"`
	AutoFailback        bool `json:"autoFailback" gorm:"default:true;comment:自动回切"`
	
	// 负载均衡
	LoadBalanceEnabled  bool   `json:"loadBalanceEnabled" gorm:"default:true;comment:启用负载均衡"`
	LoadBalanceStrategy string `json:"loadBalanceStrategy" gorm:"type:varchar(32);default:'round-robin';comment:负载均衡策略"`
	
	// 分布式锁
	LockBackend    string `json:"lockBackend" gorm:"type:varchar(32);default:'redis';comment:锁后端(redis/etcd/database)"`
	LockTTLSeconds int    `json:"lockTtlSeconds" gorm:"default:30;comment:锁TTL(秒)"`
	
	// Redis配置
	RedisMode       string `json:"redisMode" gorm:"type:varchar(32);default:'standalone';comment:Redis模式(standalone/sentinel/cluster)"`
	RedisMasterName string `json:"redisMasterName" gorm:"type:varchar(64);comment:Redis主节点名称(哨兵模式)"`
	
	// 数据库配置
	DBMode          string `json:"dbMode" gorm:"type:varchar(32);default:'standalone';comment:数据库模式(standalone/master-slave)"`
	DBReadFromSlave bool   `json:"dbReadFromSlave" gorm:"default:false;comment:从库读取"`
	
	// 会话配置
	SessionMode    string `json:"sessionMode" gorm:"type:varchar(32);default:'memory';comment:会话模式(memory/redis)"`
	SessionTTL     int    `json:"sessionTtl" gorm:"default:1800;comment:会话TTL(秒)"`
	
	// 状态
	Enabled bool `json:"enabled" gorm:"default:true;comment:是否启用"`
}

func (HAClusterConfig) TableName() string {
	return "ha_cluster_configs"
}

// FailoverRecord 故障转移记录
type FailoverRecord struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	// 故障信息
	FailoverType string `json:"failoverType" gorm:"type:varchar(32);comment:类型(node/leader/db/redis)"`
	
	// 故障节点
	FailedNodeID   string `json:"failedNodeId" gorm:"type:varchar(64);index;comment:故障节点ID"`
	FailedNodeName string `json:"failedNodeName" gorm:"type:varchar(128);comment:故障节点名称"`
	FailedNodeIP   string `json:"failedNodeIp" gorm:"type:varchar(64);comment:故障节点IP"`
	
	// 原因
	Reason       string `json:"reason" gorm:"type:varchar(255);comment:故障原因"`
	DetectedAt   *time.Time `json:"detectedAt" gorm:"comment:检测时间"`
	
	// 转移目标
	TargetNodeID   string `json:"targetNodeId" gorm:"type:varchar(64);comment:目标节点ID"`
	TargetNodeName string `json:"targetNodeName" gorm:"type:varchar(128);comment:目标节点名称"`
	TargetNodeIP   string `json:"targetNodeIp" gorm:"type:varchar(64);comment:目标节点IP"`
	
	// 执行信息
	Status       string     `json:"status" gorm:"type:varchar(16);comment:状态"`
	StartedAt    *time.Time `json:"startedAt" gorm:"comment:开始时间"`
	CompletedAt  *time.Time `json:"completedAt" gorm:"comment:完成时间"`
	Duration     int64      `json:"duration" gorm:"comment:耗时(毫秒)"`
	
	// 结果
	Success bool   `json:"success" gorm:"comment:是否成功"`
	Error   string `json:"error" gorm:"type:text;comment:错误信息"`
	
	// 自动/手动
	TriggerType string `json:"triggerType" gorm:"type:varchar(16);comment:触发类型(auto/manual)"`
	TriggeredBy string `json:"triggeredBy" gorm:"type:varchar(64);comment:触发者"`
}

func (FailoverRecord) TableName() string {
	return "failover_records"
}

// HASession HA会话
type HASession struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 会话信息
	SessionID   string `json:"sessionId" gorm:"type:varchar(128);uniqueIndex;not null;comment:会话ID"`
	UserID      uint   `json:"userId" gorm:"index;comment:用户ID"`
	Username    string `json:"username" gorm:"type:varchar(64);comment:用户名"`
	
	// 会话数据
	Data string `json:"data" gorm:"type:text;comment:会话数据(JSON)"`
	
	// 节点信息
	CreatedNodeID string `json:"createdNodeId" gorm:"type:varchar(64);comment:创建节点ID"`
	LastAccessNodeID string `json:"lastAccessNodeId" gorm:"type:varchar(64);comment:最后访问节点ID"`
	
	// 时间
	LastAccessAt *time.Time `json:"lastAccessAt" gorm:"comment:最后访问时间"`
	ExpiresAt    *time.Time `json:"expiresAt" gorm:"comment:过期时间"`
	
	// 状态
	IsActive bool `json:"isActive" gorm:"default:true;comment:是否活跃"`
	
	// 客户端信息
	ClientIP    string `json:"clientIp" gorm:"type:varchar(64);comment:客户端IP"`
	UserAgent   string `json:"userAgent" gorm:"type:varchar(255);comment:用户代理"`
}

func (HASession) TableName() string {
	return "ha_sessions"
}

// ClusterEvent 集群事件
type ClusterEvent struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt" gorm:"index"`

	// 事件信息
	EventType string `json:"eventType" gorm:"type:varchar(32);index;comment:事件类型"`
	
	// 节点信息
	NodeID   string `json:"nodeId" gorm:"type:varchar(64);index;comment:节点ID"`
	NodeName string `json:"nodeName" gorm:"type:varchar(128);comment:节点名称"`
	NodeIP   string `json:"nodeIp" gorm:"type:varchar(64);comment:节点IP"`
	
	// 事件内容
	Title   string `json:"title" gorm:"type:varchar(255);comment:标题"`
	Detail  string `json:"detail" gorm:"type:text;comment:详情"`
	
	// 级别
	Level string `json:"level" gorm:"type:varchar(16);comment:级别(info/warning/error/critical)"`
	
	// 来源
	Source string `json:"source" gorm:"type:varchar(64);comment:来源"`
}

func (ClusterEvent) TableName() string {
	return "cluster_events"
}

// NodeMetric 节点指标
type NodeMetric struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt" gorm:"index"`

	NodeID uint   `json:"nodeId" gorm:"index;comment:节点ID"`
	NodeUUID string `json:"nodeUuid" gorm:"type:varchar(64);index;comment:节点UUID"`
	
	// 资源指标
	CPUUsage       float64 `json:"cpuUsage" gorm:"comment:CPU使用率"`
	MemoryUsage    float64 `json:"memoryUsage" gorm:"comment:内存使用率"`
	MemoryUsed     uint64  `json:"memoryUsed" gorm:"comment:已用内存(MB)"`
	MemoryTotal    uint64  `json:"memoryTotal" gorm:"comment:总内存(MB)"`
	DiskUsage      float64 `json:"diskUsage" gorm:"comment:磁盘使用率"`
	DiskUsed       uint64  `json:"diskUsed" gorm:"comment:已用磁盘(GB)"`
	DiskTotal      uint64  `json:"diskTotal" gorm:"comment:总磁盘(GB)"`
	
	// 网络指标
	NetInBytes     uint64 `json:"netInBytes" gorm:"comment:网络入流量"`
	NetOutBytes    uint64 `json:"netOutBytes" gorm:"comment:网络出流量"`
	
	// 运行时指标
	GoroutineCount int   `json:"goroutineCount" gorm:"comment:协程数"`
	ThreadCount    int   `json:"threadCount" gorm:"comment:线程数"`
	HandleCount    int   `json:"handleCount" gorm:"comment:句柄数"`
	
	// 业务指标
	RequestCount     int64   `json:"requestCount" gorm:"comment:请求数"`
	RequestLatency   float64 `json:"requestLatency" gorm:"comment:请求延迟(ms)"`
	RequestQPS       float64 `json:"requestQps" gorm:"comment:QPS"`
	ConnectionCount  int     `json:"connectionCount" gorm:"comment:连接数"`
	
	// 负载
	Load1  float64 `json:"load1" gorm:"comment:1分钟负载"`
	Load5  float64 `json:"load5" gorm:"comment:5分钟负载"`
	Load15 float64 `json:"load15" gorm:"comment:15分钟负载"`
}

func (NodeMetric) TableName() string {
	return "node_metrics"
}
