import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Alert,
  App,
  Button,
  Card,
  Col,
  Descriptions,
  Form,
  Input,
  Row,
  Select,
  Space,
  Tabs,
  Typography,
  Upload,
} from 'antd'
import type { UploadFile } from 'antd/es/upload/interface'
import { InboxOutlined, SendOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import PageContainer from '@/components/PageContainer'
import {
  aiLabAudio,
  aiLabImage,
  aiLabText,
  aiLabTranscription,
  aiLabVision,
  getAiLabStatus,
  type AiLabStatus,
} from '@/api/aiLab'

function toDataUrl(mimeType: string, base64: string) {
  return `data:${mimeType};base64,${base64}`
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

  const aiEnabled = status?.ai_enabled ?? false

  const runAction = async (action: () => Promise<void>) => {
    try {
      await action()
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { error_code?: string; message?: string } } }
      const code = apiErr.response?.data?.error_code
      if (code === 'ai_lab_rate_limited') {
        message.error(t('common.ai_lab_rate_limited'))
        return
      }
      message.error(apiErr.response?.data?.message || t('common.operation_failed'))
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

  const modelItems = useMemo(() => {
    if (!status) return []
    return [
      { label: 'Text', value: status.text_model || '—' },
      { label: 'Image', value: status.image_model || '—' },
      { label: 'Audio', value: status.audio_model || '—' },
      { label: 'STT', value: status.transcription_model || '—' },
    ]
  }, [status])

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
            <div style={{ marginTop: 8, color: 'var(--ant-color-text-secondary)', fontSize: 13 }}>
              {t('ai_lab.rate_limit_hint', {
                perMinute: status.rate_limit_per_minute,
                perDay: status.rate_limit_per_day,
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
                <Card>
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
                  {textResult && (
                    <Form.Item label={t('ai_lab.result')} style={{ marginTop: 16 }}>
                      <Input.TextArea rows={8} value={textResult} readOnly />
                    </Form.Item>
                  )}
                </Card>
              ),
            },
            {
              key: 'vision',
              label: t('ai_lab.tab_vision'),
              children: (
                <Card>
                  <Row gutter={16}>
                    <Col xs={24} md={10}>
                      <Upload.Dragger
                        maxCount={1}
                        accept="image/*"
                        beforeUpload={() => false}
                        fileList={visionFile ? [visionFile] : []}
                        onChange={({ fileList }) => setVisionFile(fileList[0] ?? null)}
                      >
                        <p className="ant-upload-drag-icon">
                          <InboxOutlined />
                        </p>
                        <p className="ant-upload-text">{t('ai_lab.upload_image')}</p>
                        <p className="ant-upload-hint">{t('ai_lab.upload_image_hint')}</p>
                      </Upload.Dragger>
                    </Col>
                    <Col xs={24} md={14}>
                      <Form layout="vertical">
                        <Form.Item label={t('ai_lab.prompt')}>
                          <Input.TextArea
                            rows={4}
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
                    </Col>
                  </Row>
                  {visionResult && (
                    <Form.Item label={t('ai_lab.result')} style={{ marginTop: 16 }}>
                      <Input.TextArea rows={8} value={visionResult} readOnly />
                    </Form.Item>
                  )}
                </Card>
              ),
            },
            {
              key: 'image',
              label: t('ai_lab.tab_image'),
              children: (
                <Card>
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
                    <div style={{ marginTop: 16 }}>
                      <div style={{ marginBottom: 8 }}>{t('ai_lab.preview_image')}</div>
                      <img
                        src={imagePreview}
                        alt="generated"
                        style={{ maxWidth: '100%', maxHeight: 480, borderRadius: 8 }}
                      />
                    </div>
                  )}
                </Card>
              ),
            },
            {
              key: 'audio',
              label: t('ai_lab.tab_audio'),
              children: (
                <Card>
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
                    <div style={{ marginTop: 16 }}>
                      <div style={{ marginBottom: 8 }}>{t('ai_lab.preview_audio')}</div>
                      <audio controls src={audioPreview} style={{ width: '100%' }} />
                    </div>
                  )}
                </Card>
              ),
            },
            {
              key: 'transcription',
              label: t('ai_lab.tab_transcription'),
              children: (
                <Card>
                  <Row gutter={16}>
                    <Col xs={24} md={10}>
                      <Upload.Dragger
                        maxCount={1}
                        accept="audio/*"
                        beforeUpload={() => false}
                        fileList={transcriptionFile ? [transcriptionFile] : []}
                        onChange={({ fileList }) => setTranscriptionFile(fileList[0] ?? null)}
                      >
                        <p className="ant-upload-drag-icon">
                          <InboxOutlined />
                        </p>
                        <p className="ant-upload-text">{t('ai_lab.upload_audio')}</p>
                        <p className="ant-upload-hint">{t('ai_lab.upload_audio_hint')}</p>
                      </Upload.Dragger>
                    </Col>
                    <Col xs={24} md={14}>
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
                    </Col>
                  </Row>
                  {transcriptionResult && (
                    <Form.Item label={t('ai_lab.result')} style={{ marginTop: 16 }}>
                      <Input.TextArea rows={8} value={transcriptionResult} readOnly />
                    </Form.Item>
                  )}
                </Card>
              ),
            },
          ]}
        />
      </Space>
    </PageContainer>
  )
}
