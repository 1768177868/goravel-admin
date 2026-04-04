import request from '../utils/request'

export function getPositionList(params) {
  return request({
    url: '/positions',
    method: 'get',
    params
  })
}

export function getPositionDetail(id) {
  return request({
    url: `/positions/${id}`,
    method: 'get'
  })
}

export function createPosition(data) {
  return request({
    url: '/positions',
    method: 'post',
    data
  })
}

export function updatePosition(id, data) {
  return request({
    url: `/positions/${id}`,
    method: 'put',
    data
  })
}

export function deletePosition(id) {
  return request({
    url: `/positions/${id}`,
    method: 'delete'
  })
}
