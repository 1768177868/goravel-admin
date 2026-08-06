import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'

const permissionApi = createCRUDApi('permissions')

export async function getPermissionList(params?: Record<string, unknown>) {
  const res = await permissionApi.list(params)
  return normalizeListResponse(res)
}

export const getPermissionDetail = permissionApi.detail
export const createPermission = permissionApi.create
export const updatePermission = permissionApi.update
export const deletePermission = permissionApi.delete
