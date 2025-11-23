import request from '../utils/request'

// 获取部门列表
export function getDepartmentList(params) {
  return request({
    url: '/departments',
    method: 'get',
    params
  })
}

// 获取部门详情
export function getDepartmentDetail(id) {
  return request({
    url: `/departments/${id}`,
    method: 'get'
  })
}

// 创建部门
export function createDepartment(data) {
  return request({
    url: '/departments',
    method: 'post',
    data
  })
}

// 更新部门
export function updateDepartment(id, data) {
  return request({
    url: `/departments/${id}`,
    method: 'put',
    data
  })
}

// 删除部门
export function deleteDepartment(id) {
  return request({
    url: `/departments/${id}`,
    method: 'delete'
  })
}

