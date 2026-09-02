import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Alert,
  App,
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  Select,
  Space,
  Tabs,
  Typography,
  Upload,
} from 'antd'
import type { UploadFile } from 'antd/es/upload/interface'
import { CloseCircleFilled, InboxOutlined, SendOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import PageContainer from '@/components/PageContainer'
import type { ApiError, ApiResponse } from '@/types'
import {
  aiLabAudio,
  aiLabImage,
  aiLabText,
  aiLabTranscription,
  aiLabVision,
  getAiLabStatus,
  type AiLabStatus,
} from '@/api/aiLab'
import './AiLab.css'

function toDataUrl(mimeType: string, base64: string) {
  return `data:${mimeType};base64,${base64}`
}

function ResultBlock({ title, value }: { title: string; value: string }) {
  if (!value) return null
  return (
    <div className="ai-lab-result">
      <div className="ai-lab-result__title">{title}</div>
      <Input.TextArea rows={8} value={value} readOnly />
    </div>
  )
}

export default function AiLab() {
  const { t } = useTranslation()
  const { message } = App.useApp()

  const [status, setStatus] = useState<AiLabStatus | null>(null)
  const [loadingStatus, setLoadingStatus] = useState(true)

  const [textPrompt, setTextPrompt] = useState('用三句话介绍 Goravel 框架。')
  const [systemPrompt, setSystemPrompt] = useState('')
  const [textResult, setTextResult] = useState('')
  const [textLoading, setTextLoading] = useState(false)

  const [visionPrompt, setVisionPrompt] = useState('请描述这张图片的内容。')
  const [visionFile, setVisionFile] = useState<UploadFile | null>(null)
  const [visionPreviewUrl, setVisionPreviewUrl] = useState('')
  const [visionResult, setVisionResult] = useState('')
  const [visionLoading, setVisionLoading] = useState(false)

  const [imagePrompt, setImagePrompt] = useState('A minimal flat icon of a CPU chip, blue gradient background')
  const [imageSize, setImageSize] = useState('square')
  const [imagePreview, setImagePreview] = useState('')
  const [imageLoading, setImageLoading] = useState(false)

  const [audioPrompt, setAudioPrompt] = useState('欢迎使用 Goravel AI 实验室。')
  const [audioVoice, setAudioVoice] = useState('female')
  const [audioPreview, setAudioPreview] = useState('')
  const [audioLoading, setAudioLoading] = useState(false)

  const [transcriptionFile, setTranscriptionFile] = useState<UploadFile | null>(null)
  const [transcriptionLanguage, setTranscriptionLanguage] = useState('')
  const [transcriptionResult, setTranscriptionResult] = useState('')
  const [transcriptionLoading, setTranscriptionLoading] = useState(false)

  const loadStatus = useCallback(async () => {
    setLoadingStatus(true)
    try {
      const res = await getAiLabStatus()
      setStatus(res.data ?? null)
    } catch {
      setStatus(null)
    } finally {
      setLoadingStatus(false)
    }
  }, [])

  useEffect(() => {
    void loadStatus()
  }, [loadStatus])

  useEffect(() => {
    if (!visionFile?.originFileObj) {
      setVisionPreviewUrl('')
      return
    }
    const url = URL.createObjectURL(visionFile.originFileObj)
    setVisionPreviewUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [visionFile])

  const aiEnabled = status?.ai_enabled ?? false

  const runAction = async (action: () => Promise<void>) => {
    try {
      await action()
    } catch (err: unknown) {
      const apiErr = err as ApiError
      if (apiErr.__handled) return
      const code = apiErr.errorCode || (apiErr.response?.data as ApiResponse | undefined)?.error_code
      if (code === 'ai_lab_rate_limited') {
        message.error(t('common.ai_lab_rate_limited'))
        return
      }
      message.error(apiErr.translatedMessage || apiErr.message || t('common.operation_failed'))
    }
  }

  const handleText = () =>
    runAction(async () => {
      setTextLoading(true)
      try {
        const res = await aiLabText({ prompt: textPrompt, system_prompt: systemPrompt || undefined })
        setTextResult(res.data?.text ?? '')
      } finally {
        setTextLoading(false)
      }
    })

  const handleVision = () =>
    runAction(async () => {
      if (!visionFile?.originFileObj) {
        message.warning(t('common.file_required'))
        return
      }
      setVisionLoading(true)
      try {
        const formData = new FormData()
        formData.append('file', visionFile.originFileObj)
        formData.append('prompt', visionPrompt)
        const res = await aiLabVision(formData)
        setVisionResult(res.data?.text ?? '')
      } finally {
        setVisionLoading(false)
      }
    })

  const clearVisionFile = () => setVisionFile(null)

  const handleImage = () =>
    runAction(async () => {
      setImageLoading(true)
      try {
        const res = await aiLabImage({ prompt: imagePrompt, size: imageSize })
        const data = res.data
        if (data?.content_base64 && data.mime_type) {
          setImagePreview(toDataUrl(data.mime_type, data.content_base64))
        }
      } finally {
        setImageLoading(false)
      }
    })

  const handleAudio = () =>
    runAction(async () => {
      setAudioLoading(true)
      try {
        const res = await aiLabAudio({ prompt: audioPrompt, voice: audioVoice })
        const data = res.data
        if (data?.content_base64 && data.mime_type) {
          setAudioPreview(toDataUrl(data.mime_type, data.content_base64))
        }
      } finally {
        setAudioLoading(false)
      }
    })

  const handleTranscription = () =>
    runAction(async () => {
      if (!transcriptionFile?.originFileObj) {
        message.warning(t('common.file_required'))
        return
      }
      setTranscriptionLoading(true)
      try {
        const formData = new FormData()
        formData.append('file', transcriptionFile.originFileObj)
        if (transcriptionLanguage.trim()) {
          formData.append('language', transcriptionLanguage.trim())
        }
        const res = await aiLabTranscription(formData)
        setTranscriptionResult(res.data?.text ?? '')
      } finally {
        setTranscriptionLoading(false)
      }
    })

  const clearTranscriptionFile = () => setTranscriptionFile(null)

  const modelItems = useMemo(() => {
    if (!status) return []
    return [
      { label: 'Text', value: status.text_model || '—' },
      { label: 'Image', value: status.image_model || '—' },
      { label: 'Audio', value: status.audio_model || '—' },
      { label: 'STT', value: status.transcription_model || '—' },
    ]
  }, [status])

  const visionMediaBox = (
    <div className={`ai-lab-media-box${visionFile ? ' ai-lab-media-box--filled' : ''}`}>
      {!visionFile ? (
        <Upload.Dragger
          className="ai-lab-upload-dragger"
          maxCount={1}
          accept="image/*"
          showUploadList={false}
          beforeUpload={() => false}
          onChange={({ fileList }) => setVisionFile(fileList[0] ?? null)}
        >
          <p className="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p className="ant-upload-text">{t('ai_lab.upload_image')}</p>
          <p className="ant-upload-hint">{t('ai_lab.upload_image_hint')}</p>
        </Upload.Dragger>
      ) : (
        <div className="ai-lab-media-preview">
          {visionPreviewUrl && <img src={visionPreviewUrl} alt="preview" />}
          <Button
            type="text"
            size="small"
            className="ai-lab-media-preview__remove"
            icon={<CloseCircleFilled />}
            aria-label={t('common.delete')}
            onClick={clearVisionFile}
          />
        </div>
      )}
    </div>
  )

  const transcriptionMediaBox = (
    <div className={`ai-lab-media-box${transcriptionFile ? ' ai-lab-media-box--filled' : ''}`}>
      {!transcriptionFile ? (
        <Upload.Dragger
          className="ai-lab-upload-dragger"
          maxCount={1}
          accept="audio/*"
          showUploadList={false}
          beforeUpload={() => false}
          onChange={({ fileList }) => setTranscriptionFile(fileList[0] ?? null)}
        >
          <p className="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p className="ant-upload-text">{t('ai_lab.upload_audio')}</p>
          <p className="ant-upload-hint">{t('ai_lab.upload_audio_hint')}</p>
        </Upload.Dragger>
      ) : (
        <div className="ai-lab-file-chip">
          <span className="ai-lab-file-chip__name" title={transcriptionFile.name}>
            {transcriptionFile.name}
          </span>
          <Button type="link" danger size="small" onClick={clearTranscriptionFile}>
            {t('common.delete')}
          </Button>
        </div>
      )}
    </div>
  )

  return (
    <PageContainer title={t('ai_lab.title')}>
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
          {t('ai_lab.subtitle')}
        </Typography.Paragraph>
        {!loadingStatus && !aiEnabled && (
          <Alert type="warning" showIcon message={t('ai_lab.not_configured')} />
        )}

        {status && (
          <Card size="small">
            <Descriptions size="small" column={{ xs: 1, sm: 2, md: 4 }} title={t('ai_lab.models')}>
              {modelItems.map((item) => (
                <Descriptions.Item key={item.label} label={item.label}>
                  {item.value}
                </Descriptions.Item>
              ))}
            </Descriptions>
            <div className="ai-lab-status-hint">
              {t('ai_lab.rate_limit_hint', {
                perMinute: status.rate_limit_per_minute,
                perDay: status.rate_limit_per_day,
                maxUpload: status.max_upload_mb ?? 10,
              })}
            </div>
          </Card>
        )}

        <Tabs
          items={[
            {
              key: 'text',
              label: t('ai_lab.tab_text'),
              children: (
                <Card bordered={false} className="ai-lab-tab-body">
                  <Form layout="vertical">
                    <Form.Item label={t('ai_lab.prompt')}>
                      <Input.TextArea
                        rows={4}
                        value={textPrompt}
                        onChange={(e) => setTextPrompt(e.target.value)}
                        placeholder={t('ai_lab.prompt_placeholder')}
                      />
                    </Form.Item>
                    <Form.Item label={t('ai_lab.system_prompt')}>
                      <Input.TextArea
                        rows={2}
                        value={systemPrompt}
                        onChange={(e) => setSystemPrompt(e.target.value)}
                        placeholder={t('ai_lab.system_prompt_placeholder')}
                      />
                    </Form.Item>
                    <Button
                      type="primary"
                      icon={<SendOutlined />}
                      loading={textLoading}
                      disabled={!aiEnabled}
                      onClick={() => void handleText()}
                    >
                      {t('ai_lab.submit')}
                    </Button>
                  </Form>
                  <ResultBlock title={t('ai_lab.result')} value={textResult} />
                </Card>
              ),
            },
            {
              key: 'vision',
              label: t('ai_lab.tab_vision'),
              children: (
                <Card bordered={false} className="ai-lab-tab-body">
                  <div className="ai-lab-workspace">
                    <div className="ai-lab-media">{visionMediaBox}</div>
                    <div className="ai-lab-form-panel">
                      <Form layout="vertical">
                        <Form.Item label={t('ai_lab.prompt')}>
                          <Input.TextArea
                            rows={5}
                            value={visionPrompt}
                            onChange={(e) => setVisionPrompt(e.target.value)}
                          />
                        </Form.Item>
                        <Button
                          type="primary"
                          loading={visionLoading}
                          disabled={!aiEnabled}
                          onClick={() => void handleVision()}
                        >
                          {t('ai_lab.submit')}
                        </Button>
                      </Form>
                    </div>
                  </div>
                  <ResultBlock title={t('ai_lab.result')} value={visionResult} />
                </Card>
              ),
            },
            {
              key: 'image',
              label: t('ai_lab.tab_image'),
              children: (
                <Card bordered={false} className="ai-lab-tab-body">
                  <Form layout="vertical">
                    <Form.Item label={t('ai_lab.prompt')}>
                      <Input.TextArea
                        rows={3}
                        value={imagePrompt}
                        onChange={(e) => setImagePrompt(e.target.value)}
                      />
                    </Form.Item>
                    <Form.Item label={t('ai_lab.size')}>
                      <Select
                        style={{ width: 200 }}
                        value={imageSize}
                        onChange={setImageSize}
                        options={[
                          { value: 'square', label: t('ai_lab.size_square') },
                          { value: 'portrait', label: t('ai_lab.size_portrait') },
                          { value: 'landscape', label: t('ai_lab.size_landscape') },
                        ]}
                      />
                    </Form.Item>
                    <Button
                      type="primary"
                      loading={imageLoading}
                      disabled={!aiEnabled}
                      onClick={() => void handleImage()}
                    >
                      {t('ai_lab.generate')}
                    </Button>
                  </Form>
                  {imagePreview && (
                    <div className="ai-lab-media-preview-block">
                      <div className="ai-lab-media-preview-block__title">{t('ai_lab.preview_image')}</div>
                      <img src={imagePreview} alt="generated" />
                    </div>
                  )}
                </Card>
              ),
            },
            {
              key: 'audio',
              label: t('ai_lab.tab_audio'),
              children: (
                <Card bordered={false} className="ai-lab-tab-body">
                  <Form layout="vertical">
                    <Form.Item label={t('ai_lab.prompt')}>
                      <Input.TextArea
                        rows={3}
                        value={audioPrompt}
                        onChange={(e) => setAudioPrompt(e.target.value)}
                      />
                    </Form.Item>
                    <Form.Item label={t('ai_lab.voice')}>
                      <Select
                        style={{ width: 200 }}
                        value={audioVoice}
                        onChange={setAudioVoice}
                        options={[
                          { value: 'female', label: t('ai_lab.voice_female') },
                          { value: 'male', label: t('ai_lab.voice_male') },
                        ]}
                      />
                    </Form.Item>
                    <Button
                      type="primary"
                      loading={audioLoading}
                      disabled={!aiEnabled}
                      onClick={() => void handleAudio()}
                    >
                      {t('ai_lab.generate')}
                    </Button>
                  </Form>
                  {audioPreview && (
                    <div className="ai-lab-media-preview-block">
                      <div className="ai-lab-media-preview-block__title">{t('ai_lab.preview_audio')}</div>
                      <audio controls src={audioPreview} />
                    </div>
                  )}
                </Card>
              ),
            },
            {
              key: 'transcription',
              label: t('ai_lab.tab_transcription'),
              children: (
                <Card bordered={false} className="ai-lab-tab-body">
                  <div className="ai-lab-workspace">
                    <div className="ai-lab-media">{transcriptionMediaBox}</div>
                    <div className="ai-lab-form-panel">
                      <Form layout="vertical">
                        <Form.Item label={t('ai_lab.language')}>
                          <Input
                            value={transcriptionLanguage}
                            onChange={(e) => setTranscriptionLanguage(e.target.value)}
                            placeholder={t('ai_lab.language_placeholder')}
                          />
                        </Form.Item>
                        <Button
                          type="primary"
                          loading={transcriptionLoading}
                          disabled={!aiEnabled}
                          onClick={() => void handleTranscription()}
                        >
                          {t('ai_lab.transcribe')}
                        </Button>
                      </Form>
                    </div>
                  </div>
                  <ResultBlock title={t('ai_lab.result')} value={transcriptionResult} />
                </Card>
              ),
            },
          ]}
        />
      </Space>
    </PageContainer>
  )
}
