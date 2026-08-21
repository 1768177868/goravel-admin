import { useEffect } from 'react'
import { useRouteError } from 'react-router-dom'
import { Spin, message } from 'antd'
import i18n from '@/i18n'
import { handleChunkLoadFailure } from '@/utils/chunkLoadRecovery'
import logger from '@/utils/logger'

/**
 * React Router errorElement: chunk-load → toast + auto-reload (Vue router.onError);
 * other route errors → toast only.
 */
export function RouteErrorFallback() {
  const error = useRouteError()

  useEffect(() => {
    logger.error('Router error:', error)
    if (handleChunkLoadFailure(error)) return
    message.error({
      content: i18n.t('error.route_nav_failed'),
      duration: 3,
    })
  }, [error])

  return (
    <div className="page-fallback page-fallback--fullscreen">
      <Spin size="large" />
      <div style={{ marginTop: 16 }}>{i18n.t('error.page_load_failed')}</div>
    </div>
  )
}
