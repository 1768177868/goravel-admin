import request from '../utils/request'

// 获取菜单列表（需要菜单权限，用于菜单管理页）
export function getMenuList(params) {
  return request({
    url: '/menus',
    method: 'get',
    params
  })
}

// 获取菜单树（仅登录即可，不校验菜单权限；用于角色/权限表单、下拉等）
export function getMenuTree() {
  return request({
    url: '/menus/tree',
    method: 'get'
  })
}

// 获取菜单详情
export function getMenuDetail(id) {
  return request({
    url: `/menus/${id}`,
    method: 'get'
  })
}

// 创建菜单
export function createMenu(data) {
  return request({
    url: '/menus',
    method: 'post',
    data
  })
}

// 更新菜单
export function updateMenu(id, data) {
  return request({
    url: `/menus/${id}`,
    method: 'put',
    data
  })
}

// 删除菜单
export function deleteMenu(id) {
  return request({
    url: `/menus/${id}`,
    method: 'delete'
  })
}

