import { useEffect, useState } from 'react'
import { App, Alert, Button, Card, Form, Input, Space, Tabs, Typography } from 'antd'
import { useTranslation } from 'react-i18next'
import { updatePassword, updateProfile } from '@/api/profile'
import {
  bindGoogleAuthenticator,
  getGoogleAuthenticatorQRCode,
  getGoogleAuthenticatorStatus,
  unbindGoogleAuthenticator,
} from '@/api/auth'
import { useUserStore } from '@/stores/user'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import PageContainer from '@/components/PageContainer'

export default function ProfilePage() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const adminInfo = useUserStore((s) => s.adminInfo)
  const fetchUserInfo = useUserStore((s) => s.fetchUserInfo)
  const [profileForm] = Form.useForm()
  const [passwordForm] = Form.useForm()
  const [bindForm] = Form.useForm()
  const [unbindForm] = Form.useForm()
  const [savingProfile, setSavingProfile] = useState(false)
  const [savingPassword, setSavingPassword] = useState(false)
  const [bound, setBound] = useState(false)
  const [qr, setQr] = useState<{ qrcode?: string; secret?: string }>({})
  const [loading2fa, setLoading2fa] = useState(false)

  useEffect(() => {
    profileForm.setFieldsValue({
      nickname: adminInfo?.nickname || '',
      email: adminInfo?.email || '',
      phone: adminInfo?.phone || '',
    })
  }, [adminInfo, profileForm])

  const load2fa = async () => {
    setLoading2fa(true)
    try {
      const statusRes = await getGoogleAuthenticatorStatus()
      const isBound = !!(statusRes.data as { is_bound?: boolean } | undefined)?.is_bound
      setBound(isBound)
      if (!isBound) {
        const qrRes = await getGoogleAuthenticatorQRCode()
        setQr({
          qrcode: (qrRes.data as { qrcode?: string; qr_code?: string })?.qrcode ||
            (qrRes.data as { qr_code?: string })?.qr_code,
          secret: (qrRes.data as { secret?: string })?.secret,
        })
      } else {
        setQr({})
      }
    } catch (error) {
      showError(error, t('common.query_failed'))
    } finally {
      setLoading2fa(false)
    }
  }

  const handleProfileSave = async () => {
    try {
      const values = await profileForm.validateFields()
      setSavingProfile(true)
      await updateProfile(values)
      message.success(t('common.update_success'))
      await fetchUserInfo(true)
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    } finally {
      setSavingProfile(false)
    }
  }

  const handlePasswordSave = async () => {
    try {
      const values = await passwordForm.validateFields()
      setSavingPassword(true)
      await updatePassword({
        old_password: values.old_password,
        password: values.password,
        password_confirmation: values.password_confirmation,
      })
      message.success(t('common.update_success'))
      passwordForm.resetFields()
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    } finally {
      setSavingPassword(false)
    }
  }

  const handleBind = async () => {
    try {
      const values = await bindForm.validateFields()
      await bindGoogleAuthenticator({ code: values.code })
      message.success(t('common.update_success'))
      bindForm.resetFields()
      await load2fa()
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    }
  }

  const handleUnbind = async () => {
    try {
      const values = await unbindForm.validateFields()
      await unbindGoogleAuthenticator({ code: values.code })
      message.success(t('common.update_success'))
      unbindForm.resetFields()
      await load2fa()
    } catch (error) {
      if ((error as { errorFields?: unknown })?.errorFields) return
      showError(error, t('common.operation_failed'))
    }
  }

  return (
    <PageContainer title={t('menu.profile')}>
      <Card>
        <Tabs
          onChange={(key) => {
            if (key === '2fa') void load2fa()
          }}
          items={[
            {
              key: 'basic',
              label: t('common.basic_info'),
              children: (
                <Form
                  form={profileForm}
                  layout="vertical"
                  style={{ maxWidth: 480 }}
                  onFinish={() => void handleProfileSave()}
                >
                  <Form.Item label={t('table.username')}>
                    <Input value={adminInfo?.username || ''} disabled />
                  </Form.Item>
                  <Form.Item name="nickname" label={t('table.nickname')}>
                    <Input />
                  </Form.Item>
                  <Form.Item name="email" label={t('table.email')}>
                    <Input />
                  </Form.Item>
                  <Form.Item name="phone" label={t('table.phone')}>
                    <Input />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" htmlType="submit" loading={savingProfile}>
                      {t('common.save')}
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: 'password',
              label: t('common.change_password'),
              children: (
                <Form
                  form={passwordForm}
                  layout="vertical"
                  style={{ maxWidth: 480 }}
                  onFinish={() => void handlePasswordSave()}
                >
                  <Form.Item name="old_password" label={t('common.old_password')} rules={[{ required: true }]}>
                    <Input.Password />
                  </Form.Item>
                  <Form.Item name="password" label={t('common.new_password')} rules={[{ required: true }]}>
                    <Input.Password />
                  </Form.Item>
                  <Form.Item
                    name="password_confirmation"
                    label={t('common.confirm_password')}
                    dependencies={['password']}
                    rules={[
                      { required: true },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('password') === value) return Promise.resolve()
                          return Promise.reject(new Error(t('common.password_mismatch')))
                        },
                      }),
                    ]}
                  >
                    <Input.Password />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" htmlType="submit" loading={savingPassword}>
                      {t('common.save')}
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: '2fa',
              label: t('profile.google_authenticator', { defaultValue: '谷歌验证器' }),
              children: (
                <div style={{ maxWidth: 520 }}>
                  {bound ? (
                    <Space direction="vertical" style={{ width: '100%' }} size="middle">
                      <Alert
                        type="success"
                        showIcon
                        message={t('profile.google_auth_bound', { defaultValue: '已绑定谷歌验证器' })}
                      />
                      <Form form={unbindForm} layout="vertical" onFinish={() => void handleUnbind()}>
                        <Form.Item
                          name="code"
                          label={t('login.google_code_placeholder')}
                          rules={[
                            { required: true },
                            { pattern: /^\d{6}$/, message: t('login.google_code_format') },
                          ]}
                        >
                          <Input maxLength={6} />
                        </Form.Item>
                        <Button danger htmlType="submit" loading={loading2fa}>
                          {t('admin.unbind_google_auth', { defaultValue: '解绑谷歌验证' })}
                        </Button>
                      </Form>
                    </Space>
                  ) : (
                    <Space direction="vertical" style={{ width: '100%' }} size="middle">
                      <Alert
                        type="info"
                        showIcon
                        message={t('profile.google_auth_not_bound', { defaultValue: '尚未绑定谷歌验证器' })}
                      />
                      {qr.qrcode ? (
                        <img src={qr.qrcode} alt="qrcode" style={{ width: 180, height: 180 }} />
                      ) : null}
                      {qr.secret ? (
                        <Typography.Text type="secondary">
                          Secret: <Typography.Text code>{qr.secret}</Typography.Text>
                        </Typography.Text>
                      ) : null}
                      <Form form={bindForm} layout="vertical" onFinish={() => void handleBind()}>
                        <Form.Item
                          name="code"
                          label={t('login.google_code_placeholder')}
                          rules={[
                            { required: true },
                            { pattern: /^\d{6}$/, message: t('login.google_code_format') },
                          ]}
                        >
                          <Input maxLength={6} />
                        </Form.Item>
                        <Space>
                          <Button type="primary" htmlType="submit" loading={loading2fa}>
                            {t('common.confirm')}
                          </Button>
                          <Button onClick={() => void load2fa()}>{t('common.refresh')}</Button>
                        </Space>
                      </Form>
                    </Space>
                  )}
                </div>
              ),
            },
          ]}
        />
      </Card>
    </PageContainer>
  )
}
