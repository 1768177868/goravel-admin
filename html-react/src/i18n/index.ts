import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import Storage from '@/utils/storage'
import zhCN from './locales/zh-CN.json'
import enUS from './locales/en-US.json'

const savedLanguage = Storage.getItem<string>('language', 'zh-CN') || 'zh-CN'

void i18n.use(initReactI18next).init({
  resources: {
    'zh-CN': { translation: zhCN },
    'en-US': { translation: enUS },
  },
  lng: savedLanguage,
  fallbackLng: 'zh-CN',
  interpolation: {
    escapeValue: false,
  },
})

export function setLanguage(lang: 'zh-CN' | 'en-US') {
  Storage.setItem('language', lang)
  void i18n.changeLanguage(lang)
}

export default i18n
