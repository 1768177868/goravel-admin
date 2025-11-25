import request from '../utils/request'

export function fetchOnlineAdmins() {
  return request({
    url: '/online-admins/count',
    method: 'get'
  })
}

