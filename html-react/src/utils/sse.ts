import Storage from '@/utils/storage'

export interface SSEOptions {
  onMessage?: (data: unknown, event: MessageEvent) => void
  onError?: (error: Event, eventSource: EventSource) => void
  onOpen?: (event: Event) => void
  onClose?: () => void
}

function getBaseURL() {
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL as string | undefined
  const apiPrefix = (import.meta.env.VITE_API_PREFIX as string | undefined) || '/api/admin'

  if (apiBaseURL) {
    const base = apiBaseURL.replace(/\/+$/, '')
    const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
    return `${base}${prefix}`
  }
  return apiPrefix
}

export function createSSEConnection(url: string, options: SSEOptions = {}) {
  const { onMessage, onError, onOpen, onClose } = options
  const baseURL = getBaseURL()
  const fullURL = url.startsWith('http') ? url : `${baseURL}${url.startsWith('/') ? url : `/${url}`}`

  const token = Storage.getItem<string>('token', '')
  if (!token || typeof token !== 'string') {
    throw new Error('Token is required for SSE connection')
  }

  const separator = fullURL.includes('?') ? '&' : '?'
  const urlWithToken = `${fullURL}${separator}_token=${encodeURIComponent(token.trim())}`
  const eventSource = new EventSource(urlWithToken)

  if (onOpen) eventSource.onopen = onOpen

  if (onMessage) {
    eventSource.onmessage = (event) => {
      try {
        onMessage(JSON.parse(event.data), event)
      } catch {
        onMessage(event.data, event)
      }
    }
  }

  eventSource.onerror = (error) => {
    onError?.(error, eventSource)
    if (eventSource.readyState === EventSource.CLOSED) {
      onClose?.()
    }
  }

  return eventSource
}

export function closeSSEConnection(eventSource: EventSource | null | undefined) {
  if (eventSource && eventSource.readyState !== EventSource.CLOSED) {
    eventSource.close()
  }
}

export const SSE_STATE = {
  CONNECTING: 0,
  OPEN: 1,
  CLOSED: 2,
} as const
