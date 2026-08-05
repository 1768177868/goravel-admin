import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

const blacklistApi = createCRUDApi('blacklists')

export async function getBlacklistList(params?: Record<string, unknown>) {
  const res = await blacklistApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const getBlacklistDetail = blacklistApi.detail
export const createBlacklist = blacklistApi.create
export const updateBlacklist = blacklistApi.update
export const deleteBlacklist = blacklistApi.delete
