import { defineStore } from 'pinia'
import { getInfo, logout } from '../api/auth'
import Storage from '../utils/storage'
import logger from '../utils/logger'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: Storage.getItem('token', ''),
    adminInfo: Storage.getItem('adminInfo', null),
    permissions: [],
    menus: [], // 菜单不缓存，每次刷新都从服务器重新获取
    isSuperAdmin: false, // 是否是超级管理员
    config: {
      showButtonsWithoutPermission: false // 是否显示无权限的按钮
    }
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    hasPermission: (state) => (permission) => {
      // 超级管理员拥有所有权限
      if (state.isSuperAdmin) {
        return true
      }
      return state.permissions.includes(permission)
    },
    // 检查是否应该显示按钮（考虑权限和配置）
    shouldShowButton: (state) => (permission) => {
      // 超级管理员总是显示所有按钮
      if (state.isSuperAdmin) {
        return true
      }
      const hasPerm = state.permissions.includes(permission)
      // 如果有权限，总是显示
      if (hasPerm) {
        return true
      }
      // 如果没有权限，根据配置决定是否显示
      return state.config.showButtonsWithoutPermission
    }
  },

  actions: {
    setToken(token) {
      this.token = token
      Storage.setItem('token', token)
    },

    setAdminInfo(adminInfo) {
      // 只存储必要的基本信息，避免 localStorage 配额超限
      // 不存储 permissions, menus 等大数据字段，它们已经单独存储
      // roles 只存储基本信息（id 和 name），用于显示
      const basicInfo = {
        id: adminInfo.id || adminInfo.ID,
        username: adminInfo.username || adminInfo.Username,
        nickname: adminInfo.nickname || adminInfo.Nickname,
        avatar: adminInfo.avatar || adminInfo.Avatar,
        email: adminInfo.email || adminInfo.Email,
        phone: adminInfo.phone || adminInfo.Phone,
        department_id: adminInfo.department_id || adminInfo.DepartmentID,
        department: adminInfo.department || adminInfo.Department,
        // roles 只存储基本信息，避免存储完整的关联数据
        roles: (adminInfo.roles || adminInfo.Roles || []).map(role => ({
          id: role.id || role.ID,
          name: role.name || role.Name,
          slug: role.slug || role.Slug
        }))
      }
      this.adminInfo = basicInfo
      // 设置超级管理员标识（从后端返回或从角色判断）
      this.isSuperAdmin = adminInfo.is_super_admin === true || adminInfo.isSuperAdmin === true || 
        (adminInfo.roles || adminInfo.Roles || []).some(role => 
          (role.slug || role.Slug) === 'super-admin' && (role.status || role.Status) === 1
        )
      
      // 使用 Storage 工具保存，自动处理错误
      const saved = Storage.setItem('adminInfo', basicInfo)
      if (!saved) {
        // 如果保存失败，尝试只存储最基本信息
        const minimalInfo = {
          id: basicInfo.id,
          username: basicInfo.username,
          nickname: basicInfo.nickname,
          avatar: basicInfo.avatar
        }
        const minimalSaved = Storage.setItem('adminInfo', minimalInfo)
        if (minimalSaved) {
          this.adminInfo = minimalInfo
        } else {
          logger.error('Failed to save adminInfo even with minimal data')
        }
      }
    },

    setPermissions(permissions) {
      // 如果 permissions 是对象数组，提取 slug 字段；如果是字符串数组，直接使用
      if (Array.isArray(permissions) && permissions.length > 0) {
        if (typeof permissions[0] === 'object' && permissions[0] !== null) {
          // 对象数组，提取 slug 字段
          this.permissions = permissions.map(perm => perm.slug || perm.Slug || perm).filter(Boolean)
        } else {
          // 字符串数组，直接使用
          this.permissions = permissions
        }
      } else {
        this.permissions = []
      }
    },

    setMenus(menus) {
      this.menus = menus
      // 菜单不缓存到 localStorage，每次刷新都从服务器重新获取
    },

    setConfig(config) {
      this.config = {
        showButtonsWithoutPermission: config?.show_buttons_without_permission || config?.showButtonsWithoutPermission || false
      }
    },

    async fetchUserInfo() {
      try {
        // 清除旧的数据，确保获取最新的数据
        this.menus = []
        this.adminInfo = null
        this.permissions = []
        Storage.removeItem('adminInfo')
        
        const res = await getInfo()
        if (res.data && res.data.admin) {
          this.setAdminInfo(res.data.admin)
          // 设置权限（需要先设置，因为 setAdminInfo 可能会用到）
          this.setPermissions(res.data.admin.permissions || [])
          this.setMenus(res.data.admin.menus || [])
          // 设置超级管理员标识
          this.isSuperAdmin = res.data.admin.is_super_admin === true || res.data.admin.isSuperAdmin === true ||
            (res.data.admin.roles || []).some(role => 
              (role.slug || role.Slug) === 'super-admin' && (role.status || role.Status) === 1
            )
          
          // 调试：打印权限信息（开发环境）
          logger.debug('User permissions:', this.permissions)
          logger.debug('Is super admin:', this.isSuperAdmin)
        }
        // 保存配置信息
        if (res.data && res.data.config) {
          this.setConfig(res.data.config)
        }
        return res
      } catch (error) {
        // fetchUserInfo 失败时也应该清除状态，但这里不直接跳转，由拦截器处理
        this.logout(true)
        throw error
      }
    },

    async logout(skipApiCall = false) {
      try {
        // 如果有 token 且不需要跳过 API 调用，尝试调用后端登出接口
        if (this.token && !skipApiCall) {
          await logout()
        }
      } catch (error) {
        // 即使登出接口失败，也要清除本地状态
        logger.error('Logout error:', error)
      } finally {
        // 清除所有状态（同步执行，不等待）
        this.token = ''
        this.adminInfo = null
        this.permissions = []
        this.menus = []
        this.isSuperAdmin = false
        this.config = {
          showButtonsWithoutPermission: false
        }
        Storage.removeItem('token')
        Storage.removeItem('adminInfo')
        // 菜单不缓存，无需清除
      }
      // 返回 resolved promise 确保调用者可以继续
      return Promise.resolve()
    }
  }
})

