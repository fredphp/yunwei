<template>
  <div class="cost-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
              <el-icon size="28"><Wallet /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">${{ stats.currentMonth.toFixed(0) }}</div>
              <div class="stat-label">本月成本</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);">
              <el-icon size="28"><TrendCharts /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">${{ stats.forecast.toFixed(0) }}</div>
              <div class="stat-label">预测月终</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
              <el-icon size="28"><PieChart /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">${{ stats.waste.toFixed(0) }}</div>
              <div class="stat-label">潜在节省</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">
              <el-icon size="28"><DataLine /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.idleResources }}</div>
              <div class="stat-label">闲置资源</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <!-- 成本趋势图表 -->
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>成本趋势</span>
              <el-radio-group v-model="chartType" size="small">
                <el-radio-button label="daily">日视图</el-radio-button>
                <el-radio-button label="monthly">月视图</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container" ref="chartRef"></div>
        </el-card>

        <!-- 成本分布 -->
        <el-card class="mt-4">
          <template #header>
            <span>成本分布</span>
          </template>
          <el-row :gutter="20">
            <el-col :span="12">
              <div class="chart-container" ref="pieChartRef"></div>
            </el-col>
            <el-col :span="12">
              <el-table :data="costBreakdown" size="small">
                <el-table-column prop="category" label="类别" />
                <el-table-column prop="cost" label="成本">
                  <template #default="{ row }">
                    ${{ row.cost.toFixed(2) }}
                  </template>
                </el-table-column>
                <el-table-column label="占比" width="120">
                  <template #default="{ row }">
                    <el-progress :percentage="row.percent" :stroke-width="10" />
                  </template>
                </el-table-column>
              </el-table>
            </el-col>
          </el-row>
        </el-card>
      </el-col>

      <!-- 右侧面板 -->
      <el-col :span="8">
        <!-- 预算管理 -->
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>预算管理</span>
              <el-button size="small" type="primary" @click="showBudgetDialog = true">设置</el-button>
            </div>
          </template>
          <div v-for="budget in budgets" :key="budget.id" class="budget-item">
            <div class="budget-header">
              <span>{{ budget.name }}</span>
              <el-tag :type="budget.percent > 80 ? 'danger' : budget.percent > 60 ? 'warning' : 'success'">
                {{ budget.percent }}%
              </el-tag>
            </div>
            <el-progress :percentage="budget.percent" :color="budget.percent > 80 ? '#f56c6c' : '#67c23a'" />
            <div class="budget-detail">
              ${{ budget.spent.toFixed(0) }} / ${{ budget.total.toFixed(0) }}
            </div>
          </div>
        </el-card>

        <!-- 浪费检测 -->
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>浪费检测</span>
              <el-button size="small" type="danger">扫描</el-button>
            </div>
          </template>
          <div v-for="waste in wasteList" :key="waste.id" class="waste-item">
            <div class="waste-info">
              <el-icon :size="20" :color="getWasteColor(waste.severity)">
                <WarningFilled v-if="waste.severity === 'high'" />
                <InfoFilled v-else />
              </el-icon>
              <div class="waste-detail">
                <div class="waste-title">{{ waste.resourceName }}</div>
                <div class="waste-reason">{{ waste.reason }}</div>
              </div>
            </div>
            <div class="waste-savings">
              <span class="savings-amount">${{ waste.savings.toFixed(0) }}/月</span>
              <el-button size="small" type="primary" text @click="optimizeWaste(waste)">优化</el-button>
            </div>
          </div>
        </el-card>

        <!-- 闲置资源 -->
        <el-card>
          <template #header>
            <div class="card-header">
              <span>闲置资源</span>
              <el-button size="small" text>查看全部</el-button>
            </div>
          </template>
          <div v-for="idle in idleResources" :key="idle.id" class="idle-item">
            <div class="idle-info">
              <div class="idle-name">{{ idle.name }}</div>
              <div class="idle-stats">
                CPU: {{ idle.cpuUsage }}% | 内存: {{ idle.memUsage }}%
              </div>
            </div>
            <div class="idle-actions">
              <el-tag size="small" type="warning">{{ idle.idleDays }}天</el-tag>
              <el-dropdown trigger="click">
                <el-button size="small">操作</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="stopIdle(idle)">停止</el-dropdown-item>
                    <el-dropdown-item @click="resizeIdle(idle)">缩容</el-dropdown-item>
                    <el-dropdown-item @click="ignoreIdle(idle)">忽略</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 云账户管理 -->
    <el-card class="mt-4">
      <template #header>
        <div class="card-header">
          <span>云账户管理</span>
          <el-button type="primary" @click="showAccountDialog = true">
            <el-icon><Plus /></el-icon> 添加账户
          </el-button>
        </div>
      </template>
      <el-table :data="cloudAccounts">
        <el-table-column prop="name" label="账户名称" />
        <el-table-column prop="provider" label="云服务商" width="120">
          <template #default="{ row }">
            <el-tag>{{ row.provider }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="monthCost" label="本月成本" width="120">
          <template #default="{ row }">
            ${{ row.monthCost.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="budget" label="预算" width="120">
          <template #default="{ row }">
            ${{ row.budget?.toFixed(0) || '未设置' }}
          </template>
        </el-table-column>
        <el-table-column prop="lastSync" label="最后同步" width="180" />
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button size="small" @click="syncAccount(row)">同步</el-button>
            <el-button size="small" @click="editAccount(row)">编辑</el-button>
            <el-button size="small" type="danger" text @click="deleteAccount(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 对话框 -->
    <el-dialog v-model="showBudgetDialog" title="设置预算" width="500px">
      <el-form :model="budgetForm" label-width="100px">
        <el-form-item label="预算名称">
          <el-input v-model="budgetForm.name" />
        </el-form-item>
        <el-form-item label="预算金额">
          <el-input-number v-model="budgetForm.amount" :min="0" :step="100" />
        </el-form-item>
        <el-form-item label="周期">
          <el-select v-model="budgetForm.period" style="width: 100%;">
            <el-option label="月度" value="monthly" />
            <el-option label="季度" value="quarterly" />
            <el-option label="年度" value="yearly" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBudgetDialog = false">取消</el-button>
        <el-button type="primary" @click="saveBudget">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showAccountDialog" title="添加云账户" width="500px">
      <el-form :model="accountForm" label-width="100px">
        <el-form-item label="账户名称">
          <el-input v-model="accountForm.name" />
        </el-form-item>
        <el-form-item label="云服务商">
          <el-select v-model="accountForm.provider" style="width: 100%;">
            <el-option label="AWS" value="aws" />
            <el-option label="GCP" value="gcp" />
            <el-option label="阿里云" value="aliyun" />
            <el-option label="腾讯云" value="tencent" />
          </el-select>
        </el-form-item>
        <el-form-item label="Access Key">
          <el-input v-model="accountForm.accessKey" />
        </el-form-item>
        <el-form-item label="Secret Key">
          <el-input v-model="accountForm.secretKey" type="password" show-password />
        </el-form-item>
        <el-form-item label="预算">
          <el-input-number v-model="accountForm.budget" :min="0" :step="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAccountDialog = false">取消</el-button>
        <el-button type="primary" @click="saveAccount">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Wallet, TrendCharts, PieChart, DataLine, Plus, WarningFilled, InfoFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'
import * as echarts from 'echarts'

const chartType = ref('daily')
const chartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()
const showBudgetDialog = ref(false)
const showAccountDialog = ref(false)

const stats = ref({
  currentMonth: 12500,
  forecast: 15200,
  waste: 2800,
  idleResources: 12
})

const costBreakdown = ref([
  { category: '计算资源', cost: 5200, percent: 42 },
  { category: '存储', cost: 3100, percent: 25 },
  { category: '数据库', cost: 2500, percent: 20 },
  { category: '网络', cost: 1200, percent: 10 },
  { category: '其他', cost: 500, percent: 3 }
])

const budgets = ref([
  { id: 1, name: 'AWS 生产环境', spent: 8500, total: 10000, percent: 85 },
  { id: 2, name: 'GCP 开发环境', spent: 3200, total: 5000, percent: 64 }
])

const wasteList = ref([
  { id: 1, resourceName: 'EC2-Web-01', reason: 'CPU使用率长期低于5%', severity: 'high', savings: 156 },
  { id: 2, resourceName: 'RDS-Standby', reason: '备用实例未启用', severity: 'medium', savings: 89 },
  { id: 3, resourceName: 'EBS-Volume-03', reason: '未挂载的EBS卷', severity: 'low', savings: 23 }
])

const idleResources = ref([
  { id: 1, name: 'EC2-Test-02', cpuUsage: 3, memUsage: 15, idleDays: 14 },
  { id: 2, name: 'Pod-nginx-5', cpuUsage: 5, memUsage: 20, idleDays: 7 },
  { id: 3, name: 'VM-Dev-03', cpuUsage: 2, memUsage: 10, idleDays: 21 }
])

const cloudAccounts = ref([
  { id: 1, name: 'AWS Production', provider: 'aws', monthCost: 8500, budget: 10000, lastSync: '2024-02-23 10:30' },
  { id: 2, name: 'GCP Development', provider: 'gcp', monthCost: 3200, budget: 5000, lastSync: '2024-02-23 09:15' }
])

const budgetForm = ref({
  name: '',
  amount: 10000,
  period: 'monthly'
})

const accountForm = ref({
  name: '',
  provider: 'aws',
  accessKey: '',
  secretKey: '',
  budget: 10000
})

const initCharts = () => {
  // 成本趋势图
  if (chartRef.value) {
    const chart = echarts.init(chartRef.value)
    const option = {
      tooltip: { trigger: 'axis' },
      legend: { data: ['计算', '存储', '数据库', '网络'] },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: ['1月', '2月', '3月', '4月', '5月', '6月'] },
      yAxis: { type: 'value' },
      series: [
        { name: '计算', type: 'line', stack: 'Total', data: [3200, 3500, 3800, 4200, 4500, 5200] },
        { name: '存储', type: 'line', stack: 'Total', data: [1800, 2000, 2200, 2500, 2800, 3100] },
        { name: '数据库', type: 'line', stack: 'Total', data: [1500, 1600, 1800, 2000, 2200, 2500] },
        { name: '网络', type: 'line', stack: 'Total', data: [800, 900, 1000, 1100, 1150, 1200] }
      ]
    }
    chart.setOption(option)
  }

  // 饼图
  if (pieChartRef.value) {
    const chart = echarts.init(pieChartRef.value)
    const option = {
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
        label: { show: false },
        data: [
          { value: 5200, name: '计算资源', itemStyle: { color: '#5470c6' } },
          { value: 3100, name: '存储', itemStyle: { color: '#91cc75' } },
          { value: 2500, name: '数据库', itemStyle: { color: '#fac858' } },
          { value: 1200, name: '网络', itemStyle: { color: '#ee6666' } },
          { value: 500, name: '其他', itemStyle: { color: '#73c0de' } }
        ]
      }]
    }
    chart.setOption(option)
  }
}

const fetchCostData = async () => {
  try {
    const res = await request.get('/api/v1/cost/overview')
    // 更新数据
  } catch (error) {
    console.error('获取成本数据失败', error)
  }
}

const getWasteColor = (severity: string) => {
  return severity === 'high' ? '#f56c6c' : '#e6a23c'
}

const optimizeWaste = (waste: any) => {
  ElMessage.success(`已提交优化建议: ${waste.resourceName}`)
}

const stopIdle = (idle: any) => {
  ElMessage.success(`已停止闲置资源: ${idle.name}`)
}

const resizeIdle = (idle: any) => {
  ElMessage.success(`已提交缩容请求: ${idle.name}`)
}

const ignoreIdle = (idle: any) => {
  ElMessage.info(`已忽略: ${idle.name}`)
}

const saveBudget = async () => {
  ElMessage.success('预算设置成功')
  showBudgetDialog.value = false
}

const saveAccount = async () => {
  try {
    await request.post('/api/v1/cost/accounts', accountForm.value)
    ElMessage.success('账户添加成功')
    showAccountDialog.value = false
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const syncAccount = async (account: any) => {
  try {
    await request.post(`/api/v1/cost/accounts/${account.id}/sync`)
    ElMessage.success('同步成功')
  } catch (error) {
    ElMessage.error('同步失败')
  }
}

const editAccount = (account: any) => {
  accountForm.value = { ...account }
  showAccountDialog.value = true
}

const deleteAccount = async (account: any) => {
  ElMessage.success('账户已删除')
}

onMounted(() => {
  initCharts()
  fetchCostData()
})
</script>

<style scoped>
.cost-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-card-wrapper {
  background: transparent;
}

.stat-card {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
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

.mt-4 {
  margin-top: 16px;
}

.chart-container {
  height: 300px;
}

.budget-item {
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
}

.budget-item:last-child {
  border-bottom: none;
}

.budget-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.budget-detail {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}

.waste-item, .idle-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
}

.waste-item:last-child, .idle-item:last-child {
  border-bottom: none;
}

.waste-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.waste-detail {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.waste-title {
  font-weight: 500;
}

.waste-reason {
  font-size: 12px;
  color: #909399;
}

.waste-savings {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.savings-amount {
  font-weight: 500;
  color: #67c23a;
}

.idle-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.idle-name {
  font-weight: 500;
}

.idle-stats {
  font-size: 12px;
  color: #909399;
}

.idle-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
