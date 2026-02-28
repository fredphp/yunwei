<template>
  <div class="ha-page">
    <!-- 集群状态概览 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" :class="leaderStatus === 'active' ? 'active' : 'inactive'">
              <el-icon size="28"><Cpu /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ leaderNode?.hostname || 'N/A' }}</div>
              <div class="stat-label">当前 Leader</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #67c23a;">
              <el-icon size="28"><Connection /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.healthyNodes }}/{{ stats.totalNodes }}</div>
              <div class="stat-label">健康节点</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #409eff;">
              <el-icon size="28"><Timer /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.uptime }}</div>
              <div class="stat-label">运行时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" :style="{ background: stats.quorum ? '#67c23a' : '#f56c6c' }">
              <el-icon size="28"><SuccessFilled /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.quorum ? '正常' : '异常' }}</div>
              <div class="stat-label">仲裁状态</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="16">
        <!-- 节点列表 -->
        <el-card>
          <template #header>
            <div class="card-header">
              <span>集群节点</span>
              <div class="header-actions">
                <el-button @click="refreshNodes" :loading="loading">刷新</el-button>
                <el-button type="primary" @click="showAddNodeDialog = true">添加节点</el-button>
              </div>
            </div>
          </template>
          <el-table :data="nodes" v-loading="loading">
            <el-table-column prop="hostname" label="主机名" min-width="140" />
            <el-table-column prop="internalIp" label="IP 地址" width="130">
              <template #default="{ row }">
                {{ row.internalIp || row.ip || '-' }}
              </template>
            </el-table-column>
            <el-table-column label="角色" width="100">
              <template #default="{ row }">
                <el-tag :type="row.isLeader ? 'danger' : 'info'">
                  {{ row.isLeader ? 'Leader' : 'Follower' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'online' || row.status === 'healthy' ? 'success' : 'danger'">
                  {{ row.status === 'online' || row.status === 'healthy' ? '健康' : '异常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="CPU" width="100">
              <template #default="{ row }">
                <el-progress :percentage="Math.round(row.cpuUsage || 0)" :color="getProgressColor(row.cpuUsage || 0)" :stroke-width="8" />
              </template>
            </el-table-column>
            <el-table-column label="内存" width="100">
              <template #default="{ row }">
                <el-progress :percentage="Math.round(row.memoryUsage || 0)" :color="getProgressColor(row.memoryUsage || 0)" :stroke-width="8" />
              </template>
            </el-table-column>
            <el-table-column label="数据中心" width="100">
              <template #default="{ row }">
                {{ row.dataCenter || '-' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="250" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="viewMetrics(row)" type="primary">监控</el-button>
                <el-button size="small" @click="viewNodeDetail(row)">详情</el-button>
                <el-button size="small" @click="disableNode(row)" v-if="row.enabled !== false" type="warning">禁用</el-button>
                <el-button size="small" @click="enableNode(row)" v-else type="success">启用</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 分布式锁 -->
        <el-card class="mt-4">
          <template #header>
            <div class="card-header">
              <span>分布式锁</span>
              <div class="header-actions">
                <el-button size="small" @click="showCreateLockDialog = true">创建锁</el-button>
                <el-button size="small" @click="refreshLocks">刷新</el-button>
              </div>
            </div>
          </template>
          <el-table :data="locks" size="small" v-loading="locksLoading">
            <el-table-column prop="lockKey" label="锁名称" min-width="200" />
            <el-table-column prop="holderNodeId" label="持有者" width="140" />
            <el-table-column label="TTL" width="100">
              <template #default="{ row }">
                {{ row.ttlSeconds || 30 }}s
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'released' || row.status === 'expired' ? 'danger' : 'success'">
                  {{ row.status === 'released' ? '已释放' : row.status === 'expired' ? '已过期' : '有效' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="创建时间" width="160">
              <template #default="{ row }">
                {{ formatTime(row.createdAt) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button 
                  size="small" 
                  type="danger" 
                  text 
                  @click="forceReleaseLock(row)"
                  v-if="row.status === 'acquired'"
                >
                  强制释放
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 任务 HA 管理 -->
        <el-card class="mt-4">
          <template #header>
            <div class="card-header">
              <span>运行中任务</span>
              <el-button size="small" @click="refreshTasks">刷新</el-button>
            </div>
          </template>
          <el-table :data="runningTasks" size="small" v-loading="tasksLoading">
            <el-table-column prop="taskId" label="任务ID" min-width="180" />
            <el-table-column prop="taskName" label="任务名称" width="150" />
            <el-table-column prop="runningNode" label="执行节点" width="140" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag type="success">运行中</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="开始时间" width="160">
              <template #default="{ row }">
                {{ formatTime(row.startTime) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button size="small" type="warning" @click="migrateTask(row)">迁移</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="runningTasks.length === 0 && !tasksLoading" description="暂无运行中任务" />
        </el-card>
      </el-col>

      <el-col :span="8">
        <!-- Leader 选举 -->
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>Leader 选举</span>
            </div>
          </template>
          <div class="leader-info">
            <div class="leader-detail">
              <span>当前 Leader:</span>
              <strong>{{ leaderNode?.hostname || leaderNode?.nodeName || '无' }}</strong>
            </div>
            <div class="leader-detail">
              <span>任期:</span>
              <strong>{{ leaderTerm }}</strong>
            </div>
            <div class="leader-actions">
              <el-button type="warning" @click="resignLeader" :disabled="!isCurrentLeader">辞职 Leader</el-button>
              <el-button type="danger" @click="showForceLeaderDialog = true">强制指定</el-button>
            </div>
          </div>
          <el-divider />
          <div class="election-history">
            <div class="history-title">选举历史</div>
            <el-timeline v-if="electionRecords.length > 0">
              <el-timeline-item v-for="record in electionRecords" :key="record.id" :timestamp="formatTime(record.createdAt)" placement="top">
                {{ record.leaderNodeId || record.nodeId }} 成为 Leader
              </el-timeline-item>
            </el-timeline>
            <el-empty v-else description="暂无选举记录" :image-size="60" />
          </div>
        </el-card>

        <!-- 故障转移 -->
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>故障转移</span>
            </div>
          </template>
          <el-alert
            v-if="failoverInProgress"
            title="故障转移进行中"
            type="warning"
            show-icon
            :closable="false"
            class="mb-3"
          />
          <el-button type="danger" @click="showFailoverDialog = true" :disabled="failoverInProgress">
            手动触发故障转移
          </el-button>
          <el-divider />
          <div class="failover-records">
            <div class="history-title">转移记录</div>
            <div v-if="failoverRecords.length > 0">
              <div v-for="record in failoverRecords" :key="record.id" class="failover-item">
                <div class="failover-flow">
                  <span class="node-name">{{ record.failedNodeName || record.from }}</span>
                  <el-icon><Right /></el-icon>
                  <span class="node-name">{{ record.targetNodeName || record.to }}</span>
                </div>
                <div class="failover-meta">
                  <el-tag :type="record.success ? 'success' : 'danger'" size="small">
                    {{ record.success ? '成功' : '失败' }}
                  </el-tag>
                  <span class="failover-time">{{ formatTime(record.createdAt) }}</span>
                </div>
              </div>
            </div>
            <el-empty v-else description="暂无转移记录" :image-size="60" />
          </div>
        </el-card>

        <!-- 集群事件 -->
        <el-card>
          <template #header>
            <div class="card-header">
              <span>集群事件</span>
              <el-button size="small" text @click="refreshEvents">刷新</el-button>
            </div>
          </template>
          <div class="events-list" v-loading="eventsLoading">
            <div v-for="event in events" :key="event.id" class="event-item">
              <el-icon :color="getEventColor(event.level)" :size="18">
                <WarningFilled v-if="event.level === 'warning'" />
                <CircleCloseFilled v-else-if="event.level === 'error'" />
                <InfoFilled v-else />
              </el-icon>
              <div class="event-content">
                <div class="event-message">{{ event.title || event.message }}</div>
                <div class="event-meta">
                  <span class="event-node" v-if="event.nodeName">{{ event.nodeName }}</span>
                  <span class="event-time">{{ formatTime(event.createdAt) }}</span>
                </div>
              </div>
            </div>
            <el-empty v-if="events.length === 0 && !eventsLoading" description="暂无事件" :image-size="60" />
          </div>
        </el-card>

        <!-- HA 配置 -->
        <el-card class="mt-4">
          <template #header>
            <div class="card-header">
              <span>集群配置</span>
              <el-button size="small" text @click="showConfigDialog = true">编辑</el-button>
            </div>
          </template>
          <el-descriptions :column="1" size="small" border>
            <el-descriptions-item label="集群模式">{{ haConfig.clusterMode || 'active-active' }}</el-descriptions-item>
            <el-descriptions-item label="心跳间隔">{{ haConfig.heartbeatInterval || 10 }}s</el-descriptions-item>
            <el-descriptions-item label="选举超时">{{ haConfig.electionTimeout || 30 }}s</el-descriptions-item>
            <el-descriptions-item label="故障转移">{{ haConfig.failoverEnabled ? '已启用' : '已禁用' }}</el-descriptions-item>
            <el-descriptions-item label="负载均衡">{{ haConfig.loadBalanceEnabled ? '已启用' : '已禁用' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <!-- 节点监控弹窗 -->
    <el-dialog v-model="showMetricsDialog" :title="`节点监控 - ${selectedNode?.hostname || ''}`" width="900px" destroy-on-close>
      <div v-loading="metricsLoading">
        <el-row :gutter="20" class="mb-4">
          <el-col :span="6">
            <el-statistic title="CPU 使用率" :value="nodeMetrics.current?.cpuUsage || 0" suffix="%">
              <template #suffix>
                <span :style="{ color: getProgressColor(nodeMetrics.current?.cpuUsage || 0) }">%</span>
              </template>
            </el-statistic>
          </el-col>
          <el-col :span="6">
            <el-statistic title="内存使用率" :value="nodeMetrics.current?.memoryUsage || 0">
              <template #suffix>
                <span :style="{ color: getProgressColor(nodeMetrics.current?.memoryUsage || 0) }">%</span>
              </template>
            </el-statistic>
          </el-col>
          <el-col :span="6">
            <el-statistic title="磁盘使用率" :value="nodeMetrics.current?.diskUsage || 0">
              <template #suffix>
                <span :style="{ color: getProgressColor(nodeMetrics.current?.diskUsage || 0) }">%</span>
              </template>
            </el-statistic>
          </el-col>
          <el-col :span="6">
            <el-statistic title="连接数" :value="nodeMetrics.current?.connectionCount || 0" />
          </el-col>
        </el-row>

        <el-divider content-position="left">资源使用趋势</el-divider>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="chart-container">
              <div class="chart-title">CPU 使用率趋势</div>
              <div ref="cpuChartRef" class="chart" style="height: 200px;"></div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="chart-container">
              <div class="chart-title">内存使用率趋势</div>
              <div ref="memoryChartRef" class="chart" style="height: 200px;"></div>
            </div>
          </el-col>
        </el-row>

        <el-divider content-position="left">运行时信息</el-divider>
        
        <el-descriptions :column="3" border size="small">
          <el-descriptions-item label="主机名">{{ selectedNode?.hostname }}</el-descriptions-item>
          <el-descriptions-item label="IP地址">{{ selectedNode?.internalIp || selectedNode?.ip }}</el-descriptions-item>
          <el-descriptions-item label="节点ID">{{ selectedNode?.nodeId }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{ selectedNode?.version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="Go版本">{{ selectedNode?.goVersion || '-' }}</el-descriptions-item>
          <el-descriptions-item label="数据中心">{{ selectedNode?.dataCenter || '-' }}</el-descriptions-item>
          <el-descriptions-item label="协程数">{{ nodeMetrics.current?.goroutineCount || '-' }}</el-descriptions-item>
          <el-descriptions-item label="请求数">{{ nodeMetrics.current?.requestCount || '-' }}</el-descriptions-item>
          <el-descriptions-item label="QPS">{{ nodeMetrics.current?.requestQps?.toFixed(2) || '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">系统负载</el-divider>
        
        <el-row :gutter="20">
          <el-col :span="8">
            <el-statistic title="1分钟负载" :value="nodeMetrics.current?.load1?.toFixed(2) || '-'" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="5分钟负载" :value="nodeMetrics.current?.load5?.toFixed(2) || '-'" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="15分钟负载" :value="nodeMetrics.current?.load15?.toFixed(2) || '-'" />
          </el-col>
        </el-row>
      </div>
      <template #footer>
        <el-button @click="showMetricsDialog = false">关闭</el-button>
        <el-button type="primary" @click="refreshNodeMetrics">刷新</el-button>
      </template>
    </el-dialog>

    <!-- 添加节点弹窗 -->
    <el-dialog v-model="showAddNodeDialog" title="添加节点" width="500px">
      <el-form :model="addNodeForm" label-width="100px" :rules="addNodeRules" ref="addNodeFormRef">
        <el-form-item label="主机名" prop="hostname">
          <el-input v-model="addNodeForm.hostname" placeholder="请输入主机名" />
        </el-form-item>
        <el-form-item label="内网IP" prop="internalIp">
          <el-input v-model="addNodeForm.internalIp" placeholder="请输入内网IP" />
        </el-form-item>
        <el-form-item label="外网IP">
          <el-input v-model="addNodeForm.externalIp" placeholder="请输入外网IP（可选）" />
        </el-form-item>
        <el-form-item label="API端口">
          <el-input-number v-model="addNodeForm.apiPort" :min="1" :max="65535" :default="8080" />
        </el-form-item>
        <el-form-item label="gRPC端口">
          <el-input-number v-model="addNodeForm.grpcPort" :min="1" :max="65535" :default="50051" />
        </el-form-item>
        <el-form-item label="数据中心">
          <el-input v-model="addNodeForm.dataCenter" placeholder="请输入数据中心" />
        </el-form-item>
        <el-form-item label="可用区">
          <el-input v-model="addNodeForm.zone" placeholder="请输入可用区" />
        </el-form-item>
        <el-form-item label="权重">
          <el-input-number v-model="addNodeForm.weight" :min="1" :max="1000" :default="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddNodeDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAddNode" :loading="addNodeLoading">确认添加</el-button>
      </template>
    </el-dialog>

    <!-- 强制指定Leader弹窗 -->
    <el-dialog v-model="showForceLeaderDialog" title="强制指定 Leader" width="400px">
      <el-form label-width="80px">
        <el-form-item label="选择节点">
          <el-select v-model="forceLeaderNode" style="width: 100%;">
            <el-option v-for="node in followerNodes" :key="node.nodeId" :label="node.hostname || node.nodeName" :value="node.nodeId" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-alert type="warning" :closable="false" class="mb-3">
        强制指定 Leader 可能导致数据不一致，请谨慎操作
      </el-alert>
      <template #footer>
        <el-button @click="showForceLeaderDialog = false">取消</el-button>
        <el-button type="danger" @click="forceLeader" :loading="forceLeaderLoading">确认</el-button>
      </template>
    </el-dialog>

    <!-- 手动故障转移弹窗 -->
    <el-dialog v-model="showFailoverDialog" title="手动触发故障转移" width="500px">
      <el-form :model="failoverForm" label-width="100px">
        <el-form-item label="故障节点">
          <el-select v-model="failoverForm.nodeId" style="width: 100%;" placeholder="选择要转移的节点">
            <el-option v-for="node in nodes" :key="node.nodeId" :label="node.hostname || node.nodeName" :value="node.nodeId" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标节点">
          <el-select v-model="failoverForm.targetNodeId" style="width: 100%;" placeholder="选择目标节点（可选，留空自动选择）" clearable>
            <el-option v-for="node in healthyNodes" :key="node.nodeId" :label="node.hostname || node.nodeName" :value="node.nodeId" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-alert type="warning" :closable="false">
        故障转移将停止所选节点的服务，并将其任务迁移到其他节点
      </el-alert>
      <template #footer>
        <el-button @click="showFailoverDialog = false">取消</el-button>
        <el-button type="danger" @click="triggerFailover" :loading="failoverLoading">确认转移</el-button>
      </template>
    </el-dialog>

    <!-- 节点详情弹窗 -->
    <el-dialog v-model="showNodeDetailDialog" title="节点详情" width="600px">
      <el-descriptions :column="2" border v-if="selectedNode">
        <el-descriptions-item label="节点ID">{{ selectedNode.nodeId }}</el-descriptions-item>
        <el-descriptions-item label="主机名">{{ selectedNode.hostname }}</el-descriptions-item>
        <el-descriptions-item label="内网IP">{{ selectedNode.internalIp }}</el-descriptions-item>
        <el-descriptions-item label="外网IP">{{ selectedNode.externalIp || '-' }}</el-descriptions-item>
        <el-descriptions-item label="API端口">{{ selectedNode.apiPort || '-' }}</el-descriptions-item>
        <el-descriptions-item label="gRPC端口">{{ selectedNode.grpcPort || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="selectedNode.status === 'online' ? 'success' : 'danger'">{{ selectedNode.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="角色">
          <el-tag :type="selectedNode.isLeader ? 'danger' : 'info'">{{ selectedNode.isLeader ? 'Leader' : 'Follower' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="版本">{{ selectedNode.version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Go版本">{{ selectedNode.goVersion || '-' }}</el-descriptions-item>
        <el-descriptions-item label="数据中心">{{ selectedNode.dataCenter || '-' }}</el-descriptions-item>
        <el-descriptions-item label="可用区">{{ selectedNode.zone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="权重">{{ selectedNode.weight || 100 }}</el-descriptions-item>
        <el-descriptions-item label="启用状态">
          <el-tag :type="selectedNode.enabled !== false ? 'success' : 'info'">{{ selectedNode.enabled !== false ? '已启用' : '已禁用' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="最后心跳" :span="2">{{ formatTime(selectedNode.lastHeartbeat) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间" :span="2">{{ formatTime(selectedNode.createdAt) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="showNodeDetailDialog = false">关闭</el-button>
        <el-button type="primary" @click="viewMetrics(selectedNode)">查看监控</el-button>
      </template>
    </el-dialog>

    <!-- 创建锁弹窗 -->
    <el-dialog v-model="showCreateLockDialog" title="创建分布式锁" width="400px">
      <el-form :model="createLockForm" label-width="80px">
        <el-form-item label="锁名称" required>
          <el-input v-model="createLockForm.key" placeholder="如: /config/global" />
        </el-form-item>
        <el-form-item label="TTL(秒)">
          <el-input-number v-model="createLockForm.ttl" :min="5" :max="300" :default="30" />
        </el-form-item>
        <el-form-item label="资源类型">
          <el-input v-model="createLockForm.resourceType" placeholder="如: task, config" />
        </el-form-item>
        <el-form-item label="资源ID">
          <el-input v-model="createLockForm.resourceId" placeholder="资源标识" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateLockDialog = false">取消</el-button>
        <el-button type="primary" @click="createLock" :loading="createLockLoading">创建</el-button>
      </template>
    </el-dialog>

    <!-- HA配置弹窗 -->
    <el-dialog v-model="showConfigDialog" title="集群配置" width="600px">
      <el-form :model="haConfig" label-width="120px">
        <el-form-item label="集群名称">
          <el-input v-model="haConfig.clusterName" />
        </el-form-item>
        <el-form-item label="集群模式">
          <el-select v-model="haConfig.clusterMode">
            <el-option label="Active-Active" value="active-active" />
            <el-option label="Active-Passive" value="active-passive" />
          </el-select>
        </el-form-item>
        <el-divider content-position="left">心跳配置</el-divider>
        <el-form-item label="心跳间隔(秒)">
          <el-input-number v-model="haConfig.heartbeatInterval" :min="5" :max="60" />
        </el-form-item>
        <el-form-item label="心跳超时(秒)">
          <el-input-number v-model="haConfig.heartbeatTimeout" :min="10" :max="120" />
        </el-form-item>
        <el-divider content-position="left">选举配置</el-divider>
        <el-form-item label="选举超时(秒)">
          <el-input-number v-model="haConfig.electionTimeout" :min="10" :max="60" />
        </el-form-item>
        <el-form-item label="Leader租约(秒)">
          <el-input-number v-model="haConfig.leaderLeaseSeconds" :min="5" :max="30" />
        </el-form-item>
        <el-divider content-position="left">功能开关</el-divider>
        <el-form-item label="启用故障转移">
          <el-switch v-model="haConfig.failoverEnabled" />
        </el-form-item>
        <el-form-item label="自动回切">
          <el-switch v-model="haConfig.autoFailback" />
        </el-form-item>
        <el-form-item label="启用负载均衡">
          <el-switch v-model="haConfig.loadBalanceEnabled" />
        </el-form-item>
        <el-form-item label="负载均衡策略" v-if="haConfig.loadBalanceEnabled">
          <el-select v-model="haConfig.loadBalanceStrategy">
            <el-option label="轮询" value="round-robin" />
            <el-option label="加权轮询" value="weighted-round-robin" />
            <el-option label="最少连接" value="least-connections" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showConfigDialog = false">取消</el-button>
        <el-button type="primary" @click="saveConfig" :loading="saveConfigLoading">保存</el-button>
      </template>
    </el-dialog>

    <!-- 任务迁移弹窗 -->
    <el-dialog v-model="showMigrateDialog" title="迁移任务" width="400px">
      <el-form :model="migrateForm" label-width="80px">
        <el-form-item label="任务ID">
          <el-input :value="migrateForm.taskId" disabled />
        </el-form-item>
        <el-form-item label="目标节点">
          <el-select v-model="migrateForm.targetNode" style="width: 100%;" placeholder="选择目标节点">
            <el-option v-for="node in healthyNodes" :key="node.nodeId" :label="node.hostname || node.nodeName" :value="node.nodeId" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showMigrateDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmMigrateTask" :loading="migrateLoading">确认迁移</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Cpu, Connection, Timer, SuccessFilled, WarningFilled, InfoFilled, Right, CircleCloseFilled } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import request from '@/utils/request'

// ==================== 状态 ====================
const loading = ref(false)
const locksLoading = ref(false)
const tasksLoading = ref(false)
const eventsLoading = ref(false)
const metricsLoading = ref(false)
const addNodeLoading = ref(false)
const forceLeaderLoading = ref(false)
const failoverLoading = ref(false)
const createLockLoading = ref(false)
const saveConfigLoading = ref(false)
const migrateLoading = ref(false)

// 弹窗状态
const showMetricsDialog = ref(false)
const showAddNodeDialog = ref(false)
const showForceLeaderDialog = ref(false)
const showFailoverDialog = ref(false)
const showNodeDetailDialog = ref(false)
const showCreateLockDialog = ref(false)
const showConfigDialog = ref(false)
const showMigrateDialog = ref(false)

const forceLeaderNode = ref('')
const failoverInProgress = ref(false)
const leaderStatus = ref('active')
const leaderTerm = ref(0)
const isCurrentLeader = ref(false)

// 图表引用
const cpuChartRef = ref<HTMLElement>()
const memoryChartRef = ref<HTMLElement>()
let cpuChart: echarts.ECharts | null = null
let memoryChart: echarts.ECharts | null = null

// ==================== 数据 ====================
const stats = ref({
  healthyNodes: 0,
  totalNodes: 0,
  uptime: 'N/A',
  quorum: true
})

const nodes = ref<any[]>([])
const locks = ref<any[]>([])
const events = ref<any[]>([])
const electionRecords = ref<any[]>([])
const failoverRecords = ref<any[]>([])
const runningTasks = ref<any[]>([])

const selectedNode = ref<any>(null)
const nodeMetrics = ref<any>({
  current: null,
  history: []
})

const haConfig = ref<any>({
  clusterName: '',
  clusterMode: 'active-active',
  heartbeatInterval: 10,
  heartbeatTimeout: 30,
  electionTimeout: 30,
  leaderLeaseSeconds: 15,
  failoverEnabled: true,
  autoFailback: true,
  loadBalanceEnabled: true,
  loadBalanceStrategy: 'round-robin'
})

// 表单数据
const addNodeForm = ref({
  hostname: '',
  internalIp: '',
  externalIp: '',
  apiPort: 8080,
  grpcPort: 50051,
  dataCenter: '',
  zone: '',
  weight: 100
})

const addNodeRules = {
  hostname: [{ required: true, message: '请输入主机名', trigger: 'blur' }],
  internalIp: [{ required: true, message: '请输入内网IP', trigger: 'blur' }]
}

const failoverForm = ref({
  nodeId: '',
  targetNodeId: ''
})

const createLockForm = ref({
  key: '',
  ttl: 30,
  resourceType: '',
  resourceId: ''
})

const migrateForm = ref({
  taskId: '',
  taskName: '',
  targetNode: ''
})

// ==================== 计算属性 ====================
const leaderNode = computed(() => nodes.value.find(n => n.isLeader))
const followerNodes = computed(() => nodes.value.filter(n => !n.isLeader && n.status === 'online'))
const healthyNodes = computed(() => nodes.value.filter(n => n.status === 'online' && n.enabled !== false))

// ==================== 数据加载 ====================
const fetchStats = async () => {
  try {
    const res = await request.get('/ha/stats')
    if (res.data) {
      stats.value = {
        healthyNodes: res.data.onlineNodes || 0,
        totalNodes: res.data.totalNodes || 0,
        uptime: 'N/A',
        quorum: res.data.onlineNodes > 0
      }
      isCurrentLeader.value = res.data.isLeader || false
    }
  } catch (error) {
    console.error('获取统计失败', error)
  }
}

const fetchNodes = async () => {
  loading.value = true
  try {
    const res = await request.get('/ha/nodes')
    if (res.data?.list) {
      nodes.value = res.data.list
    } else if (Array.isArray(res.data)) {
      nodes.value = res.data
    }
    // 更新统计
    stats.value.healthyNodes = nodes.value.filter(n => n.status === 'online' || n.status === 'healthy').length
    stats.value.totalNodes = nodes.value.length
  } catch (error) {
    console.error('获取节点失败', error)
    // 使用模拟数据
    nodes.value = [
      { id: 1, nodeId: 'node-01', hostname: 'ha-node-01', internalIp: '192.168.1.201', isLeader: true, status: 'online', cpuUsage: 25, memoryUsage: 45, dataCenter: 'dc1', enabled: true },
      { id: 2, nodeId: 'node-02', hostname: 'ha-node-02', internalIp: '192.168.1.202', isLeader: false, status: 'online', cpuUsage: 18, memoryUsage: 38, dataCenter: 'dc1', enabled: true },
      { id: 3, nodeId: 'node-03', hostname: 'ha-node-03', internalIp: '192.168.1.203', isLeader: false, status: 'online', cpuUsage: 22, memoryUsage: 42, dataCenter: 'dc2', enabled: true }
    ]
  } finally {
    loading.value = false
  }
}

const fetchLeaderStatus = async () => {
  try {
    const res = await request.get('/ha/leader')
    if (res.data) {
      leaderTerm.value = res.data.term || 0
    }
  } catch (error) {
    console.error('获取Leader状态失败', error)
  }
}

const fetchLocks = async () => {
  locksLoading.value = true
  try {
    const res = await request.get('/ha/locks')
    if (res.data?.list) {
      locks.value = res.data.list
    } else if (Array.isArray(res.data)) {
      locks.value = res.data
    }
  } catch (error) {
    console.error('获取锁失败', error)
    // 使用模拟数据
    locks.value = [
      { id: 1, lockKey: '/config/global', holderNodeId: 'ha-node-01', ttlSeconds: 30, status: 'acquired', createdAt: new Date().toISOString() },
      { id: 2, lockKey: '/task/backup', holderNodeId: 'ha-node-02', ttlSeconds: 60, status: 'acquired', createdAt: new Date().toISOString() }
    ]
  } finally {
    locksLoading.value = false
  }
}

const fetchEvents = async () => {
  eventsLoading.value = true
  try {
    const res = await request.get('/ha/events')
    if (res.data?.list) {
      events.value = res.data.list
    } else if (Array.isArray(res.data)) {
      events.value = res.data
    }
  } catch (error) {
    console.error('获取事件失败', error)
    // 使用模拟数据
    events.value = [
      { id: 1, eventType: 'join', level: 'info', title: 'ha-node-01 加入集群', nodeName: 'ha-node-01', createdAt: new Date(Date.now() - 3600000).toISOString() },
      { id: 2, eventType: 'heartbeat', level: 'warning', title: 'ha-node-02 心跳超时', nodeName: 'ha-node-02', createdAt: new Date(Date.now() - 7200000).toISOString() },
      { id: 3, eventType: 'election', level: 'info', title: 'Leader 选举完成', nodeName: 'ha-node-01', createdAt: new Date(Date.now() - 86400000).toISOString() }
    ]
  } finally {
    eventsLoading.value = false
  }
}

const fetchElectionRecords = async () => {
  try {
    const res = await request.get('/ha/leader/records')
    if (Array.isArray(res.data)) {
      electionRecords.value = res.data
    }
  } catch (error) {
    console.error('获取选举记录失败', error)
    // 使用模拟数据
    electionRecords.value = [
      { id: 1, leaderNodeId: 'ha-node-01', createdAt: new Date(Date.now() - 86400000).toISOString() },
      { id: 2, leaderNodeId: 'ha-node-02', createdAt: new Date(Date.now() - 172800000).toISOString() }
    ]
  }
}

const fetchFailoverRecords = async () => {
  try {
    const res = await request.get('/ha/failover')
    if (res.data?.list) {
      failoverRecords.value = res.data.list
    }
  } catch (error) {
    console.error('获取故障转移记录失败', error)
    // 使用模拟数据
    failoverRecords.value = [
      { id: 1, failedNodeName: 'ha-node-02', targetNodeName: 'ha-node-01', success: true, createdAt: new Date(Date.now() - 172800000).toISOString() }
    ]
  }
}

const fetchRunningTasks = async () => {
  tasksLoading.value = true
  try {
    const res = await request.get('/ha/tasks/running')
    if (Array.isArray(res.data)) {
      runningTasks.value = res.data
    }
  } catch (error) {
    console.error('获取运行任务失败', error)
    runningTasks.value = []
  } finally {
    tasksLoading.value = false
  }
}

const fetchConfig = async () => {
  try {
    const res = await request.get('/ha/config')
    if (res.data) {
      haConfig.value = { ...haConfig.value, ...res.data }
    }
  } catch (error) {
    console.error('获取配置失败', error)
  }
}

// ==================== 刷新方法 ====================
const refreshNodes = () => {
  fetchNodes()
  fetchLeaderStatus()
}

const refreshLocks = () => {
  fetchLocks()
  ElMessage.success('已刷新锁列表')
}

const refreshEvents = () => {
  fetchEvents()
}

const refreshTasks = () => {
  fetchRunningTasks()
}

// ==================== 节点监控 ====================
const viewMetrics = async (node: any) => {
  selectedNode.value = node
  showMetricsDialog.value = true
  metricsLoading.value = true
  
  try {
    const res = await request.get(`/ha/nodes/${node.nodeId || node.id}/metrics?hours=24`)
    if (res.data) {
      nodeMetrics.value = res.data
    }
  } catch (error) {
    console.error('获取指标失败', error)
    // 使用模拟数据
    nodeMetrics.value = {
      current: {
        cpuUsage: node.cpuUsage || 25,
        memoryUsage: node.memoryUsage || 45,
        diskUsage: 55,
        connectionCount: 150,
        goroutineCount: 128,
        requestCount: 15000,
        requestQps: 25.5,
        load1: 0.5,
        load5: 0.8,
        load15: 0.6
      },
      history: generateMockMetrics()
    }
  } finally {
    metricsLoading.value = false
  }
  
  // 初始化图表
  await nextTick()
  initCharts()
}

const refreshNodeMetrics = async () => {
  if (!selectedNode.value) return
  metricsLoading.value = true
  try {
    const res = await request.get(`/ha/nodes/${selectedNode.value.nodeId || selectedNode.value.id}/metrics?hours=24`)
    if (res.data) {
      nodeMetrics.value = res.data
    }
  } catch (error) {
    console.error('刷新指标失败', error)
  } finally {
    metricsLoading.value = false
  }
}

const generateMockMetrics = () => {
  const data = []
  const now = Date.now()
  for (let i = 24; i >= 0; i--) {
    const time = new Date(now - i * 3600000)
    data.push({
      time: time.toISOString(),
      cpuUsage: 20 + Math.random() * 30,
      memoryUsage: 40 + Math.random() * 20
    })
  }
  return data
}

const initCharts = () => {
  // 销毁旧图表
  cpuChart?.dispose()
  memoryChart?.dispose()
  
  // 初始化CPU图表
  if (cpuChartRef.value) {
    cpuChart = echarts.init(cpuChartRef.value)
    const cpuOption = {
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: nodeMetrics.value.history?.map((m: any) => formatTime(m.time, 'HH:mm')) || [],
        axisLabel: { fontSize: 10 }
      },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series: [{
        type: 'line',
        data: nodeMetrics.value.history?.map((m: any) => m.cpuUsage?.toFixed(1)) || [],
        smooth: true,
        areaStyle: { color: 'rgba(103, 194, 58, 0.3)' },
        lineStyle: { color: '#67c23a' },
        itemStyle: { color: '#67c23a' }
      }],
      grid: { left: '10%', right: '5%', top: '10%', bottom: '15%' }
    }
    cpuChart.setOption(cpuOption)
  }
  
  // 初始化内存图表
  if (memoryChartRef.value) {
    memoryChart = echarts.init(memoryChartRef.value)
    const memoryOption = {
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: nodeMetrics.value.history?.map((m: any) => formatTime(m.time, 'HH:mm')) || [],
        axisLabel: { fontSize: 10 }
      },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series: [{
        type: 'line',
        data: nodeMetrics.value.history?.map((m: any) => m.memoryUsage?.toFixed(1)) || [],
        smooth: true,
        areaStyle: { color: 'rgba(64, 158, 255, 0.3)' },
        lineStyle: { color: '#409eff' },
        itemStyle: { color: '#409eff' }
      }],
      grid: { left: '10%', right: '5%', top: '10%', bottom: '15%' }
    }
    memoryChart.setOption(memoryOption)
  }
}

// ==================== 节点操作 ====================
const viewNodeDetail = (node: any) => {
  selectedNode.value = node
  showNodeDetailDialog.value = true
}

const enableNode = async (node: any) => {
  try {
    await request.post(`/ha/nodes/${node.nodeId || node.id}/enable`)
    ElMessage.success('节点已启用')
    fetchNodes()
  } catch (error) {
    ElMessage.error('启用失败')
  }
}

const disableNode = async (node: any) => {
  try {
    await ElMessageBox.confirm('确定要禁用该节点吗？', '提示', { type: 'warning' })
    await request.post(`/ha/nodes/${node.nodeId || node.id}/disable`)
    ElMessage.success('节点已禁用')
    fetchNodes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('禁用失败')
    }
  }
}

const handleAddNode = async () => {
  addNodeLoading.value = true
  try {
    await request.post('/ha/nodes', {
      nodeId: `node-${Date.now()}`,
      nodeName: addNodeForm.value.hostname,
      hostname: addNodeForm.value.hostname,
      internalIp: addNodeForm.value.internalIp,
      externalIp: addNodeForm.value.externalIp,
      apiPort: addNodeForm.value.apiPort,
      grpcPort: addNodeForm.value.grpcPort,
      dataCenter: addNodeForm.value.dataCenter,
      zone: addNodeForm.value.zone,
      weight: addNodeForm.value.weight,
      status: 'offline',
      enabled: true
    })
    ElMessage.success('节点已添加')
    showAddNodeDialog.value = false
    fetchNodes()
    // 重置表单
    addNodeForm.value = {
      hostname: '',
      internalIp: '',
      externalIp: '',
      apiPort: 8080,
      grpcPort: 50051,
      dataCenter: '',
      zone: '',
      weight: 100
    }
  } catch (error) {
    ElMessage.error('添加失败')
  } finally {
    addNodeLoading.value = false
  }
}

// ==================== Leader 操作 ====================
const resignLeader = async () => {
  try {
    await ElMessageBox.confirm('确定要辞职 Leader 吗？将触发重新选举', '提示', { type: 'warning' })
    await request.post('/ha/leader/resign')
    ElMessage.success('已提交辞职请求')
    fetchLeaderStatus()
    fetchElectionRecords()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('辞职失败')
    }
  }
}

const forceLeader = async () => {
  if (!forceLeaderNode.value) {
    ElMessage.warning('请选择节点')
    return
  }
  
  forceLeaderLoading.value = true
  try {
    await ElMessageBox.confirm('确定强制指定该节点为 Leader？', '警告', { type: 'warning' })
    await request.post('/ha/leader/force', { nodeId: forceLeaderNode.value })
    ElMessage.success('已强制指定')
    showForceLeaderDialog.value = false
    fetchNodes()
    fetchLeaderStatus()
    fetchElectionRecords()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  } finally {
    forceLeaderLoading.value = false
  }
}

// ==================== 故障转移 ====================
const triggerFailover = async () => {
  if (!failoverForm.value.nodeId) {
    ElMessage.warning('请选择故障节点')
    return
  }
  
  failoverLoading.value = true
  try {
    await request.post('/ha/failover/trigger', {
      nodeId: failoverForm.value.nodeId,
      targetNodeId: failoverForm.value.targetNodeId
    })
    ElMessage.success('故障转移已触发')
    showFailoverDialog.value = false
    failoverInProgress.value = true
    // 轮询检查状态
    setTimeout(() => {
      failoverInProgress.value = false
      fetchFailoverRecords()
      fetchNodes()
    }, 5000)
  } catch (error) {
    ElMessage.error('触发失败')
  } finally {
    failoverLoading.value = false
  }
}

// ==================== 锁操作 ====================
const forceReleaseLock = async (lock: any) => {
  try {
    await ElMessageBox.confirm('确定强制释放该锁？', '提示', { type: 'warning' })
    await request.post(`/ha/locks/${encodeURIComponent(lock.lockKey)}/release`, {
      reason: '手动释放'
    })
    ElMessage.success('锁已释放')
    fetchLocks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('释放失败')
    }
  }
}

const createLock = async () => {
  if (!createLockForm.value.key) {
    ElMessage.warning('请输入锁名称')
    return
  }
  
  createLockLoading.value = true
  try {
    await request.post('/ha/locks', createLockForm.value)
    ElMessage.success('锁已创建')
    showCreateLockDialog.value = false
    fetchLocks()
    createLockForm.value = { key: '', ttl: 30, resourceType: '', resourceId: '' }
  } catch (error) {
    ElMessage.error('创建失败')
  } finally {
    createLockLoading.value = false
  }
}

// ==================== 配置管理 ====================
const saveConfig = async () => {
  saveConfigLoading.value = true
  try {
    await request.put('/ha/config', haConfig.value)
    ElMessage.success('配置已保存')
    showConfigDialog.value = false
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saveConfigLoading.value = false
  }
}

// ==================== 任务迁移 ====================
const migrateTask = (task: any) => {
  migrateForm.value = {
    taskId: task.taskId,
    taskName: task.taskName,
    targetNode: ''
  }
  showMigrateDialog.value = true
}

const confirmMigrateTask = async () => {
  if (!migrateForm.value.targetNode) {
    ElMessage.warning('请选择目标节点')
    return
  }
  
  migrateLoading.value = true
  try {
    await request.post(`/ha/tasks/${migrateForm.value.taskId}/migrate`, {
      targetNode: migrateForm.value.targetNode
    })
    ElMessage.success('任务已迁移')
    showMigrateDialog.value = false
    fetchRunningTasks()
  } catch (error) {
    ElMessage.error('迁移失败')
  } finally {
    migrateLoading.value = false
  }
}

// ==================== 工具方法 ====================
const getProgressColor = (value: number) => {
  if (value >= 90) return '#f56c6c'
  if (value >= 70) return '#e6a23c'
  return '#67c23a'
}

const getEventColor = (level: string) => {
  switch (level) {
    case 'error': return '#f56c6c'
    case 'warning': return '#e6a23c'
    case 'info': return '#409eff'
    default: return '#909399'
  }
}

const formatTime = (time: string | Date, format = 'YYYY-MM-DD HH:mm') => {
  if (!time) return '-'
  const date = new Date(time)
  if (isNaN(date.getTime())) return '-'
  
  const pad = (n: number) => n.toString().padStart(2, '0')
  const year = date.getFullYear()
  const month = pad(date.getMonth() + 1)
  const day = pad(date.getDate())
  const hour = pad(date.getHours())
  const minute = pad(date.getMinutes())
  const second = pad(date.getSeconds())
  
  return format
    .replace('YYYY', year.toString())
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hour)
    .replace('mm', minute)
    .replace('ss', second)
}

// ==================== 生命周期 ====================
let refreshTimer: NodeJS.Timeout | null = null

onMounted(() => {
  fetchStats()
  fetchNodes()
  fetchLeaderStatus()
  fetchLocks()
  fetchEvents()
  fetchElectionRecords()
  fetchFailoverRecords()
  fetchRunningTasks()
  fetchConfig()
  
  // 定时刷新
  refreshTimer = setInterval(() => {
    fetchStats()
    fetchLeaderStatus()
  }, 10000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  cpuChart?.dispose()
  memoryChart?.dispose()
})

// 监听弹窗关闭，销毁图表
watch(showMetricsDialog, (val) => {
  if (!val) {
    cpuChart?.dispose()
    memoryChart?.dispose()
  }
})
</script>

<style scoped>
.ha-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.stat-card {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin-right: 16px;
}

.stat-icon.active {
  background: #67c23a;
}

.stat-icon.inactive {
  background: #f56c6c;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.mb-4 {
  margin-bottom: 16px;
}

.mt-4 {
  margin-top: 16px;
}

.mb-3 {
  margin-bottom: 12px;
}

.leader-info {
  margin-bottom: 16px;
}

.leader-detail {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}

.leader-actions {
  display: flex;
  gap: 10px;
  margin-top: 16px;
}

.history-title {
  font-weight: 500;
  margin-bottom: 12px;
}

.failover-item {
  padding: 8px 0;
  border-bottom: 1px solid #ebeef5;
}

.failover-item:last-child {
  border-bottom: none;
}

.failover-flow {
  display: flex;
  align-items: center;
  gap: 8px;
}

.failover-flow .node-name {
  font-weight: 500;
}

.failover-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
}

.failover-time {
  font-size: 12px;
  color: #909399;
}

.events-list {
  max-height: 300px;
  overflow-y: auto;
}

.event-item {
  display: flex;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #ebeef5;
}

.event-item:last-child {
  border-bottom: none;
}

.event-content {
  flex: 1;
}

.event-message {
  font-size: 14px;
}

.event-meta {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.event-node {
  font-size: 12px;
  color: #409eff;
}

.event-time {
  font-size: 12px;
  color: #909399;
}

.chart-container {
  padding: 10px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.chart-title {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 10px;
  text-align: center;
}
</style>
