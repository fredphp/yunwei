import { get, post, put, del, Response, PageResponse } from '@/utils/request'

// 登录请求参数
interface LoginParams {
  username: string
  password: string
  captcha?: string
  captchaId?: string
}

// 登录响应
interface LoginResponse {
  token: string
  user: {
    id: number
    username: string
    nickName: string
    avatar: string
    roleId: number
  }
}

// 用户信息
interface UserInfo {
  id: number
  username: string
  nickName: string
  avatar: string
  email: string
  phone: string
  status: number
  roleId: number
  role: {
    id: number
    name: string
    keyword: string
  }
  createdAt: string
  updatedAt: string
}

// 用户列表参数
interface UserListParams {
  page: number
  pageSize: number
  username?: string
  status?: number
}

// 创建用户参数
interface CreateUserParams {
  username: string
  password: string
  nickName?: string
  email?: string
  phone?: string
  roleId?: number
}

// 更新用户参数
interface UpdateUserParams {
  id: number
  nickName?: string
  email?: string
  phone?: string
  roleId?: number
  status?: number
}

// 登录
export function login(data: LoginParams): Promise<Response<LoginResponse>> {
  return post('/login', data)
}

// 注册
export function register(data: LoginParams): Promise<Response<null>> {
  return post('/register', data)
}

// 获取用户信息
export function getUserInfo(): Promise<Response<UserInfo>> {
  return get('/user/info')
}

// 获取用户列表
export function getUserList(params: UserListParams): Promise<Response<PageResponse<UserInfo>>> {
  return get('/users', params)
}

// 获取用户详情
export function getUser(id: number): Promise<Response<UserInfo>> {
  return get(`/users/${id}`)
}

// 创建用户
export function createUser(data: CreateUserParams): Promise<Response<null>> {
  return post('/users', data)
}

// 更新用户
export function updateUser(data: UpdateUserParams): Promise<Response<null>> {
  return put(`/users/${data.id}`, data)
}

// 删除用户
export function deleteUser(id: number): Promise<Response<null>> {
  return del(`/users/${id}`)
}

// 修改密码
export function changePassword(data: { password: string; newPassword: string }): Promise<Response<null>> {
  return put('/user/password', data)
}
