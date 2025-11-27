import request from '../utils/request'

/**
 * 获取下拉选项数据（统一接口）
 * @param {string} type - 选项类型：role, department, status, method, yes_no
 * @returns {Promise}
 */
export function getOptions(type) {
  return request({
    url: '/options',
    method: 'get',
    params: { type }
  })
}

