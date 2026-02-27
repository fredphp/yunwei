<template>
  <div class="lb-page">
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>负载均衡器</span>
              <el-button type="primary" size="small" @click="showAddDialog = true">
                添加
              </el-button>
            </div>
          </template>
          
          <el-table :data="loadBalancers" v-loading="loading" @row-click="selectLB">
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'danger'">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="activeConns" label="连接数" width="80" />
            <el-table-column prop="autoOptimize" label="自动优化" width="100">
              <template #default="{ row }">
                <el-switch v-model="row.autoOptimize" size="small" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card class="mb-4" v-if="currentLB">
          <template #header>
            <div class="card-header">
              <span>后端服务器</span>
              <el-button type="primary" size="small" @click="showAddBackendDialog = true">
                添加后端
              </el-button>
            </div>
          </template>
          
          <el-table :data="backends" v-loading="backendLoading">
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="host" label="地址" />
            <el-table-column prop="port" label="端口" width="80" />
            <el-table-column prop="weight" label="权重" width="80" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'up' ? 'success' : 'danger'" size="small">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button size="small" @click="editBackend(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="removeBackend(row)">移除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 统计信息 -->
    <el-card class="mb-4" v-if="currentLB">
      <template #header>
        <div class="card-header">
          <span>实时统计</span>
          <el-button type="primary" size="small" @click="optimizeLB(currentLB)">
            AI 优化
          </el-button>
        </div>
      </template>
      
      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic title="总请求数" :value="currentLB.totalRequests" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="活跃连接" :value="currentLB.activeConns" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="平均延迟" :value="currentLB.avgLatency" suffix="ms" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="错误率" :value="currentLB.errorRate" suffix="%" />
        </el-col>
      </el-row>
    </el-card>

    <!-- 优化历史 -->
    <el-card class="mb-4">
      <template #header>
        <span>优化历史</span>
      </template>
      <el-table :data="optimizationHistory" v-loading="historyLoading">
        <el-table-column prop="type" label="优化类型" width="120" />
        <el-table-column prop="triggerReason" label="触发原因" show-overflow-tooltip />
        <el-table-column prop="aiDecision" label="AI 决策" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加 LB 对话框 -->
    <el-dialog v-model="showAddDialog" title="添加负载均衡器" width="500px">
      <el-form :model="lbForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="lbForm.name" placeholder="nginx-lb" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="lbForm.type">
            <el-option label="Nginx" value="nginx" />
            <el-option label="HAProxy" value="haproxy" />
            <el-option label="Traefik" value="traefik" />
          </el-select>
        </el-form-item>
        <el-form-item label="主机">
          <el-input v-model="lbForm.host" placeholder="192.168.1.100" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="lbForm.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="配置路径">
          <el-input v-model="lbForm.configPath" placeholder="/etc/nginx/nginx.conf" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="addLB">添加</el-button>
      </template>
    </el-dialog>

    <!-- 添加后端对话框 -->
    <el-dialog v-model="showAddBackendDialog" title="添加后端服务器" width="500px">
      <el-form :model="backendForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="backendForm.name" placeholder="backend-1" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="backendForm.host" placeholder="192.168.1.101" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="backendForm.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="权重">
          <el-input-number v-model="backendForm.weight" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="最大连接数">
          <el-input-number v-model="backendForm.maxConns" :min="1" :max="10000" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddBackendDialog = false">取消</el-button>
        <el-button type="primary" @click="addBackend">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const loading = ref(false)
const backendLoading = ref(false)
const historyLoading = ref(false)
const loadBalancers = ref([])
const backends = ref([])
const optimizationHistory = ref([])
const currentLB = ref<any>(null)
const showAddDialog = ref(false)
const showAddBackendDialog = ref(false)

const lbForm = ref({
  name: '',
  type: 'nginx',
  host: '',
  port: 80,
  configPath: '/etc/nginx/nginx.conf',
  autoOptimize: true
})

const backendForm = ref({
  lbId: 0,
  name: '',
  host: '',
  port: 8080,
  weight: 10,
  maxConns: 1000
})

const fetchLoadBalancers = async () => {
  loading.value = true
  try {
    const res = await request.get('/loadbalancer')
    loadBalancers.value = res.data || []
    if (loadBalancers.value.length > 0) {
      selectLB(loadBalancers.value[0])
    }
  } catch (error) {
    ElMessage.error('获取负载均衡器列表失败')
  } finally {
    loading.value = false
  }
}

const selectLB = async (lb: any) => {
  currentLB.value = lb
  backendForm.value.lbId = lb.id
  await fetchBackends(lb.id)
  await fetchOptimizationHistory(lb.id)
}

const fetchBackends = async (lbId: number) => {
  backendLoading.value = true
  try {
    const res = await request.get(`/loadbalancer/${lbId}/backends`)
    backends.value = res.data || []
  } catch (error) {
    console.error('获取后端服务器失败', error)
  } finally {
    backendLoading.value = false
  }
}

const fetchOptimizationHistory = async (lbId: number) => {
  historyLoading.value = true
  try {
    const res = await request.get('/loadbalancer/history', { params: { lbId } })
    optimizationHistory.value = res.data || []
  } catch (error) {
    console.error('获取优化历史失败', error)
  } finally {
    historyLoading.value = false
  }
}

const addLB = async () => {
  try {
    await request.post('/loadbalancer', lbForm.value)
    ElMessage.success('添加成功')
    showAddDialog.value = false
    fetchLoadBalancers()
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const addBackend = async () => {
  try {
    await request.post(`/loadbalancer/${currentLB.value.id}/backends`, backendForm.value)
    ElMessage.success('添加成功')
    showAddBackendDialog.value = false
    fetchBackends(currentLB.value.id)
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const editBackend = (backend: any) => {
  // 编辑后端
  ElMessage.info('编辑后端: ' + backend.name)
}

const removeBackend = async (backend: any) => {
  try {
    await request.delete(`/loadbalancer/backends/${backend.id}`)
    ElMessage.success('移除成功')
    fetchBackends(currentLB.value.id)
  } catch (error) {
    ElMessage.error('移除失败')
  }
}

const optimizeLB = async (lb: any) => {
  try {
    const res = await request.post(`/loadbalancer/${lb.id}/optimize`)
    ElMessage.success('AI 优化已执行')
    fetchOptimizationHistory(lb.id)
    if (res.data) {
      ElMessage.info('优化类型: ' + res.data.type)
    }
  } catch (error) {
    ElMessage.error('优化失败')
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString()
}

onMounted(() => {
  fetchLoadBalancers()
})
</script>

<style scoped>
.lb-page {
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
</style>
