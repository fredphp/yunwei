import request from '@/utils/request'

// 租户类型定义
export interface Tenant {
  id: string
  name: string
  slug: string
  domain?: string
  logo?: string
  description?: string
  status: 'active' | 'suspended' | 'deleted'
  plan: 'free' | 'starter' | 'pro' | 'enterprise'
  billing_cycle?: string
  contact_name?: string
  contact_email?: string
  contact_phone?: string
  address?: string
  settings?: Record<string, any>
  features?: Record<string, any>
  created_at: string
  updated_at: string
  quota?: TenantQuota
}

export interface TenantQuota {
  id: string
  tenant_id: string
  max_users: number
  max_admins: number
  max_resources: number
  max_servers: number
  max_databases: number
  max_monitors: number
  max_alert_rules: number
  metrics_retention: number
  max_cloud_accounts: number
  budget_limit: number
  max_storage_gb: number
  max_backup_gb: number
  max_api_calls: number
  max_webhooks: number
  current_users: number
  current_resources: number
  current_storage_gb: number
  current_api_calls: number
}

export interface TenantUser {
  id: string
  tenant_id: string
  user_id: string
  email: string
  name: string
  avatar?: string
  role_id: string
  role_name: string
  is_owner: boolean
  is_admin: boolean
  status: 'active' | 'inactive' | 'pending'
  invited_by?: string
  joined_at: string
  last_active_at?: string
}

export interface TenantRole {
  id: string
  tenant_id: string
  name: string
  slug: string
  description?: string
  is_system: boolean
  permissions: Record<string, any>
  scope: string
  parent_id?: string
}

export interface CreateTenantRequest {
  name: string
  slug: string
  plan?: string
  owner_email: string
  owner_name: string
  domain?: string
  contact_phone?: string
}

export interface UpdateTenantRequest {
  name?: string
  domain?: string
  description?: string
  contact_name?: string
  contact_email?: string
  contact_phone?: string
  address?: string
}

export interface TenantListResponse {
  data: Tenant[]
  total: number
  page: number
  page_size: number
}

// 平台管理员接口

// 获取租户列表
export function getTenantList(params?: { page?: number; page_size?: number; status?: string; plan?: string }) {
  return request.get<any, TenantListResponse>('/admin/tenants', params)
}

// 创建租户
export function createTenant(data: CreateTenantRequest) {
  return request.post<any, { data: Tenant }>('/admin/tenants', data)
}

// 获取租户详情
export function getTenant(id: string) {
  return request.get<any, { data: Tenant }>(`/admin/tenants/${id}`)
}

// 更新租户
export function updateTenant(id: string, data: UpdateTenantRequest) {
  return request.put<any, { message: string }>(`/admin/tenants/${id}`, data)
}

// 删除租户
export function deleteTenant(id: string) {
  return request.delete<any, { message: string }>(`/admin/tenants/${id}`)
}

// 暂停租户
export function suspendTenant(id: string, reason?: string) {
  return request.post<any, { message: string }>(`/admin/tenants/${id}/suspend`, { reason })
}

// 激活租户
export function activateTenant(id: string) {
  return request.post<any, { message: string }>(`/admin/tenants/${id}/activate`)
}

// 升级套餐
export function upgradePlan(id: string, plan: string) {
  return request.post<any, { message: string }>(`/admin/tenants/${id}/upgrade`, { plan })
}

// 租户内部接口

// 获取当前租户信息
export function getCurrentTenant() {
  return request.get<any, { data: Tenant }>('/tenant/info')
}

// 更新当前租户信息
export function updateCurrentTenant(data: UpdateTenantRequest) {
  return request.put<any, { message: string }>('/tenant/info', data)
}

// 获取租户用户列表
export function getTenantUsers(params?: { page?: number; page_size?: number }) {
  return request.get<any, { data: TenantUser[]; total: number; page: number; page_size: number }>('/tenant/users', params)
}

// 添加用户到租户
export function addTenantUser(data: { email: string; name: string; role_id: string }) {
  return request.post<any, { data: TenantUser }>('/tenant/users', data)
}

// 从租户移除用户
export function removeTenantUser(userId: string) {
  return request.delete<any, { message: string }>(`/tenant/users/${userId}`)
}

// 更新用户角色
export function updateUserRole(userId: string, roleId: string) {
  return request.put<any, { message: string }>(`/tenant/users/${userId}/role`, { role_id: roleId })
}

// 更新用户状态
export function updateUserStatus(userId: string, status: string) {
  return request.put<any, { message: string }>(`/tenant/users/${userId}/status`, { status })
}

// 获取租户角色列表
export function getTenantRoles() {
  return request.get<any, { data: TenantRole[] }>('/tenant/roles')
}

// 创建租户角色
export function createTenantRole(data: { name: string; slug: string; description?: string; permissions?: string[] }) {
  return request.post<any, { data: TenantRole }>('/tenant/roles', data)
}

// 更新租户角色
export function updateTenantRole(roleId: string, data: Record<string, any>) {
  return request.put<any, { message: string }>(`/tenant/roles/${roleId}`, data)
}

// 删除租户角色
export function deleteTenantRole(roleId: string) {
  return request.delete<any, { message: string }>(`/tenant/roles/${roleId}`)
}

// 获取租户配额
export function getTenantQuota() {
  return request.get<any, { data: TenantQuota }>('/tenant/quota')
}

// 获取租户使用量
export function getTenantUsage() {
  return request.get<any, { data: Record<string, number> }>('/tenant/usage')
}

// 获取租户审计日志
export function getTenantAuditLogs(params?: { page?: number; page_size?: number }) {
  return request.get<any, { data: any[] }>('/tenant/audit-logs', params)
}

// 套餐价格映射
export const planPrices: Record<string, { name: string; price: number; description: string }> = {
  free: { name: 'Free', price: 0, description: '免费版，适合个人学习使用' },
  starter: { name: 'Starter', price: 99, description: '入门版，适合小型团队' },
  pro: { name: 'Pro', price: 299, description: '专业版，适合中型企业' },
  enterprise: { name: 'Enterprise', price: 0, description: '企业版，联系销售定制' }
}

// 状态映射
export const statusMap: Record<string, { label: string; type: 'success' | 'warning' | 'danger' | 'info' }> = {
  active: { label: '正常', type: 'success' },
  suspended: { label: '已暂停', type: 'warning' },
  deleted: { label: '已删除', type: 'danger' }
}
