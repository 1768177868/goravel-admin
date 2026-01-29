import { createApp, watch, nextTick } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// 导入 Element Plus 暗黑模式样式
import 'element-plus/theme-chalk/dark/css-vars.css'
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
import Storage from './utils/storage'
import { setupTabsStorageSync } from './store/tabs'
import { validateEnv } from './utils/env'
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

// 配置 vxe-table 国际化
const setupVxeTableI18n = () => {
  const vxeI18nMap = {
    'zh-CN': {
      'vxe.pager.goto': '前往',
      'vxe.pager.pagesize': '{size} 条/页',
      'vxe.pager.total': '共 {total} 条记录',
      'vxe.pager.pageClassifier': '页',
      'vxe.table.emptyText': '暂无数据',
      'vxe.loading.text': '加载中...',
      'vxe.toolbar.customAll': '全选',
      'vxe.toolbar.customCancel': '取消',
      'vxe.toolbar.customConfirm': '确认',
      'vxe.toolbar.customReset': '重置',
      'vxe.table.customRestore': '恢复默认',
      'vxe.table.customCancel': '取消',
      'vxe.table.customConfirm': '确认',
      'vxe.table.customReset': '重置',
      'vxe.toolbar.refresh': '刷新',
      'vxe.toolbar.zoom': '全屏',
      'vxe.toolbar.custom': '列设置',
      'vxe.table.fixedLeft': '冻结列左侧',
      'vxe.table.fixedRight': '冻结列右侧',
      'vxe.table.unfixed': '取消冻结列',
      'vxe.table.fixedLeftTitle': '冻结列左侧',
      'vxe.table.fixedRightTitle': '冻结列右侧',
      'vxe.table.unfixedTitle': '取消冻结列',
      'vxe.toolbar.fixedLeft': '冻结列左侧',
      'vxe.toolbar.fixedRight': '冻结列右侧',
      'vxe.toolbar.cancelFixed': '取消冻结列',
      'vxe.custom.setting.colVisible': '显示/隐藏列',
      'vxe.custom.setting.sortHelpTip': '拖拽调整列顺序',
      'pager.goto': '前往',
      'pager.pagesize': '{size} 条/页',
      'pager.total': '共 {total} 条记录',
      'pager.pageClassifier': '页',
      'table.emptyText': '暂无数据',
      'loading.text': '加载中...',
      'toolbar.customAll': '全选',
      'toolbar.customCancel': '取消',
      'toolbar.customConfirm': '确认',
      'toolbar.customReset': '重置',
      'table.customRestore': '恢复默认',
      'table.customCancel': '取消',
      'table.customConfirm': '确认',
      'table.customReset': '重置',
      'toolbar.refresh': '刷新',
      'toolbar.zoom': '全屏',
      'toolbar.custom': '列设置',
      'table.fixedLeft': '冻结列左侧',
      'table.fixedRight': '冻结列右侧',
      'table.unfixed': '取消冻结列',
      'table.fixedLeftTitle': '冻结列左侧',
      'table.fixedRightTitle': '冻结列右侧',
      'table.unfixedTitle': '取消冻结列',
      'toolbar.fixedLeft': '冻结列左侧',
      'toolbar.fixedRight': '冻结列右侧',
      'toolbar.cancelFixed': '取消冻结列',
      'custom.setting.colVisible': '显示/隐藏列',
      'custom.setting.sortHelpTip': '拖拽调整列顺序'
    },
    'en-US': {
      'vxe.pager.goto': 'Go to',
      'vxe.pager.pagesize': '{size} records/page',
      'vxe.pager.total': 'Total {total} records',
      'vxe.pager.pageClassifier': 'page',
      'vxe.table.emptyText': 'No Data',
      'vxe.loading.text': 'Loading...',
      'vxe.toolbar.customAll': 'Select All',
      'vxe.toolbar.customCancel': 'Cancel',
      'vxe.toolbar.customConfirm': 'Confirm',
      'vxe.toolbar.customReset': 'Reset',
      'vxe.table.customRestore': 'Restore Default',
      'vxe.table.customCancel': 'Cancel',
      'vxe.table.customConfirm': 'Confirm',
      'vxe.table.customReset': 'Reset',
      'vxe.toolbar.refresh': 'Refresh',
      'vxe.toolbar.zoom': 'Fullscreen',
      'vxe.toolbar.custom': 'Column Setting',
      'vxe.table.fixedLeft': 'Freeze Column Left',
      'vxe.table.fixedRight': 'Freeze Column Right',
      'vxe.table.unfixed': 'Unfreeze Column',
      'vxe.table.fixedLeftTitle': 'Freeze Column Left',
      'vxe.table.fixedRightTitle': 'Freeze Column Right',
      'vxe.table.unfixedTitle': 'Unfreeze Column',
      'vxe.toolbar.fixedLeft': 'Freeze Column Left',
      'vxe.toolbar.fixedRight': 'Freeze Column Right',
      'vxe.toolbar.cancelFixed': 'Unfreeze Column',
      'vxe.custom.setting.colVisible': 'Show/Hide Column',
      'vxe.custom.setting.sortHelpTip': 'Drag to reorder columns',
      'pager.goto': 'Go to',
      'pager.pagesize': '{size} records/page',
      'pager.total': 'Total {total} records',
      'pager.pageClassifier': 'page',
      'table.emptyText': 'No Data',
      'loading.text': 'Loading...',
      'toolbar.customAll': 'Select All',
      'toolbar.customCancel': 'Cancel',
      'toolbar.customConfirm': 'Confirm',
      'toolbar.customReset': 'Reset',
      'table.customRestore': 'Restore Default',
      'table.customCancel': 'Cancel',
      'table.customConfirm': 'Confirm',
      'table.customReset': 'Reset',
      'toolbar.refresh': 'Refresh',
      'toolbar.zoom': 'Fullscreen',
      'toolbar.custom': 'Column Setting',
      'table.fixedLeft': 'Freeze Column Left',
      'table.fixedRight': 'Freeze Column Right',
      'table.unfixed': 'Unfreeze Column',
      'table.fixedLeftTitle': 'Freeze Column Left',
      'table.fixedRightTitle': 'Freeze Column Right',
      'table.unfixedTitle': 'Unfreeze Column',
      'toolbar.fixedLeft': 'Freeze Column Left',
      'toolbar.fixedRight': 'Freeze Column Right',
      'toolbar.cancelFixed': 'Unfreeze Column',
      'custom.setting.colVisible': 'Show/Hide Column',
      'custom.setting.sortHelpTip': 'Drag to reorder columns'
    }
  }
  
  VXETable.setup({
    i18n: (key, args) => {
      // 动态获取当前语言，而不是使用闭包变量
      const currentLocale = i18n.global.locale.value || Storage.getItem('language', 'zh-CN') || 'zh-CN'
      
      // 尝试直接匹配
      let value = vxeI18nMap[currentLocale]?.[key]
      
      // 如果没有找到，尝试去掉 vxe. 前缀
      if (!value && key.startsWith('vxe.')) {
        value = vxeI18nMap[currentLocale]?.[key.substring(4)]
      }
      
      // 如果还是没有找到，尝试添加 vxe. 前缀
      if (!value && !key.startsWith('vxe.')) {
        value = vxeI18nMap[currentLocale]?.[`vxe.${key}`]
      }
      
      // 特殊处理：vxe-table 可能使用 table.emptyText 或 loading.text 格式
      if (!value) {
        if (key === 'table.emptyText' || key === 'emptyText') {
          value = vxeI18nMap[currentLocale]?.['vxe.table.emptyText'] || vxeI18nMap[currentLocale]?.['table.emptyText']
        } else if (key === 'loading.text' || key === 'loading') {
          value = vxeI18nMap[currentLocale]?.['vxe.loading.text'] || vxeI18nMap[currentLocale]?.['loading.text']
        } else if (key.startsWith('toolbar.')) {
          // 处理 toolbar.* 格式的键
          value = vxeI18nMap[currentLocale]?.[`vxe.${key}`] || vxeI18nMap[currentLocale]?.[key]
        } else if (key.startsWith('table.custom')) {
          // 处理 table.custom* 格式的键
          value = vxeI18nMap[currentLocale]?.[`vxe.${key}`] || vxeI18nMap[currentLocale]?.[key]
        }
      }
      
      // 调试日志（开发环境）- 扩展调试范围
      if (process.env.NODE_ENV === 'development' && (key.includes('empty') || key.includes('loading'))) {
        // console.log('[VXE i18n]', { key, currentLocale, value })
      }
      
      // 如果找到值且有参数，替换参数
      if (value && args !== undefined && args !== null) {
        // 处理 args 可能是对象、数组或其他类型的情况
        let params = {}
        
        if (Array.isArray(args)) {
          // 如果是数组，尝试从数组中提取参数
          if (args.length > 0 && typeof args[0] === 'object') {
            params = args[0]
          } else {
            // 如果数组元素不是对象，可能是按位置传递的参数
            params = { total: args[0], pageSize: args[1] }
          }
        } else if (typeof args === 'object') {
          params = args
        } else {
          // 如果是单个值，可能是 total
          params = { total: args }
        }
        
        // 调试日志（开发环境）
        if (process.env.NODE_ENV === 'development' && (key.includes('total') || key.includes('pagesize'))) {
          // console.log('[VXE i18n]', { key, args, params, value })
        }
        
        // 替换所有参数占位符
        let result = value
        for (const paramKey in params) {
          const paramValue = params[paramKey]
          if (paramValue !== undefined && paramValue !== null) {
            // 支持多种占位符格式：{key}、${key}、$key
            const regex1 = new RegExp(`\\{${paramKey}\\}`, 'g')
            const regex2 = new RegExp(`\\$\\{${paramKey}\\}`, 'g')
            const regex3 = new RegExp(`\\$${paramKey}\\b`, 'g')
            result = result.replace(regex1, String(paramValue))
            result = result.replace(regex2, String(paramValue))
            result = result.replace(regex3, String(paramValue))
          }
        }
        
        // 特殊处理：如果 key 包含 pagesize 且没有 size 参数，尝试从 args 中提取
        if (key.includes('pagesize') && !params.size && args !== undefined && args !== null) {
          // 如果 args 是数字，直接使用
          if (typeof args === 'number') {
            result = result.replace(/\{size\}/g, String(args))
          }
          // 如果 args 是数组且第一个元素是数字
          else if (Array.isArray(args) && args.length > 0 && typeof args[0] === 'number') {
            result = result.replace(/\{size\}/g, String(args[0]))
          }
          // 如果 args 是对象但没有 size 属性，尝试其他可能的属性名
          else if (typeof args === 'object' && !Array.isArray(args)) {
            const sizeValue = args.size || args.pageSize || args.pagesize || args.value
            if (sizeValue !== undefined && sizeValue !== null) {
              result = result.replace(/\{size\}/g, String(sizeValue))
            }
          }
        }
        
        return result
      }
      
      return value || key
    }
  })
}

const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(i18n)
app.use(ElementPlus, { locale: getElementLocale() })

// 初始化 vxe-table 国际化（必须在 app.use(VXETable) 之前调用）
const currentLocale = Storage.getItem('language', 'zh-CN') || 'zh-CN'
i18n.global.locale.value = currentLocale
setupVxeTableI18n()

app.use(VXETable)
app.use(VxePcUI)

// 监听语言变化，重新设置 vxe-table 国际化
watch(() => i18n.global.locale.value, (newLocale) => {
  // 重新设置 vxe-table 国际化配置
  setupVxeTableI18n()
  
  // 强制刷新所有 vxe-table 实例（通过触发全局事件）
  // 注意：vxe-table 可能不会自动刷新，我们需要手动触发
  nextTick(() => {
    // 触发一个自定义事件，让所有使用 vxe-table 的组件知道语言已更改
    window.dispatchEvent(new CustomEvent('vxe-i18n-updated', { detail: { locale: newLocale } }))
  })
})

// 初始化布局大小
const layoutSize = Storage.getItem('layoutSize', 'default')
document.body.classList.add(`layout-${layoutSize}`)

// 设置多标签页同步监听器（在 Pinia 初始化后）
setupTabsStorageSync()

// 导入错误上报器
import { reportComponentError, reportUnhandledRejection } from './utils/errorReporter'

// 全局错误处理：静默处理 Element Plus TabPane 的已知卸载错误
app.config.errorHandler = (err, instance, info) => {
  const errorMessage = err?.message || ''
  const errorStack = err?.stack || ''
  const instanceName = instance?.$?.type?.name || instance?.$.type?.__name || ''
  
  // 静默处理 Element Plus TabPane 卸载时的已知错误
  const isTabPaneError = (
    errorMessage.includes('indexOf') ||
    errorStack.includes('unregisterPane') ||
    errorStack.includes('removeChild') ||
    (instanceName === 'ElTabPane' && info === 'beforeUnmount hook')
  )
  
  if (isTabPaneError) {
    // 这是 Element Plus 的已知问题，不影响功能，静默处理
    return
  }
  
  // 上报组件错误
  reportComponentError(err, instanceName, info)
}

// 全局未捕获错误处理
window.addEventListener('error', (event) => {
  const errorMessage = event.message || ''
  const errorStack = event.error?.stack || ''
  
  // 静默处理 Element Plus TabPane 卸载时的已知错误
  if (errorMessage.includes('indexOf') && (errorStack.includes('unregisterPane') || errorStack.includes('element-plus'))) {
    event.preventDefault()
    return
  }
})

// 全局未处理的 Promise 错误
window.addEventListener('unhandledrejection', (event) => {
  const errorMessage = event.reason?.message || ''
  const errorStack = event.reason?.stack || ''
  
  // 静默处理 Element Plus TabPane 卸载时的已知错误
  if (errorMessage.includes('indexOf') && (errorStack.includes('unregisterPane') || errorStack.includes('element-plus'))) {
    event.preventDefault()
    return
  }
  
  // 上报未处理的 Promise 错误
  reportUnhandledRejection(event)
})

app.mount('#app')

