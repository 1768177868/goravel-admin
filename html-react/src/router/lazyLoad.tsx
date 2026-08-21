import { lazy, type ComponentType, type LazyExoticComponent } from 'react'
import { message } from 'antd'
import i18n from '@/i18n'
import { handleChunkLoadFailure, isChunkLoadError } from '@/utils/chunkLoadRecovery'
import logger from '@/utils/logger'

type ModuleDefault = { default: ComponentType<object> }

/**
 * Lazy import with retry + timeout.
 * Vite HMR often invalidates module URLs (`?t=...`); retries of the same import
 * rarely help in DEV, so we fail fast and trigger a recovery reload.
 */
export function lazyLoad(
  importFn: () => Promise<ModuleDefault>,
  maxRetries = import.meta.env.DEV ? 2 : 3,
  timeout = import.meta.env.DEV ? 5000 : 10000,
): LazyExoticComponent<ComponentType<object>> {
  const loadWithRetry = () =>
    new Promise<ModuleDefault>((resolve, reject) => {
      let retryCount = 0

      const attemptLoad = () => {
        const timeoutPromise = new Promise<never>((_, timeoutReject) => {
          setTimeout(() => timeoutReject(new Error('模块加载超时')), timeout)
        })

        Promise.race([importFn(), timeoutPromise])
          .then((mod) => resolve(mod))
          .catch((error: Error) => {
            retryCount += 1
            const canRetry = retryCount < maxRetries && !isChunkLoadError(error)

            if (canRetry) {
              logger.warn(`模块加载失败，正在重试 (${retryCount}/${maxRetries}):`, error.message)
              const delay = Math.min(1000 * Math.pow(2, retryCount - 1), 5000)
              setTimeout(attemptLoad, delay)
              return
            }

            // Chunk/HMR failures: one quick retry then recover via reload.
            if (isChunkLoadError(error) && retryCount < maxRetries) {
              logger.warn(`Chunk load failed, quick retry (${retryCount}/${maxRetries}):`, error.message)
              setTimeout(attemptLoad, 200)
              return
            }

            logger.error('模块加载失败，已达到最大重试次数:', error)
            if (!handleChunkLoadFailure(error)) {
              message.error({
                content: i18n.t('error.page_load_failed'),
                duration: 5,
              })
            }
            reject(error)
          })
      }

      attemptLoad()
    })

  return lazy(loadWithRetry)
}
