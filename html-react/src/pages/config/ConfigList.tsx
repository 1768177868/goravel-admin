import { useCallback, useEffect, useState } from 'react'
import { App, Card, Form, Input, InputNumber, Select, Switch, Tabs } from 'antd'
import { useTranslation } from 'react-i18next'
import PageContainer from '@/components/PageContainer'
import PermissionButton from '@/components/PermissionButton'
import AttachmentImageField from '@/components/AttachmentImageField'
import { getConfigByGroup, saveConfig, testEmail } from '@/api/config'
import { entityField } from '@/utils/normalize'
import { notifyWebsiteConfigUpdated } from '@/utils/publicImage'
import { useUnhandledError } from '@/hooks/useUnhandledError'

function configsToForm(
  configs: Array<Record<string, unknown>> | undefined,
  keys: string[],
  parsers?: Record<string, (value: string) => unknown>,
) {
  const result: Record<string, unknown> = {}
  keys.forEach((key) => {
    result[key] = parsers?.[key]?.('') ?? ''
  })

  configs?.forEach((item) => {
    const key = String(entityField(item, 'key', '') ?? entityField(item, 'Key', '') ?? '')
    const raw = String(entityField(item, 'value', '') ?? entityField(item, 'Value', '') ?? '')
    if (keys.includes(key)) {
      result[key] = parsers?.[key] ? parsers[key](raw) : raw
    }
  })

  return result
}

function WebsiteConfigPanel() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getConfigByGroup('website')
      const values = configsToForm(res.data?.configs, [
        'site_enabled',
        'site_name',
        'site_url',
        'site_logo',
        'site_icp',
        'site_keywords',
        'site_description',
        'site_copyright',
      ])
      if (values.site_enabled === '') values.site_enabled = '1'
      form.setFieldsValue(values)
    } catch (error) {
      showError(error, t('common.query_failed'))
    } finally {
      setLoading(false)
    }
  }, [form, showError, t])

  useEffect(() => {
    void loadData()
  }, [loadData])

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      await saveConfig('website', values as Record<string, unknown>)
      notifyWebsiteConfigUpdated()
      message.success(t('config.update_success'))
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Form form={form} layout="vertical" disabled={loading}>
      <Form.Item
        name="site_enabled"
        label={t('config.site_enabled')}
        valuePropName="checked"
        getValueProps={(value) => ({ checked: value === '1' || value === true })}
        getValueFromEvent={(checked: boolean) => (checked ? '1' : '0')}
      >
        <Switch checkedChildren={t('common.enabled')} unCheckedChildren={t('common.disabled')} />
      </Form.Item>
      <Form.Item name="site_name" label={t('config.site_name')}>
        <Input placeholder={t('config.site_name_placeholder')} />
      </Form.Item>
      <Form.Item name="site_url" label={t('config.site_url')}>
        <Input placeholder={t('config.site_url_placeholder')} />
      </Form.Item>
      <Form.Item name="site_logo" label={t('config.site_logo')}>
        <AttachmentImageField placeholder={t('config.site_logo_placeholder')} />
      </Form.Item>
      <Form.Item name="site_icp" label={t('config.site_icp')}>
        <Input placeholder={t('config.site_icp_placeholder')} />
      </Form.Item>
      <Form.Item name="site_keywords" label={t('config.site_keywords')}>
        <Input placeholder={t('config.site_keywords_placeholder')} />
      </Form.Item>
      <Form.Item name="site_description" label={t('config.site_description')}>
        <Input placeholder={t('config.site_description_placeholder')} />
      </Form.Item>
      <Form.Item name="site_copyright" label={t('config.site_copyright')}>
        <Input.TextArea rows={3} placeholder={t('config.site_copyright_placeholder')} />
      </Form.Item>
      <PermissionButton permission="config.save" type="primary" loading={submitting} onClick={() => void handleSubmit()}>
        {t('common.save')}
      </PermissionButton>
    </Form>
  )
}

function EmailConfigPanel() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [testing, setTesting] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getConfigByGroup('email')
      const values = configsToForm(
        res.data?.configs,
        [
          'email_host',
          'email_port',
          'email_username',
          'email_password',
          'email_from',
          'email_from_name',
          'email_encryption',
          'email_timeout',
        ],
        {
          email_port: (v) => (v ? Number(v) : 587),
          email_timeout: (v) => (v ? Number(v) : 30),
        },
      )
      if (!values.email_encryption) values.email_encryption = 'tls'
      form.setFieldsValue(values)
    } catch (error) {
      showError(error, t('common.query_failed'))
    } finally {
      setLoading(false)
    }
  }, [form, showError, t])

  useEffect(() => {
    void loadData()
  }, [loadData])

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      await saveConfig('email', values as Record<string, unknown>)
      message.success(t('config.update_success'))
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    } finally {
      setSubmitting(false)
    }
  }

  const handleTest = async () => {
    try {
      const values = await form.validateFields()
      setTesting(true)
      await testEmail(values as Record<string, unknown>)
      message.success(t('config.test_email_success'))
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('config.test_email_failed'))
    } finally {
      setTesting(false)
    }
  }

  return (
    <Form form={form} layout="vertical" disabled={loading}>
      <Form.Item name="email_host" label={t('config.email_host')}>
        <Input placeholder={t('config.email_host_placeholder')} />
      </Form.Item>
      <Form.Item name="email_port" label={t('config.email_port')}>
        <InputNumber min={1} max={65535} style={{ width: '100%' }} />
      </Form.Item>
      <Form.Item name="email_username" label={t('config.email_username')}>
        <Input autoComplete="off" placeholder={t('config.email_username_placeholder')} />
      </Form.Item>
      <Form.Item name="email_password" label={t('config.email_password')}>
        <Input.Password autoComplete="new-password" placeholder={t('config.email_password_placeholder')} />
      </Form.Item>
      <Form.Item name="email_from" label={t('config.email_from')}>
        <Input placeholder={t('config.email_from_placeholder')} />
      </Form.Item>
      <Form.Item name="email_from_name" label={t('config.email_from_name')}>
        <Input placeholder={t('config.email_from_name_placeholder')} />
      </Form.Item>
      <Form.Item name="email_encryption" label={t('config.email_encryption')}>
        <Select
          options={[
            { label: 'TLS', value: 'tls' },
            { label: 'SSL', value: 'ssl' },
            { label: 'None', value: '' },
          ]}
        />
      </Form.Item>
      <Form.Item name="email_timeout" label={t('config.email_timeout')}>
        <InputNumber min={1} max={300} style={{ width: '100%' }} addonAfter={t('config.email_timeout_unit')} />
      </Form.Item>
      <Form.Item>
        <PermissionButton permission="config.save" type="primary" loading={submitting} onClick={() => void handleSubmit()}>
          {t('common.save')}
        </PermissionButton>
        <PermissionButton
          permission="config.test_email"
          style={{ marginLeft: 8 }}
          loading={testing}
          onClick={() => void handleTest()}
        >
          {t('config.test_email')}
        </PermissionButton>
      </Form.Item>
    </Form>
  )
}

function CaptchaConfigPanel() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getConfigByGroup('captcha')
      const values = configsToForm(res.data?.configs, ['captcha_enabled', 'captcha_expire'], {
        captcha_enabled: (v) => v === '1' || v === 'true',
        captcha_expire: (v) => (v ? Number(v) : 120),
      })
      form.setFieldsValue(values)
    } catch (error) {
      showError(error, t('common.query_failed'))
    } finally {
      setLoading(false)
    }
  }, [form, showError, t])

  useEffect(() => {
    void loadData()
  }, [loadData])

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      await saveConfig('captcha', {
        captcha_enabled: values.captcha_enabled ? '1' : '0',
        captcha_expire: String(values.captcha_expire ?? 120),
      })
      message.success(t('config.update_success'))
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Form form={form} layout="vertical" disabled={loading}>
      <Form.Item name="captcha_enabled" label={t('config.captcha_enabled')} valuePropName="checked">
        <Switch checkedChildren={t('common.enabled')} unCheckedChildren={t('common.disabled')} />
      </Form.Item>
      <Form.Item name="captcha_expire" label={t('config.captcha_expire')}>
        <InputNumber min={60} max={600} style={{ width: '100%' }} addonAfter={t('config.captcha_expire_unit')} />
      </Form.Item>
      <PermissionButton permission="config.save" type="primary" loading={submitting} onClick={() => void handleSubmit()}>
        {t('common.save')}
      </PermissionButton>
    </Form>
  )
}

function StorageConfigPanel() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getConfigByGroup('storage')
      const values = configsToForm(res.data?.configs, ['file_disk', 'export_format'])
      if (!values.file_disk) values.file_disk = 'local'
      if (!values.export_format) values.export_format = 'csv'
      form.setFieldsValue(values)
    } catch (error) {
      showError(error, t('common.query_failed'))
    } finally {
      setLoading(false)
    }
  }, [form, showError, t])

  useEffect(() => {
    void loadData()
  }, [loadData])

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      await saveConfig('storage', values as Record<string, unknown>)
      message.success(t('config.update_success'))
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Form form={form} layout="vertical" disabled={loading}>
      <Form.Item name="file_disk" label={t('config.file_disk')}>
        <Select
          options={[
            { label: 'local', value: 'local' },
            { label: 's3', value: 's3' },
            { label: 'oss', value: 'oss' },
            { label: 'cos', value: 'cos' },
            { label: 'qiniu', value: 'qiniu' },
            { label: 'minio', value: 'minio' },
          ]}
        />
      </Form.Item>
      <Form.Item name="export_format" label={t('config.export_format')}>
        <Select
          options={[
            { label: 'csv', value: 'csv' },
            { label: 'xlsx', value: 'xlsx' },
          ]}
        />
      </Form.Item>
      <PermissionButton permission="config.save" type="primary" loading={submitting} onClick={() => void handleSubmit()}>
        {t('common.save')}
      </PermissionButton>
    </Form>
  )
}

export default function ConfigList() {
  const { t } = useTranslation()

  return (
    <PageContainer title={t('menu.config')}>
      <Card>
        <Tabs
          items={[
            { key: 'website', label: t('config.website_config'), children: <WebsiteConfigPanel /> },
            { key: 'email', label: t('config.email_config'), children: <EmailConfigPanel /> },
            { key: 'captcha', label: t('config.captcha_config'), children: <CaptchaConfigPanel /> },
            { key: 'storage', label: t('config.storage_config'), children: <StorageConfigPanel /> },
          ]}
        />
      </Card>
    </PageContainer>
  )
}
