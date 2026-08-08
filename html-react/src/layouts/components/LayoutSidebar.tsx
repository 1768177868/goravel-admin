import { Drawer, Layout, Menu } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useAppStore } from '@/stores/app'
import { useMenuItems } from '@/hooks/useMenuItems'
import { useLayoutWebsite } from '@/hooks/useLayoutWebsite'
import './LayoutSidebar.scss'

const { Sider } = Layout

interface LayoutSidebarProps {
  isMobile?: boolean
  drawerOpen?: boolean
  onDrawerClose?: () => void
}

export default function LayoutSidebar({
  isMobile = false,
  drawerOpen = false,
  onDrawerClose,
}: LayoutSidebarProps) {
  const navigate = useNavigate()
  const collapsed = useAppStore((s) => s.sidebarCollapsed)
  const menuMode = useAppStore((s) => s.menuMode)
  const { items, selectedKeys } = useMenuItems()
  const { systemTitle, websiteLogoUrl } = useLayoutWebsite()

  const title = systemTitle || (collapsed && !isMobile ? 'GA' : 'Goravel Admin')

  const handleMenuClick = (key: string) => {
    if (key.startsWith('/')) navigate(key)
    else navigate(`/${key}`)
    if (isMobile) onDrawerClose?.()
  }

  const logoNode = (
    <div className={`layout-sidebar__logo${collapsed && !isMobile ? ' layout-sidebar__logo--collapsed' : ''}`}>
      {websiteLogoUrl ? (
        <img src={websiteLogoUrl} alt="logo" className="layout-sidebar__logo-image" />
      ) : null}
      {(!collapsed || isMobile) ? <span className="layout-sidebar__logo-title">{title}</span> : null}
      {collapsed && !isMobile && !websiteLogoUrl ? (
        <span className="layout-sidebar__logo-title">{title}</span>
      ) : null}
    </div>
  )

  const menuNode = (
    <Menu
      theme="dark"
      mode="inline"
      selectedKeys={selectedKeys}
      defaultOpenKeys={['/system', '/log', '/payment', '/dev']}
      items={items}
      onClick={({ key }) => {
        if (typeof key === 'string') handleMenuClick(key)
      }}
    />
  )

  return (
    <>
      {!isMobile && menuMode === 'sidebar' ? (
        <Sider
          className="layout-sidebar"
          collapsible
          collapsed={collapsed}
          trigger={null}
          width={220}
          theme="dark"
        >
          {logoNode}
          {menuNode}
        </Sider>
      ) : null}

      <Drawer
        className="layout-sidebar-drawer"
        open={isMobile && drawerOpen}
        onClose={onDrawerClose}
        placement="left"
        width="min(80vw, 280px)"
        closable={false}
        styles={{ body: { padding: 0 } }}
        destroyOnClose={false}
      >
        <div className="layout-sidebar-drawer__content">
          {logoNode}
          {menuNode}
        </div>
      </Drawer>
    </>
  )
}
