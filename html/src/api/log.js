import request from '../utils/request'

// 操作日志
export function getOperationLogList(params) {
  return request({
    url: '/operation-logs',
    method: 'get',
    params
  })
}

export function getOperationLogTitleOptions() {
  return request({
    url: '/operation-logs/title-options',
    method: 'get'
  })
}

export function getOperationLogDetail(id) {
  return request({
    url: `/operation-logs/${id}`,
    method: 'get'
  })
}

export function deleteOperationLog(id) {
  return request({
    url: `/operation-logs/${id}`,
    method: 'delete'
  })
}

export function batchDeleteOperationLogs(ids) {
  return request({
    url: '/operation-logs/batch-delete',
    method: 'post',
    data: { ids: ids }
  })
}

export function cleanOperationLogs(params = {}) {
  return request({
    url: '/operation-logs/clean',
    method: 'post',
    data: params
  })
}

// 登录日志
export function getLoginLogList(params) {
  return request({
    url: '/login-logs',
    method: 'get',
    params
  })
}

export function getLoginLogDetail(id) {
  return request({
    url: `/login-logs/${id}`,
    method: 'get'
  })
}

export function deleteLoginLog(id) {
  return request({
    url: `/login-logs/${id}`,
    method: 'delete'
  })
}

export function batchDeleteLoginLogs(ids) {
  return request({
    url: '/login-logs/batch-delete',
    method: 'post',
    data: { ids: ids }
  })
}

export function cleanLoginLogs(params = {}) {
  return request({
    url: '/login-logs/clean',
    method: 'post',
    data: params
  })
}

// 系统日志
export function getSystemLogList(params) {
  return request({
    url: '/system-logs',
    method: 'get',
    params
  })
}

export function getSystemLogModuleOptions() {
  return request({
    url: '/system-logs/module-options',
    method: 'get'
  })
}

export function getSystemLogDetail(id) {
  return request({
    url: `/system-logs/${id}`,
    method: 'get'
  })
}

export function deleteSystemLog(id) {
  return request({
    url: `/system-logs/${id}`,
    method: 'delete'
  })
}

export function batchDeleteSystemLogs(ids) {
  return request({
    url: '/system-logs/batch-delete',
    method: 'post',
    data: { ids: ids }
  })
}

export function cleanSystemLogs(params = {}) {
  return request({
    url: '/system-logs/clean',
    method: 'post',
    data: params
  })
}

