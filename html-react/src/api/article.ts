import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

const articleApi = createCRUDApi('articles')

export async function getArticleList(params?: Record<string, unknown>) {
  const res = await articleApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const getArticleDetail = articleApi.detail
export const createArticle = articleApi.create
export const updateArticle = articleApi.update
export const deleteArticle = articleApi.delete
