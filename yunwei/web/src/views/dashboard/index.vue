<template>
  <div class="dashboard-container">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon user">
              <el-icon :size="30"><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.userCount }}</div>
              <div class="stat-label">用户总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon role">
              <el-icon :size="30"><UserFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.roleCount }}</div>
              <div class="stat-label">角色总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon menu">
              <el-icon :size="30"><Menu /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.menuCount }}</div>
              <div class="stat-label">菜单总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon api">
              <el-icon :size="30"><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.apiCount }}</div>
              <div class="stat-label">API总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card>
          <template #header>
            <span>访问趋势</span>
          </template>
          <div class="chart-placeholder">
            <el-icon :size="60" color="#ddd"><TrendCharts /></el-icon>
            <p>图表区域（可集成 ECharts）</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>快捷操作</span>
          </template>
          <div class="quick-actions">
            <el-button type="primary" @click="$router.push('/system/user')">
              <el-icon><User /></el-icon>用户管理
            </el-button>
            <el-button type="success" @click="$router.push('/system/role')">
              <el-icon><UserFilled /></el-icon>角色管理
            </el-button>
            <el-button type="warning" @click="$router.push('/system/menu')">
              <el-icon><Menu /></el-icon>菜单管理
            </el-button>
            <el-button @click="$router.push('/profile')">
              <el-icon><Setting /></el-icon>个人设置
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const stats = ref({
  userCount: 128,
  roleCount: 5,
  menuCount: 15,
  apiCount: 42
})

onMounted(() => {
  // 可以在这里请求统计数据
})
</script>

<style scoped lang="scss">
.dashboard-container {
  .stat-card {
    .stat-content {
      display: flex;
      align-items: center;
      
      .stat-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 60px;
        height: 60px;
        border-radius: 10px;
        color: #fff;
        
        &.user { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
        &.role { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
        &.menu { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
        &.api { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }
      }
      
      .stat-info {
        margin-left: 15px;
        
        .stat-value {
          font-size: 28px;
          font-weight: bold;
          color: #333;
        }
        
        .stat-label {
          font-size: 14px;
          color: #999;
        }
      }
    }
  }
  
  .chart-placeholder {
    height: 300px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: #999;
  }
  
  .quick-actions {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
    
    .el-button {
      width: 100%;
    }
  }
}
</style>
