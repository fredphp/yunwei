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
                <el-button @click="refreshNodes">刷新</el-button>
                <el-button type="primary" @click="showAddNodeDialog = true">添加节点</el-button>
              </div>
            </div>
          </template>
          <el-table :data="nodes" v-loading="loading">
            <el-table-column prop="hostname" label="主机名" min-width="140" />
            <el-table-column prop="ip" label="IP 地址" width="130" />
            <el-table-column label="角色" width="100">
              <template #default="{ row }">
                <el-tag :type="row.isLeader ? 'danger' : 'info'">
                  {{ row.isLeader ? 'Leader' : 'Follower' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'healthy' ? 'success' : 'danger'">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="CPU" width="100">
              <template #default="{ row }">
                <el-progress :percentage="row.cpuUsage" :color="getProgressColor(row.cpuUsage)" :stroke-width="8" />
              </template>
            </el-table-column>
            <el-table-column label="内存" width="100">
              <template #default="{ row }">
                <el-progress :percentage="row.memoryUsage" :color="getProgressColor(row.memoryUsage)" :stroke-width="8" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="viewMetrics(row)">监控</el-button>
                <el-button size="small" @click="disableNode(row)" v-if="row.status === 'healthy'">禁用</el-button>
                <el-button size="small" @click="enableNode(row)" v-else>启用</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 分布式锁 -->
        <el-card class="mt-4">
          <template #header>
            <div class="card-header">
              <span>分布式锁</span>
              <el-button size="small" @click="refreshLocks">刷新</el-button>
            </div>
          </template>
          <el-table :data="locks" size="small">
            <el-table-column prop="key" label="锁名称" min-width="200" />
            <el-table-column prop="holder" label="持有者" width="140" />
            <el-table-column prop="ttl" label="TTL" width="100" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.expired ? 'danger' : 'success'">
                  {{ row.expired ? '已过期' : '有效' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" width="160" />
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button size="small" type="danger" text @click="forceReleaseLock(row)">强制释放</el-button>
              </template>
            </el-table-column>
          </el-table>
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
              <strong>{{ leaderNode?.hostname || '无' }}</strong>
            </div>
            <div class="leader-detail">
              <span>任期:</span>
              <strong>{{ leaderTerm }}</strong>
            </div>
            <div class="leader-actions">
              <el-button type="warning" @click="resignLeader">辞职 Leader</el-button>
              <el-button type="danger" @click="showForceLeaderDialog = true">强制指定</el-button>
            </div>
          </div>
          <el-divider />
          <div class="election-history">
            <div class="history-title">选举历史</div>
            <el-timeline>
              <el-timeline-item v-for="record in electionRecords" :key="record.id" :timestamp="record.time" placement="top">
                {{ record.node }} 成为 Leader
              </el-timeline-item>
            </el-timeline>
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
          <el-button type="danger" @click="triggerFailover" :disabled="failoverInProgress">
            手动触发故障转移
          </el-button>
          <el-divider />
          <div class="failover-records">
            <div class="history-title">转移记录</div>
            <div v-for="record in failoverRecords" :key="record.id" class="failover-item">
              <div>{{ record.from }} → {{ record.to }}</div>
              <div class="failover-time">{{ record.time }}</div>
            </div>
          </div>
        </el-card>

        <!-- 集群事件 -->
        <el-card>
          <template #header>
            <span>集群事件</span>
          </template>
          <div class="events-list">
            <div v-for="event in events" :key="event.id" class="event-item">
              <el-icon :color="getEventColor(event.type)">
                <WarningFilled v-if="event.type === 'warning'" />
                <InfoFilled v-else />
              </el-icon>
              <div class="event-content">
                <div class="event-message">{{ event.message }}</div>
                <div class="event-time">{{ event.time }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 对话框 -->
    <el-dialog v-model="showForceLeaderDialog" title="强制指定 Leader" width="400px">
      <el-form label-width="80px">
        <el-form-item label="选择节点">
          <el-select v-model="forceLeaderNode" style="width: 100%;">
            <el-option v-for="node in nodes" :key="node.id" :label="node.hostname" :value="node.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-alert type="warning" :closable="false" class="mb-3">
        强制指定 Leader 可能导致数据不一致，请谨慎操作
      </el-alert>
      <template #footer>
        <el-button @click="showForceLeaderDialog = false">取消</el-button>
        <el-button type="danger" @click="forceLeader">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Cpu, Connection, Timer, SuccessFilled, WarningFilled, InfoFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const showAddNodeDialog = ref(false)
const showForceLeaderDialog = ref(false)
const forceLeaderNode = ref('')
const failoverInProgress = ref(false)

const stats = ref({
  healthyNodes: 3,
  totalNodes: 3,
  uptime: '15d 8h',
  quorum: true
})

const leaderTerm = ref(42)

const nodes = ref([
  { id: 1, hostname: 'ha-node-01', ip: '192.168.1.201', isLeader: true, status: 'healthy', cpuUsage: 25, memoryUsage: 45 },
  { id: 2, hostname: 'ha-node-02', ip: '192.168.1.202', isLeader: false, status: 'healthy', cpuUsage: 18, memoryUsage: 38 },
  { id: 3, hostname: 'ha-node-03', ip: '192.168.1.203', isLeader: false, status: 'healthy', cpuUsage: 22, memoryUsage: 42 }
])

const leaderNode = computed(() => nodes.value.find(n => n.isLeader))

const locks = ref([
  { id: 1, key: '/config/global', holder: 'ha-node-01', ttl: '30s', expired: false, createdAt: '2024-02-23 10:25' },
  { id: 2, key: '/task/backup', holder: 'ha-node-02', ttl: '60s', expired: false, createdAt: '2024-02-23 10:20' }
])

const electionRecords = ref([
  { id: 1, node: 'ha-node-01', time: '2024-02-23 08:00' },
  { id: 2, node: 'ha-node-02', time: '2024-02-20 14:30' },
  { id: 3, node: 'ha-node-03', time: '2024-02-15 09:15' }
])

const failoverRecords = ref([
  { id: 1, from: 'ha-node-02', to: 'ha-node-01', time: '2024-02-20 14:30' },
  { id: 2, from: 'ha-node-03', to: 'ha-node-02', time: '2024-02-10 11:20' }
])

const events = ref([
  { id: 1, type: 'info', message: 'ha-node-01 加入集群', time: '2小时前' },
  { id: 2, type: 'warning', message: 'ha-node-02 心跳超时', time: '5小时前' },
  { id: 3, type: 'info', message: 'Leader 选举完成', time: '1天前' }
])

const fetchNodes = async () => {
  loading.value = true
  try {
    const res = await request.get('/ha/nodes')
    // nodes.value = res.data || []
  } catch (error) {
    console.error('获取节点失败', error)
  } finally {
    loading.value = false
  }
}

const refreshNodes = () => fetchNodes()
const refreshLocks = () => {
  ElMessage.success('已刷新')
}

const viewMetrics = (node: any) => {
  ElMessage.info('监控功能开发中')
}

const enableNode = async (node: any) => {
  await request.post(`/ha/nodes/${node.id}/enable`)
  ElMessage.success('节点已启用')
}

const disableNode = async (node: any) => {
  await request.post(`/ha/nodes/${node.id}/disable`)
  ElMessage.success('节点已禁用')
}

const resignLeader = async () => {
  try {
    await ElMessageBox.confirm('确定要辞职 Leader 吗？将触发重新选举', '提示', { type: 'warning' })
    await request.post('/ha/leader/resign')
    ElMessage.success('已提交辞职请求')
  } catch {}
}

const forceLeader = async () => {
  try {
    await ElMessageBox.confirm('确定强制指定该节点为 Leader？', '警告', { type: 'warning' })
    await request.post('/ha/leader/force', { nodeId: forceLeaderNode.value })
    ElMessage.success('已强制指定')
    showForceLeaderDialog.value = false
  } catch {}
}

const triggerFailover = async () => {
  try {
    await ElMessageBox.confirm('确定触发故障转移？', '警告', { type: 'warning' })
    await request.post('/ha/failover/trigger')
    ElMessage.success('故障转移已触发')
    failoverInProgress.value = true
  } catch {}
}

const forceReleaseLock = async (lock: any) => {
  try {
    await ElMessageBox.confirm('确定强制释放该锁？', '提示', { type: 'warning' })
    await request.post(`/ha/locks/${lock.key}/release`)
    ElMessage.success('锁已释放')
  } catch {}
}

const getProgressColor = (value: number) => {
  if (value >= 90) return '#f56c6c'
  if (value >= 70) return '#e6a23c'
  return '#67c23a'
}

const getEventColor = (type: string) => {
  return type === 'warning' ? '#e6a23c' : '#409eff'
}

onMounted(() => {
  fetchNodes()
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

.failover-time {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
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

.event-message {
  font-size: 14px;
}

.event-time {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
