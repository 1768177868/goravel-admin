import { create } from 'zustand'
import Storage from '@/utils/storage'

export interface TabItem {
  name?: string
  path: string
  title: string
  titleKey?: string
  refreshKey?: number
}

interface TabsSnapshot {
  tabs: TabItem[]
  activeTab: string | null
}

function loadTabsFromStorage(): TabsSnapshot {
  try {
    const data = Storage.getItem<TabsSnapshot>('tabs', null)
    if (data && typeof data === 'object') {
      return {
        tabs: data.tabs || [],
        activeTab: data.activeTab || null,
      }
    }
  } catch (error) {
    console.error('Failed to load tabs from storage:', error)
  }
  return { tabs: [], activeTab: null }
}

function saveTabsToStorage(tabs: TabItem[], activeTab: string | null) {
  try {
    const data = { tabs, activeTab }
    Storage.setItem('tabs', data)
    window.dispatchEvent(
      new StorageEvent('storage', {
        key: 'tabs',
        newValue: JSON.stringify(data),
      }),
    )
  } catch (error) {
    console.error('Failed to save tabs to storage:', error)
  }
}

interface TabsState {
  tabs: TabItem[]
  activeTab: string | null
  addTab: (tab: Omit<TabItem, 'refreshKey'>) => void
  removeTab: (path: string) => string | null
  removeOtherTabs: (path: string) => void
  removeLeftTabs: (path: string) => void
  removeRightTabs: (path: string) => void
  removeAllTabs: () => void
  refreshTab: (path: string) => void
  getRefreshKey: (path: string) => number | string
  setActiveTab: (path: string) => void
  syncTabsFromStorage: () => void
}

const initial = loadTabsFromStorage()

export const useTabsStore = create<TabsState>((set, get) => ({
  tabs: initial.tabs,
  activeTab: initial.activeTab,

  addTab: (route) => {
    const tab: TabItem = {
      name: route.name,
      path: route.path,
      title: route.titleKey || route.title || route.name || route.path,
      titleKey: route.titleKey,
    }
    const exists = get().tabs.find((t) => t.path === tab.path)
    const tabs = exists ? get().tabs : [...get().tabs, tab]
    set({ tabs, activeTab: tab.path })
    saveTabsToStorage(tabs, tab.path)
  },

  removeTab: (path) => {
    const { tabs, activeTab } = get()
    const index = tabs.findIndex((t) => t.path === path)
    if (index < 0) return activeTab

    const nextTabs = tabs.filter((t) => t.path !== path)
    let nextActive = activeTab

    if (activeTab === path) {
      if (nextTabs.length > 0) {
        const nextIndex = index < nextTabs.length ? index : index - 1
        nextActive = nextTabs[Math.max(0, nextIndex)]?.path ?? null
      } else {
        nextActive = null
      }
    }

    set({ tabs: nextTabs, activeTab: nextActive })
    saveTabsToStorage(nextTabs, nextActive)
    return nextActive
  },

  removeOtherTabs: (path) => {
    const tabs = get().tabs.filter((t) => t.path === path)
    set({ tabs, activeTab: path })
    saveTabsToStorage(tabs, path)
  },

  removeLeftTabs: (path) => {
    const index = get().tabs.findIndex((t) => t.path === path)
    if (index < 0) return
    const tabs = get().tabs.slice(index)
    set({ tabs, activeTab: path })
    saveTabsToStorage(tabs, path)
  },

  removeRightTabs: (path) => {
    const index = get().tabs.findIndex((t) => t.path === path)
    if (index < 0) return
    const tabs = get().tabs.slice(0, index + 1)
    set({ tabs, activeTab: path })
    saveTabsToStorage(tabs, path)
  },

  removeAllTabs: () => {
    set({ tabs: [], activeTab: null })
    saveTabsToStorage([], null)
  },

  refreshTab: (path) => {
    const tabs = get().tabs.map((tab) =>
      tab.path === path ? { ...tab, refreshKey: Date.now() } : tab,
    )
    set({ tabs })
    saveTabsToStorage(tabs, get().activeTab)
  },

  getRefreshKey: (path) => get().tabs.find((t) => t.path === path)?.refreshKey || '',

  setActiveTab: (path) => {
    set({ activeTab: path })
    saveTabsToStorage(get().tabs, path)
  },

  syncTabsFromStorage: () => {
    const snapshot = loadTabsFromStorage()
    set({ tabs: snapshot.tabs, activeTab: snapshot.activeTab })
  },
}))

export function setupTabsStorageSync() {
  if (typeof window === 'undefined') return

  window.addEventListener('storage', (e) => {
    if (e.key === 'tabs' && e.newValue) {
      try {
        const data = JSON.parse(e.newValue) as TabsSnapshot
        useTabsStore.setState({
          tabs: data.tabs || [],
          activeTab: data.activeTab || null,
        })
      } catch (error) {
        console.error('Failed to sync tabs from storage:', error)
      }
    }
  })
}
