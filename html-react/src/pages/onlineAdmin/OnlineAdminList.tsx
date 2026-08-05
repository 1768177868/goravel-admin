import { useState } from 'react'
import { Avatar, Button, Space, Table } from 'antd'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import { useTranslation } from 'react-i18next'
import { SettingOutlined } from '@ant-design/icons'
import { App } from 'antd'
import {
  batchKickOutOnlineAdmins,
  getOnlineAdminList,
  kickOutOnlineAdmin,
} from '@/api/onlineAdmin'
import { useListPage } from '@/hooks/useListPage'
import { useCrudActions } from '@/hooks/useCrudActions'
import { usePermission } from '@/hooks/usePermission'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { useColumnSetting } from '@/hooks/useColumnSetting'
import PageContainer from '@/components/PageContainer'
import ColumnSettingDialog from '@/components/ColumnSettingDialog'
import SearchForm from '@/components/SearchForm'
import PermissionButton from '@/components/PermissionButton'
import { entityField } from '@/utils/normalize'

interface OnlineAdminRow {
  id: number | string
  username?: string
  nickname?: string
  avatar?: string
  browser?: string
  ip?: string
  os?: string
  session_id?: string
  last_active?: string
  name?: string
}

function formatOnlineTime(time?: string) {
  if (!time) return '-'
  const date = new Date(time)
  if (Number.isNaN(date.getTime())) return time
  return date.toLocaleString().replace(/\//g, '-')
}

export default function OnlineAdminList() {
  const { t } = useTranslation()
  const { modal, message } = App.useApp()
  const showError = useUnhandledError()
  const { getButtonState } = usePermission()
  const [selectedRowKeys, setSelectedRowKeys] = useState<Array<string | number>>([])

  const {
    tableData,
    loading,
    pagination,
    searchForm,
    setSearchForm,
    loadData,
    handleSearch,
    handleReset,
    handleSortChange,
    refresh,
  } = useListPage<OnlineAdminRow>({
    fetchApi: getOnlineAdminList as never,
    initialSearchForm: { username: '', ip: '', browser: '', os: '' },
    fieldMapping: { last_active: 'last_used_at' },
    defaultSort: 'last_used_at:desc',
    normalizeRows: false,
    transformData: (row) => {
      const record = row as unknown as Record<string, unknown>
      return {
        id: entityField(record, 'id', '')!,
        username: String(entityField(record, 'username', '') ?? ''),
        nickname: String(entityField(record, 'nickname', '') ?? ''),
        avatar: String(entityField(record, 'avatar', '') ?? ''),
        browser: String(entityField(record, 'browser', '') ?? ''),
        ip: String(entityField(record, 'ip', '') ?? ''),
        os: String(entityField(record, 'os', '') ?? ''),
        session_id: String(entityField(record, 'session_id', '') ?? ''),
        last_active: String(
          entityField(record, 'last_active', '') ?? entityField(record, 'last_used_at', '') ?? '',
        ),
        name: String(entityField(record, 'username', '') ?? ''),
      }
    },
  })

  const { toolbar } = useCrudActions({ onRefresh: refresh })

  const handleKickOut = (row: OnlineAdminRow) => {
    modal.confirm({
      title: t('common.confirm'),
      content: t('online_admin.kick_out_confirm', { username: row.username }),
      okType: 'danger',
      onOk: async () => {
        try {
          await kickOutOnlineAdmin(row.id)
          message.success(t('online_admin.kick_out_success'))
          await refresh()
        } catch (error) {
          showError(error, t('online_admin.kick_out_failed'))
        }
      },
    })
  }

  const handleBatchKickOut = () => {
    if (!selectedRowKeys.length) return
    modal.confirm({
      title: t('common.confirm'),
      content: t('online_admin.batch_kick_out_confirm', { count: selectedRowKeys.length }),
      okType: 'danger',
      onOk: async () => {
        try {
          await batchKickOutOnlineAdmins(selectedRowKeys)
          message.success(t('online_admin.batch_kick_out_success'))
          setSelectedRowKeys([])
          await refresh()
        } catch (error) {
          showError(error, t('online_admin.batch_kick_out_failed'))
        }
      },
    })
  }

  const columns: ColumnsType<OnlineAdminRow> = [
    { title: t('online_admin.username'), dataIndex: 'username', width: 120 },
    { title: t('online_admin.nickname'), dataIndex: 'nickname', width: 120 },
    {
      title: t('online_admin.avatar'),
      dataIndex: 'avatar',
      width: 80,
      render: (_, row) => (
        <Avatar src={row.avatar || undefined} size={32}>
          {row.nickname?.charAt(0) || row.username?.charAt(0) || 'U'}
        </Avatar>
      ),
    },
    { title: t('online_admin.browser'), dataIndex: 'browser', width: 150 },
    { title: t('online_admin.ip'), dataIndex: 'ip', width: 150 },
    { title: t('online_admin.os'), dataIndex: 'os', width: 150 },
    { title: t('online_admin.session_id'), dataIndex: 'session_id', width: 200, ellipsis: true },
    {
      title: t('online_admin.last_active'),
      dataIndex: 'last_active',
      width: 180,
      sorter: true,
      render: (time: string) => formatOnlineTime(time),
    },
    {
      title: t('common.operation'),
      key: 'operation',
      width: 120,
      fixed: 'right',
      render: (_, row) =>
        getButtonState('admin.kick_out').show ? (
          <PermissionButton permission="admin.kick_out" type="link" danger onClick={() => handleKickOut(row)}>
            {t('online_admin.kick_out')}
          </PermissionButton>
        ) : null,
    },
  ]

  const {
    filteredColumns,
    open: columnSettingOpen,
    openColumnSetting,
    closeColumnSetting,
    allColumns,
    visibleColumns,
    columnOrder,
    fixedColumns,
    handleConfirm: handleColumnSettingConfirm,
  } = useColumnSetting('online_admin', columns)

  return (
    <PageContainer
      title={t('menu.online_admin')}
      extra={
        <Space>
          {selectedRowKeys.length > 0 && getButtonState('admin.kick_out').show ? (
            <Button danger onClick={handleBatchKickOut}>
              {t('online_admin.batch_kick_out')} ({selectedRowKeys.length})
            </Button>
          ) : null}
          {toolbar}
          <Button icon={<SettingOutlined />} onClick={openColumnSetting}>
            {t('common.column_setting')}
          </Button>
        </Space>
      }
    >
      <SearchForm
        fields={[
          { name: 'username', label: t('online_admin.username') },
          { name: 'ip', label: t('online_admin.ip') },
          { name: 'browser', label: t('online_admin.browser') },
          { name: 'os', label: t('online_admin.os') },
        ]}
        values={searchForm}
        onChange={(values) => setSearchForm(values as never)}
        onSearch={handleSearch}
        onReset={handleReset}
      />
      <Table<OnlineAdminRow>
        rowKey="id"
        loading={loading}
        columns={filteredColumns}
        dataSource={tableData}
        scroll={{ x: 1200 }}
        rowSelection={{
          selectedRowKeys,
          onChange: (keys) => setSelectedRowKeys(keys as Array<string | number>),
        }}
        pagination={{
          current: pagination.page,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
          showTotal: (total) => t('common.total', { total }),
        }}
        onChange={(pager: TablePaginationConfig, _f, sorter) => {
          const sort = Array.isArray(sorter) ? sorter[0] : sorter
          const sortObj = sort as { field?: string; order?: 'ascend' | 'descend' | null; column?: unknown }
          if (sortObj?.column && sortObj.field) {
            handleSortChange(String(sortObj.field), sortObj.order)
            return
          }
          void loadData({
            currentPage: pager.current || 1,
            pageSize: pager.pageSize || pagination.pageSize,
          })
        }}
      />
      <ColumnSettingDialog
        open={columnSettingOpen}
        onClose={closeColumnSetting}
        allColumns={allColumns}
        visibleColumns={visibleColumns}
        columnOrder={columnOrder}
        fixedColumns={fixedColumns}
        onConfirm={handleColumnSettingConfirm}
      />
    </PageContainer>
  )
}
