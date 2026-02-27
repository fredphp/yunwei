<template>
  <div class="alerts-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #f56c6c;">
              <el-icon size="28"><Bell /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total }}</div>
              <div class="stat-label">告警总数</div>
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
              <div class="stat-value">{{ stats.pending }}</div>
              <div class="stat-label">待处理</div>
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
              <div class="stat-value">{{ stats.resolved }}</div>
              <div class="stat-label">已处理</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #909399;">
              <el-icon size="28"><Timer /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.avgTime }}m</div>
              <div class="stat-label">平均处理时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <!-- 告警列表 -->
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>告警列表</span>
              <div class="header-actions">
                <el-select v-model="filterLevel" placeholder="级别" style="width: 100px; margin-right: 10px;" clearable>
                  <el-option label="紧急" value="critical" />
                  <el-option label="警告" value="warning" />
                  <el-option label="信息" value="info" />
                </el-select>
                <el-select v-model="filterStatus" placeholder="状态" style="width: 100px;" clearable>
                  <el-option label="待处理" value="pending" />
                  <el-option label="处理中" value="processing" />
                  <el-option label="已解决" value="resolved" />
                </el-select>
              </div>
            </div>
          </template>

          <el-table :data="filteredAlerts" v-loading="loading" max-height="500">
            <el-table-column label="级别" width="80">
              <template #default="{ row }">
                <el-tag :type="getLevelType(row.level)" effect="dark" size="small">
                  {{ row.level }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="title" label="告警标题" min-width="180" show-overflow-tooltip />
            <el-table-column prop="source" label="来源" width="120" />
            <el-table-column prop="serverName" label="服务器" width="120" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="时间" width="160">
              <template #default="{ row }">
                {{ formatDate(row.createdAt) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="viewAlertDetail(row)">详情</el-button>
                <el-button size="small" type="primary" @click="acknowledgeAlert(row)" v-if="row.status === 'pending'">
                  处理
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- 右侧面板 -->
      <el-col :span="8">
        <!-- 检测规则 -->
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>检测规则</span>
              <el-button size="small" type="primary" @click="showRuleDialog = true">
                <el-icon><Plus /></el-icon>
              </el-button>
            </div>
          </template>
          <el-table :data="rules" size="small">
            <el-table-column prop="name" label="规则名称" />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" @change="updateRule(row)" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- AI决策 -->
        <el-card>
          <template #header>
            <span>AI 决策建议</span>
          </template>
          <div v-if="decisions.length">
            <div v-for="decision in decisions" :key="decision.id" class="decision-item">
              <div class="decision-header">
                <span class="decision-title">{{ decision.title }}</span>
                <el-tag :type="getDecisionType(decision.confidence)" size="small">
                  {{ (decision.confidence * 100).toFixed(0) }}% 置信度
                </el-tag>
              </div>
              <p class="decision-desc">{{ decision.description }}</p>
              <div class="decision-actions">
                <el-button size="small" type="primary" @click="approveDecision(decision)">采纳</el-button>
                <el-button size="small" @click="rejectDecision(decision)">忽略</el-button>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无决策建议" :image-size="80" />
        </el-card>

        <!-- 自动操作 -->
        <el-card class="mt-4">
          <template #header>
            <span>自动操作</span>
          </template>
          <div v-if="actions.length">
            <div v-for="action in actions" :key="action.id" class="action-item">
              <div class="action-info">
                <span>{{ action.name }}</span>
                <el-tag size="small">{{ action.triggerCount }}次触发</el-tag>
              </div>
              <el-button size="small" type="primary" @click="executeAction(action)">执行</el-button>
            </div>
          </div>
          <el-empty v-else description="暂无自动操作" :image-size="80" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 告警详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="告警详情" width="700px">
      <el-descriptions :column="2" border v-if="currentAlert">
        <el-descriptions-item label="告警标题" :span="2">{{ currentAlert.title }}</el-descriptions-item>
        <el-descriptions-item label="告警级别">
          <el-tag :type="getLevelType(currentAlert.level)">{{ currentAlert.level }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentAlert.status)">{{ currentAlert.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="来源">{{ currentAlert.source }}</el-descriptions-item>
        <el-descriptions-item label="服务器">{{ currentAlert.serverName }}</el-descriptions-item>
        <el-descriptions-item label="触发时间">{{ formatDate(currentAlert.createdAt) }}</el-descriptions-item>
        <el-descriptions-item label="处理时间">{{ currentAlert.resolvedAt ? formatDate(currentAlert.resolvedAt) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="告警详情" :span="2">{{ currentAlert.description }}</el-descriptions-item>
        <el-descriptions-item label="处理备注" :span="2">
          <el-input v-model="ackNote" type="textarea" :rows="3" placeholder="请输入处理备注" />
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
        <el-button type="primary" @click="resolveAlert" v-if="currentAlert?.status !== 'resolved'">标记已解决</el-button>
      </template>
    </el-dialog>

    <!-- 规则编辑对话框 -->
    <el-dialog v-model="showRuleDialog" title="添加检测规则" width="500px">
      <el-form :model="ruleForm" label-width="100px">
        <el-form-item label="规则名称">
          <el-input v-model="ruleForm.name" />
        </el-form-item>
        <el-form-item label="监控指标">
          <el-select v-model="ruleForm.metric" style="width: 100%;">
            <el-option label="CPU 使用率" value="cpu" />
            <el-option label="内存使用率" value="memory" />
            <el-option label="磁盘使用率" value="disk" />
            <el-option label="网络流量" value="network" />
          </el-select>
        </el-form-item>
        <el-form-item label="条件">
          <el-select v-model="ruleForm.operator" style="width: 100px;">
            <el-option label=">" value="gt" />
            <el-option label="<" value="lt" />
            <el-option label="=" value="eq" />
          </el-select>
          <el-input-number v-model="ruleForm.threshold" :min="0" :max="100" style="width: 150px; margin-left: 10px;" />
          <span style="margin-left: 10px;">%</span>
        </el-form-item>
        <el-form-item label="告警级别">
          <el-select v-model="ruleForm.level" style="width: 100%;">
            <el-option label="紧急" value="critical" />
            <el-option label="警告" value="warning" />
            <el-option label="信息" value="info" />
          </el-select>
        </el-form-item>
        <el-form-item label="持续时间">
          <el-input-number v-model="ruleForm.duration" :min="1" :max="60" />
          <span style="margin-left: 10px;">分钟</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRuleDialog = false">取消</el-button>
        <el-button type="primary" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Bell, Warning, CircleCheck, Timer, Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const alerts = ref([])
const rules = ref([])
const decisions = ref([])
const actions = ref([])
const filterLevel = ref('')
const filterStatus = ref('')
const showDetailDialog = ref(false)
const showRuleDialog = ref(false)
const currentAlert = ref<any>(null)
const ackNote = ref('')

const stats = ref({
  total: 0,
  pending: 0,
  resolved: 0,
  avgTime: 0
})

const ruleForm = ref({
  name: '',
  metric: 'cpu',
  operator: 'gt',
  threshold: 80,
  level: 'warning',
  duration: 5,
  enabled: true
})

const filteredAlerts = computed(() => {
  let result = alerts.value
  if (filterLevel.value) {
    result = result.filter((a: any) => a.level === filterLevel.value)
  }
  if (filterStatus.value) {
    result = result.filter((a: any) => a.status === filterStatus.value)
  }
  return result
})

const fetchAlerts = async () => {
  loading.value = true
  try {
    const res = await request.get('/api/v1/alerts')
    alerts.value = res.data || []
    
    stats.value.total = alerts.value.length
    stats.value.pending = alerts.value.filter((a: any) => a.status === 'pending').length
    stats.value.resolved = alerts.value.filter((a: any) => a.status === 'resolved').length
    stats.value.avgTime = 15 // 模拟平均处理时间
  } catch (error) {
    ElMessage.error('获取告警列表失败')
  } finally {
    loading.value = false
  }
}

const fetchRules = async () => {
  try {
    const res = await request.get('/api/v1/rules')
    rules.value = res.data || []
  } catch (error) {
    console.error('获取规则失败', error)
  }
}

const fetchDecisions = async () => {
  try {
    const res = await request.get('/api/v1/decisions')
    decisions.value = res.data || []
  } catch (error) {
    console.error('获取决策失败', error)
  }
}

const fetchActions = async () => {
  try {
    const res = await request.get('/api/v1/actions')
    actions.value = res.data || []
  } catch (error) {
    console.error('获取自动操作失败', error)
  }
}

const viewAlertDetail = (alert: any) => {
  currentAlert.value = alert
  showDetailDialog.value = true
}

const acknowledgeAlert = async (alert: any) => {
  try {
    await request.post(`/api/v1/alerts/${alert.id}/acknowledge`)
    ElMessage.success('告警已确认')
    fetchAlerts()
  } catch (error) {
    ElMessage.error('处理失败')
  }
}

const resolveAlert = async () => {
  try {
    await request.post(`/api/v1/alerts/${currentAlert.value.id}/acknowledge`, { note: ackNote.value })
    ElMessage.success('已标记为已解决')
    showDetailDialog.value = false
    fetchAlerts()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const updateRule = async (rule: any) => {
  try {
    await request.put(`/api/v1/rules/${rule.id}`, rule)
    ElMessage.success('规则已更新')
  } catch (error) {
    ElMessage.error('更新失败')
  }
}

const saveRule = async () => {
  try {
    await request.post('/api/v1/rules', ruleForm.value)
    ElMessage.success('规则已添加')
    showRuleDialog.value = false
    fetchRules()
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const approveDecision = async (decision: any) => {
  try {
    await request.post(`/api/v1/decisions/${decision.id}/approve`)
    ElMessage.success('决策已采纳')
    fetchDecisions()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const rejectDecision = async (decision: any) => {
  try {
    await request.post(`/api/v1/decisions/${decision.id}/reject`)
    ElMessage.success('决策已忽略')
    fetchDecisions()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const executeAction = async (action: any) => {
  try {
    await request.post(`/api/v1/actions/${action.id}/execute`)
    ElMessage.success('操作已执行')
  } catch (error) {
    ElMessage.error('执行失败')
  }
}

const getLevelType = (level: string) => {
  const types: Record<string, string> = {
    critical: 'danger',
    warning: 'warning',
    info: 'info'
  }
  return types[level] || 'info'
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    pending: 'warning',
    processing: 'primary',
    resolved: 'success'
  }
  return types[status] || 'info'
}

const getDecisionType = (confidence: number) => {
  if (confidence >= 0.8) return 'success'
  if (confidence >= 0.6) return 'warning'
  return 'info'
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString()
}

onMounted(() => {
  fetchAlerts()
  fetchRules()
  fetchDecisions()
  fetchActions()
})
</script>

<style scoped>
.alerts-page {
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

.decision-item, .action-item {
  padding: 12px;
  border-bottom: 1px solid #ebeef5;
}

.decision-item:last-child, .action-item:last-child {
  border-bottom: none;
}

.decision-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.decision-title {
  font-weight: 500;
}

.decision-desc {
  font-size: 12px;
  color: #909399;
  margin-bottom: 10px;
}

.decision-actions {
  display: flex;
  gap: 8px;
}

.action-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.action-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
</style>
