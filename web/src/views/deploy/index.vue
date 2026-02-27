<template>
  <div class="deploy-page">
    <el-row :gutter="20">
      <!-- 左侧：步骤导航 -->
      <el-col :span="6">
        <el-card>
          <template #header>
            <span>部署向导</span>
          </template>
          <el-steps :active="currentStep" direction="vertical" finish-status="success">
            <el-step title="上传项目" description="上传或指定项目路径" />
            <el-step title="项目分析" description="分析项目类型和依赖" />
            <el-step title="选择服务器" description="选择部署服务器" />
            <el-step title="生成方案" description="生成部署配置" />
            <el-step title="执行部署" description="一键部署到服务器" />
          </el-steps>
        </el-card>
      </el-col>

      <!-- 右侧：内容区 -->
      <el-col :span="18">
        <!-- 步骤1: 上传项目 -->
        <el-card v-if="currentStep === 0" class="mb-4">
          <template #header>
            <span>上传项目</span>
          </template>
          
          <el-tabs v-model="uploadType">
            <el-tab-pane label="上传文件" name="file">
              <el-upload
                drag
                action="/deploy/upload"
                :headers="uploadHeaders"
                :on-success="handleUploadSuccess"
                :on-error="handleUploadError"
                accept=".zip,.tar.gz,.tgz"
              >
                <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                <div class="el-upload__text">
                  拖拽项目文件到这里，或 <em>点击上传</em>
                </div>
                <template #tip>
                  <div class="el-upload__tip">
                    支持 .zip, .tar.gz 格式的项目文件
                  </div>
                </template>
              </el-upload>
            </el-tab-pane>
            
            <el-tab-pane label="指定路径" name="path">
              <el-form :model="pathForm" label-width="100px">
                <el-form-item label="项目路径">
                  <el-input v-model="pathForm.path" placeholder="/path/to/your/project" />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="analyzeByPath">分析项目</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </el-card>

        <!-- 步骤2: 项目分析结果 -->
        <el-card v-if="currentStep === 1" class="mb-4">
          <template #header>
            <div class="card-header">
              <span>项目分析结果</span>
              <el-button type="primary" @click="nextStep">下一步</el-button>
            </div>
          </template>
          
          <el-descriptions :column="2" border v-if="projectAnalysis">
            <el-descriptions-item label="项目名称">{{ projectAnalysis.projectName }}</el-descriptions-item>
            <el-descriptions-item label="项目类型">
              <el-tag>{{ projectAnalysis.projectType }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="技术栈">
              <el-tag v-for="tech in parseTechStack(projectAnalysis.techStacks)" :key="tech" class="mr-1">
                {{ tech }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="服务数量">
              {{ parseServices(projectAnalysis.services).length }} 个
            </el-descriptions-item>
            <el-descriptions-item label="最小资源">
              CPU: {{ projectAnalysis.minCpu }}核, 内存: {{ projectAnalysis.minMemory }}MB, 磁盘: {{ projectAnalysis.minDisk }}GB
            </el-descriptions-item>
            <el-descriptions-item label="推荐资源">
              CPU: {{ projectAnalysis.recCpu }}核, 内存: {{ projectAnalysis.recMemory }}MB, 磁盘: {{ projectAnalysis.recDisk }}GB
            </el-descriptions-item>
            <el-descriptions-item label="需要集群">
              <el-tag :type="projectAnalysis.needCluster ? 'success' : 'info'">
                {{ projectAnalysis.needCluster ? '是' : '否' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="集群规模">
              {{ projectAnalysis.clusterSize || 1 }} 节点
            </el-descriptions-item>
            <el-descriptions-item label="需要负载均衡">
              <el-tag :type="projectAnalysis.needLb ? 'success' : 'info'">
                {{ projectAnalysis.needLb ? '是' : '否' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="需要数据库集群">
              <el-tag :type="projectAnalysis.needDbCluster ? 'success' : 'info'">
                {{ projectAnalysis.needDbCluster ? '是' : '否' }}
              </el-tag>
            </el-descriptions-item>
          </el-descriptions>

          <!-- 服务列表 -->
          <div class="mt-4" v-if="parseServices(projectAnalysis.services).length > 0">
            <h4>服务列表</h4>
            <el-table :data="parseServices(projectAnalysis.services)" size="small">
              <el-table-column prop="name" label="服务名称" />
              <el-table-column prop="type" label="类型" />
              <el-table-column prop="port" label="端口" />
              <el-table-column prop="replicas" label="副本数" />
            </el-table>
          </div>

          <!-- AI 建议 -->
          <div class="mt-4" v-if="projectAnalysis.aiSuggestion">
            <h4>AI 建议</h4>
            <el-alert type="info" :closable="false">
              {{ projectAnalysis.aiSuggestion }}
            </el-alert>
          </div>
        </el-card>

        <!-- 步骤3: 选择服务器 -->
        <el-card v-if="currentStep === 2" class="mb-4">
          <template #header>
            <div class="card-header">
              <span>选择服务器</span>
              <div>
                <el-button @click="analyzeServers" :loading="analyzing">刷新服务器状态</el-button>
                <el-button type="primary" @click="findBestServers" :disabled="!projectAnalysis">智能推荐</el-button>
              </div>
            </div>
          </template>
          
          <el-table :data="serverMatches" v-loading="findingServers" @selection-change="handleServerSelect">
            <el-table-column type="selection" width="50" />
            <el-table-column prop="serverId" label="服务器ID" width="100" />
            <el-table-column label="服务器名称">
              <template #default="{ row }">
                {{ getServerName(row.serverId) }}
              </template>
            </el-table-column>
            <el-table-column prop="role" label="推荐角色" width="120" />
            <el-table-column prop="score" label="匹配分数" width="120">
              <template #default="{ row }">
                <el-progress :percentage="row.score" :stroke-width="10" />
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="推荐原因" show-overflow-tooltip />
          </el-table>
          
          <div class="mt-4">
            <el-button @click="prevStep">上一步</el-button>
            <el-button type="primary" @click="nextStep" :disabled="selectedServers.length === 0">
              下一步
            </el-button>
          </div>
        </el-card>

        <!-- 步骤4: 生成方案 -->
        <el-card v-if="currentStep === 3" class="mb-4">
          <template #header>
            <div class="card-header">
              <span>部署方案</span>
              <div>
                <el-button @click="generatePlan" :loading="generating" :disabled="!projectAnalysis">
                  生成方案
                </el-button>
                <el-button type="primary" @click="nextStep" :disabled="!deployPlan">
                  确认方案
                </el-button>
              </div>
            </div>
          </template>

          <div v-if="deployPlan">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="方案名称">{{ deployPlan.name }}</el-descriptions-item>
              <el-descriptions-item label="部署类型">
                <el-tag>{{ deployPlan.planType }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="服务数量">
                {{ parseServices(deployPlan.services).length }} 个
              </el-descriptions-item>
              <el-descriptions-item label="预估成本">
                ${{ deployPlan.estimatedCost?.toFixed(2) }}/月
              </el-descriptions-item>
            </el-descriptions>

            <!-- 服务拓扑 -->
            <div class="mt-4">
              <h4>服务拓扑</h4>
              <div id="topology-container" style="height: 300px; background: #f5f7fa; border-radius: 4px;">
                <!-- 这里可以集成拓扑图可视化 -->
                <el-table :data="parseAssignments(deployPlan.serverAssignments)" size="small">
                  <el-table-column prop="serverName" label="服务器" />
                  <el-table-column prop="role" label="角色" />
                  <el-table-column label="服务">
                    <template #default="{ row }">
                      <el-tag v-for="svc in row.services" :key="svc" class="mr-1" size="small">{{ svc }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="资源">
                    <template #default="{ row }">
                      CPU: {{ row.resources?.cpu }}核, 内存: {{ row.resources?.memory }}MB
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </div>

            <!-- AI 建议 -->
            <div class="mt-4" v-if="deployPlan.aiSuggestion">
              <h4>AI 优化建议</h4>
              <el-alert type="success" :closable="false">
                {{ deployPlan.aiSuggestion }}
              </el-alert>
            </div>

            <!-- 配置预览 -->
            <div class="mt-4">
              <el-button @click="previewConfigs" :loading="previewing">预览配置文件</el-button>
            </div>
          </div>
          <el-empty v-else description="请先生成部署方案" />
          
          <div class="mt-4">
            <el-button @click="prevStep">上一步</el-button>
          </div>
        </el-card>

        <!-- 步骤5: 执行部署 -->
        <el-card v-if="currentStep === 4" class="mb-4">
          <template #header>
            <div class="card-header">
              <span>执行部署</span>
              <div>
                <el-button @click="prevStep">上一步</el-button>
                <el-button type="primary" @click="executeDeploy" :loading="executing" :disabled="!deployPlan">
                  开始部署
                </el-button>
              </div>
            </div>
          </template>

          <div v-if="deployTask">
            <!-- 部署进度 -->
            <el-progress :percentage="deployTask.progress" :status="getTaskStatus(deployTask.status)" />
            <p class="mt-2">当前步骤: {{ deployTask.currentStep }}</p>
            
            <!-- 部署步骤 -->
            <div class="mt-4">
              <h4>执行步骤</h4>
              <el-timeline>
                <el-timeline-item
                  v-for="step in taskSteps"
                  :key="step.id"
                  :type="getStepType(step.status)"
                  :timestamp="formatDate(step.createdAt)">
                  <p><strong>{{ step.name }}</strong></p>
                  <p v-if="step.serverName">服务器: {{ step.serverName }}</p>
                  <p v-if="step.output" class="output">{{ step.output }}</p>
                  <p v-if="step.error" class="error">{{ step.error }}</p>
                </el-timeline-item>
              </el-timeline>
            </div>
            
            <!-- 执行日志 -->
            <div class="mt-4">
              <h4>执行日志</h4>
              <el-input type="textarea" :rows="10" :model-value="deployTask.logs" readonly />
            </div>
            
            <!-- 操作按钮 -->
            <div class="mt-4">
              <el-button 
                v-if="deployTask.status === 'running'" 
                type="warning" 
                @click="pauseDeploy">
                暂停
              </el-button>
              <el-button 
                v-if="deployTask.status === 'paused'" 
                type="success" 
                @click="resumeDeploy">
                继续
              </el-button>
              <el-button 
                v-if="deployTask.status === 'failed'" 
                type="danger" 
                @click="rollbackDeploy">
                回滚
              </el-button>
            </div>
          </div>
          <el-empty v-else description="点击"开始部署"执行部署任务" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 配置预览对话框 -->
    <el-dialog v-model="showConfigPreview" title="配置文件预览" width="80%">
      <el-tabs>
        <el-tab-pane v-for="cfg in configPreviews" :key="cfg.serverId" :label="cfg.serverName">
          <el-collapse>
            <el-collapse-item v-for="file in cfg.configs" :key="file.path" :title="file.path">
              <pre class="config-content">{{ file.content }}</pre>
            </el-collapse-item>
          </el-collapse>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'

const currentStep = ref(0)
const uploadType = ref('file')
const analyzing = ref(false)
const findingServers = ref(false)
const generating = ref(false)
const previewing = ref(false)
const executing = ref(false)
const showConfigPreview = ref(false)

const pathForm = ref({ path: '' })
const projectAnalysis = ref<any>(null)
const serverMatches = ref<any[]>([])
const selectedServers = ref<any[]>([])
const serverCapabilities = ref<any[]>([])
const deployPlan = ref<any>(null)
const deployTask = ref<any>(null)
const taskSteps = ref<any[]>([])
const configPreviews = ref<any[]>([])

const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem('token')}`
}))

// 上传成功
const handleUploadSuccess = (response: any) => {
  if (response.code === 0) {
    projectAnalysis.value = response.data.analysis
    ElMessage.success('项目上传并分析成功')
    currentStep.value = 1
  }
}

// 上传失败
const handleUploadError = () => {
  ElMessage.error('上传失败')
}

// 通过路径分析
const analyzeByPath = async () => {
  if (!pathForm.value.path) {
    ElMessage.warning('请输入项目路径')
    return
  }
  
  analyzing.value = true
  try {
    const res = await request.post('/deploy/analyze', { path: pathForm.value.path })
    projectAnalysis.value = res.data
    ElMessage.success('项目分析成功')
    currentStep.value = 1
  } catch (error) {
    ElMessage.error('项目分析失败')
  } finally {
    analyzing.value = false
  }
}

// 分析服务器
const analyzeServers = async () => {
  analyzing.value = true
  try {
    const res = await request.get('/deploy/servers/analyze')
    // 更新服务器能力
  } catch (error) {
    console.error('分析服务器失败', error)
  } finally {
    analyzing.value = false
  }
}

// 查找最佳服务器
const findBestServers = async () => {
  if (!projectAnalysis.value) return
  
  findingServers.value = true
  try {
    const res = await request.post('/deploy/servers/find-best', {
      minCpu: projectAnalysis.value.minCpu,
      minMemory: projectAnalysis.value.minMemory,
      minDisk: projectAnalysis.value.minDisk,
      needDocker: true
    })
    serverMatches.value = res.data || []
  } catch (error) {
    ElMessage.error('查找服务器失败')
  } finally {
    findingServers.value = false
  }
}

// 获取服务器能力
const fetchServerCapabilities = async () => {
  try {
    const res = await request.get('/deploy/servers/capabilities')
    serverCapabilities.value = res.data || []
  } catch (error) {
    console.error('获取服务器能力失败', error)
  }
}

// 处理服务器选择
const handleServerSelect = (selection: any[]) => {
  selectedServers.value = selection
}

// 生成部署方案
const generatePlan = async () => {
  if (!projectAnalysis.value) return
  
  generating.value = true
  try {
    const res = await request.post('/deploy/plans', {
      projectAnalysisId: projectAnalysis.value.id
    })
    deployPlan.value = res.data
    ElMessage.success('部署方案生成成功')
  } catch (error) {
    ElMessage.error('生成方案失败')
  } finally {
    generating.value = false
  }
}

// 预览配置
const previewConfigs = async () => {
  if (!deployPlan.value) return
  
  previewing.value = true
  try {
    const res = await request.get(`/deploy/plans/${deployPlan.value.id}/preview`)
    configPreviews.value = res.data || []
    showConfigPreview.value = true
  } catch (error) {
    ElMessage.error('获取配置失败')
  } finally {
    previewing.value = false
  }
}

// 执行部署
const executeDeploy = async () => {
  if (!deployPlan.value) return
  
  executing.value = true
  try {
    const res = await request.post(`/deploy/plans/${deployPlan.value.id}/execute`)
    deployTask.value = res.data
    ElMessage.success('部署任务已启动')
    
    // 轮询任务状态
    pollTaskStatus()
  } catch (error) {
    ElMessage.error('启动部署失败')
  } finally {
    executing.value = false
  }
}

// 轮询任务状态
const pollTaskStatus = async () => {
  if (!deployTask.value) return
  
  const poll = async () => {
    try {
      const res = await request.get(`/deploy/tasks/${deployTask.value.id}`)
      deployTask.value = res.data
      
      // 获取步骤
      const stepsRes = await request.get(`/deploy/tasks/${deployTask.value.id}/steps`)
      taskSteps.value = stepsRes.data || []
      
      // 如果任务还在运行，继续轮询
      if (deployTask.value.status === 'running') {
        setTimeout(poll, 2000)
      }
    } catch (error) {
      console.error('获取任务状态失败', error)
    }
  }
  
  poll()
}

// 暂停部署
const pauseDeploy = async () => {
  try {
    await request.post(`/deploy/tasks/${deployTask.value.id}/pause`)
    ElMessage.success('已暂停')
    deployTask.value.status = 'paused'
  } catch (error) {
    ElMessage.error('暂停失败')
  }
}

// 恢复部署
const resumeDeploy = async () => {
  try {
    await request.post(`/deploy/tasks/${deployTask.value.id}/resume`)
    ElMessage.success('已恢复')
    deployTask.value.status = 'running'
    pollTaskStatus()
  } catch (error) {
    ElMessage.error('恢复失败')
  }
}

// 回滚部署
const rollbackDeploy = async () => {
  try {
    await request.post(`/deploy/tasks/${deployTask.value.id}/rollback`)
    ElMessage.warning('已回滚')
  } catch (error) {
    ElMessage.error('回滚失败')
  }
}

// 辅助函数
const parseTechStack = (json: string) => {
  try {
    return JSON.parse(json) || []
  } catch {
    return []
  }
}

const parseServices = (json: string) => {
  try {
    return JSON.parse(json) || []
  } catch {
    return []
  }
}

const parseAssignments = (json: string) => {
  try {
    return JSON.parse(json) || []
  } catch {
    return []
  }
}

const getServerName = (id: number) => {
  const cap = serverCapabilities.value.find(c => c.serverId === id)
  return cap?.server?.name || `服务器 ${id}`
}

const getTaskStatus = (status: string) => {
  const statusMap: Record<string, string> = {
    completed: 'success',
    failed: 'exception',
    running: ''
  }
  return statusMap[status] || ''
}

const getStepType = (status: string) => {
  const typeMap: Record<string, string> = {
    success: 'success',
    running: 'primary',
    failed: 'danger',
    pending: 'info'
  }
  return typeMap[status] || 'info'
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString()
}

const nextStep = () => {
  currentStep.value++
}

const prevStep = () => {
  currentStep.value--
}

onMounted(() => {
  fetchServerCapabilities()
})
</script>

<style scoped>
.deploy-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mb-4 {
  margin-bottom: 16px;
}

.mt-2 {
  margin-top: 8px;
}

.mt-4 {
  margin-top: 16px;
}

.mr-1 {
  margin-right: 4px;
}

.config-content {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 4px;
  overflow-x: auto;
  font-family: monospace;
  font-size: 12px;
  white-space: pre-wrap;
}

.output {
  color: #606266;
  font-size: 12px;
}

.error {
  color: #f56c6c;
  font-size: 12px;
}
</style>
