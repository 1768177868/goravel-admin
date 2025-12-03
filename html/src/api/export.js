import request from '../utils/request'

// 获取导出记录列表
// 注意：request 已经配置了 baseURL 为 /api/admin，这里只写相对路径
export function getExportList(params) {
  return request({
    url: '/exports',
    method: 'get',
    params
  })
}

// 删除导出记录（同时删除源文件）
export function deleteExport(id) {
  return request({
    url: `/exports/${id}`,
    method: 'delete'
  })
}

// 批量删除导出记录（同时删除源文件）
export function batchDeleteExports(ids) {
  return request({
    url: '/exports/batch-delete',
    method: 'post',
    data: { ids }
  })
}

// 创建导出任务进度推送 SSE URL
export function createExportProgressSSE(exportID, options = {}) {
  const { interval = 1000 } = options
  const url = `/exports/${exportID}/progress?interval=${interval}`
  return url
}


