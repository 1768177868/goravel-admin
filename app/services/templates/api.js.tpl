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

export function create<<.ModelName>>(data) {
  return request({
    url: '/<<.ModuleName>>s',
    method: 'post',
    data
  })
}

export function update<<.ModelName>>(id, data) {
  return request({
    url: `/<<.ModuleName>>s/${id}`,
    method: 'put',
    data
  })
}

export function delete<<.ModelName>>(id) {
  return request({
    url: `/<<.ModuleName>>s/${id}`,
    method: 'delete'
  })
}
