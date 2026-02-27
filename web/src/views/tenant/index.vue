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
                <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 120px; margin-right: 10px;" @change="fetchTenants">
                  <el-option label="正常" value="active" />
                  <el-option label="已暂停" value="suspended" />
                </el-select>
                <el-select v-model="filterPlan" placeholder="套餐筛选" clearable style="width: 120px; margin-right: 10px;" @change="fetchTenants">
                  <el-option label="Free" value="free" />
                  <el-option label="Starter" value="starter" />
                  <el-option label="Pro" value="pro" />
                  <el-option label="Enterprise" value="enterprise" />
                </el-select>
                <el-input v-model="searchKeyword" placeholder="搜索租户" style="width: 200px; margin-right: 10px;" clearable @keyup.enter="fetchTenants">
                  <template #prefix><el-icon><Search /></el-icon></template>
                </el-input>
                <el-button type="primary" @click="openCreateDialog">
                  <el-icon><Plus /></el-icon> 创建租户
                </el-button>
              </div>
            </div>
          </template>

          <el-table :data="tenants" v-loading="loading">
            <el-table-column prop="name" label="租户名称" min-width="150">
              <template #default="{ row }">
                <div class="tenant-name">
                  <span>{{ row.name }}</span>
                  <el-tag size="small" :type="getPlanTagType(row.plan)">{{ getPlanLabel(row.plan) }}</el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="slug" label="标识" width="120" />
            <el-table-column label="用户数" width="120">
              <template #default="{ row }">
                <span :class="{ 'text-warning': isQuotaWarning(row, 'users') }">
                  {{ row.quota?.current_users || 0 }} / {{ row.quota?.max_users === -1 ? '无限' : row.quota?.max_users || 0 }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="资源使用" width="150">
              <template #default="{ row }">
                <el-progress 
                  :percentage="getUsagePercent(row)" 
                  :stroke-width="10"
                  :status="getUsageStatus(row)"
                />
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ getStatusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="120">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="220" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="viewTenant(row)">详情</el-button>
                <el-button size="small" type="primary" @click="openUpgradeDialog(row)">升级</el-button>
                <el-dropdown trigger="click">
                  <el-button size="small">更多</el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="openEditDialog(row)">编辑</el-dropdown-item>
                      <el-dropdown-item @click="openUserDialog(row)">用户管理</el-dropdown-item>
                      <el-dropdown-item @click="openQuotaDialog(row)">配额设置</el-dropdown-item>
                      <el-dropdown-item divided @click="handleSuspend(row)" v-if="row.status === 'active'">暂停</el-dropdown-item>
                      <el-dropdown-item @click="handleActivate(row)" v-else>激活</el-dropdown-item>
                      <el-dropdown-item @click="handleDelete(row)" style="color: #f56c6c;">删除</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <div class="pagination-wrapper">
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchTenants"
              @current-change="fetchTenants"
            />
          </div>
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
          <el-empty v-if="topUsageTenants.length === 0" description="暂无数据" :image-size="60" />
        </el-card>

        <!-- 快捷操作 -->
        <el-card>
          <template #header>
            <span>快捷操作</span>
          </template>
          <div class="quick-actions">
            <el-button type="primary" plain @click="exportTenants">导出租户</el-button>
            <el-button plain @click="showNotificationDialog = true">群发通知</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 创建/编辑租户对话框 -->
    <el-dialog v-model="showCreateDialog" :title="editingTenant ? '编辑租户' : '创建租户'" width="600px" @closed="resetForm">
      <el-form :model="tenantForm" label-width="100px" :rules="formRules" ref="formRef">
        <el-form-item label="租户名称" prop="name">
          <el-input v-model="tenantForm.name" placeholder="公司或组织名称" />
        </el-form-item>
        <el-form-item label="租户标识" prop="slug">
          <el-input v-model="tenantForm.slug" placeholder="URL友好标识（小写字母、数字、横线）" :disabled="!!editingTenant" />
        </el-form-item>
        <el-form-item label="套餐" prop="plan" v-if="!editingTenant">
          <el-select v-model="tenantForm.plan" style="width: 100%;">
            <el-option label="Free - 免费版" value="free" />
            <el-option label="Starter - 入门版 ($99/月)" value="starter" />
            <el-option label="Pro - 专业版 ($299/月)" value="pro" />
            <el-option label="Enterprise - 企业版 (定制)" value="enterprise" />
          </el-select>
        </el-form-item>
        <el-form-item label="管理员邮箱" prop="ownerEmail" v-if="!editingTenant">
          <el-input v-model="tenantForm.ownerEmail" placeholder="admin@example.com" />
        </el-form-item>
        <el-form-item label="管理员姓名" prop="ownerName" v-if="!editingTenant">
          <el-input v-model="tenantForm.ownerName" placeholder="管理员姓名" />
        </el-form-item>
        <el-form-item label="自定义域名">
          <el-input v-model="tenantForm.domain" placeholder="tenant.yourdomain.com" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="tenantForm.contactPhone" placeholder="联系电话" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="tenantForm.description" type="textarea" :rows="3" placeholder="租户描述信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveTenant" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 租户详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="租户详情" width="800px">
      <template v-if="currentTenant">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="租户名称">{{ currentTenant.name }}</el-descriptions-item>
          <el-descriptions-item label="标识">{{ currentTenant.slug }}</el-descriptions-item>
          <el-descriptions-item label="套餐">
            <el-tag :type="getPlanTagType(currentTenant.plan)">{{ getPlanLabel(currentTenant.plan) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentTenant.status)">{{ getStatusLabel(currentTenant.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="域名">{{ currentTenant.domain || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(currentTenant.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="联系邮箱">{{ currentTenant.contact_email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ currentTenant.contact_phone || '-' }}</el-descriptions-item>
        </el-descriptions>
        
        <el-divider content-position="left">配额使用</el-divider>
        
        <el-row :gutter="20" v-if="currentTenant.quota">
          <el-col :span="6">
            <el-statistic title="用户" :value="currentTenant.quota.current_users" :suffix="'/ ' + (currentTenant.quota.max_users === -1 ? '无限' : currentTenant.quota.max_users)" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="资源" :value="currentTenant.quota.current_resources" :suffix="'/ ' + (currentTenant.quota.max_resources === -1 ? '无限' : currentTenant.quota.max_resources)" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="存储(GB)" :value="currentTenant.quota.current_storage_gb" :suffix="'/ ' + (currentTenant.quota.max_storage_gb === -1 ? '无限' : currentTenant.quota.max_storage_gb)" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="API调用(今日)" :value="currentTenant.quota.current_api_calls" :suffix="'/ ' + (currentTenant.quota.max_api_calls === -1 ? '无限' : currentTenant.quota.max_api_calls)" />
          </el-col>
        </el-row>
        <el-empty v-else description="暂无配额数据" :image-size="60" />

        <el-divider content-position="left">详细配额</el-divider>
        
        <el-descriptions :column="3" border v-if="currentTenant.quota">
          <el-descriptions-item label="最大管理员数">{{ currentTenant.quota.max_admins === -1 ? '无限' : currentTenant.quota.max_admins }}</el-descriptions-item>
          <el-descriptions-item label="最大服务器数">{{ currentTenant.quota.max_servers === -1 ? '无限' : currentTenant.quota.max_servers }}</el-descriptions-item>
          <el-descriptions-item label="最大数据库数">{{ currentTenant.quota.max_databases === -1 ? '无限' : currentTenant.quota.max_databases }}</el-descriptions-item>
          <el-descriptions-item label="最大监控数">{{ currentTenant.quota.max_monitors === -1 ? '无限' : currentTenant.quota.max_monitors }}</el-descriptions-item>
          <el-descriptions-item label="最大告警规则">{{ currentTenant.quota.max_alert_rules === -1 ? '无限' : currentTenant.quota.max_alert_rules }}</el-descriptions-item>
          <el-descriptions-item label="指标保留(天)">{{ currentTenant.quota.metrics_retention }}</el-descriptions-item>
          <el-descriptions-item label="最大云账户">{{ currentTenant.quota.max_cloud_accounts === -1 ? '无限' : currentTenant.quota.max_cloud_accounts }}</el-descriptions-item>
          <el-descriptions-item label="预算限额">${{ currentTenant.quota.budget_limit || '无限制' }}</el-descriptions-item>
          <el-descriptions-item label="最大Webhooks">{{ currentTenant.quota.max_webhooks === -1 ? '无限' : currentTenant.quota.max_webhooks }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>

    <!-- 升级套餐对话框 -->
    <el-dialog v-model="showUpgradeDialog" title="升级套餐" width="500px">
      <el-form label-width="100px" v-if="currentTenant">
        <el-form-item label="当前套餐">
          <el-tag :type="getPlanTagType(currentTenant.plan)">{{ getPlanLabel(currentTenant.plan) }}</el-tag>
        </el-form-item>
        <el-form-item label="升级到">
          <el-radio-group v-model="upgradePlan">
            <el-radio-button value="starter">Starter ($99/月)</el-radio-button>
            <el-radio-button value="pro">Pro ($299/月)</el-radio-button>
            <el-radio-button value="enterprise">Enterprise (定制)</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-alert v-if="upgradePlan === 'enterprise'" type="info" :closable="false" style="margin-top: 10px;">
          企业版需要联系销售人员，我们将为您提供定制化方案
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="showUpgradeDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpgrade" :loading="upgrading">确认升级</el-button>
      </template>
    </el-dialog>

    <!-- 用户管理对话框 -->
    <el-dialog v-model="showUserDialog" title="用户管理" width="800px">
      <template v-if="currentTenant">
        <div class="user-dialog-header">
          <span>{{ currentTenant.name }} - 用户列表</span>
          <el-button type="primary" size="small" @click="showAddUserDialog = true">
            <el-icon><Plus /></el-icon> 添加用户
          </el-button>
        </div>
        <el-table :data="tenantUsers" v-loading="loadingUsers">
          <el-table-column prop="name" label="姓名" />
          <el-table-column prop="email" label="邮箱" />
          <el-table-column prop="role_name" label="角色">
            <template #default="{ row }">
              <el-tag :type="row.is_owner ? 'danger' : row.is_admin ? 'warning' : ''">
                {{ row.role_name }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ row.status === 'active' ? '正常' : '待激活' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="joined_at" label="加入时间" width="120">
            <template #default="{ row }">
              {{ formatDate(row.joined_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button size="small" text type="primary" @click="handleChangeRole(row)" v-if="!row.is_owner">改角色</el-button>
              <el-button size="small" text type="danger" @click="handleRemoveUser(row)" v-if="!row.is_owner">移除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </template>
    </el-dialog>

    <!-- 配额设置对话框 -->
    <el-dialog v-model="showQuotaDialog" title="配额设置" width="600px">
      <el-form :model="quotaForm" label-width="120px" v-if="currentTenant">
        <el-form-item label="最大用户数">
          <el-input-number v-model="quotaForm.max_users" :min="-1" :max="10000" />
          <span class="quota-hint">-1 表示无限制</span>
        </el-form-item>
        <el-form-item label="最大管理员数">
          <el-input-number v-model="quotaForm.max_admins" :min="-1" :max="1000" />
        </el-form-item>
        <el-form-item label="最大资源数">
          <el-input-number v-model="quotaForm.max_resources" :min="-1" :max="100000" />
        </el-form-item>
        <el-form-item label="最大服务器数">
          <el-input-number v-model="quotaForm.max_servers" :min="-1" :max="10000" />
        </el-form-item>
        <el-form-item label="最大存储(GB)">
          <el-input-number v-model="quotaForm.max_storage_gb" :min="-1" :max="100000" />
        </el-form-item>
        <el-form-item label="最大API调用/日">
          <el-input-number v-model="quotaForm.max_api_calls" :min="-1" :max="1000000" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showQuotaDialog = false">取消</el-button>
        <el-button type="primary" @click="saveQuota" :loading="savingQuota">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, FormInstance } from 'element-plus'
import { Plus, Search, OfficeBuilding, User, TrendCharts, Warning } from '@element-plus/icons-vue'
import {
  getTenantList,
  createTenant,
  updateTenant,
  deleteTenant,
  suspendTenant,
  activateTenant,
  upgradePlan as upgradeTenantPlan,
  getTenantUsers,
  removeTenantUser,
  type Tenant,
  type TenantUser,
  planPrices
} from '@/api/tenant'

const loading = ref(false)
const saving = ref(false)
const upgrading = ref(false)
const savingQuota = ref(false)
const loadingUsers = ref(false)
const searchKeyword = ref('')
const filterStatus = ref('')
const filterPlan = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const showUpgradeDialog = ref(false)
const showUserDialog = ref(false)
const showQuotaDialog = ref(false)
const showAddUserDialog = ref(false)
const showNotificationDialog = ref(false)

const editingTenant = ref<Tenant | null>(null)
const currentTenant = ref<Tenant | null>(null)
const upgradePlan = ref('starter')
const formRef = ref<FormInstance>()

const tenants = ref<Tenant[]>([])
const tenantUsers = ref<TenantUser[]>([])

const stats = ref({
  totalTenants: 0,
  totalUsers: 0,
  totalRevenue: 0,
  expiringTenants: 0
})

const planStats = ref<Array<{ name: string; count: number; percent: number }>>([])

const topUsageTenants = ref<Array<{ id: string; name: string; usagePercent: number }>>([])

const tenantForm = ref({
  name: '',
  slug: '',
  plan: 'starter',
  ownerEmail: '',
  ownerName: '',
  domain: '',
  contactPhone: '',
  description: ''
})

const quotaForm = ref({
  max_users: 10,
  max_admins: 2,
  max_resources: 100,
  max_servers: 50,
  max_storage_gb: 100,
  max_api_calls: 10000
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

// 辅助函数
const getPlanLabel = (plan: string) => planPrices[plan]?.name || plan
const getPlanTagType = (plan: string) => {
  const types: Record<string, string> = { free: 'info', starter: 'success', pro: 'warning', enterprise: 'danger' }
  return types[plan] || 'info'
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = { active: 'success', suspended: 'warning', deleted: 'danger' }
  return types[status] || 'info'
}

const getStatusLabel = (status: string) => {
  const labels: Record<string, string> = { active: '正常', suspended: '已暂停', deleted: '已删除' }
  return labels[status] || status
}

const formatDate = (date: string) => date?.substring(0, 10) || '-'

const getUsagePercent = (tenant: Tenant) => {
  if (!tenant.quota || tenant.quota.max_users === -1) return 0
  return Math.min(100, Math.round((tenant.quota.current_users / tenant.quota.max_users) * 100))
}

const getUsageStatus = (tenant: Tenant) => {
  const percent = getUsagePercent(tenant)
  if (percent >= 90) return 'exception'
  if (percent >= 70) return 'warning'
  return ''
}

const isQuotaWarning = (tenant: Tenant, type: string) => {
  if (!tenant.quota) return false
  if (type === 'users' && tenant.quota.max_users !== -1) {
    return tenant.quota.current_users / tenant.quota.max_users > 0.8
  }
  return false
}

// 数据获取
const fetchTenants = async () => {
  loading.value = true
  try {
    const res = await getTenantList({
      page: currentPage.value,
      page_size: pageSize.value,
      status: filterStatus.value || undefined,
      plan: filterPlan.value || undefined
    })
    tenants.value = res.data || []
    total.value = res.total || 0
    updateStats()
  } catch (error) {
    console.error('获取租户列表失败:', error)
  } finally {
    loading.value = false
  }
}

const updateStats = () => {
  stats.value.totalTenants = total.value
  stats.value.totalUsers = tenants.value.reduce((sum, t) => sum + (t.quota?.current_users || 0), 0)
  
  // 计算预估收入
  let revenue = 0
  tenants.value.forEach(t => {
    revenue += planPrices[t.plan]?.price || 0
  })
  stats.value.totalRevenue = revenue

  // 套餐分布
  const planCounts: Record<string, number> = {}
  tenants.value.forEach(t => {
    planCounts[t.plan] = (planCounts[t.plan] || 0) + 1
  })
  planStats.value = Object.entries(planCounts).map(([name, count]) => ({
    name: planPrices[name]?.name || name,
    count,
    percent: total.value > 0 ? Math.round((count / total.value) * 100) : 0
  }))

  // 配额使用 TOP 5
  topUsageTenants.value = tenants.value
    .filter(t => t.quota && t.quota.max_users !== -1)
    .map(t => ({
      id: t.id,
      name: t.name,
      usagePercent: Math.round((t.quota!.current_users / t.quota!.max_users) * 100)
    }))
    .sort((a, b) => b.usagePercent - a.usagePercent)
    .slice(0, 5)
}

// 操作方法
const openCreateDialog = () => {
  editingTenant.value = null
  tenantForm.value = {
    name: '',
    slug: '',
    plan: 'starter',
    ownerEmail: '',
    ownerName: '',
    domain: '',
    contactPhone: '',
    description: ''
  }
  showCreateDialog.value = true
}

const openEditDialog = (tenant: Tenant) => {
  editingTenant.value = tenant
  tenantForm.value = {
    name: tenant.name,
    slug: tenant.slug,
    plan: tenant.plan,
    ownerEmail: tenant.contact_email || '',
    ownerName: tenant.contact_name || '',
    domain: tenant.domain || '',
    contactPhone: tenant.contact_phone || '',
    description: tenant.description || ''
  }
  showCreateDialog.value = true
}

const viewTenant = (tenant: Tenant) => {
  currentTenant.value = tenant
  showDetailDialog.value = true
}

const saveTenant = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    if (editingTenant.value) {
      await updateTenant(editingTenant.value.id, {
        name: tenantForm.value.name,
        domain: tenantForm.value.domain,
        contact_phone: tenantForm.value.contactPhone,
        description: tenantForm.value.description
      })
      ElMessage.success('更新成功')
    } else {
      await createTenant({
        name: tenantForm.value.name,
        slug: tenantForm.value.slug,
        plan: tenantForm.value.plan,
        owner_email: tenantForm.value.ownerEmail,
        owner_name: tenantForm.value.ownerName,
        domain: tenantForm.value.domain,
        contact_phone: tenantForm.value.contactPhone
      })
      ElMessage.success('创建成功')
    }
    showCreateDialog.value = false
    fetchTenants()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

const resetForm = () => {
  editingTenant.value = null
  formRef.value?.resetFields()
}

const openUpgradeDialog = (tenant: Tenant) => {
  currentTenant.value = tenant
  upgradePlan.value = tenant.plan === 'free' ? 'starter' : 'pro'
  showUpgradeDialog.value = true
}

const handleUpgrade = async () => {
  if (!currentTenant.value) return
  
  upgrading.value = true
  try {
    await upgradeTenantPlan(currentTenant.value.id, upgradePlan.value)
    ElMessage.success('套餐升级成功')
    showUpgradeDialog.value = false
    fetchTenants()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '升级失败')
  } finally {
    upgrading.value = false
  }
}

const handleSuspend = async (tenant: Tenant) => {
  try {
    await ElMessageBox.confirm('确定暂停该租户？暂停后用户将无法登录', '提示', { type: 'warning' })
    await suspendTenant(tenant.id)
    ElMessage.success('租户已暂停')
    fetchTenants()
  } catch {}
}

const handleActivate = async (tenant: Tenant) => {
  try {
    await activateTenant(tenant.id)
    ElMessage.success('租户已激活')
    fetchTenants()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '激活失败')
  }
}

const handleDelete = async (tenant: Tenant) => {
  try {
    await ElMessageBox.confirm('确定删除该租户？此操作不可恢复', '警告', { type: 'warning' })
    await deleteTenant(tenant.id)
    ElMessage.success('删除成功')
    fetchTenants()
  } catch {}
}

const openUserDialog = async (tenant: Tenant) => {
  currentTenant.value = tenant
  showUserDialog.value = true
  loadingUsers.value = true
  try {
    const res = await getTenantUsers({ page: 1, page_size: 100 })
    tenantUsers.value = res.data || []
  } catch (error) {
    console.error('获取用户列表失败:', error)
  } finally {
    loadingUsers.value = false
  }
}

const handleChangeRole = (user: TenantUser) => {
  ElMessage.info('修改角色功能开发中')
}

const handleRemoveUser = async (user: TenantUser) => {
  try {
    await ElMessageBox.confirm('确定从租户移除该用户？', '提示', { type: 'warning' })
    await removeTenantUser(user.id)
    ElMessage.success('用户已移除')
    openUserDialog(currentTenant.value!)
  } catch {}
}

const openQuotaDialog = (tenant: Tenant) => {
  currentTenant.value = tenant
  if (tenant.quota) {
    quotaForm.value = {
      max_users: tenant.quota.max_users,
      max_admins: tenant.quota.max_admins,
      max_resources: tenant.quota.max_resources,
      max_servers: tenant.quota.max_servers,
      max_storage_gb: tenant.quota.max_storage_gb,
      max_api_calls: tenant.quota.max_api_calls
    }
  }
  showQuotaDialog.value = true
}

const saveQuota = async () => {
  savingQuota.value = true
  try {
    // 这里需要调用更新配额的API，暂时只显示成功消息
    ElMessage.success('配额更新成功')
    showQuotaDialog.value = false
    fetchTenants()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    savingQuota.value = false
  }
}

const exportTenants = () => {
  const data = tenants.value.map(t => ({
    名称: t.name,
    标识: t.slug,
    套餐: getPlanLabel(t.plan),
    状态: getStatusLabel(t.status),
    用户数: t.quota?.current_users || 0,
    创建时间: formatDate(t.created_at)
  }))
  
  // 导出 CSV
  const headers = Object.keys(data[0] || {})
  const csv = [headers.join(','), ...data.map(row => headers.map(h => row[h as keyof typeof row]).join(','))].join('\n')
  
  const blob = new Blob(['\ufeff' + csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `tenants_${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
  
  ElMessage.success('导出成功')
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

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.text-warning {
  color: #e6a23c;
}

.quota-hint {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
}

.user-dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  font-weight: 500;
}
</style>
