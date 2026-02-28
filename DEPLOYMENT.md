# AI 自动化运维管理系统 - 部署文档

## 目录
- [环境要求](#环境要求)
- [配置说明](#配置说明)
- [快速启动](#快速启动)
- [详细步骤](#详细步骤)
- [常见问题](#常见问题)

---

## 环境要求

| 组件 | 版本要求 | 说明 |
|------|---------|------|
| Go | >= 1.20 | 后端运行环境 |
| Node.js | >= 18.0 | 前端运行环境 |
| MySQL | >= 5.7 / 8.0 | 数据库 |
| Redis | >= 6.0 | 缓存（可选） |
| npm/pnpm | 最新版 | 前端包管理器 |

---

## 配置说明

### 1. 后端配置 (server/config/config.yaml)

```yaml
# ==================== 系统配置 ====================
system:
  port: 8080                    # 后端服务端口
  grpc-port: 50051              # gRPC 服务端口
  env: develop                  # 环境: develop, test, production
  name: yunwei                  # 服务名称

# ==================== MySQL 数据库配置 ====================
mysql:
  host: 127.0.0.1               # 数据库地址
  port: 3306                    # 数据库端口
  username: root                # 数据库用户名
  password: your_password       # 数据库密码 ⚠️ 请修改
  database: yunwei              # 数据库名称
  max-idle-conns: 10            # 最大空闲连接数
  max-open-conns: 100           # 最大打开连接数

# ==================== Redis 配置（可选） ====================
redis:
  host: 127.0.0.1
  port: 6379
  password: ""                  # Redis 密码，没有则留空
  db: 0

# ==================== JWT 配置 ====================
jwt:
  signing-key: yunwei-secret-key-change-me  # JWT 密钥 ⚠️ 请修改为复杂字符串
  expires-time: 24h             # Token 过期时间
  issuer: yunwei                # 签发者

# ==================== AI 配置 ====================
ai:
  enabled: true                 # 是否启用 AI 功能
  api-key: ""                   # AI API Key（如智谱 GLM）
  base-url: "https://open.bigmodel.cn/api/paas/v4"
  model: "glm-4"                # 使用的模型
  max-tokens: 4096              # 最大 Token 数
  temperature: 0.7              # 温度参数
  auto-execute: false           # 是否自动执行低风险命令

# ==================== 安全配置 ====================
security:
  enable-whitelist: true        # 启用白名单
  enable-blacklist: true        # 启用黑名单
  require-approval: true        # 高危命令需要审批
  audit-enabled: true           # 启用审计日志
  audit-retention-days: 90      # 审计日志保留天数
```

### 2. 前端配置 (web/.env.development / .env.production)

```bash
# 开发环境配置文件: web/.env.development
VITE_API_BASE_URL=http://localhost:8080/api/v1

# 生产环境配置文件: web/.env.production
VITE_API_BASE_URL=/api/v1
```

---

## 快速启动

### 一键启动脚本 (Linux/Mac)

```bash
#!/bin/bash

# 1. 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS yunwei DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 2. 启动后端
cd server
go mod tidy
go run main.go &

# 3. 启动前端
cd ../web
npm install
npm run dev &
```

---

## 详细步骤

### 第一步：安装依赖环境

#### 1.1 安装 Go
```bash
# Ubuntu/Debian
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Mac
brew install go

# 验证安装
go version
```

#### 1.2 安装 Node.js
```bash
# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Mac
brew install node

# 验证安装
node -v
npm -v
```

#### 1.3 安装 MySQL
```bash
# Ubuntu/Debian
sudo apt install mysql-server
sudo mysql_secure_installation

# Mac
brew install mysql
brew services start mysql

# 创建数据库
mysql -u root -p
```

```sql
CREATE DATABASE IF NOT EXISTS yunwei DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'yunwei'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON yunwei.* TO 'yunwei'@'localhost';
FLUSH PRIVILEGES;
```

### 第二步：克隆项目

```bash
git clone https://github.com/fredphp/yunwei.git
cd yunwei
```

### 第三步：配置后端

#### 3.1 修改配置文件
```bash
cd server
cp config/config.yaml.example config/config.yaml  # 如果有示例文件
# 或直接编辑 config/config.yaml
```

**重要配置项修改：**
```yaml
mysql:
  host: 你的MySQL地址
  port: 3306
  username: 你的用户名
  password: 你的密码
  database: yunwei

jwt:
  signing-key: 请修改为复杂的随机字符串
```

#### 3.2 安装 Go 依赖
```bash
cd server
go mod tidy
go mod download
```

#### 3.3 初始化数据库
数据库表结构会在首次启动时自动创建，无需手动执行 SQL。

如需手动执行迁移：
```bash
# 连接 MySQL
mysql -u root -p yunwei < sql/init.sql
```

#### 3.4 启动后端服务
```bash
# 开发环境
go run main.go

# 生产环境
go build -o yunwei-server
./yunwei-server
```

**启动成功提示：**
```
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     AI 自动化运维管理系统 启动成功!                       ║
║                                                           ║
║     HTTP:  http://localhost:8080                          ║
║     gRPC:  localhost:50051                                ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝

===========================================
  超级管理员账号创建成功!
  用户名: admin
  密码: admin123
===========================================
```

### 第四步：配置前端

#### 4.1 安装依赖
```bash
cd web
npm install
# 或使用 pnpm
pnpm install
```

#### 4.2 配置后端地址
创建 `web/.env.development`：
```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

创建 `web/.env.production`：
```bash
VITE_API_BASE_URL=/api/v1
```

#### 4.3 启动前端开发服务器
```bash
npm run dev
```

访问：http://localhost:5173

### 第五步：登录系统

1. 打开浏览器访问 http://localhost:5173
2. 使用默认管理员账号登录：
   - **用户名**: `admin`
   - **密码**: `admin123`
3. ⚠️ 登录后请立即修改密码！

---

## 生产环境部署

### 方式一：Docker 部署

#### 1. 构建镜像
```bash
# 后端
cd server
docker build -t yunwei-server:latest .

# 前端
cd web
docker build -t yunwei-web:latest .
```

#### 2. 使用 docker-compose
```yaml
# docker-compose.yml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: your_root_password
      MYSQL_DATABASE: yunwei
      MYSQL_USER: yunwei
      MYSQL_PASSWORD: your_password
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  server:
    image: yunwei-server:latest
    depends_on:
      - mysql
      - redis
    environment:
      - MYSQL_HOST=mysql
      - MYSQL_USER=yunwei
      - MYSQL_PASSWORD=your_password
      - MYSQL_DATABASE=yunwei
      - REDIS_HOST=redis
    ports:
      - "8080:8080"
      - "50051:50051"

  web:
    image: yunwei-web:latest
    depends_on:
      - server
    ports:
      - "80:80"

volumes:
  mysql_data:
```

```bash
docker-compose up -d
```

### 方式二：手动部署

#### 1. 后端部署
```bash
# 编译
cd server
GOOS=linux GOARCH=amd64 go build -o yunwei-server

# 使用 systemd 管理
sudo cp yunwei-server /usr/local/bin/
sudo cp deploy/yunwei.service /etc/systemd/system/

sudo systemctl daemon-reload
sudo systemctl enable yunwei
sudo systemctl start yunwei
```

#### 2. 前端部署
```bash
cd web
npm run build

# 使用 Nginx 托管
sudo cp -r dist/* /var/www/html/
sudo cp deploy/nginx.conf /etc/nginx/sites-available/yunwei
sudo ln -s /etc/nginx/sites-available/yunwei /etc/nginx/sites-enabled/
sudo nginx -s reload
```

---

## Nginx 配置示例

```nginx
# /etc/nginx/sites-available/yunwei
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    location / {
        root /var/www/html;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # 后端 API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # WebSocket 代理
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

---

## 常见问题

### 1. 数据库连接失败
```
错误: 数据库连接失败: Error 1045 (28000): Access denied for user 'root'@'localhost'
```
**解决方案**：
- 检查 MySQL 服务是否启动
- 确认用户名密码正确
- 检查用户是否有数据库访问权限

### 2. 前端无法连接后端
```
Network Error / CORS Error
```
**解决方案**：
- 确认后端服务已启动
- 检查 `.env.development` 中的 API 地址
- 检查后端 CORS 配置

### 3. JWT Token 无效
```
token signature is invalid
```
**解决方案**：
- 确保 `jwt.signing-key` 配置正确
- 清除浏览器缓存重新登录

### 4. 端口被占用
```
bind: address already in use
```
**解决方案**：
```bash
# 查看端口占用
lsof -i :8080
# 或
netstat -tlnp | grep 8080

# 杀死进程
kill -9 <PID>
```

### 5. Go 依赖下载失败
**解决方案**：
```bash
# 设置代理
go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy
```

---

## 目录结构

```
yunwei/
├── server/                 # 后端代码
│   ├── api/               # API 接口
│   │   └── v1/            # v1 版本 API
│   ├── config/            # 配置文件
│   │   └── config.yaml    # 主配置文件
│   ├── global/            # 全局变量和初始化
│   ├── middleware/        # 中间件
│   ├── model/             # 数据模型
│   ├── router/            # 路由
│   ├── service/           # 业务逻辑
│   ├── migrations/        # 数据库迁移
│   └── main.go            # 入口文件
├── web/                   # 前端代码
│   ├── src/
│   │   ├── api/          # API 请求
│   │   ├── components/   # 组件
│   │   ├── views/        # 页面
│   │   ├── router/       # 路由
│   │   ├── store/        # 状态管理
│   │   └── utils/        # 工具函数
│   ├── public/           # 静态资源
│   └── package.json
├── sql/                   # SQL 脚本
│   └── init.sql          # 初始化 SQL
├── deploy/               # 部署配置
│   ├── nginx.conf
│   └── supervisor.conf
└── docker-compose.yml    # Docker 编排
```

---

## 技术支持

- GitHub Issues: https://github.com/fredphp/yunwei/issues
- 默认管理员: admin / admin123
