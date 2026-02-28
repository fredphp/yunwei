<template>
  <div class="cert-page">
    <el-row :gutter="20">
      <el-col :span="16">
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>SSL 证书管理</span>
              <div>
                <el-button type="success" @click="checkAllCerts">
                  检查所有证书
                </el-button>
                <el-button type="primary" @click="showAddDialog = true">
                  添加证书
                </el-button>
              </div>
            </div>
          </template>
          
          <el-table :data="certificates" v-loading="loading">
            <el-table-column prop="domain" label="域名" />
            <el-table-column prop="provider" label="提供商" width="100" />
            <el-table-column label="有效期" width="200">
              <template #default="{ row }">
                <div>
                  {{ formatDate(row.notBefore) }} 至
                  <br />
                  {{ formatDate(row.notAfter) }}
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="daysLeft" label="剩余天数" width="100">
              <template #default="{ row }">
                <el-tag :type="getDaysLeftType(row.daysLeft)">
                  {{ row.daysLeft }} 天
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="autoRenew" label="自动续期" width="100">
              <template #default="{ row }">
                <el-switch v-model="row.autoRenew" @change="updateCert(row)" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button size="small" @click="checkCert(row)">检查</el-button>
                <el-button size="small" type="primary" @click="renewCert(row)" 
                  :disabled="row.status !== 'expiring' && row.status !== 'expired'">
                  续期
                </el-button>
                <el-button size="small" type="danger" @click="deleteCert(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card class="mb-4">
          <template #header>
            <span>证书概览</span>
          </template>
          
          <el-statistic title="证书总数" :value="certStats.total" class="mb-4" />
          <el-statistic title="即将过期" :value="certStats.expiring" class="mb-4">
            <template #suffix>
              <span class="text-warning">个</span>
            </template>
          </el-statistic>
          <el-statistic title="已过期" :value="certStats.expired" class="mb-4">
            <template #suffix>
              <span class="text-danger">个</span>
            </template>
          </el-statistic>
          <el-statistic title="自动续期开启" :value="certStats.autoRenew" />
        </el-card>

        <el-card>
          <template #header>
            <span>申请新证书</span>
          </template>
          <el-form :model="requestForm" label-width="100px">
            <el-form-item label="域名">
              <el-input v-model="requestForm.domain" placeholder="example.com" />
            </el-form-item>
            <el-form-item label="邮箱">
              <el-input v-model="requestForm.email" placeholder="admin@example.com" />
            </el-form-item>
            <el-form-item label="DNS 提供商">
              <el-select v-model="requestForm.dnsProvider">
                <el-option label="HTTP 验证" value="http" />
                <el-option label="Cloudflare" value="cloudflare" />
                <el-option label="阿里云" value="aliyun" />
                <el-option label="腾讯云" value="tencent" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="requestCert" :loading="requesting">
                申请证书
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <!-- 续期历史 -->
    <el-card class="mb-4">
      <template #header>
        <span>续期历史</span>
      </template>
      <el-table :data="renewalHistory" v-loading="historyLoading">
        <el-table-column prop="cert.domain" label="域名" />
        <el-table-column label="旧证书序列" width="200">
          <template #default="{ row }">
            {{ row.oldSerialNumber?.substring(0, 16) }}...
          </template>
        </el-table-column>
        <el-table-column label="新证书序列" width="200">
          <template #default="{ row }">
            {{ row.newSerialNumber?.substring(0, 16) }}...
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="method" label="方式" width="100" />
        <el-table-column prop="createdAt" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加证书对话框 -->
    <el-dialog v-model="showAddDialog" title="添加证书" width="600px">
      <el-form :model="certForm" label-width="120px">
        <el-form-item label="名称">
          <el-input v-model="certForm.name" placeholder="主站证书" />
        </el-form-item>
        <el-form-item label="域名">
          <el-input v-model="certForm.domain" placeholder="example.com" />
        </el-form-item>
        <el-form-item label="证书文件路径">
          <el-input v-model="certForm.certPath" placeholder="/etc/ssl/certs/example.com.crt" />
        </el-form-item>
        <el-form-item label="私钥文件路径">
          <el-input v-model="certForm.keyPath" placeholder="/etc/ssl/private/example.com.key" />
        </el-form-item>
        <el-form-item label="证书链路径">
          <el-input v-model="certForm.chainPath" placeholder="/etc/ssl/certs/chain.crt" />
        </el-form-item>
        <el-form-item label="提供商">
          <el-select v-model="certForm.provider">
            <el-option label="Let's Encrypt" value="letsencrypt" />
            <el-option label="ZeroSSL" value="zerossl" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="提前续期天数">
          <el-input-number v-model="certForm.renewBefore" :min="7" :max="90" />
        </el-form-item>
        <el-form-item label="自动续期">
          <el-switch v-model="certForm.autoRenew" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="addCert">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'

const loading = ref(false)
const historyLoading = ref(false)
const requesting = ref(false)
const certificates = ref([])
const renewalHistory = ref([])
const showAddDialog = ref(false)

const certForm = ref({
  name: '',
  domain: '',
  certPath: '',
  keyPath: '',
  chainPath: '',
  provider: 'letsencrypt',
  renewBefore: 30,
  autoRenew: true
})

const requestForm = ref({
  domain: '',
  email: '',
  dnsProvider: 'http'
})

const certStats = computed(() => {
  const certs = certificates.value
  return {
    total: certs.length,
    expiring: certs.filter((c: any) => c.status === 'expiring').length,
    expired: certs.filter((c: any) => c.status === 'expired').length,
    autoRenew: certs.filter((c: any) => c.autoRenew).length
  }
})

const fetchCertificates = async () => {
  loading.value = true
  try {
    const res = await request.get('/certificates')
    certificates.value = res.data || []
  } catch (error) {
    ElMessage.error('获取证书列表失败')
  } finally {
    loading.value = false
  }
}

const fetchRenewalHistory = async () => {
  historyLoading.value = true
  try {
    const res = await request.get('/certificates/history')
    renewalHistory.value = res.data || []
  } catch (error) {
    console.error('获取续期历史失败', error)
  } finally {
    historyLoading.value = false
  }
}

const addCert = async () => {
  try {
    await request.post('/certificates', certForm.value)
    ElMessage.success('添加成功')
    showAddDialog.value = false
    fetchCertificates()
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const checkCert = async (cert: any) => {
  try {
    const res = await request.post(`/certificates/${cert.id}/check`)
    ElMessage.success('检查完成')
    // 更新列表中的证书信息
    fetchCertificates()
  } catch (error) {
    ElMessage.error('检查失败')
  }
}

const checkAllCerts = async () => {
  try {
    await request.post('/certificates/check-all')
    ElMessage.success('所有证书检查完成')
    fetchCertificates()
  } catch (error) {
    ElMessage.error('检查失败')
  }
}

const renewCert = async (cert: any) => {
  try {
    await ElMessageBox.confirm('确定要续期该证书吗？', '确认')
    const res = await request.post(`/certificates/${cert.id}/renew`)
    ElMessage.success('续期请求已提交')
    fetchRenewalHistory()
    fetchCertificates()
  } catch (error) {
    // 用户取消或其他错误
  }
}

const deleteCert = async (cert: any) => {
  try {
    await ElMessageBox.confirm('确定要删除该证书吗？', '确认', { type: 'warning' })
    await request.delete(`/certificates/${cert.id}`)
    ElMessage.success('删除成功')
    fetchCertificates()
  } catch (error) {
    // 用户取消
  }
}

const requestCert = async () => {
  requesting.value = true
  try {
    await request.post('/certificates/request', requestForm.value)
    ElMessage.success('证书申请已提交，请稍后刷新查看结果')
    requestForm.value = { domain: '', email: '', dnsProvider: 'http' }
  } catch (error) {
    ElMessage.error('申请失败')
  } finally {
    requesting.value = false
  }
}

const updateCert = async (cert: any) => {
  try {
    await request.put(`/certificates/${cert.id}`, cert)
    ElMessage.success('更新成功')
  } catch (error) {
    ElMessage.error('更新失败')
  }
}

const getDaysLeftType = (days: number) => {
  if (days <= 0) return 'danger'
  if (days <= 30) return 'warning'
  return 'success'
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    valid: 'success',
    expiring: 'warning',
    expired: 'danger',
    renewing: 'primary',
    failed: 'danger'
  }
  return types[status] || 'info'
}

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString()
}

onMounted(() => {
  fetchCertificates()
  fetchRenewalHistory()
})
</script>

<style scoped>
.cert-page {
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

.text-warning {
  color: var(--el-color-warning);
}

.text-danger {
  color: var(--el-color-danger);
}
</style>
