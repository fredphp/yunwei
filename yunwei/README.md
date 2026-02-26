# AI 自动化运维管理系统 (Yunwei)

基于 Go + Gin 的 AI 自动化运维管理系统，实现多服务器集中管理、实时监控、AI 分析和自动修复。

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        管理平台                                  │
│   Dashboard │ Servers │ Monitor │ Alerts │ AI │ Tasks          │
├─────────────────────────────────────────────────────────────────┤
│                    WebSocket 实时推送                            │
├─────────────────────────────────────────────────────────────────┤
│                   Backend API (Go + Gin)                         │
│         SSH客户端 │ gRPC服务 │ AI引擎                           │
├─────────────────────────────────────────────────────────────────┤
│                       gRPC 通信层                                │
├─────────────────────────────────────────────────────────────────┤
│     Agent 1    │    Agent 2    │    Agent 3    │    Agent N    │
└─────────────────────────────────────────────────────────────────┘
```

## 功能模块

### 1. 服务器管理模块

| 功能 | 说明 |
|------|------|
| 添加服务器 | 支持 SSH 密码/密钥认证 |
| SSH 自动检测 | 添加时自动检测连接状态 |
| 服务器分组 | 树形分组管理 |
| 状态监控 | 在线/离线状态实时更新 |
| 远程命令 | 在线执行远程命令 |

### 2. Agent 节点程序

#### 功能
- 收集系统数据
- 执行远程命令
- 返回日志
- 监听 AI 指令

#### 采集指标

| 类型 | 指标 |
|------|------|
| CPU | 使用率、用户态、内核态、空闲、IO等待 |
| 内存 | 使用率、已用、空闲、缓存 |
| 磁盘 | 使用率、已用、空闲、IO读写 |
| 网络 | 入流量、出流量、包数 |
| 负载 | 1分钟、5分钟、15分钟 |
| 进程 | 进程数、TOP进程列表 |
| Docker | 容器列表、状态、资源使用 |
| 端口 | 监听端口、占用进程 |

## 技术栈

### 后端
- Go 1.22+
- Gin
- GORM
- MySQL 8.0
- gRPC
- WebSocket

### Agent
- Go 1.22+
- gRPC
- 系统命令执行

## 项目结构

```
yunwei/
├── server/                      # 后端服务
│   ├── api/v1/                  # API 接口
│   │   ├── auth/               # 认证
│   │   └── server/             # 服务器管理
│   ├── config/                  # 配置
│   ├── global/                  # 全局变量
│   ├── middleware/              # 中间件
│   ├── model/                   # 数据模型
│   ├── router/                  # 路由
│   ├── service/
│   │   ├── ssh/                # SSH 客户端
│   │   └── ai/                 # AI 服务
│   └── main.go
├── agent/                       # Agent 程序
│   ├── collector/              # 指标采集
│   ├── executor/               # 命令执行
│   ├── reporter/               # 数据上报
│   └── main.go
├── web/                         # 前端
├── sql/                         # 数据库脚本
└── deploy/                      # 部署配置
```

## 快速启动

### 后端服务

```bash
cd server

# 安装依赖
go mod tidy

# 运行
go run main.go
```

### Agent 程序

```bash
cd agent

# 安装依赖
go mod tidy

# 运行
go run main.go --server=localhost:50051 --name=agent-1
```

### Docker 部署

```bash
docker-compose up -d
```

## API 接口

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/login | 登录 |
| POST | /api/v1/register | 注册 |
| GET | /api/v1/user/info | 获取用户信息 |

### 服务器管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/servers | 服务器列表 |
| GET | /api/v1/servers/:id | 服务器详情 |
| POST | /api/v1/servers | 添加服务器 |
| PUT | /api/v1/servers/:id | 更新服务器 |
| DELETE | /api/v1/servers/:id | 删除服务器 |
| GET | /api/v1/servers/:id/metrics | 获取指标 |
| POST | /api/v1/servers/:id/command | 执行命令 |
| POST | /api/v1/servers/:id/refresh | 刷新状态 |
| POST | /api/v1/ssh/test | 测试 SSH 连接 |

### 分组管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/groups | 分组列表 |
| POST | /api/v1/groups | 创建分组 |
| DELETE | /api/v1/groups/:id | 删除分组 |

## 配置说明

```yaml
# server/config/config.yaml
system:
  port: 8080
  grpc-port: 50051
  env: develop

mysql:
  host: 127.0.0.1
  port: 3306
  username: root
  password: 123456
  database: yunwei

jwt:
  signing-key: yunwei-secret-key
```

## 默认账号

- 用户名: `admin`
- 密码: `admin123`

## License

MIT

---

# 项目部署文档

> 本文档详细介绍 AI 自动化运维管理系统的完整部署流程，涵盖开发环境、测试环境和生产环境。

---

## 目录

- [一、环境要求](#一环境要求)
- [二、安装部署](#二安装部署)
- [三、配置详解](#三配置详解)
- [四、数据库配置](#四数据库配置)
- [五、后端部署](#五后端部署)
- [六、前端部署](#六前端部署)
- [七、Agent 部署](#七agent-部署)
- [八、Docker 部署](#八docker-部署)
- [九、Kubernetes 部署](#九kubernetes-部署)
- [十、生产环境配置](#十生产环境配置)
- [十一、Nginx 配置](#十一nginx-配置)
- [十二、SSL/HTTPS 配置](#十二sslhttps-配置)
- [十三、性能调优](#十三性能调优)
- [十四、监控告警配置](#十四监控告警配置)
- [十五、日志管理](#十五日志管理)
- [十六、备份恢复](#十六备份恢复)
- [十七、安全加固](#十七安全加固)
- [十八、常见问题 FAQ](#十八常见问题-faq)
- [十九、故障排查](#十九故障排查)

---

## 一、环境要求

### 1.1 操作系统支持

| 操作系统 | 版本要求 | 架构 |
|---------|---------|------|
| CentOS | 7.x / 8.x / 9.x | x86_64, ARM64 |
| Ubuntu | 18.04 / 20.04 / 22.04 / 24.04 | x86_64, ARM64 |
| Debian | 10 / 11 / 12 | x86_64, ARM64 |
| Rocky Linux | 8.x / 9.x | x86_64, ARM64 |
| macOS | 11+ | x86_64, ARM64 (M1/M2) |
| Windows Server | 2019 / 2022 | x86_64 |
| Windows | 10 / 11 | x86_64 |

### 1.2 软件依赖

#### 后端依赖

| 软件 | 最低版本 | 推荐版本 | 说明 |
|------|---------|---------|------|
| Go | 1.21 | 1.22+ | 后端开发语言 |
| MySQL | 8.0 | 8.0.35+ | 主数据库 |
| PostgreSQL | 14 | 16+ | 可选数据库 |
| Redis | 6.0 | 7.2+ | 缓存/会话存储 |
| Git | 2.30 | 最新版 | 版本控制 |

#### 前端依赖

| 软件 | 最低版本 | 推荐版本 | 说明 |
|------|---------|---------|------|
| Node.js | 18.0 | 20.x LTS | JavaScript 运行时 |
| npm | 9.0 | 10.x | 包管理器 |
| pnpm | 8.0 | 最新版 | 可选，更快的包管理器 |

#### 容器化依赖

| 软件 | 最低版本 | 推荐版本 | 说明 |
|------|---------|---------|------|
| Docker | 24.0 | 25.x+ | 容器运行时 |
| Docker Compose | 2.20 | 2.x 最新版 | 容器编排 |
| Kubernetes | 1.26 | 1.29+ | 可选，K8s 部署 |

### 1.3 硬件要求

#### 最低配置（开发/测试环境）

| 资源 | 要求 |
|------|------|
| CPU | 2 核 |
| 内存 | 4 GB |
| 磁盘 | 50 GB SSD |
| 网络 | 100 Mbps |

#### 推荐配置（生产环境）

| 资源 | 要求 |
|------|------|
| CPU | 8 核+ |
| 内存 | 16 GB+ |
| 磁盘 | 500 GB+ SSD（系统盘）+ 数据盘 |
| 网络 | 1 Gbps |

#### 高可用配置（大规模生产）

| 资源 | 要求 |
|------|------|
| CPU | 16 核+ |
| 内存 | 32 GB+ |
| 磁盘 | 1 TB+ NVMe SSD |
| 网络 | 10 Gbps |
| 服务器数量 | 3+（负载均衡 + 高可用） |

### 1.4 网络端口

| 端口 | 服务 | 协议 | 说明 |
|------|------|------|------|
| 80 | Nginx | HTTP | Web 入口（重定向到 HTTPS） |
| 443 | Nginx | HTTPS | Web 安全入口 |
| 8080 | Go Backend | HTTP | 后端 API 服务 |
| 50051 | gRPC | gRPC | Agent 通信端口 |
| 3306 | MySQL | TCP | 数据库服务 |
| 6379 | Redis | TCP | 缓存服务 |
| 9090 | Prometheus | HTTP | 监控指标（可选） |
| 3000 | Grafana | HTTP | 监控面板（可选） |

### 1.5 环境检查脚本

```bash
#!/bin/bash
# check_environment.sh - 环境检查脚本

echo "===== 系统信息 ====="
echo "操作系统: $(cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2)"
echo "内核版本: $(uname -r)"
echo "架构: $(uname -m)"

echo ""
echo "===== 软件版本 ====="

# Go
if command -v go &> /dev/null; then
    echo "Go: $(go version)"
else
    echo "Go: 未安装"
fi

# Node.js
if command -v node &> /dev/null; then
    echo "Node.js: $(node -v)"
else
    echo "Node.js: 未安装"
fi

# MySQL
if command -v mysql &> /dev/null; then
    echo "MySQL: $(mysql --version)"
else
    echo "MySQL: 未安装"
fi

# Redis
if command -v redis-server &> /dev/null; then
    echo "Redis: $(redis-server --version)"
else
    echo "Redis: 未安装"
fi

# Docker
if command -v docker &> /dev/null; then
    echo "Docker: $(docker --version)"
else
    echo "Docker: 未安装"
fi

echo ""
echo "===== 系统资源 ====="
echo "CPU 核心数: $(nproc)"
echo "总内存: $(free -h | grep Mem | awk '{print $2}')"
echo "可用内存: $(free -h | grep Mem | awk '{print $7}')"
echo "磁盘空间: $(df -h / | tail -1 | awk '{print $4}') 可用"

echo ""
echo "===== 网络端口 ====="
for port in 80 443 8080 50051 3306 6379; do
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        echo "端口 $port: 已占用"
    else
        echo "端口 $port: 可用"
    fi
done
```

---

## 二、安装部署

### 2.1 方式一：源码安装

#### 2.1.1 克隆代码

```bash
# 克隆仓库
git clone https://github.com/fredphp/yunwei.git
cd yunwei
```

#### 2.1.2 安装 Go 环境

```bash
# CentOS/RHEL
sudo yum install -y golang

# Ubuntu/Debian
sudo apt update
sudo apt install -y golang-go

# 或使用官方安装方式（推荐）
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### 2.1.3 安装 Node.js 环境

```bash
# 使用 nvm 安装（推荐）
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc
nvm install 20
nvm use 20

# 或使用包管理器安装
# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# CentOS/RHEL
curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash -
sudo yum install -y nodejs

# 验证安装
node -v
npm -v
```

#### 2.1.4 安装数据库

**MySQL 安装：**

```bash
# CentOS 8+
sudo dnf install -y mysql-server
sudo systemctl start mysqld
sudo systemctl enable mysqld
sudo mysql_secure_installation

# Ubuntu/Debian
sudo apt install -y mysql-server
sudo systemctl start mysql
sudo systemctl enable mysql
sudo mysql_secure_installation

# Docker 方式安装 MySQL（推荐）
docker run -d \
  --name yunwei-mysql \
  -e MYSQL_ROOT_PASSWORD=your_strong_password \
  -e MYSQL_DATABASE=yunwei \
  -e MYSQL_USER=yunwei \
  -e MYSQL_PASSWORD=yunwei_password \
  -p 3306:3306 \
  -v /data/mysql:/var/lib/mysql \
  --restart=always \
  mysql:8.0 \
  --character-set-server=utf8mb4 \
  --collation-server=utf8mb4_unicode_ci
```

**PostgreSQL 安装（可选）：**

```bash
# Ubuntu/Debian
sudo apt install -y postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql

# 创建数据库和用户
sudo -u postgres psql
CREATE DATABASE yunwei;
CREATE USER yunwei WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE yunwei TO yunwei;
\q
```

#### 2.1.5 安装 Redis

```bash
# CentOS
sudo yum install -y redis
sudo systemctl start redis
sudo systemctl enable redis

# Ubuntu/Debian
sudo apt install -y redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server

# Docker 方式安装 Redis（推荐）
docker run -d \
  --name yunwei-redis \
  -p 6379:6379 \
  -v /data/redis:/data \
  --restart=always \
  redis:7.2-alpine \
  redis-server --appendonly yes --requirepass your_redis_password

# 验证连接
redis-cli ping
```

### 2.2 方式二：Docker 安装

#### 2.2.1 安装 Docker

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# CentOS
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动 Docker
sudo systemctl start docker
sudo systemctl enable docker

# 验证安装
docker --version
docker compose version
```

#### 2.2.2 使用 Docker Compose 部署

```bash
# 创建部署目录
mkdir -p /opt/yunwei && cd /opt/yunwei

# 创建 docker-compose.yml（见第八节）

# 启动服务
docker compose up -d

# 查看日志
docker compose logs -f

# 停止服务
docker compose down
```

---

## 三、配置详解

### 3.1 后端配置文件

创建配置文件 `server/config/config.yaml`：

```yaml
# ==================== 系统配置 ====================
system:
  port: "8080"              # HTTP 服务端口
  grpc-port: "50051"        # gRPC 服务端口
  env: "production"         # 环境: develop, test, production
  name: "yunwei"            # 应用名称
  mode: "release"           # Gin 模式: debug, release, test

# ==================== 数据库配置 ====================
mysql:
  host: "127.0.0.1"
  port: 3306
  username: "yunwei"
  password: "your_mysql_password"
  database: "yunwei"
  charset: "utf8mb4"
  max-idle-conns: 20        # 空闲连接池最大连接数
  max-open-conns: 100       # 数据库最大连接数
  conn-max-lifetime: 3600   # 连接最大存活时间(秒)
  log-mode: false           # 是否开启 SQL 日志

# ==================== Redis 配置 ====================
redis:
  host: "127.0.0.1"
  port: 6379
  password: "your_redis_password"
  db: 0
  pool-size: 100            # 连接池大小
  min-idle-conns: 10        # 最小空闲连接数

# ==================== JWT 配置 ====================
jwt:
  signing-key: "your-super-secret-key-at-least-32-characters-long"
  expires-time: "24h"       # Token 过期时间
  issuer: "yunwei"
  refresh-expires-time: "168h"  # 刷新 Token 过期时间(7天)

# ==================== AI 配置 ====================
ai:
  enabled: true
  api-key: "your_ai_api_key"
  base-url: "https://open.bigmodel.cn/api/paas/v4"
  model: "glm-4"
  max-tokens: 4096
  temperature: 0.7
  auto-execute: false       # 是否自动执行低风险命令
  timeout: 60               # AI 请求超时时间(秒)

# ==================== 安全配置 ====================
security:
  enable-whitelist: true    # 启用命令白名单
  enable-blacklist: true    # 启用命令黑名单
  require-approval: true    # 高危命令需要审批
  audit-enabled: true       # 启用审计日志
  audit-retention-days: 90  # 审计日志保留天数
  max-login-attempts: 5     # 最大登录尝试次数
  login-lock-duration: 30   # 登录锁定时间(分钟)
  password-min-length: 8    # 密码最小长度
  password-require-special: true  # 密码要求特殊字符

# ==================== 日志配置 ====================
log:
  level: "info"             # 日志级别: debug, info, warn, error
  format: "json"            # 日志格式: json, text
  output: "stdout"          # 输出: stdout, file
  file-path: "/var/log/yunwei/app.log"
  max-size: 100             # 单个日志文件最大大小(MB)
  max-backups: 10           # 保留旧日志文件最大数量
  max-age: 30               # 保留旧日志文件最大天数
  compress: true            # 是否压缩旧日志文件

# ==================== CORS 配置 ====================
cors:
  allow-origins:
    - "https://yunwei.example.com"
    - "https://admin.yunwei.example.com"
  allow-methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow-headers:
    - "Origin"
    - "Content-Type"
    - "Authorization"
    - "X-Tenant-ID"
    - "X-Request-ID"
  expose-headers:
    - "Content-Length"
    - "X-Request-ID"
  allow-credentials: true
  max-age: 43200            # 预检请求缓存时间(秒)

# ==================== 限流配置 ====================
rate-limit:
  enabled: true
  requests-per-second: 100  # 每秒请求数限制
  burst: 200                # 突发请求数
  by-ip: true               # 按 IP 限流
  by-user: true             # 按用户限流

# ==================== WebSocket 配置 ====================
websocket:
  path: "/ws"
  ping-period: 30           # Ping 间隔(秒)
  pong-wait: 60             # Pong 等待时间(秒)
  write-wait: 10            # 写等待时间(秒)
  max-message-size: 65536   # 最大消息大小(字节)

# ==================== 存储配置 ====================
storage:
  type: "local"             # 存储类型: local, s3, oss
  local:
    path: "/data/yunwei/uploads"
  s3:
    endpoint: "s3.amazonaws.com"
    region: "us-east-1"
    bucket: "yunwei"
    access-key: "your_access_key"
    secret-key: "your_secret_key"
  oss:
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    bucket: "yunwei"
    access-key: "your_access_key"
    secret-key: "your_secret_key"

# ==================== 邮件配置 ====================
email:
  enabled: true
  host: "smtp.example.com"
  port: 587
  username: "noreply@example.com"
  password: "your_smtp_password"
  from: "Yunwei <noreply@example.com>"
  use-tls: true

# ==================== 多租户配置 ====================
tenant:
  enabled: true
  default-plan: "free"
  trial-days: 14
  max-free-tenants: 10

# ==================== 备份配置 ====================
backup:
  enabled: true
  schedule: "0 2 * * *"     # 每天凌晨2点执行
  retention-days: 30
  storage-path: "/data/yunwei/backups"
  compress: true
  encrypt: true
  encrypt-key: "your_backup_encrypt_key"
```

### 3.2 前端配置文件

创建环境配置文件 `.env.production`：

```bash
# API 配置
NEXT_PUBLIC_API_URL=https://api.yunwei.example.com
NEXT_PUBLIC_WS_URL=wss://api.yunwei.example.com/ws

# 应用配置
NEXT_PUBLIC_APP_NAME=Yunwei
NEXT_PUBLIC_APP_VERSION=1.0.0

# 认证配置
NEXT_PUBLIC_JWT_STORAGE_KEY=yunwei_token
NEXT_PUBLIC_TOKEN_REFRESH_THRESHOLD=300

# 功能开关
NEXT_PUBLIC_ENABLE_AI=true
NEXT_PUBLIC_ENABLE_MONITORING=true
NEXT_PUBLIC_ENABLE_MULTI_TENANT=true

# 第三方服务
NEXT_PUBLIC_SENTRY_DSN=https://xxx@sentry.io/xxx
NEXT_PUBLIC_GA_ID=UA-XXXXXXXXX-X
```

### 3.3 环境变量配置

创建 `.env` 文件：

```bash
# 数据库
DB_HOST=localhost
DB_PORT=3306
DB_USER=yunwei
DB_PASSWORD=your_mysql_password
DB_NAME=yunwei

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# JWT
JWT_SECRET=your-super-secret-key-at-least-32-characters-long
JWT_EXPIRE=24h

# AI 服务
AI_API_KEY=your_ai_api_key
AI_BASE_URL=https://open.bigmodel.cn/api/paas/v4
AI_MODEL=glm-4

# 应用
APP_ENV=production
APP_PORT=8080
APP_GRPC_PORT=50051

# 日志
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## 四、数据库配置

### 4.1 创建数据库

```sql
-- 连接 MySQL
mysql -u root -p

-- 创建数据库
CREATE DATABASE yunwei DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户
CREATE USER 'yunwei'@'%' IDENTIFIED BY 'your_strong_password';

-- 授权
GRANT ALL PRIVILEGES ON yunwei.* TO 'yunwei'@'%';
FLUSH PRIVILEGES;

-- 验证
SHOW DATABASES;
SELECT user, host FROM mysql.user;
```

### 4.2 数据库初始化

```bash
# 进入项目目录
cd yunwei/server

# 方式一：使用 GORM 自动迁移（推荐）
# 程序启动时会自动创建表结构

# 方式二：使用 SQL 脚本初始化
mysql -u yunwei -p yunwei < sql/init.sql
```

### 4.3 MySQL 配置优化

编辑 `/etc/my.cnf` 或 `/etc/mysql/mysql.conf.d/mysqld.cnf`：

```ini
[mysqld]
# 基础配置
port = 3306
bind-address = 0.0.0.0
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
default-time-zone = '+08:00'

# InnoDB 配置
innodb_buffer_pool_size = 4G           # 系统内存的 50-70%
innodb_log_file_size = 512M
innodb_log_buffer_size = 64M
innodb_flush_log_at_trx_commit = 1
innodb_lock_wait_timeout = 50
innodb_flush_method = O_DIRECT

# 连接配置
max_connections = 500
max_connect_errors = 100
wait_timeout = 28800
interactive_timeout = 28800

# 查询缓存（MySQL 8.0 已移除）
# query_cache_type = 1
# query_cache_size = 64M

# 日志配置
log_error = /var/log/mysql/error.log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
log_queries_not_using_indexes = 1

# 二进制日志（主从复制需要）
log_bin = mysql-bin
binlog_format = ROW
expire_logs_days = 7
server_id = 1

# 安全配置
skip-name-resolve
local_infile = 0

[mysql]
default-character-set = utf8mb4

[client]
default-character-set = utf8mb4
```

重启 MySQL 服务：

```bash
sudo systemctl restart mysqld
# 或
sudo systemctl restart mysql
```

### 4.4 Redis 配置优化

编辑 `/etc/redis/redis.conf`：

```conf
# 基础配置
bind 127.0.0.1
port 6379
daemonize yes
pidfile /var/run/redis/redis-server.pid
logfile /var/log/redis/redis-server.log

# 内存配置
maxmemory 2gb
maxmemory-policy allkeys-lru

# 持久化配置
save 900 1
save 300 10
save 60 10000
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec

# 安全配置
requirepass your_redis_password
rename-command FLUSHALL ""
rename-command FLUSHDB ""
rename-command KEYS ""

# 性能配置
tcp-keepalive 300
timeout 0
tcp-backlog 511
```

重启 Redis 服务：

```bash
sudo systemctl restart redis
# 或
sudo systemctl restart redis-server
```

---

## 五、后端部署

### 5.1 编译构建

```bash
cd yunwei/server

# 下载依赖
go mod download
go mod tidy

# 编译 Linux amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o yunwei-server

# 编译 Linux arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o yunwei-server-arm64

# 编译 macOS
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o yunwei-server-darwin

# 编译 Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o yunwei-server.exe

# 查看编译结果
ls -la yunwei-server
```

### 5.2 目录结构

```bash
# 创建部署目录
sudo mkdir -p /opt/yunwei/{bin,config,logs,data}

# 复制文件
sudo cp yunwei-server /opt/yunwei/bin/
sudo cp -r config/* /opt/yunwei/config/

# 设置权限
sudo chmod +x /opt/yunwei/bin/yunwei-server
sudo chown -R yunwei:yunwei /opt/yunwei
```

### 5.3 创建系统用户

```bash
# 创建用户
sudo useradd -r -s /bin/false -d /opt/yunwei yunwei

# 设置目录权限
sudo chown -R yunwei:yunwei /opt/yunwei
```

### 5.4 Systemd 服务配置

创建服务文件 `/etc/systemd/system/yunwei.service`：

```ini
[Unit]
Description=Yunwei AI Operations Management System
Documentation=https://github.com/fredphp/yunwei
After=network.target mysql.service redis.service
Wants=mysql.service redis.service

[Service]
Type=simple
User=yunwei
Group=yunwei
WorkingDirectory=/opt/yunwei
ExecStart=/opt/yunwei/bin/yunwei-server
ExecReload=/bin/kill -HUP $MAINPID
ExecStop=/bin/kill -s QUIT $MAINPID
Restart=on-failure
RestartSec=5s
LimitNOFILE=65535
LimitNPROC=65535
StandardOutput=journal
StandardError=journal
Environment="GIN_MODE=release"
Environment="PATH=/usr/local/bin:/usr/bin:/bin"

# 安全配置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/yunwei/logs /opt/yunwei/data
ReadOnlyPaths=/opt/yunwei/config

[Install]
WantedBy=multi-user.target
```

### 5.5 启动服务

```bash
# 重载 systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start yunwei

# 开机自启
sudo systemctl enable yunwei

# 查看状态
sudo systemctl status yunwei

# 查看日志
sudo journalctl -u yunwei -f
```

### 5.6 健康检查

```bash
# 检查服务健康状态
curl http://localhost:8080/health

# 预期响应
# {"status":"ok","message":"Yunwei Server is running"}
```

---

## 六、前端部署

### 6.1 构建前端

```bash
cd yunwei/web

# 安装依赖
npm install
# 或使用 pnpm
pnpm install

# 构建生产版本
npm run build
# 或
pnpm build

# 查看构建结果
ls -la .next/
```

### 6.2 方式一：Node.js 运行

```bash
# 直接运行
npm run start

# 指定端口
PORT=3000 npm run start
```

### 6.3 方式二：PM2 管理

```bash
# 安装 PM2
npm install -g pm2

# 创建 PM2 配置文件 ecosystem.config.js
cat > ecosystem.config.js << 'EOF'
module.exports = {
  apps: [{
    name: 'yunwei-web',
    script: 'npm',
    args: 'start',
    cwd: '/opt/yunwei/web',
    instances: 'max',
    exec_mode: 'cluster',
    autorestart: true,
    watch: false,
    max_memory_restart: '1G',
    env: {
      NODE_ENV: 'production',
      PORT: 3000
    }
  }]
}
EOF

# 启动
pm2 start ecosystem.config.js

# 开机自启
pm2 startup
pm2 save

# 查看状态
pm2 status
pm2 logs yunwei-web
```

### 6.4 方式三：静态文件部署

```bash
# 导出静态文件
npm run export

# 输出目录为 out/
# 使用 Nginx 托管静态文件
```

### 6.5 Systemd 服务配置

创建服务文件 `/etc/systemd/system/yunwei-web.service`：

```ini
[Unit]
Description=Yunwei Web Frontend
After=network.target

[Service]
Type=simple
User=yunwei
Group=yunwei
WorkingDirectory=/opt/yunwei/web
ExecStart=/usr/bin/node /opt/yunwei/web/node_modules/.bin/next start -p 3000
Restart=on-failure
RestartSec=5s
Environment="NODE_ENV=production"

[Install]
WantedBy=multi-user.target
```

---

## 七、Agent 部署

### 7.1 编译 Agent

```bash
cd yunwei/agent

# 下载依赖
go mod download

# 编译各平台版本
# Linux amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o yunwei-agent

# Linux arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o yunwei-agent-arm64

# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o yunwei-agent.exe
```

### 7.2 Agent 配置文件

创建 `/opt/yunwei-agent/config.yaml`：

```yaml
# Agent 配置
server:
  address: "yunwei.example.com:50051"
  tls:
    enabled: true
    cert: "/opt/yunwei-agent/certs/client.crt"
    key: "/opt/yunwei-agent/certs/client.key"
    ca: "/opt/yunwei-agent/certs/ca.crt"

agent:
  id: ""                      # 留空自动生成
  name: "agent-01"            # Agent 名称
  secret: "your_agent_secret" # Agent 密钥
  
  # 心跳配置
  heartbeat:
    interval: 30              # 心跳间隔(秒)
    timeout: 10               # 超时时间(秒)

  # 采集配置
  collector:
    enabled: true
    interval: 60              # 采集间隔(秒)
    metrics:
      - cpu
      - memory
      - disk
      - network
      - load
      - process
      - docker
      - port

  # 执行配置
  executor:
    enabled: true
    timeout: 300              # 命令超时(秒)
    max_concurrent: 5         # 最大并发数
    work_dir: "/tmp/yunwei"   # 工作目录

  # 日志配置
  log:
    level: "info"
    path: "/var/log/yunwei-agent/agent.log"
    max_size: 50
    max_backups: 5
    max_age: 7
    compress: true
```

### 7.3 Agent Systemd 服务

创建 `/etc/systemd/system/yunwei-agent.service`：

```ini
[Unit]
Description=Yunwei Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/yunwei-agent
ExecStart=/opt/yunwei-agent/yunwei-agent -config /opt/yunwei-agent/config.yaml
Restart=on-failure
RestartSec=5s
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### 7.4 部署到目标服务器

```bash
# 创建目录
ssh root@target-server "mkdir -p /opt/yunwei-agent/{config,certs,logs}"

# 复制文件
scp yunwei-agent root@target-server:/opt/yunwei-agent/
scp config.yaml root@target-server:/opt/yunwei-agent/config/

# 设置权限
ssh root@target-server "chmod +x /opt/yunwei-agent/yunwei-agent"

# 启动服务
ssh root@target-server "systemctl daemon-reload && systemctl start yunwei-agent && systemctl enable yunwei-agent"

# 查看状态
ssh root@target-server "systemctl status yunwei-agent"
```

---

## 八、Docker 部署

### 8.1 后端 Dockerfile

创建 `server/Dockerfile`：

```dockerfile
# 构建阶段
FROM golang:1.22-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o yunwei-server .

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装必要工具
RUN apk add --no-cache ca-certificates tzdata

# 复制二进制文件
COPY --from=builder /build/yunwei-server .
COPY --from=builder /build/config ./config

# 创建非 root 用户
RUN adduser -D -u 1000 yunwei
USER yunwei

EXPOSE 8080 50051

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./yunwei-server"]
```

### 8.2 前端 Dockerfile

创建 `web/Dockerfile`：

```dockerfile
# 构建阶段
FROM node:20-alpine AS builder

WORKDIR /app

# 复制 package 文件
COPY package.json package-lock.json* ./

# 安装依赖
RUN npm ci

# 复制源代码
COPY . .

# 构建
RUN npm run build

# 运行阶段
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# 创建非 root 用户
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# 复制构建产物
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/api/health || exit 1

CMD ["node", "server.js"]
```

### 8.3 Docker Compose 完整配置

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  # ==================== 后端服务 ====================
  backend:
    build:
      context: ./server
      dockerfile: Dockerfile
    container_name: yunwei-backend
    restart: always
    ports:
      - "8080:8080"
      - "50051:50051"
    environment:
      - GIN_MODE=release
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=yunwei
      - DB_PASSWORD=${MYSQL_PASSWORD:-yunwei_password}
      - DB_NAME=yunwei
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD:-redis_password}
      - JWT_SECRET=${JWT_SECRET:-your-secret-key}
    volumes:
      - ./server/config:/app/config:ro
      - backend-data:/app/data
      - backend-logs:/app/logs
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - yunwei-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # ==================== 前端服务 ====================
  frontend:
    build:
      context: ./web
      dockerfile: Dockerfile
    container_name: yunwei-frontend
    restart: always
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - NEXT_PUBLIC_API_URL=${API_URL:-http://localhost:8080}
    depends_on:
      - backend
    networks:
      - yunwei-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3000"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ==================== MySQL 数据库 ====================
  mysql:
    image: mysql:8.0
    container_name: yunwei-mysql
    restart: always
    ports:
      - "3306:3306"
    environment:
      - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-root_password}
      - MYSQL_DATABASE=yunwei
      - MYSQL_USER=yunwei
      - MYSQL_PASSWORD=${MYSQL_PASSWORD:-yunwei_password}
    volumes:
      - mysql-data:/var/lib/mysql
      - ./sql/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
      - --default-authentication-plugin=mysql_native_password
      - --max_connections=500
      - --innodb_buffer_pool_size=1G
    networks:
      - yunwei-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-p${MYSQL_ROOT_PASSWORD:-root_password}"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

  # ==================== Redis 缓存 ====================
  redis:
    image: redis:7.2-alpine
    container_name: yunwei-redis
    restart: always
    ports:
      - "6379:6379"
    command: >
      redis-server
      --appendonly yes
      --requirepass ${REDIS_PASSWORD:-redis_password}
      --maxmemory 1gb
      --maxmemory-policy allkeys-lru
    volumes:
      - redis-data:/data
    networks:
      - yunwei-network
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD:-redis_password}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # ==================== Nginx 反向代理 ====================
  nginx:
    image: nginx:1.25-alpine
    container_name: yunwei-nginx
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - nginx-logs:/var/log/nginx
    depends_on:
      - backend
      - frontend
    networks:
      - yunwei-network
    healthcheck:
      test: ["CMD", "nginx", "-t"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ==================== Prometheus 监控（可选） ====================
  prometheus:
    image: prom/prometheus:v2.48.0
    container_name: yunwei-prometheus
    restart: always
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
    networks:
      - yunwei-network
    profiles:
      - monitoring

  # ==================== Grafana 可视化（可选） ====================
  grafana:
    image: grafana/grafana:10.2.0
    container_name: yunwei-grafana
    restart: always
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin123}
      - GF_INSTALL_PLUGINS=redis-datasource
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning:ro
    depends_on:
      - prometheus
    networks:
      - yunwei-network
    profiles:
      - monitoring

networks:
  yunwei-network:
    driver: bridge

volumes:
  mysql-data:
  redis-data:
  backend-data:
  backend-logs:
  nginx-logs:
  prometheus-data:
  grafana-data:
```

### 8.4 启动 Docker 服务

```bash
# 创建环境变量文件
cat > .env << 'EOF'
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_PASSWORD=your_mysql_password
REDIS_PASSWORD=your_redis_password
JWT_SECRET=your-super-secret-key-at-least-32-characters
API_URL=https://api.yunwei.example.com
GRAFANA_PASSWORD=your_grafana_password
EOF

# 启动核心服务
docker compose up -d

# 启动包含监控的完整服务
docker compose --profile monitoring up -d

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f backend

# 停止服务
docker compose down

# 停止并删除数据卷
docker compose down -v
```

---

## 九、Kubernetes 部署

### 9.1 Namespace 配置

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: yunwei
  labels:
    app: yunwei
```

### 9.2 ConfigMap 配置

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: yunwei-config
  namespace: yunwei
data:
  config.yaml: |
    system:
      port: "8080"
      grpc-port: "50051"
      env: "production"
    mysql:
      host: "mysql-service"
      port: 3306
      username: "yunwei"
      database: "yunwei"
    redis:
      host: "redis-service"
      port: 6379
      db: 0
```

### 9.3 Secret 配置

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: yunwei-secret
  namespace: yunwei
type: Opaque
stringData:
  mysql-password: your_mysql_password
  redis-password: your_redis_password
  jwt-secret: your-super-secret-key-at-least-32-characters
  ai-api-key: your_ai_api_key
```

### 9.4 后端 Deployment

```yaml
# k8s/backend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: yunwei-backend
  namespace: yunwei
spec:
  replicas: 3
  selector:
    matchLabels:
      app: yunwei-backend
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: yunwei-backend
    spec:
      containers:
      - name: backend
        image: yunwei/backend:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 50051
          name: grpc
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: yunwei-secret
              key: mysql-password
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: yunwei-secret
              key: redis-password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: yunwei-secret
              key: jwt-secret
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "2000m"
            memory: "2Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: yunwei-config
```

### 9.5 Service 配置

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: yunwei-backend
  namespace: yunwei
spec:
  selector:
    app: yunwei-backend
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: grpc
    port: 50051
    targetPort: 50051
  type: ClusterIP

---
apiVersion: v1
kind: Service
metadata:
  name: yunwei-frontend
  namespace: yunwei
spec:
  selector:
    app: yunwei-frontend
  ports:
  - name: http
    port: 3000
    targetPort: 3000
  type: ClusterIP
```

### 9.6 Ingress 配置

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: yunwei-ingress
  namespace: yunwei
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - yunwei.example.com
    - api.yunwei.example.com
    secretName: yunwei-tls
  rules:
  - host: yunwei.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: yunwei-frontend
            port:
              number: 3000
  - host: api.yunwei.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: yunwei-backend
            port:
              number: 8080
```

### 9.7 部署命令

```bash
# 创建命名空间
kubectl apply -f k8s/namespace.yaml

# 创建配置
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# 部署应用
kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml

# 创建服务
kubectl apply -f k8s/service.yaml

# 创建 Ingress
kubectl apply -f k8s/ingress.yaml

# 查看部署状态
kubectl get pods -n yunwei
kubectl get services -n yunwei
kubectl get ingress -n yunwei

# 查看日志
kubectl logs -f deployment/yunwei-backend -n yunwei
```

---

## 十、生产环境配置

### 10.1 负载均衡架构

```
                    ┌─────────────────┐
                    │   Load Balancer │
                    │   (Nginx/HAProxy)│
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
    ┌─────────┐         ┌─────────┐         ┌─────────┐
    │ Backend │         │ Backend │         │ Backend │
    │   #1    │         │   #2    │         │   #3    │
    └────┬────┘         └────┬────┘         └────┬────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
         ┌─────────┐   ┌─────────┐   ┌─────────┐
         │ MySQL   │   │ Redis   │   │ Storage │
         │ Primary │   │ Cluster │   │  (S3)   │
         │ + Slave │   │         │   │         │
         └─────────┘   └─────────┘   └─────────┘
```

### 10.2 高可用配置要点

#### 数据库高可用

```bash
# MySQL 主从复制配置

# 主库配置 (my.cnf)
[mysqld]
server-id = 1
log_bin = mysql-bin
binlog_format = ROW
gtid_mode = ON
enforce_gtid_consistency = ON

# 从库配置 (my.cnf)
[mysqld]
server-id = 2
log_bin = mysql-bin
binlog_format = ROW
gtid_mode = ON
enforce_gtid_consistency = ON
read_only = ON

# 配置复制
CHANGE MASTER TO
  MASTER_HOST='master-ip',
  MASTER_USER='repl',
  MASTER_PASSWORD='repl_password',
  MASTER_AUTO_POSITION=1;
START SLAVE;
```

#### Redis 高可用

```bash
# Redis Sentinel 配置
sentinel monitor mymaster 192.168.1.1 6379 2
sentinel down-after-milliseconds mymaster 30000
sentinel parallel-syncs mymaster 1
sentinel failover-timeout mymaster 180000
```

---

## 十一、Nginx 配置

### 11.1 主配置文件

创建 `/etc/nginx/nginx.conf`：

```nginx
user nginx;
worker_processes auto;
worker_rlimit_nofile 65535;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    use epoll;
    worker_connections 65535;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # 日志格式
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for" '
                    'rt=$request_time uct="$upstream_connect_time" '
                    'uht="$upstream_header_time" urt="$upstream_response_time"';

    access_log /var/log/nginx/access.log main;

    # 性能优化
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
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml application/json application/javascript 
               application/xml application/xml+rss text/javascript application/x-javascript;

    # 缓冲区设置
    client_body_buffer_size 128k;
    client_max_body_size 50m;
    large_client_header_buffers 4 16k;

    # 代理缓冲
    proxy_buffer_size 128k;
    proxy_buffers 4 256k;
    proxy_busy_buffers_size 256k;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # 上游服务器
    upstream backend {
        least_conn;
        server 127.0.0.1:8080 weight=5 max_fails=3 fail_timeout=30s;
        server 127.0.0.1:8081 weight=5 max_fails=3 fail_timeout=30s backup;
        keepalive 32;
    }

    upstream frontend {
        server 127.0.0.1:3000;
        keepalive 16;
    }

    # 包含站点配置
    include /etc/nginx/conf.d/*.conf;
}
```

### 11.2 站点配置

创建 `/etc/nginx/conf.d/yunwei.conf`：

```nginx
# 后端 API 服务
server {
    listen 80;
    server_name api.yunwei.example.com;

    # 重定向到 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yunwei.example.com;

    # SSL 配置
    ssl_certificate /etc/nginx/ssl/yunwei.example.com.crt;
    ssl_certificate_key /etc/nginx/ssl/yunwei.example.com.key;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;

    # 现代 SSL 配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    # HSTS
    add_header Strict-Transport-Security "max-age=63072000" always;

    # 日志
    access_log /var/log/nginx/api-access.log main;
    error_log /var/log/nginx/api-error.log;

    # API 代理
    location / {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";

        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # 限流
        limit_req zone=api_limit burst=100 nodelay;
    }

    # WebSocket 代理
    location /ws {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # 健康检查
    location /health {
        proxy_pass http://backend/health;
        access_log off;
    }
}

# 前端服务
server {
    listen 80;
    server_name yunwei.example.com;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yunwei.example.com;

    # SSL 配置
    ssl_certificate /etc/nginx/ssl/yunwei.example.com.crt;
    ssl_certificate_key /etc/nginx/ssl/yunwei.example.com.key;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    add_header Strict-Transport-Security "max-age=63072000" always;

    # 日志
    access_log /var/log/nginx/web-access.log main;
    error_log /var/log/nginx/web-error.log;

    # 静态资源缓存
    location /_next/static {
        proxy_pass http://frontend;
        proxy_cache static_cache;
        proxy_cache_valid 200 365d;
        proxy_cache_key $uri;
        add_header Cache-Control "public, max-age=31536000, immutable";
    }

    # 前端代理
    location / {
        proxy_pass http://frontend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### 11.3 限流配置

在 `http` 块中添加：

```nginx
# 限流区域
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/s;
limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

# 应用限流
limit_req zone=api_limit burst=200 nodelay;
limit_conn conn_limit 50;
```

---

## 十二、SSL/HTTPS 配置

### 12.1 Let's Encrypt 免费证书

```bash
# 安装 Certbot
# Ubuntu/Debian
sudo apt install -y certbot python3-certbot-nginx

# CentOS/RHEL
sudo dnf install -y certbot python3-certbot-nginx

# 申请证书
sudo certbot --nginx -d yunwei.example.com -d api.yunwei.example.com

# 自动续期
sudo certbot renew --dry-run

# 设置自动续期定时任务
sudo crontab -e
# 添加以下行
0 0 1 * * /usr/bin/certbot renew --quiet --post-hook "systemctl reload nginx"
```

### 12.2 自签名证书（测试环境）

```bash
# 生成私钥
openssl genrsa -out ca.key 4096

# 生成 CA 证书
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=Yunwei/CN=Yunwei CA"

# 生成服务器私钥
openssl genrsa -out server.key 2048

# 生成 CSR
openssl req -new -key server.key -out server.csr \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=Yunwei/CN=yunwei.example.com"

# 使用 CA 签发证书
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out server.crt

# 复制证书到 Nginx 目录
sudo mkdir -p /etc/nginx/ssl
sudo cp server.crt server.key /etc/nginx/ssl/
sudo chmod 600 /etc/nginx/ssl/server.key
```

### 12.3 SSL 配置最佳实践

```nginx
# SSL 配置
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;

# SSL 会话
ssl_session_timeout 1d;
ssl_session_cache shared:SSL:50m;
ssl_session_tickets off;

# OCSP Stapling
ssl_stapling on;
ssl_stapling_verify on;
resolver 8.8.8.8 8.8.4.4 valid=300s;
resolver_timeout 5s;

# HSTS
add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
```

---

## 十三、性能调优

### 13.1 系统内核调优

编辑 `/etc/sysctl.conf`：

```ini
# 网络优化
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_keepalive_probes = 5
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_max_tw_buckets = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_max_syn_backlog = 262144
net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syn_retries = 2

# 文件描述符
fs.file-max = 2097152
fs.nr_open = 2097152

# 内存优化
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
vm.overcommit_memory = 1
vm.max_map_count = 262144

# 共享内存
kernel.shmmax = 68719476736
kernel.shmall = 4294967296
```

应用配置：

```bash
sudo sysctl -p
```

### 13.2 文件描述符限制

编辑 `/etc/security/limits.conf`：

```
# 软限制和硬限制
* soft nofile 65535
* hard nofile 65535
* soft nproc 65535
* hard nproc 65535
root soft nofile 65535
root hard nofile 65535
```

### 13.3 Go 应用调优

```go
// main.go 中添加
import (
    "runtime"
)

func init() {
    // 设置使用所有 CPU 核心
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // 设置 GC 目标百分比
    // debug.SetGCPercent(100)
}
```

### 13.4 数据库连接池

```go
// 数据库连接池配置
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(20)      // 空闲连接数
sqlDB.SetMaxOpenConns(100)     // 最大连接数
sqlDB.SetConnMaxLifetime(time.Hour)  // 连接最大生命周期
```

### 13.5 Redis 连接池

```go
// Redis 连接池配置
rdb := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     100,
    MinIdleConns: 10,
    MaxRetries:   3,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
})
```

---

## 十四、监控告警配置

### 14.1 Prometheus 配置

创建 `prometheus/prometheus.yml`：

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - alertmanager:9093

rule_files:
  - /etc/prometheus/rules/*.yml

scrape_configs:
  # Prometheus 自身
  - job_name: 'prometheus'
    static_configs:
    - targets: ['localhost:9090']

  # Yunwei 后端
  - job_name: 'yunwei-backend'
    static_configs:
    - targets: ['backend:8080']
    metrics_path: /metrics

  # MySQL Exporter
  - job_name: 'mysql'
    static_configs:
    - targets: ['mysql-exporter:9104']

  # Redis Exporter
  - job_name: 'redis'
    static_configs:
    - targets: ['redis-exporter:9121']

  # Node Exporter
  - job_name: 'node'
    static_configs:
    - targets: ['node-exporter:9100']

  # Nginx Exporter
  - job_name: 'nginx'
    static_configs:
    - targets: ['nginx-exporter:9113']
```

### 14.2 告警规则

创建 `prometheus/rules/alerts.yml`：

```yaml
groups:
- name: yunwei-alerts
  rules:
  # 服务不可用
  - alert: ServiceDown
    expr: up == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "服务 {{ $labels.job }} 不可用"
      description: "{{ $labels.instance }} 已经停止超过 1 分钟"

  # 高 CPU 使用率
  - alert: HighCPUUsage
    expr: 100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "CPU 使用率过高"
      description: "实例 {{ $labels.instance }} CPU 使用率超过 80%，当前值: {{ $value }}%"

  # 高内存使用率
  - alert: HighMemoryUsage
    expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 85
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "内存使用率过高"
      description: "实例 {{ $labels.instance }} 内存使用率超过 85%，当前值: {{ $value }}%"

  # 磁盘空间不足
  - alert: DiskSpaceLow
    expr: (node_filesystem_avail_bytes{fstype!~"tmpfs|overlay"} / node_filesystem_size_bytes{fstype!~"tmpfs|overlay"}) * 100 < 15
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "磁盘空间不足"
      description: "实例 {{ $labels.instance }} 磁盘 {{ $labels.mountpoint }} 可用空间不足 15%"

  # HTTP 错误率
  - alert: HighErrorRate
    expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100 > 5
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "HTTP 错误率过高"
      description: "5xx 错误率超过 5%，当前值: {{ $value }}%"
```

### 14.3 Grafana Dashboard

导入 JSON Dashboard 或创建自定义面板：

```json
{
  "dashboard": {
    "title": "Yunwei Monitoring",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ]
      }
    ]
  }
}
```

---

## 十五、日志管理

### 15.1 日志轮转配置

创建 `/etc/logrotate.d/yunwei`：

```
/var/log/yunwei/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0640 yunwei yunwei
    sharedscripts
    postrotate
        systemctl reload yunwei > /dev/null 2>&1 || true
    endscript
}
```

### 15.2 ELK 集成（可选）

```yaml
# Filebeat 配置
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/yunwei/*.log
  fields:
    app: yunwei
  fields_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "yunwei-%{+yyyy.MM.dd}"

setup.kibana:
  host: "kibana:5601"
```

### 15.3 日志格式

使用 JSON 格式日志：

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "info",
  "msg": "Request processed",
  "request_id": "abc123",
  "method": "GET",
  "path": "/api/v1/servers",
  "status": 200,
  "latency": 0.023,
  "client_ip": "192.168.1.100",
  "user_id": "user123"
}
```

---

## 十六、备份恢复

### 16.1 数据库备份脚本

创建 `/opt/yunwei/scripts/backup.sh`：

```bash
#!/bin/bash
# 数据库备份脚本

set -e

# 配置
DB_HOST="localhost"
DB_PORT="3306"
DB_USER="yunwei"
DB_PASS="your_password"
DB_NAME="yunwei"
BACKUP_DIR="/data/backups/mysql"
RETENTION_DAYS=30

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份文件名
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${DATE}.sql.gz"

# 执行备份
mysqldump -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASS \
  --single-transaction \
  --routines \
  --triggers \
  --events \
  $DB_NAME | gzip > $BACKUP_FILE

# 检查备份是否成功
if [ $? -eq 0 ]; then
    echo "[$(date)] Backup completed: $BACKUP_FILE"
    # 记录到日志
    logger "Yunwei MySQL backup completed: $BACKUP_FILE"
else
    echo "[$(date)] Backup failed!"
    logger "Yunwei MySQL backup failed!"
    exit 1
fi

# 删除旧备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete

echo "[$(date)] Old backups cleaned"
```

### 16.2 定时备份

```bash
# 添加定时任务
crontab -e

# 每天凌晨 2 点执行备份
0 2 * * * /opt/yunwei/scripts/backup.sh >> /var/log/yunwei/backup.log 2>&1
```

### 16.3 数据恢复

```bash
# 解压并恢复
gunzip < /data/backups/mysql/yunwei_20240115_020000.sql.gz | mysql -u yunwei -p yunwei

# 或分步执行
gunzip yunwei_20240115_020000.sql.gz
mysql -u yunwei -p yunwei < yunwei_20240115_020000.sql
```

### 16.4 Redis 备份

```bash
# 手动触发 RDB 快照
redis-cli -a your_password BGSAVE

# 复制 RDB 文件
cp /var/lib/redis/dump.rdb /data/backups/redis/dump_$(date +%Y%m%d).rdb

# 恢复
# 停止 Redis
systemctl stop redis
# 复制备份文件
cp /data/backups/redis/dump_20240115.rdb /var/lib/redis/dump.rdb
# 启动 Redis
systemctl start redis
```

---

## 十七、安全加固

### 17.1 防火墙配置

```bash
# CentOS/RHEL (firewalld)
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=50051/tcp
sudo firewall-cmd --reload

# Ubuntu (ufw)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8080/tcp
sudo ufw allow 50051/tcp
sudo ufw enable
```

### 17.2 SSH 加固

编辑 `/etc/ssh/sshd_config`：

```ini
# 禁用 root 登录
PermitRootLogin no

# 使用密钥认证
PubkeyAuthentication yes
PasswordAuthentication no

# 更改默认端口
Port 2222

# 限制登录用户
AllowUsers yunwei admin

# 登录尝试限制
MaxAuthTries 3
MaxSessions 5

# 空闲超时
ClientAliveInterval 300
ClientAliveCountMax 2
```

### 17.3 Fail2ban 配置

```bash
# 安装 Fail2ban
sudo apt install -y fail2ban  # Ubuntu
sudo yum install -y fail2ban  # CentOS

# 创建配置
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3

[yunwei]
enabled = true
port = 8080,443
filter = yunwei
logpath = /var/log/yunwei/app.log
maxretry = 10
EOF

# 创建应用过滤器
cat > /etc/fail2ban/filter.d/yunwei.conf << 'EOF'
[Definition]
failregex = ^.*"status":\s*(401|403).*"client_ip":\s*"<HOST>".*$
ignoreregex =
EOF

# 启动服务
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### 17.4 安全扫描

```bash
# 安装安全扫描工具
sudo apt install -y lynis clamav rkhunter

# 运行安全审计
sudo lynis audit system

# 扫描恶意软件
sudo freshclam
sudo clamscan -r /opt/yunwei

# Rootkit 检测
sudo rkhunter --update
sudo rkhunter --check
```

---

## 十八、常见问题 FAQ

### 18.1 数据库连接问题

**问题**: `Error: dial tcp 127.0.0.1:3306: connect: connection refused`

**解决方案**:
```bash
# 检查 MySQL 服务状态
sudo systemctl status mysql

# 检查端口监听
sudo netstat -tlnp | grep 3306

# 检查防火墙
sudo firewall-cmd --list-ports

# 测试连接
mysql -h 127.0.0.1 -u yunwei -p
```

### 18.2 Redis 连接问题

**问题**: `redis: connection refused`

**解决方案**:
```bash
# 检查 Redis 服务
sudo systemctl status redis

# 检查配置
redis-cli ping

# 检查密码
redis-cli -a your_password ping

# 检查绑定地址
grep "bind" /etc/redis/redis.conf
```

### 18.3 权限问题

**问题**: `permission denied`

**解决方案**:
```bash
# 检查文件权限
ls -la /opt/yunwei/

# 修复权限
sudo chown -R yunwei:yunwei /opt/yunwei
sudo chmod -R 755 /opt/yunwei/bin
sudo chmod 600 /opt/yunwei/config/config.yaml
```

### 18.4 内存不足

**问题**: `out of memory`

**解决方案**:
```bash
# 检查内存使用
free -h

# 检查进程内存
ps aux --sort=-%mem | head -10

# 增加交换空间
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# 添加到 fstab
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

### 18.5 WebSocket 连接失败

**问题**: WebSocket 连接失败或频繁断开

**解决方案**:
```nginx
# Nginx 配置
location /ws {
    proxy_pass http://backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 3600s;
    proxy_send_timeout 3600s;
}
```

### 18.6 SSL 证书问题

**问题**: SSL 证书过期或无效

**解决方案**:
```bash
# 检查证书有效期
openssl x509 -enddate -noout -in /etc/nginx/ssl/server.crt

# 续期 Let's Encrypt 证书
sudo certbot renew

# 重启 Nginx
sudo systemctl reload nginx
```

---

## 十九、故障排查

### 19.1 服务无法启动

```bash
# 查看服务状态
sudo systemctl status yunwei

# 查看详细日志
sudo journalctl -u yunwei -n 100 --no-pager

# 检查端口占用
sudo netstat -tlnp | grep 8080

# 手动启动测试
cd /opt/yunwei && ./bin/yunwei-server
```

### 19.2 API 响应慢

```bash
# 检查系统负载
top
htop

# 检查数据库慢查询
mysql -e "SHOW PROCESSLIST;"

# 检查 Redis 延迟
redis-cli --latency

# 启用 API 日志分析
grep "latency" /var/log/yunwei/app.log | tail -100
```

### 19.3 数据库性能问题

```sql
-- 查看慢查询
SHOW VARIABLES LIKE 'slow_query%';
SELECT * FROM mysql.slow_log ORDER BY start_time DESC LIMIT 10;

-- 查看锁等待
SHOW ENGINE INNODB STATUS\G

-- 查看表大小
SELECT 
    table_name,
    ROUND((data_length + index_length) / 1024 / 1024, 2) AS size_mb
FROM information_schema.tables
WHERE table_schema = 'yunwei'
ORDER BY size_mb DESC;

-- 优化表
ANALYZE TABLE servers;
OPTIMIZE TABLE server_metrics;
```

### 19.4 网络问题排查

```bash
# 检查网络连接
ping -c 5 google.com
traceroute google.com

# 检查 DNS 解析
nslookup yunwei.example.com
dig yunwei.example.com

# 检查端口连通性
telnet localhost 8080
nc -zv localhost 8080

# 抓包分析
sudo tcpdump -i eth0 port 8080 -w capture.pcap
```

### 19.5 日志分析

```bash
# 查看错误日志
grep -i error /var/log/yunwei/app.log | tail -50

# 统计错误类型
grep -o '"error":"[^"]*"' /var/log/yunwei/app.log | sort | uniq -c | sort -rn

# 分析请求延迟
awk -F'"latency":' '{print $2}' /var/log/yunwei/app.log | awk -F',' '{sum+=$1; count++} END {print "avg:", sum/count}'

# 实时监控日志
tail -f /var/log/yunwei/app.log | grep --color=auto "error\|warn"
```

### 19.6 紧急恢复流程

```bash
# 1. 停止服务
sudo systemctl stop yunwei

# 2. 检查数据完整性
mysqlcheck --all-databases -u root -p

# 3. 恢复最近备份
gunzip < /data/backups/mysql/yunwei_latest.sql.gz | mysql -u yunwei -p yunwei

# 4. 清理日志
> /var/log/yunwei/app.log

# 5. 重启服务
sudo systemctl start yunwei

# 6. 验证服务
curl http://localhost:8080/health
```

---

## 联系支持

- **项目地址**: https://github.com/fredphp/yunwei
- **问题反馈**: https://github.com/fredphp/yunwei/issues
- **文档更新**: 查看 `document.md` 获取功能文档

---

**部署文档版本**: v1.0.0  
**最后更新**: 2024年
