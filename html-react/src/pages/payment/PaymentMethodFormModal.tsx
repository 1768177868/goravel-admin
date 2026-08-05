import { useEffect, useMemo, useState } from 'react'
import { App, Divider, Form, Input, InputNumber, Modal, Select, Switch } from 'antd'
import { useTranslation } from 'react-i18next'
import { createPaymentMethod, getPaymentMethodDetail, updatePaymentMethod } from '@/api/paymentMethod'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { entityField } from '@/utils/normalize'
import {
  createEmptyConfig,
  createPaymentMethodTypeOptions,
  getConfigFieldsForType,
  type PaymentMethodType,
} from './paymentMethod.config'

interface PaymentMethodFormModalProps {
  open: boolean
  editId?: string | number | null
  onClose: () => void
  onSuccess: () => void
}

interface FormValues {
  name: string
  code?: string
  type?: PaymentMethodType
  is_active: boolean
  sort: number
  description?: string
  config: Record<string, string>
  config_json?: string
}

export default function PaymentMethodFormModal({
  open,
  editId,
  onClose,
  onSuccess,
}: PaymentMethodFormModalProps) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const [form] = Form.useForm<FormValues>()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const isEdit = Boolean(editId)

  const typeValue = Form.useWatch('type', form)
  const typeOptions = useMemo(() => createPaymentMethodTypeOptions(t), [t])
  const configFields = useMemo(() => getConfigFieldsForType(typeValue), [typeValue])

  useEffect(() => {
    if (!open) return

    if (!editId) {
      form.setFieldsValue({
        name: '',
        code: '',
        type: undefined,
        is_active: true,
        sort: 0,
        description: '',
        config: {},
        config_json: '',
      })
      return
    }

    setLoading(true)
    getPaymentMethodDetail(editId)
      .then((res) => {
        const data = (res.data || {}) as Record<string, unknown>
        const type = String(entityField(data, 'type', '') ?? '') as PaymentMethodType
        const configData = (entityField(data, 'config', {}) || {}) as Record<string, unknown>
        const fields = getConfigFieldsForType(type)
        const config: Record<string, string> = {}

        fields.forEach((field) => {
          const value = configData[field.key]
          config[field.key] =
            value !== undefined && value !== null ? String(value) : ''
        })

        const hasUnknownConfig = Object.keys(configData).length > 0 && fields.length === 0

        form.setFieldsValue({
          name: String(entityField(data, 'name', '') ?? ''),
          code: String(entityField(data, 'code', '') ?? ''),
          type: type || undefined,
          is_active: Boolean(entityField(data, 'is_active', true)),
          sort: Number(entityField(data, 'sort', 0) ?? 0),
          description: String(entityField(data, 'description', '') ?? ''),
          config,
          config_json: hasUnknownConfig ? JSON.stringify(configData, null, 2) : '',
        })
      })
      .catch((error) => showError(error, t('common.query_failed')))
      .finally(() => setLoading(false))
  }, [open, editId, form, showError, t])

  const handleTypeChange = (type?: PaymentMethodType) => {
    form.setFieldValue('config', createEmptyConfig(type))
    form.setFieldValue('config_json', '')
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      let config: Record<string, string> = {}
      const fields = getConfigFieldsForType(values.type)

      if (fields.length > 0) {
        fields.forEach((field) => {
          const value = values.config?.[field.key]
          if (value !== undefined && value !== null && value !== '') {
            config[field.key] = value
          }
        })
      } else if (values.config_json?.trim()) {
        try {
          config = JSON.parse(values.config_json.trim()) as Record<string, string>
        } catch {
          message.error(t('payment_method.config_invalid'))
          return
        }
      }

      if (Object.keys(config).length === 0) {
        message.error(t('payment_method.config_required'))
        return
      }

      setSubmitting(true)
      const payload: Record<string, unknown> = {
        name: values.name.trim(),
        is_active: values.is_active,
        sort: values.sort ?? 0,
        config,
        description: values.description?.trim() || '',
      }

      if (isEdit) {
        await updatePaymentMethod(editId!, payload)
        message.success(t('payment_method.update_success'))
      } else {
        await createPaymentMethod({
          ...payload,
          code: values.code?.trim(),
          type: values.type,
        })
        message.success(t('payment_method.create_success'))
      }

      onClose()
      onSuccess()
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
      title={isEdit ? t('payment_method.edit_payment_method') : t('payment_method.add_payment_method')}
      width={800}
      confirmLoading={submitting}
      onCancel={onClose}
      onOk={() => void handleSubmit()}
      destroyOnClose
    >
      <Form form={form} layout="vertical" disabled={loading}>
        <Form.Item
          name="name"
          label={t('payment_method.name')}
          rules={[{ required: true, message: t('payment_method.name_required') }]}
        >
          <Input placeholder={t('payment_method.name_placeholder')} />
        </Form.Item>

        {!isEdit ? (
          <>
            <Form.Item
              name="code"
              label={t('payment_method.code')}
              rules={[{ required: true, message: t('payment_method.code_required') }]}
            >
              <Input placeholder={t('payment_method.code_placeholder')} />
            </Form.Item>
            <Form.Item
              name="type"
              label={t('payment_method.type')}
              rules={[{ required: true, message: t('payment_method.type_required') }]}
            >
              <Select
                allowClear
                placeholder={t('payment_method.type_placeholder')}
                options={typeOptions}
                onChange={handleTypeChange}
              />
            </Form.Item>
          </>
        ) : null}

        <Form.Item name="is_active" label={t('table.status')} valuePropName="checked">
          <Switch checkedChildren={t('common.enabled')} unCheckedChildren={t('common.disabled')} />
        </Form.Item>

        <Form.Item name="sort" label={t('table.sort')}>
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="description" label={t('table.description')}>
          <Input.TextArea rows={3} placeholder={t('payment_method.description_placeholder')} />
        </Form.Item>

        {(typeValue || isEdit) && configFields.length > 0 ? (
          <>
            <Divider plain>{t('payment_method.config')}</Divider>
            {configFields.map((field) => (
              <Form.Item
                key={field.key}
                name={['config', field.key]}
                label={field.label}
                rules={
                  field.required
                    ? [{ required: true, message: t('payment_method.config_field_required', { field: field.label }) }]
                    : undefined
                }
              >
                {field.type === 'textarea' ? (
                  <Input.TextArea rows={field.rows || 3} placeholder={field.placeholder} />
                ) : (
                  <Input
                    type={field.inputType === 'password' ? 'password' : 'text'}
                    placeholder={field.placeholder}
                  />
                )}
              </Form.Item>
            ))}
          </>
        ) : (typeValue || isEdit) && configFields.length === 0 ? (
          <Form.Item
            name="config_json"
            label={t('payment_method.config')}
            rules={[{ required: true, message: t('payment_method.config_required') }]}
          >
            <Input.TextArea rows={10} placeholder={t('payment_method.config_placeholder')} />
          </Form.Item>
        ) : !typeValue && !isEdit ? (
          <Form.Item name="config_json" label={t('payment_method.config')}>
            <Input.TextArea rows={10} placeholder={t('payment_method.config_placeholder')} />
          </Form.Item>
        ) : null}
      </Form>
    </Modal>
  )
}
