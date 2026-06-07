import { describe, expect, it } from 'vitest'
import { extractTextFromMarkdown, renderContent, sanitizeHtml } from '../markdown'

describe('markdown utils', () => {
  it('应该净化 HTML 中的脚本、事件属性和危险标签', () => {
    const dirtyHtml = `
      <p onclick="alert(1)">hello</p>
      <img src="/api/admin/public/images/1" onerror="alert(2)">
      <iframe src="https://example.com"></iframe>
      <script>alert(3)</script>
    `

    const html = renderContent(dirtyHtml, 'html')

    expect(html).toContain('<p>hello</p>')
    expect(html).not.toContain('onclick')
    expect(html).not.toContain('onerror')
    expect(html).not.toContain('<iframe')
    expect(html).not.toContain('<script')
  })

  it('应该净化 Markdown 转换后的 HTML，同时保留安全标签', () => {
    const markdown = '**hello** <img src=x onerror=alert(1)> <script>alert(2)</script>'
    const html = renderContent(markdown, 'markdown')

    expect(html).toContain('<strong>hello</strong>')
    expect(html).not.toContain('onerror')
    expect(html).not.toContain('<script')
  })

  it('提取 HTML 文本前应先净化内容', () => {
    const text = extractTextFromMarkdown('<p>safe</p><script>alert(1)</script>')

    expect(text).toBe('safe')
  })

  it('sanitizeHtml 应该处理空内容', () => {
    expect(sanitizeHtml('')).toBe('')
    expect(sanitizeHtml(null)).toBe('')
  })
})
