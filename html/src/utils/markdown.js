/**
 * Markdown 和 HTML 内容处理工具
 * 
 * 功能：
 * 1. 自动判断内容是 HTML 还是 Markdown
 * 2. 将 Markdown 转换为 HTML
 * 3. 处理图片路径（相对路径转绝对路径）
 * 4. 提取纯文本（用于列表显示）
 * 
 * 使用示例：
 * ```javascript
 * import { renderContent, extractTextFromMarkdown } from '@/utils/markdown'
 * 
 * // 自动判断并渲染
 * const html = renderContent(content, 'auto')
 * 
 * // 强制指定类型
 * const html = renderContent(content, 'html')  // 或 'markdown'
 * 
 * // 提取纯文本
 * const text = extractTextFromMarkdown(content)
 * ```
 */

import { marked } from 'marked'
import { getApiBaseURL } from './env'

// 配置 marked 选项
marked.setOptions({
  breaks: true, // 支持 GitHub 风格的换行
  gfm: true, // 启用 GitHub 风格的 Markdown
  sanitize: false, // 允许 HTML（因为我们会在显示时处理）
  silent: true // 静默模式，不抛出错误
})

/**
 * 判断内容是 HTML 还是 Markdown
 * @param {string} content - 内容文本
 * @returns {'html'|'markdown'} 内容类型
 */
export function detectContentType(content) {
  if (!content) return 'markdown'
  
  // 检查是否包含常见的 HTML 标签（排除 markdown 可能产生的简单标签）
  const htmlTagPattern = /<(p|div|span|img|br|hr|h[1-6]|ul|ol|li|table|tr|td|th|thead|tbody|blockquote|pre|code|strong|em|a|iframe)[\s>]/i
  const hasHtmlTags = htmlTagPattern.test(content)
  
  // 检查是否包含完整的 HTML 结构（如 <p>...</p>）
  const htmlStructurePattern = /<[a-z]+[^>]*>.*<\/[a-z]+>/i
  const hasHtmlStructure = htmlStructurePattern.test(content)
  
  // 检查是否包含 HTML 属性（如 class, style, src 等）
  const htmlAttributePattern = /<[^>]+(class|style|src|href|id|data-)=["'][^"']*["'][^>]*>/i
  const hasHtmlAttributes = htmlAttributePattern.test(content)
  
  // 如果同时满足多个 HTML 特征，认为是 HTML
  if (hasHtmlTags && (hasHtmlStructure || hasHtmlAttributes)) {
    return 'html'
  }
  
  // 检查是否包含明显的 markdown 语法
  const markdownPatterns = [
    /^#{1,6}\s+/m,           // 标题
    /!\[.*?\]\(.*?\)/,       // 图片
    /\[.*?\]\(.*?\)/,        // 链接
    /^\s*[-*+]\s+/m,         // 无序列表
    /^\s*\d+\.\s+/m,         // 有序列表
    /```[\s\S]*?```/,       // 代码块
    /`[^`]+`/,               // 行内代码
    /^\s*>\s+/m,             // 引用
    /\*\*.*?\*\*/,           // 加粗
    /\*.*?\*/,               // 斜体
  ]
  
  const markdownScore = markdownPatterns.filter(pattern => pattern.test(content)).length
  
  // 如果 markdown 特征明显多于 HTML 特征，认为是 markdown
  if (markdownScore >= 2 && !hasHtmlStructure) {
    return 'markdown'
  }
  
  // 默认：如果有 HTML 标签，认为是 HTML；否则认为是 markdown
  return hasHtmlTags ? 'html' : 'markdown'
}

/**
 * 将 markdown 转换为 HTML，并处理图片路径
 * @param {string} markdown - markdown 文本
 * @returns {string} HTML 字符串
 */
export function markdownToHtml(markdown) {
  if (!markdown) return ''
  
  // 先转换 markdown 为 HTML
  let html = marked.parse(markdown)
  
  // 处理图片路径
  html = processImageUrls(html)
  
  return html
}

/**
 * 渲染内容（自动判断 HTML 或 Markdown）
 * @param {string} content - 内容文本
 * @param {string} contentType - 可选：强制指定内容类型 'html' | 'markdown' | 'auto'
 * @returns {string} HTML 字符串
 */
export function renderContent(content, contentType = 'auto') {
  if (!content) return ''
  
  let type = contentType
  if (type === 'auto') {
    type = detectContentType(content)
  }
  
  if (type === 'html') {
    // 直接处理 HTML 中的图片路径
    return processImageUrls(content)
  } else {
    // 转换为 HTML 并处理图片路径
    return markdownToHtml(content)
  }
}

/**
 * 处理 HTML 中的图片 URL，将相对路径转换为绝对路径
 * @param {string} html - HTML 字符串
 * @returns {string} 处理后的 HTML
 */
export function processImageUrls(html) {
  if (!html) return ''
  
  const baseURL = getApiBaseURL()
  if (!baseURL) return html
  
  const cleanBaseURL = baseURL.replace(/\/+$/, '')
  
  // 处理 markdown 图片语法 ![alt](url) 转换后的 <img src="url">
  // 匹配 src="/api/..." 或 src='/api/...' 的相对路径（不包括 http/https 开头的完整 URL）
  html = html.replace(
    /src=["']((?!https?:\/\/)(\/api\/[^"']+))["']/g,
    `src="${cleanBaseURL}$1"`
  )
  
  // 也处理 markdown 中直接写的图片链接（如果已经是完整 URL 则跳过）
  // 匹配 markdown 图片语法中的相对路径
  html = html.replace(
    /!\[([^\]]*)\]\((?!https?:\/\/)(\/api\/[^)]+)\)/g,
    (match, alt, url) => {
      return `![${alt}](${cleanBaseURL}${url})`
    }
  )
  
  return html
}

/**
 * 从 markdown 或 HTML 中提取纯文本（用于列表显示）
 * @param {string} content - markdown 或 HTML 内容
 * @returns {string} 纯文本
 */
export function extractTextFromMarkdown(content) {
  if (!content) return ''
  
  // 如果是 HTML，先提取文本
  if (content.includes('<')) {
    const temp = document.createElement('div')
    temp.innerHTML = content
    return temp.textContent || temp.innerText || ''
  }
  
  // 如果是 markdown，移除 markdown 语法
  return content
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, '') // 移除图片
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1') // 移除链接，保留文本
    .replace(/#{1,6}\s+/g, '') // 移除标题标记
    .replace(/\*\*([^*]+)\*\*/g, '$1') // 移除加粗
    .replace(/\*([^*]+)\*/g, '$1') // 移除斜体
    .replace(/`([^`]+)`/g, '$1') // 移除行内代码
    .replace(/```[\s\S]*?```/g, '') // 移除代码块
    .replace(/^\s*[-*+]\s+/gm, '') // 移除列表标记
    .replace(/^\s*\d+\.\s+/gm, '') // 移除有序列表标记
    .replace(/>\s+/g, '') // 移除引用标记
    .replace(/\n{2,}/g, '\n') // 合并多个换行
    .trim()
}
