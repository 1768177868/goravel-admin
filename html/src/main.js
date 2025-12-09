import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import VXETable from 'vxe-table'
import 'vxe-table/lib/style.css'
import VxePcUI from 'vxe-pc-ui'
import 'vxe-pc-ui/lib/style.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import en from 'element-plus/dist/locale/en.mjs'

import App from './App.vue'
import router from './router'
import i18n from './i18n'
import { setupTabsStorageSync } from './store/tabs'
import { validateEnv } from './utils/env'
import Storage from './utils/storage'
import logger from './utils/logger'
import './style.css'

// 验证环境变量

try {
  validateEnv(false) // 非严格模式，只警告
} catch (error) {
  logger.error('Environment validation failed:', error)
}

// 检查 localStorage 是否可用
if (!Storage.isAvailable()) {
  logger.warn('localStorage is not available. Some features may not work properly.')
}

const app = createApp(App)

// 注册 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 根据当前语言设置 Element Plus 语言
const getElementLocale = () => {
  const savedLocale = Storage.getItem('language', 'zh-CN')
  return savedLocale === 'zh-CN' ? zhCn : en
}

const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(i18n)
app.use(ElementPlus, { locale: getElementLocale() })
app.use(VXETable)
app.use(VxePcUI)

// 初始化布局大小
const layoutSize = Storage.getItem('layoutSize', 'default')
document.body.classList.add(`layout-${layoutSize}`)

// 设置多标签页同步监听器（在 Pinia 初始化后）
setupTabsStorageSync()

app.mount('#app')

