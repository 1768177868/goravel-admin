import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

const roleApi = createCRUDApi('roles')

export async function getRoleList(params?: Record<string, unknown>) {
  const res = await roleApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const getRoleDetail = roleApi.detail
export const createRole = roleApi.create
export const updateRole = roleApi.update
export const deleteRole = roleApi.delete
