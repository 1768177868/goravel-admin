import { message } from 'antd'
import i18n from '@/i18n'
import logger from '@/utils/logger'

const RELOAD_KEY = 'chunk_load_reload_at'
const RELOAD_COOLDOWN_MS = 15_000

let reloadScheduled = false

/** Matches Vite HMR / deploy chunk-load failures. */
export function isChunkLoadError(error: unknown): boolean {
  if (!error) return false

  const err = error as { message?: string; name?: string; stack?: string }
  const msg = `${err.message || ''} ${err.name || ''} ${err.stack || ''}`

  return (
    msg.includes('Failed to fetch dynamically imported module') ||
    msg.includes('error loading dynamically imported module') ||
    msg.includes('Importing a module script failed') ||
    msg.includes('Loading chunk') ||
    msg.includes('Loading CSS chunk') ||
    err.name === 'ChunkLoadError'
  )
}

function canReloadNow(): boolean {
  try {
    const last = Number(sessionStorage.getItem(RELOAD_KEY) || 0)
    return !last || Date.now() - last > RELOAD_COOLDOWN_MS
  } catch {
    return true
  }
}

function markReload() {
  try {
    sessionStorage.setItem(RELOAD_KEY, String(Date.now()))
  } catch {
    /* ignore */
  }
}

/**
 * Toast + auto-reload (deduped, with cooldown to avoid loops).
 * Returns true when handled as a chunk-load failure.
 */
export function handleChunkLoadFailure(error: unknown): boolean {
  if (!isChunkLoadError(error)) return false
  if (reloadScheduled) return true

  reloadScheduled = true
  logger.error('Chunk load failed, scheduling page reload:', error)

  if (!canReloadNow()) {
    message.error({
      content: i18n.t('error.page_load_failed'),
      duration: 8,
    })
    return true
  }

  markReload()
  message.error({
    content: i18n.t('error.page_load_failed'),
    duration: 3,
  })

  // Dev HMR stale module URLs almost never recover via retry — reload ASAP.
  const delay = import.meta.env.DEV ? 300 : 1500
  setTimeout(() => {
    const url = new URL(window.location.href)
    url.searchParams.set('_chunk_recovery', String(Date.now()))
    window.location.replace(url.toString())
  }, delay)

  return true
}

/** Global listeners for chunk failures that bypass route errorElement. */
export function installChunkLoadRecovery(): void {
  if (typeof window === 'undefined') return
  const w = window as Window & { __chunkLoadRecoveryInstalled?: boolean }
  if (w.__chunkLoadRecoveryInstalled) return
  w.__chunkLoadRecoveryInstalled = true

  try {
    const url = new URL(window.location.href)
    if (url.searchParams.has('_chunk_recovery')) {
      url.searchParams.delete('_chunk_recovery')
      window.history.replaceState(null, '', url.toString())
    }
  } catch {
    /* ignore */
  }

  window.addEventListener('unhandledrejection', (event) => {
    if (isChunkLoadError(event.reason)) {
      event.preventDefault()
      handleChunkLoadFailure(event.reason)
    }
  })

  window.addEventListener('error', (event) => {
    if (isChunkLoadError(event.error || event.message)) {
      handleChunkLoadFailure(event.error || event.message)
    }
  })
}
