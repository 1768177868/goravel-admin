import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../store/user'
import logger from '../utils/logger'

/**
 * 带重试和错误处理的动态导入包装函数
 * @param {Function} importFn - 动态导入函数
 * @param {number} maxRetries - 最大重试次数，默认 3
 * @param {number} timeout - 超时时间（毫秒），默认 30 秒
 * @returns {Promise} 导入的模块
 */
function lazyLoad(importFn, maxRetries = 3, timeout = 10000) {
  return new Promise((resolve, reject) => {
    let retryCount = 0
    
    const attemptLoad = () => {
      // 创建超时 Promise
      const timeoutPromise = new Promise((_, timeoutReject) => {
        setTimeout(() => {
          timeoutReject(new Error('模块加载超时'))
        }, timeout)
      })
      
      // 创建加载 Promise
      const loadPromise = importFn().catch(err => {
        // 如果是网络错误或加载失败，可以重试
        if (err.message && (
          err.message.includes('Failed to fetch') ||
          err.message.includes('Loading chunk') ||
          err.message.includes('Loading CSS chunk')
        )) {
          throw err
        }
        throw err
      })
      
      // 竞争加载和超时
      Promise.race([loadPromise, timeoutPromise])
        .then(module => {
          resolve(module)
        })
        .catch(error => {
          retryCount++
          
          if (retryCount < maxRetries) {
            logger.warn(`模块加载失败，正在重试 (${retryCount}/${maxRetries}):`, error.message)
            // 指数退避：1秒、2秒、4秒
            const delay = Math.min(1000 * Math.pow(2, retryCount - 1), 5000)
            setTimeout(() => {
              attemptLoad()
            }, delay)
          } else {
            logger.error('模块加载失败，已达到最大重试次数:', error)
            ElMessage.error({
              message: '页面加载失败，请刷新页面重试',
              duration: 5000,
              showClose: true
            })
            reject(error)
          }
        })
    }
    
    attemptLoad()
  })
}

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => lazyLoad(() => import('../views/Login.vue')),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => lazyLoad(() => import('../layouts/MainLayout.vue')),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => lazyLoad(() => import('../views/Dashboard.vue')),
        meta: { titleKey: 'menu.dashboard' }
      },
      {
        path: 'admins',
        name: 'Admins',
        component: () => lazyLoad(() => import('../views/admin/AdminList.vue')),
        meta: { titleKey: 'menu.admin_management' }
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => lazyLoad(() => import('../views/role/RoleList.vue')),
        meta: { titleKey: 'menu.role_management' }
      },
      {
        path: 'permissions',
        name: 'Permissions',
        component: () => lazyLoad(() => import('../views/permission/PermissionList.vue')),
        meta: { titleKey: 'menu.permission_management' }
      },
      {
        path: 'menus',
        name: 'Menus',
        component: () => lazyLoad(() => import('../views/menu/MenuList.vue')),
        meta: { titleKey: 'menu.menu_management' }
      },
      {
        path: 'departments',
        name: 'Departments',
        component: () => lazyLoad(() => import('../views/department/DepartmentList.vue')),
        meta: { titleKey: 'menu.department_management' }
      },
      {
        path: 'dictionaries',
        name: 'Dictionaries',
        component: () => lazyLoad(() => import('../views/dictionary/DictionaryList.vue')),
        meta: { titleKey: 'menu.dictionary_management' }
      },
      {
        path: 'configs',
        name: 'Configs',
        component: () => lazyLoad(() => import('../views/config/ConfigList.vue')),
        meta: { titleKey: 'menu.config_management' }
      },
      {
        path: 'exports',
        name: 'Exports',
        component: () => lazyLoad(() => import('../views/export/ExportList.vue')),
        meta: { titleKey: 'menu.export_management' }
      },
      {
        path: 'attachments',
        name: 'Attachments',
        component: () => lazyLoad(() => import('../views/attachment/AttachmentList.vue')),
        meta: { titleKey: 'menu.attachment_management' }
      },
      {
        path: 'blacklists',
        name: 'Blacklists',
        component: () => lazyLoad(() => import('../views/blacklist/BlacklistList.vue')),
        meta: { titleKey: 'menu.blacklist_management' }
      },
      {
        path: 'online-users',
        name: 'OnlineUser',
        component: () => lazyLoad(() => import('../views/onlineUser/OnlineUserList.vue')),
        meta: { titleKey: 'menu.online_user_management' }
      },
      {
        path: 'operation-logs',
        name: 'OperationLogs',
        component: () => lazyLoad(() => import('../views/log/OperationLogList.vue')),
        meta: { titleKey: 'menu.operation_log' }
      },
      {
        path: 'login-logs',
        name: 'LoginLogs',
        component: () => lazyLoad(() => import('../views/log/LoginLogList.vue')),
        meta: { titleKey: 'menu.login_log' }
      },
      {
        path: 'system-logs',
        name: 'SystemLogs',
        component: () => lazyLoad(() => import('../views/log/SystemLogList.vue')),
        meta: { titleKey: 'menu.system_log' }
      },
      {
        path: 'notifications',
        name: 'Notifications',
        component: () => lazyLoad(() => import('../views/notification/NotificationList.vue')),
        meta: { titleKey: 'menu.notification_center' }
      },
      {
        path: 'monitor',
        name: 'Monitor',
        component: () => lazyLoad(() => import('../views/monitor/Monitor.vue')),
        meta: { titleKey: 'menu.service_monitor' }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => lazyLoad(() => import('../views/profile/Profile.vue')),
        meta: { titleKey: 'menu.profile' }
      },
      {
        path: 'iframe',
        name: 'Iframe',
        component: () => lazyLoad(() => import('../views/iframe/IframeView.vue')),
        meta: { titleKey: 'menu.external_link' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth === false) {
    // 登录页面，如果已登录则跳转到首页
    if (userStore.isLoggedIn) {
      next('/')
    } else {
      next()
    }
  } else {
    // 需要认证的页面
    if (!userStore.isLoggedIn) {
      // 如果没有token，直接跳转到登录页
      next('/login')
    } else {
      // 优化：只在首次加载（从登录页或刷新页面）时才阻塞导航
      // 如果用户信息已获取过，允许导航继续，菜单可以在后台异步加载
      const isFirstLoad = !from.name || from.name === 'Login'
      
      // 检查菜单是否为空（即使 userInfoFetched 为 true，菜单也可能被意外清空）
      const menusEmpty = !userStore.menus || userStore.menus.length === 0
      
      // 如果用户信息已获取过，但菜单为空，需要重新获取菜单（不阻塞导航）
      if (userStore.userInfoFetched && menusEmpty && userStore.adminInfo) {
        // 后台异步获取菜单，不阻塞导航
        userStore.fetchUserInfo(true).catch((error) => {
          logger.error('Failed to refresh menus in background:', error)
        })
        next()
        return
      }
      
      // 如果用户信息已获取过且菜单不为空，直接允许导航
      if (userStore.userInfoFetched && !menusEmpty) {
        next()
        return
      }
      
      // 首次加载：如果用户信息不存在或菜单为空，需要获取
      if (!userStore.adminInfo || menusEmpty) {
        // 阻塞导航，等待用户信息加载完成
        userStore.fetchUserInfo().then(() => {
          next()
        }).catch((error) => {
          // 如果获取用户信息失败（可能是401），拦截器会处理跳转
          // 这里只需要阻止导航
          next(false)
        })
      } else {
        // 用户信息已存在，直接允许导航
        // 标记为已获取，避免后续路由切换时重复检查
        userStore.userInfoFetched = true
        next()
      }
    }
  }
})

// 捕获路由错误（包括动态导入失败）
router.onError((error) => {
  logger.error('Router error:', error)
  
  // 检查是否是动态导入失败
  if (error.message && (
    error.message.includes('Failed to fetch dynamically imported module') ||
    error.message.includes('Loading chunk') ||
    error.message.includes('Loading CSS chunk') ||
    error.name === 'ChunkLoadError'
  )) {
    ElMessage.error({
      message: '页面加载失败，请刷新页面重试',
      duration: 5000,
      showClose: true
    })
    
    // 可以尝试重新加载页面
    const retry = () => {
      window.location.reload()
    }
    
    // 延迟 2 秒后自动刷新，给用户时间看到错误提示
    setTimeout(retry, 2000)
  } else {
    ElMessage.error({
      message: '路由导航失败，请刷新页面',
      duration: 3000
    })
  }
})

export default router

