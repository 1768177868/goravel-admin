import { useEffect, useState } from 'react'
import { App, Form, Input, Modal, Radio, Select } from 'antd'
import { useTranslation } from 'react-i18next'
import { createNotification } from '@/api/notification'
import MarkdownEditor from '@/components/MarkdownEditor'
import { useOptions } from '@/hooks/useOptions'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { useNotificationStore } from '@/stores/notification'

interface NotificationFormModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

export default function NotificationFormModal({ open, onClose, onSuccess }: NotificationFormModalProps) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const refreshBell = useNotificationStore((s) => s.refresh)
  const [form] = Form.useForm()
  const [submitting, setSubmitting] = useState(false)
  const notifyType = Form.useWatch('type', form)
  const { selectOptions: adminOptions, loading: adminLoading } = useOptions(
    'admin',
    open && notifyType === 'message',
  )

  useEffect(() => {
    if (!open) return
    form.setFieldsValue({
      type: 'announcement',
      receiver_id: undefined,
      title: '',
      content: '',
    })
  }, [open, form])

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      const payload: Record<string, unknown> = {
        type: values.type,
        title: String(values.title || '').trim(),
        content: values.content,
      }
      if (values.type === 'message') {
        if (!values.receiver_id) {
          message.error(t('notification.receiver_required'))
          return
        }
        payload.receiver_id = values.receiver_id
      }
      await createNotification(payload)
      message.success(t('notification.create_success'))
      await refreshBell()
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
      title={t('notification.create')}
      onCancel={onClose}
      onOk={() => void handleOk()}
      confirmLoading={submitting}
      destroyOnHidden
      width={800}
    >
      <Form form={form} layout="vertical">
        <Form.Item name="type" label={t('notification.table.type')} rules={[{ required: true }]}>
          <Radio.Group
            options={[
              { label: t('notification.types.announcement'), value: 'announcement' },
              { label: t('notification.types.notice'), value: 'notice' },
              { label: t('notification.types.message'), value: 'message' },
            ]}
          />
        </Form.Item>
        {notifyType === 'message' && (
          <Form.Item
            name="receiver_id"
            label={t('notification.receiver')}
            rules={[{ required: true, message: t('notification.receiver_required') }]}
          >
            <Select
              allowClear
              showSearch
              loading={adminLoading}
              options={adminOptions}
              optionFilterProp="label"
              placeholder={t('notification.select_receiver')}
            />
          </Form.Item>
        )}
        <Form.Item
          name="title"
          label={t('notification.table.title')}
          rules={[
            { required: true, message: t('notification.title_required') },
            { max: 150, message: t('notification.title_max_length') },
          ]}
        >
          <Input maxLength={150} showCount placeholder={t('notification.title_placeholder')} />
        </Form.Item>
        <Form.Item
          name="content"
          label={t('notification.table.content')}
          rules={[{ required: true, message: t('notification.content_required') }]}
        >
          <MarkdownEditor height={400} placeholder={t('notification.content_placeholder')} />
        </Form.Item>
      </Form>
    </Modal>
  )
}
