import { createCRUDApi } from '../utils/apiFactory'
import { normalizeListResponse } from '../utils/normalize'

const departmentApi = createCRUDApi('departments')

async function listWithNormalize(params) {
  const res = await departmentApi.list(params)
  return normalizeListResponse(res)
}

export const getDepartmentList = listWithNormalize
export const getDepartmentDetail = departmentApi.detail
export const createDepartment = departmentApi.create
export const updateDepartment = departmentApi.update
export const deleteDepartment = departmentApi.delete
