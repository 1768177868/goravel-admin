import { Badge, Button, Empty, List, Popover, Space, Typography } from 'antd'
import { BellOutlined } from '@ant-design/icons'
import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useNotificationStore } from '@/stores/notification'
import { useUserStore } from '@/stores/user'

export default function NotificationBell() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const token = useUserStore((s) => s.token)
  const items = useNotificationStore((s) => s.items)
  const unreadCount = useNotificationStore((s) => s.unreadCount)
  const loading = useNotificationStore((s) => s.loading)
  const init = useNotificationStore((s) => s.init)
  const markAsRead = useNotificationStore((s) => s.markAsRead)
  const markAllRead = useNotificationStore((s) => s.markAllRead)
  const disconnect = useNotificationStore((s) => s.disconnect)

  useEffect(() => {
    if (token) {
      void init()
    } else {
      disconnect()
    }
    return () => {
      // keep connection while layout mounted
    }
  }, [token, init, disconnect])

  const content = (
    <div style={{ width: 320 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
        <Typography.Text strong>{t('menu.notification')}</Typography.Text>
        <Button type="link" size="small" onClick={() => void markAllRead()} disabled={!unreadCount}>
          {t('notification.mark_all', { defaultValue: '全部已读' })}
        </Button>
      </div>
      <List
        loading={loading}
        locale={{ emptyText: <Empty description={t('notification.empty', { defaultValue: '暂无通知' })} /> }}
        dataSource={items.slice(0, 7)}
        renderItem={(item) => (
          <List.Item
            style={{ cursor: 'pointer', opacity: item.is_read ? 0.65 : 1 }}
            onClick={() => {
              if (!item.is_read) void markAsRead(item.id)
            }}
          >
            <List.Item.Meta
              title={item.title || '-'}
              description={
                <Space direction="vertical" size={0}>
                  <Typography.Text type="secondary" ellipsis style={{ maxWidth: 280 }}>
                    {item.content}
                  </Typography.Text>
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    {item.created_at}
                  </Typography.Text>
                </Space>
              }
            />
          </List.Item>
        )}
      />
      <Button type="link" block onClick={() => navigate('/notifications')}>
        {t('notification.view_all', { defaultValue: '查看全部' })}
      </Button>
    </div>
  )

  return (
    <Popover content={content} trigger="click" placement="bottomRight">
      <Badge count={unreadCount} size="small" offset={[-2, 2]}>
        <Button type="text" icon={<BellOutlined />} />
      </Badge>
    </Popover>
  )
}
