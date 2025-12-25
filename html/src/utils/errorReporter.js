/**
 * 错误上报工具
 * 
 * 用于收集和上报前端错误，便于监控和排查问题
 */

import logger from './logger'

/**
 * 错误级别
 */
export const ErrorLevel = {
  INFO: 'info',
  WARNING: 'warning',
  ERROR: 'error',
  FATAL: 'fatal'
}

/**
 * 错误上报配置
 */
const config = {
  // 是否启用错误上报
  enabled: import.meta.env.PROD,
  // 上报地址 (如 Sentry DSN)
  reportUrl: import.meta.env.VITE_ERROR_REPORT_URL || '',
  // 采样率 (0-1)
  sampleRate: 1.0,
  // 忽略的错误类型
  ignoreErrors: [
    'ResizeObserver loop limit exceeded',
    'ResizeObserver loop completed with undelivered notifications',
    'Cannot read properties of undefined (reading \'indexOf\')', // Element Plus TabPane 已知问题
  ],
  // 最大错误堆栈长度
  maxStackLength: 50
}

/**
 * 错误队列 (用于批量上报)
 */
const errorQueue = []
const MAX_QUEUE_SIZE = 10
let flushTimer = null

/**
 * 判断错误是否应该被忽略
 * @param {Error|string} error 
 * @returns {boolean}
 */
function shouldIgnore(error) {
  const message = error?.message || String(error)
  return config.ignoreErrors.some(pattern => message.includes(pattern))
}

/**
 * 格式化错误信息
 * @param {Error|string} error 
 * @param {Object} context 
 * @returns {Object}
 */
function formatError(error, context = {}) {
  const errorInfo = {
    timestamp: new Date().toISOString(),
    level: context.level || ErrorLevel.ERROR,
    message: error?.message || String(error),
    stack: error?.stack?.split('\n').slice(0, config.maxStackLength).join('\n') || '',
    url: window.location.href,
    userAgent: navigator.userAgent,
    ...context
  }

  // 添加用户信息 (如果有)
  try {
    const userStore = JSON.parse(localStorage.getItem('user') || '{}')
    if (userStore.userInfo) {
      errorInfo.userId = userStore.userInfo.id
      errorInfo.username = userStore.userInfo.username
    }
  } catch {
    // 忽略
  }

  return errorInfo
}

/**
 * 发送错误到服务器
 * @param {Object[]} errors 
 */
async function sendErrors(errors) {
  if (!config.reportUrl || errors.length === 0) {
    return
  }

  try {
    await fetch(config.reportUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ errors }),
      keepalive: true // 页面关闭时也能发送
    })
  } catch (e) {
    // 上报失败不影响用户体验
    console.warn('[ErrorReporter] Failed to send errors:', e)
  }
}

/**
 * 刷新错误队列
 */
function flushQueue() {
  if (errorQueue.length === 0) return
  
  const errors = errorQueue.splice(0)
  sendErrors(errors)
}

/**
 * 添加错误到队列
 * @param {Object} errorInfo 
 */
function addToQueue(errorInfo) {
  errorQueue.push(errorInfo)

  if (errorQueue.length >= MAX_QUEUE_SIZE) {
    flushQueue()
  } else if (!flushTimer) {
    // 5秒后自动刷新
    flushTimer = setTimeout(() => {
      flushQueue()
      flushTimer = null
    }, 5000)
  }
}

/**
 * 上报错误
 * @param {Error|string} error 错误对象或错误消息
 * @param {Object} context 上下文信息
 * @param {string} context.level 错误级别
 * @param {string} context.component 组件名
 * @param {string} context.action 操作名
 * @param {Object} context.extra 额外信息
 * 
 * @example
 * // 上报普通错误
 * reportError(new Error('Something went wrong'))
 * 
 * // 上报带上下文的错误
 * reportError(error, {
 *   level: ErrorLevel.WARNING,
 *   component: 'UserList',
 *   action: 'fetchUsers',
 *   extra: { userId: 123 }
 * })
 */
export function reportError(error, context = {}) {
  // 检查是否应该忽略
  if (shouldIgnore(error)) {
    return
  }

  // 采样
  if (Math.random() > config.sampleRate) {
    return
  }

  // 格式化错误
  const errorInfo = formatError(error, context)

  // 本地日志
  if (context.level === ErrorLevel.FATAL) {
    logger.error('[FATAL]', errorInfo.message, errorInfo)
  } else if (context.level === ErrorLevel.WARNING) {
    logger.warn('[WARN]', errorInfo.message, errorInfo)
  } else {
    logger.error('[ERROR]', errorInfo.message, errorInfo)
  }

  // 生产环境上报
  if (config.enabled) {
    addToQueue(errorInfo)
  }
}

/**
 * 上报 API 错误
 * @param {Object} error Axios 错误对象
 * @param {Object} requestConfig 请求配置
 */
export function reportApiError(error, requestConfig = {}) {
  const context = {
    level: ErrorLevel.ERROR,
    type: 'api',
    url: requestConfig.url,
    method: requestConfig.method,
    status: error.response?.status,
    responseData: error.response?.data
  }

  reportError(error, context)
}

/**
 * 上报组件错误
 * @param {Error} error 
 * @param {string} componentName 
 * @param {string} hook 生命周期钩子名
 */
export function reportComponentError(error, componentName, hook = '') {
  reportError(error, {
    level: ErrorLevel.ERROR,
    type: 'component',
    component: componentName,
    hook
  })
}

/**
 * 上报 Promise 未处理的 rejection
 * @param {PromiseRejectionEvent} event 
 */
export function reportUnhandledRejection(event) {
  reportError(event.reason || 'Unhandled Promise Rejection', {
    level: ErrorLevel.ERROR,
    type: 'unhandledRejection'
  })
}

/**
 * 配置错误上报
 * @param {Object} options 
 */
export function configureErrorReporter(options = {}) {
  Object.assign(config, options)
}

/**
 * 手动刷新队列 (页面关闭前调用)
 */
export function flush() {
  flushQueue()
}

// 页面关闭前刷新队列
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', flush)
  window.addEventListener('pagehide', flush)
}

export default {
  reportError,
  reportApiError,
  reportComponentError,
  reportUnhandledRejection,
  configureErrorReporter,
  flush,
  ErrorLevel
}

