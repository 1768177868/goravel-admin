import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN.json'
import enUS from './locales/en-US.json'

const messages = {
  'zh-CN': zhCN,
  'en-US': enUS
}

// 从 localStorage 获取语言设置，默认为中文
const locale = localStorage.getItem('language') || 'zh-CN'

const i18n = createI18n({
  legacy: false,
  locale,
  fallbackLocale: 'zh-CN',
  messages
})

export default i18n

