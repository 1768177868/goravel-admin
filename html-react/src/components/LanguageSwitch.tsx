import { Select } from 'antd'
import { GlobalOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { setLanguage } from '@/i18n'

export default function LanguageSwitch() {
  const { i18n, t } = useTranslation()

  return (
    <Select
      size="small"
      variant="borderless"
      value={i18n.language === 'en-US' ? 'en-US' : 'zh-CN'}
      style={{ width: 110 }}
      suffixIcon={<GlobalOutlined />}
      options={[
        { value: 'zh-CN', label: t('common.language_zh') },
        { value: 'en-US', label: t('common.language_en') },
      ]}
      onChange={(value) => setLanguage(value as 'zh-CN' | 'en-US')}
    />
  )
}
