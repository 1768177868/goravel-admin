import { Tag } from 'antd'
import { useTranslation } from 'react-i18next'

interface StatusTagProps {
  status: number | string | boolean | undefined | null
  enabledText?: string
  disabledText?: string
}

export default function StatusTag({ status, enabledText, disabledText }: StatusTagProps) {
  const { t } = useTranslation()
  const enabled = status === 1 || status === '1' || status === true

  return (
    <Tag color={enabled ? 'success' : 'default'}>
      {enabled ? enabledText || t('common.enabled') : disabledText || t('common.disabled')}
    </Tag>
  )
}
