import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

const permissionApi = createCRUDApi('permissions')

export async function getPermissionList(params?: Record<string, unknown>) {
  const res = await permissionApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const getPermissionDetail = permissionApi.detail
export const createPermission = permissionApi.create
export const updatePermission = permissionApi.update
export const deletePermission = permissionApi.delete
