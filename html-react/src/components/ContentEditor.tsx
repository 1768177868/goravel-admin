import { useState } from 'react'
import { Radio } from 'antd'
import { useTranslation } from 'react-i18next'
import MarkdownEditor from '@/components/MarkdownEditor'
import WangEditor from '@/components/WangEditor'
import { detectContentType } from '@/utils/markdown'

export type ContentEditorMode = 'markdown' | 'rich'

interface ContentEditorProps {
  value?: string
  onChange?: (value: string) => void
  height?: number
  placeholder?: string
  defaultMode?: ContentEditorMode
}

function inferMode(content: string, defaultMode: ContentEditorMode): ContentEditorMode {
  if (!content) return defaultMode
  return detectContentType(content) === 'html' ? 'rich' : 'markdown'
}

export default function ContentEditor({
  value = '',
  onChange,
  height = 400,
  placeholder,
  defaultMode = 'markdown',
}: ContentEditorProps) {
  const { t } = useTranslation()
  const [mode, setMode] = useState<ContentEditorMode>(() => inferMode(value, defaultMode))

  const handleModeChange = (nextMode: ContentEditorMode) => {
    setMode(nextMode)
  }

  return (
    <div className="content-editor">
      <Radio.Group
        value={mode}
        optionType="button"
        buttonStyle="solid"
        size="small"
        style={{ marginBottom: 8 }}
        onChange={(e) => handleModeChange(e.target.value as ContentEditorMode)}
        options={[
          { label: t('editor.markdown'), value: 'markdown' },
          { label: t('editor.rich_text'), value: 'rich' },
        ]}
      />
      {mode === 'markdown' ? (
        <MarkdownEditor value={value} onChange={onChange} height={height} placeholder={placeholder} />
      ) : (
        <WangEditor value={value} onChange={onChange} height={height} placeholder={placeholder} />
      )}
    </div>
  )
}
