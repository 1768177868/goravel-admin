import { MoonOutlined, SunOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'

interface DarkModeSwitchProps {
  className?: string
}

export default function DarkModeSwitch({ className }: DarkModeSwitchProps) {
  const { t } = useTranslation()
  const darkMode = useAppStore((s) => s.darkMode)
  const toggleDarkMode = useAppStore((s) => s.toggleDarkMode)

  return (
    <button
      type="button"
      className={className || 'layout-header__icon-btn'}
      title={darkMode ? t('header.switch_to_light') : t('header.switch_to_dark')}
      onClick={toggleDarkMode}
    >
      {darkMode ? <SunOutlined /> : <MoonOutlined />}
    </button>
  )
}
