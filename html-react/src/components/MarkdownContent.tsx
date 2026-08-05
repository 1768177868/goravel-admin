import { useMemo } from 'react'
import { renderContent } from '@/utils/markdown'
import '@/components/MarkdownEditor.scss'

interface MarkdownContentProps {
  content?: string
  className?: string
}

export default function MarkdownContent({ content, className }: MarkdownContentProps) {
  const html = useMemo(() => renderContent(content || '', 'auto'), [content])

  if (!html) return null

  return (
    <div
      className={['markdown-content', className].filter(Boolean).join(' ')}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}
