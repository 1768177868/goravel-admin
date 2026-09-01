import request from '../utils/request'

const AI_LAB_TIMEOUT = 180000

export function getAiLabStatus() {
  return request({
    url: '/ai-lab/status',
    method: 'get',
  })
}

export function aiLabText(data) {
  return request({
    url: '/ai-lab/text',
    method: 'post',
    data,
    timeout: AI_LAB_TIMEOUT,
  })
}

export function aiLabVision(formData) {
  return request({
    url: '/ai-lab/vision',
    method: 'post',
    data: formData,
    timeout: AI_LAB_TIMEOUT,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function aiLabImage(data) {
  return request({
    url: '/ai-lab/image',
    method: 'post',
    data,
    timeout: AI_LAB_TIMEOUT,
  })
}

export function aiLabAudio(data) {
  return request({
    url: '/ai-lab/audio',
    method: 'post',
    data,
    timeout: AI_LAB_TIMEOUT,
  })
}

export function aiLabTranscription(formData) {
  return request({
    url: '/ai-lab/transcription',
    method: 'post',
    data: formData,
    timeout: AI_LAB_TIMEOUT,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
