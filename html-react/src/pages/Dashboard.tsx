import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Col, List, Row, Space, Statistic, Typography } from 'antd'
import {
  ReloadOutlined,
  ShoppingCartOutlined,
  TeamOutlined,
  UserOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { getCount, getRecentActivities } from '@/api/dashboard'
import { useUserStore } from '@/stores/user'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import PageContainer from '@/components/PageContainer'
import { entityField } from '@/utils/normalize'

export default function DashboardPage() {
  const { t } = useTranslation()
  const showError = useUnhandledError()
  const adminInfo = useUserStore((s) => s.adminInfo)
  const [loading, setLoading] = useState(false)
  const [counts, setCounts] = useState<Record<string, number>>({})
  const [activities, setActivities] = useState<Array<Record<string, unknown>>>([])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [countRes, activityRes] = await Promise.all([
        getCount(),
        getRecentActivities().catch(() => null),
      ])
      const data = (countRes.data || {}) as Record<string, unknown>
      setCounts({
        admins: Number(entityField(data, 'admin_count', entityField(data, 'admins', 0)) ?? 0),
        users: Number(entityField(data, 'user_count', entityField(data, 'users', 0)) ?? 0),
        orders: Number(entityField(data, 'order_count', entityField(data, 'orders', 0)) ?? 0),
        roles: Number(entityField(data, 'role_count', entityField(data, 'roles', 0)) ?? 0),
      })

      const raw =
        (activityRes as { data?: unknown } | null)?.data ??
        []
      const list = Array.isArray(raw)
        ? raw
        : Array.isArray((raw as { list?: unknown[] })?.list)
          ? ((raw as { list: unknown[] }).list)
          : []
      setActivities(list as Array<Record<string, unknown>>)
    } catch (error) {
      showError(error, t('common.query_failed'))
    } finally {
      setLoading(false)
    }
  }, [showError, t])

  useEffect(() => {
    void load()
  }, [load])

  const cards = [
    { key: 'admins', title: t('dashboard.admins'), value: counts.admins, icon: <SafetyCertificateOutlined />, color: '#1677ff' },
    { key: 'users', title: t('dashboard.users'), value: counts.users, icon: <UserOutlined />, color: '#52c41a' },
    { key: 'orders', title: t('dashboard.orders'), value: counts.orders, icon: <ShoppingCartOutlined />, color: '#fa8c16' },
    { key: 'roles', title: t('dashboard.roles'), value: counts.roles, icon: <TeamOutlined />, color: '#722ed1' },
  ]

  return (
    <PageContainer
      title={t('menu.dashboard')}
      extra={
        <Button icon={<ReloadOutlined />} loading={loading} onClick={() => void load()}>
          {t('dashboard.refresh')}
        </Button>
      }
    >
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        {t('dashboard.welcome')}
        {adminInfo?.nickname || adminInfo?.username
          ? `，${adminInfo.nickname || adminInfo.username}`
          : ''}
      </Typography.Title>
      <Typography.Paragraph type="secondary">{t('dashboard.subtitle')}</Typography.Paragraph>

      <Row gutter={[16, 16]}>
        {cards.map((card) => (
          <Col xs={24} sm={12} lg={6} key={card.key}>
            <Card loading={loading}>
              <Space align="start" size="large">
                <div
                  style={{
                    width: 48,
                    height: 48,
                    borderRadius: 12,
                    display: 'grid',
                    placeItems: 'center',
                    color: '#fff',
                    background: card.color,
                    fontSize: 22,
                  }}
                >
                  {card.icon}
                </div>
                <Statistic title={card.title} value={card.value || 0} />
              </Space>
            </Card>
          </Col>
        ))}
      </Row>

      <Card title={t('dashboard.recent')} style={{ marginTop: 16 }} loading={loading}>
        <List
          locale={{ emptyText: t('common.no_data') }}
          dataSource={activities.slice(0, 8)}
          renderItem={(item) => (
            <List.Item>
              <List.Item.Meta
                title={String(entityField(item, 'title', entityField(item, 'content', '-')) ?? '-')}
                description={String(entityField(item, 'created_at', '') ?? '')}
              />
            </List.Item>
          )}
        />
      </Card>
    </PageContainer>
  )
}
