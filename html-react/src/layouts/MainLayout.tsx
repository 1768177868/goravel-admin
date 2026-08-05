import { Layout, theme } from 'antd'
import { Outlet, useLocation, useMatches } from 'react-router-dom'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import LayoutHeader from './components/LayoutHeader'
import LayoutSidebar from './components/LayoutSidebar'
import TabsView from './components/TabsView'
import { useAppStore } from '@/stores/app'
import { useTabsStore } from '@/stores/tabs'
import type { AppRouteMeta } from '@/router/dynamicRoutes'
import './MainLayout.scss'

const { Content } = Layout

export default function MainLayout() {
  const { t } = useTranslation()
  const location = useLocation()
  const matches = useMatches()
  const collapsed = useAppStore((s) => s.sidebarCollapsed)
  const darkMode = useAppStore((s) => s.darkMode)
  const addTab = useTabsStore((s) => s.addTab)
  const { token } = theme.useToken()

  useEffect(() => {
    const match = [...matches].reverse().find((m) => (m.handle as AppRouteMeta | undefined)?.titleKey)
    const handle = match?.handle as AppRouteMeta | undefined
    const titleKey = handle?.titleKey
    addTab({
      path: location.pathname,
      title: titleKey ? t(titleKey) : location.pathname,
      titleKey,
      name: String(match?.id || location.pathname),
    })
  }, [location.pathname, matches, addTab, t])

  return (
    <Layout className={`main-layout ${darkMode ? 'main-layout--dark' : ''}`}>
      <LayoutSidebar />
      <Layout className="main-layout__body" style={{ marginLeft: collapsed ? 64 : 220 }}>
        <LayoutHeader />
        <TabsView />
        <Content
          className="main-layout__content"
          style={{ background: token.colorBgLayout }}
        >
          <div className="main-layout__page">
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}
