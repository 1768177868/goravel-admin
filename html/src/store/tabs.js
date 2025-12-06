import { defineStore } from 'pinia'
import Storage from '../utils/storage'

// 从 localStorage 加载标签页数据
const loadTabsFromStorage = () => {
  try {
    const data = Storage.getItem('tabs', null)
    if (data && typeof data === 'object') {
      return {
        tabs: data.tabs || [],
        activeTab: data.activeTab || null
      }
    }
  } catch (error) {
    console.error('Failed to load tabs from storage:', error)
  }
  return {
    tabs: [],
    activeTab: null
  }
}

// 保存标签页数据到 localStorage
const saveTabsToStorage = (tabs, activeTab) => {
  try {
    const data = { tabs, activeTab }
    Storage.setItem('tabs', data)
    // 触发 storage 事件，通知其他标签页更新
    window.dispatchEvent(new StorageEvent('storage', {
      key: 'tabs',
      newValue: JSON.stringify(data)
    }))
  } catch (error) {
    console.error('Failed to save tabs to storage:', error)
  }
}

export const useTabsStore = defineStore('tabs', {
  state: () => {
    const { tabs, activeTab } = loadTabsFromStorage()
    return {
      tabs,
      activeTab
    }
  },

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
      saveTabsToStorage(this.tabs, this.activeTab)
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
      saveTabsToStorage(this.tabs, this.activeTab)
    },

    removeOtherTabs(path) {
      this.tabs = this.tabs.filter(t => t.path === path)
      this.activeTab = path
      saveTabsToStorage(this.tabs, this.activeTab)
    },

    removeLeftTabs(path) {
      const index = this.tabs.findIndex(t => t.path === path)
      if (index > -1) {
        this.tabs = this.tabs.slice(index)
        this.activeTab = path
      }
      saveTabsToStorage(this.tabs, this.activeTab)
    },

    removeRightTabs(path) {
      const index = this.tabs.findIndex(t => t.path === path)
      if (index > -1) {
        this.tabs = this.tabs.slice(0, index + 1)
        this.activeTab = path
      }
      saveTabsToStorage(this.tabs, this.activeTab)
    },

    removeAllTabs() {
      this.tabs = []
      this.activeTab = null
      saveTabsToStorage(this.tabs, this.activeTab)
    },

    refreshTab(path) {
      const tab = this.tabs.find(t => t.path === path)
      if (tab) {
        tab.refreshKey = Date.now()
        saveTabsToStorage(this.tabs, this.activeTab)
      }
    },

    getRefreshKey(path) {
      const tab = this.tabs.find(t => t.path === path)
      return tab?.refreshKey || ''
    },

    setActiveTab(path) {
      this.activeTab = path
      saveTabsToStorage(this.tabs, this.activeTab)
    },

    // 从 localStorage 同步标签页（用于多标签页同步）
    syncTabsFromStorage() {
      const { tabs, activeTab } = loadTabsFromStorage()
      this.tabs = tabs
      this.activeTab = activeTab
    }
  }
})

// 监听 storage 事件，实现多标签页同步
// 注意：storage 事件只在其他标签页修改 localStorage 时触发，不会在当前标签页触发
// 这个监听器会在 store 初始化后设置
let storageListener = null

export const setupTabsStorageSync = () => {
  if (typeof window === 'undefined' || storageListener) {
    return
  }
  
  storageListener = (e) => {
    if (e.key === 'tabs' && e.newValue) {
      try {
        const data = JSON.parse(e.newValue)
        const tabsStore = useTabsStore()
        if (tabsStore) {
          tabsStore.tabs = data.tabs || []
          tabsStore.activeTab = data.activeTab || null
        }
      } catch (error) {
        console.error('Failed to sync tabs from storage:', error)
      }
    }
  }
  
  window.addEventListener('storage', storageListener)
}

// 在浏览器环境中自动设置监听器
if (typeof window !== 'undefined') {
  // 延迟设置，确保 Pinia 已经初始化
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', setupTabsStorageSync)
  } else {
    // DOM 已经加载完成，延迟一下确保 Pinia 初始化
    setTimeout(setupTabsStorageSync, 100)
  }
}

