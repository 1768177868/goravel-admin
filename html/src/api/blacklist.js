import { createCRUDApi } from '../utils/apiFactory'
import { normalizeListResponse } from '../utils/normalize'

const blacklistApi = createCRUDApi('blacklists')

export async function getBlacklistList(params) {
  const res = await blacklistApi.list(params)
  return normalizeListResponse(res)
}

export const getBlacklistDetail = blacklistApi.detail
export const createBlacklist = blacklistApi.create
export const updateBlacklist = blacklistApi.update
export const deleteBlacklist = blacklistApi.delete
