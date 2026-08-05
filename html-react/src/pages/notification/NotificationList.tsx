import { Table, Tag } from 'antd'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import { useTranslation } from 'react-i18next'
import { getNotificationList } from '@/api/notification'
import { useListPage } from '@/hooks/useListPage'
import { useCrudActions } from '@/hooks/useCrudActions'
import { useNotificationStore } from '@/stores/notification'
import PageContainer from '@/components/PageContainer'
import SearchForm from '@/components/SearchForm'
import { entityField } from '@/utils/normalize'
import { Button } from 'antd'

interface Row {
  id: number | string
  title?: string
  content?: string
  type?: string
  is_read?: boolean
  created_at?: string
  name?: string
}

export default function NotificationList() {
  const { t } = useTranslation()
  const markAsRead = useNotificationStore((s) => s.markAsRead)
  const markAllRead = useNotificationStore((s) => s.markAllRead)
  const refreshBell = useNotificationStore((s) => s.refresh)

  const {
    tableData,
    loading,
    pagination,
    searchForm,
    setSearchForm,
    loadData,
    handleSearch,
    handleReset,
    refresh,
  } = useListPage<Row>({
    fetchApi: getNotificationList as never,
    initialSearchForm: { title: '' },
    normalizeRows: true,
    transformData: (row) => {
      const record = row as unknown as Record<string, unknown>
      return {
        id: entityField(record, 'id', '')!,
        title: String(entityField(record, 'title', '') ?? ''),
        content: String(entityField(record, 'content', '') ?? ''),
        type: String(entityField(record, 'type', '') ?? ''),
        is_read: !!(entityField(record, 'is_read', false) || entityField(record, 'IsRead', false)),
        created_at: String(entityField(record, 'created_at', '') ?? ''),
        name: String(entityField(record, 'title', '') ?? ''),
      }
    },
  })

  const { toolbar } = useCrudActions({
    onRefresh: async () => {
      await refresh()
      await refreshBell()
    },
  })

  const columns: ColumnsType<Row> = [
    { title: t('table.id'), dataIndex: 'id', width: 80 },
    { title: t('table.title'), dataIndex: 'title' },
    { title: t('common.description'), dataIndex: 'content', ellipsis: true },
    { title: t('common.type'), dataIndex: 'type', width: 100 },
    {
      title: t('common.status'),
      dataIndex: 'is_read',
      width: 100,
      render: (read: boolean) =>
        read ? (
          <Tag>{t('notification.read', { defaultValue: '已读' })}</Tag>
        ) : (
          <Tag color="processing">{t('notification.unread', { defaultValue: '未读' })}</Tag>
        ),
    },
    { title: t('table.created_at'), dataIndex: 'created_at', width: 180 },
    {
      title: t('common.operation'),
      key: 'operation',
      width: 120,
      render: (_, row) =>
        !row.is_read ? (
          <Button
            type="link"
            onClick={async () => {
              await markAsRead(row.id)
              await refresh()
            }}
          >
            {t('notification.mark_read', { defaultValue: '标记已读' })}
          </Button>
        ) : null,
    },
  ]

  return (
    <PageContainer
      title={t('menu.notification')}
      extra={
        <>
          {toolbar}
          <Button onClick={() => void markAllRead().then(() => refresh())}>
            {t('notification.mark_all', { defaultValue: '全部已读' })}
          </Button>
        </>
      }
    >
      <SearchForm
        fields={[{ name: 'title', label: t('table.title') }]}
        values={searchForm}
        onChange={(values) => setSearchForm(values as never)}
        onSearch={handleSearch}
        onReset={handleReset}
      />
      <Table<Row>
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={tableData}
        pagination={{
          current: pagination.page,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
          showTotal: (total) => t('common.total', { total }),
        }}
        onChange={(pager: TablePaginationConfig) => {
          void loadData({
            currentPage: pager.current || 1,
            pageSize: pager.pageSize || pagination.pageSize,
          })
        }}
      />
    </PageContainer>
  )
}
