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
            <el-radio value="sshKey">SSH 密钥</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="serverForm.authType === 'password'">
          <el-input v-model="serverForm.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="SSH 密钥" prop="sshKeyId" v-if="serverForm.authType === 'sshKey'">
          <div class="ssh-key-selector">
            <el-select v-model="serverForm.sshKeyId" placeholder="选择 SSH 密钥" style="width: 100%;">
              <el-option v-for="key in sshKeys" :key="key.id" :label="`${key.name} (${key.filename})`" :value="key.id">
                <div style="display: flex; justify-content: space-between;">
                  <span>{{ key.name }}</span>
                  <span style="color: #999; font-size: 12px;">{{ key.filename }}</span>
                </div>
              </el-option>
            </el-select>
            <el-button type="primary" link @click="showAddSshKeyDialog = true" style="margin-left: 10px;">
              <el-icon><Plus /></el-icon>
              添加密钥
            </el-button>
          </div>
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

    <!-- 添加 SSH 密钥对话框 -->
    <el-dialog v-model="showAddSshKeyDialog" title="添加 SSH 密钥" width="550px">
      <el-form :model="sshKeyForm" label-width="100px" ref="sshKeyFormRef">
        <el-form-item label="密钥名称" required>
          <el-input v-model="sshKeyForm.name" placeholder="例如：生产服务器密钥" />
        </el-form-item>
        <el-form-item label="上传文件">
          <el-upload
            ref="pemUploadRef"
            :auto-upload="false"
            :show-file-list="false"
            accept=".pem,.key"
            :on-change="handlePemFileChange"
          >
            <template #trigger>
              <el-button type="primary">
                <el-icon><Upload /></el-icon>
                选择 .pem 文件
              </el-button>
            </template>
            <template #tip>
              <div class="el-upload__tip" style="margin-top: 8px;">
                支持 .pem 或 .key 格式的 SSH 私钥文件
              </div>
            </template>
          </el-upload>
          <div v-if="sshKeyForm.filename" class="file-info">
            <el-icon><Document /></el-icon>
            <span>{{ sshKeyForm.filename }}</span>
            <el-button type="danger" link @click="clearPemFile">清除</el-button>
          </div>
        </el-form-item>
        <el-form-item label="或输入内容">
          <el-input
            v-model="sshKeyForm.keyContent"
            type="textarea"
            :rows="6"
            placeholder="-----BEGIN RSA PRIVATE KEY-----&#10;...&#10;-----END RSA PRIVATE KEY-----"
          />
        </el-form-item>
        <el-form-item label="密钥密码">
          <el-input v-model="sshKeyForm.passphrase" type="password" placeholder="如果有密码的话" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="sshKeyForm.description" placeholder="密钥描述（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddSshKeyDialog = false">取消</el-button>
        <el-button type="primary" @click="saveSshKey" :loading="sshKeyLoading">保存</el-button>
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
        
        <el-descriptions title="服务器信息" :column="3" border>
          <el-descriptions-item label="名称">{{ currentServer.name }}</el-descriptions-item>
          <el-descriptions-item label="IP">{{ currentServer.ip }}</el-descriptions-item>
          <el-descriptions-item label="系统">{{ currentServer.os || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentServer.status)">{{ currentServer.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="分组">{{ currentServer.groupId || '-' }}</el-descriptions-item>
          <el-descriptions-item label="标签">
            <el-tag v-for="tag in (currentServer.tags || [])" :key="tag" size="small" class="mr-1">{{ tag }}</el-tag>
          </el-descriptions-item>
        </el-descriptions>
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

    <!-- 日志查看对话框 -->
    <el-dialog v-model="showLogsDialog" :title="`系统日志 - ${currentServer?.name || ''}`" width="900px">
      <div v-loading="logsLoading">
        <div class="logs-toolbar">
          <el-select v-model="logsFilter.type" placeholder="日志类型" style="width: 150px;" @change="loadLogs">
            <el-option label="系统日志" value="system" />
            <el-option label="安全日志" value="security" />
            <el-option label="应用日志" value="application" />
          </el-select>
          <el-input v-model="logsFilter.keyword" placeholder="搜索关键词" style="width: 200px; margin-left: 10px;" clearable @keyup.enter="loadLogs" />
          <el-button type="primary" style="margin-left: 10px;" @click="loadLogs">刷新</el-button>
        </div>
        
        <el-table :data="logsData" max-height="400" class="mt-3">
          <el-table-column prop="timestamp" label="时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.timestamp) }}
            </template>
          </el-table-column>
          <el-table-column prop="level" label="级别" width="100">
            <template #default="{ row }">
              <el-tag :type="getLogLevelType(row.level)" size="small">{{ row.level }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="source" label="来源" width="120" />
          <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
        </el-table>
        
        <el-empty v-if="logsData.length === 0 && !logsLoading" description="暂无日志数据" />
      </div>
      <template #footer>
        <el-button @click="showLogsDialog = false">关闭</el-button>
        <el-button type="primary" @click="exportLogs">导出日志</el-button>
      </template>
    </el-dialog>

    <!-- 容器列表对话框 -->
    <el-dialog v-model="showContainersDialog" :title="`Docker容器 - ${currentServer?.name || ''}`" width="900px">
      <div v-loading="containersLoading">
        <div class="toolbar">
          <el-button type="primary" @click="loadContainers">刷新</el-button>
        </div>
        
        <el-table :data="containersData" class="mt-3">
          <el-table-column prop="name" label="容器名称" min-width="180" />
          <el-table-column prop="image" label="镜像" width="200" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="row.status === 'running' ? 'success' : row.status === 'exited' ? 'info' : 'warning'" size="small">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="CPU %" width="100">
            <template #default="{ row }">
              {{ row.cpuPercent?.toFixed(1) || 0 }}%
            </template>
          </el-table-column>
          <el-table-column label="内存" width="120">
            <template #default="{ row }">
              {{ formatBytes(row.memUsage) }} / {{ formatBytes(row.memLimit) }}
            </template>
          </el-table-column>
          <el-table-column prop="ports" label="端口" width="180" show-overflow-tooltip />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="containerAction(row, 'start')" v-if="row.status !== 'running'" type="success">启动</el-button>
              <el-button size="small" @click="containerAction(row, 'stop')" v-if="row.status === 'running'" type="warning">停止</el-button>
              <el-button size="small" @click="containerAction(row, 'restart')" type="primary">重启</el-button>
            </template>
          </el-table-column>
        </el-table>
        
        <el-empty v-if="containersData.length === 0 && !containersLoading" description="暂无容器" />
      </div>
    </el-dialog>

    <!-- 端口信息对话框 -->
    <el-dialog v-model="showPortsDialog" :title="`端口信息 - ${currentServer?.name || ''}`" width="900px">
      <div v-loading="portsLoading">
        <div class="toolbar">
          <el-input v-model="portsFilter" placeholder="过滤端口" style="width: 200px;" clearable />
          <el-button type="primary" style="margin-left: 10px;" @click="loadPorts">刷新</el-button>
        </div>
        
        <el-table :data="filteredPorts" class="mt-3">
          <el-table-column prop="port" label="端口" width="100" />
          <el-table-column prop="protocol" label="协议" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ row.protocol || 'TCP' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="state" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.state === 'LISTEN' ? 'success' : 'info'" size="small">{{ row.state }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="process" label="进程" width="150" />
          <el-table-column prop="pid" label="PID" width="100" />
          <el-table-column prop="user" label="用户" width="100" />
          <el-table-column prop="service" label="服务" min-width="150" show-overflow-tooltip />
        </el-table>
        
        <el-empty v-if="portsData.length === 0 && !portsLoading" description="暂无端口信息" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Monitor, CircleCheck, CircleClose, Warning, ArrowDown, Upload, Document, Key } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const servers = ref([])
const groups = ref([])
const sshKeys = ref<any[]>([])
const tags = ref(['production', 'development', 'staging', 'database', 'web', 'api'])
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const showAddDialog = ref(false)
const showAIDialog = ref(false)
const showMetricsDialog = ref(false)
const showCommandDialog = ref(false)
const showLogsDialog = ref(false)
const showContainersDialog = ref(false)
const showPortsDialog = ref(false)
const showAddSshKeyDialog = ref(false)
const editingServer = ref(null)
const currentServer = ref(null)
const aiResult = ref<any>(null)
const commandResult = ref('')
const commandLoading = ref(false)
const sshKeyLoading = ref(false)

// 日志相关
const logsLoading = ref(false)
const logsData = ref<any[]>([])
const logsFilter = ref({
  type: 'system',
  keyword: ''
})

// 容器相关
const containersLoading = ref(false)
const containersData = ref<any[]>([])

// 端口相关
const portsLoading = ref(false)
const portsData = ref<any[]>([])
const portsFilter = ref('')

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
  sshKeyId: null as number | null,
  groupId: '',
  tags: []
})

const sshKeyForm = ref({
  name: '',
  filename: '',
  keyContent: '',
  passphrase: '',
  description: ''
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

const filteredPorts = computed(() => {
  if (!portsFilter.value) return portsData.value
  const keyword = portsFilter.value.toLowerCase()
  return portsData.value.filter((p: any) => 
    p.port?.toString().includes(keyword) ||
    p.process?.toLowerCase().includes(keyword) ||
    p.service?.toLowerCase().includes(keyword)
  )
})

const fetchServers = async () => {
  loading.value = true
  try {
    const res = await request.get('/servers', {
      params: { page: currentPage.value, pageSize: pageSize.value }
    })
    // 后端 OkWithPage 返回: { code: 0, data: { list: [], total: 10, page: 1, pageSize: 10 }, msg: "" }
    const data = res.data || {}
    servers.value = data.list || []
    total.value = data.total || 0
    
    // 计算统计
    stats.value.total = total.value
    stats.value.online = servers.value.filter((s: any) => s.status === 'online').length
    stats.value.offline = servers.value.filter((s: any) => s.status === 'offline').length
    stats.value.warning = servers.value.filter((s: any) => s.cpuUsage > 80 || s.memUsage > 80).length
  } catch (error) {
    console.error('获取服务器列表失败', error)
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

// 获取 SSH 密钥列表
const fetchSshKeys = async () => {
  try {
    const res = await request.get('/ssh-keys')
    sshKeys.value = res.data?.list || []
  } catch (error) {
    console.error('获取 SSH 密钥列表失败', error)
  }
}

// 处理 .pem 文件上传
const handlePemFileChange = (file: any) => {
  const rawFile = file.raw
  if (!rawFile) return
  
  // 验证文件扩展名
  if (!rawFile.name.endsWith('.pem') && !rawFile.name.endsWith('.key')) {
    ElMessage.error('请上传 .pem 或 .key 格式的文件')
    return
  }
  
  // 读取文件内容
  const reader = new FileReader()
  reader.onload = (event) => {
    const content = event.target?.result as string
    sshKeyForm.value.filename = rawFile.name
    sshKeyForm.value.keyContent = content
    // 自动填充名称
    if (!sshKeyForm.value.name) {
      sshKeyForm.value.name = rawFile.name.replace(/\.(pem|key)$/, '')
    }
  }
  reader.readAsText(rawFile)
}

// 清除 .pem 文件
const clearPemFile = () => {
  sshKeyForm.value.filename = ''
  sshKeyForm.value.keyContent = ''
}

// 保存 SSH 密钥
const saveSshKey = async () => {
  if (!sshKeyForm.value.name) {
    ElMessage.error('请输入密钥名称')
    return
  }
  if (!sshKeyForm.value.keyContent) {
    ElMessage.error('请上传 .pem 文件或输入密钥内容')
    return
  }
  
  sshKeyLoading.value = true
  try {
    await request.post('/ssh-keys', sshKeyForm.value)
    ElMessage.success('SSH 密钥添加成功')
    showAddSshKeyDialog.value = false
    // 重置表单
    sshKeyForm.value = {
      name: '',
      filename: '',
      keyContent: '',
      passphrase: '',
      description: ''
    }
    // 刷新密钥列表
    await fetchSshKeys()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '添加失败')
  } finally {
    sshKeyLoading.value = false
  }
}

const saveServer = async () => {
  try {
    // 构建请求数据，映射字段名
    const data: any = {
      name: serverForm.value.name,
      host: serverForm.value.ip,  // 前端使用 ip，后端使用 host
      port: serverForm.value.port,
      user: serverForm.value.username,  // 前端使用 username，后端使用 user
      authType: serverForm.value.authType,
      groupId: serverForm.value.groupId,
      tags: serverForm.value.tags
    }
    
    // 根据认证方式添加认证信息
    if (serverForm.value.authType === 'password') {
      data.password = serverForm.value.password
    } else if (serverForm.value.authType === 'sshKey') {
      data.sshKeyId = serverForm.value.sshKeyId
    }
    
    if (editingServer.value) {
      await request.put(`/servers/${editingServer.value.id}`, data)
      ElMessage.success('更新成功')
    } else {
      await request.post('/servers', data)
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
  serverForm.value = {
    name: server.name || '',
    ip: server.host || server.ip || '',
    port: server.port || 22,
    username: server.user || server.username || 'root',
    authType: server.authType || 'password',
    password: '',
    privateKey: '',
    sshKeyId: server.sshKeyId || server.sshKey?.id || null,
    groupId: server.groupId || '',
    tags: server.tags || []
  }
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

// 日志功能
const viewLogs = async (server: any) => {
  currentServer.value = server
  showLogsDialog.value = true
  await loadLogs()
}

const loadLogs = async () => {
  if (!currentServer.value) return
  logsLoading.value = true
  try {
    const res = await request.get(`/servers/${currentServer.value.id}/logs`, {
      params: {
        type: logsFilter.value.type,
        keyword: logsFilter.value.keyword
      }
    })
    logsData.value = res.data || []
  } catch (error) {
    console.error('获取日志失败', error)
    // 使用模拟数据
    logsData.value = generateMockLogs()
  } finally {
    logsLoading.value = false
  }
}

const generateMockLogs = () => {
  const levels = ['INFO', 'WARN', 'ERROR', 'DEBUG']
  const sources = ['systemd', 'kernel', 'sshd', 'nginx', 'mysql']
  const messages = [
    '服务启动完成',
    '连接建立成功',
    '检测到异常登录尝试',
    '内存使用率超过80%',
    '定时任务执行完成',
    '配置文件已重新加载',
    'SSL证书即将过期',
    '数据库连接池已满'
  ]
  
  return Array.from({ length: 20 }, (_, i) => ({
    timestamp: new Date(Date.now() - i * 3600000).toISOString(),
    level: levels[Math.floor(Math.random() * levels.length)],
    source: sources[Math.floor(Math.random() * sources.length)],
    message: messages[Math.floor(Math.random() * messages.length)]
  }))
}

const exportLogs = () => {
  const content = logsData.value.map(log => 
    `${log.timestamp} [${log.level}] ${log.source}: ${log.message}`
  ).join('\n')
  
  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `logs_${currentServer.value?.name}_${new Date().toISOString().slice(0, 10)}.txt`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('日志已导出')
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

// 容器功能
const viewContainers = async (server: any) => {
  currentServer.value = server
  showContainersDialog.value = true
  await loadContainers()
}

const loadContainers = async () => {
  if (!currentServer.value) return
  containersLoading.value = true
  try {
    const res = await request.get(`/servers/${currentServer.value.id}/containers`)
    containersData.value = res.data || []
  } catch (error) {
    console.error('获取容器失败', error)
    // 使用模拟数据
    containersData.value = generateMockContainers()
  } finally {
    containersLoading.value = false
  }
}

const generateMockContainers = () => {
  return [
    { id: 'abc123', name: 'nginx-proxy', image: 'nginx:1.24', status: 'running', cpuPercent: 2.5, memUsage: 52428800, memLimit: 104857600, ports: '80:80, 443:443' },
    { id: 'def456', name: 'mysql-db', image: 'mysql:8.0', status: 'running', cpuPercent: 15.3, memUsage: 524288000, memLimit: 1073741824, ports: '3306:3306' },
    { id: 'ghi789', name: 'redis-cache', image: 'redis:7-alpine', status: 'running', cpuPercent: 1.2, memUsage: 31457280, memLimit: 104857600, ports: '6379:6379' },
    { id: 'jkl012', name: 'app-backup', image: 'backup-tool:latest', status: 'exited', cpuPercent: 0, memUsage: 0, memLimit: 0, ports: '-' }
  ]
}

const containerAction = async (container: any, action: string) => {
  try {
    await request.post(`/servers/${currentServer.value.id}/containers/${container.id}/${action}`)
    ElMessage.success(`容器${action}命令已发送`)
    await loadContainers()
  } catch (error) {
    ElMessage.error(`容器${action}失败`)
  }
}

// 端口功能
const viewPorts = async (server: any) => {
  currentServer.value = server
  showPortsDialog.value = true
  await loadPorts()
}

const loadPorts = async () => {
  if (!currentServer.value) return
  portsLoading.value = true
  try {
    const res = await request.get(`/servers/${currentServer.value.id}/ports`)
    portsData.value = res.data || []
  } catch (error) {
    console.error('获取端口失败', error)
    // 使用模拟数据
    portsData.value = generateMockPorts()
  } finally {
    portsLoading.value = false
  }
}

const generateMockPorts = () => {
  return [
    { port: 22, protocol: 'TCP', state: 'LISTEN', process: 'sshd', pid: 1234, user: 'root', service: 'SSH Server' },
    { port: 80, protocol: 'TCP', state: 'LISTEN', process: 'nginx', pid: 2345, user: 'www-data', service: 'HTTP Server' },
    { port: 443, protocol: 'TCP', state: 'LISTEN', process: 'nginx', pid: 2345, user: 'www-data', service: 'HTTPS Server' },
    { port: 3306, protocol: 'TCP', state: 'LISTEN', process: 'mysqld', pid: 3456, user: 'mysql', service: 'MySQL Server' },
    { port: 6379, protocol: 'TCP', state: 'LISTEN', process: 'redis-server', pid: 4567, user: 'redis', service: 'Redis Server' },
    { port: 9090, protocol: 'TCP', state: 'LISTEN', process: 'prometheus', pid: 5678, user: 'prometheus', service: 'Prometheus' }
  ]
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
    sshKeyId: null,
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

const getLogLevelType = (level: string) => {
  const types: Record<string, string> = {
    INFO: 'info',
    WARN: 'warning',
    ERROR: 'danger',
    DEBUG: ''
  }
  return types[level] || ''
}

const formatTime = (time: string) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN')
}

const formatBytes = (bytes: number) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return `${bytes.toFixed(1)} ${units[i]}`
}

onMounted(() => {
  fetchServers()
  fetchGroups()
  fetchSshKeys()
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

.mt-3 {
  margin-top: 12px;
}

.mt-2 {
  margin-top: 8px;
}

.mb-2 {
  margin-bottom: 8px;
}

.mr-1 {
  margin-right: 4px;
}

.text-gray-500 {
  color: #909399;
}

.ai-result {
  padding: 10px;
}

.command-result {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 15px;
  border-radius: 4px;
  margin-top: 10px;
  max-height: 300px;
  overflow: auto;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
}

.command-result pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.logs-toolbar,
.toolbar {
  display: flex;
  align-items: center;
}

.ssh-key-selector {
  display: flex;
  align-items: center;
  width: 100%;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 14px;
}

.file-info .el-icon {
  color: #409eff;
}
</style>
