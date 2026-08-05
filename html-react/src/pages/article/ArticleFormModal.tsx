import { useEffect, useState } from 'react'
import { App, Form, Input, InputNumber, Modal, Spin, Switch } from 'antd'
import { useTranslation } from 'react-i18next'
import { createArticle, getArticleDetail, updateArticle } from '@/api/article'
import WangEditor from '@/components/WangEditor'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { entityField } from '@/utils/normalize'

interface ArticleFormModalProps {
  open: boolean
  editId?: string | number | null
  onClose: () => void
  onSuccess: () => void
}

export default function ArticleFormModal({ open, editId, onClose, onSuccess }: ArticleFormModalProps) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [content, setContent] = useState('')

  useEffect(() => {
    if (!open) return

    let cancelled = false
    const boot = async () => {
      if (!editId) {
        form.setFieldsValue({ admin_id: undefined, title: '', status: 1 })
        setContent('')
        return
      }

      setLoading(true)
      try {
        const res = await getArticleDetail(editId)
        if (cancelled) return
        const raw = (res.data || {}) as Record<string, unknown>
        const data = (entityField(raw, 'article', raw) || {}) as Record<string, unknown>
        form.setFieldsValue({
          admin_id: entityField(data, 'admin_id', undefined),
          title: entityField(data, 'title', ''),
          status: Number(entityField(data, 'status', 1)),
        })
        setContent(String(entityField(data, 'content', '') ?? ''))
      } catch (error) {
        showError(error, t('common.query_failed'))
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void boot()
    return () => {
      cancelled = true
    }
  }, [open, editId, form, showError, t])

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      const payload: Record<string, unknown> = {
        admin_id: values.admin_id,
        title: values.title,
        content,
        status: Number(values.status),
      }

      if (editId) {
        await updateArticle(editId, payload)
        message.success(t('common.update_success'))
      } else {
        await createArticle(payload)
        message.success(t('common.create_success'))
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
      title={editId ? t('article.edit_article') : t('article.add_article')}
      onCancel={onClose}
      onOk={() => void handleOk()}
      confirmLoading={submitting}
      width={900}
      destroyOnHidden
      styles={{ body: { maxHeight: '75vh', overflow: 'auto' } }}
    >
      <Spin spinning={loading}>
        <Form form={form} layout="vertical">
          <Form.Item name="admin_id" label={t('admin_id')} rules={[{ required: true }]}>
            <InputNumber min={1} style={{ width: '100%' }} placeholder={t('admin_id')} />
          </Form.Item>
          <Form.Item name="title" label={t('title')} rules={[{ required: true }]}>
            <Input placeholder={t('title')} />
          </Form.Item>
          <Form.Item label={t('content')}>
            <WangEditor value={content} onChange={setContent} placeholder={t('content_placeholder')} height={400} />
          </Form.Item>
          <Form.Item
            name="status"
            label={t('common.status')}
            getValueProps={(value) => ({ checked: Number(value) === 1 })}
            getValueFromEvent={(checked: boolean) => (checked ? 1 : 0)}
          >
            <Switch checkedChildren={t('common.enabled')} unCheckedChildren={t('common.disabled')} />
          </Form.Item>
        </Form>
      </Spin>
    </Modal>
  )
}
