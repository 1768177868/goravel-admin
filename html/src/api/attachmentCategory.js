import { createCRUDApi } from '../utils/apiFactory'
import { normalizeListResponse } from '../utils/normalize'

const categoryApi = createCRUDApi('attachment-categories')

export async function getAttachmentCategoryList(params) {
  const res = await categoryApi.list(params)
  return normalizeListResponse(res)
}

export const getAttachmentCategoryDetail = categoryApi.detail
export const createAttachmentCategory = categoryApi.create
export const updateAttachmentCategory = categoryApi.update
export const deleteAttachmentCategory = categoryApi.delete
