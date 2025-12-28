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

// 固定路由（不需要从接口获取）
const staticRoutes = [
  {
    path: '/login',
    name: 'Login',
    component: () => lazyLoad(() => import('../views/Login.vue')),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'MainLayout',
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

/**
 * 组件导入映射表
 * 将所有可能的组件路径都明确列出，以便 Vite 可以静态分析
 */
const componentImportMap = {
  'admin/index': () => lazyLoad(() => import('../views/admin/AdminList.vue')),
  'role/index': () => lazyLoad(() => import('../views/role/RoleList.vue')),
  'permission/index': () => lazyLoad(() => import('../views/permission/PermissionList.vue')),
  'menu/index': () => lazyLoad(() => import('../views/menu/MenuList.vue')),
  'department/index': () => lazyLoad(() => import('../views/department/DepartmentList.vue')),
  'dictionary/index': () => lazyLoad(() => import('../views/dictionary/DictionaryList.vue')),
  'config/index': () => lazyLoad(() => import('../views/config/ConfigList.vue')),
  'export/index': () => lazyLoad(() => import('../views/export/ExportList.vue')),
  'attachment/index': () => lazyLoad(() => import('../views/attachment/AttachmentList.vue')),
  'blacklist/index': () => lazyLoad(() => import('../views/blacklist/BlacklistList.vue')),
  'order/index': () => lazyLoad(() => import('../views/order/OrderList.vue')),
  'user/index': () => lazyLoad(() => import('../views/user/UserList.vue')),
  'user-balance-log/index': () => lazyLoad(() => import('../views/user/UserBalanceLogList.vue')),
  'onlineAdmin/index': () => lazyLoad(() => import('../views/onlineAdmin/OnlineAdminList.vue')),
  'log/operation/index': () => lazyLoad(() => import('../views/log/OperationLogList.vue')),
  'log/login/index': () => lazyLoad(() => import('../views/log/LoginLogList.vue')),
  'log/system/index': () => lazyLoad(() => import('../views/log/SystemLogList.vue')),
  'notification/index': () => lazyLoad(() => import('../views/notification/NotificationList.vue')),
  'monitor/index': () => lazyLoad(() => import('../views/monitor/Monitor.vue')),
  'profile/index': () => lazyLoad(() => import('../views/profile/Profile.vue'))
}

/**
 * 获取组件的导入函数
 * @param {string} component - 菜单的 component 字段，如 "admin/index", "log/operation/index"
 * @returns {Function|null} 组件导入函数，如果不存在则返回 null
 */
function getComponentImport(component) {
  if (!component || component === 'Layout') {
    return null
  }

  // 如果映射表中存在，直接返回
  if (componentImportMap[component]) {
    return componentImportMap[component]
  }

  // 如果映射表中不存在，返回 null（可能是外部链接或其他特殊类型）
  // 注意：为了支持 Vite 静态分析，我们不使用动态路径拼接
  // 如果需要添加新的组件，请在 componentImportMap 中添加对应的映射
  logger.warn(`Component import not found for: ${component}, please add it to componentImportMap`)
  return null
}

/**
 * 将菜单数据转换为路由配置
 * @param {Array} menus - 菜单数组（扁平结构）
 * @returns {Array} 路由配置数组
 */
function convertMenusToRoutes(menus) {
  if (!menus || !Array.isArray(menus)) {
    return []
  }

  const routes = []
  const processedPaths = new Set() // 避免重复路由

  menus.forEach(menu => {
    // 只处理类型为菜单（type === 2）且状态为启用（status === 1）的菜单
    const type = menu.Type !== undefined ? menu.Type : (menu.type !== undefined ? menu.type : 1)
    const status = menu.Status !== undefined ? menu.Status : (menu.status !== undefined ? menu.status : 1)
    const linkType = menu.LinkType !== undefined ? menu.LinkType : (menu.link_type !== undefined ? menu.link_type : 1)
    
    // 只处理菜单类型（type === 2）且启用的菜单
    if (type !== 2 || status !== 1) {
      return
    }

    const path = menu.Path || menu.path || ''
    const component = menu.Component || menu.component || ''
    
    // 如果没有路径，跳过
    if (!path || path === '/') {
      return
    }

    // 避免重复路由
    if (processedPaths.has(path)) {
      return
    }
    processedPaths.add(path)

    // 处理路径：移除前导斜杠，子路由使用相对路径（不带前导斜杠）
    // 静态路由中的子路由都是相对路径，如 "dashboard", "profile"
    const routePath = path.startsWith('/') ? path.slice(1) : path
    
    // 生成路由名称（从路径转换，如 "admins" -> "Admins", "user-balance-logs" -> "UserBalanceLogs"）
    const routeName = routePath
      .split('-')
      .map(part => part.charAt(0).toUpperCase() + part.slice(1))
      .join('')

    // 生成 titleKey
    // 翻译文件中的键通常是 menu.xxx_management 格式
    // 但有些菜单的 slug 可能已经包含了 _management，所以需要智能处理
    const slug = menu.Slug || menu.slug || routePath
    let titleKey = `menu.${slug}`
    
    // 如果 slug 不包含 _management，尝试添加后缀
    // 但先检查原始键是否存在，如果存在就不添加后缀
    // 注意：这里我们无法直接检查翻译键，所以先尝试添加 _management
    // BreadcrumbView 会使用智能翻译函数来处理
    
    // 构建路由配置
    const route = {
      path: routePath,
      name: routeName,
      meta: {
        titleKey: titleKey,
        menuId: menu.id || menu.ID,
        menuSlug: slug // 保存 slug，供 BreadcrumbView 使用
      }
    }

    // 如果是外部链接（linkType === 2），使用 iframe 组件
    if (linkType === 2) {
      route.component = () => lazyLoad(() => import('../views/iframe/IframeView.vue'))
      route.meta.externalUrl = path
    } else {
      // 内部页面，根据 component 字段获取组件导入函数
      const componentImport = getComponentImport(component)
      if (componentImport) {
        route.component = componentImport
      } else {
        // 如果无法获取组件导入函数，跳过（可能是目录类型或其他特殊类型）
        logger.warn(`Skipping route ${path} due to missing component import for: ${component}`)
        return
      }
    }

    routes.push(route)
  })

  return routes
}

// 标记是否已经添加过动态路由
let dynamicRoutesAdded = false

// 初始路由（只包含固定路由）
const routes = [...staticRoutes]

const router = createRouter({
  history: createWebHistory(),
  routes
})

/**
 * 动态添加路由
 * @param {Array} menus - 菜单数组
 */
function addDynamicRoutes(menus) {
  if (!menus || menus.length === 0) {
    return
  }

  const dynamicRoutes = convertMenusToRoutes(menus)
  
  if (dynamicRoutes.length === 0) {
    logger.warn('No dynamic routes to add')
    return
  }

  // 检查路由是否已存在，避免重复添加
  const existingRoutes = router.getRoutes()
  const existingPaths = new Set(
    existingRoutes
      .filter(route => route.path !== '/' && route.path !== '/login')
      .flatMap(route => route.children || [])
      .map(child => child.path)
  )

  // 只添加不存在的路由
  const routesToAdd = dynamicRoutes.filter(route => !existingPaths.has(route.path))
  
  if (routesToAdd.length === 0) {
    logger.debug('All dynamic routes already exist')
    return
  }

  // 找到主布局路由（path === '/' 或 name === 'MainLayout'）
  const mainLayoutRoute = existingRoutes.find(route => route.path === '/' || route.name === 'MainLayout')
  
  if (!mainLayoutRoute) {
    logger.error('Main layout route not found')
    return
  }

  // 添加新路由到主布局路由
  // 根据错误信息，Vue Router 4.2.5 要求路径必须以 "/" 开头
  // 即使对于子路由也是如此（这可能是一个特殊要求或版本差异）
  const parentName = mainLayoutRoute.name || 'MainLayout'
  routesToAdd.forEach(route => {
    // 确保路径以 "/" 开头（根据错误信息的要求）
    const routeWithSlash = {
      ...route,
      path: route.path.startsWith('/') ? route.path : '/' + route.path
    }
    
    try {
      router.addRoute(parentName, routeWithSlash)
    } catch (error) {
      logger.error(`Failed to add route ${route.path}:`, error)
    }
  })
  
  logger.debug(`Added ${routesToAdd.length} dynamic routes:`, routesToAdd.map(r => r.path))
}

/**
 * 重置动态路由标志（在登出时调用）
 */
export function resetDynamicRoutes() {
  dynamicRoutesAdded = false
}

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth === false) {
    // 登录页面，如果已登录则跳转到首页
    if (userStore.isLoggedIn) {
      next('/')
    } else {
      // 如果未登录，重置动态路由标志
      if (dynamicRoutesAdded) {
        dynamicRoutesAdded = false
      }
      next()
    }
  } else {
    // 需要认证的页面
    if (!userStore.isLoggedIn) {
      // 如果没有token，重置动态路由标志并跳转到登录页
      if (dynamicRoutesAdded) {
        dynamicRoutesAdded = false
      }
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
        userStore.fetchUserInfo(true).then(() => {
          // 获取菜单后添加动态路由
          if (userStore.menus && userStore.menus.length > 0 && !dynamicRoutesAdded) {
            addDynamicRoutes(userStore.menus)
            dynamicRoutesAdded = true
          }
        }).catch((error) => {
          logger.error('Failed to refresh menus in background:', error)
        })
        next()
        return
      }
      
      // 如果用户信息已获取过且菜单不为空，检查是否需要添加动态路由
      if (userStore.userInfoFetched && !menusEmpty) {
        // 如果还没有添加动态路由，现在添加
        if (!dynamicRoutesAdded && userStore.menus && userStore.menus.length > 0) {
          addDynamicRoutes(userStore.menus)
          dynamicRoutesAdded = true
        }
        next()
        return
      }
      
      // 首次加载：如果用户信息不存在或菜单为空，需要获取
      if (!userStore.adminInfo || menusEmpty) {
        // 阻塞导航，等待用户信息加载完成
        userStore.fetchUserInfo().then(() => {
          // 获取菜单后添加动态路由
          if (userStore.menus && userStore.menus.length > 0 && !dynamicRoutesAdded) {
            addDynamicRoutes(userStore.menus)
            dynamicRoutesAdded = true
          }
          next()
        }).catch((error) => {
          // 如果获取用户信息失败（可能是401），拦截器会处理跳转
          // 这里只需要阻止导航
          next(false)
        })
      } else {
        // 用户信息已存在，检查是否需要添加动态路由
        if (!dynamicRoutesAdded && userStore.menus && userStore.menus.length > 0) {
          addDynamicRoutes(userStore.menus)
          dynamicRoutesAdded = true
        }
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

