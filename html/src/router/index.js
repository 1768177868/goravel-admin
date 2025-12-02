import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../store/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { titleKey: 'menu.dashboard' }
      },
      {
        path: 'admins',
        name: 'Admins',
        component: () => import('../views/admin/AdminList.vue'),
        meta: { titleKey: 'menu.admin_management' }
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('../views/role/RoleList.vue'),
        meta: { titleKey: 'menu.role_management' }
      },
      {
        path: 'permissions',
        name: 'Permissions',
        component: () => import('../views/permission/PermissionList.vue'),
        meta: { titleKey: 'menu.permission_management' }
      },
      {
        path: 'menus',
        name: 'Menus',
        component: () => import('../views/menu/MenuList.vue'),
        meta: { titleKey: 'menu.menu_management' }
      },
      {
        path: 'departments',
        name: 'Departments',
        component: () => import('../views/department/DepartmentList.vue'),
        meta: { titleKey: 'menu.department_management' }
      },
      {
        path: 'dictionaries',
        name: 'Dictionaries',
        component: () => import('../views/dictionary/DictionaryList.vue'),
        meta: { titleKey: 'menu.dictionary_management' }
      },
      {
        path: 'configs',
        name: 'Configs',
        component: () => import('../views/config/ConfigList.vue'),
        meta: { titleKey: 'menu.config_management' }
      },
      {
        path: 'exports',
        name: 'Exports',
        component: () => import('../views/export/ExportList.vue'),
        meta: { titleKey: 'menu.export_management' }
      },
      {
        path: 'attachments',
        name: 'Attachments',
        component: () => import('../views/attachment/AttachmentList.vue'),
        meta: { titleKey: 'menu.attachment_management' }
      },
      {
        path: 'blacklists',
        name: 'Blacklists',
        component: () => import('../views/blacklist/BlacklistList.vue'),
        meta: { titleKey: 'menu.blacklist_management' }
      },
      {
        path: 'online-users',
        name: 'OnlineUser',
        component: () => import('../views/onlineUser/OnlineUserList.vue'),
        meta: { titleKey: 'menu.online_user_management' }
      },
      {
        path: 'operation-logs',
        name: 'OperationLogs',
        component: () => import('../views/log/OperationLogList.vue'),
        meta: { titleKey: 'menu.operation_log' }
      },
      {
        path: 'login-logs',
        name: 'LoginLogs',
        component: () => import('../views/log/LoginLogList.vue'),
        meta: { titleKey: 'menu.login_log' }
      },
      {
        path: 'system-logs',
        name: 'SystemLogs',
        component: () => import('../views/log/SystemLogList.vue'),
        meta: { titleKey: 'menu.system_log' }
      },
      {
        path: 'notifications',
        name: 'Notifications',
        component: () => import('../views/notification/NotificationList.vue'),
        meta: { titleKey: 'menu.notification_center' }
      },
      {
        path: 'monitor',
        name: 'Monitor',
        component: () => import('../views/monitor/Monitor.vue'),
        meta: { titleKey: 'menu.service_monitor' }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('../views/profile/Profile.vue'),
        meta: { titleKey: 'menu.profile' }
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
        // 如果用户信息不存在或菜单为空（菜单不缓存，刷新后需要重新获取），尝试获取
        if (!userStore.adminInfo || userStore.menus.length === 0) {
          userStore.fetchUserInfo().then(() => {
            next()
          }).catch((error) => {
            // 如果获取用户信息失败（可能是401），拦截器会处理跳转
            // 这里只需要阻止导航
            next(false)
          })
        } else {
          next()
        }
      }
    }
})

export default router

