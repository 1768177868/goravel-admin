import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

const departmentApi = createCRUDApi('departments')

export async function getDepartmentList(params?: Record<string, unknown>) {
  const res = await departmentApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const getDepartmentDetail = departmentApi.detail
export const createDepartment = departmentApi.create
export const updateDepartment = departmentApi.update
export const deleteDepartment = departmentApi.delete
