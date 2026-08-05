import { createCRUDApi, extendApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import request from '@/utils/request'
import type { ApiResponse, PaginatedData } from '@/types'

const baseExportApi = createCRUDApi('exports')

const exportApi = extendApi(baseExportApi, {
  batchDelete: (ids: Array<string | number>) =>
    request({
      url: '/exports/batch-delete',
      method: 'post',
      data: { ids },
    }),
})

export async function getExportList(params?: Record<string, unknown>) {
  const res = await exportApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const deleteExport = exportApi.delete
export const batchDeleteExports = exportApi.batchDelete
