<template>
  <div class="scheduler-page">
    <!-- 仪表盘统计 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="等待执行" :value="dashboard.taskStats?.pending || 0">
            <template #suffix>
              <el-icon class="text-warning"><Clock /></el-icon>
            </template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="执行中" :value="dashboard.taskStats?.running || 0">
            <template #suffix>
              <el-icon class="text-primary"><Loading /></el-icon>
            </template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="执行成功" :value="dashboard.taskStats?.success || 0">
            <template #suffix>
              <el-icon class="text-success"><CircleCheck /></el-icon>
            </template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="执行失败" :value="dashboard.taskStats?.failed || 0">
            <template #suffix>
              <el-icon class="text-danger"><CircleClose /></el-icon>
            </template>
          </el-statistic>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <!-- 左侧：任务列表 -->
      <el-col :span="16">
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>任务列表</span>
              <div>
                <el-button type="primary" @click="showSubmitDialog = true">
                  <el-icon><Plus /></el-icon>
                  提交任务
                </el-button>
                <el-button @click="showBatchDialog = true">
                  <el-icon><Files /></el-icon>
                  批量任务
                </el-button>
              </div>
            </div>
          </template>
          
          <el-tabs v-model="activeTab">
            <el-tab-pane label="任务列表" name="tasks">
              <el-table :data="tasks" v-loading="loading" @row-click="viewTask">
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="name" label="任务名称" />
                <el-table-column prop="type" label="类型" width="100">
                  <template #default="{ row }">
                    <el-tag size="small">{{ row.type }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="100">
                  <template #default="{ row }">
                    <el-tag :type="getStatusType(row.status)" size="small">
                      {{ row.status }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="priority" label="优先级" width="80">
                  <template #default="{ row }">
                    <el-rate v-model="row.priority" disabled :max="20" />
                  </template>
                </el-table-column>
                <el-table-column prop="queueName" label="队列" width="100" />
                <el-table-column prop="retryCount" label="重试" width="60">
                  <template #default="{ row }">
                    {{ row.retryCount }}/{{ row.maxRetry }}
                  </template>
                </el-table-column>
                <el-table-column prop="createdAt" label="创建时间" width="180">
                  <template #default="{ row }">
                    {{ formatDate(row.createdAt) }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="180">
                  <template #default="{ row }">
                    <el-button-group>
                      <el-button size="small" 
                        v-if="row.status === 'failed'"
                        type="warning"
                        @click.stop="retryTask(row.id)">
                        重试
                      </el-button>
                      <el-button size="small"
                        v-if="row.rollbackEnabled && row.status === 'success'"
                        type="danger"
                        @click.stop="rollbackTask(row.id)">
                        回滚
                      </el-button>
                      <el-button size="small"
                        v-if="row.status === 'pending'"
                        type="info"
                        @click.stop="cancelTask(row.id)">
                        取消
                      </el-button>
                    </el-button-group>
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
            
            <el-tab-pane label="批量任务" name="batches">
              <el-table :data="batches" v-loading="batchesLoading">
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="name" label="批次名称" />
                <el-table-column prop="status" label="状态" width="100">
                  <template #default="{ row }">
                    <el-tag :type="getStatusType(row.status)" size="small">
                      {{ row.status }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="进度" width="200">
                  <template #default="{ row }">
                    <el-progress 
                      :percentage="Math.round((row.successTasks + row.failedTasks) / row.totalTasks * 100)"
                      :status="row.status === 'success' ? 'success' : row.status === 'failed' ? 'exception' : ''" />
                    <small>{{ row.successTasks }}/{{ row.totalTasks }} 完成</small>
                  </template>
                </el-table-column>
                <el-table-column prop="createdAt" label="创建时间" width="180">
                  <template #default="{ row }">
                    {{ formatDate(row.createdAt) }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="100">
                  <template #default="{ row }">
                    <el-button size="small" @click="viewBatch(row.id)">详情</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
            
            <el-tab-pane label="定时任务" name="cron">
              <el-table :data="cronJobs" v-loading="cronLoading">
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="name" label="任务名称" />
                <el-table-column prop="cronExpr" label="Cron 表达式" width="120" />
                <el-table-column prop="enabled" label="启用" width="80">
                  <template #default="{ row }">
                    <el-switch v-model="row.enabled" @change="toggleCron(row)" />
                  </template>
                </el-table-column>
                <el-table-column prop="lastStatus" label="上次状态" width="100">
                  <template #default="{ row }">
                    <el-tag v-if="row.lastStatus" :type="getStatusType(row.lastStatus)" size="small">
                      {{ row.lastStatus }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="nextRunAt" label="下次执行" width="180">
                  <template #default="{ row }">
                    {{ row.nextRunAt ? formatDate(row.nextRunAt) : '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="150">
                  <template #default="{ row }">
                    <el-button size="small" @click="triggerCron(row.id)">立即执行</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>

      <!-- 右侧：队列和 Worker 状态 -->
      <el-col :span="8">
        <el-card class="mb-4">
          <template #header>
            <span>队列状态</span>
          </template>
          
          <el-table :data="queueStats" size="small">
            <el-table-column prop="name" label="队列" />
            <el-table-column prop="pending" label="等待" width="80" />
            <el-table-column prop="running" label="运行" width="80" />
          </el-table>
        </el-card>

        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>Worker 状态</span>
              <el-button size="small" @click="showScaleDialog = true">扩缩容</el-button>
            </div>
          </template>
          
          <el-table :data="workerStats" size="small">
            <el-table-column prop="queueName" label="队列" />
            <el-table-column label="Worker" width="100">
              <template #default="{ row }">
                {{ row.idleWorkers }}/{{ row.totalWorkers }}
              </template>
            </el-table-column>
            <el-table-column prop="totalTasksHandled" label="已处理" width="80" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 提交任务对话框 -->
    <el-dialog v-model="showSubmitDialog" title="提交任务" width="600px">
      <el-form :model="taskForm" label-width="100px">
        <el-form-item label="任务名称">
          <el-input v-model="taskForm.name" placeholder="任务名称" />
        </el-form-item>
        <el-form-item label="任务类型">
          <el-select v-model="taskForm.type">
            <el-option label="命令执行" value="command" />
            <el-option label="脚本执行" value="script" />
            <el-option label="部署任务" value="deploy" />
            <el-option label="备份任务" value="backup" />
            <el-option label="清理任务" value="cleanup" />
            <el-option label="批量任务" value="batch" />
          </el-select>
        </el-form-item>
        <el-form-item label="执行命令">
          <el-input v-model="taskForm.command" type="textarea" :rows="3" placeholder="命令或脚本" />
        </el-form-item>
        <el-form-item label="目标服务器">
          <el-select v-model="taskForm.targetIds" multiple placeholder="选择服务器">
            <el-option v-for="s in servers" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="队列">
          <el-select v-model="taskForm.queueName">
            <el-option label="默认队列" value="default" />
            <el-option label="关键任务" value="critical" />
            <el-option label="后台任务" value="background" />
            <el-option label="部署任务" value="deploy" />
            <el-option label="批量任务" value="batch" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-rate v-model="taskForm.priority" :max="20" />
        </el-form-item>
        <el-form-item label="超时(秒)">
          <el-input-number v-model="taskForm.timeout" :min="10" :max="7200" />
        </el-form-item>
        <el-form-item label="重试次数">
          <el-input-number v-model="taskForm.maxRetry" :min="0" :max="10" />
        </el-form-item>
        <el-form-item label="回滚命令">
          <el-input v-model="taskForm.rollbackCommand" placeholder="可选：回滚时执行的命令" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSubmitDialog = false">取消</el-button>
        <el-button type="primary" @click="submitTask" :loading="submitting">提交</el-button>
      </template>
    </el-dialog>

    <!-- 批量任务对话框 -->
    <el-dialog v-model="showBatchDialog" title="批量任务" width="700px">
      <el-form :model="batchForm" label-width="100px">
        <el-form-item label="批次名称">
          <el-input v-model="batchForm.name" placeholder="批次名称" />
        </el-form-item>
        <el-form-item label="并行数">
          <el-input-number v-model="batchForm.parallelism" :min="1" :max="50" />
        </el-form-item>
        <el-form-item label="失败时停止">
          <el-switch v-model="batchForm.stopOnFail" />
        </el-form-item>
        <el-form-item label="任务列表">
          <div v-for="(task, index) in batchForm.tasks" :key="index" class="batch-task-item">
            <el-input v-model="task.name" placeholder="任务名称" style="width: 150px; margin-right: 8px" />
            <el-input v-model="task.command" placeholder="命令" style="flex: 1; margin-right: 8px" />
            <el-button type="danger" @click="batchForm.tasks.splice(index, 1)">删除</el-button>
          </div>
          <el-button type="primary" plain @click="batchForm.tasks.push({ name: '', command: '', type: 'command' })">
            添加任务
          </el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatchDialog = false">取消</el-button>
        <el-button type="primary" @click="submitBatch" :loading="submitting">提交</el-button>
      </template>
    </el-dialog>

    <!-- 任务详情对话框 -->
    <el-dialog v-model="showTaskDetail" title="任务详情" width="700px">
      <el-descriptions :column="2" border v-if="currentTask">
        <el-descriptions-item label="任务ID">{{ currentTask.id }}</el-descriptions-item>
        <el-descriptions-item label="任务名称">{{ currentTask.name }}</el-descriptions-item>
        <el-descriptions-item label="任务类型">{{ currentTask.type }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentTask.status)">{{ currentTask.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="队列">{{ currentTask.queueName }}</el-descriptions-item>
        <el-descriptions-item label="Worker">{{ currentTask.workerId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="重试">{{ currentTask.retryCount }}/{{ currentTask.maxRetry }}</el-descriptions-item>
        <el-descriptions-item label="超时">{{ currentTask.timeout }}秒</el-descriptions-item>
        <el-descriptions-item label="命令" :span="2">
          <pre>{{ currentTask.command }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="输出" :span="2">
          <pre class="task-output">{{ currentTask.stdout || currentTask.result }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="错误" :span="2" v-if="currentTask.error">
          <pre class="task-error">{{ currentTask.error }}</pre>
        </el-descriptions-item>
      </el-descriptions>

      <div class="mt-4" v-if="taskExecutions.length > 0">
        <h4>执行历史</h4>
        <el-timeline>
          <el-timeline-item
            v-for="exec in taskExecutions"
            :key="exec.id"
            :type="getStatusType(exec.status)"
            :timestamp="formatDate(exec.createdAt)">
            <p>尝试 #{{ exec.attempt }} - {{ exec.status }}</p>
            <p v-if="exec.duration">耗时: {{ exec.duration }}ms</p>
          </el-timeline-item>
        </el-timeline>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Files, Clock, Loading, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const batchesLoading = ref(false)
const cronLoading = ref(false)
const submitting = ref(false)
const activeTab = ref('tasks')
const showSubmitDialog = ref(false)
const showBatchDialog = ref(false)
const showScaleDialog = ref(false)
const showTaskDetail = ref(false)

const dashboard = ref<any>({})
const tasks = ref<any[]>([])
const batches = ref<any[]>([])
const cronJobs = ref<any[]>([])
const servers = ref<any[]>([])
const currentTask = ref<any>(null)
const taskExecutions = ref<any[]>([])

const taskForm = ref({
  name: '',
  type: 'command',
  command: '',
  queueName: 'default',
  priority: 5,
  timeout: 300,
  maxRetry: 3,
  retryDelay: 10,
  targetIds: [] as number[],
  rollbackCommand: ''
})

const batchForm = ref({
  name: '',
  parallelism: 5,
  stopOnFail: false,
  tasks: [] as any[]
})

const queueStats = computed(() => {
  const stats = dashboard.value.queueStats || {}
  return Object.entries(stats).map(([name, data]: [string, any]) => ({
    name,
    pending: data.pending || 0,
    running: data.running || 0
  }))
})

const workerStats = computed(() => {
  const stats = dashboard.value.workerStats || {}
  return Object.entries(stats).map(([name, data]: [string, any]) => ({
    queueName: name,
    ...data
  }))
})

const fetchDashboard = async () => {
  try {
    const res = await request.get('/api/v1/scheduler/dashboard')
    dashboard.value = res.data || {}
  } catch (error) {
    console.error('获取仪表盘失败', error)
  }
}

const fetchTasks = async () => {
  loading.value = true
  try {
    const res = await request.get('/api/v1/scheduler/tasks')
    tasks.value = res.data?.list || []
  } catch (error) {
    console.error('获取任务列表失败', error)
  } finally {
    loading.value = false
  }
}

const fetchBatches = async () => {
  batchesLoading.value = true
  try {
    const res = await request.get('/api/v1/scheduler/batches')
    batches.value = res.data || []
  } catch (error) {
    console.error('获取批次列表失败', error)
  } finally {
    batchesLoading.value = false
  }
}

const fetchCronJobs = async () => {
  cronLoading.value = true
  try {
    const res = await request.get('/api/v1/scheduler/cron')
    cronJobs.value = res.data || []
  } catch (error) {
    console.error('获取定时任务失败', error)
  } finally {
    cronLoading.value = false
  }
}

const submitTask = async () => {
  submitting.value = true
  try {
    await request.post('/api/v1/scheduler/tasks/options', taskForm.value)
    ElMessage.success('任务已提交')
    showSubmitDialog.value = false
    fetchTasks()
    fetchDashboard()
  } catch (error) {
    ElMessage.error('提交失败')
  } finally {
    submitting.value = false
  }
}

const submitBatch = async () => {
  submitting.value = true
  try {
    await request.post('/api/v1/scheduler/batches', batchForm.value)
    ElMessage.success('批量任务已提交')
    showBatchDialog.value = false
    fetchBatches()
    fetchDashboard()
  } catch (error) {
    ElMessage.error('提交失败')
  } finally {
    submitting.value = false
  }
}

const viewTask = async (row: any) => {
  currentTask.value = row
  showTaskDetail.value = true
  
  // 获取执行历史
  try {
    const res = await request.get(`/api/v1/scheduler/tasks/${row.id}/executions`)
    taskExecutions.value = res.data || []
  } catch (error) {
    console.error('获取执行历史失败', error)
  }
}

const viewBatch = async (id: number) => {
  // 查看批次详情
  ElMessage.info(`查看批次 ${id}`)
}

const cancelTask = async (id: number) => {
  try {
    await request.post(`/api/v1/scheduler/tasks/${id}/cancel`)
    ElMessage.success('任务已取消')
    fetchTasks()
  } catch (error) {
    ElMessage.error('取消失败')
  }
}

const retryTask = async (id: number) => {
  try {
    await request.post(`/api/v1/scheduler/tasks/${id}/retry`)
    ElMessage.success('任务已重试')
    fetchTasks()
  } catch (error) {
    ElMessage.error('重试失败')
  }
}

const rollbackTask = async (id: number) => {
  try {
    await request.post(`/api/v1/scheduler/tasks/${id}/rollback`)
    ElMessage.warning('任务已回滚')
    fetchTasks()
  } catch (error) {
    ElMessage.error('回滚失败')
  }
}

const triggerCron = async (id: number) => {
  try {
    await request.post(`/api/v1/scheduler/cron/${id}/trigger`)
    ElMessage.success('已触发执行')
  } catch (error) {
    ElMessage.error('触发失败')
  }
}

const toggleCron = async (job: any) => {
  try {
    await request.put(`/api/v1/scheduler/cron/${job.id}`, job)
    ElMessage.success(job.enabled ? '已启用' : '已禁用')
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    pending: 'info',
    queued: 'info',
    running: 'primary',
    success: 'success',
    failed: 'danger',
    retrying: 'warning',
    canceled: 'info',
    timeout: 'danger',
    rolledback: 'warning'
  }
  return types[status] || 'info'
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString()
}

onMounted(() => {
  fetchDashboard()
  fetchTasks()
  fetchBatches()
  fetchCronJobs()
  
  // 定时刷新
  setInterval(() => {
    fetchDashboard()
    if (activeTab.value === 'tasks') fetchTasks()
    if (activeTab.value === 'batches') fetchBatches()
    if (activeTab.value === 'cron') fetchCronJobs()
  }, 5000)
})
</script>

<style scoped>
.scheduler-page {
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

.batch-task-item {
  display: flex;
  margin-bottom: 8px;
  align-items: center;
}

.task-output {
  background: #f5f7fa;
  padding: 8px;
  border-radius: 4px;
  max-height: 200px;
  overflow: auto;
}

.task-error {
  background: #fef0f0;
  padding: 8px;
  border-radius: 4px;
  color: #f56c6c;
}
</style>
