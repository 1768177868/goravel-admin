/**
 * 日志工具
 * 开发环境输出到控制台，生产环境可发送到日志服务
 */
const isDev = import.meta.env.DEV

export const logger = {
  /**
   * 输出日志信息
   * @param {...any} args - 日志参数
   */
  log: (...args) => {
    if (isDev) {
      console.log('[LOG]', ...args)
    }
  },

  /**
   * 输出错误信息
   * @param {...any} args - 错误参数
   */
  error: (...args) => {
    if (isDev) {
      console.error('[ERROR]', ...args)
    } else {
      // 生产环境可以发送到日志服务
      // sendToLogService('error', ...args)
    }
  },

  /**
   * 输出警告信息
   * @param {...any} args - 警告参数
   */
  warn: (...args) => {
    if (isDev) {
      console.warn('[WARN]', ...args)
    }
  },

  /**
   * 输出调试信息
   * @param {...any} args - 调试参数
   */
  debug: (...args) => {
    if (isDev) {
      console.debug('[DEBUG]', ...args)
    }
  },

  /**
   * 输出信息
   * @param {...any} args - 信息参数
   */
  info: (...args) => {
    if (isDev) {
      console.info('[INFO]', ...args)
    }
  }
}

export default logger

