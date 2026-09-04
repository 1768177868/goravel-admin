import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'

const articleApi = createCRUDApi('articles')

export async function getArticleList(params?: Record<string, unknown>) {
  return normalizeListResponse(await articleApi.list(params))
}

export const getArticleDetail = articleApi.detail

export const createArticle = articleApi.create

export const updateArticle = articleApi.update

export const deleteArticle = articleApi.delete
