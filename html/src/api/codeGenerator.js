import request from '../utils/request'

export function getFieldTypes() {
  return request({
    url: '/code-generator/field-types',
    method: 'get'
  })
}

export function getTables() {
  return request({
    url: '/code-generator/tables',
    method: 'get'
  })
}

export function getTableColumns(tableName) {
  return request({
    url: '/code-generator/table-columns',
    method: 'get',
    params: { table_name: tableName }
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

export function installGeneratedModule(data) {
  return request({
    url: '/code-generator/install-module',
    method: 'post',
    data
  })
}

export function generateWithAI(data) {
  return request({
    url: '/code-generator/generate-with-ai',
    method: 'post',
    data,
    timeout: 300000 // 5 分钟（300 秒）
  })
}
