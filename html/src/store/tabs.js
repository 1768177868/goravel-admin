import { defineStore } from 'pinia'

export const useTabsStore = defineStore('tabs', {
  state: () => ({
    tabs: [],
    activeTab: null
  }),

  getters: {
    hasTabs: (state) => state.tabs.length > 0
  },

  actions: {
    addTab(route) {
      const tab = {
        name: route.name,
        path: route.path,
        title: route.meta?.titleKey || route.meta?.title || route.name,
        titleKey: route.meta?.titleKey
      }

      // 检查是否已存在
      const exists = this.tabs.find(t => t.path === tab.path)
      if (!exists) {
        this.tabs.push(tab)
      }

      this.activeTab = tab.path
    },

    removeTab(path) {
      const index = this.tabs.findIndex(t => t.path === path)
      if (index > -1) {
        this.tabs.splice(index, 1)
      }

      // 如果删除的是当前激活的标签，需要外部处理路由跳转
      // 这里不自动切换，由组件处理
      if (this.activeTab === path) {
        if (this.tabs.length > 0) {
          // 优先选择右侧标签，如果没有则选择左侧
          const nextIndex = index < this.tabs.length ? index : index - 1
          if (nextIndex >= 0 && nextIndex < this.tabs.length) {
            this.activeTab = this.tabs[nextIndex].path
          } else if (this.tabs.length > 0) {
            this.activeTab = this.tabs[this.tabs.length - 1].path
          } else {
            this.activeTab = null
          }
        } else {
          this.activeTab = null
        }
      }
    },

    removeOtherTabs(path) {
      this.tabs = this.tabs.filter(t => t.path === path)
      this.activeTab = path
    },

    removeLeftTabs(path) {
      const index = this.tabs.findIndex(t => t.path === path)
      if (index > -1) {
        this.tabs = this.tabs.slice(index)
        this.activeTab = path
      }
    },

    removeRightTabs(path) {
      const index = this.tabs.findIndex(t => t.path === path)
      if (index > -1) {
        this.tabs = this.tabs.slice(0, index + 1)
        this.activeTab = path
      }
    },

    removeAllTabs() {
      this.tabs = []
      this.activeTab = null
    },

    refreshTab(path) {
      const tab = this.tabs.find(t => t.path === path)
      if (tab) {
        tab.refreshKey = Date.now()
      }
    },

    getRefreshKey(path) {
      const tab = this.tabs.find(t => t.path === path)
      return tab?.refreshKey || ''
    },

    setActiveTab(path) {
      this.activeTab = path
    }
  }
})

