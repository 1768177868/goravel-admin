import { ElMessage, ElNotification } from 'element-plus'
import { useI18n } from 'vue-i18n'
import logger from './logger'

/**
 * 统一错误处理工具
 */
export class ErrorHandler {
  /**
   * 处理错误
   * @param {Error} error - 错误对象
   * @param {Object} options - 配置选项
   * @param {boolean} options.silent - 是否静默处理（不显示消息）
   * @param {boolean} options.showNotification - 是否使用通知而不是消息
   * @param {string} options.customMessage - 自定义错误消息
   * @param {string} options.title - 通知标题（仅在使用通知时）
   */
  static handle(error, options = {}) {
    const {
      silent = false,
      showNotification = false,
      customMessage = null,
      title = null
    } = options

    if (silent) {
      logger.debug('Error handled silently:', error)
      return
    }

    // 如果错误已经被处理过，不再重复处理
    if (error?.__handled) {
      logger.debug('Error already handled:', error)
      return
    }

    // 获取错误消息
    const message = customMessage || 
                   error.translatedMessage || 
                   error.message || 
                   this.getDefaultMessage(error)

    // 记录错误
    logger.error('Error occurred:', {
      message,
      errorCode: error.errorCode,
      code: error.code,
      error
    })

    // 显示错误消息
    if (showNotification) {
      ElNotification.error({
        title: title || '错误',
        message,
        duration: 5000,
        showClose: true
      })
    } else {
      ElMessage.error(message)
    }

    // 标记为已处理
    if (error && typeof error === 'object') {
      error.__handled = true
    }
  }

  /**
   * 获取默认错误消息
   * @param {Error} error - 错误对象
   * @returns {string} 默认错误消息
   */
  static getDefaultMessage(error) {
    // 根据错误类型返回默认消息
    if (error.code === 'ERR_NETWORK' || error.message === 'Network Error') {
      return '网络连接失败，请检查网络设置'
    }
    if (error.code === 'ECONNABORTED') {
      return '请求超时，请稍后重试'
    }
    if (error.code === 401) {
      return '未授权，请重新登录'
    }
    if (error.code === 403) {
      return '没有权限执行此操作'
    }
    if (error.code === 404) {
      return '请求的资源不存在'
    }
    if (error.code === 500) {
      return '服务器内部错误，请稍后重试'
    }
    return '操作失败，请稍后重试'
  }

  /**
   * 处理 API 错误
   * @param {Error} error - 错误对象
   * @param {Object} options - 配置选项
   */
  static handleApiError(error, options = {}) {
    const { 
      silent = false,
      showNotification = false,
      customMessage = null
    } = options

    // 提取错误信息
    const message = customMessage || 
                   error.translatedMessage || 
                   error.response?.data?.message || 
                   error.message ||
                   this.getDefaultMessage(error)

    const errorCode = error.errorCode || error.response?.data?.error_code

    // 创建错误对象
    const apiError = {
      ...error,
      message,
      errorCode,
      __handled: false
    }

    this.handle(apiError, { silent, showNotification, customMessage: message })
    
    return apiError
  }
}

export default ErrorHandler

