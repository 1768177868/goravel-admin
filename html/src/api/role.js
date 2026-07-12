import { createCRUDApi } from '../utils/apiFactory'
import { normalizeListResponse } from '../utils/normalize'

const roleApi = createCRUDApi('roles')

export async function getRoleList(params) {
  const res = await roleApi.list(params)
  return normalizeListResponse(res)
}

export const getRoleDetail = roleApi.detail
export const createRole = roleApi.create
export const updateRole = roleApi.update
export const deleteRole = roleApi.delete
