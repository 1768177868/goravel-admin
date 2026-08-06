import { Dropdown, Tabs, type MenuProps } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useTabsStore } from '@/stores/tabs'
import { resolveMenuTitle } from '@/utils/menuTitle'
import './TabsView.scss'

export default function TabsView() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const tabs = useTabsStore((s) => s.tabs)
  const activeTab = useTabsStore((s) => s.activeTab)
  const removeTab = useTabsStore((s) => s.removeTab)
  const removeOtherTabs = useTabsStore((s) => s.removeOtherTabs)
  const removeAllTabs = useTabsStore((s) => s.removeAllTabs)
  const setActiveTab = useTabsStore((s) => s.setActiveTab)

  if (tabs.length === 0) return null

  const contextMenu = (path: string): MenuProps => ({
    items: [
      {
        key: 'close',
        label: t('common.close'),
        onClick: () => {
          const next = removeTab(path)
          if (next) navigate(next)
          else navigate('/dashboard')
        },
      },
      {
        key: 'close-others',
        label: t('common.close_others'),
        onClick: () => {
          removeOtherTabs(path)
          navigate(path)
        },
      },
      {
        key: 'close-all',
        label: t('common.close_all'),
        onClick: () => {
          removeAllTabs()
          navigate('/dashboard')
        },
      },
    ],
  })

  return (
    <div className="tabs-view">
      <Tabs
        type="editable-card"
        hideAdd
        size="small"
        activeKey={activeTab || undefined}
        onChange={(key) => {
          setActiveTab(key)
          navigate(key)
        }}
        onEdit={(key, action) => {
          if (action === 'remove' && typeof key === 'string') {
            const next = removeTab(key)
            if (activeTab === key) {
              if (next) navigate(next)
              else navigate('/dashboard')
            }
          }
        }}
        items={tabs.map((tab) => ({
          key: tab.path,
          label: (
            <Dropdown menu={contextMenu(tab.path)} trigger={['contextMenu']}>
              <span>
                {resolveMenuTitle(t, {
                  titleKey: tab.titleKey,
                  slug: tab.titleKey?.replace(/^menu\./, ''),
                  fallback: tab.title,
                })}
              </span>
            </Dropdown>
          ),
          closable: tab.path !== '/dashboard',
        }))}
      />
    </div>
  )
}
