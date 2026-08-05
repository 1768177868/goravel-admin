import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import PageContainer from '@/components/PageContainer'

export default function IframeView() {
  const [params] = useSearchParams()
  const url = useMemo(() => params.get('url') || '', [params])

  return (
    <PageContainer title="IFrame">
      {url ? (
        <iframe
          title="external"
          src={url}
          style={{ width: '100%', height: 'calc(100vh - 220px)', border: 0, borderRadius: 8 }}
        />
      ) : (
        <div>Missing url</div>
      )}
    </PageContainer>
  )
}
