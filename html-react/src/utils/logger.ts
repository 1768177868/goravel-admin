const isDev = import.meta.env.DEV

export const logger = {
  log: (...args: unknown[]) => {
    if (isDev) console.log('[LOG]', ...args)
  },
  error: (...args: unknown[]) => {
    if (isDev) console.error('[ERROR]', ...args)
  },
  warn: (...args: unknown[]) => {
    if (isDev) console.warn('[WARN]', ...args)
  },
  debug: (...args: unknown[]) => {
    if (isDev) console.debug('[DEBUG]', ...args)
  },
  info: (...args: unknown[]) => {
    if (isDev) console.info('[INFO]', ...args)
  },
}

export default logger
