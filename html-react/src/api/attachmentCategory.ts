import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

const categoryApi = createCRUDApi('attachment-categories')

export async function getAttachmentCategoryList(params?: Record<string, unknown>) {
  const res = await categoryApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const getAttachmentCategoryDetail = categoryApi.detail
export const createAttachmentCategory = categoryApi.create
export const updateAttachmentCategory = categoryApi.update
export const deleteAttachmentCategory = categoryApi.delete
