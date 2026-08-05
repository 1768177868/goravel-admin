import { Layout, Menu, type MenuProps } from 'antd'
import { useMemo } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { DashboardOutlined } from '@ant-design/icons'
import { useAppStore } from '@/stores/app'
import { useUserStore } from '@/stores/user'
import { resolveMenuIcon } from '@/utils/menuIcons'
import type { MenuNode } from '@/types'
import './LayoutSidebar.scss'

const { Sider } = Layout

type AntdItem = Required<MenuProps>['items'][number]

function menuTitle(node: MenuNode, t: (key: string) => string) {
  const slug = node.Slug || node.slug
  if (slug) {
    const clean = String(slug).replace(/^\//, '')
    for (const key of [`menu.${clean}`, `menu.${clean}_management`]) {
      const translated = t(key)
      if (translated !== key) return translated
    }
  }
  return node.Title || node.title || node.Name || node.name || slug || 'menu'
}

function toMenuItems(nodes: MenuNode[], t: (key: string) => string): AntdItem[] {
  return nodes
    .filter((node) => {
      const status = node.Status ?? node.status ?? 1
      const type = node.Type ?? node.type ?? 1
      const hidden =
        (node as { IsHidden?: number; is_hidden?: number }).IsHidden ??
        (node as { is_hidden?: number }).is_hidden ??
        0
      return status === 1 && type !== 3 && hidden !== 1
    })
    .sort((a, b) => (a.Sort ?? a.sort ?? 0) - (b.Sort ?? b.sort ?? 0))
    .map((node) => {
      const children = node.children || node.Children || []
      const path = node.Path || node.path || ''
      const type = node.Type ?? node.type ?? 1
      const key = path || String(node.id || node.ID || menuTitle(node, t))
      const item: AntdItem = {
        key,
        icon: resolveMenuIcon(node.Icon || node.icon),
        label: menuTitle(node, t),
      }

      if (type === 1 && Array.isArray(children) && children.length > 0) {
        ;(item as { children?: AntdItem[] }).children = toMenuItems(children, t)
      }

      return item
    })
}

export default function LayoutSidebar() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()
  const collapsed = useAppStore((s) => s.sidebarCollapsed)
  const menus = useUserStore((s) => s.menus)

  const items = useMemo(() => {
    const dashboardItem: AntdItem = {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: t('menu.dashboard'),
    }
    return [dashboardItem, ...toMenuItems(menus, t)]
  }, [menus, t])

  const selectedKeys = useMemo(() => {
    const path = location.pathname
    return [path.startsWith('/') ? path : `/${path}`]
  }, [location.pathname])

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
