<template>
  <div class="cdn-page">
    <el-row :gutter="20">
      <el-col :span="16">
        <el-card class="mb-4">
          <template #header>
            <div class="card-header">
              <span>CDN 域名管理</span>
              <el-button type="primary" @click="showAddDialog = true">
                添加域名
              </el-button>
            </div>
          </template>
          
          <el-table :data="domains" v-loading="loading">
            <el-table-column prop="domain" label="域名" />
            <el-table-column prop="provider" label="提供商" width="100" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'danger'">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="hitRate" label="命中率" width="100">
              <template #default="{ row }">
                <el-progress :percentage="row.hitRate" :stroke-width="8" />
              </template>
            </el-table-column>
            <el-table-column prop="bandwidth" label="带宽" width="100">
              <template #default="{ row }">
                {{ row.bandwidth }} Mbps
              </template>
            </el-table-column>
            <el-table-column prop="monthlyCost" label="月成本" width="100">
              <template #default="{ row }">
                ${{ row.monthlyCost?.toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="280">
              <template #default="{ row }">
                <el-button-group>
                  <el-button size="small" @click="viewDetail(row)">详情</el-button>
                  <el-button size="small" type="primary" @click="optimizeCDN(row)">优化</el-button>
                  <el-button size="small" type="warning" @click="showPurgeDialog(row)">刷新</el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card class="mb-4">
          <template #header>
            <span>CDN 概览</span>
          </template>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-statistic title="总带宽" :value="cdnStats.totalBandwidth" suffix="Mbps" />
            </el-col>
            <el-col :span="12">
              <el-statistic title="总流量" :value="cdnStats.totalTraffic" suffix="GB" />
            </el-col>
          </el-row>
          <el-row :gutter="20" class="mt-4">
            <el-col :span="12">
              <el-statistic title="总请求数" :value="cdnStats.totalRequests" />
            </el-col>
            <el-col :span="12">
              <el-statistic title="月成本" :value="cdnStats.totalCost" prefix="$" />
            </el-col>
          </el-row>
        </el-card>

        <el-card>
          <template #header>
            <span>快速操作</span>
          </template>
          <el-space direction="vertical" style="width: 100%">
            <el-button type="primary" style="width: 100%" @click="costOptimize">
              成本优化分析
            </el-button>
            <el-button type="success" style="width: 100%" @click="preheatHotContent">
              预热热点内容
            </el-button>
          </el-space>
        </el-card>
      </el-col>
    </el-row>

    <!-- 缓存规则 -->
    <el-card class="mb-4" v-if="currentDomain">
      <template #header>
        <div class="card-header">
          <span>缓存规则 - {{ currentDomain.domain }}</span>
          <el-button type="primary" size="small" @click="showRuleDialog = true">
            添加规则
          </el-button>
        </div>
      </template>
      
      <el-table :data="cacheRules" v-loading="rulesLoading">
        <el-table-column prop="name" label="规则名称" />
        <el-table-column prop="pathPattern" label="路径匹配" />
        <el-table-column prop="ttl" label="缓存时间" width="120">
          <template #default="{ row }">
            {{ formatTTL(row.ttl) }}
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" />
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button size="small" @click="editRule(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteRule(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 节点状态 -->
    <el-card class="mb-4" v-if="currentDomain">
      <template #header>
        <span>节点状态</span>
      </template>
      <el-table :data="cdnNodes" v-loading="nodesLoading">
        <el-table-column prop="name" label="节点名称" />
        <el-table-column prop="region" label="区域" />
        <el-table-column prop="isp" label="运营商" />
        <el-table-column prop="latency" label="延迟" width="100">
          <template #default="{ row }">
            {{ row.latency }} ms
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'danger'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 刷新缓存对话框 -->
    <el-dialog v-model="showPurgeCacheDialog" title="刷新缓存" width="500px">
      <el-form label-width="100px">
        <el-form-item label="URL 列表">
          <el-input type="textarea" :rows="5" v-model="purgeURLs" 
            placeholder="每行一个URL，支持通配符" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPurgeCacheDialog = false">取消</el-button>
        <el-button type="primary" @click="purgeCache">刷新</el-button>
      </template>
    </el-dialog>

    <!-- 添加域名对话框 -->
    <el-dialog v-model="showAddDialog" title="添加 CDN 域名" width="600px">
      <el-form :model="domainForm" label-width="120px">
        <el-form-item label="名称">
          <el-input v-model="domainForm.name" placeholder="主站CDN" />
        </el-form-item>
        <el-form-item label="域名">
          <el-input v-model="domainForm.domain" placeholder="cdn.example.com" />
        </el-form-item>
        <el-form-item label="提供商">
          <el-select v-model="domainForm.provider">
            <el-option label="阿里云" value="aliyun" />
            <el-option label="腾讯云" value="tencent" />
            <el-option label="AWS" value="aws" />
            <el-option label="Cloudflare" value="cloudflare" />
            <el-option label="七牛云" value="qiniu" />
          </el-select>
        </el-form-item>
        <el-form-item label="源站类型">
          <el-select v-model="domainForm.originType">
            <el-option label="域名" value="domain" />
            <el-option label="IP" value="ip" />
            <el-option label="OSS" value="oss" />
          </el-select>
        </el-form-item>
        <el-form-item label="源站地址">
          <el-input v-model="domainForm.originHost" placeholder="origin.example.com" />
        </el-form-item>
        <el-form-item label="源站端口">
          <el-input-number v-model="domainForm.originPort" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="启用 HTTPS">
          <el-switch v-model="domainForm.enableHttps" />
        </el-form-item>
        <el-form-item label="默认缓存时间">
          <el-input-number v-model="domainForm.defaultTtl" :min="0" :max="31536000" />
          <span class="ml-2">秒</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="addDomain">添加</el-button>
      </template>
    </el-dialog>

    <!-- 添加规则对话框 -->
    <el-dialog v-model="showRuleDialog" title="添加缓存规则" width="500px">
      <el-form :model="ruleForm" label-width="120px">
        <el-form-item label="规则名称">
          <el-input v-model="ruleForm.name" placeholder="图片缓存" />
        </el-form-item>
        <el-form-item label="路径匹配">
          <el-input v-model="ruleForm.pathPattern" placeholder="*.jpg,*.png,*.gif" />
        </el-form-item>
        <el-form-item label="缓存时间">
          <el-input-number v-model="ruleForm.ttl" :min="0" :max="31536000" />
          <span class="ml-2">秒</span>
        </el-form-item>
        <el-form-item label="优先级">
          <el-input-number v-model="ruleForm.priority" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="忽略参数">
          <el-switch v-model="ruleForm.ignoreParam" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRuleDialog = false">取消</el-button>
        <el-button type="primary" @click="addRule">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const loading = ref(false)
const rulesLoading = ref(false)
const nodesLoading = ref(false)
const domains = ref([])
const cacheRules = ref([])
const cdnNodes = ref([])
const currentDomain = ref<any>(null)
const showAddDialog = ref(false)
const showPurgeCacheDialog = ref(false)
const showRuleDialog = ref(false)
const purgeURLs = ref('')

const domainForm = ref({
  name: '',
  domain: '',
  provider: 'aliyun',
  originType: 'domain',
  originHost: '',
  originPort: 80,
  enableHttps: true,
  defaultTtl: 86400,
  autoOptimize: true
})

const ruleForm = ref({
  domainId: 0,
  name: '',
  pathPattern: '',
  ttl: 86400,
  priority: 10,
  ignoreParam: false,
  enabled: true
})

const cdnStats = computed(() => {
  const ds = domains.value
  return {
    totalBandwidth: ds.reduce((sum: number, d: any) => sum + (d.bandwidth || 0), 0),
    totalTraffic: ds.reduce((sum: number, d: any) => sum + (d.traffic || 0), 0),
    totalRequests: ds.reduce((sum: number, d: any) => sum + (d.requestCount || 0), 0),
    totalCost: ds.reduce((sum: number, d: any) => sum + (d.monthlyCost || 0), 0)
  }
})

const fetchDomains = async () => {
  loading.value = true
  try {
    const res = await request.get('/api/v1/cdn/domains')
    domains.value = res.data || []
    if (domains.value.length > 0) {
      viewDetail(domains.value[0])
    }
  } catch (error) {
    ElMessage.error('获取域名列表失败')
  } finally {
    loading.value = false
  }
}

const viewDetail = async (domain: any) => {
  currentDomain.value = domain
  ruleForm.value.domainId = domain.id
  await fetchCacheRules(domain.id)
  await fetchNodes(domain.id)
}

const fetchCacheRules = async (domainId: number) => {
  rulesLoading.value = true
  try {
    const res = await request.get(`/api/v1/cdn/domains/${domainId}/rules`)
    cacheRules.value = res.data || []
  } catch (error) {
    console.error('获取缓存规则失败', error)
  } finally {
    rulesLoading.value = false
  }
}

const fetchNodes = async (domainId: number) => {
  nodesLoading.value = true
  try {
    const res = await request.get(`/api/v1/cdn/domains/${domainId}/nodes`)
    cdnNodes.value = res.data || []
  } catch (error) {
    console.error('获取节点状态失败', error)
  } finally {
    nodesLoading.value = false
  }
}

const addDomain = async () => {
  try {
    await request.post('/api/v1/cdn/domains', domainForm.value)
    ElMessage.success('添加成功')
    showAddDialog.value = false
    fetchDomains()
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const optimizeCDN = async (domain: any) => {
  try {
    const res = await request.post(`/api/v1/cdn/domains/${domain.id}/optimize`)
    ElMessage.success('AI 优化已执行')
    fetchDomains()
  } catch (error) {
    ElMessage.error('优化失败')
  }
}

const showPurgeDialog = (domain: any) => {
  currentDomain.value = domain
  purgeURLs.value = ''
  showPurgeCacheDialog.value = true
}

const purgeCache = async () => {
  const urls = purgeURLs.value.split('\n').filter(u => u.trim())
  if (urls.length === 0) {
    ElMessage.warning('请输入要刷新的URL')
    return
  }
  try {
    await request.post(`/api/v1/cdn/domains/${currentDomain.value.id}/purge`, { urls })
    ElMessage.success('刷新请求已提交')
    showPurgeCacheDialog.value = false
  } catch (error) {
    ElMessage.error('刷新失败')
  }
}

const costOptimize = async () => {
  if (!currentDomain.value) {
    ElMessage.warning('请先选择一个域名')
    return
  }
  try {
    await request.post(`/api/v1/cdn/domains/${currentDomain.value.id}/cost-optimize`)
    ElMessage.success('成本优化分析已完成')
  } catch (error) {
    ElMessage.error('分析失败')
  }
}

const preheatHotContent = async () => {
  if (!currentDomain.value) {
    ElMessage.warning('请先选择一个域名')
    return
  }
  try {
    await request.post(`/api/v1/cdn/domains/${currentDomain.value.id}/preheat`, { urls: [] })
    ElMessage.success('热点内容预热已启动')
  } catch (error) {
    ElMessage.error('预热失败')
  }
}

const addRule = async () => {
  try {
    await request.post(`/api/v1/cdn/domains/${currentDomain.value.id}/rules`, ruleForm.value)
    ElMessage.success('添加成功')
    showRuleDialog.value = false
    fetchCacheRules(currentDomain.value.id)
  } catch (error) {
    ElMessage.error('添加失败')
  }
}

const editRule = (rule: any) => {
  ElMessage.info('编辑规则: ' + rule.name)
}

const deleteRule = async (rule: any) => {
  try {
    await request.delete(`/api/v1/cdn/rules/${rule.id}`)
    ElMessage.success('删除成功')
    fetchCacheRules(currentDomain.value.id)
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const formatTTL = (seconds: number) => {
  if (seconds >= 86400) return `${seconds / 86400} 天`
  if (seconds >= 3600) return `${seconds / 3600} 小时`
  if (seconds >= 60) return `${seconds / 60} 分钟`
  return `${seconds} 秒`
}

onMounted(() => {
  fetchDomains()
})
</script>

<style scoped>
.cdn-page {
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

.ml-2 {
  margin-left: 8px;
}
</style>
