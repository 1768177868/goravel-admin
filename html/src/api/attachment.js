import request from '../utils/request'

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

// 初始化分片上传
export function initChunkUpload(filename, totalSize, chunkSize, totalChunks) {
  return request({
    url: '/attachments/chunk/init',
    method: 'post',
    data: {
      filename,
      total_size: totalSize,
      chunk_size: chunkSize,
      total_chunks: totalChunks
    }
  })
}

// 上传分片
export function uploadChunk(chunkID, chunkIndex, chunk, onProgress) {
  const formData = new FormData()
  formData.append('chunk_id', chunkID)
  formData.append('chunk_index', chunkIndex)
  formData.append('chunk', chunk)

  return request({
    url: '/attachments/chunk/upload',
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

// 合并分片
export function mergeChunks(chunkID, filename, mimeType) {
  return request({
    url: '/attachments/chunk/merge',
    method: 'post',
    data: {
      chunk_id: chunkID,
      filename,
      mime_type: mimeType
    }
  })
}

// 获取分片上传进度
export function getChunkProgress(chunkID) {
  return request({
    url: '/attachments/chunk/progress',
    method: 'get',
    params: { chunk_id: chunkID }
  })
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

