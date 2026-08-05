import { useState } from 'react'
import { App, Button, Dropdown, Form, Input, Space, Switch, Table, Tag } from 'antd'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import type { MenuProps } from 'antd'
import { DownOutlined, SettingOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import {
  deleteAdmin,
  exportAdmin,
  getAdminList,
  kickOutUser,
  resetAdminGoogleAuth,
  resetPassword,
  unbindAdminGoogleAuth,
  updateAdmin,
} from '@/api/admin'
import { useListPage } from '@/hooks/useListPage'
import { useCrudActions } from '@/hooks/useCrudActions'
import { usePermission } from '@/hooks/usePermission'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { useColumnSetting } from '@/hooks/useColumnSetting'
import PageContainer from '@/components/PageContainer'
import SearchForm from '@/components/SearchForm'
import PermissionButton from '@/components/PermissionButton'
import ColumnSettingDialog from '@/components/ColumnSettingDialog'
import { entityField } from '@/utils/normalize'
import AdminFormModal from './AdminFormModal'

interface AdminRow {
  id: number | string
  username: string
  nickname?: string
  email?: string
  phone?: string
  status?: number
  created_at?: string
  is_2fa_bound?: boolean
  department?: { name?: string } | string
  position?: { name?: string } | string
  roles?: Array<{ id?: number | string; name?: string }>
}

const protectedIds = new Set([1, 2])

export default function AdminList() {
  const { t } = useTranslation()
  const { message, modal } = App.useApp()
  const showError = useUnhandledError()
  const { getButtonState } = usePermission()
  const [formOpen, setFormOpen] = useState(false)
  const [editId, setEditId] = useState<string | number | null>(null)
  const [exporting, setExporting] = useState(false)
  const [pwdForm] = Form.useForm()

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
  } = useListPage<AdminRow, Record<string, unknown>>({
    fetchApi: getAdminList as never,
    initialSearchForm: { username: '', status: '', is_2fa_bound: '' },
    defaultSort: 'id:desc',
    normalizeRows: true,
    transformData: (row) => {
      const record = row as unknown as Record<string, unknown>
      return {
        id: entityField(record, 'id', '')!,
        username: String(entityField(record, 'username', '') ?? ''),
        nickname: String(entityField(record, 'nickname', '') ?? ''),
        email: String(entityField(record, 'email', '') ?? ''),
        phone: String(entityField(record, 'phone', '') ?? ''),
        status: Number(entityField(record, 'status', 0) ?? 0),
        created_at: String(entityField(record, 'created_at', '') ?? ''),
        is_2fa_bound: !!(
          entityField(record, 'is_2fa_bound', false) || entityField(record, 'Is2faBound', false)
        ),
        department: (entityField(record, 'department', null) as AdminRow['department']) || undefined,
        position: (entityField(record, 'position', null) as AdminRow['position']) || undefined,
        roles: (entityField(record, 'roles', []) as AdminRow['roles']) || [],
      }
    },
  })

  const { toolbar, confirmDelete } = useCrudActions({
    createPermission: 'admin.store',
    onRefresh: refresh,
    onCreate: () => {
      setEditId(null)
      setFormOpen(true)
    },
    deleteApi: deleteAdmin,
  })

  const deptName = (dept: AdminRow['department']) => {
    if (!dept) return '-'
    if (typeof dept === 'string') return dept
    return dept.name || '-'
  }

  const posName = (pos: AdminRow['position']) => {
    if (!pos) return '-'
    if (typeof pos === 'string') return pos
    return pos.name || '-'
  }

  const handleStatusChange = async (row: AdminRow, checked: boolean) => {
    if (protectedIds.has(Number(row.id)) && !checked) {
      message.warning(t('admin.protected_cannot_disable', { defaultValue: '受保护账号不能禁用' }))
      return
    }
    try {
      await updateAdmin(row.id, { status: checked ? 1 : 0 })
      message.success(t('common.update_success'))
      await refresh()
    } catch (error) {
      showError(error, t('common.operation_failed'))
    }
  }

  const handleResetPassword = (row: AdminRow) => {
    pwdForm.resetFields()
    modal.confirm({
      title: t('admin.reset_password'),
      content: (
        <Form form={pwdForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="password"
            label={t('admin.new_password', { defaultValue: '新密码' })}
            rules={[{ required: true, message: t('admin.password_required') }]}
          >
            <Input.Password />
          </Form.Item>
        </Form>
      ),
      onOk: async () => {
        const values = await pwdForm.validateFields()
        try {
          await resetPassword(row.id, { password: values.password })
          message.success(t('admin.reset_password_success', { defaultValue: '密码已重置' }))
        } catch (error) {
          showError(error, t('common.operation_failed'))
          throw error
        }
      },
    })
  }

  const runConfirm = (title: string, content: string, action: () => Promise<void>) => {
    modal.confirm({
      title,
      content,
      onOk: async () => {
        try {
          await action()
        } catch (error) {
          showError(error, t('common.operation_failed'))
          throw error
        }
      },
    })
  }

  const moreMenu = (row: AdminRow): MenuProps['items'] =>
    [
      getButtonState('admin.password').show
        ? {
            key: 'resetPassword',
            label: t('admin.reset_password'),
            onClick: () => handleResetPassword(row),
          }
        : null,
      getButtonState('admin.kick_out').show
        ? {
            key: 'kickOut',
            label: t('admin.kick_out', { defaultValue: '踢出登录' }),
            onClick: () =>
              runConfirm(
                t('admin.kick_out', { defaultValue: '踢出登录' }),
                t('admin.kick_out_confirm', {
                  username: row.username,
                  defaultValue: `确定踢出 ${row.username} 的全部登录吗？`,
                }),
                async () => {
                  await kickOutUser(row.id)
                  message.success(t('admin.kick_out_success', { defaultValue: '已踢出' }))
                },
              ),
          }
        : null,
      row.is_2fa_bound && getButtonState('admin.unbind_google_auth').show
        ? {
            key: 'unbind2fa',
            label: t('admin.unbind_google_auth', { defaultValue: '解绑谷歌验证' }),
            onClick: () =>
              runConfirm(
                t('admin.unbind_google_auth', { defaultValue: '解绑谷歌验证' }),
                t('admin.unbind_google_auth_confirm', {
                  defaultValue: '确定解绑该管理员的谷歌验证码吗？',
                }),
                async () => {
                  await unbindAdminGoogleAuth(row.id)
                  message.success(t('common.update_success'))
                  await refresh()
                },
              ),
          }
        : null,
      getButtonState('admin.reset_google_auth').show
        ? {
            key: 'reset2fa',
            label: t('admin.reset_google_auth', { defaultValue: '重置谷歌验证' }),
            onClick: () =>
              runConfirm(
                t('admin.reset_google_auth', { defaultValue: '重置谷歌验证' }),
                t('admin.reset_google_auth_confirm', {
                  defaultValue: '确定重置该管理员的谷歌验证码吗？',
                }),
                async () => {
                  await resetAdminGoogleAuth(row.id)
                  message.success(t('common.update_success'))
                  await refresh()
                },
              ),
          }
        : null,
    ].filter(Boolean) as MenuProps['items']

  const columns: ColumnsType<AdminRow> = [
    { title: t('table.id'), dataIndex: 'id', width: 80, sorter: true },
    { title: t('table.username'), dataIndex: 'username' },
    { title: t('table.nickname'), dataIndex: 'nickname' },
    { title: t('table.email'), dataIndex: 'email' },
    { title: t('table.department'), dataIndex: 'department', render: (v) => deptName(v) },
    { title: t('table.position'), dataIndex: 'position', render: (v) => posName(v) },
    {
      title: t('table.roles'),
      dataIndex: 'roles',
      render: (roles: AdminRow['roles']) =>
        roles?.length ? (
          <Space size={[4, 4]} wrap>
            {roles.map((role) => (
              <Tag key={String(role.id)}>{role.name}</Tag>
            ))}
          </Space>
        ) : (
          '-'
        ),
    },
    {
      title: t('admin.google_auth_status', { defaultValue: '谷歌验证' }),
      dataIndex: 'is_2fa_bound',
      width: 110,
      render: (bound: boolean) =>
        bound
          ? t('admin.google_auth_bound', { defaultValue: '已绑定' })
          : t('admin.google_auth_not_bound', { defaultValue: '未绑定' }),
    },
    {
      title: t('table.status'),
      dataIndex: 'status',
      width: 100,
      render: (status: number, row) => (
        <Switch
          checked={status === 1}
          disabled={protectedIds.has(Number(row.id)) || getButtonState('admin.update').disabled}
          onChange={(checked) => void handleStatusChange(row, checked)}
        />
      ),
    },
    { title: t('table.created_at'), dataIndex: 'created_at', width: 180, sorter: true },
    {
      title: t('common.operation'),
      key: 'operation',
      width: 220,
      fixed: 'right',
      render: (_, row) => (
        <Space>
          {getButtonState('admin.update').show && (
            <PermissionButton
              permission="admin.update"
              type="link"
              onClick={() => {
                setEditId(row.id)
                setFormOpen(true)
              }}
            >
              {t('common.edit')}
            </PermissionButton>
          )}
          {getButtonState('admin.destroy').show && !protectedIds.has(Number(row.id)) && (
            <PermissionButton
              permission="admin.destroy"
              type="link"
              danger
              onClick={() => confirmDelete(row.id, row.username)}
            >
              {t('common.delete')}
            </PermissionButton>
          )}
          <Dropdown menu={{ items: moreMenu(row) }}>
            <a onClick={(e) => e.preventDefault()}>
              {t('common.more', { defaultValue: '更多' })} <DownOutlined />
            </a>
          </Dropdown>
        </Space>
      ),
    },
  ]

  const handleExport = async () => {
    setExporting(true)
    try {
      await exportAdmin(searchForm)
      message.success(t('common.queued', { defaultValue: '导出任务已提交，请稍后查看导出记录' }))
    } catch (error) {
      showError(error, t('common.operation_failed'))
    } finally {
      setExporting(false)
    }
  }

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
  } = useColumnSetting('admin', columns)

  return (
    <PageContainer
      title={t('menu.admin_management')}
      extra={
        <Space>
          {toolbar}
          <PermissionButton permission="admin.export" loading={exporting} onClick={() => void handleExport()}>
            {t('common.export', { defaultValue: '导出' })}
          </PermissionButton>
          <Button icon={<SettingOutlined />} onClick={openColumnSetting}>
            {t('common.column_setting')}
          </Button>
        </Space>
      }
    >
      <SearchForm
        fields={[
          { name: 'username', label: t('table.username') },
          {
            name: 'status',
            label: t('table.status'),
            type: 'select',
            options: [
              { label: t('common.enabled'), value: 1 },
              { label: t('common.disabled'), value: 0 },
            ],
          },
          {
            name: 'is_2fa_bound',
            label: t('admin.google_auth_status', { defaultValue: '谷歌验证' }),
            type: 'select',
            options: [
              { label: t('admin.google_auth_bound', { defaultValue: '已绑定' }), value: '1' },
              { label: t('admin.google_auth_not_bound', { defaultValue: '未绑定' }), value: '0' },
            ],
          },
        ]}
        values={searchForm}
        onChange={(values) => setSearchForm(values as typeof searchForm)}
        onSearch={handleSearch}
        onReset={handleReset}
      />
      <Table<AdminRow>
        rowKey="id"
        loading={loading}
        columns={filteredColumns}
        dataSource={tableData}
        scroll={{ x: 1400 }}
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
      <AdminFormModal
        open={formOpen}
        editId={editId}
        onClose={() => setFormOpen(false)}
        onSuccess={() => void refresh()}
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
