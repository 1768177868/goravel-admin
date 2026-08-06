import { useMemo, useState } from 'react'
import { Button, Col, DatePicker, Form, Input, Row, Select, Space } from 'antd'
import { DownOutlined, UpOutlined } from '@ant-design/icons'
import dayjs, { type Dayjs } from 'dayjs'
import { useTranslation } from 'react-i18next'
import './SearchForm.scss'

export interface SearchField {
  name: string
  label: string
  placeholder?: string
  span?: number
  type?: 'input' | 'select' | 'datetime'
  options?: Array<{ label: string; value: string | number }>
  allowClear?: boolean
  /** Hidden until user expands advanced search */
  advanced?: boolean
}

interface SearchFormProps {
  fields: SearchField[]
  values: Record<string, unknown>
  onChange: (values: Record<string, unknown>) => void
  onSearch: () => void
  onReset: () => void
  /** Show expand when basic fields exceed this count (default 3) */
  collapseThreshold?: number
}

export default function SearchForm({
  fields,
  values,
  onChange,
  onSearch,
  onReset,
  collapseThreshold = 3,
}: SearchFormProps) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)

  const hasAdvancedFlag = fields.some((field) => field.advanced)
  const basicFields = useMemo(
    () => (hasAdvancedFlag ? fields.filter((field) => !field.advanced) : fields.slice(0, collapseThreshold)),
    [fields, hasAdvancedFlag, collapseThreshold],
  )
  const advancedFields = useMemo(
    () => (hasAdvancedFlag ? fields.filter((field) => field.advanced) : fields.slice(collapseThreshold)),
    [fields, hasAdvancedFlag, collapseThreshold],
  )
  const showExpand = advancedFields.length > 0
  const visibleFields = expanded ? fields : basicFields

  const renderField = (field: SearchField) => (
    <Col key={field.name} xs={24} sm={12} md={field.span || 6}>
      <Form.Item label={field.label} className="search-form__item">
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
            value={values[field.name] ? dayjs(String(values[field.name])) : null}
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
  )

  return (
    <Form className="search-form" layout="vertical" onFinish={onSearch}>
      <Row gutter={[16, 0]}>
        {visibleFields.map(renderField)}
        <Col xs={24} sm={12} md={6} className="search-form__actions">
          <Form.Item className="search-form__item" label=" ">
            <Space wrap>
              <Button type="primary" htmlType="submit">
                {t('common.search')}
              </Button>
              <Button onClick={onReset}>{t('common.reset')}</Button>
              {showExpand ? (
                <Button type="link" onClick={() => setExpanded((v) => !v)}>
                  {expanded ? t('common.collapse') : t('common.expand')}
                  {expanded ? <UpOutlined /> : <DownOutlined />}
                </Button>
              ) : null}
            </Space>
          </Form.Item>
        </Col>
      </Row>
    </Form>
  )
}
