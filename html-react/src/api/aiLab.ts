import request from '@/utils/request'

export interface AiLabStatus {
  ai_enabled: boolean
  text_model: string
  image_model: string
  audio_model: string
  transcription_model: string
  rate_limit_per_minute: number
  rate_limit_per_day: number
}

export interface AiLabTextResult {
  text: string
}

export interface AiLabMediaResult {
  mime_type: string
  content_base64: string
}

const aiLabTimeout = 180000

export function getAiLabStatus() {
  return request.get<AiLabStatus>('/ai-lab/status')
}

export function aiLabText(data: { prompt: string; system_prompt?: string }) {
  return request.post<AiLabTextResult>('/ai-lab/text', data, { timeout: aiLabTimeout })
}

export function aiLabVision(formData: FormData) {
  return request.post<AiLabTextResult>('/ai-lab/vision', formData, {
    timeout: aiLabTimeout,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function aiLabImage(data: { prompt: string; size?: string }) {
  return request.post<AiLabMediaResult>('/ai-lab/image', data, { timeout: aiLabTimeout })
}

export function aiLabAudio(data: { prompt: string; voice?: string }) {
  return request.post<AiLabMediaResult>('/ai-lab/audio', data, { timeout: aiLabTimeout })
}

export function aiLabTranscription(formData: FormData) {
  return request.post<AiLabTextResult>('/ai-lab/transcription', formData, {
    timeout: aiLabTimeout,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
