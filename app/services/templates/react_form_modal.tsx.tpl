import { useEffect, useState } from 'react'
import { App, Form, Input, InputNumber, Modal, Radio, Select, Switch } from 'antd'
import { useTranslation } from 'react-i18next'
import {
  <<if .HasCreate>>create<<.ModelName>>,<<end>>
  get<<.ModelName>>Detail,
  <<if .HasEdit>>update<<.ModelName>>,<<end>>
} from '@/api/<<.ModuleNameK>>'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { entityField } from '@/utils/normalize'

interface <<.ModelName>>FormModalProps {
  open: boolean
  editId?: string | number | null
  onClose: () => void
  onSuccess: () => void
}

export default function <<.ModelName>>FormModal({ open, editId, onClose, onSuccess }: <<.ModelName>>FormModalProps) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!open) return

    if (!editId) {
      form.setFieldsValue({
<<range .FormFields>>
<<- if and .ShowInForm (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
        <<.Name>>: <<if eq .FormType "switch">>1<<else if eq .FormType "number">>0<<else if eq .FormType "checkbox">>[]<<else>>''<<end>>,
<<- end>>
<<- end>>
      })
      return
    }

    setLoading(true)
    get<<.ModelName>>Detail(editId)
      .then((res) => {
        const raw = (res.data || {}) as Record<string, unknown>
        const data = (entityField(raw, '<<.ModuleName>>', raw) || {}) as Record<string, unknown>
        form.setFieldsValue({
<<range .FormFields>>
<<- if and .ShowInForm (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
          <<.Name>>: <<if eq .FormType "switch">>Number(entityField(data, '<<.Name>>', 1))<<else if eq .FormType "number">>Number(entityField(data, '<<.Name>>', 0))<<else>>entityField(data, '<<.Name>>', '')<<end>>,
<<- end>>
<<- end>>
        })
      })
      .catch((error) => showError(error, t('common.query_failed')))
      .finally(() => setLoading(false))
  }, [open, editId, form, showError, t])

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      if (editId) {
        <<if .HasEdit>>
        await update<<.ModelName>>(editId, values)
        message.success(t('common.update_success'))
        <<end>>
      } else {
        <<if .HasCreate>>
        await create<<.ModelName>>(values)
        message.success(t('common.create_success'))
        <<end>>
      }
      onSuccess()
      onClose()
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal
      open={open}
      title={editId ? t('<<.ModuleName>>.edit', { defaultValue: t('common.edit') }) : t('<<.ModuleName>>.add', { defaultValue: t('common.add') })}
      onCancel={onClose}
      onOk={() => void handleOk()}
      confirmLoading={submitting}
      destroyOnHidden
      width={640}
    >
      <Form form={form} layout="vertical" disabled={loading}>
<<range .FormFields>>
<<- if and .ShowInForm (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
        <<if eq .FormType "textarea">>
        <Form.Item name="<<.Name>>" label={t('<<.Name>>', { defaultValue: '<<.Label>>' })}<<if .Required>> rules={[{ required: true }]}<<end>>>
          <Input.TextArea rows={3} />
        </Form.Item>
        <<else if eq .FormType "number">>
        <Form.Item name="<<.Name>>" label={t('<<.Name>>', { defaultValue: '<<.Label>>' })}<<if .Required>> rules={[{ required: true }]}<<end>>>
          <InputNumber style={{ width: '100%' }} />
        </Form.Item>
        <<else if eq .FormType "switch">>
        <Form.Item name="<<.Name>>" label={t('<<.Name>>', { defaultValue: '<<.Label>>' })} valuePropName="checked">
          <Switch />
        </Form.Item>
        <<else if eq .FormType "select">>
        <Form.Item name="<<.Name>>" label={t('<<.Name>>', { defaultValue: '<<.Label>>' })}<<if .Required>> rules={[{ required: true }]}<<end>>>
          <Select allowClear />
        </Form.Item>
        <<else>>
        <Form.Item name="<<.Name>>" label={t('<<.Name>>', { defaultValue: '<<.Label>>' })}<<if .Required>> rules={[{ required: true }]}<<end>>>
          <Input />
        </Form.Item>
        <<end>>
<<- end>>
<<- end>>
      </Form>
    </Modal>
  )
}
