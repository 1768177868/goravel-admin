import { Switch, Tooltip } from 'antd'
import { MoonOutlined, SunOutlined } from '@ant-design/icons'
import { useAppStore } from '@/stores/app'

export default function DarkModeSwitch() {
  const darkMode = useAppStore((s) => s.darkMode)
  const toggleDarkMode = useAppStore((s) => s.toggleDarkMode)

  return (
    <Tooltip title={darkMode ? 'Light' : 'Dark'}>
      <Switch
        size="small"
        checked={darkMode}
        onChange={toggleDarkMode}
        checkedChildren={<MoonOutlined />}
        unCheckedChildren={<SunOutlined />}
      />
    </Tooltip>
  )
}
