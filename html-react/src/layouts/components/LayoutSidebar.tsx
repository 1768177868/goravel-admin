import { Layout, Menu } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useAppStore } from '@/stores/app'
import { useMenuItems } from '@/hooks/useMenuItems'
import './LayoutSidebar.scss'

const { Sider } = Layout

export default function LayoutSidebar() {
  const navigate = useNavigate()
  const collapsed = useAppStore((s) => s.sidebarCollapsed)
  const { items, selectedKeys } = useMenuItems()

  return (
    <Sider
      className="layout-sidebar"
      collapsible
      collapsed={collapsed}
      trigger={null}
      width={220}
      theme="dark"
    >
      <div className="layout-sidebar__logo">{collapsed ? 'GA' : 'Goravel Admin'}</div>
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
