import { useEffect, useMemo, useState } from 'react'
import { App, Button, Form, Input, InputNumber, Modal, Radio, Space, Table, TreeSelect } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useTranslation } from 'react-i18next'
import { SettingOutlined } from '@ant-design/icons'
import {
  createDepartment,
  deleteDepartment,
  getDepartmentDetail,
  getDepartmentList,
  updateDepartment,
} from '@/api/department'
import { useCrudActions } from '@/hooks/useCrudActions'
import { usePermission } from '@/hooks/usePermission'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { useColumnSetting } from '@/hooks/useColumnSetting'
import PageContainer from '@/components/PageContainer'
import ColumnSettingDialog from '@/components/ColumnSettingDialog'
import SearchForm from '@/components/SearchForm'
import StatusTag from '@/components/StatusTag'
import PermissionButton from '@/components/PermissionButton'
import { entityField, normalizeTreeList } from '@/utils/normalize'
import { excludeNodeAndChildren, flattenTree } from '@/utils/tree'

interface DepartmentRow {
  id: number | string
  name?: string
  parent_id?: number | string
  description?: string
  sort?: number
  status?: number
  created_at?: string
  children?: DepartmentRow[]
}

type TreeSelectNode = { title: string; value: number | string; children?: TreeSelectNode[] }

function mapRows(list: unknown[]): DepartmentRow[] {
  return normalizeTreeList(list as Record<string, unknown>[]).map((item) => {
    const row = item as Record<string, unknown>
    const children = Array.isArray(row.children) ? mapRows(row.children as unknown[]) : undefined
    return {
      id: entityField(row, 'id', '')!,
      name: String(entityField(row, 'name', '') ?? ''),
      parent_id: entityField(row, 'parent_id', 0) as number | string,
      description: String(entityField(row, 'description', '') ?? entityField(row, 'remark', '') ?? ''),
      sort: Number(entityField(row, 'sort', 0) ?? 0),
      status: Number(entityField(row, 'status', 1) ?? 1),
      created_at: String(entityField(row, 'created_at', '') ?? ''),
      children: children?.length ? children : undefined,
    }
  })
}

function toTreeSelect(items: DepartmentRow[]): TreeSelectNode[] {
  return items.map((item) => ({
    title: String(item.name || item.id),
    value: item.id,
    children: item.children?.length ? toTreeSelect(item.children) : undefined,
  }))
}

function collectIds(items: DepartmentRow[]): Array<string | number> {
  return flattenTree(items).map((node) => (node as DepartmentRow).id)
}

export default function DepartmentList() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const { getButtonState } = usePermission()
  const [form] = Form.useForm()
  const [data, setData] = useState<DepartmentRow[]>([])
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<string | number | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [expandedKeys, setExpandedKeys] = useState<Array<string | number>>([])
  const [searchForm, setSearchForm] = useState<Record<string, unknown>>({ name: '', status: '' })

  const load = async (params: Record<string, unknown> = searchForm) => {
    setLoading(true)
    try {
      const res = await getDepartmentList({ ...params, page_size: 1000 })
      const list = (res.data?.list ?? res.data?.data ?? res.data ?? []) as unknown[]
      const rows = mapRows(Array.isArray(list) ? list : [])
      setData(rows)
      setExpandedKeys(collectIds(rows))
    } catch (error) {
      showError(error, t('common.query_failed'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const { toolbar, confirmDelete } = useCrudActions({
    createPermission: 'department.store',
    onRefresh: () => void load(),
    onCreate: () => {
      setEditId(null)
      form.setFieldsValue({
        name: '',
        parent_id: 0,
        description: '',
        status: 1,
        sort: 0,
      })
      setOpen(true)
    },
    deleteApi: deleteDepartment,
  })

  const parentTreeData = useMemo(() => {
    const filtered = editId ? excludeNodeAndChildren(data, editId) : data
    return [{ title: t('department.top_department'), value: 0 }, ...toTreeSelect(filtered)]
  }, [data, editId, t])

  const allExpanded = expandedKeys.length > 0

  const columns: ColumnsType<DepartmentRow> = [
    { title: t('common.name'), dataIndex: 'name', width: 220 },
    { title: t('common.description'), dataIndex: 'description', ellipsis: true },
    { title: t('common.sort'), dataIndex: 'sort', width: 80 },
    {
      title: t('common.status'),
      dataIndex: 'status',
      width: 100,
      render: (status: number) => <StatusTag status={status} />,
    },
    { title: t('table.created_at'), dataIndex: 'created_at', width: 180 },
    {
      title: t('common.operation'),
      key: 'operation',
      width: 160,
      fixed: 'right',
      render: (_, row) => (
        <Space>
          {getButtonState('department.update').show && (
            <PermissionButton
              permission="department.update"
              type="link"
              onClick={() => {
                void (async () => {
                  setEditId(row.id)
                  setOpen(true)
                  try {
                    const res = await getDepartmentDetail(row.id)
                    const detail = (res.data || {}) as Record<string, unknown>
                    form.setFieldsValue({
                      name: entityField(detail, 'name', row.name),
                      parent_id: Number(entityField(detail, 'parent_id', row.parent_id) || 0),
                      description: entityField(detail, 'description', row.description),
                      status: Number(entityField(detail, 'status', row.status)),
                      sort: Number(entityField(detail, 'sort', row.sort)),
                    })
                  } catch (error) {
                    showError(error, t('common.query_failed'))
                    form.setFieldsValue({
                      name: row.name,
                      parent_id: Number(row.parent_id || 0),
                      description: row.description,
                      status: row.status,
                      sort: row.sort,
                    })
                  }
                })()
              }}
            >
              {t('common.edit')}
            </PermissionButton>
          )}
          {getButtonState('department.destroy').show && (
            <PermissionButton
              permission="department.destroy"
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
  } = useColumnSetting('department', columns)

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      const payload = {
        ...values,
        parent_id: Number(values.parent_id ?? 0),
        status: Number(values.status),
        sort: Number(values.sort || 0),
      }
      if (editId) {
        await updateDepartment(editId, payload)
        message.success(t('common.update_success'))
      } else {
        await createDepartment(payload)
        message.success(t('common.create_success'))
      }
      setOpen(false)
      await load()
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <PageContainer
      title={t('menu.department_management')}
      extra={
        <Space>
          <Button
            onClick={() => setExpandedKeys(allExpanded ? [] : collectIds(data))}
          >
            {allExpanded ? t('common.collapse') : t('common.expand')}
          </Button>
          {toolbar}
          <Button icon={<SettingOutlined />} onClick={openColumnSetting}>
            {t('common.column_setting')}
          </Button>
        </Space>
      }
    >
      <SearchForm
        fields={[
          { name: 'name', label: t('common.name') },
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
        onChange={setSearchForm}
        onSearch={() => void load(searchForm)}
        onReset={() => {
          const reset = { name: '', status: '' }
          setSearchForm(reset)
          void load(reset)
        }}
      />
      <Table<DepartmentRow>
        rowKey="id"
        loading={loading}
        columns={filteredColumns}
        dataSource={data}
        pagination={false}
        scroll={{ x: 1000 }}
        expandable={{
          expandedRowKeys: expandedKeys,
          onExpandedRowsChange: (keys) => setExpandedKeys(keys as Array<string | number>),
        }}
      />
      <Modal
        open={open}
        title={editId ? t('department.edit') : t('department.add')}
        onCancel={() => setOpen(false)}
        onOk={() => void handleSubmit()}
        confirmLoading={submitting}
        destroyOnHidden
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label={t('common.name')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="parent_id" label={t('department.parent_department')}>
            <TreeSelect
              allowClear
              treeDefaultExpandAll
              treeData={parentTreeData}
              placeholder={t('department.parent_department')}
            />
          </Form.Item>
          <Form.Item name="description" label={t('common.description')}>
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="sort" label={t('common.sort')}>
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item name="status" label={t('common.status')}>
            <Radio.Group
              options={[
                { label: t('common.enabled'), value: 1 },
                { label: t('common.disabled'), value: 0 },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
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
