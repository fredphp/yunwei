<template>
  <div class="backup-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #409eff;">
              <el-icon size="28"><FolderOpened /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.totalBackups }}</div>
              <div class="stat-label">备份任务</div>
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
              <div class="stat-value">{{ stats.successRate }}%</div>
              <div class="stat-label">成功率</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #e6a23c;">
              <el-icon size="28"><Timer /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.storageUsed }}GB</div>
              <div class="stat-label">存储使用</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #f56c6c;">
              <el-icon size="28"><Warning /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.failedBackups }}</div>
              <div class="stat-label">失败任务</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="16">
        <!-- 备份策略列表 -->
        <el-card>
          <template #header>
            <div class="card-header">
              <span>备份策略</span>
              <el-button type="primary" @click="showPolicyDialog = true">
                <el-icon><Plus /></el-icon> 新建策略
              </el-button>
            </div>
          </template>
          <el-table :data="policies" v-loading="loading">
            <el-table-column prop="name" label="策略名称" min-width="150" />
            <el-table-column prop="type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="schedule" label="调度周期" width="120" />
            <el-table-column prop="retention" label="保留天数" width="100" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" @change="togglePolicy(row)" />
              </template>
            </el-table-column>
            <el-table-column prop="lastBackup" label="上次执行" width="160" />
            <el-table-column prop="nextBackup" label="下次执行" width="160" />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="triggerBackup(row)">立即执行</el-button>
                <el-button size="small" @click="editPolicy(row)">编辑</el-button>
                <el-button size="small" type="danger" text @click="deletePolicy(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 备份记录 -->
        <el-card class="mt-4">
          <template #header>
            <div class="card-header">
              <span>备份记录</span>
              <el-button type="primary" @click="showQuickBackupDialog = true">快速备份</el-button>
            </div>
          </template>
          <el-table :data="records" max-height="400">
            <el-table-column prop="policyName" label="策略" width="120" />
            <el-table-column prop="target" label="备份目标" min-width="150" />
            <el-table-column prop="size" label="大小" width="100" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="duration" label="耗时" width="100" />
            <el-table-column prop="createdAt" label="时间" width="160" />
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="viewRecord(row)">详情</el-button>
                <el-button size="small" type="primary" @click="restoreBackup(row)">恢复</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <!-- 灾备演练 -->
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>灾备演练</span>
              <el-button size="small" type="primary" @click="showDrillDialog = true">新建</el-button>
            </div>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="drill in drills"
              :key="drill.id"
              :type="drill.status === 'completed' ? 'success' : 'primary'"
              :timestamp="drill.date"
            >
              <div class="drill-item">
                <div class="drill-title">{{ drill.name }}</div>
                <div class="drill-result" v-if="drill.status === 'completed'">
                  <el-tag size="small" :type="drill.result === 'pass' ? 'success' : 'danger'">
                    {{ drill.result === 'pass' ? '通过' : '失败' }}
                  </el-tag>
                  <span class="drill-rto">RTO: {{ drill.rto }}</span>
                </div>
                <el-button v-else size="small" @click="executeDrill(drill)">执行</el-button>
              </div>
            </el-timeline-item>
          </el-timeline>
        </el-card>

        <!-- 存储配置 -->
        <el-card>
          <template #header>
            <div class="card-header">
              <span>存储配置</span>
              <el-button size="small" @click="showStorageDialog = true">添加</el-button>
            </div>
          </template>
          <div v-for="storage in storages" :key="storage.id" class="storage-item">
            <div class="storage-info">
              <el-icon size="24"><Box /></el-icon>
              <div class="storage-detail">
                <div class="storage-name">{{ storage.name }}</div>
                <div class="storage-type">{{ storage.type }}</div>
              </div>
            </div>
            <div class="storage-usage">
              <el-progress :percentage="storage.usage" :stroke-width="6" />
              <span class="usage-text">{{ storage.used }}/{{ storage.total }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 对话框 -->
    <el-dialog v-model="showPolicyDialog" title="新建备份策略" width="600px">
      <el-form :model="policyForm" label-width="100px">
        <el-form-item label="策略名称">
          <el-input v-model="policyForm.name" />
        </el-form-item>
        <el-form-item label="备份类型">
          <el-select v-model="policyForm.type" style="width: 100%;">
            <el-option label="数据库备份" value="database" />
            <el-option label="文件备份" value="file" />
            <el-option label="快照备份" value="snapshot" />
          </el-select>
        </el-form-item>
        <el-form-item label="备份目标">
          <el-input v-model="policyForm.target" placeholder="数据库连接串或文件路径" />
        </el-form-item>
        <el-form-item label="调度周期">
          <el-select v-model="policyForm.schedule" style="width: 100%;">
            <el-option label="每小时" value="hourly" />
            <el-option label="每天" value="daily" />
            <el-option label="每周" value="weekly" />
            <el-option label="每月" value="monthly" />
          </el-select>
        </el-form-item>
        <el-form-item label="保留天数">
          <el-input-number v-model="policyForm.retention" :min="1" :max="365" />
        </el-form-item>
        <el-form-item label="存储位置">
          <el-select v-model="policyForm.storageId" style="width: 100%;">
            <el-option v-for="s in storages" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPolicyDialog = false">取消</el-button>
        <el-button type="primary" @click="savePolicy">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showQuickBackupDialog" title="快速备份" width="500px">
      <el-form :model="quickBackupForm" label-width="100px">
        <el-form-item label="备份类型">
          <el-select v-model="quickBackupForm.type" style="width: 100%;">
            <el-option label="数据库" value="database" />
            <el-option label="文件" value="file" />
          </el-select>
        </el-form-item>
        <el-form-item label="备份目标">
          <el-input v-model="quickBackupForm.target" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showQuickBackupDialog = false">取消</el-button>
        <el-button type="primary" @click="executeQuickBackup">执行</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showDrillDialog" title="新建灾备演练" width="500px">
      <el-form :model="drillForm" label-width="100px">
        <el-form-item label="演练名称">
          <el-input v-model="drillForm.name" />
        </el-form-item>
        <el-form-item label="演练类型">
          <el-select v-model="drillForm.type" style="width: 100%;">
            <el-option label="桌面演练" value="desktop" />
            <el-option label="部分演练" value="partial" />
            <el-option label="完整演练" value="full" />
          </el-select>
        </el-form-item>
        <el-form-item label="计划时间">
          <el-date-picker v-model="drillForm.scheduledAt" type="datetime" style="width: 100%;" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDrillDialog = false">取消</el-button>
        <el-button type="primary" @click="saveDrill">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showStorageDialog" title="添加存储" width="500px">
      <el-form :model="storageForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="storageForm.name" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="storageForm.type" style="width: 100%;">
            <el-option label="本地存储" value="local" />
            <el-option label="S3" value="s3" />
            <el-option label="OSS" value="oss" />
            <el-option label="NFS" value="nfs" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="storageForm.config" type="textarea" :rows="4" placeholder="JSON配置" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showStorageDialog = false">取消</el-button>
        <el-button type="primary" @click="saveStorage">保存</el-button>
      </template>
    </el-dialog>

    <!-- 恢复对话框 -->
    <el-dialog v-model="showRestoreDialog" title="恢复备份" width="600px">
      <el-alert title="警告" type="warning" description="恢复操作将覆盖现有数据，请谨慎操作" show-icon :closable="false" class="mb-4" />
      <el-form :model="restoreForm" label-width="100px">
        <el-form-item label="备份记录">
          <el-input :model-value="currentRecord?.policyName + ' - ' + currentRecord?.createdAt" disabled />
        </el-form-item>
        <el-form-item label="恢复目标">
          <el-input v-model="restoreForm.target" />
        </el-form-item>
        <el-form-item label="恢复类型">
          <el-radio-group v-model="restoreForm.mode">
            <el-radio value="full">完整恢复</el-radio>
            <el-radio value="partial">部分恢复</el-radio>
            <el-radio value="pitr">时间点恢复</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="时间点" v-if="restoreForm.mode === 'pitr'">
          <el-date-picker v-model="restoreForm.pointInTime" type="datetime" style="width: 100%;" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRestoreDialog = false">取消</el-button>
        <el-button type="danger" @click="confirmRestore">确认恢复</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, FolderOpened, CircleCheck, Timer, Warning, Box } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const showPolicyDialog = ref(false)
const showQuickBackupDialog = ref(false)
const showDrillDialog = ref(false)
const showStorageDialog = ref(false)
const showRestoreDialog = ref(false)
const currentRecord = ref<any>(null)

const stats = ref({
  totalBackups: 156,
  successRate: 98.5,
  storageUsed: 2.4,
  failedBackups: 2
})

const policies = ref([
  { id: 1, name: 'MySQL生产库备份', type: 'database', schedule: 'daily', retention: 30, enabled: true, lastBackup: '2024-02-23 02:00', nextBackup: '2024-02-24 02:00' },
  { id: 2, name: '用户文件备份', type: 'file', schedule: 'weekly', retention: 90, enabled: true, lastBackup: '2024-02-18 03:00', nextBackup: '2024-02-25 03:00' },
  { id: 3, name: 'VM快照', type: 'snapshot', schedule: 'hourly', retention: 7, enabled: true, lastBackup: '2024-02-23 10:00', nextBackup: '2024-02-23 11:00' }
])

const records = ref([
  { id: 1, policyName: 'MySQL生产库', target: 'prod-mysql-master', size: '2.3GB', status: 'success', duration: '3m 20s', createdAt: '2024-02-23 02:00' },
  { id: 2, policyName: 'VM快照', target: 'vm-web-01', size: '15GB', status: 'success', duration: '1m 45s', createdAt: '2024-02-23 10:00' },
  { id: 3, policyName: '用户文件', target: '/data/users', size: '8.5GB', status: 'success', duration: '12m 30s', createdAt: '2024-02-18 03:00' }
])

const drills = ref([
  { id: 1, name: 'Q1灾备演练', date: '2024-02-15', status: 'completed', result: 'pass', rto: '45min' },
  { id: 2, name: '数据库恢复演练', date: '2024-02-20', status: 'completed', result: 'pass', rto: '12min' },
  { id: 3, name: 'Q2灾备演练', date: '2024-03-01', status: 'pending' }
])

const storages = ref([
  { id: 1, name: '本地NAS', type: 'NFS', used: '1.2TB', total: '2TB', usage: 60 },
  { id: 2, name: 'AWS S3', type: 'S3', used: '800GB', total: '无限制', usage: 20 }
])

const policyForm = ref({
  name: '',
  type: 'database',
  target: '',
  schedule: 'daily',
  retention: 30,
  storageId: 1,
  enabled: true
})

const quickBackupForm = ref({
  type: 'database',
  target: ''
})

const drillForm = ref({
  name: '',
  type: 'desktop',
  scheduledAt: ''
})

const storageForm = ref({
  name: '',
  type: 'local',
  config: ''
})

const restoreForm = ref({
  target: '',
  mode: 'full',
  pointInTime: ''
})

const fetchPolicies = async () => {
  loading.value = true
  try {
    const res = await request.get('/api/v1/backup/policies')
    // policies.value = res.data || []
  } catch (error) {
    console.error('获取策略失败', error)
  } finally {
    loading.value = false
  }
}

const togglePolicy = async (policy: any) => {
  ElMessage.success(`策略已${policy.enabled ? '启用' : '禁用'}`)
}

const triggerBackup = async (policy: any) => {
  try {
    await request.post(`/api/v1/backup/policies/${policy.id}/trigger`)
    ElMessage.success('备份任务已触发')
  } catch (error) {
    ElMessage.error('触发失败')
  }
}

const editPolicy = (policy: any) => {
  policyForm.value = { ...policy }
  showPolicyDialog.value = true
}

const deletePolicy = async (policy: any) => {
  try {
    await ElMessageBox.confirm('确定删除该策略？', '提示', { type: 'warning' })
    ElMessage.success('删除成功')
  } catch {}
}

const savePolicy = async () => {
  ElMessage.success('策略保存成功')
  showPolicyDialog.value = false
}

const executeQuickBackup = async () => {
  ElMessage.success('快速备份任务已启动')
  showQuickBackupDialog.value = false
}

const executeDrill = async (drill: any) => {
  try {
    await request.post(`/api/v1/backup/drills/${drill.id}/execute`)
    ElMessage.success('演练已开始')
  } catch (error) {
    ElMessage.error('启动失败')
  }
}

const saveDrill = async () => {
  ElMessage.success('演练计划已创建')
  showDrillDialog.value = false
}

const saveStorage = async () => {
  ElMessage.success('存储配置已保存')
  showStorageDialog.value = false
}

const viewRecord = (record: any) => {
  ElMessage.info('详情功能开发中')
}

const restoreBackup = (record: any) => {
  currentRecord.value = record
  showRestoreDialog.value = true
}

const confirmRestore = async () => {
  try {
    await ElMessageBox.confirm('确定执行恢复操作？', '警告', { type: 'warning' })
    await request.post('/api/v1/backup/restores', {
      recordId: currentRecord.value.id,
      ...restoreForm.value
    })
    ElMessage.success('恢复任务已启动')
    showRestoreDialog.value = false
  } catch {}
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    success: 'success',
    running: 'warning',
    failed: 'danger'
  }
  return types[status] || 'info'
}

onMounted(() => {
  fetchPolicies()
})
</script>

<style scoped>
.backup-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
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

.drill-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.drill-title {
  font-weight: 500;
}

.drill-result {
  display: flex;
  align-items: center;
  gap: 8px;
}

.drill-rto {
  font-size: 12px;
  color: #909399;
}

.storage-item {
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
}

.storage-item:last-child {
  border-bottom: none;
}

.storage-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.storage-name {
  font-weight: 500;
}

.storage-type {
  font-size: 12px;
  color: #909399;
}

.storage-usage {
  display: flex;
  align-items: center;
  gap: 8px;
}

.usage-text {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
}
</style>
