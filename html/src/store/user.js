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

    async logout() {
      try {
        await logout()
      } catch (error) {
        console.error('Logout error:', error)
      } finally {
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

