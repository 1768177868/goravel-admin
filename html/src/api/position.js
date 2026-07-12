import { createCRUDApi } from '../utils/apiFactory'
import { normalizeListResponse } from '../utils/normalize'

const positionApi = createCRUDApi('positions')

export async function getPositionList(params) {
  const res = await positionApi.list(params)
  return normalizeListResponse(res)
}

export const getPositionDetail = positionApi.detail
export const createPosition = positionApi.create
export const updatePosition = positionApi.update
export const deletePosition = positionApi.delete
