import { defineStore } from 'pinia'
import Storage from '../utils/storage'
import { VxeUI } from 'vxe-table'

const detectBrowserTimezone = () => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  } catch {
    return 'UTC'
  }
}

// 主题色预设：key 存 localStorage，color 用于设置 CSS 变量
export const THEME_COLORS = [
  { key: 'blue', color: '#409EFF' },
  { key: 'green', color: '#67C23A' },
  { key: 'orange', color: '#E6A23C' },
  { key: 'red', color: '#F56C6C' },
  { key: 'purple', color: '#9C27B0' },
  { key: 'cyan', color: '#00BCD4' }
]

const DEFAULT_THEME_KEY = 'blue'

/** 将 hex 转为 [r, g, b] */
function hexToRgb(hex) {
  const m = hex.match(/^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i)
  if (!m) return [64, 158, 255]
  return [parseInt(m[1], 16), parseInt(m[2], 16), parseInt(m[3], 16)]
}

/** 混合两种 RGB，weight 为 color2 的占比 0–100，返回 #rrggbb */
function mixRgb(rgb1, rgb2, weight) {
  const w = weight / 100
  const r = Math.round(rgb1[0] * (1 - w) + rgb2[0] * w)
  const g = Math.round(rgb1[1] * (1 - w) + rgb2[1] * w)
  const b = Math.round(rgb1[2] * (1 - w) + rgb2[2] * w)
  return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`
}

const WHITE = [255, 255, 255]
const BLACK = [0, 0, 0]

export const useAppStore = defineStore('app', {
  state: () => ({
    sidebarCollapsed: Storage.getItem('sidebarCollapsed', 'false') === 'true',
    layoutSize: Storage.getItem('layoutSize', 'default') || 'default', // default, large, small
    isFullscreen: false,
    timezone: Storage.getItem('timezone', detectBrowserTimezone()) || detectBrowserTimezone(),
    darkMode: Storage.getItem('darkMode', 'false') === 'true',
    // 导航模式: sidebar 左侧菜单, top 顶部菜单
    menuMode: Storage.getItem('menuMode', 'sidebar') || 'sidebar',
    // 是否开启水印
    watermarkEnabled: Storage.getItem('watermarkEnabled', 'false') === 'true',
    // 主题色 key，对应 THEME_COLORS[].key
    themeColor: Storage.getItem('themeColor', DEFAULT_THEME_KEY) || DEFAULT_THEME_KEY
  }),

  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      Storage.setItem('sidebarCollapsed', this.sidebarCollapsed.toString())
    },

    setSidebarCollapsed(collapsed) {
      this.sidebarCollapsed = collapsed
      Storage.setItem('sidebarCollapsed', collapsed.toString())
    },

    setLayoutSize(size) {
      this.layoutSize = size
      Storage.setItem('layoutSize', size)
      // 应用布局大小到 body
      document.body.className = document.body.className.replace(/layout-\w+/g, '')
      document.body.classList.add(`layout-${size}`)
      
      // 映射 size 值到 Element Plus: default -> default, large -> large, small -> small
      // Element Plus 支持 'large', 'default', 'small'
      const elementSize = size === 'default' ? 'default' : size
      
      // 映射 size 值到 VXE Table: default -> medium, large -> '' (空字符串，最大), small -> small
      // VXE Table 支持 'medium', 'small', 'mini', '' (空字符串表示默认/最大)
      let vxeSize = ''
      if (size === 'small') {
        vxeSize = 'small'
      } else if (size === 'default') {
        vxeSize = 'medium'
      } else if (size === 'large') {
        vxeSize = '' // 空字符串表示默认/最大尺寸
      }
      
      // 设置 VXE Table 大小
      try {
        if (VxeUI && typeof VxeUI.setSize === 'function') {
          VxeUI.setSize(vxeSize)
        }
      } catch (error) {
        console.error('Failed to set VXE Table size:', error)
      }
    },

    toggleFullscreen() {
      if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen().then(() => {
          this.isFullscreen = true
        }).catch(() => {
          console.error('无法进入全屏模式')
        })
      } else {
        document.exitFullscreen().then(() => {
          this.isFullscreen = false
        }).catch(() => {
          console.error('无法退出全屏模式')
        })
      }
    },

    setTimezone(timezone) {
      this.timezone = timezone || 'UTC'
      Storage.setItem('timezone', this.timezone)
    },

    toggleDarkMode() {
      this.darkMode = !this.darkMode
      Storage.setItem('darkMode', this.darkMode.toString())
      // 使用 Element Plus 官方暗黑模式方式：在 html 元素上添加/移除 dark 类
      if (this.darkMode) {
        document.documentElement.classList.add('dark')
        VxeUI.setTheme('dark')
      } else {
        document.documentElement.classList.remove('dark')
        VxeUI.setTheme('light')
      }
    },

    initDarkMode() {
      // 初始化时应用夜间模式
      if (this.darkMode) {
        document.documentElement.classList.add('dark')
        VxeUI.setTheme('dark')
      } else {
        document.documentElement.classList.remove('dark')
        VxeUI.setTheme('light')
      }
    },

    setMenuMode(mode) {
      this.menuMode = mode === 'top' ? 'top' : 'sidebar'
      Storage.setItem('menuMode', this.menuMode)
    },

    setWatermarkEnabled(enabled) {
      this.watermarkEnabled = !!enabled
      Storage.setItem('watermarkEnabled', this.watermarkEnabled.toString())
    },

    toggleWatermark() {
      this.watermarkEnabled = !this.watermarkEnabled
      Storage.setItem('watermarkEnabled', this.watermarkEnabled.toString())
    },

    /** 将当前 themeColor 对应的色值应用到根节点 CSS 变量（含 hover 用的 light/dark 变体） */
    applyThemeColor() {
      const preset = THEME_COLORS.find((t) => t.key === this.themeColor) || THEME_COLORS[0]
      const color = preset.color
      const rgb = hexToRgb(color)
      const root = document.documentElement
      root.style.setProperty('--el-color-primary', color)
      root.style.setProperty('--el-menu-active-color', color)
      root.style.setProperty('--sidebar-active', color)
      // Element Plus hover 使用 --el-color-primary-light-*，需一并设置（light-i = i*10% 白色 + base）
      for (let i = 1; i <= 9; i++) {
        const mixed = mixRgb(rgb, WHITE, i * 10)
        root.style.setProperty(`--el-color-primary-light-${i}`, mixed)
      }
      // dark-2 = 80% 主色 + 20% 黑（按下时略深）
      root.style.setProperty('--el-color-primary-dark-2', mixRgb(rgb, BLACK, 20))
    },

    setThemeColor(key) {
      const preset = THEME_COLORS.find((t) => t.key === key)
      if (!preset) return
      this.themeColor = key
      Storage.setItem('themeColor', key)
      this.applyThemeColor()
    },

    initThemeColor() {
      this.applyThemeColor()
    }
  }
})

