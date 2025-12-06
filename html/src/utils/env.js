import logger from './logger'

/**
 * 环境变量配置
 */
const requiredEnvVars = []
const optionalEnvVars = ['VITE_API_BASE_URL', 'VITE_API_PREFIX']

/**
 * 验证环境变量
 * @param {boolean} strict - 是否严格模式（生产环境应该为 true）
 * @returns {boolean} 是否验证通过
 */
export function validateEnv(strict = import.meta.env.PROD) {
  const missing = []
  const warnings = []

  // 检查必需的环境变量
  requiredEnvVars.forEach(key => {
    if (!import.meta.env[key]) {
      missing.push(key)
    }
  })

  // 检查可选但推荐的环境变量
  optionalEnvVars.forEach(key => {
    if (!import.meta.env[key]) {
      warnings.push(key)
    }
  })

  // 输出警告
  if (warnings.length > 0) {
    logger.warn('Optional environment variables not set:', warnings)
  }

  // 检查必需变量
  if (missing.length > 0) {
    const message = `Missing required environment variables: ${missing.join(', ')}`
    logger.error(message)
    
    if (strict) {
      throw new Error(message)
    }
  }

  return missing.length === 0
}

/**
 * 获取环境变量值
 * @param {string} key - 环境变量键名
 * @param {any} defaultValue - 默认值
 * @returns {any} 环境变量值或默认值
 */
export function getEnv(key, defaultValue = null) {
  return import.meta.env[key] || defaultValue
}

/**
 * 获取 API 基础 URL
 * @returns {string} API 基础 URL
 */
export function getApiBaseURL() {
  return getEnv('VITE_API_BASE_URL', '')
}

/**
 * 获取 API 前缀
 * @returns {string} API 前缀
 */
export function getApiPrefix() {
  return getEnv('VITE_API_PREFIX', '/api/admin')
}

/**
 * 是否为开发环境
 * @returns {boolean} 是否为开发环境
 */
export function isDev() {
  return import.meta.env.DEV
}

/**
 * 是否为生产环境
 * @returns {boolean} 是否为生产环境
 */
export function isProd() {
  return import.meta.env.PROD
}

// 自动验证环境变量（仅在开发环境）
if (import.meta.env.DEV) {
  validateEnv(false)
}

export default {
  validateEnv,
  getEnv,
  getApiBaseURL,
  getApiPrefix,
  isDev,
  isProd
}

