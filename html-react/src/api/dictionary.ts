import { createCRUDApi, extendApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import request from '@/utils/request'
import type { ApiResponse, PaginatedData } from '@/types'

const baseDictionaryApi = createCRUDApi('dictionaries')

const dictionaryApi = extendApi(baseDictionaryApi, {
  getByType: (type: string) =>
    request({
      url: `/dictionaries/type/${type}`,
      method: 'get',
    }),
  getTypes: () =>
    request({
      url: '/dictionaries/types',
      method: 'get',
    }),
})

export async function getDictionaryList(params?: Record<string, unknown>) {
  const res = await dictionaryApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const getDictionaryDetail = dictionaryApi.detail
export const getDictionaryByType = dictionaryApi.getByType
export const getDictionaryTypes = dictionaryApi.getTypes
export const createDictionary = dictionaryApi.create
export const updateDictionary = dictionaryApi.update
export const deleteDictionary = dictionaryApi.delete
