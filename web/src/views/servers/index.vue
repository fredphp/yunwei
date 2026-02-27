<template>
  <div class="servers-page">
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
              <div class="stat-label">服务器总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #67c23a;">
              <el-icon size="28"><CircleCheck /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.online }}</div>
              <div class="stat-label">在线服务器</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #f56c6c;">
              <el-icon size="28"><CircleClose /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.offline }}</div>
              <div class="stat-label">离线服务器</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #e6a23c;">
              <el-icon size="28"><Warning /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.warning }}</div>
              <div class="stat-label">告警服务器</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 服务器列表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>服务器管理</span>
          <div class="header-actions">
            <el-input v-model="searchKeyword" placeholder="搜索服务器" style="width: 200px; margin-right: 10px;" clearable>
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-button type="primary" @click="showAddDialog = true">
              <el-icon><Plus /></el-icon>
              添加服务器
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="filteredServers" v-loading="loading" @row-click="showServerDetail">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="ip" label="IP 地址" width="140" />
        <el-table-column prop="os" label="系统" width="100" />
        <el-table-column label="CPU" width="120">
          <template #default="{ row }">
            <el-progress :percentage="row.cpuUsage" :color="getProgressColor(row.cpuUsage)" :stroke-width="8" />
          </template>
        </el-table-column>
        <el-table-column label="内存" width="120">
          <template #default="{ row }">
            <el-progress :percentage="row.memUsage" :color="getProgressColor(row.memUsage)" :stroke-width="8" />
          </template>
        </el-table-column>
        <el-table-column label="磁盘" width="120">
          <template #default="{ row }">
            <el-progress :percentage="row.diskUsage" :color="getProgressColor(row.diskUsage)" :stroke-width="8" />
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="groupId" label="分组" width="100" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click.stop="viewMetrics(row)">监控</el-button>
            <el-button size="small" @click.stop="viewLogs(row)">日志</el-button>
            <el-button size="small" type="primary" @click.stop="analyzeServer(row)">AI分析</el-button>
            <el-dropdown @click.stop trigger="click">
              <el-button size="small">更多<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="executeCommand(row)">执行命令</el-dropdown-item>
                  <el-dropdown-item @click="refreshStatus(row)">刷新状态</el-dropdown-item>
                  <el-dropdown-item @click="viewContainers(row)">容器列表</el-dropdown-item>
                  <el-dropdown-item @click="viewPorts(row)">端口信息</el-dropdown-item>
                  <el-dropdown-item divided @click="editServer(row)">编辑</el-dropdown-item>
                  <el-dropdown-item @click="deleteServer(row)" style="color: #f56c6c;">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        class="mt-4"
      />
    </el-card>

    <!-- 添加/编辑服务器对话框 -->
    <el-dialog v-model="showAddDialog" :title="editingServer ? '编辑服务器' : '添加服务器'" width="600px">
      <el-form :model="serverForm" label-width="100px" :rules="formRules" ref="formRef">
        <el-form-item label="名称" prop="name">
          <el-input v-model="serverForm.name" placeholder="服务器名称" />
        </el-form-item>
        <el-form-item label="IP 地址" prop="ip">
          <el-input v-model="serverForm.ip" placeholder="192.168.1.1" />
        </el-form-item>
        <el-form-item label="SSH 端口" prop="port">
          <el-input-number v-model="serverForm.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="serverForm.username" placeholder="root" />
        </el-form-item>
        <el-form-item label="认证方式" prop="authType">
          <el-radio-group v-model="serverForm.authType">
            <el-radio value="password">密码</el-radio>
            <el-radio value="key">密钥</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="serverForm.authType === 'password'">
          <el-input v-model="serverForm.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="私钥" prop="privateKey" v-else>
          <el-input v-model="serverForm.privateKey" type="textarea" :rows="5" placeholder="SSH 私钥" />
        </el-form-item>
        <el-form-item label="分组" prop="groupId">
          <el-select v-model="serverForm.groupId" placeholder="选择分组" clearable>
            <el-option v-for="group in groups" :key="group.id" :label="group.name" :value="group.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签" prop="tags">
          <el-select v-model="serverForm.tags" multiple placeholder="选择标签" allow-create>
            <el-option v-for="tag in tags" :key="tag" :label="tag" :value="tag" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="saveServer">保存</el-button>
      </template>
    </el-dialog>

    <!-- AI 分析对话框 -->
    <el-dialog v-model="showAIDialog" title="AI 智能分析" width="700px">
      <div v-if="aiResult" class="ai-result">
        <el-alert :title="aiResult.summary" type="info" show-icon :closable="false" class="mb-4" />
        
        <el-descriptions title="分析详情" :column="2" border class="mb-4">
          <el-descriptions-item label="健康评分">
            <el-progress :percentage="aiResult.healthScore" :color="getProgressColor(aiResult.healthScore)" />
          </el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <el-tag :type="getRiskType(aiResult.riskLevel)">{{ aiResult.riskLevel }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="CPU 分析" :span="2">{{ aiResult.cpuAnalysis }}</el-descriptions-item>
          <el-descriptions-item label="内存分析" :span="2">{{ aiResult.memoryAnalysis }}</el-descriptions-item>
          <el-descriptions-item label="磁盘分析" :span="2">{{ aiResult.diskAnalysis }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="aiResult.recommendations?.length">
          <h4 class="mb-2">优化建议:</h4>
          <el-timeline>
            <el-timeline-item v-for="(rec, idx) in aiResult.recommendations" :key="idx" :type="rec.priority === 'high' ? 'danger' : 'primary'">
              <strong>{{ rec.title }}</strong>
              <p class="text-gray-500 mt-1">{{ rec.description }}</p>
            </el-timeline-item>
          </el-timeline>
        </div>
      </div>
    </el-dialog>

    <!-- 监控详情对话框 -->
    <el-dialog v-model="showMetricsDialog" title="服务器监控" width="900px">
      <div v-if="currentServer">
        <el-row :gutter="20" class="mb-4">
          <el-col :span="8">
            <el-card shadow="never">
              <el-statistic title="CPU 使用率" :value="currentServer.cpuUsage" suffix="%" />
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="never">
              <el-statistic title="内存使用率" :value="currentServer.memUsage" suffix="%" />
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="never">
              <el-statistic title="磁盘使用率" :value="currentServer.diskUsage" suffix="%" />
            </el-card>
          </el-col>
        </el-row>
        <!-- 这里可以添加图表 -->
        <el-empty description="图表加载中..." />
      </div>
    </el-dialog>

    <!-- 执行命令对话框 -->
    <el-dialog v-model="showCommandDialog" title="执行命令" width="600px">
      <el-form :model="commandForm" label-width="80px">
        <el-form-item label="命令">
          <el-input v-model="commandForm.command" type="textarea" :rows="4" placeholder="输入要执行的命令" />
        </el-form-item>
      </el-form>
      <div v-if="commandResult" class="command-result">
        <pre>{{ commandResult }}</pre>
      </div>
      <template #footer>
        <el-button @click="showCommandDialog = false">关闭</el-button>
        <el-button type="primary" @click="runCommand" :loading="commandLoading">执行</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Monitor, CircleCheck, CircleClose, Warning, ArrowDown } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const servers = ref([])
const groups = ref([])
const tags = ref(['production', 'development', 'staging', 'database', 'web', 'api'])
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const showAddDialog = ref(false)
const showAIDialog = ref(false)
const showMetricsDialog = ref(false)
const showCommandDialog = ref(false)
const editingServer = ref(null)
const currentServer = ref(null)
const aiResult = ref<any>(null)
const commandResult = ref('')
const commandLoading = ref(false)

const stats = ref({
  total: 0,
  online: 0,
  offline: 0,
  warning: 0
})

const serverForm = ref({
  name: '',
  ip: '',
  port: 22,
  username: 'root',
  authType: 'password',
  password: '',
  privateKey: '',
  groupId: '',
  tags: []
})

const commandForm = ref({
  serverId: '',
  command: ''
})

const formRules = {
  name: [{ required: true, message: '请输入服务器名称', trigger: 'blur' }],
  ip: [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入SSH端口', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
}

const filteredServers = computed(() => {
  if (!searchKeyword.value) return servers.value
  const keyword = searchKeyword.value.toLowerCase()
  return servers.value.filter((s: any) => 
    s.name.toLowerCase().includes(keyword) ||
    s.ip.includes(keyword)
  )
})

const fetchServers = async () => {
  loading.value = true
  try {
    const res = await request.get('/servers', {
      params: { page: currentPage.value, pageSize: pageSize.value }
    })
    servers.value = res.data || []
    total.value = res.total || 0
    
    // 计算统计
    stats.value.total = total.value
    stats.value.online = servers.value.filter((s: any) => s.status === 'online').length
    stats.value.offline = servers.value.filter((s: any) => s.status === 'offline').length
    stats.value.warning = servers.value.filter((s: any) => s.cpuUsage > 80 || s.memUsage > 80).length
  } catch (error) {
    ElMessage.error('获取服务器列表失败')
  } finally {
    loading.value = false
  }
}

const fetchGroups = async () => {
  try {
    const res = await request.get('/groups')
    groups.value = res.data || []
  } catch (error) {
    console.error('获取分组失败', error)
  }
}

const saveServer = async () => {
  try {
    if (editingServer.value) {
      await request.put(`/servers/${editingServer.value.id}`, serverForm.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/servers', serverForm.value)
      ElMessage.success('添加成功')
    }
    showAddDialog.value = false
    resetForm()
    fetchServers()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const editServer = (server: any) => {
  editingServer.value = server
  serverForm.value = { ...server }
  showAddDialog.value = true
}

const deleteServer = async (server: any) => {
  try {
    await ElMessageBox.confirm('确定要删除该服务器吗？', '提示', { type: 'warning' })
    await request.delete(`/servers/${server.id}`)
    ElMessage.success('删除成功')
    fetchServers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const viewMetrics = (server: any) => {
  currentServer.value = server
  showMetricsDialog.value = true
}

const viewLogs = async (server: any) => {
  try {
    const res = await request.get(`/servers/${server.id}/logs`)
    ElMessage.info('日志功能开发中')
  } catch (error) {
    ElMessage.error('获取日志失败')
  }
}

const analyzeServer = async (server: any) => {
  try {
    const res = await request.post(`/servers/${server.id}/analyze`)
    aiResult.value = res.data
    showAIDialog.value = true
  } catch (error) {
    ElMessage.error('AI分析失败')
  }
}

const executeCommand = (server: any) => {
  commandForm.value.serverId = server.id
  commandResult.value = ''
  showCommandDialog.value = true
}

const runCommand = async () => {
  commandLoading.value = true
  try {
    const res = await request.post(`/servers/${commandForm.value.serverId}/command`, {
      command: commandForm.value.command
    })
    commandResult.value = res.data?.output || '执行成功'
  } catch (error) {
    commandResult.value = '执行失败'
  } finally {
    commandLoading.value = false
  }
}

const refreshStatus = async (server: any) => {
  try {
    await request.post(`/servers/${server.id}/refresh`)
    ElMessage.success('刷新成功')
    fetchServers()
  } catch (error) {
    ElMessage.error('刷新失败')
  }
}

const viewContainers = async (server: any) => {
  try {
    const res = await request.get(`/servers/${server.id}/containers`)
    ElMessage.info('容器列表功能开发中')
  } catch (error) {
    ElMessage.error('获取容器失败')
  }
}

const viewPorts = async (server: any) => {
  try {
    const res = await request.get(`/servers/${server.id}/ports`)
    ElMessage.info('端口信息功能开发中')
  } catch (error) {
    ElMessage.error('获取端口失败')
  }
}

const showServerDetail = (row: any) => {
  currentServer.value = row
  showMetricsDialog.value = true
}

const resetForm = () => {
  editingServer.value = null
  serverForm.value = {
    name: '',
    ip: '',
    port: 22,
    username: 'root',
    authType: 'password',
    password: '',
    privateKey: '',
    groupId: '',
    tags: []
  }
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    online: 'success',
    offline: 'danger',
    warning: 'warning'
  }
  return types[status] || 'info'
}

const getProgressColor = (value: number) => {
  if (value >= 90) return '#f56c6c'
  if (value >= 70) return '#e6a23c'
  return '#67c23a'
}

const getRiskType = (level: string) => {
  const types: Record<string, string> = {
    high: 'danger',
    medium: 'warning',
    low: 'success'
  }
  return types[level] || 'info'
}

onMounted(() => {
  fetchServers()
  fetchGroups()
})
</script>

<style scoped>
.servers-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
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

.mt-2 {
  margin-top: 8px;
}

.mb-2 {
  margin-bottom: 8px;
}

.text-gray-500 {
  color: #909399;
}

.ai-result {
  padding: 10px;
}

.command-result {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  margin-top: 10px;
  max-height: 300px;
  overflow: auto;
}

.command-result pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
