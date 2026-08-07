import { useState } from 'react'
import { App, Button, Dropdown, Form, Input, Space, Switch, Table, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
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
import { handlePaginatedTableChange } from '@/utils/tableChange'
import { useCrudActions } from '@/hooks/useCrudActions'
import { usePermission } from '@/hooks/usePermission'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { useColumnSetting } from '@/hooks/useColumnSetting'
import PageContainer from '@/components/PageContainer'
import SearchForm from '@/components/SearchForm'
import PermissionButton from '@/components/PermissionButton'
import ColumnSettingDialog from '@/components/ColumnSettingDialog'
import AdminFormModal from './AdminFormModal'
import {
  adminInitialSearchForm,
  adminProtectedIds,
  getAdminDeptName,
  getAdminPositionName,
  transformAdminRow,
  type AdminRow,
} from './admin.config'

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
    onSearchFormChange,
    loadData,
    handleSearch,
    handleReset,
    handleSortChange,
    refresh,
  } = useListPage<AdminRow, typeof adminInitialSearchForm>({
    fetchApi: getAdminList,
    initialSearchForm: adminInitialSearchForm,
    defaultSort: 'id:desc',
    normalizeRows: true,
    transformData: transformAdminRow,
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

  const handleStatusChange = async (row: AdminRow, checked: boolean) => {
    if (adminProtectedIds.has(Number(row.id)) && !checked) {
      message.warning(t('admin.protected_cannot_disable', { defaultValue: '�ܱ����˺Ų��ܽ���' }))
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
            label={t('admin.new_password', { defaultValue: '������' })}
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
          message.success(t('admin.reset_password_success', { defaultValue: '����������' }))
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
            label: t('admin.kick_out', { defaultValue: '�߳���¼' }),
            onClick: () =>
              runConfirm(
                t('admin.kick_out', { defaultValue: '�߳���¼' }),
                t('admin.kick_out_confirm', {
                  username: row.username,
                  defaultValue: `ȷ���߳� ${row.username} ��ȫ����¼��`,
                }),
                async () => {
                  await kickOutUser(row.id)
                  message.success(t('admin.kick_out_success', { defaultValue: '���߳�' }))
                },
              ),
          }
        : null,
      row.is_2fa_bound && getButtonState('admin.unbind_google_auth').show
        ? {
            key: 'unbind2fa',
            label: t('admin.unbind_google_auth', { defaultValue: '���ȸ���֤' }),
            onClick: () =>
              runConfirm(
                t('admin.unbind_google_auth', { defaultValue: '���ȸ���֤' }),
                t('admin.unbind_google_auth_confirm', {
                  defaultValue: 'ȷ�����ù���Ա�Ĺȸ���֤����',
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
            label: t('admin.reset_google_auth', { defaultValue: '���ùȸ���֤' }),
            onClick: () =>
              runConfirm(
                t('admin.reset_google_auth', { defaultValue: '���ùȸ���֤' }),
                t('admin.reset_google_auth_confirm', {
                  defaultValue: 'ȷ�����øù���Ա�Ĺȸ���֤����',
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
    { title: t('table.department'), dataIndex: 'department', render: (_, row) => getAdminDeptName(row) },
    { title: t('table.position'), dataIndex: 'position', render: (_, row) => getAdminPositionName(row) },
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
      title: t('admin.google_auth_status', { defaultValue: '�ȸ���֤' }),
      dataIndex: 'is_2fa_bound',
      width: 110,
      render: (bound: boolean) =>
        bound
          ? t('admin.google_auth_bound', { defaultValue: '�Ѱ�' })
          : t('admin.google_auth_not_bound', { defaultValue: 'δ��' }),
    },
    {
      title: t('table.status'),
      dataIndex: 'status',
      width: 100,
      render: (status: number, row) => (
        <Switch
          checked={status === 1}
          disabled={adminProtectedIds.has(Number(row.id)) || getButtonState('admin.update').disabled}
          onChange={(checked) => void handleStatusChange(row, checked)}
        />
      ),
    },
    { title: t('table.created_at'), dataIndex: 'created_at', width: 180, sorter: true },
    {
      title: t('common.operation'),
      key: 'operation',
      width: 220,
      fixed: 'end',
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
          {getButtonState('admin.destroy').show && !adminProtectedIds.has(Number(row.id)) && (
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
              {t('common.more', { defaultValue: '����' })} <DownOutlined />
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
      message.success(t('common.queued', { defaultValue: '�����������ύ�����Ժ�鿴������¼' }))
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
            {t('common.export', { defaultValue: '����' })}
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
            label: t('admin.google_auth_status', { defaultValue: '�ȸ���֤' }),
            type: 'select',
            options: [
              { label: t('admin.google_auth_bound', { defaultValue: '�Ѱ�' }), value: '1' },
              { label: t('admin.google_auth_not_bound', { defaultValue: 'δ��' }), value: '0' },
            ],
          },
        ]}
        values={searchForm}
        onChange={onSearchFormChange}
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
        onChange={(pager, _f, sorter) =>
          handlePaginatedTableChange({ pager, sorter, pagination, loadData, handleSortChange })
        }
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
