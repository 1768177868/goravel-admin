import { useEffect, useMemo, useState } from 'react'
import { App, Form, Input, Modal, TreeSelect<<if .HasFormNumber>>, InputNumber<<end>><<if .HasFormSelect>>, Select<<end>><<if .HasFormSwitch>>, Switch<<end>><<if .HasFormRadio>>, Radio<<end>><<if .HasFormCheckbox>>, Checkbox<<end>><<if .HasFormDatePicker>>, DatePicker<<end>> } from 'antd'
import { useTranslation } from 'react-i18next'
import {
  <<if .HasCreate>>create<<.ModelName>>,<<end>>
  get<<.ModelName>>Detail,
  <<if .HasEdit>>update<<.ModelName>>,<<end>>
} from '@/api/<<.ModuleNameK>>'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { entityField } from '@/utils/normalize'
import { excludeNodeAndChildren } from '@/utils/tree'
<<if .HasEditor>>
import WangEditor from '@/components/WangEditor'
<<end>>
<<if .HasMarkdown>>
import MarkdownEditor from '@/components/MarkdownEditor'
<<end>>
import {
  to<<.ModelName>>TreeSelect,
  type <<.ModelName>>Row,
} from './<<.ModuleNameCamel>>.config'

interface <<.ModelName>>FormModalProps {
  open: boolean
  editId?: string | number | null
  treeData: <<.ModelName>>Row[]
  onClose: () => void
  onSuccess: () => void
}

export default function <<.ModelName>>FormModal({
  open,
  editId,
  treeData,
  onClose,
  onSuccess,
}: <<.ModelName>>FormModalProps) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const parentTreeData = useMemo(() => {
    const filtered = editId ? excludeNodeAndChildren(treeData, editId) : treeData
    return [
      { title: t('<<.ModuleName>>.top_parent', { defaultValue: t('common.top_level', { defaultValue: 'Top Level' }) }), value: 0 },
      ...to<<.ModelName>>TreeSelect(filtered),
    ]
  }, [editId, treeData, t])

  useEffect(() => {
    if (!open) return

    if (!editId) {
      form.setFieldsValue({
<<range .FormFields>>
<<- if and .ShowInForm (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
        <<.Name>>: <<if eq .Name "parent_id">>0<<else if eq .FormType "switch">>true<<else if eq .FormType "number">>0<<else if eq .FormType "checkbox">>[]<<else>>''<<end>>,
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
          <<.Name>>: <<if eq .Name "parent_id">>Number(entityField(data, '<<.Name>>', 0))<<else if eq .FormType "switch">>(Number(entityField(data, '<<.Name>>', 1)) === 1)<<else if eq .FormType "number">>Number(entityField(data, '<<.Name>>', 0))<<else>>entityField(data, '<<.Name>>', '')<<end>>,
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
      const payload = {
        ...values,
<<range .FormFields>>
<<- if and .ShowInForm (eq .Name "parent_id")>>
        parent_id: Number(values.parent_id ?? 0),
<<- end>>
<<- if and .ShowInForm (eq .FormType "switch")>>
        <<.Name>>: values.<<.Name>> ? 1 : 0,
<<- end>>
<<- end>>
      }
      if (editId) {
        <<if .HasEdit>>
        await update<<.ModelName>>(editId, payload)
        message.success(t('common.update_success'))
        <<end>>
      } else {
        <<if .HasCreate>>
        await create<<.ModelName>>(payload)
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
      width={<<if or .HasEditor .HasMarkdown>>800<<else>>640<<end>>}
    >
      <Form form={form} layout="vertical" disabled={loading}>
<<range .FormFields>>
<<- if and .ShowInForm (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
        <<if eq .Name "parent_id">>
        <Form.Item name="parent_id" label={t('<<.Name>>', { defaultValue: '<<.Label>>' })}>
          <TreeSelect
            allowClear
            treeDefaultExpandAll
            treeData={parentTreeData}
            placeholder={t('<<.Name>>', { defaultValue: '<<.Label>>' })}
          />
        </Form.Item>
        <<else if eq .FormType "editor">>
        <Form.Item name="<<.Name>>" label={t('<<.Name>>', { defaultValue: '<<.Label>>' })}<<if .Required>> rules={[{ required: true }]}<<end>>>
          <WangEditor height={400} placeholder={t('<<.Name>>', { defaultValue: '<<.Label>>' })} />
        </Form.Item>
        <<else if eq .FormType "markdown">>
        <Form.Item name="<<.Name>>" label={t('<<.Name>>', { defaultValue: '<<.Label>>' })}<<if .Required>> rules={[{ required: true }]}<<end>>>
          <MarkdownEditor height={400} placeholder={t('<<.Name>>', { defaultValue: '<<.Label>>' })} />
        </Form.Item>
        <<else if eq .FormType "textarea">>
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
