import { get, post, put, del, Response, PageResponse } from '@/utils/request'

// 角色信息
interface RoleInfo {
  id: number
  name: string
  keyword: string
  description: string
  status: number
  createdAt: string
  updatedAt: string
}

// 角色详情
interface RoleDetail {
  role: RoleInfo
  menuIds: number[]
  apiIds: number[]
}

// 角色列表参数
interface RoleListParams {
  page: number
  pageSize: number
}

// 创建角色参数
interface CreateRoleParams {
  name: string
  keyword: string
  description?: string
  menuIds?: number[]
  apiIds?: number[]
}

// 更新角色参数
interface UpdateRoleParams {
  id: number
  name: string
  keyword: string
  description?: string
  status?: number
  menuIds?: number[]
  apiIds?: number[]
}

// 获取角色列表
export function getRoleList(params: RoleListParams): Promise<Response<PageResponse<RoleInfo>>> {
  return get('/roles', params)
}

// 获取角色详情
export function getRole(id: number): Promise<Response<RoleDetail>> {
  return get(`/roles/${id}`)
}

// 创建角色
export function createRole(data: CreateRoleParams): Promise<Response<null>> {
  return post('/roles', data)
}

// 更新角色
export function updateRole(data: UpdateRoleParams): Promise<Response<null>> {
  return put(`/roles/${data.id}`, data)
}

// 删除角色
export function deleteRole(id: number): Promise<Response<null>> {
  return del(`/roles/${id}`)
}
