<template>
  <div class="tenant-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #409eff;">
              <el-icon size="28"><OfficeBuilding /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.totalTenants }}</div>
              <div class="stat-label">租户总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #67c23a;">
              <el-icon size="28"><User /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.totalUsers }}</div>
              <div class="stat-label">用户总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #e6a23c;">
              <el-icon size="28"><TrendCharts /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">${{ stats.totalRevenue }}/月</div>
              <div class="stat-label">预估收入</div>
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
              <div class="stat-value">{{ stats.expiringTenants }}</div>
              <div class="stat-label">即将到期</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <!-- 租户列表 -->
      <el-col :span="18">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>租户管理</span>
              <div class="header-actions">
                <el-input v-model="searchKeyword" placeholder="搜索租户" style="width: 200px; margin-right: 10px;" clearable>
                  <template #prefix><el-icon><Search /></el-icon></template>
                </el-input>
                <el-button type="primary" @click="showCreateDialog = true">
                  <el-icon><Plus /></el-icon> 创建租户
                </el-button>
              </div>
            </div>
          </template>

          <el-table :data="filteredTenants" v-loading="loading">
            <el-table-column prop="name" label="租户名称" min-width="150">
              <template #default="{ row }">
                <div class="tenant-name">
                  <span>{{ row.name }}</span>
                  <el-tag size="small" v-if="row.plan">{{ row.plan }}</el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="slug" label="标识" width="120" />
            <el-table-column label="用户数" width="100">
              <template #default="{ row }">
                {{ row.userCount }} / {{ row.maxUsers }}
              </template>
            </el-table-column>
            <el-table-column prop="usagePercent" label="资源使用" width="120">
              <template #default="{ row }">
                <el-progress :percentage="row.usagePercent" :stroke-width="10" />
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" width="120" />
            <el-table-column label="操作" width="220" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="viewTenant(row)">详情</el-button>
                <el-button size="small" type="primary" @click="upgradeTenant(row)">升级</el-button>
                <el-dropdown trigger="click">
                  <el-button size="small">更多</el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="editTenant(row)">编辑</el-dropdown-item>
                      <el-dropdown-item @click="viewUsers(row)">用户管理</el-dropdown-item>
                      <el-dropdown-item @click="viewQuota(row)">配额设置</el-dropdown-item>
                      <el-dropdown-item @click="viewAuditLogs(row)">审计日志</el-dropdown-item>
                      <el-dropdown-item divided @click="suspendTenant(row)" v-if="row.status === 'active'">暂停</el-dropdown-item>
                      <el-dropdown-item @click="activateTenant(row)" v-else>激活</el-dropdown-item>
                      <el-dropdown-item @click="deleteTenant(row)" style="color: #f56c6c;">删除</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- 右侧面板 -->
      <el-col :span="6">
        <!-- 套餐分布 -->
        <el-card class="mb-4">
          <template #header>
            <span>套餐分布</span>
          </template>
          <div v-for="plan in planStats" :key="plan.name" class="plan-item">
            <div class="plan-info">
              <span class="plan-name">{{ plan.name }}</span>
              <span class="plan-count">{{ plan.count }} 个租户</span>
            </div>
            <el-progress :percentage="plan.percent" :stroke-width="8" />
          </div>
        </el-card>

        <!-- 配额使用 TOP 5 -->
        <el-card class="mb-4">
          <template #header>
            <span>配额使用 TOP 5</span>
          </template>
          <div v-for="tenant in topUsageTenants" :key="tenant.id" class="usage-item">
            <div class="usage-header">
              <span>{{ tenant.name }}</span>
              <el-tag size="small" :type="tenant.usagePercent > 80 ? 'danger' : 'warning'">
                {{ tenant.usagePercent }}%
              </el-tag>
            </div>
            <el-progress :percentage="tenant.usagePercent" :show-text="false" />
          </div>
        </el-card>

        <!-- 快捷操作 -->
        <el-card>
          <template #header>
            <span>快捷操作</span>
          </template>
          <div class="quick-actions">
            <el-button type="primary" plain @click="exportTenants">导出租户</el-button>
            <el-button plain @click="sendNotification">群发通知</el-button>
            <el-button plain @click="viewReports">收入报表</el-button>
            <el-button plain @click="viewAuditLogs">全局审计</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 创建/编辑租户对话框 -->
    <el-dialog v-model="showCreateDialog" :title="editingTenant ? '编辑租户' : '创建租户'" width="600px">
      <el-form :model="tenantForm" label-width="100px" :rules="formRules" ref="formRef">
        <el-form-item label="租户名称" prop="name">
          <el-input v-model="tenantForm.name" placeholder="公司或组织名称" />
        </el-form-item>
        <el-form-item label="租户标识" prop="slug">
          <el-input v-model="tenantForm.slug" placeholder="URL友好标识" />
        </el-form-item>
        <el-form-item label="套餐" prop="plan">
          <el-select v-model="tenantForm.plan" style="width: 100%;">
            <el-option label="Free - 免费版" value="free" />
            <el-option label="Starter - 入门版 ($99/月)" value="starter" />
            <el-option label="Pro - 专业版 ($299/月)" value="pro" />
            <el-option label="Enterprise - 企业版 (定制)" value="enterprise" />
          </el-select>
        </el-form-item>
        <el-form-item label="管理员邮箱" prop="ownerEmail">
          <el-input v-model="tenantForm.ownerEmail" placeholder="admin@example.com" />
        </el-form-item>
        <el-form-item label="管理员姓名" prop="ownerName">
          <el-input v-model="tenantForm.ownerName" />
        </el-form-item>
        <el-form-item label="自定义域名">
          <el-input v-model="tenantForm.domain" placeholder="tenant.yourdomain.com" />
        </el-form-item>
        <el-form-item label="联系信息">
          <el-input v-model="tenantForm.contactPhone" placeholder="联系电话" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveTenant">保存</el-button>
      </template>
    </el-dialog>

    <!-- 租户详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="租户详情" width="800px">
      <el-descriptions :column="2" border v-if="currentTenant">
        <el-descriptions-item label="租户名称">{{ currentTenant.name }}</el-descriptions-item>
        <el-descriptions-item label="标识">{{ currentTenant.slug }}</el-descriptions-item>
        <el-descriptions-item label="套餐">
          <el-tag>{{ currentTenant.plan }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentTenant.status)">{{ currentTenant.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="域名">{{ currentTenant.domain || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ currentTenant.createdAt }}</el-descriptions-item>
      </el-descriptions>
      
      <el-divider content-position="left">配额使用</el-divider>
      
      <el-row :gutter="20">
        <el-col :span="8">
          <el-statistic title="用户" :value="currentTenant.userCount" :suffix="'/ ' + currentTenant.maxUsers" />
        </el-col>
        <el-col :span="8">
          <el-statistic title="资源" :value="currentTenant.resourceCount" :suffix="'/ ' + currentTenant.maxResources" />
        </el-col>
        <el-col :span="8">
          <el-statistic title="存储" :value="currentTenant.storageUsed" suffix="GB" />
        </el-col>
      </el-row>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, OfficeBuilding, User, TrendCharts, Warning } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const searchKeyword = ref('')
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const editingTenant = ref<any>(null)
const currentTenant = ref<any>(null)

const stats = ref({
  totalTenants: 128,
  totalUsers: 1256,
  totalRevenue: 45600,
  expiringTenants: 5
})

const tenants = ref([
  { id: 1, name: 'Acme Corp', slug: 'acme', plan: 'pro', userCount: 45, maxUsers: 50, usagePercent: 75, status: 'active', createdAt: '2024-01-15' },
  { id: 2, name: 'TechStart Inc', slug: 'techstart', plan: 'starter', userCount: 8, maxUsers: 10, usagePercent: 60, status: 'active', createdAt: '2024-02-01' },
  { id: 3, name: 'Global Solutions', slug: 'global', plan: 'enterprise', userCount: 200, maxUsers: -1, usagePercent: 45, status: 'active', createdAt: '2023-06-20' },
  { id: 4, name: 'Demo Company', slug: 'demo', plan: 'free', userCount: 2, maxUsers: 3, usagePercent: 90, status: 'active', createdAt: '2024-02-10' }
])

const planStats = ref([
  { name: 'Enterprise', count: 8, percent: 6 },
  { name: 'Pro', count: 45, percent: 35 },
  { name: 'Starter', count: 52, percent: 41 },
  { name: 'Free', count: 23, percent: 18 }
])

const topUsageTenants = ref([
  { id: 1, name: 'Demo Company', usagePercent: 95 },
  { id: 2, name: 'TechStart Inc', usagePercent: 88 },
  { id: 3, name: 'Acme Corp', usagePercent: 82 },
  { id: 4, name: 'StartupXYZ', usagePercent: 78 },
  { id: 5, name: 'SmallBiz', usagePercent: 72 }
])

const tenantForm = ref({
  name: '',
  slug: '',
  plan: 'starter',
  ownerEmail: '',
  ownerName: '',
  domain: '',
  contactPhone: ''
})

const formRules = {
  name: [{ required: true, message: '请输入租户名称', trigger: 'blur' }],
  slug: [
    { required: true, message: '请输入租户标识', trigger: 'blur' },
    { pattern: /^[a-z0-9-]+$/, message: '只能包含小写字母、数字和横线', trigger: 'blur' }
  ],
  plan: [{ required: true, message: '请选择套餐', trigger: 'change' }],
  ownerEmail: [
    { required: true, message: '请输入管理员邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  ownerName: [{ required: true, message: '请输入管理员姓名', trigger: 'blur' }]
}

const filteredTenants = computed(() => {
  if (!searchKeyword.value) return tenants.value
  const keyword = searchKeyword.value.toLowerCase()
  return tenants.value.filter((t: any) => 
    t.name.toLowerCase().includes(keyword) ||
    t.slug.toLowerCase().includes(keyword)
  )
})

const fetchTenants = async () => {
  loading.value = true
  try {
    const res = await request.get('/admin/tenants')
    if (res.data && Array.isArray(res.data)) {
      tenants.value = res.data.map((t: any) => ({
        id: t.id,
        name: t.name,
        slug: t.slug,
        plan: t.plan || 'free',
        userCount: t.quota?.current_users || 0,
        maxUsers: t.quota?.max_users || 10,
        usagePercent: t.quota?.max_users ? Math.round((t.quota?.current_users || 0) / t.quota.max_users * 100) : 0,
        status: t.status || 'active',
        createdAt: t.created_at?.substring(0, 10) || ''
      }))
    }
    // 更新统计数据
    stats.value.totalTenants = res.total || tenants.value.length
  } catch (error) {
    // 请求失败时使用本地模拟数据，不显示错误提示
    console.log('使用本地模拟数据')
  } finally {
    loading.value = false
  }
}

const viewTenant = (tenant: any) => {
  currentTenant.value = tenant
  showDetailDialog.value = true
}

const editTenant = (tenant: any) => {
  editingTenant.value = tenant
  tenantForm.value = { ...tenant }
  showCreateDialog.value = true
}

const saveTenant = async () => {
  try {
    if (editingTenant.value) {
      await request.put(`/admin/tenants/${editingTenant.value.id}`, tenantForm.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/admin/tenants', tenantForm.value)
      ElMessage.success('创建成功')
    }
    showCreateDialog.value = false
    resetForm()
    fetchTenants()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const resetForm = () => {
  editingTenant.value = null
  tenantForm.value = {
    name: '',
    slug: '',
    plan: 'starter',
    ownerEmail: '',
    ownerName: '',
    domain: '',
    contactPhone: ''
  }
}

const upgradeTenant = (tenant: any) => {
  ElMessage.info('升级功能开发中')
}

const viewUsers = (tenant: any) => {
  ElMessage.info('用户管理功能开发中')
}

const viewQuota = (tenant: any) => {
  ElMessage.info('配额设置功能开发中')
}

const viewAuditLogs = (tenant: any) => {
  ElMessage.info('审计日志功能开发中')
}

const suspendTenant = async (tenant: any) => {
  try {
    await ElMessageBox.confirm('确定暂停该租户？暂停后用户将无法登录', '提示', { type: 'warning' })
    await request.post(`/admin/tenants/${tenant.id}/suspend`)
    ElMessage.success('租户已暂停')
    fetchTenants()
  } catch {}
}

const activateTenant = async (tenant: any) => {
  try {
    await request.post(`/admin/tenants/${tenant.id}/activate`)
    ElMessage.success('租户已激活')
    fetchTenants()
  } catch {}
}

const deleteTenant = async (tenant: any) => {
  try {
    await ElMessageBox.confirm('确定删除该租户？此操作不可恢复', '警告', { type: 'warning' })
    await request.delete(`/admin/tenants/${tenant.id}`)
    ElMessage.success('删除成功')
    fetchTenants()
  } catch {}
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    active: 'success',
    suspended: 'warning',
    deleted: 'danger'
  }
  return types[status] || 'info'
}

const exportTenants = () => {
  ElMessage.success('导出功能开发中')
}

const sendNotification = () => {
  ElMessage.info('群发通知功能开发中')
}

const viewReports = () => {
  ElMessage.info('收入报表功能开发中')
}

onMounted(() => {
  fetchTenants()
})
</script>

<style scoped>
.tenant-page {
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
  font-size: 24px;
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

.tenant-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.plan-item {
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
}

.plan-item:last-child {
  border-bottom: none;
}

.plan-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.plan-name {
  font-weight: 500;
}

.plan-count {
  font-size: 12px;
  color: #909399;
}

.usage-item {
  padding: 10px 0;
  border-bottom: 1px solid #ebeef5;
}

.usage-item:last-child {
  border-bottom: none;
}

.usage-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
  font-size: 13px;
}

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
</style>
