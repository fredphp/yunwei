<template>
  <div class="tenant-audit-page">
    <!-- 筛选条件 -->
    <el-card class="mb-4">
      <el-form :inline="true" :model="filterForm">
        <el-form-item label="租户">
          <el-select v-model="filterForm.tenantId" placeholder="选择租户" clearable style="width: 200px;">
            <el-option v-for="t in tenants" :key="t.id" :label="t.name" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作类型">
          <el-select v-model="filterForm.action" placeholder="选择类型" clearable style="width: 150px;">
            <el-option label="登录" value="login" />
            <el-option label="创建" value="create" />
            <el-option label="更新" value="update" />
            <el-option label="删除" value="delete" />
            <el-option label="导出" value="export" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源类型">
          <el-select v-model="filterForm.resource" placeholder="选择资源" clearable style="width: 150px;">
            <el-option label="服务器" value="server" />
            <el-option label="用户" value="user" />
            <el-option label="角色" value="role" />
            <el-option label="配置" value="config" />
            <el-option label="账单" value="billing" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 260px;"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filterForm.status" placeholder="选择状态" clearable style="width: 120px;">
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchLogs">
            <el-icon><Search /></el-icon> 查询
          </el-button>
          <el-button @click="resetFilter">
            <el-icon><Refresh /></el-icon> 重置
          </el-button>
          <el-button @click="exportLogs">
            <el-icon><Download /></el-icon> 导出
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 日志列表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>审计日志</span>
          <div class="header-stats">
            <el-tag type="success">成功: {{ stats.success }}</el-tag>
            <el-tag type="danger">失败: {{ stats.failed }}</el-tag>
            <el-tag type="info">总计: {{ total }}</el-tag>
          </div>
        </div>
      </template>

      <el-table :data="logs" v-loading="loading" @row-click="viewLogDetail">
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="tenant_name" label="租户" width="150">
          <template #default="{ row }">
            <el-tag size="small">{{ row.tenant_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="user_name" label="操作用户" width="120">
          <template #default="{ row }">
            <div class="user-info">
              <el-avatar :size="24" style="margin-right: 8px;">{{ row.user_name?.charAt(0) }}</el-avatar>
              <div>
                <div>{{ row.user_name }}</div>
                <div class="user-email">{{ row.user_email }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="action" label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getActionType(row.action)" size="small">
              {{ getActionLabel(row.action) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resource" label="资源类型" width="100" />
        <el-table-column prop="resource_name" label="资源名称" min-width="150">
          <template #default="{ row }">
            <span class="resource-name">{{ row.resource_name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.status === 'success'" class="status-success"><CircleCheck /></el-icon>
            <el-icon v-else class="status-failed"><CircleClose /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP地址" width="130" />
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click.stop="viewLogDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100, 200]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchLogs"
          @current-change="fetchLogs"
        />
      </div>
    </el-card>

    <!-- 日志详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="日志详情" width="700px">
      <template v-if="currentLog">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="时间">{{ formatDateTime(currentLog.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="租户">{{ currentLog.tenant_name }}</el-descriptions-item>
          <el-descriptions-item label="操作用户">{{ currentLog.user_name }}</el-descriptions-item>
          <el-descriptions-item label="用户邮箱">{{ currentLog.user_email }}</el-descriptions-item>
          <el-descriptions-item label="操作类型">
            <el-tag :type="getActionType(currentLog.action)">{{ getActionLabel(currentLog.action) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="currentLog.status === 'success' ? 'success' : 'danger'">
              {{ currentLog.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="资源类型">{{ currentLog.resource }}</el-descriptions-item>
          <el-descriptions-item label="资源ID">{{ currentLog.resource_id }}</el-descriptions-item>
          <el-descriptions-item label="资源名称" :span="2">{{ currentLog.resource_name }}</el-descriptions-item>
          <el-descriptions-item label="IP地址">{{ currentLog.ip_address }}</el-descriptions-item>
          <el-descriptions-item label="请求ID">{{ currentLog.request_id }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left" v-if="currentLog.old_value || currentLog.new_value">变更内容</el-divider>

        <el-row :gutter="20" v-if="currentLog.old_value || currentLog.new_value">
          <el-col :span="12" v-if="currentLog.old_value">
            <div class="change-label">变更前:</div>
            <el-input type="textarea" :rows="6" :model-value="JSON.stringify(currentLog.old_value, null, 2)" readonly />
          </el-col>
          <el-col :span="12" v-if="currentLog.new_value">
            <div class="change-label">变更后:</div>
            <el-input type="textarea" :rows="6" :model-value="JSON.stringify(currentLog.new_value, null, 2)" readonly />
          </el-col>
        </el-row>

        <el-divider content-position="left" v-if="currentLog.error_msg">错误信息</el-divider>

        <el-alert v-if="currentLog.error_msg" type="error" :closable="false" show-icon>
          {{ currentLog.error_msg }}
        </el-alert>

        <el-divider content-position="left">请求信息</el-divider>

        <el-descriptions :column="1" border>
          <el-descriptions-item label="User-Agent">{{ currentLog.user_agent }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh, Download, CircleCheck, CircleClose } from '@element-plus/icons-vue'

const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const showDetailDialog = ref(false)
const currentLog = ref<any>(null)

const filterForm = ref({
  tenantId: '',
  action: '',
  resource: '',
  status: '',
  dateRange: [] as Date[]
})

const stats = ref({
  success: 156,
  failed: 12
})

const tenants = ref([
  { id: '1', name: 'Acme Corporation' },
  { id: '2', name: 'TechStart Inc' },
  { id: '3', name: 'Global Solutions' },
  { id: '4', name: 'Demo Company' },
  { id: '5', name: 'StartupXYZ' }
])

const logs = ref([
  {
    id: '1',
    created_at: '2024-03-15T14:30:00',
    tenant_id: '1',
    tenant_name: 'Acme Corporation',
    user_id: '101',
    user_name: 'John Smith',
    user_email: 'john@acme.com',
    action: 'login',
    resource: 'user',
    resource_id: '101',
    resource_name: 'John Smith',
    status: 'success',
    ip_address: '192.168.1.100',
    user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    request_id: 'req-001'
  },
  {
    id: '2',
    created_at: '2024-03-15T14:25:00',
    tenant_id: '1',
    tenant_name: 'Acme Corporation',
    user_id: '101',
    user_name: 'John Smith',
    user_email: 'john@acme.com',
    action: 'create',
    resource: 'server',
    resource_id: 's-001',
    resource_name: 'production-web-01',
    status: 'success',
    ip_address: '192.168.1.100',
    old_value: null,
    new_value: { name: 'production-web-01', ip: '10.0.0.10', type: 'web' },
    user_agent: 'Mozilla/5.0',
    request_id: 'req-002'
  },
  {
    id: '3',
    created_at: '2024-03-15T14:20:00',
    tenant_id: '2',
    tenant_name: 'TechStart Inc',
    user_id: '201',
    user_name: 'Jane Doe',
    user_email: 'jane@techstart.io',
    action: 'update',
    resource: 'config',
    resource_id: 'cfg-001',
    resource_name: '系统配置',
    status: 'success',
    ip_address: '192.168.1.101',
    old_value: { max_users: 10 },
    new_value: { max_users: 15 },
    user_agent: 'Mozilla/5.0',
    request_id: 'req-003'
  },
  {
    id: '4',
    created_at: '2024-03-15T14:15:00',
    tenant_id: '3',
    tenant_name: 'Global Solutions',
    user_id: '301',
    user_name: 'Mike Johnson',
    user_email: 'mike@global.com',
    action: 'delete',
    resource: 'server',
    resource_id: 's-002',
    resource_name: 'test-server-01',
    status: 'failed',
    ip_address: '192.168.1.102',
    error_msg: '服务器正在运行，无法删除',
    user_agent: 'Mozilla/5.0',
    request_id: 'req-004'
  },
  {
    id: '5',
    created_at: '2024-03-15T14:10:00',
    tenant_id: '4',
    tenant_name: 'Demo Company',
    user_id: '401',
    user_name: 'Demo User',
    user_email: 'demo@example.com',
    action: 'export',
    resource: 'billing',
    resource_id: '',
    resource_name: '账单数据导出',
    status: 'success',
    ip_address: '192.168.1.103',
    user_agent: 'Mozilla/5.0',
    request_id: 'req-005'
  }
])

const fetchLogs = async () => {
  loading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))
    total.value = logs.value.length
  } catch (error) {
    console.error('获取日志失败:', error)
  } finally {
    loading.value = false
  }
}

const resetFilter = () => {
  filterForm.value = {
    tenantId: '',
    action: '',
    resource: '',
    status: '',
    dateRange: []
  }
  fetchLogs()
}

const exportLogs = () => {
  ElMessage.success('日志导出中...')
}

const viewLogDetail = (log: any) => {
  currentLog.value = log
  showDetailDialog.value = true
}

const getActionType = (action: string) => {
  const types: Record<string, string> = {
    login: 'primary',
    create: 'success',
    update: 'warning',
    delete: 'danger',
    export: 'info'
  }
  return types[action] || 'info'
}

const getActionLabel = (action: string) => {
  const labels: Record<string, string> = {
    login: '登录',
    create: '创建',
    update: '更新',
    delete: '删除',
    export: '导出'
  }
  return labels[action] || action
}

const formatDateTime = (date: string) => date?.replace('T', ' ').substring(0, 19) || '-'

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.tenant-audit-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-stats {
  display: flex;
  gap: 10px;
}

.user-info {
  display: flex;
  align-items: center;
}

.user-email {
  font-size: 12px;
  color: #909399;
}

.resource-name {
  color: #409eff;
  cursor: pointer;
}

.resource-name:hover {
  text-decoration: underline;
}

.status-success {
  color: #67c23a;
  font-size: 18px;
}

.status-failed {
  color: #f56c6c;
  font-size: 18px;
}

.change-label {
  margin-bottom: 8px;
  font-weight: 500;
  color: #606266;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.mb-4 {
  margin-bottom: 16px;
}
</style>
