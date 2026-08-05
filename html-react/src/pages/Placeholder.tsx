import { Result } from 'antd'
import { useTranslation } from 'react-i18next'
import PageContainer from '@/components/PageContainer'

/** Fallback for menu routes whose React page is not migrated yet. */
export default function PlaceholderPage() {
  const { t } = useTranslation()

  return (
    <PageContainer title={t('common.coming_soon')}>
      <Result
        status="info"
        title={t('common.coming_soon')}
        subTitle={t('common.coming_soon_desc')}
      />
    </PageContainer>
  )
}
