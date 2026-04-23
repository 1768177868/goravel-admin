import request from '../utils/request'

export function getTraceAggregate(params) {
  return request({
    url: '/observability/trace',
    method: 'get',
    params
  })
}

export function getSlowSqlTop(params) {
  return request({
    url: '/observability/slow-sql/top',
    method: 'get',
    params
  })
}

export function getAuditTimeline(params) {
  return request({
    url: '/observability/audit-timeline',
    method: 'get',
    params
  })
}

export function getQueueDashboard() {
  return request({
    url: '/observability/queue-dashboard',
    method: 'get'
  })
}
