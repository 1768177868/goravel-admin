import { ref, onUnmounted } from 'vue'
import logger from '../utils/logger'

/**
 * API 请求 composable
 * 提供请求取消机制和加载状态管理
 */
export function useApiRequest() {
  const abortController = ref(null)
  const loading = ref(false)
  const error = ref(null)

  /**
   * 执行 API 请求
   * @param {Function} apiCall - API 调用函数（应该返回 Promise，可以是 axios 请求）
   * @param {Object} options - 配置选项
   * @param {boolean} options.cancelPrevious - 是否取消之前的请求
   * @returns {Promise} 请求结果
   */
  const request = async (apiCall, options = {}) => {
    const { cancelPrevious = true } = options

    // 取消之前的请求
    if (cancelPrevious && abortController.value) {
      abortController.value.abort()
      logger.debug('Previous request cancelled')
    }

    // 创建新的 AbortController
    abortController.value = new AbortController()
    loading.value = true
    error.value = null

    try {
      // 调用 API 函数
      // 注意：如果 apiCall 返回的是 axios 请求对象，需要手动添加 signal
      const promise = apiCall()
      
      // 如果返回的是 axios 请求对象（有 cancel 方法），添加取消支持
      if (promise && typeof promise.cancel === 'function') {
        // 存储取消函数
        const originalCancel = promise.cancel
        promise.cancel = () => {
          abortController.value?.abort()
          originalCancel()
        }
      }
      
      const result = await promise
      
      // 检查请求是否被取消
      if (abortController.value?.signal?.aborted) {
        logger.debug('Request was cancelled during execution')
        return null
      }
      
      return result
    } catch (err) {
      // 如果是取消错误，不处理
      if (err.name === 'AbortError' || 
          err.message === 'canceled' || 
          err.code === 'ERR_CANCELED' ||
          err.message?.includes('canceled')) {
        logger.debug('Request was cancelled')
        return null
      }

      // 其他错误
      error.value = err
      throw err
    } finally {
      // 只有在请求未被取消时才重置 loading
      if (!abortController.value?.signal?.aborted) {
        loading.value = false
      }
    }
  }

  /**
   * 取消当前请求
   */
  const cancel = () => {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
      loading.value = false
      logger.debug('Request cancelled manually')
    }
  }

  /**
   * 重置状态
   */
  const reset = () => {
    cancel()
    error.value = null
  }

  // 组件卸载时取消请求
  onUnmounted(() => {
    cancel()
  })

  return {
    request,
    cancel,
    reset,
    loading,
    error
  }
}

export default useApiRequest

