/**
 * XSS 防护工具
 * 提供 HTML 转义和清理功能
 */

/**
 * HTML 转义
 * @param {string} text - 要转义的文本
 * @returns {string} 转义后的文本
 */
export function escapeHtml(text) {
  if (typeof text !== 'string') {
    return text
  }

  const map = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  }

  return text.replace(/[&<>"']/g, (m) => map[m])
}

/**
 * 反转义 HTML
 * @param {string} text - 要反转义的文本
 * @returns {string} 反转义后的文本
 */
export function unescapeHtml(text) {
  if (typeof text !== 'string') {
    return text
  }

  const map = {
    '&amp;': '&',
    '&lt;': '<',
    '&gt;': '>',
    '&quot;': '"',
    '&#039;': "'"
  }

  return text.replace(/&amp;|&lt;|&gt;|&quot;|&#039;/g, (m) => map[m])
}

/**
 * 清理 HTML（移除危险标签和属性）
 * @param {string} html - 要清理的 HTML
 * @returns {string} 清理后的 HTML
 */
export function sanitizeHtml(html) {
  if (typeof html !== 'string') {
    return html
  }

  // 移除 script 标签
  html = html.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
  
  // 移除 on* 事件属性
  html = html.replace(/\s*on\w+\s*=\s*["'][^"']*["']/gi, '')
  html = html.replace(/\s*on\w+\s*=\s*[^\s>]*/gi, '')
  
  // 移除 javascript: 协议
  html = html.replace(/javascript:/gi, '')
  
  // 移除 data: 协议（可能包含恶意代码）
  html = html.replace(/data:text\/html/gi, '')
  
  return html
}

/**
 * 验证 URL 是否安全
 * @param {string} url - URL 地址
 * @returns {boolean} 是否安全
 */
export function isSafeUrl(url) {
  if (typeof url !== 'string') {
    return false
  }

  // 检查危险协议
  const dangerousProtocols = ['javascript:', 'data:', 'vbscript:', 'file:']
  const lowerUrl = url.toLowerCase().trim()
  
  for (const protocol of dangerousProtocols) {
    if (lowerUrl.startsWith(protocol)) {
      return false
    }
  }

  // 允许 http、https、mailto、tel 等安全协议
  const safeProtocols = ['http:', 'https:', 'mailto:', 'tel:']
  try {
    const urlObj = new URL(url)
    return safeProtocols.includes(urlObj.protocol.toLowerCase())
  } catch {
    // 相对 URL 认为是安全的
    return !url.includes('javascript:') && !url.includes('data:')
  }
}

/**
 * 清理用户输入
 * @param {any} input - 用户输入
 * @returns {any} 清理后的输入
 */
export function sanitizeInput(input) {
  if (typeof input === 'string') {
    return escapeHtml(input)
  }
  
  if (Array.isArray(input)) {
    return input.map(item => sanitizeInput(item))
  }
  
  if (input && typeof input === 'object') {
    const sanitized = {}
    for (const key in input) {
      if (Object.prototype.hasOwnProperty.call(input, key)) {
        sanitized[key] = sanitizeInput(input[key])
      }
    }
    return sanitized
  }
  
  return input
}

export default {
  escapeHtml,
  unescapeHtml,
  sanitizeHtml,
  isSafeUrl,
  sanitizeInput
}

