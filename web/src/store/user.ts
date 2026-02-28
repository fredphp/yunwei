import { defineStore } from 'pinia'
import { login, getUserInfo, logout } from '@/api/user'
import { ref } from 'vue'

interface UserInfo {
  id: number
  username: string
  nickName: string
  avatar: string
  email: string
  phone: string
  roleId: number
}

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo>({
    id: 0,
    username: '',
    nickName: '',
    avatar: '',
    email: '',
    phone: '',
    roleId: 0
  })

  // 登录
  const loginAction = async (username: string, password: string) => {
    try {
      const res = await login({ username, password })
      token.value = res.data.token
      localStorage.setItem('token', res.data.token)
      return res
    } catch (error) {
      throw error
    }
  }

  // 获取用户信息
  const getUserInfoAction = async () => {
    try {
      const res = await getUserInfo()
      userInfo.value = res.data
      return res
    } catch (error) {
      throw error
    }
  }

  // 登出
  const logoutAction = () => {
    token.value = ''
    userInfo.value = {
      id: 0,
      username: '',
      nickName: '',
      avatar: '',
      email: '',
      phone: '',
      roleId: 0
    }
    localStorage.removeItem('token')
  }

  return {
    token,
    userInfo,
    login: loginAction,
    getUserInfo: getUserInfoAction,
    logout: logoutAction
  }
})
