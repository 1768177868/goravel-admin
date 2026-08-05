import { Card, Form, Input, InputNumber, Radio, Select, Space, Button, App } from 'antd'
import { useTranslation } from 'react-i18next'
import PageContainer from '@/components/PageContainer'

/**
 * Lightweight form demo for React admin (parity placeholder for Vue FormDemo).
 */
export default function FormDemo() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const [form] = Form.useForm()

  return (
    <PageContainer title={t('menu.form_demo', { defaultValue: '表单演示' })}>
      <Card>
        <Form
          form={form}
          layout="vertical"
          style={{ maxWidth: 560 }}
          initialValues={{ status: 1, sort: 0 }}
          onFinish={() => {
            message.success(t('common.success'))
            form.resetFields()
          }}
        >
          <Form.Item name="name" label={t('common.name')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="slug" label={t('common.slug')}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label={t('common.type')}>
            <Select
              options={[
                { label: 'A', value: 'a' },
                { label: 'B', value: 'b' },
              ]}
            />
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
          <Form.Item name="description" label={t('common.description')}>
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {t('common.submit')}
              </Button>
              <Button onClick={() => form.resetFields()}>{t('common.reset')}</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  )
}
