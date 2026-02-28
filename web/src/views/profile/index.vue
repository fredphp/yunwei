<template>
  <div class="profile-container">
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>个人信息</span>
          </template>
          <div class="user-profile">
            <el-avatar :size="100" :src="userInfo.avatar">
              {{ userInfo.nickName?.charAt(0) }}
            </el-avatar>
            <h3>{{ userInfo.nickName }}</h3>
            <p>{{ userInfo.email }}</p>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="16">
        <el-card>
          <template #header>
            <span>修改密码</span>
          </template>
          <el-form ref="formRef" :model="formData" :rules="rules" label-width="100px">
            <el-form-item label="当前密码" prop="password">
              <el-input v-model="formData.password" type="password" show-password placeholder="请输入当前密码" />
            </el-form-item>
            <el-form-item label="新密码" prop="newPassword">
              <el-input v-model="formData.newPassword" type="password" show-password placeholder="请输入新密码" />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirmPassword">
              <el-input v-model="formData.confirmPassword" type="password" show-password placeholder="请确认新密码" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSubmit">保存</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useUserStore } from '@/store/user'
import { changePassword } from '@/api/user'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const userStore = useUserStore()
const formRef = ref<FormInstance>()

const userInfo = computed(() => userStore.userInfo)

const formData = reactive({
  password: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (rule: any, value: string, callback: any) => {
  if (value !== formData.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  password: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        await changePassword({
          password: formData.password,
          newPassword: formData.newPassword
        })
        ElMessage.success('密码修改成功')
        formRef.value.resetFields()
      } catch (error) {
        console.error(error)
      }
    }
  })
}
</script>

<style scoped lang="scss">
.profile-container {
  .user-profile {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px 0;
    
    h3 {
      margin: 15px 0 5px;
      color: #333;
    }
    
    p {
      margin: 0;
      color: #999;
    }
  }
}
</style>
