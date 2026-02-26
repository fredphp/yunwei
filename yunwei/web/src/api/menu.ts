import { get, post, put, del, Response } from '@/utils/request'

// 菜单信息
interface MenuInfo {
  id: number
  parentId: number
  title: string
  name: string
  path: string
  component: string
  icon: string
  sort: number
  status: number
  hidden: number
  children?: MenuInfo[]
}

// 创建菜单参数
interface CreateMenuParams {
  parentId?: number
  title: string
  name: string
  path?: string
  component?: string
  icon?: string
  sort?: number
  hidden?: number
}

// 更新菜单参数
interface UpdateMenuParams {
  id: number
  parentId?: number
  title: string
  name: string
  path?: string
  component?: string
  icon?: string
  sort?: number
  hidden?: number
  status?: number
}

// 获取菜单列表
export function getMenuList(): Promise<Response<MenuInfo[]>> {
  return get('/menus')
}

// 获取菜单详情
export function getMenu(id: number): Promise<Response<MenuInfo>> {
  return get(`/menus/${id}`)
}

// 创建菜单
export function createMenu(data: CreateMenuParams): Promise<Response<null>> {
  return post('/menus', data)
}

// 更新菜单
export function updateMenu(data: UpdateMenuParams): Promise<Response<null>> {
  return put(`/menus/${data.id}`, data)
}

// 删除菜单
export function deleteMenu(id: number): Promise<Response<null>> {
  return del(`/menus/${id}`)
}

// 获取用户菜单
export function getUserMenus(): Promise<Response<MenuInfo[]>> {
  return get('/user/menus')
}
