import { useState } from 'react'
import { Space, Table } from 'antd'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import { useTranslation } from 'react-i18next'
import { deleteRole, getRoleList } from '@/api/role'
import { useListPage } from '@/hooks/useListPage'
import { useCrudActions } from '@/hooks/useCrudActions'
import { usePermission } from '@/hooks/usePermission'
import PageContainer from '@/components/PageContainer'
import SearchForm from '@/components/SearchForm'
import StatusTag from '@/components/StatusTag'
import PermissionButton from '@/components/PermissionButton'
import { entityField } from '@/utils/normalize'
import RoleFormModal from './RoleFormModal'

interface RoleRow {
  id: number | string
  name?: string
  slug?: string
  description?: string
  status?: number
  sort?: number
  created_at?: string
}

export default function RoleList() {
  const { t } = useTranslation()
  const { getButtonState } = usePermission()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<string | number | null>(null)

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
  } = useListPage<RoleRow>({
    fetchApi: getRoleList as never,
    initialSearchForm: { name: '', status: '' },
    normalizeRows: true,
    transformData: (row) => {
      const record = row as unknown as Record<string, unknown>
      return {
        id: entityField(record, 'id', '')!,
        name: String(entityField(record, 'name', '') ?? ''),
        slug: String(entityField(record, 'slug', '') ?? ''),
        description: String(entityField(record, 'description', '') ?? ''),
        status: Number(entityField(record, 'status', 0) ?? 0),
        sort: Number(entityField(record, 'sort', 0) ?? 0),
        created_at: String(entityField(record, 'created_at', '') ?? ''),
      }
    },
  })

  const { toolbar, confirmDelete } = useCrudActions({
    createPermission: 'role.store',
    onRefresh: refresh,
    onCreate: () => {
      setEditId(null)
      setOpen(true)
    },
    deleteApi: deleteRole,
  })

  const columns: ColumnsType<RoleRow> = [
    { title: t('table.id'), dataIndex: 'id', width: 80, sorter: true },
    { title: t('role.name'), dataIndex: 'name' },
    { title: t('role.slug'), dataIndex: 'slug' },
    { title: t('common.description'), dataIndex: 'description', ellipsis: true },
    { title: t('common.sort'), dataIndex: 'sort', width: 80, sorter: true },
    {
      title: t('common.status'),
      dataIndex: 'status',
      width: 100,
      render: (status: number) => <StatusTag status={status} />,
    },
    { title: t('table.created_at'), dataIndex: 'created_at', width: 180, sorter: true },
    {
      title: t('common.operation'),
      key: 'operation',
      width: 160,
      fixed: 'right',
      render: (_, row) => (
        <Space>
          {getButtonState('role.update').show && (
            <PermissionButton
              permission="role.update"
              type="link"
              onClick={() => {
                setEditId(row.id)
                setOpen(true)
              }}
            >
              {t('common.edit')}
            </PermissionButton>
          )}
          {getButtonState('role.destroy').show && row.slug !== 'super-admin' && (
            <PermissionButton
              permission="role.destroy"
              type="link"
              danger
              onClick={() => confirmDelete(row.id, row.name)}
            >
              {t('common.delete')}
            </PermissionButton>
          )}
        </Space>
      ),
    },
  ]

  return (
    <PageContainer title={t('menu.role_management')} extra={toolbar}>
      <SearchForm
        fields={[
          { name: 'name', label: t('role.name') },
          {
            name: 'status',
            label: t('common.status'),
            type: 'select',
            options: [
              { label: t('common.enabled'), value: 1 },
              { label: t('common.disabled'), value: 0 },
            ],
          },
        ]}
        values={searchForm}
        onChange={(values) => setSearchForm(values as never)}
        onSearch={handleSearch}
        onReset={handleReset}
      />
      <Table<RoleRow>
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={tableData}
        scroll={{ x: 1000 }}
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
      <RoleFormModal
        open={open}
        editId={editId}
        onClose={() => setOpen(false)}
        onSuccess={() => void refresh()}
      />
    </PageContainer>
  )
}
