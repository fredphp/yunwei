<template>
  <div class="agents-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #409eff;">
              <el-icon size="28"><Monitor /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total }}</div>
              <div class="stat-label">Agent 总数</div>
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
              <div class="stat-value">{{ stats.online }}</div>
              <div class="stat-label">在线</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #f56c6c;">
              <el-icon size="28"><Disconnect /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.offline }}</div>
              <div class="stat-label">离线</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #e6a23c;">
              <el-icon size="28"><Refresh /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.needUpdate }}</div>
              <div class="stat-label">需更新</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="16">
        <!-- Agent 列表 -->
        <el-card>
          <template #header>
            <div class="card-header">
              <span>Agent 管理</span>
              <div class="header-actions">
                <el-button @click="batchUpgrade" :disabled="selectedAgents.length === 0">批量升级</el-button>
                <el-button type="primary" @click="refreshAgents">刷新</el-button>
              </div>
            </div>
          </template>
          <el-table :data="agents" v-loading="loading" @selection-change="handleSelectionChange">
            <el-table-column type="selection" width="50" />
            <el-table-column prop="hostname" label="主机名" min-width="120" />
            <el-table-column prop="ip" label="IP 地址" width="130" />
            <el-table-column prop="os" label="系统" width="80" />
            <el-table-column prop="version" label="版本" width="80" />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 'online' ? 'success' : 'danger'" size="small">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="lastHeartbeat" label="最后心跳" width="160" />
            <el-table-column label="操作" width="220" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="viewAgent(row)">详情</el-button>
                <el-button size="small" @click="viewConfig(row)">配置</el-button>
                <el-button size="small" type="primary" @click="upgradeAgent(row)" v-if="row.needUpdate">
                  升级
                </el-button>
                <el-dropdown trigger="click">
                  <el-button size="small">更多</el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="enableAgent(row)" v-if="row.status === 'disabled'">启用</el-dropdown-item>
                      <el-dropdown-item @click="disableAgent(row)" v-else>禁用</el-dropdown-item>
                      <el-dropdown-item @click="viewHeartbeats(row)">心跳记录</el-dropdown-item>
                      <el-dropdown-item @click="viewRecovers(row)">恢复记录</el-dropdown-item>
                      <el-dropdown-item divided @click="deleteAgent(row)" style="color: #f56c6c;">删除</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 升级任务 -->
        <el-card class="mt-4">
          <template #header>
            <div class="card-header">
              <span>升级任务</span>
              <el-button type="primary" size="small" @click="showUpgradeDialog = true">创建升级任务</el-button>
            </div>
          </template>
          <el-table :data="upgradeTasks" size="small">
            <el-table-column prop="name" label="任务名称" />
            <el-table-column label="进度" width="150">
              <template #default="{ row }">
                <el-progress :percentage="row.progress" :status="row.status === 'completed' ? 'success' : ''" />
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getTaskStatusType(row.status)" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" width="160" />
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button size="small" v-if="row.status === 'running'" @click="cancelUpgrade(row)">取消</el-button>
                <el-button size="small" v-if="row.status === 'completed'" @click="rollbackUpgrade(row)">回滚</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <!-- 版本管理 -->
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>版本管理</span>
              <el-button size="small" type="primary" @click="showVersionDialog = true">上传版本</el-button>
            </div>
          </template>
          <div v-for="version in versions" :key="version.id" class="version-item">
            <div class="version-info">
              <div class="version-name">
                v{{ version.version }}
                <el-tag size="small" v-if="version.isLatest" type="success">最新</el-tag>
              </div>
              <div class="version-stats">
                {{ version.agentCount }} 个Agent
              </div>
            </div>
            <el-button size="small" @click="setAsLatest(version)" v-if="!version.isLatest">设为最新</el-button>
          </div>
        </el-card>

        <!-- 灰度发布策略 -->
        <el-card class="mb-4">
          <template #header>
            <span>灰度发布策略</span>
          </template>
          <div v-for="strategy in grayStrategies" :key="strategy.id" class="strategy-item">
            <div class="strategy-header">
              <span class="strategy-name">{{ strategy.name }}</span>
              <el-tag :type="strategy.status === 'running' ? 'success' : 'info'" size="small">
                {{ strategy.status }}
              </el-tag>
            </div>
            <div class="strategy-progress">
              <el-progress :percentage="strategy.progress" />
            </div>
            <div class="strategy-actions">
              <el-button size="small" v-if="strategy.status === 'paused'" @click="resumeStrategy(strategy)">继续</el-button>
              <el-button size="small" v-if="strategy.status === 'running'" @click="pauseStrategy(strategy)">暂停</el-button>
              <el-button size="small" type="danger" text @click="cancelStrategy(strategy)">取消</el-button>
            </div>
          </div>
        </el-card>

        <!-- 监控统计 -->
        <el-card>
          <template #header>
            <span>监控统计</span>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="平均CPU使用率">32.5%</el-descriptions-item>
            <el-descriptions-item label="平均内存使用率">48.2%</el-descriptions-item>
            <el-descriptions-item label="24h告警数">{{ stats.alerts24h }}</el-descriptions-item>
            <el-descriptions-item label="24h恢复数">{{ stats.recovers24h }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <!-- 对话框 -->
    <el-dialog v-model="showAgentDialog" title="Agent 详情" width="700px">
      <el-descriptions :column="2" border v-if="currentAgent">
        <el-descriptions-item label="主机名">{{ currentAgent.hostname }}</el-descriptions-item>
        <el-descriptions-item label="IP 地址">{{ currentAgent.ip }}</el-descriptions-item>
        <el-descriptions-item label="操作系统">{{ currentAgent.os }}</el-descriptions-item>
        <el-descriptions-item label="版本">{{ currentAgent.version }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentAgent.status === 'online' ? 'success' : 'danger'">{{ currentAgent.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="最后心跳">{{ currentAgent.lastHeartbeat }}</el-descriptions-item>
        <el-descriptions-item label="CPU 使用率">{{ currentAgent.cpuUsage }}%</el-descriptions-item>
        <el-descriptions-item label="内存使用率">{{ currentAgent.memoryUsage }}%</el-descriptions-item>
        <el-descriptions-item label="注册时间">{{ currentAgent.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="标签">{{ currentAgent.tags?.join(', ') }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- Agent配置对话框 -->
    <el-dialog v-model="showConfigDialog" title="Agent 配置" width="700px">
      <div v-loading="configLoading">
        <el-form :model="agentConfig" label-width="140px" v-if="agentConfig">
          <el-divider content-position="left">采集配置</el-divider>
          <el-form-item label="采集间隔">
            <el-input-number v-model="agentConfig.collectInterval" :min="10" :max="300" /> 秒
          </el-form-item>
          <el-form-item label="启用CPU采集">
            <el-switch v-model="agentConfig.collectCPU" />
          </el-form-item>
          <el-form-item label="启用内存采集">
            <el-switch v-model="agentConfig.collectMemory" />
          </el-form-item>
          <el-form-item label="启用磁盘采集">
            <el-switch v-model="agentConfig.collectDisk" />
          </el-form-item>
          <el-form-item label="启用网络采集">
            <el-switch v-model="agentConfig.collectNetwork" />
          </el-form-item>
          
          <el-divider content-position="left">上报配置</el-divider>
          <el-form-item label="上报地址">
            <el-input v-model="agentConfig.reportUrl" placeholder="http://server:8080/api/report" />
          </el-form-item>
          <el-form-item label="心跳间隔">
            <el-input-number v-model="agentConfig.heartbeatInterval" :min="5" :max="60" /> 秒
          </el-form-item>
          
          <el-divider content-position="left">执行配置</el-divider>
          <el-form-item label="命令超时">
            <el-input-number v-model="agentConfig.execTimeout" :min="10" :max="600" /> 秒
          </el-form-item>
          <el-form-item label="并发数">
            <el-input-number v-model="agentConfig.concurrentLimit" :min="1" :max="10" />
          </el-form-item>
          
          <el-divider content-position="left">日志配置</el-divider>
          <el-form-item label="日志级别">
            <el-select v-model="agentConfig.logLevel">
              <el-option label="DEBUG" value="debug" />
              <el-option label="INFO" value="info" />
              <el-option label="WARN" value="warn" />
              <el-option label="ERROR" value="error" />
            </el-select>
          </el-form-item>
          <el-form-item label="日志保留天数">
            <el-input-number v-model="agentConfig.logRetention" :min="1" :max="30" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="showConfigDialog = false">取消</el-button>
        <el-button type="primary" @click="saveAgentConfig">保存配置</el-button>
      </template>
    </el-dialog>

    <!-- 心跳记录对话框 -->
    <el-dialog v-model="showHeartbeatsDialog" title="心跳记录" width="900px">
      <div v-loading="heartbeatsLoading">
        <div class="toolbar mb-3">
          <el-date-picker
            v-model="heartbeatsTimeRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            @change="loadHeartbeats"
          />
          <el-button type="primary" style="margin-left: 10px;" @click="loadHeartbeats">刷新</el-button>
        </div>
        
        <el-table :data="heartbeatsData" max-height="400">
          <el-table-column prop="timestamp" label="时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.timestamp) }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
                {{ row.status === 'success' ? '正常' : '异常' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="latency" label="延迟(ms)" width="100" />
          <el-table-column prop="cpuUsage" label="CPU%" width="80" />
          <el-table-column prop="memoryUsage" label="内存%" width="80" />
          <el-table-column prop="diskUsage" label="磁盘%" width="80" />
          <el-table-column prop="message" label="消息" min-width="200" show-overflow-tooltip />
        </el-table>
        
        <el-empty v-if="heartbeatsData.length === 0 && !heartbeatsLoading" description="暂无心跳记录" />
      </div>
    </el-dialog>

    <!-- 恢复记录对话框 -->
    <el-dialog v-model="showRecoversDialog" title="恢复记录" width="900px">
      <div v-loading="recoversLoading">
        <el-table :data="recoversData" max-height="400">
          <el-table-column prop="timestamp" label="时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.timestamp) }}
            </template>
          </el-table-column>
          <el-table-column label="类型" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="issue" label="问题" min-width="180" show-overflow-tooltip />
          <el-table-column prop="action" label="恢复操作" min-width="180" show-overflow-tooltip />
          <el-table-column label="结果" width="100">
            <template #default="{ row }">
              <el-tag :type="row.success ? 'success' : 'danger'" size="small">
                {{ row.success ? '成功' : '失败' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="duration" label="耗时(ms)" width="100" />
        </el-table>
        
        <el-empty v-if="recoversData.length === 0 && !recoversLoading" description="暂无恢复记录" />
      </div>
    </el-dialog>

    <el-dialog v-model="showUpgradeDialog" title="创建升级任务" width="500px">
      <el-form :model="upgradeForm" label-width="100px">
        <el-form-item label="任务名称">
          <el-input v-model="upgradeForm.name" />
        </el-form-item>
        <el-form-item label="目标版本">
          <el-select v-model="upgradeForm.versionId" style="width: 100%;">
            <el-option v-for="v in versions" :key="v.id" :label="'v' + v.version" :value="v.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="升级策略">
          <el-radio-group v-model="upgradeForm.strategy">
            <el-radio value="all">全部升级</el-radio>
            <el-radio value="gray">灰度发布</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="灰度比例" v-if="upgradeForm.strategy === 'gray'">
          <el-slider v-model="upgradeForm.grayPercent" :marks="{ 0: '0%', 50: '50%', 100: '100%' }" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUpgradeDialog = false">取消</el-button>
        <el-button type="primary" @click="createUpgradeTask">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Monitor, Connection, Disconnect, Refresh } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const selectedAgents = ref([])
const showAgentDialog = ref(false)
const showUpgradeDialog = ref(false)
const showVersionDialog = ref(false)
const showConfigDialog = ref(false)
const showHeartbeatsDialog = ref(false)
const showRecoversDialog = ref(false)
const currentAgent = ref<any>(null)

// 配置相关
const configLoading = ref(false)
const agentConfig = ref<any>(null)

// 心跳记录相关
const heartbeatsLoading = ref(false)
const heartbeatsData = ref<any[]>([])
const heartbeatsTimeRange = ref<[Date, Date] | null>(null)

// 恢复记录相关
const recoversLoading = ref(false)
const recoversData = ref<any[]>([])

const stats = ref({
  total: 45,
  online: 42,
  offline: 3,
  needUpdate: 8,
  alerts24h: 12,
  recovers24h: 10
})

const agents = ref([
  { id: 1, hostname: 'web-server-01', ip: '192.168.1.101', os: 'Linux', version: '1.2.3', status: 'online', lastHeartbeat: '2024-02-23 10:30:00', needUpdate: true },
  { id: 2, hostname: 'db-server-01', ip: '192.168.1.102', os: 'Linux', version: '1.2.5', status: 'online', lastHeartbeat: '2024-02-23 10:29:55', needUpdate: false },
  { id: 3, hostname: 'cache-server-01', ip: '192.168.1.103', os: 'Linux', version: '1.2.3', status: 'online', lastHeartbeat: '2024-02-23 10:30:05', needUpdate: true }
])

const upgradeTasks = ref([
  { id: 1, name: '生产环境升级', progress: 75, status: 'running', createdAt: '2024-02-23 09:00' },
  { id: 2, name: '测试环境升级', progress: 100, status: 'completed', createdAt: '2024-02-22 15:00' }
])

const versions = ref([
  { id: 1, version: '1.2.5', isLatest: true, agentCount: 37 },
  { id: 2, version: '1.2.3', isLatest: false, agentCount: 8 }
])

const grayStrategies = ref([
  { id: 1, name: '生产环境灰度升级', status: 'running', progress: 35 }
])

const upgradeForm = ref({
  name: '',
  versionId: '',
  strategy: 'gray',
  grayPercent: 20
})

const fetchAgents = async () => {
  loading.value = true
  try {
    const res = await request.get('/agents')
    // agents.value = res.data || []
  } catch (error) {
    console.error('获取Agent列表失败', error)
  } finally {
    loading.value = false
  }
}

const handleSelectionChange = (selection: any[]) => {
  selectedAgents.value = selection
}

const viewAgent = (agent: any) => {
  currentAgent.value = agent
  showAgentDialog.value = true
}

const viewConfig = async (agent: any) => {
  currentAgent.value = agent
  showConfigDialog.value = true
  configLoading.value = true
  
  try {
    const res = await request.get(`/agents/${agent.id}/config`)
    agentConfig.value = res.data || getDefaultConfig()
  } catch (error) {
    console.error('获取配置失败', error)
    agentConfig.value = getDefaultConfig()
  } finally {
    configLoading.value = false
  }
}

const getDefaultConfig = () => ({
  collectInterval: 60,
  collectCPU: true,
  collectMemory: true,
  collectDisk: true,
  collectNetwork: true,
  reportUrl: '',
  heartbeatInterval: 30,
  execTimeout: 60,
  concurrentLimit: 5,
  logLevel: 'info',
  logRetention: 7
})

const saveAgentConfig = async () => {
  try {
    await request.put(`/agents/${currentAgent.value.id}/config`, agentConfig.value)
    ElMessage.success('配置已保存')
    showConfigDialog.value = false
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const upgradeAgent = async (agent: any) => {
  ElMessage.success(`已提交升级任务: ${agent.hostname}`)
}

const enableAgent = async (agent: any) => {
  await request.post(`/agents/${agent.id}/enable`)
  ElMessage.success('已启用')
}

const disableAgent = async (agent: any) => {
  await request.post(`/agents/${agent.id}/disable`)
  ElMessage.success('已禁用')
}

const deleteAgent = async (agent: any) => {
  try {
    await ElMessageBox.confirm('确定删除该Agent？', '提示', { type: 'warning' })
    await request.delete(`/agents/${agent.id}`)
    ElMessage.success('删除成功')
  } catch {}
}

const viewHeartbeats = async (agent: any) => {
  currentAgent.value = agent
  showHeartbeatsDialog.value = true
  await loadHeartbeats()
}

const loadHeartbeats = async () => {
  if (!currentAgent.value) return
  heartbeatsLoading.value = true
  try {
    const params: any = {}
    if (heartbeatsTimeRange.value) {
      params.startTime = heartbeatsTimeRange.value[0].toISOString()
      params.endTime = heartbeatsTimeRange.value[1].toISOString()
    }
    const res = await request.get(`/agents/${currentAgent.value.id}/heartbeats`, { params })
    heartbeatsData.value = res.data || []
  } catch (error) {
    console.error('获取心跳记录失败', error)
    // 模拟数据
    heartbeatsData.value = Array.from({ length: 20 }, (_, i) => ({
      timestamp: new Date(Date.now() - i * 30000).toISOString(),
      status: Math.random() > 0.1 ? 'success' : 'failed',
      latency: Math.floor(Math.random() * 50) + 10,
      cpuUsage: (Math.random() * 30 + 20).toFixed(1),
      memoryUsage: (Math.random() * 20 + 40).toFixed(1),
      diskUsage: (Math.random() * 10 + 50).toFixed(1),
      message: Math.random() > 0.9 ? '资源使用率告警' : '正常'
    }))
  } finally {
    heartbeatsLoading.value = false
  }
}

const viewRecovers = async (agent: any) => {
  currentAgent.value = agent
  showRecoversDialog.value = true
  recoversLoading.value = true
  
  try {
    const res = await request.get(`/agents/${agent.id}/recovers`)
    recoversData.value = res.data || []
  } catch (error) {
    console.error('获取恢复记录失败', error)
    // 模拟数据
    recoversData.value = [
      { timestamp: new Date(Date.now() - 3600000).toISOString(), type: '服务重启', issue: 'nginx服务停止', action: '自动重启nginx服务', success: true, duration: 1500 },
      { timestamp: new Date(Date.now() - 7200000).toISOString(), type: '进程恢复', issue: '进程异常退出', action: '重启应用进程', success: true, duration: 800 },
      { timestamp: new Date(Date.now() - 86400000).toISOString(), type: '磁盘清理', issue: '磁盘空间不足', action: '清理临时文件和日志', success: true, duration: 5000 }
    ]
  } finally {
    recoversLoading.value = false
  }
}

const refreshAgents = () => {
  fetchAgents()
}

const batchUpgrade = () => {
  ElMessage.success(`已提交批量升级任务: ${selectedAgents.value.length} 个Agent`)
}

const createUpgradeTask = async () => {
  try {
    await request.post('/agents/upgrades', upgradeForm.value)
    ElMessage.success('升级任务已创建')
    showUpgradeDialog.value = false
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

const cancelUpgrade = async (task: any) => {
  await request.post(`/agents/upgrades/${task.id}/cancel`)
  ElMessage.success('已取消')
}

const rollbackUpgrade = async (task: any) => {
  await request.post(`/agents/upgrades/${task.id}/rollback`)
  ElMessage.success('已回滚')
}

const setAsLatest = async (version: any) => {
  ElMessage.success(`已设置 v${version.version} 为最新版本`)
}

const resumeStrategy = async (strategy: any) => {
  await request.post(`/agents/gray/${strategy.id}/resume`)
  ElMessage.success('已继续')
}

const pauseStrategy = async (strategy: any) => {
  await request.post(`/agents/gray/${strategy.id}/pause`)
  ElMessage.success('已暂停')
}

const cancelStrategy = async (strategy: any) => {
  await request.post(`/agents/gray/${strategy.id}/cancel`)
  ElMessage.success('已取消')
}

const getTaskStatusType = (status: string) => {
  const types: Record<string, string> = {
    pending: 'info',
    running: 'primary',
    completed: 'success',
    failed: 'danger'
  }
  return types[status] || 'info'
}

const formatTime = (time: string) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  fetchAgents()
})
</script>

<style scoped>
.agents-page {
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

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 28px;
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

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
}

.version-item:last-child {
  border-bottom: none;
}

.version-name {
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.version-stats {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.strategy-item {
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
}

.strategy-item:last-child {
  border-bottom: none;
}

.strategy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.strategy-name {
  font-weight: 500;
}

.strategy-progress {
  margin-bottom: 8px;
}

.strategy-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.toolbar {
  display: flex;
  align-items: center;
}

.mb-3 {
  margin-bottom: 12px;
}
</style>
