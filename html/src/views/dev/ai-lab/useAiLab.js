import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  aiLabAudio,
  aiLabImage,
  aiLabText,
  aiLabTranscription,
  aiLabVision,
  getAiLabStatus,
} from '@/api/aiLab'

function toDataUrl(mimeType, base64) {
  return `data:${mimeType};base64,${base64}`
}

function extractErrorMessage(err, fallback) {
  if (err?.__handled) return null
  return err?.translatedMessage || err?.message || fallback
}

export function useAiLab() {
  const { t } = useI18n()

  const status = ref(null)
  const loadingStatus = ref(true)
  const activeTab = ref('text')

  const textPrompt = ref('用三句话介绍 Goravel 框架。')
  const systemPrompt = ref('')
  const textResult = ref('')
  const textLoading = ref(false)

  const visionPrompt = ref('请描述这张图片的内容。')
  const visionFile = ref(null)
  const visionResult = ref('')
  const visionLoading = ref(false)

  const imagePrompt = ref('A minimal flat icon of a CPU chip, blue gradient background')
  const imageSize = ref('square')
  const imagePreview = ref('')
  const imageLoading = ref(false)

  const audioPrompt = ref('欢迎使用 Goravel AI 实验室。')
  const audioVoice = ref('female')
  const audioPreview = ref('')
  const audioLoading = ref(false)

  const transcriptionFile = ref(null)
  const transcriptionLanguage = ref('')
  const transcriptionResult = ref('')
  const transcriptionLoading = ref(false)

  const aiEnabled = computed(() => status.value?.ai_enabled ?? false)

  const modelItems = computed(() => {
    if (!status.value) return []
    return [
      { label: 'Text', value: status.value.text_model || '—' },
      { label: 'Image', value: status.value.image_model || '—' },
      { label: 'Audio', value: status.value.audio_model || '—' },
      { label: 'STT', value: status.value.transcription_model || '—' },
    ]
  })

  const sizeOptions = computed(() => [
    { label: t('ai_lab.size_square'), value: 'square' },
    { label: t('ai_lab.size_portrait'), value: 'portrait' },
    { label: t('ai_lab.size_landscape'), value: 'landscape' },
  ])

  const voiceOptions = computed(() => [
    { label: t('ai_lab.voice_female'), value: 'female' },
    { label: t('ai_lab.voice_male'), value: 'male' },
  ])

  async function loadStatus() {
    loadingStatus.value = true
    try {
      const res = await getAiLabStatus()
      status.value = res.data ?? null
    } catch {
      status.value = null
    } finally {
      loadingStatus.value = false
    }
  }

  function showActionError(err) {
    if (err?.__handled) return
    const code = err?.errorCode || err?.response?.data?.error_code
    if (code === 'ai_lab_rate_limited') {
      ElMessage.error(t('common.ai_lab_rate_limited'))
      return
    }
    const msg = extractErrorMessage(err, t('common.operation_failed'))
    if (msg) ElMessage.error(msg)
  }

  async function handleText() {
    textLoading.value = true
    try {
      const res = await aiLabText({
        prompt: textPrompt.value,
        system_prompt: systemPrompt.value || undefined,
      })
      textResult.value = res.data?.text ?? ''
    } catch (err) {
      showActionError(err)
    } finally {
      textLoading.value = false
    }
  }

  function onVisionFileChange(uploadFile) {
    if (uploadFile?.status === 'ready' && uploadFile.raw) {
      visionFile.value = uploadFile.raw
    }
  }

  function onVisionFileRemove() {
    visionFile.value = null
  }

  async function handleVision() {
    if (!visionFile.value) {
      ElMessage.warning(t('common.file_required'))
      return
    }
    visionLoading.value = true
    try {
      const formData = new FormData()
      formData.append('file', visionFile.value)
      formData.append('prompt', visionPrompt.value)
      const res = await aiLabVision(formData)
      visionResult.value = res.data?.text ?? ''
    } catch (err) {
      showActionError(err)
    } finally {
      visionLoading.value = false
    }
  }

  async function handleImage() {
    imageLoading.value = true
    try {
      const res = await aiLabImage({ prompt: imagePrompt.value, size: imageSize.value })
      const data = res.data
      if (data?.content_base64 && data.mime_type) {
        imagePreview.value = toDataUrl(data.mime_type, data.content_base64)
      }
    } catch (err) {
      showActionError(err)
    } finally {
      imageLoading.value = false
    }
  }

  async function handleAudio() {
    audioLoading.value = true
    try {
      const res = await aiLabAudio({ prompt: audioPrompt.value, voice: audioVoice.value })
      const data = res.data
      if (data?.content_base64 && data.mime_type) {
        audioPreview.value = toDataUrl(data.mime_type, data.content_base64)
      }
    } catch (err) {
      showActionError(err)
    } finally {
      audioLoading.value = false
    }
  }

  function onTranscriptionFileChange(uploadFile) {
    if (uploadFile?.status === 'ready' && uploadFile.raw) {
      transcriptionFile.value = uploadFile.raw
    }
  }

  function onTranscriptionFileRemove() {
    transcriptionFile.value = null
  }

  async function handleTranscription() {
    if (!transcriptionFile.value) {
      ElMessage.warning(t('common.file_required'))
      return
    }
    transcriptionLoading.value = true
    try {
      const formData = new FormData()
      formData.append('file', transcriptionFile.value)
      if (transcriptionLanguage.value.trim()) {
        formData.append('language', transcriptionLanguage.value.trim())
      }
      const res = await aiLabTranscription(formData)
      transcriptionResult.value = res.data?.text ?? ''
    } catch (err) {
      showActionError(err)
    } finally {
      transcriptionLoading.value = false
    }
  }

  onMounted(() => {
    loadStatus()
  })

  return {
    status,
    loadingStatus,
    activeTab,
    aiEnabled,
    modelItems,
    sizeOptions,
    voiceOptions,
    textPrompt,
    systemPrompt,
    textResult,
    textLoading,
    visionPrompt,
    visionResult,
    visionLoading,
    imagePrompt,
    imageSize,
    imagePreview,
    imageLoading,
    audioPrompt,
    audioVoice,
    audioPreview,
    audioLoading,
    transcriptionLanguage,
    transcriptionResult,
    transcriptionLoading,
    handleText,
    handleVision,
    handleImage,
    handleAudio,
    handleTranscription,
    onVisionFileChange,
    onVisionFileRemove,
    onTranscriptionFileChange,
    onTranscriptionFileRemove,
  }
}
