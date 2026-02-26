# AI 自动化运维管理系统

## 项目概述

本项目是一个基于 Go + Gin + Next.js 的 AI 自动化运维管理系统，提供服务器管理、Kubernetes管理、灰度发布、负载均衡、证书管理、CDN管理、智能部署、任务调度、Agent管理、高可用、灾备备份、成本控制、多租户等完整的运维解决方案。

### 技术栈

| 层级 | 技术 |
|------|------|
| 后端框架 | Go 1.21+ / Gin |
| 前端框架 | Next.js 16 / React 18 / TypeScript |
| 数据库 | MySQL / PostgreSQL / SQLite |
| ORM | GORM |
| 缓存 | Redis |
| 消息队列 | 内置队列系统 |
| 容器化 | Docker / Kubernetes |
| 代码仓库 | https://github.com/fredphp/yunwei |

---

## 功能模块

### 一、服务器管理模块

#### 1.1 服务器基础管理
| 功能 | API | 说明 |
|------|-----|------|
| 服务器列表 | GET /api/v1/servers | 支持分页、搜索、过滤 |
| 服务器详情 | GET /api/v1/servers/:id | 获取服务器详细信息 |
| 添加服务器 | POST /api/v1/servers | 添加新服务器 |
| 更新服务器 | PUT /api/v1/servers/:id | 更新服务器信息 |
| 删除服务器 | DELETE /api/v1/servers/:id | 删除服务器 |
| SSH测试 | POST /api/v1/ssh/test | 测试SSH连接 |

#### 1.2 服务器监控
| 功能 | API | 说明 |
|------|-----|------|
| 服务器指标 | GET /api/v1/servers/:id/metrics | CPU、内存、磁盘、网络等 |
| 服务器日志 | GET /api/v1/servers/:id/logs | 操作日志查询 |
| Docker容器 | GET /api/v1/servers/:id/containers | 容器列表和状态 |
| 端口信息 | GET /api/v1/servers/:id/ports | 端口占用情况 |
| 刷新状态 | POST /api/v1/servers/:id/refresh | 刷新服务器状态 |

#### 1.3 服务器分组
| 功能 | API | 说明 |
|------|-----|------|
| 分组列表 | GET /api/v1/groups | 获取所有分组 |
| 创建分组 | POST /api/v1/groups | 创建新分组 |
| 删除分组 | DELETE /api/v1/groups/:id | 删除分组 |

---

### 二、Kubernetes 管理模块

#### 2.1 集群管理
| 功能 | API | 说明 |
|------|-----|------|
| 集群列表 | GET /api/v1/kubernetes/clusters | 获取所有集群 |
| 集群详情 | GET /api/v1/kubernetes/clusters/:id | 获取集群详情 |
| 添加集群 | POST /api/v1/kubernetes/clusters | 添加新集群 |
| 更新集群 | PUT /api/v1/kubernetes/clusters/:id | 更新集群配置 |
| 删除集群 | DELETE /api/v1/kubernetes/clusters/:id | 删除集群 |

#### 2.2 部署与扩缩容
| 功能 | API | 说明 |
|------|-----|------|
| Deployment状态 | GET /api/v1/kubernetes/clusters/:clusterId/deployments | 获取部署状态 |
| 扩容历史 | GET /api/v1/kubernetes/scale/history | 扩缩容历史记录 |
| 手动扩缩容 | POST /api/v1/kubernetes/scale/manual | 手动触发扩缩容 |
| 集群分析 | POST /api/v1/kubernetes/clusters/:clusterId/analyze | AI分析集群 |

#### 2.3 HPA 配置
| 功能 | API | 说明 |
|------|-----|------|
| HPA配置列表 | GET /api/v1/kubernetes/hpa | 获取HPA配置 |
| 更新HPA | POST /api/v1/kubernetes/hpa | 更新HPA配置 |

---

### 三、灰度发布模块

#### 3.1 发布管理
| 功能 | API | 说明 |
|------|-----|------|
| 发布列表 | GET /api/v1/canary/releases | 获取所有灰度发布 |
| 发布详情 | GET /api/v1/canary/releases/:id | 获取发布详情 |
| 发布步骤 | GET /api/v1/canary/releases/:id/steps | 获取发布步骤 |
| 创建发布 | POST /api/v1/canary/releases | 创建灰度发布 |

#### 3.2 发布操作
| 功能 | API | 说明 |
|------|-----|------|
| 推进发布 | POST /api/v1/canary/releases/:id/promote | 推进到下一阶段 |
| 完成发布 | POST /api/v1/canary/releases/:id/complete | 完成发布 |
| 回滚发布 | POST /api/v1/canary/releases/:id/rollback | 回滚发布 |
| 暂停发布 | POST /api/v1/canary/releases/:id/pause | 暂停发布 |
| 终止发布 | POST /api/v1/canary/releases/:id/abort | 终止发布 |

#### 3.3 配置管理
| 功能 | API | 说明 |
|------|-----|------|
| 配置列表 | GET /api/v1/canary/configs | 获取灰度配置 |
| 更新配置 | POST /api/v1/canary/configs | 更新配置 |

---

### 四、负载均衡模块

#### 4.1 负载均衡器管理
| 功能 | API | 说明 |
|------|-----|------|
| LB列表 | GET /api/v1/loadbalancer | 获取所有负载均衡器 |
| LB详情 | GET /api/v1/loadbalancer/:id | 获取详情 |
| 添加LB | POST /api/v1/loadbalancer | 添加负载均衡器 |
| 更新LB | PUT /api/v1/loadbalancer/:id | 更新配置 |
| 删除LB | DELETE /api/v1/loadbalancer/:id | 删除负载均衡器 |

#### 4.2 后端服务器管理
| 功能 | API | 说明 |
|------|-----|------|
| 后端列表 | GET /api/v1/loadbalancer/:id/backends | 获取后端服务器 |
| 添加后端 | POST /api/v1/loadbalancer/:id/backends | 添加后端服务器 |
| 更新后端 | PUT /api/v1/loadbalancer/backends/:id | 更新后端配置 |
| 删除后端 | DELETE /api/v1/loadbalancer/backends/:id | 删除后端服务器 |

#### 4.3 优化与操作
| 功能 | API | 说明 |
|------|-----|------|
| 优化LB | POST /api/v1/loadbalancer/:id/optimize | AI优化负载均衡 |
| 自动均衡 | POST /api/v1/loadbalancer/:id/autobalance | 自动负载均衡 |
| 健康检查 | POST /api/v1/loadbalancer/:id/healthcheck | 执行健康检查 |
| 优化历史 | GET /api/v1/loadbalancer/history | 优化历史记录 |
| 算法配置 | GET/POST /api/v1/loadbalancer/algorithm | 负载均衡算法配置 |

---

### 五、证书管理模块

#### 5.1 证书管理
| 功能 | API | 说明 |
|------|-----|------|
| 证书列表 | GET /api/v1/certificates | 获取所有证书 |
| 证书详情 | GET /api/v1/certificates/:id | 获取证书详情 |
| 添加证书 | POST /api/v1/certificates | 添加证书 |
| 更新证书 | PUT /api/v1/certificates/:id | 更新证书 |
| 删除证书 | DELETE /api/v1/certificates/:id | 删除证书 |

#### 5.2 证书操作
| 功能 | API | 说明 |
|------|-----|------|
| 续签证书 | POST /api/v1/certificates/:id/renew | 续签证书 |
| 检查证书 | POST /api/v1/certificates/:id/check | 检查证书状态 |
| 批量检查 | POST /api/v1/certificates/check-all | 批量检查所有证书 |
| 续签历史 | GET /api/v1/certificates/history | 续签历史记录 |
| 申请新证书 | POST /api/v1/certificates/request | 申请Let's Encrypt证书 |

---

### 六、CDN 管理模块

#### 6.1 域名管理
| 功能 | API | 说明 |
|------|-----|------|
| 域名列表 | GET /api/v1/cdn/domains | 获取所有CDN域名 |
| 域名详情 | GET /api/v1/cdn/domains/:id | 获取域名详情 |
| 添加域名 | POST /api/v1/cdn/domains | 添加CDN域名 |
| 更新域名 | PUT /api/v1/cdn/domains/:id | 更新域名配置 |
| 删除域名 | DELETE /api/v1/cdn/domains/:id | 删除域名 |

#### 6.2 缓存操作
| 功能 | API | 说明 |
|------|-----|------|
| 刷新缓存 | POST /api/v1/cdn/domains/:id/purge | URL刷新 |
| 预热缓存 | POST /api/v1/cdn/domains/:id/preheat | URL预热 |
| 节点状态 | GET /api/v1/cdn/domains/:id/nodes | CDN节点状态 |
| 成本计算 | GET /api/v1/cdn/domains/:id/cost | CDN成本计算 |

#### 6.3 优化与规则
| 功能 | API | 说明 |
|------|-----|------|
| CDN优化 | POST /api/v1/cdn/domains/:id/optimize | AI优化CDN配置 |
| 成本优化 | POST /api/v1/cdn/domains/:id/cost-optimize | 成本优化建议 |
| 缓存规则 | GET /api/v1/cdn/domains/:id/rules | 缓存规则列表 |
| 添加规则 | POST /api/v1/cdn/domains/:id/rules | 添加缓存规则 |
| 更新规则 | PUT /api/v1/cdn/rules/:id | 更新缓存规则 |
| 删除规则 | DELETE /api/v1/cdn/rules/:id | 删除缓存规则 |
| 优化历史 | GET /api/v1/cdn/history | 优化历史记录 |

---

### 七、智能部署模块

#### 7.1 项目分析
| 功能 | API | 说明 |
|------|-----|------|
| 上传项目 | POST /api/v1/deploy/upload | 上传项目文件 |
| 项目分析 | POST /api/v1/deploy/analyze | AI分析项目结构 |
| 项目列表 | GET /api/v1/deploy/projects | 分析项目列表 |
| 项目详情 | GET /api/v1/deploy/projects/:id | 项目分析详情 |

#### 7.2 服务器分析
| 功能 | API | 说明 |
|------|-----|------|
| 服务器分析 | GET /api/v1/deploy/servers/analyze | 分析服务器资源 |
| 服务器能力 | GET /api/v1/deploy/servers/capabilities | 获取服务器能力 |
| 最佳服务器 | POST /api/v1/deploy/servers/find-best | 查找最佳部署服务器 |

#### 7.3 部署方案
| 功能 | API | 说明 |
|------|-----|------|
| 生成方案 | POST /api/v1/deploy/plans | 生成部署方案 |
| 方案列表 | GET /api/v1/deploy/plans | 部署方案列表 |
| 方案详情 | GET /api/v1/deploy/plans/:id | 方案详情 |
| 删除方案 | DELETE /api/v1/deploy/plans/:id | 删除方案 |
| 服务拓扑 | GET /api/v1/deploy/plans/:id/topology | 服务拓扑图 |
| 配置预览 | GET /api/v1/deploy/plans/:id/preview | 预览生成配置 |

#### 7.4 部署执行
| 功能 | API | 说明 |
|------|-----|------|
| 执行部署 | POST /api/v1/deploy/plans/:id/execute | 执行部署方案 |
| 任务列表 | GET /api/v1/deploy/tasks | 部署任务列表 |
| 任务详情 | GET /api/v1/deploy/tasks/:id | 任务详情 |
| 任务步骤 | GET /api/v1/deploy/tasks/:id/steps | 任务执行步骤 |
| 暂停部署 | POST /api/v1/deploy/tasks/:id/pause | 暂停部署 |
| 恢复部署 | POST /api/v1/deploy/tasks/:id/resume | 恢复部署 |
| 回滚部署 | POST /api/v1/deploy/tasks/:id/rollback | 回滚部署 |

---

### 八、任务调度中心

#### 8.1 仪表盘
| 功能 | API | 说明 |
|------|-----|------|
| 仪表盘 | GET /api/v1/scheduler/dashboard | 调度中心概览 |

#### 8.2 任务管理
| 功能 | API | 说明 |
|------|-----|------|
| 提交任务 | POST /api/v1/scheduler/tasks | 提交新任务 |
| 任务选项 | POST /api/v1/scheduler/tasks/options | 带选项提交任务 |
| 任务列表 | GET /api/v1/scheduler/tasks | 任务列表 |
| 任务详情 | GET /api/v1/scheduler/tasks/:id | 任务详情 |
| 取消任务 | POST /api/v1/scheduler/tasks/:id/cancel | 取消任务 |
| 重试任务 | POST /api/v1/scheduler/tasks/:id/retry | 重试任务 |
| 回滚任务 | POST /api/v1/scheduler/tasks/:id/rollback | 回滚任务 |
| 执行记录 | GET /api/v1/scheduler/tasks/:id/executions | 任务执行记录 |

#### 8.3 批量任务
| 功能 | API | 说明 |
|------|-----|------|
| 提交批量任务 | POST /api/v1/scheduler/batches | 提交批量任务 |
| 批量任务列表 | GET /api/v1/scheduler/batches | 批量任务列表 |
| 批量任务详情 | GET /api/v1/scheduler/batches/:id | 批量任务详情 |
| 批量任务子任务 | GET /api/v1/scheduler/batches/:id/tasks | 子任务列表 |

#### 8.4 定时任务
| 功能 | API | 说明 |
|------|-----|------|
| 创建定时任务 | POST /api/v1/scheduler/cron | 创建CronJob |
| 定时任务列表 | GET /api/v1/scheduler/cron | CronJob列表 |
| 定时任务详情 | GET /api/v1/scheduler/cron/:id | CronJob详情 |
| 更新定时任务 | PUT /api/v1/scheduler/cron/:id | 更新CronJob |
| 删除定时任务 | DELETE /api/v1/scheduler/cron/:id | 删除CronJob |
| 手动触发 | POST /api/v1/scheduler/cron/:id/trigger | 手动触发执行 |
| 执行记录 | GET /api/v1/scheduler/cron/:id/executions | 执行记录 |

#### 8.5 队列与Worker
| 功能 | API | 说明 |
|------|-----|------|
| 队列列表 | GET /api/v1/scheduler/queues | 队列列表 |
| 队列统计 | GET /api/v1/scheduler/queues/stats | 队列统计信息 |
| Worker列表 | GET /api/v1/scheduler/workers | Worker列表 |
| 扩缩容Worker | POST /api/v1/scheduler/workers/scale | 扩缩容Worker |

#### 8.6 模板管理
| 功能 | API | 说明 |
|------|-----|------|
| 创建模板 | POST /api/v1/scheduler/templates | 创建任务模板 |
| 模板列表 | GET /api/v1/scheduler/templates | 模板列表 |
| 从模板提交 | POST /api/v1/scheduler/templates/submit | 从模板提交任务 |

---

### 九、Agent 管理模块

#### 9.1 Agent 管理
| 功能 | API | 说明 |
|------|-----|------|
| Agent列表 | GET /api/v1/agents | 获取所有Agent |
| Agent统计 | GET /api/v1/agents/stats | Agent统计数据 |
| Agent详情 | GET /api/v1/agents/:id | Agent详情 |
| 更新Agent | PUT /api/v1/agents/:id | 更新Agent配置 |
| 删除Agent | DELETE /api/v1/agents/:id | 删除Agent |
| 禁用Agent | POST /api/v1/agents/:id/disable | 禁用Agent |
| 启用Agent | POST /api/v1/agents/:id/enable | 启用Agent |
| Agent配置 | GET /api/v1/agents/:id/config | 获取Agent配置 |
| 检查升级 | GET /api/v1/agents/:id/check-upgrade | 检查可用升级 |
| 心跳记录 | GET /api/v1/agents/:id/heartbeats | 心跳记录历史 |
| 恢复记录 | GET /api/v1/agents/:id/recovers | 自动恢复记录 |
| 批量操作 | POST /api/v1/agents/batch | 批量操作Agent |

#### 9.2 版本管理
| 功能 | API | 说明 |
|------|-----|------|
| 版本列表 | GET /api/v1/agents/versions | Agent版本列表 |
| 版本统计 | GET /api/v1/agents/versions/stats | 版本分布统计 |
| 版本详情 | GET /api/v1/agents/versions/:id | 版本详情 |
| 创建版本 | POST /api/v1/agents/versions | 创建新版本 |
| 更新版本 | PUT /api/v1/agents/versions/:id | 更新版本信息 |
| 删除版本 | DELETE /api/v1/agents/versions/:id | 删除版本 |

#### 9.3 升级管理
| 功能 | API | 说明 |
|------|-----|------|
| 升级任务列表 | GET /api/v1/agents/upgrades | 升级任务列表 |
| 升级统计 | GET /api/v1/agents/upgrades/stats | 升级统计 |
| 创建升级任务 | POST /api/v1/agents/upgrades | 创建升级任务 |
| 批量升级 | POST /api/v1/agents/upgrades/batch | 批量创建升级任务 |
| 升级任务详情 | GET /api/v1/agents/upgrades/:id | 升级任务详情 |
| 执行升级 | POST /api/v1/agents/upgrades/:id/execute | 执行升级 |
| 取消升级 | POST /api/v1/agents/upgrades/:id/cancel | 取消升级 |
| 回滚升级 | POST /api/v1/agents/upgrades/:id/rollback | 回滚升级 |

#### 9.4 灰度发布
| 功能 | API | 说明 |
|------|-----|------|
| 灰度策略列表 | GET /api/v1/agents/gray | 灰度策略列表 |
| 灰度监控统计 | GET /api/v1/agents/gray/stats | 监控统计 |
| 灰度策略详情 | GET /api/v1/agents/gray/:id | 策略详情 |
| 创建灰度策略 | POST /api/v1/agents/gray | 创建灰度策略 |
| 启动灰度 | POST /api/v1/agents/gray/:id/start | 启动灰度发布 |
| 暂停灰度 | POST /api/v1/agents/gray/:id/pause | 暂停灰度 |
| 恢复灰度 | POST /api/v1/agents/gray/:id/resume | 恢复灰度 |
| 取消灰度 | POST /api/v1/agents/gray/:id/cancel | 取消灰度 |
| 灰度进度 | GET /api/v1/agents/gray/:id/progress | 灰度进度 |

#### 9.5 离线监控
| 功能 | API | 说明 |
|------|-----|------|
| 监控统计 | GET /api/v1/agents/monitor/stats | 监控统计数据 |
| 离线Agent | GET /api/v1/agents/monitor/offline | 离线Agent列表 |

---

### 十、高可用(HA)管理模块

#### 10.1 集群状态
| 功能 | API | 说明 |
|------|-----|------|
| 集群统计 | GET /api/v1/ha/stats | 集群状态统计 |
| 节点列表 | GET /api/v1/ha/nodes | 集群节点列表 |
| 节点详情 | GET /api/v1/ha/nodes/:id | 节点详情 |
| 启用节点 | POST /api/v1/ha/nodes/:id/enable | 启用节点 |
| 禁用节点 | POST /api/v1/ha/nodes/:id/disable | 禁用节点 |
| 节点指标 | GET /api/v1/ha/nodes/:id/metrics | 节点监控指标 |

#### 10.2 Leader 选举
| 功能 | API | 说明 |
|------|-----|------|
| Leader状态 | GET /api/v1/ha/leader | 当前Leader状态 |
| Leader让位 | POST /api/v1/ha/leader/resign | Leader主动让位 |
| 强制Leader | POST /api/v1/ha/leader/force | 强制指定Leader |
| 选举记录 | GET /api/v1/ha/leader/records | 选举历史记录 |

#### 10.3 分布式锁
| 功能 | API | 说明 |
|------|-----|------|
| 锁列表 | GET /api/v1/ha/locks | 分布式锁列表 |
| 锁记录 | GET /api/v1/ha/locks/records | 锁操作记录 |
| 锁详情 | GET /api/v1/ha/locks/:key | 获取锁详情 |
| 强制释放锁 | POST /api/v1/ha/locks/:key/release | 强制释放锁 |

#### 10.4 会话管理
| 功能 | API | 说明 |
|------|-----|------|
| 会话列表 | GET /api/v1/ha/sessions | 会话列表 |
| 会话统计 | GET /api/v1/ha/sessions/stats | 会话统计 |
| 删除会话 | DELETE /api/v1/ha/sessions/:id | 删除会话 |

#### 10.5 配置与故障转移
| 功能 | API | 说明 |
|------|-----|------|
| 配置列表 | GET /api/v1/ha/configs | HA配置列表 |
| 获取配置 | GET /api/v1/ha/config | 当前HA配置 |
| 更新配置 | PUT /api/v1/ha/config | 更新HA配置 |
| 创建配置 | POST /api/v1/ha/configs | 创建HA配置 |
| 删除配置 | DELETE /api/v1/ha/configs/:id | 删除HA配置 |
| 故障转移记录 | GET /api/v1/ha/failover | 故障转移历史 |
| 触发故障转移 | POST /api/v1/ha/failover/trigger | 手动触发故障转移 |
| 集群事件 | GET /api/v1/ha/events | 集群事件列表 |
| 运行中任务 | GET /api/v1/ha/tasks/running | 运行中任务列表 |

---

### 十一、灾备与备份管理模块

#### 11.1 备份策略管理
| 功能 | 说明 |
|------|------|
| 策略CRUD | 创建、查询、更新、删除备份策略 |
| 备份类型 | 数据库备份、文件备份、快照备份、全量备份 |
| 备份源 | MySQL、PostgreSQL、MongoDB、Redis、文件/目录 |
| 调度类型 | Cron表达式、间隔执行、手动触发 |
| 存储类型 | 本地存储、S3、OSS、NFS、FTP |
| 压缩加密 | 支持gzip/zstd压缩、AES加密 |
| 脚本支持 | 备份前/后脚本执行 |

#### 11.2 备份记录管理
| 功能 | 说明 |
|------|------|
| 备份类型 | 全量备份、增量备份、差异备份 |
| 触发方式 | 手动触发、定时调度、自动触发 |
| 状态跟踪 | pending、running、success、failed、cancelled |
| 文件管理 | 文件名、路径、大小、校验和 |
| 存储信息 | 存储类型、存储路径、存储节点 |

#### 11.3 恢复管理
| 功能 | 说明 |
|------|------|
| 恢复类型 | 全量恢复、部分恢复、时间点恢复 |
| 恢复进度 | 实时进度跟踪、文件级恢复统计 |
| 恢复验证 | 完整性验证、一致性检查 |
| 脚本支持 | 恢复前/后脚本执行 |

#### 11.4 快照管理
| 功能 | 说明 |
|------|------|
| 快照类型 | 全量快照、增量快照 |
| 目标类型 | 虚拟机、卷、文件系统、数据库 |
| 一致性 | 静默快照、一致性快照 |
| 保留策略 | 按天数、按数量、按容量限制 |

#### 11.5 灾备演练
| 功能 | 说明 |
|------|------|
| 演练类型 | 桌面演练、部分演练、全面演练 |
| 演练场景 | 服务器故障、数据库故障、数据损坏、勒索病毒、自然灾害 |
| RTO/RPO验证 | 目标值与实际值对比 |
| 演练报告 | 自动生成演练报告 |

#### 11.6 恢复验证
| 功能 | 说明 |
|------|------|
| 验证类型 | 完整性验证、一致性验证、可恢复性验证 |
| 验证项目 | 校验和、文件数量、文件大小、目录结构、内容抽样 |
| 评分机制 | 0-100分验证评分 |

#### 11.7 恢复脚本
| 功能 | 说明 |
|------|------|
| 脚本类型 | 备份脚本、恢复脚本、验证脚本、演练脚本 |
| 脚本语言 | Bash、Python、Go |
| 参数配置 | 支持参数化脚本执行 |
| 执行统计 | 执行次数、成功率、平均耗时 |

---

### 十二、云成本控制模块

#### 12.1 成本统计
| 功能 | 说明 |
|------|------|
| 云厂商支持 | AWS、Azure、GCP、阿里云、腾讯云、华为云 |
| 资源类型 | EC2、RDS、S3、EBS、Lambda、EKS、ECS等 |
| 计费模式 | 按需、预留实例、Spot、Savings Plan |
| 成本维度 | 按项目、部门、环境、成本中心统计 |
| 环比分析 | 环比变化、变化率计算 |

#### 12.2 成本预测
| 功能 | 说明 |
|------|------|
| 预测模型 | 线性回归、ARIMA、Prophet、机器学习 |
| 置信区间 | 预测值上下限 |
| 趋势分析 | 上升、下降、稳定趋势判断 |
| 季节性分析 | 周、月、年季节性模式 |

#### 12.3 资源浪费检测
| 浪费类型 | 说明 |
|---------|------|
| 闲置资源 | 长期低利用率资源 |
| 过大配置 | 配置远超实际需求 |
| 孤立资源 | 未挂载的磁盘、未关联的资源 |
| 过期资源 | 已过期但仍产生的费用 |

#### 12.4 闲置资源识别
| 功能 | 说明 |
|------|------|
| 指标监控 | CPU、内存、网络、磁盘IOPS、连接数 |
| 闲置评分 | 0-100分闲置评分 |
| 成本统计 | 小时/日/月成本、累计闲置成本 |
| 建议操作 | 终止、调整配置、快照后释放 |

#### 12.5 预算管理
| 功能 | 说明 |
|------|------|
| 预算范围 | 全局、按云厂商、按项目、按部门 |
| 预算周期 | 月度、季度、年度 |
| 告警阈值 | 多级阈值告警(如50%、80%、100%) |
| 预测支出 | 基于当前趋势预测月末支出 |

#### 12.6 成本优化建议
| 优化类型 | 说明 |
|---------|------|
| 配置调整 | 缩减过大配置 |
| 购买方式 | 预留实例、Savings Plan |
| 终止闲置 | 终止闲置资源 |
| 存储优化 | 冷数据归档、删除无用数据 |

#### 12.7 Kubernetes 成本
| 功能 | 说明 |
|------|------|
| 命名空间成本 | 按命名空间统计成本 |
| 工作负载成本 | Deployment、StatefulSet等成本 |
| 资源效率 | CPU/内存请求与实际使用对比 |
| 过配检测 | 过度配置检测与优化建议 |

---

### 十三、多租户系统模块

#### 13.1 租户管理
| 功能 | 说明 |
|------|------|
| 租户CRUD | 创建、查询、更新、删除租户 |
| 租户标识 | 名称、Slug(URL友好)、自定义域名 |
| 租户状态 | active、suspended、deleted |
| 套餐管理 | free、starter、pro、enterprise |
| 联系信息 | 联系人、邮箱、电话、地址 |

#### 13.2 配额管理
| 配额类型 | 说明 |
|---------|------|
| 用户配额 | 最大用户数、最大管理员数 |
| 资源配额 | 最大资源数、最大服务器数、最大数据库数 |
| 监控配额 | 最大监控数、最大告警规则数、指标保留天数 |
| 成本配额 | 最大云账户数、预算限制 |
| 存储配额 | 最大存储空间、最大备份空间 |
| API配额 | 每日API调用限制、Webhook数量 |

#### 13.3 用户成员管理
| 功能 | 说明 |
|------|------|
| 成员CRUD | 添加、查询、更新、删除成员 |
| 角色分配 | 为成员分配角色 |
| 邀请机制 | 邮箱邀请加入租户 |
| 状态管理 | active、inactive、pending |

#### 13.4 角色权限(RBAC)
| 预设角色 | 权限范围 |
|---------|---------|
| owner | 租户所有者，完全控制权 |
| admin | 租户管理员，可管理所有资源 |
| operator | 运维人员，可操作资源 |
| viewer | 只读用户，仅查看 |

#### 13.5 数据隔离
| 隔离层级 | 说明 |
|---------|------|
| 租户隔离 | 不同租户数据完全隔离 |
| 数据隔离 | 数据库级别的租户标识 |
| 资源隔离 | 资源按租户ID隔离 |
| API隔离 | API请求自动注入租户上下文 |

#### 13.6 审计日志
| 功能 | 说明 |
|------|------|
| 操作记录 | 所有操作自动记录 |
| 变更追踪 | 记录变更前后值 |
| 请求信息 | IP地址、User-Agent、RequestID |
| 状态跟踪 | 操作成功/失败状态 |

---

### 十四、权限管理系统(RBAC)

#### 14.1 权限定义

| 模块 | 权限代码 | 权限名称 | 风险等级 |
|------|---------|---------|---------|
| **服务器管理** | server:view | 查看服务器 | 1 |
| | server:add | 添加服务器 | 2 |
| | server:edit | 编辑服务器 | 2 |
| | server:delete | 删除服务器 | 4 |
| | server:ssh | SSH连接 | 3 |
| | server:execute | 执行服务器命令 | 3 |
| | server:analyze | 服务器AI分析 | 2 |
| **Kubernetes** | k8s:view | 查看K8s集群 | 1 |
| | k8s:add | 添加K8s集群 | 3 |
| | k8s:edit | 编辑K8s集群 | 3 |
| | k8s:delete | 删除K8s集群 | 4 |
| | k8s:deploy | K8s部署操作 | 3 |
| | k8s:scale | K8s扩缩容 | 3 |
| **灰度发布** | canary:view | 查看灰度发布 | 1 |
| | canary:add | 创建灰度发布 | 3 |
| | canary:deploy | 执行灰度发布 | 4 |
| | canary:rollback | 灰度回滚 | 4 |
| **负载均衡** | lb:view | 查看负载均衡 | 1 |
| | lb:add | 添加负载均衡 | 2 |
| | lb:edit | 编辑负载均衡 | 2 |
| | lb:delete | 删除负载均衡 | 4 |
| | lb:operate | 负载均衡操作 | 3 |
| | lb:optimize | 负载均衡优化 | 3 |
| **证书管理** | cert:view | 查看证书 | 1 |
| | cert:add | 添加证书 | 2 |
| | cert:edit | 编辑证书 | 2 |
| | cert:delete | 删除证书 | 4 |
| | cert:renew | 续签证书 | 3 |
| **CDN管理** | cdn:view | 查看CDN | 1 |
| | cdn:add | 添加CDN域名 | 2 |
| | cdn:edit | 编辑CDN域名 | 2 |
| | cdn:delete | 删除CDN域名 | 4 |
| | cdn:operate | CDN操作 | 3 |
| | cdn:optimize | CDN优化 | 3 |
| **智能部署** | deploy:view | 查看部署方案 | 1 |
| | deploy:add | 创建部署方案 | 2 |
| | deploy:execute | 执行部署 | 4 |
| | deploy:rollback | 部署回滚 | 4 |
| **任务调度** | scheduler:view | 查看调度任务 | 1 |
| | scheduler:add | 创建调度任务 | 2 |
| | scheduler:operate | 调度任务操作 | 3 |
| | scheduler:trigger | 触发任务执行 | 3 |
| **Agent管理** | agent:view | 查看Agent | 1 |
| | agent:edit | 编辑Agent | 2 |
| | agent:delete | 删除Agent | 4 |
| | agent:operate | Agent操作 | 3 |
| | agent:upgrade | Agent升级 | 3 |
| **高可用** | ha:view | 查看高可用状态 | 1 |
| | ha:operate | 高可用操作 | 4 |
| | ha:failover | 故障转移 | 5 |
| | ha:config | 高可用配置 | 4 |
| **备份管理** | backup:view | 查看备份 | 1 |
| | backup:add | 创建备份 | 2 |
| | backup:execute | 执行备份 | 3 |
| | backup:restore | 恢复备份 | 5 |
| | backup:delete | 删除备份 | 4 |
| **成本控制** | cost:view | 查看成本数据 | 1 |
| | cost:analyze | 成本分析 | 2 |
| | cost:optimize | 成本优化 | 3 |
| | cost:config | 成本配置 | 3 |
| **用户管理** | user:view | 查看用户 | 2 |
| | user:add | 添加用户 | 3 |
| | user:edit | 编辑用户 | 3 |
| | user:delete | 删除用户 | 4 |
| **角色管理** | role:view | 查看角色 | 2 |
| | role:add | 添加角色 | 4 |
| | role:edit | 编辑角色 | 4 |
| | role:delete | 删除角色 | 5 |
| **系统设置** | system:config | 系统配置 | 4 |
| | system:backup | 系统备份 | 3 |
| | system:restore | 系统恢复 | 5 |

#### 14.2 角色权限对照表

| 权限模块 | 超级管理员 | 管理员 | 操作员 | 只读用户 |
|---------|:--------:|:-----:|:-----:|:------:|
| 服务器管理 | 全部 | 增删改查+SSH+执行+分析 | 查看+SSH+执行+分析 | 仅查看 |
| Kubernetes | 全部 | 增改查+部署+扩缩容 | 查看+部署+扩缩容 | 仅查看 |
| 灰度发布 | 全部 | 创建+执行+回滚+配置 | 仅查看 | 仅查看 |
| 负载均衡 | 全部 | 增改查+操作+优化 | 查看+操作 | 仅查看 |
| 证书管理 | 全部 | 增改查+续签+检查 | 查看+检查 | 仅查看 |
| CDN管理 | 全部 | 增改查+操作+优化 | 查看+操作 | 仅查看 |
| 智能部署 | 全部 | 创建+执行+回滚+分析 | 查看+分析 | 仅查看 |
| 任务调度 | 全部 | 创建+操作+触发 | 查看+触发 | 仅查看 |
| Agent管理 | 全部 | 增删改+操作+升级 | 查看+操作 | 仅查看 |
| 高可用 | 全部 | 操作+配置+故障转移 | 仅查看 | 仅查看 |
| 备份管理 | 全部 | 创建+执行+恢复 | 查看+执行 | 仅查看 |
| 成本控制 | 全部 | 分析+优化+配置 | 查看+分析 | 仅查看 |
| 用户管理 | 全部 | 增改查 | 增改查 | 仅查看 |
| 角色管理 | 全部 | 仅查看 | 无权限 | 无权限 |
| 系统设置 | 全部 | 配置+备份 | 无权限 | 无权限 |

#### 14.3 权限管理API
| 功能 | API | 说明 |
|------|-----|------|
| 获取权限列表 | GET /api/v1/permissions | 获取所有权限定义 |
| 获取权限分组 | GET /api/v1/permissions/groups | 获取权限分组 |
| 检查权限 | POST /api/v1/permissions/check | 检查用户权限 |
| 批量检查 | POST /api/v1/permissions/check-batch | 批量检查权限 |
| 获取角色列表 | GET /api/v1/roles | 获取所有角色 |
| 获取角色详情 | GET /api/v1/roles/:id | 获取角色详情 |
| 创建角色 | POST /api/v1/roles | 创建自定义角色 |
| 更新角色 | PUT /api/v1/roles/:id | 更新角色信息 |
| 删除角色 | DELETE /api/v1/roles/:id | 删除自定义角色 |
| 获取角色权限 | GET /api/v1/roles/:id/permissions | 获取角色权限列表 |
| 更新角色权限 | PUT /api/v1/roles/:id/permissions | 更新角色权限 |
| 分配角色 | POST /api/v1/users/assign-role | 为用户分配角色 |
| 撤销角色 | POST /api/v1/users/revoke-role | 撤销用户角色 |
| 获取用户权限 | GET /api/v1/user/permissions | 获取当前用户权限 |

---

### 十五、告警管理模块

#### 15.1 告警管理
| 功能 | API | 说明 |
|------|-----|------|
| 告警列表 | GET /api/v1/alerts | 获取告警列表 |
| 确认告警 | POST /api/v1/alerts/:id/acknowledge | 确认告警 |

#### 15.2 检测规则
| 功能 | API | 说明 |
|------|-----|------|
| 规则列表 | GET /api/v1/rules | 获取检测规则 |
| 更新规则 | PUT /api/v1/rules/:id | 更新检测规则 |

---

### 十六、AI 运维模块

#### 16.1 AI 分析
| 功能 | API | 说明 |
|------|-----|------|
| 服务器分析 | POST /api/v1/servers/:id/analyze | AI分析服务器 |
| 集群分析 | POST /api/v1/kubernetes/clusters/:clusterId/analyze | AI分析K8s集群 |

#### 16.2 AI 决策
| 功能 | API | 说明 |
|------|-----|------|
| 决策列表 | GET /api/v1/decisions | AI决策列表 |
| 批准决策 | POST /api/v1/decisions/:id/approve | 批准AI决策 |
| 拒绝决策 | POST /api/v1/decisions/:id/reject | 拒绝AI决策 |
| 执行决策 | POST /api/v1/decisions/:id/execute | 执行AI决策 |

#### 16.3 自动操作
| 功能 | API | 说明 |
|------|-----|------|
| 操作列表 | GET /api/v1/actions | 自动操作列表 |
| 执行操作 | POST /api/v1/actions/:id/execute | 执行自动操作 |

---

### 十七、系统管理模块

#### 17.1 用户管理
| 功能 | 说明 |
|------|------|
| 用户CRUD | 创建、查询、更新、删除用户 |
| 用户认证 | 登录、注册、登出 |
| 密码管理 | 密码修改、重置 |
| 用户状态 | 启用、禁用用户 |

#### 17.2 菜单管理
| 功能 | 说明 |
|------|------|
| 菜单CRUD | 创建、查询、更新、删除菜单 |
| 菜单树 | 树形菜单结构 |
| 权限关联 | 菜单与权限关联 |

---

### 十八、WebSocket 实时通信

| 功能 | 说明 |
|------|------|
| 连接端点 | GET /ws |
| 实时推送 | 服务器状态、告警、任务进度等 |
| 心跳机制 | 保持连接活跃 |
| 自动重连 | 断线自动重连 |

---

## 数据库模型

### 核心数据表

| 模块 | 主要数据表 |
|------|-----------|
| 服务器管理 | servers, groups, server_metrics, server_logs, docker_containers, port_infos |
| Kubernetes | k8s_clusters, hpa_configs, scale_histories |
| 灰度发布 | canary_releases, canary_steps, canary_configs |
| 负载均衡 | load_balancers, lb_backends, lb_optimization_history, lb_algorithms |
| 证书管理 | certificates, cert_renewal_history |
| CDN管理 | cdn_domains, cdn_cache_rules, cdn_optimization_history |
| 智能部署 | deploy_projects, deploy_plans, deploy_tasks, deploy_steps |
| 任务调度 | scheduler_tasks, scheduler_batches, cron_jobs, scheduler_templates, task_executions |
| Agent管理 | agents, agent_versions, agent_upgrade_tasks, agent_heartbeat_records, gray_release_strategies |
| 高可用 | ha_nodes, ha_sessions, ha_locks, ha_configs, failover_records, leader_election_records |
| 备份管理 | backup_policies, backup_records, restore_records, snapshot_policies, snapshot_records, drill_plans |
| 成本控制 | cost_records, cost_summaries, cost_forecasts, waste_detections, idle_resources, budgets, k8s_cost_records |
| 多租户 | tenants, tenant_quotas, tenant_users, tenant_roles, tenant_invitations, tenant_audit_logs |
| 权限管理 | permissions, roles, user_roles, users |
| 告警管理 | alerts, detect_rules, auto_actions, ai_decisions |
| 系统管理 | sys_users, sys_roles, sys_menus, sys_role_menus, sys_role_apis |

---

## API 权限控制

### 权限中间件

系统使用权限中间件对每个 API 进行权限检查：

```go
// 示例：服务器管理权限控制
servers.GET("", middleware.RequirePermission("server:view"), server.GetServerList)
servers.POST("", middleware.RequirePermission("server:add"), server.AddServer)
servers.PUT("/:id", middleware.RequirePermission("server:edit"), server.UpdateServer)
servers.DELETE("/:id", middleware.RequirePermission("server:delete"), server.DeleteServer)
servers.POST("/:id/command", middleware.RequirePermission("server:execute"), server.ExecuteCommand)
```

### 权限检查流程

1. 用户登录获取 JWT Token
2. 请求携带 Token 访问 API
3. 中间件解析 Token 获取用户 ID
4. 查询用户角色和权限
5. 检查是否具有所需权限
6. 允许或拒绝请求

---

## 部署说明

### 环境要求

- Go 1.21+
- Node.js 18+
- MySQL 8.0+ / PostgreSQL 14+ / SQLite 3
- Redis 6.0+

### 后端启动

```bash
cd server
go mod download
go run main.go
```

### 前端启动

```bash
npm install
npm run dev
```

### Docker 部署

```bash
docker-compose up -d
```

---

## 项目结构

```
├── server/                    # Go 后端
│   ├── api/v1/               # API 处理器
│   │   ├── agent/            # Agent 管理
│   │   ├── auth/             # 认证
│   │   ├── backup/           # 备份管理
│   │   ├── canary/           # 灰度发布
│   │   ├── cert/             # 证书管理
│   │   ├── cdn/              # CDN 管理
│   │   ├── cost/             # 成本控制
│   │   ├── deploy/           # 智能部署
│   │   ├── ha/               # 高可用
│   │   ├── kubernetes/       # K8s 管理
│   │   ├── loadbalancer/     # 负载均衡
│   │   ├── permission/       # 权限管理
│   │   ├── scheduler/        # 任务调度
│   │   ├── server/           # 服务器管理
│   │   ├── system/           # 系统管理
│   │   └── tenant/           # 多租户
│   ├── model/                # 数据模型
│   ├── service/              # 业务逻辑
│   ├── middleware/           # 中间件
│   ├── router/               # 路由配置
│   ├── config/               # 配置
│   ├── global/               # 全局变量
│   ├── utils/                # 工具函数
│   └── websocket/            # WebSocket
├── src/                       # Next.js 前端
│   ├── app/                  # 页面路由
│   ├── components/           # 组件
│   └── lib/                  # 工具库
└── document.md                # 本文档
```

---

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.0.0 | 2024-01 | 初始版本，基础服务器管理 |
| 1.1.0 | 2024-02 | 添加 Kubernetes 管理 |
| 1.2.0 | 2024-03 | 添加灰度发布、负载均衡 |
| 1.3.0 | 2024-04 | 添加证书管理、CDN 管理 |
| 1.4.0 | 2024-05 | 添加智能部署、任务调度 |
| 1.5.0 | 2024-06 | 添加 Agent 管理、高可用 |
| 1.6.0 | 2024-07 | 添加灾备备份系统 |
| 1.7.0 | 2024-08 | 添加成本控制系统 |
| 1.8.0 | 2024-09 | 添加多租户系统 |
| 1.9.0 | 2024-10 | 完善RBAC权限控制系统 |

---

## 许可证

MIT License

---

## 联系方式

- 项目地址: https://github.com/fredphp/yunwei
- 问题反馈: https://github.com/fredphp/yunwei/issues
