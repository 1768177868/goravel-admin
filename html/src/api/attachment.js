import request from '../utils/request'
import Storage from '../utils/storage'

// 获取附件列表
export function getAttachmentList(params) {
  return request({
    url: '/attachments',
    method: 'get',
    params
  })
}

// 普通文件上传（小文件）
export function uploadFile(file, onProgress) {
  const formData = new FormData()
  formData.append('file', file)

  return request({
    url: '/attachments/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    onUploadProgress: (progressEvent) => {
      if (onProgress && progressEvent.total) {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        onProgress(percentCompleted)
      }
    }
  })
}

// 大文件分片上传统一接口
// action: init（初始化）、upload（上传分片）、merge（合并分片）、progress（获取进度）
export function chunkUpload(action, data = {}, onProgress) {
  const isGet = action === 'progress'
  const config = {
    url: '/attachments/chunk',
    method: isGet ? 'get' : 'post',
    ...(isGet ? { params: { action, ...data } } : { data: { action, ...data } })
  }

  // 如果是上传分片，需要特殊处理 FormData
  if (action === 'upload') {
    const formData = new FormData()
    formData.append('action', 'upload')
    formData.append('chunk_id', data.chunk_id)
    formData.append('chunk_index', data.chunk_index)
    formData.append('chunk', data.chunk)
    
    config.data = formData
    config.headers = {
      'Content-Type': 'multipart/form-data'
    }
    
    if (onProgress) {
      config.onUploadProgress = (progressEvent) => {
        if (progressEvent.total) {
          const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(percentCompleted)
        }
      }
    }
  }

  return request(config)
}

// 初始化分片上传
export function initChunkUpload(filename, totalSize, chunkSize, totalChunks) {
  return chunkUpload('init', {
    filename,
    total_size: totalSize,
    chunk_size: chunkSize,
    total_chunks: totalChunks
  })
}

// 上传分片
export function uploadChunk(chunkID, chunkIndex, chunk, onProgress) {
  return chunkUpload('upload', {
    chunk_id: chunkID,
    chunk_index: chunkIndex,
    chunk
  }, onProgress)
}

// 合并分片
export function mergeChunks(chunkID, filename, mimeType, totalChunks) {
  // 如果 totalChunks 未提供，尝试从 localStorage 获取（断点续传场景）
  if (!totalChunks) {
    try {
      const chunkInfo = Storage.getItem(`chunk_${chunkID}`, null)
      if (chunkInfo && typeof chunkInfo === 'object') {
        totalChunks = chunkInfo.total_chunks
      }
    } catch (e) {
      console.warn('Failed to get totalChunks from storage:', e)
    }
  }
  if (!totalChunks || totalChunks <= 0) {
    return Promise.reject(new Error('Total chunks is required'))
  }
  return chunkUpload('merge', {
    chunk_id: chunkID,
    filename,
    mime_type: mimeType,
    total_chunks: totalChunks
  })
}

// 获取分片上传进度
export function getChunkProgress(chunkID, totalChunks) {
  // 如果 chunkID 为空，直接返回，不调用后端接口
  if (!chunkID) {
    return Promise.reject(new Error('Chunk ID is empty'))
  }
  // 如果 totalChunks 未提供，尝试从 localStorage 获取（断点续传场景）
  if (!totalChunks) {
    try {
      const chunkInfo = Storage.getItem(`chunk_${chunkID}`, null)
      if (chunkInfo && typeof chunkInfo === 'object') {
        totalChunks = chunkInfo.total_chunks
      }
    } catch (e) {
      console.warn('Failed to get totalChunks from storage:', e)
    }
  }
  if (!totalChunks || totalChunks <= 0) {
    return Promise.reject(new Error('Total chunks is required'))
  }
  return chunkUpload('progress', { chunk_id: chunkID, total_chunks: totalChunks })
}

// 删除附件
export function deleteAttachment(id) {
  return request({
    url: `/attachments/${id}`,
    method: 'delete'
  })
}

// 批量删除附件
export function batchDeleteAttachments(ids) {
  return request({
    url: '/attachments/batch-delete',
    method: 'post',
    data: { ids }
  })
}

// 更新显示名称
export function updateDisplayName(id, displayName) {
  return request({
    url: `/attachments/${id}/display-name`,
    method: 'put',
    data: { display_name: displayName }
  })
}

// 创建上传进度 SSE URL
export function createUploadProgressSSE(chunkID, totalChunks, options = {}) {
  const { interval = 500 } = options
  const url = `/attachments/upload/progress?chunk_id=${chunkID}&total_chunks=${totalChunks}&interval=${interval}`
  return url
}

