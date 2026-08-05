import { create } from 'zustand'
import { theme as antdTheme } from 'antd'
import type { ThemeConfig } from 'antd'
import Storage from '@/utils/storage'

export const THEME_COLORS = [
  { key: 'blue', color: '#1677ff' },
  { key: 'green', color: '#52c41a' },
  { key: 'orange', color: '#fa8c16' },
  { key: 'red', color: '#f5222d' },
  { key: 'purple', color: '#722ed1' },
  { key: 'cyan', color: '#13c2c2' },
] as const

export type ThemeColorKey = (typeof THEME_COLORS)[number]['key']
export type LayoutSize = 'default' | 'large' | 'small'
export type MenuMode = 'sidebar' | 'top'

function detectBrowserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  } catch {
    return 'UTC'
  }
}

interface AppState {
  sidebarCollapsed: boolean
  layoutSize: LayoutSize
  isFullscreen: boolean
  timezone: string
  darkMode: boolean
  menuMode: MenuMode
  watermarkEnabled: boolean
  themeColor: ThemeColorKey

  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setLayoutSize: (size: LayoutSize) => void
  toggleFullscreen: () => void
  setTimezone: (timezone: string) => void
  toggleDarkMode: () => void
  initAppearance: () => void
  setMenuMode: (mode: MenuMode) => void
  setWatermarkEnabled: (enabled: boolean) => void
  toggleWatermark: () => void
  setThemeColor: (key: ThemeColorKey) => void
  getAntdTheme: () => ThemeConfig
}

export const useAppStore = create<AppState>((set, get) => ({
  sidebarCollapsed: Storage.getItem<string>('sidebarCollapsed', 'false') === 'true',
  layoutSize: (Storage.getItem<LayoutSize>('layoutSize', 'default') || 'default') as LayoutSize,
  isFullscreen: false,
  timezone: Storage.getItem<string>('timezone', detectBrowserTimezone()) || detectBrowserTimezone(),
  darkMode: Storage.getItem<string>('darkMode', 'false') === 'true',
  menuMode: (Storage.getItem<MenuMode>('menuMode', 'sidebar') || 'sidebar') as MenuMode,
  watermarkEnabled: Storage.getItem<string>('watermarkEnabled', 'false') === 'true',
  themeColor: (Storage.getItem<ThemeColorKey>('themeColor', 'blue') || 'blue') as ThemeColorKey,

  toggleSidebar: () => {
    const next = !get().sidebarCollapsed
    set({ sidebarCollapsed: next })
    Storage.setItem('sidebarCollapsed', String(next))
  },

  setSidebarCollapsed: (collapsed) => {
    set({ sidebarCollapsed: collapsed })
    Storage.setItem('sidebarCollapsed', String(collapsed))
  },

  setLayoutSize: (size) => {
    set({ layoutSize: size })
    Storage.setItem('layoutSize', size)
    document.body.className = document.body.className.replace(/layout-\w+/g, '')
    document.body.classList.add(`layout-${size}`)
  },

  toggleFullscreen: () => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().then(() => set({ isFullscreen: true })).catch(() => {
        console.error('无法进入全屏模式')
      })
    } else {
      document.exitFullscreen().then(() => set({ isFullscreen: false })).catch(() => {
        console.error('无法退出全屏模式')
      })
    }
  },

  setTimezone: (timezone) => {
    const value = timezone || 'UTC'
    set({ timezone: value })
    Storage.setItem('timezone', value)
  },

  toggleDarkMode: () => {
    const next = !get().darkMode
    set({ darkMode: next })
    Storage.setItem('darkMode', String(next))
    document.documentElement.classList.toggle('dark', next)
  },

  initAppearance: () => {
    const { darkMode, layoutSize, themeColor } = get()
    document.documentElement.classList.toggle('dark', darkMode)
    document.body.classList.add(`layout-${layoutSize}`)
    const preset = THEME_COLORS.find((t) => t.key === themeColor) || THEME_COLORS[0]
    document.documentElement.style.setProperty('--app-color-primary', preset.color)
  },

  setMenuMode: (mode) => {
    const next = mode === 'top' ? 'top' : 'sidebar'
    set({ menuMode: next })
    Storage.setItem('menuMode', next)
  },

  setWatermarkEnabled: (enabled) => {
    set({ watermarkEnabled: !!enabled })
    Storage.setItem('watermarkEnabled', String(!!enabled))
  },

  toggleWatermark: () => {
    const next = !get().watermarkEnabled
    set({ watermarkEnabled: next })
    Storage.setItem('watermarkEnabled', String(next))
  },

  setThemeColor: (key) => {
    const preset = THEME_COLORS.find((t) => t.key === key)
    if (!preset) return
    set({ themeColor: key })
    Storage.setItem('themeColor', key)
    document.documentElement.style.setProperty('--app-color-primary', preset.color)
  },

  getAntdTheme: () => {
    const { darkMode, themeColor, layoutSize } = get()
    const preset = THEME_COLORS.find((t) => t.key === themeColor) || THEME_COLORS[0]
    const sizeMap = { large: 16, default: 14, small: 12 } as const

    return {
      algorithm: darkMode ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
      token: {
        colorPrimary: preset.color,
        fontSize: sizeMap[layoutSize],
        borderRadius: 6,
      },
    }
  },
}))
