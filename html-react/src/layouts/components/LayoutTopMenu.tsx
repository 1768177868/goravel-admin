import { Menu } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useMenuItems } from '@/hooks/useMenuItems'
import './LayoutTopMenu.scss'

export default function LayoutTopMenu() {
  const navigate = useNavigate()
  const { items, selectedKeys } = useMenuItems()

  return (
    <div className="layout-top-menu">
      <Menu
        mode="horizontal"
        className="layout-top-menu__menu"
        selectedKeys={selectedKeys}
        items={items}
        onClick={({ key }) => {
          if (typeof key === 'string' && key.startsWith('/')) navigate(key)
          else if (typeof key === 'string') navigate(`/${key}`)
        }}
      />
    </div>
  )
}
