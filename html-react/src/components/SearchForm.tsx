import { Button, Col, DatePicker, Form, Input, Row, Select, Space } from 'antd'
import dayjs, { type Dayjs } from 'dayjs'
import { useTranslation } from 'react-i18next'

export interface SearchField {
  name: string
  label: string
  placeholder?: string
  span?: number
  type?: 'input' | 'select' | 'datetime'
  options?: Array<{ label: string; value: string | number }>
  allowClear?: boolean
}

interface SearchFormProps {
  fields: SearchField[]
  values: Record<string, unknown>
  onChange: (values: Record<string, unknown>) => void
  onSearch: () => void
  onReset: () => void
}

export default function SearchForm({ fields, values, onChange, onSearch, onReset }: SearchFormProps) {
  const { t } = useTranslation()

  return (
    <Form layout="vertical" style={{ marginBottom: 16 }} onFinish={onSearch}>
      <Row gutter={16}>
        {fields.map((field) => (
          <Col key={field.name} xs={24} sm={12} md={field.span || 6}>
            <Form.Item label={field.label} style={{ marginBottom: 12 }}>
              {field.type === 'select' ? (
                <Select
                  allowClear={field.allowClear !== false}
                  placeholder={field.placeholder || field.label}
                  value={(values[field.name] as string | number | undefined) ?? undefined}
                  options={field.options || []}
                  onChange={(value) => onChange({ ...values, [field.name]: value ?? '' })}
                />
              ) : field.type === 'datetime' ? (
                <DatePicker
                  showTime
                  style={{ width: '100%' }}
                  allowClear={field.allowClear !== false}
                  placeholder={field.placeholder || field.label}
                  value={
                    values[field.name]
                      ? dayjs(String(values[field.name]))
                      : null
                  }
                  onChange={(value: Dayjs | null) =>
                    onChange({
                      ...values,
                      [field.name]: value ? value.format('YYYY-MM-DD HH:mm:ss') : '',
                    })
                  }
                />
              ) : (
                <Input
                  allowClear
                  placeholder={field.placeholder || field.label}
                  value={(values[field.name] as string) ?? ''}
                  onChange={(e) => onChange({ ...values, [field.name]: e.target.value })}
                  onPressEnter={onSearch}
                />
              )}
            </Form.Item>
          </Col>
        ))}
        <Col xs={24} sm={12} md={6} style={{ display: 'flex', alignItems: 'end' }}>
          <Form.Item style={{ marginBottom: 12 }}>
            <Space>
              <Button type="primary" htmlType="submit">
                {t('common.search')}
              </Button>
              <Button onClick={onReset}>{t('common.reset')}</Button>
            </Space>
          </Form.Item>
        </Col>
      </Row>
    </Form>
  )
}
