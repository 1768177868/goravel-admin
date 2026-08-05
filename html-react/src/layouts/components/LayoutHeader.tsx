import { Avatar, Dropdown, Layout, Space, theme, type MenuProps } from 'antd'
import {
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { useUserStore } from '@/stores/user'
import { useTabsStore } from '@/stores/tabs'
import { useNotificationStore } from '@/stores/notification'
import LanguageSwitch from '@/components/LanguageSwitch'
import DarkModeSwitch from '@/components/DarkModeSwitch'
import NotificationBell from '@/components/NotificationBell'
import './LayoutHeader.scss'

const { Header } = Layout

export default function LayoutHeader() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { token } = theme.useToken()
  const collapsed = useAppStore((s) => s.sidebarCollapsed)
  const toggleSidebar = useAppStore((s) => s.toggleSidebar)
  const adminInfo = useUserStore((s) => s.adminInfo)
  const logout = useUserStore((s) => s.logout)
  const removeAllTabs = useTabsStore((s) => s.removeAllTabs)
  const disconnectNotifications = useNotificationStore((s) => s.disconnect)

  const userMenu: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: t('common.profile'),
      onClick: () => navigate('/profile'),
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: t('common.logout'),
      onClick: async () => {
        disconnectNotifications()
        removeAllTabs()
        await logout()
        navigate('/login', { replace: true })
      },
    },
  ]

  return (
    <Header
      className="layout-header"
      style={{
        background: token.colorBgContainer,
        borderBottom: `1px solid ${token.colorBorderSecondary}`,
      }}
    >
      <div className="layout-header__left">
        <button type="button" className="layout-header__trigger" onClick={toggleSidebar}>
          {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        </button>
      </div>
      <Space size="middle" className="layout-header__right">
        <NotificationBell />
        <DarkModeSwitch />
        <LanguageSwitch />
        <Dropdown menu={{ items: userMenu }} placement="bottomRight">
          <Space className="layout-header__user">
            <Avatar size="small" src={adminInfo?.avatar} icon={<UserOutlined />} />
            <span>{adminInfo?.nickname || adminInfo?.username || 'Admin'}</span>
          </Space>
        </Dropdown>
      </Space>
    </Header>
  )
}
