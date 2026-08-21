import { useEffect } from 'react'
import { useRouteError } from 'react-router-dom'
import { Spin } from 'antd'
import i18n from '@/i18n'
import { handleChunkLoadFailure, isChunkLoadError } from '@/utils/chunkLoadRecovery'
import logger from '@/utils/logger'

/**
 * React Router errorElement: chunk-load → auto-reload;
 * other route errors → friendly fallback (never RR default "Unexpected Application Error").
 */
export function RouteErrorFallback() {
  const error = useRouteError()
  const chunkError = isChunkLoadError(error)

  useEffect(() => {
    logger.error('Router error:', error)
    if (handleChunkLoadFailure(error)) return
  }, [error])

  return (
    <div className="page-fallback page-fallback--fullscreen">
      <Spin size="large" />
      <div style={{ marginTop: 16 }}>
        {chunkError ? i18n.t('error.page_load_failed') : i18n.t('error.route_nav_failed')}
      </div>
    </div>
  )
}
