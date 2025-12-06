/**
 * 输入验证工具
 * 提供常用的表单验证规则
 */

/**
 * 验证器函数类型
 * @typedef {function(any): boolean|string} Validator
 */

/**
 * 验证器集合
 */
export const validators = {
  /**
   * 必填验证
   * @param {any} value - 值
   * @returns {boolean|string} true 或错误消息
   */
  required: (value) => {
    if (value === '' || value === null || value === undefined) {
      return '此字段为必填项'
    }
    if (Array.isArray(value) && value.length === 0) {
      return '此字段为必填项'
    }
    return true
  },

  /**
   * 邮箱验证
   * @param {any} value - 值
   * @returns {boolean|string} true 或错误消息
   */
  email: (value) => {
    if (!value) return true // 空值由 required 验证
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return emailRegex.test(value) || '请输入有效的邮箱地址'
  },

  /**
   * 手机号验证（中国）
   * @param {any} value - 值
   * @returns {boolean|string} true 或错误消息
   */
  phone: (value) => {
    if (!value) return true
    const phoneRegex = /^1[3-9]\d{9}$/
    return phoneRegex.test(value) || '请输入有效的手机号码'
  },

  /**
   * 最小长度验证
   * @param {number} min - 最小长度
   * @returns {Validator} 验证器函数
   */
  minLength: (min) => (value) => {
    if (!value) return true
    const length = typeof value === 'string' ? value.length : (Array.isArray(value) ? value.length : 0)
    return length >= min || `最少需要 ${min} 个字符`
  },

  /**
   * 最大长度验证
   * @param {number} max - 最大长度
   * @returns {Validator} 验证器函数
   */
  maxLength: (max) => (value) => {
    if (!value) return true
    const length = typeof value === 'string' ? value.length : (Array.isArray(value) ? value.length : 0)
    return length <= max || `最多允许 ${max} 个字符`
  },

  /**
   * 长度范围验证
   * @param {number} min - 最小长度
   * @param {number} max - 最大长度
   * @returns {Validator} 验证器函数
   */
  length: (min, max) => (value) => {
    if (!value) return true
    const length = typeof value === 'string' ? value.length : (Array.isArray(value) ? value.length : 0)
    if (length < min || length > max) {
      return `长度必须在 ${min} 到 ${max} 个字符之间`
    }
    return true
  },

  /**
   * 数字验证
   * @param {any} value - 值
   * @returns {boolean|string} true 或错误消息
   */
  number: (value) => {
    if (!value) return true
    return !isNaN(value) && !isNaN(parseFloat(value)) || '请输入有效的数字'
  },

  /**
   * 整数验证
   * @param {any} value - 值
   * @returns {boolean|string} true 或错误消息
   */
  integer: (value) => {
    if (!value) return true
    return Number.isInteger(Number(value)) || '请输入整数'
  },

  /**
   * 最小值验证
   * @param {number} min - 最小值
   * @returns {Validator} 验证器函数
   */
  min: (min) => (value) => {
    if (!value) return true
    const num = Number(value)
    return !isNaN(num) && num >= min || `值不能小于 ${min}`
  },

  /**
   * 最大值验证
   * @param {number} max - 最大值
   * @returns {Validator} 验证器函数
   */
  max: (max) => (value) => {
    if (!value) return true
    const num = Number(value)
    return !isNaN(num) && num <= max || `值不能大于 ${max}`
  },

  /**
   * 范围验证
   * @param {number} min - 最小值
   * @param {number} max - 最大值
   * @returns {Validator} 验证器函数
   */
  range: (min, max) => (value) => {
    if (!value) return true
    const num = Number(value)
    if (isNaN(num)) return '请输入有效的数字'
    return (num >= min && num <= max) || `值必须在 ${min} 到 ${max} 之间`
  },

  /**
   * URL 验证
   * @param {any} value - 值
   * @returns {boolean|string} true 或错误消息
   */
  url: (value) => {
    if (!value) return true
    try {
      new URL(value)
      return true
    } catch {
      return '请输入有效的 URL 地址'
    }
  },

  /**
   * 正则表达式验证
   * @param {RegExp} pattern - 正则表达式
   * @param {string} message - 错误消息
   * @returns {Validator} 验证器函数
   */
  pattern: (pattern, message = '格式不正确') => (value) => {
    if (!value) return true
    return pattern.test(value) || message
  },

  /**
   * 密码强度验证
   * @param {number} minLength - 最小长度
   * @returns {Validator} 验证器函数
   */
  password: (minLength = 6) => (value) => {
    if (!value) return true
    if (value.length < minLength) {
      return `密码长度至少为 ${minLength} 个字符`
    }
    // 可以添加更多密码强度规则
    // if (!/[A-Z]/.test(value)) return '密码必须包含至少一个大写字母'
    // if (!/[a-z]/.test(value)) return '密码必须包含至少一个小写字母'
    // if (!/[0-9]/.test(value)) return '密码必须包含至少一个数字'
    return true
  },

  /**
   * 确认密码验证
   * @param {string} password - 原始密码
   * @returns {Validator} 验证器函数
   */
  confirmPassword: (password) => (value) => {
    if (!value) return true
    return value === password || '两次输入的密码不一致'
  }
}

/**
 * 创建自定义验证器
 * @param {function} validatorFn - 验证函数
 * @param {string} message - 错误消息
 * @returns {Validator} 验证器函数
 */
export function createValidator(validatorFn, message = '验证失败') {
  return (value) => {
    if (!value) return true
    return validatorFn(value) || message
  }
}

export default validators

