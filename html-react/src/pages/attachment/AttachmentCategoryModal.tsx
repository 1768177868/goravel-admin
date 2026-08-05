import { useEffect, useState } from 'react'
import { App, Form, Input, InputNumber, Modal, Space, Switch, Table } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useTranslation } from 'react-i18next'
import {
  createAttachmentCategory,
  deleteAttachmentCategory,
  getAttachmentCategoryList,
  updateAttachmentCategory,
} from '@/api/attachmentCategory'
import PermissionButton from '@/components/PermissionButton'
import StatusTag from '@/components/StatusTag'
import { usePermission } from '@/hooks/usePermission'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { entityField } from '@/utils/normalize'

interface CategoryRow {
  id: number | string
  name?: string
  sort?: number
  status?: number
  remark?: string
  is_system?: number
}

interface AttachmentCategoryModalProps {
  open: boolean
  onClose: () => void
  onChanged?: () => void
}

export default function AttachmentCategoryModal({
  open,
  onClose,
  onChanged,
}: AttachmentCategoryModalProps) {
  const { t } = useTranslation()
  const { message, modal } = App.useApp()
  const showError = useUnhandledError()
  const { getButtonState } = usePermission()
  const [loading, setLoading] = useState(false)
  const [list, setList] = useState<CategoryRow[]>([])
  const [formOpen, setFormOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm()
  const [editing, setEditing] = useState<CategoryRow | null>(null)

  const loadData = async () => {
    setLoading(true)
    try {
      const res = await getAttachmentCategoryList({ page: 1, page_size: 100 })
      const rows = (res.data?.list || []) as Array<Record<string, unknown>>
      setList(
        rows.map((row) => ({
          id: entityField(row, 'id', '')!,
          name: String(entityField(row, 'name', '') ?? ''),
          sort: Number(entityField(row, 'sort', 0) ?? 0),
          status: Number(entityField(row, 'status', 1) ?? 1),
          remark: String(entityField(row, 'remark', '') ?? ''),
          is_system: Number(entityField(row, 'is_system', 0) ?? 0),
        })),
      )
    } catch {
      setList([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (open) void loadData()
  }, [open])

  const openForm = (row?: CategoryRow) => {
    setEditing(row || null)
    form.setFieldsValue({
      name: row?.name || '',
      sort: Number(row?.sort ?? 0),
      status: row ? Number(row.status ?? 1) === 1 : true,
      remark: row?.remark || '',
    })
    setFormOpen(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      const payload = {
        name: values.name,
        sort: values.sort,
        status: values.status ? 1 : 0,
        remark: values.remark || '',
      }
      if (editing?.id) {
        await updateAttachmentCategory(editing.id, payload)
      } else {
        await createAttachmentCategory(payload)
      }
      message.success(t('common.save_success'))
      setFormOpen(false)
      await loadData()
      onChanged?.()
    } catch (error) {
      if (error && typeof error === 'object' && 'errorFields' in error) return
      showError(error, t('common.operation_failed'))
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = (row: CategoryRow) => {
    if (Number(row.is_system) === 1) {
      message.warning(t('attachment.category_system_cannot_delete'))
      return
    }
    modal.confirm({
      title: t('common.confirm'),
      content: t('attachment.category_delete_confirm'),
      okType: 'danger',
      onOk: async () => {
        try {
          await deleteAttachmentCategory(row.id)
          message.success(t('common.delete_success'))
          await loadData()
          onChanged?.()
        } catch (error) {
          showError(error, t('common.operation_failed'))
        }
      },
    })
  }

  const columns: ColumnsType<CategoryRow> = [
    { title: t('table.id'), dataIndex: 'id', width: 70 },
    { title: t('attachment.category_name'), dataIndex: 'name', minWidth: 140 },
    { title: t('common.sort'), dataIndex: 'sort', width: 80 },
    {
      title: t('common.status'),
      dataIndex: 'status',
      width: 90,
      render: (status: number) => <StatusTag status={status} />,
    },
    {
      title: t('common.operation'),
      key: 'operation',
      width: 160,
      fixed: 'right',
      render: (_, row) => (
        <Space>
          {getButtonState('attachment_category.update').show ? (
            <PermissionButton
              permission="attachment_category.update"
              type="link"
              onClick={() => openForm(row)}
            >
              {t('common.edit')}
            </PermissionButton>
          ) : null}
          {getButtonState('attachment_category.destroy').show ? (
            <PermissionButton
              permission="attachment_category.destroy"
              type="link"
              danger
              disabled={Number(row.is_system) === 1}
              onClick={() => handleDelete(row)}
            >
              {t('common.delete')}
            </PermissionButton>
          ) : null}
        </Space>
      ),
    },
  ]

  return (
    <>
      <Modal
        open={open}
        title={t('attachment.category_manage')}
        width={720}
        onCancel={onClose}
        footer={null}
        destroyOnHidden
      >
        <div style={{ marginBottom: 12 }}>
          {getButtonState('attachment_category.store').show ? (
            <PermissionButton
              permission="attachment_category.store"
              type="primary"
              onClick={() => openForm()}
            >
              {t('attachment.category_add')}
            </PermissionButton>
          ) : null}
        </div>
        <Table<CategoryRow>
          rowKey="id"
          size="small"
          loading={loading}
          columns={columns}
          dataSource={list}
          pagination={false}
          scroll={{ x: 600 }}
        />
      </Modal>

      <Modal
        open={formOpen}
        title={editing ? t('attachment.category_edit') : t('attachment.category_add')}
        width={420}
        confirmLoading={submitting}
        onCancel={() => setFormOpen(false)}
        onOk={() => void handleSubmit()}
        destroyOnHidden
      >
        <Form form={form} layout="vertical" initialValues={{ status: true, sort: 0 }}>
          <Form.Item
            name="name"
            label={t('attachment.category_name')}
            rules={[{ required: true, message: t('attachment.category_name_required') }]}
          >
            <Input maxLength={50} />
          </Form.Item>
          <Form.Item name="sort" label={t('common.sort')}>
            <InputNumber min={0} max={9999} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label={t('common.status')} valuePropName="checked">
            <Switch
              checkedChildren={t('common.enabled')}
              unCheckedChildren={t('common.disabled')}
              disabled={Number(editing?.is_system) === 1}
            />
          </Form.Item>
          <Form.Item name="remark" label={t('common.description')}>
            <Input.TextArea rows={2} maxLength={500} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
