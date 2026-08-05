import { App } from 'antd'
import type { ApiError } from '@/types'

/** Show error only when interceptor did not already toast. */
export function useUnhandledError() {
  const { message } = App.useApp()

  return (error: unknown, fallback: string) => {
    const err = error as ApiError
    if (!err?.__handled) {
      message.error(err?.translatedMessage || err?.message || fallback)
    }
  }
}
