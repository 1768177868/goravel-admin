import { createCRUDApi, extendApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import request from '@/utils/request'

const baseArticleApi = createCRUDApi('articles')

const articleApi = baseArticleApi

export async function getArticleList(params?: Record<string, unknown>) {
  return normalizeListResponse(await articleApi.list(params))
}

export const getArticleDetail = articleApi.detail

export const createArticle = articleApi.create

export const updateArticle = articleApi.update

export const deleteArticle = articleApi.delete
