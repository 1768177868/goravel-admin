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
import './style.css'

const app = createApp(App)

// 注册 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 根据当前语言设置 Element Plus 语言
const getElementLocale = () => {
  const savedLocale = localStorage.getItem('language') || 'zh-CN'
  return savedLocale === 'zh-CN' ? zhCn : en
}

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(ElementPlus, { locale: getElementLocale() })
app.use(VXETable)
app.use(VxePcUI)

// 初始化布局大小
const layoutSize = localStorage.getItem('layoutSize') || 'default'
document.body.classList.add(`layout-${layoutSize}`)

app.mount('#app')

