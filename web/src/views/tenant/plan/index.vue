<template>
  <div class="tenant-plan-page">
    <!-- 套餐卡片 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6" v-for="plan in plans" :key="plan.key">
        <el-card class="plan-card" :class="{ 'plan-popular': plan.popular }">
          <div class="plan-header">
            <h3>{{ plan.name }}</h3>
            <div class="plan-price">
              <span class="currency">$</span>
              <span class="amount">{{ plan.price }}</span>
              <span class="period">/月</span>
            </div>
            <p class="plan-desc">{{ plan.description }}</p>
          </div>
          <el-divider />
          <div class="plan-features">
            <div class="feature-item" v-for="(feature, idx) in plan.features" :key="idx">
              <el-icon class="feature-icon"><Check /></el-icon>
              <span>{{ feature }}</span>
            </div>
          </div>
          <div class="plan-footer">
            <el-button 
              :type="plan.popular ? 'primary' : 'default'" 
              @click="showEditPlan(plan)"
              style="width: 100%;"
            >
              编辑套餐
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 套餐对比表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>套餐对比</span>
        </div>
      </template>
      <el-table :data="comparisonData" border>
        <el-table-column prop="feature" label="功能特性" width="200" fixed />
        <el-table-column prop="free" label="Free" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.free === true" class="check-icon"><Check /></el-icon>
            <el-icon v-else-if="row.free === false" class="close-icon"><Close /></el-icon>
            <span v-else>{{ row.free }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="starter" label="Starter" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.starter === true" class="check-icon"><Check /></el-icon>
            <el-icon v-else-if="row.starter === false" class="close-icon"><Close /></el-icon>
            <span v-else>{{ row.starter }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="pro" label="Pro" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.pro === true" class="check-icon"><Check /></el-icon>
            <el-icon v-else-if="row.pro === false" class="close-icon"><Close /></el-icon>
            <span v-else>{{ row.pro }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="enterprise" label="Enterprise" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.enterprise === true" class="check-icon"><Check /></el-icon>
            <el-icon v-else-if="row.enterprise === false" class="close-icon"><Close /></el-icon>
            <span v-else>{{ row.enterprise }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑套餐对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑套餐" width="600px">
      <el-form :model="editForm" label-width="120px" v-if="currentPlan">
        <el-form-item label="套餐名称">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="月费价格">
          <el-input-number v-model="editForm.price" :min="0" :precision="2" />
        </el-form-item>
        <el-form-item label="年费价格">
          <el-input-number v-model="editForm.yearlyPrice" :min="0" :precision="2" />
          <span class="hint">（年付优惠价）</span>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-divider content-position="left">配额设置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="最大用户数">
              <el-input-number v-model="editForm.maxUsers" :min="-1" />
              <span class="hint">-1表示无限制</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大管理员">
              <el-input-number v-model="editForm.maxAdmins" :min="-1" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大资源数">
              <el-input-number v-model="editForm.maxResources" :min="-1" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大服务器">
              <el-input-number v-model="editForm.maxServers" :min="-1" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="存储空间(GB)">
              <el-input-number v-model="editForm.maxStorageGB" :min="-1" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="API调用/日">
              <el-input-number v-model="editForm.maxAPICalls" :min="-1" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="savePlan">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, Close } from '@element-plus/icons-vue'

const showEditDialog = ref(false)
const currentPlan = ref<any>(null)

const plans = ref([
  {
    key: 'free',
    name: 'Free',
    price: 0,
    yearlyPrice: 0,
    description: '适合个人学习使用',
    popular: false,
    features: [
      '最多 3 个用户',
      '最多 20 个资源',
      '10GB 存储空间',
      '7 天指标保留',
      '社区支持'
    ]
  },
  {
    key: 'starter',
    name: 'Starter',
    price: 99,
    yearlyPrice: 990,
    description: '适合小型团队',
    popular: false,
    features: [
      '最多 10 个用户',
      '最多 100 个资源',
      '50GB 存储空间',
      '30 天指标保留',
      '邮件支持'
    ]
  },
  {
    key: 'pro',
    name: 'Pro',
    price: 299,
    yearlyPrice: 2990,
    description: '适合中型企业',
    popular: true,
    features: [
      '最多 50 个用户',
      '最多 500 个资源',
      '200GB 存储空间',
      '90 天指标保留',
      '优先技术支持'
    ]
  },
  {
    key: 'enterprise',
    name: 'Enterprise',
    price: 0,
    yearlyPrice: 0,
    description: '适合大型企业定制',
    popular: false,
    features: [
      '无限用户',
      '无限资源',
      '无限存储空间',
      '365 天指标保留',
      '专属客户经理'
    ]
  }
])

const comparisonData = ref([
  { feature: '用户数量', free: '3', starter: '10', pro: '50', enterprise: '无限' },
  { feature: '管理员数量', free: '1', starter: '3', pro: '10', enterprise: '无限' },
  { feature: '资源数量', free: '20', starter: '100', pro: '500', enterprise: '无限' },
  { feature: '服务器数量', free: '10', starter: '50', pro: '200', enterprise: '无限' },
  { feature: '数据库数量', free: '5', starter: '20', pro: '100', enterprise: '无限' },
  { feature: '存储空间', free: '10GB', starter: '50GB', pro: '200GB', enterprise: '无限' },
  { feature: 'API调用/日', free: '1,000', starter: '10,000', pro: '100,000', enterprise: '无限' },
  { feature: '监控规则', free: '20', starter: '100', pro: '500', enterprise: '无限' },
  { feature: '告警规则', free: '10', starter: '50', pro: '200', enterprise: '无限' },
  { feature: '指标保留天数', free: '7', starter: '30', pro: '90', enterprise: '365' },
  { feature: '云账户数量', free: '1', starter: '3', pro: '10', enterprise: '无限' },
  { feature: 'Webhooks', free: '2', starter: '10', pro: '50', enterprise: '无限' },
  { feature: '高可用支持', free: false, starter: false, pro: true, enterprise: true },
  { feature: '灰度发布', free: false, starter: true, pro: true, enterprise: true },
  { feature: 'AI分析', free: false, starter: false, pro: true, enterprise: true },
  { feature: '专属客户经理', free: false, starter: false, pro: false, enterprise: true },
])

const editForm = ref({
  name: '',
  price: 0,
  yearlyPrice: 0,
  description: '',
  maxUsers: 0,
  maxAdmins: 0,
  maxResources: 0,
  maxServers: 0,
  maxStorageGB: 0,
  maxAPICalls: 0
})

const showEditPlan = (plan: any) => {
  currentPlan.value = plan
  editForm.value = {
    name: plan.name,
    price: plan.price,
    yearlyPrice: plan.yearlyPrice,
    description: plan.description,
    maxUsers: plan.key === 'free' ? 3 : plan.key === 'starter' ? 10 : plan.key === 'pro' ? 50 : -1,
    maxAdmins: plan.key === 'free' ? 1 : plan.key === 'starter' ? 3 : plan.key === 'pro' ? 10 : -1,
    maxResources: plan.key === 'free' ? 20 : plan.key === 'starter' ? 100 : plan.key === 'pro' ? 500 : -1,
    maxServers: plan.key === 'free' ? 10 : plan.key === 'starter' ? 50 : plan.key === 'pro' ? 200 : -1,
    maxStorageGB: plan.key === 'free' ? 10 : plan.key === 'starter' ? 50 : plan.key === 'pro' ? 200 : -1,
    maxAPICalls: plan.key === 'free' ? 1000 : plan.key === 'starter' ? 10000 : plan.key === 'pro' ? 100000 : -1
  }
  showEditDialog.value = true
}

const savePlan = () => {
  ElMessage.success('套餐配置已保存')
  showEditDialog.value = false
}
</script>

<style scoped>
.tenant-plan-page {
  padding: 20px;
}

.plan-card {
  text-align: center;
  transition: all 0.3s;
}

.plan-card:hover {
  transform: translateY(-5px);
}

.plan-popular {
  border: 2px solid #409eff;
}

.plan-popular::before {
  content: '推荐';
  position: absolute;
  top: 10px;
  right: -30px;
  background: #409eff;
  color: white;
  padding: 2px 30px;
  transform: rotate(45deg);
  font-size: 12px;
}

.plan-header h3 {
  margin: 0 0 10px;
  font-size: 24px;
}

.plan-price {
  margin: 15px 0;
}

.plan-price .currency {
  font-size: 18px;
  vertical-align: top;
}

.plan-price .amount {
  font-size: 48px;
  font-weight: bold;
  color: #303133;
}

.plan-price .period {
  font-size: 14px;
  color: #909399;
}

.plan-desc {
  color: #909399;
  font-size: 14px;
}

.plan-features {
  text-align: left;
  padding: 10px 0;
}

.feature-item {
  display: flex;
  align-items: center;
  padding: 8px 0;
}

.feature-icon {
  color: #67c23a;
  margin-right: 10px;
}

.plan-footer {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.check-icon {
  color: #67c23a;
  font-size: 18px;
}

.close-icon {
  color: #f56c6c;
  font-size: 18px;
}

.hint {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
}

.mb-4 {
  margin-bottom: 16px;
}
</style>
