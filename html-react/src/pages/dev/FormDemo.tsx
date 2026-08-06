import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Cascader,
  Checkbox,
  ColorPicker,
  Col,
  DatePicker,
  Form,
  Input,
  InputNumber,
  Radio,
  Rate,
  Row,
  Select,
  Slider,
  Space,
  Switch,
} from 'antd'
import dayjs, { type Dayjs } from 'dayjs'
import { useTranslation } from 'react-i18next'
import PageContainer from '@/components/PageContainer'
import WangEditor from '@/components/WangEditor'
import MarkdownEditor from '@/components/MarkdownEditor'
import request from '@/utils/request'

function toDayjs(value: unknown): Dayjs | undefined {
  if (value == null || value === '') return undefined
  if (dayjs.isDayjs(value)) return value.isValid() ? value : undefined
  const parsed = dayjs(value as string | number | Date)
  return parsed.isValid() ? parsed : undefined
}

function toDayjsRange(value: unknown): [Dayjs, Dayjs] | undefined {
  if (!Array.isArray(value) || value.length < 2) return undefined
  const start = toDayjs(value[0])
  const end = toDayjs(value[1])
  if (!start || !end) return undefined
  return [start, end]
}

/** Convert API/sample date strings into dayjs for Ant Design DatePicker. */
function normalizeFormDemoValues(raw: Record<string, unknown>) {
  return {
    ...raw,
    birthday: toDayjs(raw.birthday),
    meeting_at: toDayjs(raw.meeting_at),
    active_days: toDayjsRange(raw.active_days),
    active_period: toDayjsRange(raw.active_period),
  }
}

function serializePreview(values: Record<string, unknown>) {
  const out: Record<string, unknown> = {}
  Object.entries(values).forEach(([key, value]) => {
    if (dayjs.isDayjs(value)) {
      out[key] = value.format(key === 'birthday' ? 'YYYY-MM-DD' : 'YYYY-MM-DD HH:mm:ss')
      return
    }
    if (Array.isArray(value) && value.every((item) => dayjs.isDayjs(item))) {
      out[key] = value.map((item) =>
        (item as Dayjs).format(key === 'active_days' ? 'YYYY-MM-DD' : 'YYYY-MM-DD HH:mm:ss'),
      )
      return
    }
    out[key] = value
  })
  return out
}

export default function FormDemo() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const [form] = Form.useForm()
  const [preview, setPreview] = useState<Record<string, unknown>>({})
  const [loadingApi, setLoadingApi] = useState(false)

  const regionOptions = [
    {
      value: 'china',
      label: t('form_demo.country_china'),
      children: [
        {
          value: 'guangdong',
          label: t('form_demo.province_guangdong'),
          children: [{ value: 'guangzhou', label: t('form_demo.city_guangzhou') }],
        },
      ],
    },
  ]

  const applyValues = (raw: Record<string, unknown>) => {
    const values = normalizeFormDemoValues(raw)
    form.setFieldsValue(values)
    setPreview(serializePreview(form.getFieldsValue(true)))
  }

  const fillSample = () => {
    applyValues({
      username: 'demo_user',
      password: '123456',
      intro: t('form_demo.intro'),
      gender: 'male',
      hobbies: ['reading', 'music'],
      role: 'admin',
      birthday: '2026-03-31',
      meeting_at: '2026-03-31 09:30:00',
      active_days: ['2026-03-01', '2026-03-31'],
      score: 88,
      satisfaction: 4,
      volume: 60,
      enabled: true,
      color: '#1677ff',
      region: ['china', 'guangdong', 'guangzhou'],
      rich_content_wang: '<p>WangEditor demo content</p>',
      rich_content_markdown: '**Markdown** demo content',
    })
  }

  const loadApiData = async () => {
    setLoadingApi(true)
    try {
      const res = await request({ url: '/form-demo/data', method: 'get' })
      const data = (res.data || {}) as Record<string, unknown>
      applyValues(data)
      message.success(t('form_demo.loaded_from_api'))
    } catch {
      message.warning(t('common.query_failed'))
    } finally {
      setLoadingApi(false)
    }
  }

  return (
    <PageContainer title={t('form_demo.title', { defaultValue: t('menu.form_demo') })}>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={14}>
          <Card>
            <Form
              form={form}
              layout="vertical"
              initialValues={{
                gender: 'male',
                hobbies: ['reading'],
                role: 'viewer',
                score: 60,
                satisfaction: 3,
                volume: 40,
                enabled: true,
                color: '#1677ff',
                rich_content_wang: '',
                rich_content_markdown: '',
              }}
              onValuesChange={(_, all) => setPreview(serializePreview(all))}
              onFinish={(values) => {
                setPreview(serializePreview(values))
                message.success(t('common.success'))
              }}
            >
              <Form.Item name="username" label={t('form_demo.username')} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="password" label={t('common.password')}>
                <Input.Password />
              </Form.Item>
              <Form.Item name="intro" label={t('form_demo.intro')}>
                <Input.TextArea rows={3} />
              </Form.Item>
              <Form.Item
                name="gender"
                label={t('form_demo.gender')}
                rules={[{ required: true, message: t('form_demo.gender_required') }]}
              >
                <Radio.Group
                  options={[
                    { label: t('form_demo.gender_male'), value: 'male' },
                    { label: t('form_demo.gender_female'), value: 'female' },
                  ]}
                />
              </Form.Item>
              <Form.Item name="hobbies" label={t('form_demo.hobbies')}>
                <Checkbox.Group
                  options={[
                    { label: t('form_demo.hobby_reading'), value: 'reading' },
                    { label: t('form_demo.hobby_sports'), value: 'sports' },
                    { label: t('form_demo.hobby_music'), value: 'music' },
                  ]}
                />
              </Form.Item>
              <Form.Item name="role" label={t('form_demo.role')}>
                <Select
                  options={[
                    { label: t('form_demo.role_admin'), value: 'admin' },
                    { label: t('form_demo.role_editor'), value: 'editor' },
                    { label: t('form_demo.role_viewer'), value: 'viewer' },
                  ]}
                />
              </Form.Item>
              <Form.Item name="birthday" label={t('form_demo.birthday')}>
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name="meeting_at" label={t('form_demo.meeting_at')}>
                <DatePicker showTime style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name="active_days" label={t('form_demo.active_days')}>
                <DatePicker.RangePicker style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name="score" label={t('form_demo.score_input_number')}>
                <InputNumber min={0} max={100} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name="satisfaction" label={t('form_demo.satisfaction')}>
                <Rate />
              </Form.Item>
              <Form.Item name="volume" label={t('form_demo.volume')}>
                <Slider />
              </Form.Item>
              <Form.Item name="enabled" label={t('common.status')} valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item
                name="color"
                label={t('form_demo.color')}
                getValueFromEvent={(color) => (typeof color === 'string' ? color : color?.toHexString?.())}
              >
                <ColorPicker showText />
              </Form.Item>
              <Form.Item name="region" label={t('form_demo.region')}>
                <Cascader options={regionOptions} />
              </Form.Item>
              <Form.Item name="rich_content_wang" label={t('form_demo.rich_content_wang')}>
                <WangEditor height={280} placeholder={t('form_demo.rich_content_wang_placeholder')} />
              </Form.Item>
              <Form.Item name="rich_content_markdown" label={t('form_demo.rich_content_markdown')}>
                <MarkdownEditor height={280} placeholder={t('form_demo.rich_content_markdown_placeholder')} />
              </Form.Item>
              <Form.Item>
                <Space wrap>
                  <Button type="primary" htmlType="submit">
                    {t('common.submit')}
                  </Button>
                  <Button onClick={() => form.resetFields()}>{t('common.reset')}</Button>
                  <Button onClick={fillSample}>{t('form_demo.fill_sample')}</Button>
                  <Button loading={loadingApi} onClick={() => void loadApiData()}>
                    {t('form_demo.load_api_data')}
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card title={t('form_demo.preview')}>
            <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
              {JSON.stringify(preview, null, 2)}
            </pre>
          </Card>
        </Col>
      </Row>
    </PageContainer>
  )
}
