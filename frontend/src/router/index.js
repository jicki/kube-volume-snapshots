import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Dashboard from '../views/Dashboard.vue'
import SnapshotClasses from '../views/SnapshotClasses.vue'
import Snapshots from '../views/Snapshots.vue'
import ScheduledTasks from '../views/ScheduledTasks.vue'
import PVCs from '../views/PVCs.vue'
import CephCluster from '../views/CephCluster.vue'
import Login from '../views/Login.vue'
import UserManagement from '../views/UserManagement.vue'
import ClusterManagement from '../views/ClusterManagement.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: {
      requiresAuth: false,
      title: '登录'
    }
  },
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: {
      requiresAuth: true,
      title: '仪表板'
    }
  },
  {
    path: '/snapshot-classes',
    name: 'SnapshotClasses',
    component: SnapshotClasses,
    meta: {
      requiresAuth: true,
      title: '快照类'
    }
  },
  {
    path: '/snapshots',
    name: 'Snapshots',
    component: Snapshots,
    meta: {
      requiresAuth: true,
      title: '快照管理'
    }
  },
  {
    path: '/scheduled',
    name: 'ScheduledTasks',
    component: ScheduledTasks,
    meta: {
      requiresAuth: true,
      title: '定时任务'
    }
  },
  {
    path: '/pvcs',
    name: 'PVCs',
    component: PVCs,
    meta: {
      requiresAuth: true,
      title: 'PVC 管理'
    }
  },
  {
    path: '/ceph',
    name: 'CephCluster',
    component: CephCluster,
    meta: {
      requiresAuth: true,
      title: 'Ceph集群'
    }
  },
  {
    path: '/users',
    name: 'UserManagement',
    component: UserManagement,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: '用户管理'
    }
  },
  {
    path: '/clusters',
    name: 'ClusterManagement',
    component: ClusterManagement,
    meta: {
      requiresAuth: true,
      title: '集群管理'
    }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  console.log(`[路由守卫] 导航到: ${to.path}, 来自: ${from.path}`)

  const authStore = useAuthStore()
  console.log(`[路由守卫] 当前认证状态: ${authStore.isAuthenticated}`)

  // 如果访问登录页且已登录，重定向到首页
  if (to.path === '/login' && authStore.isAuthenticated) {
    console.log('[路由守卫] 已登录用户访问登录页，重定向到首页')
    next('/')
    return
  }

  // 如果路由不需要认证，直接通过
  if (to.meta.requiresAuth === false) {
    console.log('[路由守卫] 路由不需要认证，直接通过')
    next()
    return
  }

  // 默认情况下，路由需要认证（除非明确设置为false）
  const requiresAuth = to.meta.requiresAuth !== false

  if (requiresAuth) {
    console.log('[路由守卫] 路由需要认证，检查认证状态')

    // 检查认证状态
    if (!authStore.isAuthenticated) {
      console.log('[路由守卫] 用户未认证，尝试从本地存储恢复')

      // 尝试从本地存储恢复认证状态
      const restored = authStore.restoreAuth()
      console.log(`[路由守卫] 恢复认证状态结果: ${restored}`)

      if (restored) {
        console.log('[路由守卫] 找到本地token，验证有效性')
        // 验证token是否有效
        const isValid = await authStore.checkToken()
        console.log(`[路由守卫] Token验证结果: ${isValid}`)

        if (!isValid) {
          console.log('[路由守卫] Token无效，重定向到登录页')
          next('/login')
          return
        }
        console.log('[路由守卫] Token有效，继续导航')
      } else {
        console.log('[路由守卫] 没有找到有效的认证信息，重定向到登录页')
        next('/login')
        return
      }
    }

    // 检查管理员权限
    if (to.meta.requiresAdmin && !authStore.isAdmin) {
      console.log('[路由守卫] 需要管理员权限但用户不是管理员，重定向到仪表板')
      next('/dashboard')
      return
    }
  }

  console.log('[路由守卫] 认证检查通过，继续导航')
  next()
})

export default router
