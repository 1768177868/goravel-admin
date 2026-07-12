import request from '../utils/request'
import { createCRUDApi, extendApi } from '../utils/apiFactory'
import { normalizeListResponse } from '../utils/normalize'

const baseDictionaryApi = createCRUDApi('dictionaries')

const dictionaryApi = extendApi(baseDictionaryApi, {
  getByType: (type) => {
    return request({
      url: `/dictionaries/type/${type}`,
      method: 'get'
    })
  },
  getTypes: () => {
    return request({
      url: '/dictionaries/types',
      method: 'get'
    })
  }
})

export async function getDictionaryList(params) {
  const res = await dictionaryApi.list(params)
  return normalizeListResponse(res)
}

export const getDictionaryDetail = dictionaryApi.detail
export const getDictionaryByType = dictionaryApi.getByType
export const getDictionaryTypes = dictionaryApi.getTypes
export const createDictionary = dictionaryApi.create
export const updateDictionary = dictionaryApi.update
export const deleteDictionary = dictionaryApi.delete
