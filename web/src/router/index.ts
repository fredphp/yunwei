import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/store/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/',
    component: () => import('@/layout/index.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' }
      },
      // 服务器管理
      {
        path: 'servers/list',
        name: 'ServerList',
        component: () => import('@/views/servers/index.vue'),
        meta: { title: '服务器列表', icon: 'Monitor' }
      },
      {
        path: 'servers/alerts',
        name: 'Alerts',
        component: () => import('@/views/alerts/index.vue'),
        meta: { title: '告警管理', icon: 'Bell' }
      },
      // Kubernetes
      {
        path: 'kubernetes/clusters',
        name: 'Clusters',
        component: () => import('@/views/kubernetes/index.vue'),
        meta: { title: '集群管理', icon: 'Grid' }
      },
      // 其他功能模块
      {
        path: 'canary',
        name: 'Canary',
        component: () => import('@/views/canary/index.vue'),
        meta: { title: '灰度发布', icon: 'Promotion' }
      },
      {
        path: 'loadbalancer',
        name: 'LoadBalancer',
        component: () => import('@/views/loadbalancer/index.vue'),
        meta: { title: '负载均衡', icon: 'Connection' }
      },
      {
        path: 'certificate',
        name: 'Certificate',
        component: () => import('@/views/certificate/index.vue'),
        meta: { title: '证书管理', icon: 'Key' }
      },
      {
        path: 'cdn',
        name: 'CDN',
        component: () => import('@/views/cdn/index.vue'),
        meta: { title: 'CDN 管理', icon: 'Cloudy' }
      },
      {
        path: 'deploy',
        name: 'Deploy',
        component: () => import('@/views/deploy/index.vue'),
        meta: { title: '智能部署', icon: 'Position' }
      },
      {
        path: 'scheduler',
        name: 'Scheduler',
        component: () => import('@/views/scheduler/index.vue'),
        meta: { title: '任务调度', icon: 'Timer' }
      },
      {
        path: 'agents',
        name: 'Agents',
        component: () => import('@/views/agents/index.vue'),
        meta: { title: 'Agent 管理', icon: 'Cpu' }
      },
      {
        path: 'ha',
        name: 'HA',
        component: () => import('@/views/ha/index.vue'),
        meta: { title: '高可用', icon: 'Connection' }
      },
      {
        path: 'backup',
        name: 'Backup',
        component: () => import('@/views/backup/index.vue'),
        meta: { title: '灾备备份', icon: 'FolderOpened' }
      },
      {
        path: 'cost',
        name: 'Cost',
        component: () => import('@/views/cost/index.vue'),
        meta: { title: '成本控制', icon: 'Wallet' }
      },
      // 系统管理
      {
        path: 'system/user',
        name: 'UserManage',
        component: () => import('@/views/system/user/index.vue'),
        meta: { title: '用户管理', icon: 'User' }
      },
      {
        path: 'system/role',
        name: 'RoleManage',
        component: () => import('@/views/system/role/index.vue'),
        meta: { title: '角色管理', icon: 'UserFilled' }
      },
      {
        path: 'system/menu',
        name: 'MenuManage',
        component: () => import('@/views/system/menu/index.vue'),
        meta: { title: '菜单管理', icon: 'Menu' }
      },
      // 其他页面
      {
        path: 'tenant',
        name: 'Tenant',
        component: () => import('@/views/tenant/index.vue'),
        meta: { title: '租户管理', icon: 'OfficeBuilding' }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile/index.vue'),
        meta: { title: '个人中心', icon: 'UserFilled' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { title: '404' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const token = userStore.token || localStorage.getItem('token')

  // 设置页面标题
  document.title = (to.meta.title as string) || 'AI 运维管理系统'

  // 白名单路由
  const whiteList = ['/login', '/register']
  
  if (token) {
    if (to.path === '/login') {
      next('/')
    } else {
      // 检查是否已获取用户信息
      if (!userStore.userInfo.id) {
        userStore.getUserInfo().then(() => {
          next()
        }).catch(() => {
          userStore.logout()
          next('/login')
        })
      } else {
        next()
      }
    }
  } else {
    if (whiteList.includes(to.path)) {
      next()
    } else {
      next('/login')
    }
  }
})

export default router
