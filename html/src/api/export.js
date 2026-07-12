import request from '../utils/request'
import { createCRUDApi, extendApi } from '../utils/apiFactory'
import { normalizeListResponse } from '../utils/normalize'

const baseExportApi = createCRUDApi('exports')

const exportApi = extendApi(baseExportApi, {
  batchDelete: (ids) => {
    return request({
      url: '/exports/batch-delete',
      method: 'post',
      data: { ids }
    })
  }
})

export async function getExportList(params) {
  const res = await exportApi.list(params)
  return normalizeListResponse(res)
}

export const {
  delete: deleteExport,
  batchDelete: batchDeleteExports
} = exportApi

export function createExportProgressSSE(exportID, options = {}) {
  const { interval = 1000 } = options
  return `/exports/${exportID}/progress?interval=${interval}`
}
