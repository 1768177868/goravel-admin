import request from '../utils/request'

export function getFieldTypes() {
  return request({
    url: '/code-generator/field-types',
    method: 'get'
  })
}

export function previewCode(data) {
  return request({
    url: '/code-generator/preview',
    method: 'post',
    data
  })
}

export function generateCode(data) {
  return request({
    url: '/code-generator/generate',
    method: 'post',
    data
  })
}

export function saveCode(data) {
  return request({
    url: '/code-generator/save',
    method: 'post',
    data
  })
}
