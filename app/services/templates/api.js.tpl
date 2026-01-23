import request from '../utils/request'

export function get<<.ModelName>>List(params) {
  return request({
    url: '/<<.ModuleName>>s',
    method: 'get',
    params
  })
}

export function get<<.ModelName>>Detail(id) {
  return request({
    url: `/<<.ModuleName>>s/${id}`,
    method: 'get'
  })
}

<<if .HasCreate>>
export function create<<.ModelName>>(data) {
  return request({
    url: '/<<.ModuleName>>s',
    method: 'post',
    data
  })
}
<<end>>

<<if .HasEdit>>
export function update<<.ModelName>>(id, data) {
  return request({
    url: `/<<.ModuleName>>s/${id}`,
    method: 'put',
    data
  })
}
<<end>>

<<if .HasDelete>>
export function delete<<.ModelName>>(id) {
  return request({
    url: `/<<.ModuleName>>s/${id}`,
    method: 'delete'
  })
}
<<end>>

<<if .HasExport>>
export function export<<.ModelName>>(params) {
  return request({
    url: '/<<.ModuleName>>s/export',
    method: 'post',
    data: params
  })
}
<<end>>
