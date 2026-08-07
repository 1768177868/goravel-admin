import { Layout, Menu } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useAppStore } from '@/stores/app'
import { useMenuItems } from '@/hooks/useMenuItems'
import { useLayoutWebsite } from '@/hooks/useLayoutWebsite'
import './LayoutSidebar.scss'

const { Sider } = Layout

export default function LayoutSidebar() {
  const navigate = useNavigate()
  const collapsed = useAppStore((s) => s.sidebarCollapsed)
  const { items, selectedKeys } = useMenuItems()
  const { systemTitle, websiteLogoUrl } = useLayoutWebsite()

  const title = systemTitle || (collapsed ? 'GA' : 'Goravel Admin')

  return (
    <Sider
      className="layout-sidebar"
      collapsible
      collapsed={collapsed}
      trigger={null}
      width={220}
      theme="dark"
    >
      <div className={`layout-sidebar__logo${collapsed ? ' layout-sidebar__logo--collapsed' : ''}`}>
        {websiteLogoUrl ? (
          <img src={websiteLogoUrl} alt="logo" className="layout-sidebar__logo-image" />
        ) : null}
        {!collapsed ? <span className="layout-sidebar__logo-title">{title}</span> : null}
        {collapsed && !websiteLogoUrl ? <span className="layout-sidebar__logo-title">{title}</span> : null}
      </div>
      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={selectedKeys}
        defaultOpenKeys={['/system', '/log', '/payment', '/dev']}
        items={items}
        onClick={({ key }) => {
          if (typeof key === 'string' && key.startsWith('/')) navigate(key)
          else if (typeof key === 'string') navigate(`/${key}`)
        }}
      />
    </Sider>
  )
}
