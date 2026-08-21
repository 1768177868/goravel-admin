import { lazy, type ComponentType, type LazyExoticComponent } from 'react'
import { message } from 'antd'
import i18n from '@/i18n'
import { handleChunkLoadFailure } from '@/utils/chunkLoadRecovery'
import logger from '@/utils/logger'

type ModuleDefault = { default: ComponentType<object> }

/**
 * Lazy import with retry + timeout (mirrors Vue router lazyLoad).
 * On chunk-load failure after retries: toast + auto-reload (mirrors Vue router.onError).
 */
export function lazyLoad(
  importFn: () => Promise<ModuleDefault>,
  maxRetries = 3,
  timeout = 10000,
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
            if (retryCount < maxRetries) {
              logger.warn(`模块加载失败，正在重试 (${retryCount}/${maxRetries}):`, error.message)
              const delay = Math.min(1000 * Math.pow(2, retryCount - 1), 5000)
              setTimeout(attemptLoad, delay)
            } else {
              logger.error('模块加载失败，已达到最大重试次数:', error)
              if (!handleChunkLoadFailure(error)) {
                message.error({
                  content: i18n.t('error.page_load_failed'),
                  duration: 5,
                })
              }
              reject(error)
            }
          })
      }

      attemptLoad()
    })

  return lazy(loadWithRetry)
}
