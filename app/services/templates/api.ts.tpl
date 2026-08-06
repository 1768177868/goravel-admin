import { createCRUDApi, extendApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import request from '@/utils/request'

const base<<.ModelName>>Api = createCRUDApi('<<.ModuleName>>s')

<<if .HasExport>>
const <<.ModuleName>>Api = extendApi(base<<.ModelName>>Api, {
  export: (params?: Record<string, unknown>) =>
    request({
      url: '/<<.ModuleName>>s/export',
      method: 'post',
      data: params,
    }),
})
<<else>>
const <<.ModuleName>>Api = base<<.ModelName>>Api
<<end>>

export async function get<<.ModelName>>List(params?: Record<string, unknown>) {
  return normalizeListResponse(await <<.ModuleName>>Api.list(params))
}

export const get<<.ModelName>>Detail = <<.ModuleName>>Api.detail

<<if .HasCreate>>
export const create<<.ModelName>> = <<.ModuleName>>Api.create
<<end>>

<<if .HasEdit>>
export const update<<.ModelName>> = <<.ModuleName>>Api.update
<<end>>

<<if .HasDelete>>
export const delete<<.ModelName>> = <<.ModuleName>>Api.delete
<<end>>

<<if .HasExport>>
export const export<<.ModelName>> = <<.ModuleName>>Api.export
<<end>>
