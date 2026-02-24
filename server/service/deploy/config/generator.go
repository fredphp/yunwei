package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"yunwei/server/service/deploy/planner"
)

// ConfigGenerator 配置生成器
type ConfigGenerator struct{}

// NewConfigGenerator 创建配置生成器
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// GeneratedConfig 生成的配置
type GeneratedConfig struct {
	ServerID   uint            `json:"serverId"`
	ServerName string          `json:"serverName"`
	Configs    []ConfigFile    `json:"configs"`
	Commands   []ExecCommand   `json:"commands"`
}

// ConfigFile 配置文件
type ConfigFile struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Mode     string `json:"mode"` // 0644, 0755, etc.
	Owner    string `json:"owner"`
	Group    string `json:"group"`
	Reload   bool   `json:"reload"` // 是否需要重载服务
}

// ExecCommand 执行命令
type ExecCommand struct {
	Command   string `json:"command"`
	Order     int    `json:"order"`
	Check     string `json:"check"`     // 检查命令是否成功的命令
	IgnoreErr bool   `json:"ignoreErr"` // 是否忽略错误
}

// GenerateAllConfigs 生成所有配置
func (g *ConfigGenerator) GenerateAllConfigs(plan *planner.DeployPlan) ([]GeneratedConfig, error) {
	var configs []GeneratedConfig
	
	// 解析服务器分配
	var assignments []planner.ServerAssignment
	json.Unmarshal([]byte(plan.ServerAssignments), &assignments)
	
	// 解析服务配置
	var services []planner.ServiceConfig
	json.Unmarshal([]byte(plan.Services), &services)
	
	// 解析网络配置
	var networkConfig map[string]interface{}
	json.Unmarshal([]byte(plan.NetworkConfig), &networkConfig)
	
	// 为每个服务器生成配置
	for _, assignment := range assignments {
		config := GeneratedConfig{
			ServerID:   assignment.ServerID,
			ServerName: assignment.ServerName,
		}
		
		// 根据角色生成配置
		switch {
		case assignment.Role == "lb":
			config.Configs = g.generateLBConfig(plan, assignment, services)
			
		case strings.HasPrefix(assignment.Role, "db"):
			config.Configs = g.generateDBConfig(plan, assignment)
			
		case assignment.Role == "cache":
			config.Configs = g.generateCacheConfig(plan, assignment)
			
		case assignment.Role == "mq":
			config.Configs = g.generateMQConfig(plan, assignment)
			
		default:
			config.Configs = g.generateAppConfig(plan, assignment, services)
		}
		
		// 生成主机名和 hosts 配置
		config.Configs = append(config.Configs, g.generateHostsConfig(assignments)...)
		
		// 生成防火墙配置
		config.Configs = append(config.Configs, g.generateFirewallConfig(networkConfig)...)
		
		// 生成环境变量文件
		config.Configs = append(config.Configs, g.generateEnvConfig(plan, assignment)...)
		
		configs = append(configs, config)
	}
	
	return configs, nil
}

// generateLBConfig 生成负载均衡配置
func (g *ConfigGenerator) generateLBConfig(plan *planner.DeployPlan, assignment planner.ServerAssignment, services []planner.ServiceConfig) []ConfigFile {
	var configs []ConfigFile
	
	// 解析 LB 配置
	var lbConfig map[string]interface{}
	json.Unmarshal([]byte(plan.LoadBalancer), &lbConfig)
	
	// 生成 Nginx 配置
	nginxConf := g.generateNginxConfig(lbConfig, services)
	configs = append(configs, ConfigFile{
		Path:    "/etc/nginx/nginx.conf",
		Content: nginxConf,
		Mode:    "0644",
		Owner:   "root",
		Group:   "root",
		Reload:  true,
	})
	
	// 生成 upstream 配置
	upstreamConf := g.generateUpstreamConfig(lbConfig, services)
	configs = append(configs, ConfigFile{
		Path:    "/etc/nginx/conf.d/upstream.conf",
		Content: upstreamConf,
		Mode:    "0644",
		Owner:   "root",
		Group:   "root",
		Reload:  true,
	})
	
	// 生成 SSL 配置（如果启用）
	if ssl, ok := lbConfig["ssl"].(map[string]interface{}); ok {
		if enabled, ok := ssl["enabled"].(bool); ok && enabled {
			sslConf := g.generateSSLConfig(ssl)
			configs = append(configs, ConfigFile{
				Path:    "/etc/nginx/conf.d/ssl.conf",
				Content: sslConf,
				Mode:    "0644",
				Owner:   "root",
				Group:   "root",
				Reload:  true,
			})
		}
	}
	
	// 生成 keepalived 配置（高可用）
	if len(assignment.Services) > 1 {
		keepalivedConf := g.generateKeepalivedConfig(assignment)
		configs = append(configs, ConfigFile{
			Path:    "/etc/keepalived/keepalived.conf",
			Content: keepalivedConf,
			Mode:    "0644",
			Owner:   "root",
			Group:   "root",
			Reload:  true,
		})
	}
	
	return configs
}

// generateNginxConfig 生成 Nginx 主配置
func (g *ConfigGenerator) generateNginxConfig(lbConfig map[string]interface{}, services []planner.ServiceConfig) string {
	tmpl := `user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 10240;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';
    
    access_log /var/log/nginx/access.log main;
    
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    
    # Gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml application/json application/javascript application/xml;
    
    # 请求体大小限制
    client_max_body_size 50m;
    
    # 代理缓冲
    proxy_buffer_size 128k;
    proxy_buffers 4 256k;
    proxy_busy_buffers_size 256k;
    
    # 包含其他配置
    include /etc/nginx/conf.d/*.conf;
    
    # 负载均衡状态页
    server {
        listen 8080;
        server_name localhost;
        
        location /nginx_status {
            stub_status on;
            access_log off;
            allow 127.0.0.1;
            allow 10.0.0.0/8;
            deny all;
        }
        
        location /health {
            return 200 'OK';
            add_header Content-Type text/plain;
        }
    }
}
`
	return tmpl
}

// generateUpstreamConfig 生成 Upstream 配置
func (g *ConfigGenerator) generateUpstreamConfig(lbConfig map[string]interface{}, services []planner.ServiceConfig) string {
	var sb strings.Builder
	
	// 获取后端服务器
	backends, _ := lbConfig["backends"].([]interface{})
	strategy, _ := lbConfig["strategy"].(string)
	
	// 为每个服务创建 upstream
	for _, svc := range services {
		sb.WriteString(fmt.Sprintf("upstream %s {\n", svc.Name))
		sb.WriteString(fmt.Sprintf("    %s;\n", g.getLBStrategy(strategy)))
		
		// 添加后端服务器
		for _, backend := range backends {
			if b, ok := backend.(map[string]interface{}); ok {
				server, _ := b["server"].(string)
				port, _ := b["port"].(float64)
				weight, _ := b["weight"].(float64)
				sb.WriteString(fmt.Sprintf("    server %s:%d weight=%d max_fails=3 fail_timeout=30s;\n", 
					server, int(port), int(weight)))
			}
		}
		
		sb.WriteString("}\n\n")
	}
	
	// 创建主服务器配置
	sb.WriteString(`server {
    listen 80;
    server_name _;
    
    # HTTP 重定向到 HTTPS（如果启用）
    # return 301 https://$server_name$request_uri;
    
`)
	
	// 为每个服务创建 location
	for _, svc := range services {
		sb.WriteString(fmt.Sprintf(`    location /%s {
        proxy_pass http://%s;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

`, svc.Name, svc.Name))
	}
	
	sb.WriteString("}\n")
	
	return sb.String()
}

// getLBStrategy 获取负载均衡策略
func (g *ConfigGenerator) getLBStrategy(strategy string) string {
	switch strategy {
	case "least_conn":
		return "least_conn"
	case "ip_hash":
		return "ip_hash"
	case "random":
		return "random"
	default:
		return "least_conn"
	}
}

// generateSSLConfig 生成 SSL 配置
func (g *ConfigGenerator) generateSSLConfig(ssl map[string]interface{}) string {
	return `server {
    listen 443 ssl http2;
    server_name _;
    
    ssl_certificate /etc/nginx/ssl/server.crt;
    ssl_certificate_key /etc/nginx/ssl/server.key;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;
    
    # HSTS
    add_header Strict-Transport-Security "max-age=31536000" always;
    
    # 包含其他 location 配置
    include /etc/nginx/conf.d/upstream.conf;
}
`
}

// generateKeepalivedConfig 生成 Keepalived 配置
func (g *ConfigGenerator) generateKeepalivedConfig(assignment planner.ServerAssignment) string {
	state := "BACKUP"
	priority := 100
	
	// 如果是主节点
	if strings.Contains(assignment.Role, "master") {
		state = "MASTER"
		priority = 101
	}
	
	return fmt.Sprintf(`global_defs {
    router_id %s
    vrrp_skip_check_adv_addr
    vrrp_strict
    vrrp_garp_interval 0
    vrrp_gna_interval 0
}

vrrp_instance VI_1 {
    state %s
    interface eth0
    virtual_router_id 51
    priority %d
    advert_int 1
    
    authentication {
        auth_type PASS
        auth_pass 1111
    }
    
    virtual_ipaddress {
        10.0.0.100/24
    }
    
    track_script {
        check_nginx
    }
}

vrrp_script check_nginx {
    script "/usr/bin/killall -0 nginx"
    interval 2
    weight -5
    fall 3
    rise 2
}
`, assignment.ServerName, state, priority)
}

// generateDBConfig 生成数据库配置
func (g *ConfigGenerator) generateDBConfig(plan *planner.DeployPlan, assignment planner.ServerAssignment) []ConfigFile {
	var configs []ConfigFile
	
	// 解析数据库配置
	var dbConfig map[string]interface{}
	json.Unmarshal([]byte(plan.DatabaseConfig), &dbConfig)
	
	// 判断是主库还是从库
	isMaster := assignment.Role == "db-master"
	
	// 生成 MySQL 配置
	mysqlConf := g.generateMySQLConfig(dbConfig, isMaster, assignment)
	configs = append(configs, ConfigFile{
		Path:    "/etc/mysql/mysql.conf.d/mysqld.cnf",
		Content: mysqlConf,
		Mode:    "0644",
		Owner:   "mysql",
		Group:   "mysql",
		Reload:  true,
	})
	
	// 如果是从库，生成复制配置
	if !isMaster {
		replConf := g.generateReplicationConfig(dbConfig, assignment)
		configs = append(configs, ConfigFile{
			Path:    "/etc/mysql/mysql.conf.d/replication.cnf",
			Content: replConf,
			Mode:    "0644",
			Owner:   "mysql",
			Group:   "mysql",
			Reload:  true,
		})
	}
	
	// 生成备份脚本
	backupScript := g.generateBackupScript(dbConfig)
	configs = append(configs, ConfigFile{
		Path:    "/usr/local/bin/mysql-backup.sh",
		Content: backupScript,
		Mode:    "0755",
		Owner:   "root",
		Group:   "root",
		Reload:  false,
	})
	
	return configs
}

// generateMySQLConfig 生成 MySQL 配置
func (g *ConfigGenerator) generateMySQLConfig(dbConfig map[string]interface{}, isMaster bool, assignment planner.ServerAssignment) string {
	serverID := assignment.ServerID
	
	tmpl := `[mysqld]
# 基本设置
user = mysql
port = 3306
datadir = /var/lib/mysql
socket = /var/run/mysqld/mysqld.sock
pid-file = /var/run/mysqld/mysqld.pid

# 字符集
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci

# 连接设置
max_connections = 500
max_connect_errors = 100
wait_timeout = 28800
interactive_timeout = 28800

# 缓冲设置
innodb_buffer_pool_size = 2G
innodb_buffer_pool_instances = 4
innodb_log_buffer_size = 64M

# 日志设置
log_error = /var/log/mysql/error.log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2

# 二进制日志（主从复制）
log_bin = mysql-bin
binlog_format = ROW
expire_logs_days = 7
max_binlog_size = 100M

# 服务器 ID（主从复制）
server-id = {{.ServerID}}

# 复制设置
{{if .IsMaster}}
log_slave_updates = ON
binlog_do_db = app_db
{{else}}
relay_log = mysql-relay-bin
relay_log_index = mysql-relay-bin.index
read_only = ON
{{end}}

# 性能优化
innodb_flush_log_at_trx_commit = 1
innodb_lock_wait_timeout = 50
innodb_flush_method = O_DIRECT

[client]
port = 3306
socket = /var/run/mysqld/mysqld.sock
default-character-set = utf8mb4
`
	
	t, _ := template.New("mysql").Parse(tmpl)
	var result strings.Builder
	t.Execute(&result, struct {
		ServerID  uint
		IsMaster  bool
	}{ServerID: serverID, IsMaster: isMaster})
	
	return result.String()
}

// generateReplicationConfig 生成复制配置
func (g *ConfigGenerator) generateReplicationConfig(dbConfig map[string]interface{}, assignment planner.ServerAssignment) string {
	master, _ := dbConfig["master"].(map[string]interface{})
	masterServer, _ := master["server"].(string)
	
	return fmt.Sprintf(`# 从库复制配置
# 在 MySQL 中执行以下命令设置复制:

CHANGE MASTER TO
    MASTER_HOST='%s',
    MASTER_PORT=3306,
    MASTER_USER='repl',
    MASTER_PASSWORD='repl_password',
    MASTER_LOG_FILE='mysql-bin.000001',
    MASTER_LOG_POS=4;

START SLAVE;

# 检查复制状态
SHOW SLAVE STATUS\G
`, masterServer)
}

// generateBackupScript 生成备份脚本
func (g *ConfigGenerator) generateBackupScript(dbConfig map[string]interface{}) string {
	return `#!/bin/bash
# MySQL 自动备份脚本

BACKUP_DIR="/backup/mysql"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="app_db"
DB_USER="backup"
DB_PASS="backup_password"
RETAIN_DAYS=7

# 创建备份目录
mkdir -p $BACKUP_DIR

# 执行备份
mysqldump -u$DB_USER -p$DB_PASS --single-transaction --routines --triggers $DB_NAME | gzip > $BACKUP_DIR/${DB_NAME}_${DATE}.sql.gz

# 清理旧备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +$RETAIN_DAYS -delete

echo "Backup completed: ${DB_NAME}_${DATE}.sql.gz"
`
}

// generateCacheConfig 生成缓存配置
func (g *ConfigGenerator) generateCacheConfig(plan *planner.DeployPlan, assignment planner.ServerAssignment) []ConfigFile {
	var configs []ConfigFile
	
	// 解析缓存配置
	var cacheConfig map[string]interface{}
	json.Unmarshal([]byte(plan.CacheConfig), &cacheConfig)
	
	// 生成 Redis 配置
	redisConf := g.generateRedisConfig(cacheConfig, assignment)
	configs = append(configs, ConfigFile{
		Path:    "/etc/redis/redis.conf",
		Content: redisConf,
		Mode:    "0644",
		Owner:   "redis",
		Group:   "redis",
		Reload:  true,
	})
	
	// 如果是集群模式，生成集群配置
	mode, _ := cacheConfig["mode"].(string)
	if mode == "cluster" {
		clusterConf := g.generateRedisClusterConfig(cacheConfig)
		configs = append(configs, ConfigFile{
			Path:    "/etc/redis/cluster.conf",
			Content: clusterConf,
			Mode:    "0644",
			Owner:   "redis",
			Group:   "redis",
			Reload:  true,
		})
	}
	
	return configs
}

// generateRedisConfig 生成 Redis 配置
func (g *ConfigGenerator) generateRedisConfig(cacheConfig map[string]interface{}, assignment planner.ServerAssignment) string {
	memory, _ := cacheConfig["memory"].(string)
	if memory == "" {
		memory = "4gb"
	}
	
	return fmt.Sprintf(`# Redis 配置文件

# 网络
bind 0.0.0.0
port 6379
protected-mode yes

# 内存
maxmemory %s
maxmemory-policy allkeys-lru

# 持久化
save 900 1
save 300 10
save 60 10000

appendonly yes
appendfsync everysec
appendfilename "appendonly.aof"

# 日志
loglevel notice
logfile /var/log/redis/redis.log

# 性能
timeout 0
tcp-keepalive 300
tcp-backlog 511

# 安全
# requirepass your_password_here

# 慢查询日志
slowlog-log-slower-than 10000
slowlog-max-len 128

# 客户端
maxclients 10000
`, memory)
}

// generateRedisClusterConfig 生成 Redis 集群配置
func (g *ConfigGenerator) generateRedisClusterConfig(cacheConfig map[string]interface{}) string {
	nodes, _ := cacheConfig["nodes"].([]interface{})
	
	var nodeList []string
	for _, node := range nodes {
		if n, ok := node.(map[string]interface{}); ok {
			server, _ := n["server"].(string)
			port, _ := n["port"].(float64)
			nodeList = append(nodeList, fmt.Sprintf("%s:%d", server, int(port)))
		}
	}
	
	return fmt.Sprintf(`# Redis 集群配置

cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
cluster-announce-ip 0.0.0.0
cluster-announce-port 6379
cluster-announce-bus-port 16379

# 集群节点列表
# %s

# 集群要求全覆盖
cluster-require-full-coverage no
`, strings.Join(nodeList, ", "))
}

// generateMQConfig 生成消息队列配置
func (g *ConfigGenerator) generateMQConfig(plan *planner.DeployPlan, assignment planner.ServerAssignment) []ConfigFile {
	var configs []ConfigFile
	
	// 解析 MQ 配置
	var mqConfig map[string]interface{}
	json.Unmarshal([]byte(plan.MQConfig), &mqConfig)
	
	// 生成 RabbitMQ 配置
	rabbitmqConf := g.generateRabbitMQConfig(mqConfig)
	configs = append(configs, ConfigFile{
		Path:    "/etc/rabbitmq/rabbitmq.conf",
		Content: rabbitmqConf,
		Mode:    "0644",
		Owner:   "rabbitmq",
		Group:   "rabbitmq",
		Reload:  true,
	})
	
	// 如果是集群模式，生成集群配置
	isCluster, _ := mqConfig["cluster"].(bool)
	if isCluster {
		clusterConf := g.generateRabbitMQClusterConfig(mqConfig)
		configs = append(configs, ConfigFile{
			Path:    "/etc/rabbitmq/rabbitmq-cluster.conf",
			Content: clusterConf,
			Mode:    "0644",
			Owner:   "rabbitmq",
			Group:   "rabbitmq",
			Reload:  true,
		})
	}
	
	return configs
}

// generateRabbitMQConfig 生成 RabbitMQ 配置
func (g *ConfigGenerator) generateRabbitMQConfig(mqConfig map[string]interface{}) string {
	return `# RabbitMQ 配置

# 监听端口
listeners.tcp.default = 5672

# 管理界面
management.tcp.port = 15672

# 内存限制
total_memory_available_override_value = 2GB
vm_memory_high_watermark.relative = 0.6

# 磁盘限制
disk_free_limit.absolute = 1GB

# 日志
log.console.level = info
log.file.level = info
log.file = /var/log/rabbitmq/rabbit.log

# 心跳
heartbeat = 60

# 消息持久化
default_user = admin
default_pass = admin123

# 队列默认配置
queue.default_auto_delete = false
queue.default_durable = true
`
}

// generateRabbitMQClusterConfig 生成 RabbitMQ 集群配置
func (g *ConfigGenerator) generateRabbitMQClusterConfig(mqConfig map[string]interface{}) string {
	nodes, _ := mqConfig["nodes"].([]interface{})
	
	var nodeList []string
	for _, node := range nodes {
		if n, ok := node.(map[string]interface{}); ok {
			server, _ := n["server"].(string)
			nodeList = append(nodeList, fmt.Sprintf("rabbit@%s", server))
		}
	}
	
	return fmt.Sprintf(`# RabbitMQ 集群配置

# 集群节点
cluster_formation.peer_discovery_backend = rabbit_peer_discovery_classic_config
cluster_formation.classic_config.nodes.1 = %s

# 集群名称
cluster_name = app_cluster

# 自动同步
cluster_formation.node_cleanup.interval = 10
cluster_formation.node_cleanup.only_log_warning = true

# 镜像队列策略
# ha-mode = all
# ha-sync-mode = automatic
`, strings.Join(nodeList, "\ncluster_formation.classic_config.nodes.2 = "))
}

// generateAppConfig 生成应用配置
func (g *ConfigGenerator) generateAppConfig(plan *planner.DeployPlan, assignment planner.ServerAssignment, services []planner.ServiceConfig) []ConfigFile {
	var configs []ConfigFile
	
	// 生成 systemd 服务文件
	for _, svc := range assignment.Services {
		for _, s := range services {
			if s.Name == svc {
				serviceFile := g.generateSystemdService(s, assignment)
				configs = append(configs, ConfigFile{
					Path:    fmt.Sprintf("/etc/systemd/system/%s.service", s.Name),
					Content: serviceFile,
					Mode:    "0644",
					Owner:   "root",
					Group:   "root",
					Reload:  false,
				})
			}
		}
	}
	
	// 生成 Docker Compose 文件（如果使用 Docker）
	composeFile := g.generateDockerCompose(plan, assignment, services)
	configs = append(configs, ConfigFile{
		Path:    "/opt/app/docker-compose.yml",
		Content: composeFile,
		Mode:    "0644",
		Owner:   "root",
		Group:   "root",
		Reload:  false,
	})
	
	return configs
}

// generateSystemdService 生成 systemd 服务文件
func (g *ConfigGenerator) generateSystemdService(svc planner.ServiceConfig, assignment planner.ServerAssignment) string {
	return fmt.Sprintf(`[Unit]
Description=%s Service
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/app/%s
ExecStartPre=/usr/bin/docker pull %s
ExecStart=/usr/bin/docker run --name %s -p %d:%d %s
ExecStop=/usr/bin/docker stop %s
ExecStopPost=/usr/bin/docker rm %s
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
`, svc.Name, svc.Name, svc.Image, svc.Name, 
	svc.Ports[0].HostPort, svc.Ports[0].ContainerPort, svc.Image,
	svc.Name, svc.Name)
}

// generateDockerCompose 生成 Docker Compose 文件
func (g *ConfigGenerator) generateDockerCompose(plan *planner.DeployPlan, assignment planner.ServerAssignment, services []planner.ServiceConfig) string {
	var sb strings.Builder
	
	sb.WriteString("version: '3.8'\n\n")
	sb.WriteString("services:\n")
	
	for _, svcName := range assignment.Services {
		for _, svc := range services {
			if svc.Name == svcName {
				sb.WriteString(fmt.Sprintf("  %s:\n", svc.Name))
				sb.WriteString(fmt.Sprintf("    image: %s\n", svc.Image))
				sb.WriteString(fmt.Sprintf("    container_name: %s\n", svc.Name))
				sb.WriteString("    restart: always\n")
				
				// 端口
				if len(svc.Ports) > 0 {
					sb.WriteString("    ports:\n")
					for _, port := range svc.Ports {
						sb.WriteString(fmt.Sprintf("      - \"%d:%d\"\n", port.HostPort, port.ContainerPort))
					}
				}
				
				// 环境变量
				if len(svc.Env) > 0 {
					sb.WriteString("    environment:\n")
					for k, v := range svc.Env {
						sb.WriteString(fmt.Sprintf("      - %s=%s\n", k, v))
					}
				}
				
				// 资源限制
				sb.WriteString("    deploy:\n")
				sb.WriteString("      resources:\n")
				sb.WriteString("        limits:\n")
				sb.WriteString(fmt.Sprintf("          cpus: '%d'\n", svc.Resources.CPU))
				sb.WriteString(fmt.Sprintf("          memory: %dM\n", svc.Resources.Memory))
				
				// 健康检查
				if svc.HealthCheck.Type != "" {
					sb.WriteString("    healthcheck:\n")
					if svc.HealthCheck.Type == "http" {
						sb.WriteString(fmt.Sprintf("      test: [\"CMD\", \"curl\", \"-f\", \"http://localhost:%d%s\"]\n", 
							svc.HealthCheck.Port, svc.HealthCheck.Path))
					}
					sb.WriteString(fmt.Sprintf("      interval: %ds\n", svc.HealthCheck.Interval))
					sb.WriteString(fmt.Sprintf("      timeout: %ds\n", svc.HealthCheck.Timeout))
					sb.WriteString(fmt.Sprintf("      retries: %d\n", svc.HealthCheck.Retries))
				}
				
				sb.WriteString("\n")
			}
		}
	}
	
	sb.WriteString("networks:\n")
	sb.WriteString("  default:\n")
	sb.WriteString("    external:\n")
	sb.WriteString("      name: app-network\n")
	
	return sb.String()
}

// generateHostsConfig 生成 hosts 配置
func (g *ConfigGenerator) generateHostsConfig(assignments []planner.ServerAssignment) []ConfigFile {
	var sb strings.Builder
	
	sb.WriteString("# 自动生成的 hosts 配置\n")
	sb.WriteString("# 各服务器之间的关联配置\n\n")
	
	for _, assignment := range assignments {
		sb.WriteString(fmt.Sprintf("%s %s\n", assignment.ServerIP, assignment.ServerName))
	}
	
	return []ConfigFile{{
		Path:    "/etc/hosts.d/app",
		Content: sb.String(),
		Mode:    "0644",
		Owner:   "root",
		Group:   "root",
		Reload:  false,
	}}
}

// generateFirewallConfig 生成防火墙配置
func (g *ConfigGenerator) generateFirewallConfig(networkConfig map[string]interface{}) []ConfigFile {
	var sb strings.Builder
	
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# 自动生成的防火墙规则\n\n")
	
	// 允许内部通信
	sb.WriteString("# 允许内部网络通信\n")
	sb.WriteString("iptables -A INPUT -s 10.0.0.0/8 -j ACCEPT\n")
	sb.WriteString("iptables -A INPUT -s 172.16.0.0/12 -j ACCEPT\n")
	sb.WriteString("iptables -A INPUT -s 192.168.0.0/16 -j ACCEPT\n\n")
	
	// 允许常用端口
	sb.WriteString("# 允许常用端口\n")
	sb.WriteString("iptables -A INPUT -p tcp --dport 22 -j ACCEPT\n")
	sb.WriteString("iptables -A INPUT -p tcp --dport 80 -j ACCEPT\n")
	sb.WriteString("iptables -A INPUT -p tcp --dport 443 -j ACCEPT\n\n")
	
	// 允许服务端口
	sb.WriteString("# 允许服务端口\n")
	sb.WriteString("iptables -A INPUT -p tcp --dport 3306 -j ACCEPT\n")
	sb.WriteString("iptables -A INPUT -p tcp --dport 6379 -j ACCEPT\n")
	sb.WriteString("iptables -A INPUT -p tcp --dport 5672 -j ACCEPT\n\n")
	
	// 默认策略
	sb.WriteString("# 默认策略\n")
	sb.WriteString("iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT\n")
	sb.WriteString("iptables -A INPUT -j DROP\n")
	
	return []ConfigFile{{
		Path:    "/etc/firewall/rules.sh",
		Content: sb.String(),
		Mode:    "0755",
		Owner:   "root",
		Group:   "root",
		Reload:  false,
	}}
}

// generateEnvConfig 生成环境变量配置
func (g *ConfigGenerator) generateEnvConfig(plan *planner.DeployPlan, assignment planner.ServerAssignment) []ConfigFile {
	var sb strings.Builder
	
	sb.WriteString("# 自动生成的环境变量\n\n")
	sb.WriteString(fmt.Sprintf("SERVER_ROLE=%s\n", assignment.Role))
	sb.WriteString(fmt.Sprintf("SERVER_ID=%d\n", assignment.ServerID))
	sb.WriteString(fmt.Sprintf("DEPLOY_PLAN_ID=%d\n", plan.ID))
	
	// 解析数据库配置并添加连接信息
	var dbConfig map[string]interface{}
	json.Unmarshal([]byte(plan.DatabaseConfig), &dbConfig)
	if master, ok := dbConfig["master"].(map[string]interface{}); ok {
		sb.WriteString(fmt.Sprintf("DB_MASTER_HOST=%s\n", master["server"]))
		sb.WriteString("DB_MASTER_PORT=3306\n")
	}
	
	// 解析缓存配置
	var cacheConfig map[string]interface{}
	json.Unmarshal([]byte(plan.CacheConfig), &cacheConfig)
	if nodes, ok := cacheConfig["nodes"].([]interface{}); ok && len(nodes) > 0 {
		if node, ok := nodes[0].(map[string]interface{}); ok {
			sb.WriteString(fmt.Sprintf("REDIS_HOST=%s\n", node["server"]))
			sb.WriteString("REDIS_PORT=6379\n")
		}
	}
	
	// 解析 MQ 配置
	var mqConfig map[string]interface{}
	json.Unmarshal([]byte(plan.MQConfig), &mqConfig)
	if nodes, ok := mqConfig["nodes"].([]interface{}); ok && len(nodes) > 0 {
		if node, ok := nodes[0].(map[string]interface{}); ok {
			sb.WriteString(fmt.Sprintf("MQ_HOST=%s\n", node["server"]))
			sb.WriteString("MQ_PORT=5672\n")
		}
	}
	
	return []ConfigFile{{
		Path:    "/opt/app/.env",
		Content: sb.String(),
		Mode:    "0600",
		Owner:   "root",
		Group:   "root",
		Reload:  false,
	}}
}
