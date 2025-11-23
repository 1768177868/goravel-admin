import request from '../utils/request'

// 获取字典列表
export function getDictionaryList(params) {
  return request({
    url: '/dictionaries',
    method: 'get',
    params
  })
}

// 获取字典详情
export function getDictionaryDetail(id) {
  return request({
    url: `/dictionaries/${id}`,
    method: 'get'
  })
}

// 根据类型获取字典
export function getDictionaryByType(type) {
  return request({
    url: `/dictionaries/type/${type}`,
    method: 'get'
  })
}

// 创建字典
export function createDictionary(data) {
  return request({
    url: '/dictionaries',
    method: 'post',
    data
  })
}

// 更新字典
export function updateDictionary(id, data) {
  return request({
    url: `/dictionaries/${id}`,
    method: 'put',
    data
  })
}

// 删除字典
export function deleteDictionary(id) {
  return request({
    url: `/dictionaries/${id}`,
    method: 'delete'
  })
}

