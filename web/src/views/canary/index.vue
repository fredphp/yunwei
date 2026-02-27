<template>
  <div class="canary-page">
    <el-card class="mb-4">
      <template #header>
        <div class="card-header">
          <span>灰度发布管理</span>
          <el-button type="primary" @click="showStartDialog = true">
            <el-icon><Plus /></el-icon>
            新建发布
          </el-button>
        </div>
      </template>
      
      <el-table :data="releases" v-loading="loading">
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="serviceName" label="服务名称" />
        <el-table-column label="版本" width="200">
          <template #default="{ row }">
            {{ row.currentVersion }} → {{ row.newVersion }}
          </template>
        </el-table-column>
        <el-table-column prop="strategy" label="策略" width="100">
          <template #default="{ row }">
            <el-tag>{{ row.strategy }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="进度" width="150">
          <template #default="{ row }">
            <el-progress 
              :percentage="(row.canaryWeight || 0)" 
              :status="getProgressStatus(row.status)"
            />
            <small>{{ row.currentStep }}/{{ row.totalSteps }} 步</small>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280">
          <template #default="{ row }">
            <el-button-group>
              <el-button size="small" type="primary" 
                :disabled="row.status !== 'running' && row.status !== 'paused'"
                @click="promoteRelease(row)">
                推进
              </el-button>
              <el-button size="small" type="success" 
                :disabled="row.status !== 'running'"
                @click="completeRelease(row)">
                完成
              </el-button>
              <el-button size="small" type="warning"
                :disabled="row.status !== 'running'"
                @click="pauseRelease(row)">
                暂停
              </el-button>
              <el-button size="small" type="danger"
                :disabled="row.status !== 'running' && row.status !== 'paused'"
                @click="rollbackRelease(row)">
                回滚
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 发布详情 -->
    <el-card class="mb-4" v-if="currentRelease">
      <template #header>
        <span>发布详情 - {{ currentRelease.serviceName }}</span>
      </template>
      
      <el-descriptions :column="3" border>
        <el-descriptions-item label="命名空间">{{ currentRelease.namespace }}</el-descriptions-item>
        <el-descriptions-item label="服务名称">{{ currentRelease.serviceName }}</el-descriptions-item>
        <el-descriptions-item label="策略">{{ currentRelease.strategy }}</el-descriptions-item>
        <el-descriptions-item label="当前版本">{{ currentRelease.currentVersion }}</el-descriptions-item>
        <el-descriptions-item label="目标版本">{{ currentRelease.newVersion }}</el-descriptions-item>
        <el-descriptions-item label="新镜像">{{ currentRelease.newImage }}</el-descriptions-item>
        <el-descriptions-item label="当前权重">{{ currentRelease.canaryWeight }}%</el-descriptions-item>
        <el-descriptions-item label="当前步骤">{{ currentRelease.currentStep }}/{{ currentRelease.totalSteps }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentRelease.status)">{{ currentRelease.status }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 监控指标 -->
      <div class="metrics mt-4">
        <h4>实时监控指标</h4>
        <el-row :gutter="20">
          <el-col :span="6">
            <el-statistic title="错误率" :value="currentRelease.currentErrorRate" suffix="%" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="平均延迟" :value="currentRelease.currentLatency" suffix="ms" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="成功率" :value="currentRelease.currentSuccessRate" suffix="%" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="AI 置信度" :value="(currentRelease.aiConfidence * 100).toFixed(1)" suffix="%" />
          </el-col>
        </el-row>
      </div>

      <!-- 步骤历史 -->
      <div class="steps mt-4">
        <h4>发布步骤</h4>
        <el-timeline>
          <el-timeline-item 
            v-for="step in releaseSteps" 
            :key="step.id"
            :type="step.status === 'success' ? 'success' : 'primary'"
            :timestamp="formatDate(step.createdAt)">
            <p>步骤 {{ step.stepNum }}: 权重 {{ step.weight }}%</p>
            <p>错误率: {{ step.errorRate }}%, 延迟: {{ step.latency }}ms</p>
          </el-timeline-item>
        </el-timeline>
      </div>
    </el-card>

    <!-- 新建发布对话框 -->
    <el-dialog v-model="showStartDialog" title="新建灰度发布" width="600px">
      <el-form :model="releaseForm" label-width="120px">
        <el-form-item label="集群">
          <el-select v-model="releaseForm.clusterId" placeholder="选择集群">
            <el-option label="生产集群" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="命名空间">
          <el-input v-model="releaseForm.namespace" placeholder="default" />
        </el-form-item>
        <el-form-item label="服务名称">
          <el-input v-model="releaseForm.serviceName" placeholder="my-service" />
        </el-form-item>
        <el-form-item label="新镜像">
          <el-input v-model="releaseForm.newImage" placeholder="registry/image:tag" />
        </el-form-item>
        <el-form-item label="发布策略">
          <el-select v-model="releaseForm.config.strategy">
            <el-option label="金丝雀发布" value="canary" />
            <el-option label="蓝绿发布" value="bluegreen" />
            <el-option label="A/B 测试" value="ab" />
          </el-select>
        </el-form-item>
        <el-form-item label="总步骤数">
          <el-input-number v-model="releaseForm.config.totalSteps" :min="2" :max="10" />
        </el-form-item>
        <el-form-item label="每步权重增量">
          <el-input-number v-model="releaseForm.config.weightStep" :min="5" :max="50" />
        </el-form-item>
        <el-form-item label="错误率阈值(%)">
          <el-input-number v-model="releaseForm.config.errorRateThreshold" :min="0" :max="100" />
        </el-form-item>
        <el-form-item label="延迟阈值(ms)">
          <el-input-number v-model="releaseForm.config.latencyThreshold" :min="100" :max="10000" />
        </el-form-item>
        <el-form-item label="自动推进">
          <el-switch v-model="releaseForm.config.autoPromote" />
        </el-form-item>
        <el-form-item label="自动回滚">
          <el-switch v-model="releaseForm.config.autoRollback" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showStartDialog = false">取消</el-button>
        <el-button type="primary" @click="startRelease">开始发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const releases = ref([])
const currentRelease = ref<any>(null)
const releaseSteps = ref([])
const showStartDialog = ref(false)

const releaseForm = ref({
  clusterId: 1,
  namespace: 'default',
  serviceName: '',
  newImage: '',
  config: {
    strategy: 'canary',
    totalSteps: 5,
    weightStep: 20,
    errorRateThreshold: 5,
    latencyThreshold: 500,
    successRateThreshold: 95,
    autoPromote: false,
    autoRollback: true
  }
})

const fetchReleases = async () => {
  loading.value = true
  try {
    const res = await request.get('/api/v1/canary/releases')
    releases.value = res.data || []
  } catch (error) {
    ElMessage.error('获取发布列表失败')
  } finally {
    loading.value = false
  }
}

const fetchReleaseDetail = async (id: number) => {
  try {
    const [releaseRes, stepsRes] = await Promise.all([
      request.get(`/api/v1/canary/releases/${id}`),
      request.get(`/api/v1/canary/releases/${id}/steps`)
    ])
    currentRelease.value = releaseRes.data
    releaseSteps.value = stepsRes.data || []
  } catch (error) {
    console.error('获取详情失败', error)
  }
}

const startRelease = async () => {
  try {
    await request.post('/api/v1/canary/releases', releaseForm.value)
    ElMessage.success('发布已启动')
    showStartDialog.value = false
    fetchReleases()
  } catch (error) {
    ElMessage.error('启动发布失败')
  }
}

const promoteRelease = async (release: any) => {
  try {
    await request.post(`/api/v1/canary/releases/${release.id}/promote`)
    ElMessage.success('发布已推进')
    fetchReleases()
    if (currentRelease.value?.id === release.id) {
      fetchReleaseDetail(release.id)
    }
  } catch (error) {
    ElMessage.error('推进失败')
  }
}

const completeRelease = async (release: any) => {
  try {
    await ElMessageBox.confirm('确定要完成发布吗？这将把新版本设为正式版本。', '确认')
    await request.post(`/api/v1/canary/releases/${release.id}/complete`)
    ElMessage.success('发布已完成')
    fetchReleases()
  } catch (error) {
    // 用户取消
  }
}

const pauseRelease = async (release: any) => {
  try {
    await request.post(`/api/v1/canary/releases/${release.id}/pause`)
    ElMessage.success('发布已暂停')
    fetchReleases()
  } catch (error) {
    ElMessage.error('暂停失败')
  }
}

const rollbackRelease = async (release: any) => {
  try {
    await ElMessageBox.confirm('确定要回滚吗？', '确认', { type: 'warning' })
    await request.post(`/api/v1/canary/releases/${release.id}/rollback`, { reason: '用户手动回滚' })
    ElMessage.warning('发布已回滚')
    fetchReleases()
  } catch (error) {
    // 用户取消
  }
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    success: 'success',
    running: 'primary',
    paused: 'warning',
    failed: 'danger',
    rollback: 'info'
  }
  return types[status] || 'info'
}

const getProgressStatus = (status: string) => {
  if (status === 'success') return 'success'
  if (status === 'failed' || status === 'rollback') return 'exception'
  return null
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString()
}

onMounted(() => {
  fetchReleases()
})
</script>

<style scoped>
.canary-page {
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

.mt-4 {
  margin-top: 16px;
}
</style>
