import { useEffect } from 'react'
import { App as AntdApp, ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import enUS from 'antd/locale/en_US'
import { useTranslation } from 'react-i18next'
import { AppRouter } from '@/router'
import { useAppStore } from '@/stores/app'
import { setupTabsStorageSync } from '@/stores/tabs'

export default function App() {
  const { i18n } = useTranslation()
  const darkMode = useAppStore((s) => s.darkMode)
  const themeColor = useAppStore((s) => s.themeColor)
  const layoutSize = useAppStore((s) => s.layoutSize)
  const getAntdTheme = useAppStore((s) => s.getAntdTheme)
  const initAppearance = useAppStore((s) => s.initAppearance)

  useEffect(() => {
    initAppearance()
    setupTabsStorageSync()
  }, [initAppearance])

  // re-subscribe theme deps
  void darkMode
  void themeColor
  void layoutSize

  return (
    <ConfigProvider
      locale={i18n.language === 'en-US' ? enUS : zhCN}
      theme={getAntdTheme()}
    >
      <AntdApp>
        <AppRouter />
      </AntdApp>
    </ConfigProvider>
  )
}
