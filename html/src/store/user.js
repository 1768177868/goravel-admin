import { defineStore } from 'pinia'
import { getInfo, logout } from '../api/auth'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    adminInfo: JSON.parse(localStorage.getItem('adminInfo') || 'null'),
    permissions: [],
    menus: JSON.parse(localStorage.getItem('menus') || '[]')
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    hasPermission: (state) => (permission) => {
      return state.permissions.includes(permission)
    }
  },

  actions: {
    setToken(token) {
      this.token = token
      localStorage.setItem('token', token)
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
      try {
        localStorage.setItem('adminInfo', JSON.stringify(basicInfo))
      } catch (error) {
        // 如果仍然超出配额，尝试清除其他可能的缓存
        console.warn('Failed to save adminInfo to localStorage:', error)
        // 尝试只存储最基本信息
        const minimalInfo = {
          id: basicInfo.id,
          username: basicInfo.username,
          nickname: basicInfo.nickname,
          avatar: basicInfo.avatar
        }
        try {
          localStorage.setItem('adminInfo', JSON.stringify(minimalInfo))
          this.adminInfo = minimalInfo
        } catch (e) {
          console.error('Failed to save minimal adminInfo:', e)
        }
      }
    },

    setPermissions(permissions) {
      this.permissions = permissions
    },

    setMenus(menus) {
      this.menus = menus
      localStorage.setItem('menus', JSON.stringify(menus || []))
    },

    async fetchUserInfo() {
      try {
        // 清除旧的缓存，确保获取最新的数据
        this.menus = []
        this.adminInfo = null
        this.permissions = []
        localStorage.removeItem('menus')
        localStorage.removeItem('adminInfo')
        
        const res = await getInfo()
        if (res.data && res.data.admin) {
          this.setAdminInfo(res.data.admin)
          this.setPermissions(res.data.admin.permissions || [])
          this.setMenus(res.data.admin.menus || [])
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
        console.error('Logout error:', error)
      } finally {
        // 清除所有状态（同步执行，不等待）
        this.token = ''
        this.adminInfo = null
        this.permissions = []
        this.menus = []
        localStorage.removeItem('token')
        localStorage.removeItem('adminInfo')
        localStorage.removeItem('menus')
      }
      // 返回 resolved promise 确保调用者可以继续
      return Promise.resolve()
    }
  }
})

