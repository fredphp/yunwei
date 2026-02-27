<template>
  <div class="error-container">
    <div class="error-content">
      <h1>404</h1>
      <p>抱歉，您访问的页面不存在</p>
      <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.error-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100vh;
  background-color: #f0f2f5;
  
  .error-content {
    text-align: center;
    
    h1 {
      font-size: 120px;
      font-weight: bold;
      color: #409EFF;
      margin: 0;
    }
    
    p {
      font-size: 18px;
      color: #666;
      margin: 20px 0 30px;
    }
  }
}
</style>
