import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'

const articleApi = createCRUDApi('articles')

export async function getArticleList(params?: Record<string, unknown>) {
  const res = await articleApi.list(params)
  return normalizeListResponse(res)
}

export const getArticleDetail = articleApi.detail
export const createArticle = articleApi.create
export const updateArticle = articleApi.update
export const deleteArticle = articleApi.delete
