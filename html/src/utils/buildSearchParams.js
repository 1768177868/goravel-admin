import { forOwn } from 'lodash-es'

/**
 * 构建搜索参数
 * 自动过滤空值，统一处理搜索表单字段
 * 
 * @param {Object} searchForm - 搜索表单对象
 * @param {Object} extraParams - 额外的参数（如排序等）
 * @returns {Object} 处理后的参数对象
 */
export function buildSearchParams(searchForm = {}, extraParams = {}) {
  const params = { ...extraParams }

  // 遍历搜索表单，只添加有值的字段
  forOwn(searchForm, (value, key) => {
    // 跳过空值
    if (value === '' || value === null || value === undefined) {
      return
    }

    // 如果是字符串，去除首尾空格后判断
    if (typeof value === 'string') {
      const trimmed = value.trim()
      if (trimmed) {
        params[key] = trimmed
      }
    } else {
      // 非字符串类型直接添加
      params[key] = value
    }
  })

  return params
}

