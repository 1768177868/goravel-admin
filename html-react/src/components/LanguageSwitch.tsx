import { Dropdown, type MenuProps } from 'antd'
import { GlobalOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { setLanguage } from '@/i18n'

interface LanguageSwitchProps {
  className?: string
}

export default function LanguageSwitch({ className }: LanguageSwitchProps) {
  const { i18n, t } = useTranslation()
  const current = i18n.language === 'en-US' ? 'en-US' : 'zh-CN'

  const items: MenuProps['items'] = [
    {
      key: 'zh-CN',
      label: t('common.language_zh'),
      onClick: () => setLanguage('zh-CN'),
    },
    {
      key: 'en-US',
      label: t('common.language_en'),
      onClick: () => setLanguage('en-US'),
    },
  ]

  return (
    <Dropdown menu={{ items, selectedKeys: [current] }} placement="bottomRight" trigger={['click']}>
      <button
        type="button"
        className={className || 'layout-header__icon-btn'}
        title={current === 'en-US' ? t('common.language_en') : t('common.language_zh')}
      >
        <GlobalOutlined />
      </button>
    </Dropdown>
  )
}
