import { defineStore } from 'pinia'
import { getInfo, logout } from '../api/auth'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    adminInfo: JSON.parse(localStorage.getItem('adminInfo') || 'null'),
    permissions: [],
    menus: []
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
      this.adminInfo = adminInfo
      localStorage.setItem('adminInfo', JSON.stringify(adminInfo))
    },

    setPermissions(permissions) {
      this.permissions = permissions
    },

    setMenus(menus) {
      this.menus = menus
    },

    async fetchUserInfo() {
      try {
        const res = await getInfo()
        if (res.data && res.data.admin) {
          this.setAdminInfo(res.data.admin)
          this.setPermissions(res.data.admin.permissions || [])
          this.setMenus(res.data.admin.menus || [])
        }
        return res
      } catch (error) {
        this.logout()
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
        // 清除所有状态
        this.token = ''
        this.adminInfo = null
        this.permissions = []
        this.menus = []
        localStorage.removeItem('token')
        localStorage.removeItem('adminInfo')
      }
    }
  }
})

