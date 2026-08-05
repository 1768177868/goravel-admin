import { Avatar, Dropdown, Layout, Popover, Segmented, Space, Switch, theme, type MenuProps } from 'antd'
import {
  CheckOutlined,
  ColumnHeightOutlined,
  FullscreenExitOutlined,
  FullscreenOutlined,
  LockOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SettingOutlined,
  UnorderedListOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAppStore, THEME_COLORS, type LayoutSize, type MenuMode } from '@/stores/app'
import { useUserStore } from '@/stores/user'
import { useTabsStore } from '@/stores/tabs'
import { useNotificationStore } from '@/stores/notification'
import LanguageSwitch from '@/components/LanguageSwitch'
import DarkModeSwitch from '@/components/DarkModeSwitch'
import NotificationBell from '@/components/NotificationBell'
import TimezoneSwitch from '@/components/TimezoneSwitch'
import './LayoutHeader.scss'

const { Header } = Layout

interface LayoutHeaderProps {
  onLockScreen: () => void
}

export default function LayoutHeader({ onLockScreen }: LayoutHeaderProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { token } = theme.useToken()
  const collapsed = useAppStore((s) => s.sidebarCollapsed)
  const toggleSidebar = useAppStore((s) => s.toggleSidebar)
  const menuMode = useAppStore((s) => s.menuMode)
  const setMenuMode = useAppStore((s) => s.setMenuMode)
  const isFullscreen = useAppStore((s) => s.isFullscreen)
  const toggleFullscreen = useAppStore((s) => s.toggleFullscreen)
  const layoutSize = useAppStore((s) => s.layoutSize)
  const setLayoutSize = useAppStore((s) => s.setLayoutSize)
  const watermarkEnabled = useAppStore((s) => s.watermarkEnabled)
  const setWatermarkEnabled = useAppStore((s) => s.setWatermarkEnabled)
  const themeColor = useAppStore((s) => s.themeColor)
  const setThemeColor = useAppStore((s) => s.setThemeColor)
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

  const layoutSizeMenu: MenuProps['items'] = (['large', 'default', 'small'] as LayoutSize[]).map((size) => ({
    key: size,
    label: (
      <span className="layout-size-option">
        <span>{t(`header.layout_size_${size}`)}</span>
        {layoutSize === size && <CheckOutlined className="layout-size-option__check" />}
      </span>
    ),
  }))

  const settingsContent = (
    <div className="header-settings">
      <div className="header-settings__title">{t('header.settings')}</div>
      <div className="header-settings__item header-settings__item--column">
        <span className="header-settings__label">{t('header.menu_mode')}</span>
        <Segmented
          block
          value={menuMode}
          onChange={(value) => setMenuMode(value as MenuMode)}
          options={[
            { label: t('header.menu_mode_sidebar'), value: 'sidebar', icon: <MenuFoldOutlined /> },
            { label: t('header.menu_mode_top'), value: 'top', icon: <UnorderedListOutlined /> },
          ]}
        />
      </div>
      <div className="header-settings__item">
        <span className="header-settings__label">{t('header.watermark')}</span>
        <Switch checked={watermarkEnabled} onChange={setWatermarkEnabled} />
      </div>
      <div className="header-settings__item header-settings__item--column">
        <span className="header-settings__label">{t('header.theme_color')}</span>
        <Space size={10} wrap>
          {THEME_COLORS.map((color) => (
            <button
              key={color.key}
              type="button"
              className={`theme-swatch${themeColor === color.key ? ' theme-swatch--active' : ''}`}
              style={{ backgroundColor: color.color }}
              title={color.key}
              onClick={() => setThemeColor(color.key)}
            />
          ))}
        </Space>
      </div>
    </div>
  )

  return (
    <Header
      className="layout-header"
      style={{
        background: token.colorBgContainer,
        borderBottom: `1px solid ${token.colorBorderSecondary}`,
      }}
    >
      <div className="layout-header__left">
        {menuMode === 'sidebar' && (
          <button type="button" className="layout-header__trigger" onClick={toggleSidebar}>
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </button>
        )}
      </div>
      <Space size="middle" className="layout-header__right">
        <button
          type="button"
          className="layout-header__icon-btn"
          title={t('header.fullscreen')}
          onClick={toggleFullscreen}
        >
          {isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
        </button>
        <Dropdown menu={{ items: layoutSizeMenu, onClick: ({ key }) => setLayoutSize(key as LayoutSize) }}>
          <button type="button" className="layout-header__icon-btn" title={t('header.layout_size')}>
            <ColumnHeightOutlined />
          </button>
        </Dropdown>
        <Popover
          content={settingsContent}
          trigger="click"
          placement="bottomRight"
          classNames={{ root: 'header-settings-popover' }}
        >
          <button type="button" className="layout-header__icon-btn" title={t('header.settings')}>
            <SettingOutlined />
          </button>
        </Popover>
        <NotificationBell />
        <DarkModeSwitch />
        <LanguageSwitch />
        <button
          type="button"
          className="layout-header__icon-btn"
          title={t('header.lock_screen')}
          onClick={onLockScreen}
        >
          <LockOutlined />
        </button>
        <TimezoneSwitch />
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
