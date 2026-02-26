import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'

// 创建 axios 实例
const service: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['x-token'] = token
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data
    
    // 业务错误
    if (res.code !== 0) {
      ElMessage.error(res.msg || '请求失败')
      
      // 登录过期
      if (res.code === 401) {
        localStorage.removeItem('token')
        window.location.href = '/login'
      }
      
      return Promise.reject(new Error(res.msg || 'Error'))
    }
    
    return res
  },
  (error) => {
    let message = '请求失败'
    if (error.response) {
      switch (error.response.status) {
        case 401:
          message = '未授权，请重新登录'
          localStorage.removeItem('token')
          window.location.href = '/login'
          break
        case 403:
          message = '拒绝访问'
          break
        case 404:
          message = '请求资源不存在'
          break
        case 500:
          message = '服务器内部错误'
          break
        default:
          message = error.response.data?.msg || '请求失败'
      }
    }
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

// 封装请求方法
export interface Response<T = any> {
  code: number
  data: T
  msg: string
}

export interface PageResponse<T = any> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export function get<T>(url: string, params?: any, config?: AxiosRequestConfig): Promise<Response<T>> {
  return service.get(url, { params, ...config })
}

export function post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<Response<T>> {
  return service.post(url, data, config)
}

export function put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<Response<T>> {
  return service.put(url, data, config)
}

export function del<T>(url: string, config?: AxiosRequestConfig): Promise<Response<T>> {
  return service.delete(url, config)
}

export default service
