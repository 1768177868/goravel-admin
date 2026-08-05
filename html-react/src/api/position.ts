import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

const positionApi = createCRUDApi('positions')

export async function getPositionList(params?: Record<string, unknown>) {
  const res = await positionApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const getPositionDetail = positionApi.detail
export const createPosition = positionApi.create
export const updatePosition = positionApi.update
export const deletePosition = positionApi.delete
