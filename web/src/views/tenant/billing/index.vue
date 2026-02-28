<template>
  <div class="tenant-billing-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #409eff;">
              <el-icon size="28"><Wallet /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">${{ stats.totalRevenue }}</div>
              <div class="stat-label">本月收入</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #67c23a;">
              <el-icon size="28"><Money /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">${{ stats.paidAmount }}</div>
              <div class="stat-label">已收款</div>
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
              <div class="stat-value">${{ stats.pendingAmount }}</div>
              <div class="stat-label">待收款</div>
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
              <div class="stat-value">{{ stats.overdueCount }}</div>
              <div class="stat-label">逾期账单</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 账单列表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>账单管理</span>
          <div class="header-actions">
            <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 120px; margin-right: 10px;" @change="fetchBillings">
              <el-option label="待支付" value="pending" />
              <el-option label="已支付" value="paid" />
              <el-option label="已逾期" value="overdue" />
            </el-select>
            <el-date-picker
              v-model="filterMonth"
              type="month"
              placeholder="选择月份"
              style="margin-right: 10px;"
              @change="fetchBillings"
            />
            <el-button type="primary" @click="generateBillings">
              <el-icon><DocumentAdd /></el-icon> 生成账单
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="billings" v-loading="loading">
        <el-table-column prop="billing_period" label="账单周期" width="120" />
        <el-table-column prop="tenant_name" label="租户" min-width="150">
          <template #default="{ row }">
            <div class="tenant-info">
              <span>{{ row.tenant_name }}</span>
              <el-tag size="small" type="info">{{ row.tenant_slug }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="费用明细">
          <el-table-column prop="base_amount" label="基础费用" width="100" align="right">
            <template #default="{ row }">
              ${{ row.base_amount }}
            </template>
          </el-table-column>
          <el-table-column prop="usage_amount" label="用量费用" width="100" align="right">
            <template #default="{ row }">
              ${{ row.usage_amount }}
            </template>
          </el-table-column>
          <el-table-column prop="overage_amount" label="超额费用" width="100" align="right">
            <template #default="{ row }">
              <span :class="{ 'text-danger': row.overage_amount > 0 }">
                ${{ row.overage_amount }}
              </span>
            </template>
          </el-table-column>
        </el-table-column>
        <el-table-column prop="total_amount" label="总计" width="100" align="right">
          <template #default="{ row }">
            <span class="total-amount">${{ row.total_amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ getStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="due_date" label="到期日" width="120">
          <template #default="{ row }">
            {{ formatDate(row.due_date) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewBilling(row)">详情</el-button>
            <el-button size="small" type="primary" @click="markAsPaid(row)" v-if="row.status === 'pending'">
              确认支付
            </el-button>
            <el-button size="small" @click="downloadInvoice(row)" v-if="row.status === 'paid'">
              发票
            </el-button>
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
          @size-change="fetchBillings"
          @current-change="fetchBillings"
        />
      </div>
    </el-card>

    <!-- 账单详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="账单详情" width="700px">
      <template v-if="currentBilling">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="账单周期">{{ currentBilling.billing_period }}</el-descriptions-item>
          <el-descriptions-item label="租户">{{ currentBilling.tenant_name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentBilling.status)">{{ getStatusLabel(currentBilling.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="到期日">{{ formatDate(currentBilling.due_date) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">费用明细</el-divider>

        <el-descriptions :column="2" border>
          <el-descriptions-item label="基础费用">${{ currentBilling.base_amount }}</el-descriptions-item>
          <el-descriptions-item label="用量费用">${{ currentBilling.usage_amount }}</el-descriptions-item>
          <el-descriptions-item label="超额费用">${{ currentBilling.overage_amount }}</el-descriptions-item>
          <el-descriptions-item label="折扣">-${{ currentBilling.discount_amount }}</el-descriptions-item>
          <el-descriptions-item label="税费">${{ currentBilling.tax_amount }}</el-descriptions-item>
          <el-descriptions-item label="总计">
            <span class="total-amount">${{ currentBilling.total_amount }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">用量详情</el-divider>

        <el-descriptions :column="3" border v-if="currentBilling.usage_details">
          <el-descriptions-item label="用户数">{{ currentBilling.usage_details.users }}</el-descriptions-item>
          <el-descriptions-item label="资源数">{{ currentBilling.usage_details.resources }}</el-descriptions-item>
          <el-descriptions-item label="存储(GB)">{{ currentBilling.usage_details.storage }}</el-descriptions-item>
          <el-descriptions-item label="API调用">{{ currentBilling.usage_details.api_calls }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left" v-if="currentBilling.status === 'paid'">支付信息</el-divider>

        <el-descriptions :column="2" border v-if="currentBilling.status === 'paid'">
          <el-descriptions-item label="支付方式">{{ currentBilling.payment_method }}</el-descriptions-item>
          <el-descriptions-item label="支付时间">{{ formatDateTime(currentBilling.paid_at) }}</el-descriptions-item>
          <el-descriptions-item label="发票号">{{ currentBilling.invoice_number }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Wallet, Money, Timer, Warning, DocumentAdd } from '@element-plus/icons-vue'

const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const filterStatus = ref('')
const filterMonth = ref('')
const showDetailDialog = ref(false)
const currentBilling = ref<any>(null)

const stats = ref({
  totalRevenue: 45600,
  paidAmount: 42000,
  pendingAmount: 3600,
  overdueCount: 3
})

const billings = ref([
  {
    id: '1',
    billing_period: '2024-03',
    tenant_name: 'Acme Corporation',
    tenant_slug: 'acme',
    base_amount: 299,
    usage_amount: 45,
    overage_amount: 0,
    discount_amount: 0,
    tax_amount: 0,
    total_amount: 344,
    status: 'paid',
    due_date: '2024-03-15',
    paid_at: '2024-03-10',
    payment_method: '信用卡',
    invoice_number: 'INV-2024-001'
  },
  {
    id: '2',
    billing_period: '2024-03',
    tenant_name: 'TechStart Inc',
    tenant_slug: 'techstart',
    base_amount: 99,
    usage_amount: 15,
    overage_amount: 20,
    discount_amount: 0,
    tax_amount: 0,
    total_amount: 134,
    status: 'pending',
    due_date: '2024-03-15',
    usage_details: { users: 12, resources: 120, storage: 55, api_calls: 12000 }
  },
  {
    id: '3',
    billing_period: '2024-03',
    tenant_name: 'Demo Company',
    tenant_slug: 'demo',
    base_amount: 0,
    usage_amount: 0,
    overage_amount: 0,
    discount_amount: 0,
    tax_amount: 0,
    total_amount: 0,
    status: 'paid',
    due_date: '2024-03-15',
    paid_at: '2024-03-01',
    payment_method: '免费套餐',
    invoice_number: ''
  },
  {
    id: '4',
    billing_period: '2024-02',
    tenant_name: 'Global Solutions',
    tenant_slug: 'global',
    base_amount: 0,
    usage_amount: 0,
    overage_amount: 0,
    discount_amount: 0,
    tax_amount: 0,
    total_amount: 5000,
    status: 'overdue',
    due_date: '2024-02-15',
    usage_details: { users: 150, resources: 1200, storage: 500, api_calls: 50000 }
  }
])

const fetchBillings = async () => {
  loading.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    total.value = billings.value.length
  } catch (error) {
    console.error('获取账单列表失败:', error)
  } finally {
    loading.value = false
  }
}

const generateBillings = async () => {
  try {
    await ElMessageBox.confirm('确定生成本月账单？', '提示', { type: 'warning' })
    ElMessage.success('账单生成中...')
  } catch {}
}

const viewBilling = (billing: any) => {
  currentBilling.value = billing
  showDetailDialog.value = true
}

const markAsPaid = async (billing: any) => {
  try {
    await ElMessageBox.confirm('确认该账单已支付？', '提示', { type: 'info' })
    billing.status = 'paid'
    billing.paid_at = new Date().toISOString()
    billing.payment_method = '管理员确认'
    ElMessage.success('已确认支付')
  } catch {}
}

const downloadInvoice = (billing: any) => {
  ElMessage.success(`正在下载发票 ${billing.invoice_number}`)
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = { paid: 'success', pending: 'warning', overdue: 'danger' }
  return types[status] || 'info'
}

const getStatusLabel = (status: string) => {
  const labels: Record<string, string> = { paid: '已支付', pending: '待支付', overdue: '已逾期' }
  return labels[status] || status
}

const formatDate = (date: string) => date?.substring(0, 10) || '-'
const formatDateTime = (date: string) => date?.replace('T', ' ').substring(0, 19) || '-'

onMounted(() => {
  fetchBillings()
})
</script>

<style scoped>
.tenant-billing-page {
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

.tenant-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.total-amount {
  font-weight: bold;
  font-size: 16px;
  color: #409eff;
}

.text-danger {
  color: #f56c6c;
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
