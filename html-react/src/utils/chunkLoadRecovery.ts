import { message } from 'antd'
import i18n from '@/i18n'
import logger from '@/utils/logger'

let reloadScheduled = false

/** Matches Vue router.onError chunk-load detection after deploy. */
export function isChunkLoadError(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false
  const err = error as { message?: string; name?: string }
  const msg = err.message || ''
  return (
    msg.includes('Failed to fetch dynamically imported module') ||
    msg.includes('Loading chunk') ||
    msg.includes('Loading CSS chunk') ||
    err.name === 'ChunkLoadError'
  )
}

/**
 * Toast + auto-reload (deduped). Returns true when handled as a chunk-load failure.
 */
export function handleChunkLoadFailure(error: unknown): boolean {
  if (!isChunkLoadError(error)) return false
  if (reloadScheduled) return true

  reloadScheduled = true
  logger.error('Chunk load failed, scheduling page reload:', error)
  message.error({
    content: i18n.t('error.page_load_failed'),
    duration: 5,
  })
  setTimeout(() => {
    window.location.reload()
  }, 2000)
  return true
}
