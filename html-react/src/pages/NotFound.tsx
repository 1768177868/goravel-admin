import { Button, Result } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function NotFoundPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  return (
    <Result
      status="404"
      title="404"
      subTitle={t('notFound.title')}
      extra={
        <Button type="primary" onClick={() => navigate('/dashboard')}>
          {t('notFound.back_home')}
        </Button>
      }
    />
  )
}
