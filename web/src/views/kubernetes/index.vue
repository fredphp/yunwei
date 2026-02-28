<template>
  <div class="kubernetes-page">
    <el-card class="mb-4">
      <template #header>
        <div class="card-header">
          <span>Kubernetes 集群管理</span>
          <el-button type="primary" @click="showAddDialog = true">
            <el-icon><Plus /></el-icon>
            添加集群
          </el-button>
        </div>
      </template>
      
      <el-table :data="clusters" v-loading="loading">
        <el-table-column prop="name" label="集群名称" />
        <el-table-column prop="apiEndpoint" label="API 地址" />
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="nodeCount" label="节点数" width="80" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'connected' ? 'success' : 'danger'">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="autoScaleEnabled" label="自动扩容" width="100">
          <template #default="{ row }">
            <el-switch v-model="row.autoScaleEnabled" @change="updateCluster(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" @click="viewDeployments(row)">查看</el-button>
            <el-button size="small" type="primary" @click="analyzeCluster(row)">AI分析</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 扩容历史 -->
    <el-card class="mb-4">
      <template #header>
        <span>扩容历史</span>
      </template>
      <el-table :data="scaleHistory" v-loading="historyLoading">
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="deployment" label="Deployment" />
        <el-table-column prop="scaleType" label="类型" width="100" />
        <el-table-column label="副本数变化" width="150">
          <template #default="{ row }">
            {{ row.replicasBefore }} → {{ row.replicasAfter || row.replicasTarget }}
          </template>
        </el-table-column>
        <el-table-column prop="triggerReason" label="触发原因" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getScaleStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 手动扩容对话框 -->
    <el-dialog v-model="showScaleDialog" title="手动扩容" width="500px">
      <el-form :model="scaleForm" label-width="100px">
        <el-form-item label="命名空间">
          <el-input v-model="scaleForm.namespace" placeholder="default" />
        </el-form-item>
        <el-form-item label="Deployment">
          <el-input v-model="scaleForm.deployment" placeholder="deployment名称" />
        </el-form-item>
        <el-form-item label="目标副本数">
          <el-input-number v-model="scaleForm.replicas" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="原因">
          <el-input v-model="scaleForm.reason" type="textarea" placeholder="扩容原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showScaleDialog = false">取消</el-button>
        <el-button type="primary" @click="executeScale">执行扩容</el-button>
      </template>
    </el-dialog>

    <!-- 添加集群对话框 -->
    <el-dialog v-model="showAddDialog" title="添加集群" width="600px">
      <el-form :model="clusterForm" label-width="120px">
        <el-form-item label="集群名称">
          <el-input v-model="clusterForm.name" placeholder="生产环境" />
        </el-form-item>
        <el-form-item label="API 地址">
          <el-input v-model="clusterForm.apiEndpoint" placeholder="https://k8s-api.example.com" />
        </el-form-item>
        <el-form-item label="Token">
          <el-input v-model="clusterForm.token" type="textarea" placeholder="ServiceAccount Token" />
        </el-form-item>
        <el-form-item label="最小副本数">
          <el-input-number v-model="clusterForm.minReplicas" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="最大副本数">
          <el-input-number v-model="clusterForm.maxReplicas" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="CPU 阈值(%)">
          <el-input-number v-model="clusterForm.cpuThreshold" :min="50" :max="100" />
        </el-form-item>
        <el-form-item label="内存阈值(%)">
          <el-input-number v-model="clusterForm.memThreshold" :min="50" :max="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="addCluster">添加</el-button>
      </template>
    </el-dialog>

    <!-- AI 分析结果 -->
    <el-dialog v-model="showAIDialog" title="AI 分析结果" width="700px">
      <div class="ai-result" v-if="aiResult">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="分析结果">{{ aiResult.summary }}</el-descriptions-item>
          <el-descriptions-item label="建议操作">{{ aiResult.suggestions }}</el-descriptions-item>
          <el-descriptions-item label="AI 置信度">{{ (aiResult.confidence * 100).toFixed(1) }}%</el-descriptions-item>
        </el-descriptions>
        <div class="mt-4" v-if="aiResult.commands">
          <h4>推荐命令:</h4>
          <el-input type="textarea" :rows="5" :model-value="aiResult.commands" readonly />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const historyLoading = ref(false)
const clusters = ref([])
const scaleHistory = ref([])
const showAddDialog = ref(false)
const showScaleDialog = ref(false)
const showAIDialog = ref(false)
const aiResult = ref<any>(null)

const clusterForm = ref({
  name: '',
  apiEndpoint: '',
  token: '',
  minReplicas: 2,
  maxReplicas: 10,
  cpuThreshold: 80,
  memThreshold: 80,
  autoScaleEnabled: true
})

const scaleForm = ref({
  clusterId: 0,
  namespace: 'default',
  deployment: '',
  replicas: 3,
  reason: ''
})

const fetchClusters = async () => {
  loading.value = true
  try {
    const res = await request.get('/kubernetes/clusters')
    clusters.value = res.data || []
  } catch (error) {
    ElMessage.error('获取集群列表失败')
  } finally {
    loading.value = false
  }
}

const fetchScaleHistory = async () => {
  historyLoading.value = true
  try {
    const res = await request.get('/kubernetes/scale/history')
    scaleHistory.value = res.data || []
  } catch (error) {
    console.error('获取扩容历史失败', error)
  } finally {
    historyLoading.value = false
  }
}

const addCluster = async () => {
  try {
    await request.post('/kubernetes/clusters', clusterForm.value)
    ElMessage.success('添加成功')
    showAddDialog.value = false
    fetchClusters()
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const updateCluster = async (cluster: any) => {
  try {
    await request.put(`/kubernetes/clusters/${cluster.id}`, cluster)
    ElMessage.success('更新成功')
  } catch (error) {
    ElMessage.error('更新失败')
  }
}

const analyzeCluster = async (cluster: any) => {
  try {
    const res = await request.post(`/kubernetes/clusters/${cluster.id}/analyze`, {
      namespace: 'default',
      deployment: ''
    })
    aiResult.value = res.data
    showAIDialog.value = true
  } catch (error) {
    ElMessage.error('AI分析失败')
  }
}

const executeScale = async () => {
  try {
    await request.post('/kubernetes/scale/manual', scaleForm.value)
    ElMessage.success('扩容请求已提交')
    showScaleDialog.value = false
    fetchScaleHistory()
  } catch (error) {
    ElMessage.error('扩容失败')
  }
}

const viewDeployments = (cluster: any) => {
  // 查看集群的 Deployments
  ElMessage.info(`查看集群 ${cluster.name} 的 Deployments`)
}

const getScaleStatusType = (status: string) => {
  const types: Record<string, string> = {
    success: 'success',
    running: 'warning',
    failed: 'danger',
    pending: 'info'
  }
  return types[status] || 'info'
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString()
}

onMounted(() => {
  fetchClusters()
  fetchScaleHistory()
})
</script>

<style scoped>
.kubernetes-page {
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
