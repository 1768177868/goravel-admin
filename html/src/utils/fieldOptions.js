/**
 * 通用的搜索表单字段选项配置
 * 用于在各个列表页面的搜索表单中复用
 */

/**
 * 获取状态选项（启用/禁用）
 * @param {Function} t - i18n 的 t 函数
 * @returns {Array} 选项数组
 */
export const getStatusOptions = (t) => {
  return [
    { label: t('common.enabled'), value: '1' },
    { label: t('common.disabled'), value: '0' }
  ]
}

/**
 * 获取 HTTP 方法选项
 * @returns {Array} 选项数组
 */
export const getMethodOptions = () => {
  return [
    { label: 'GET', value: 'GET' },
    { label: 'POST', value: 'POST' },
    { label: 'PUT', value: 'PUT' },
    { label: 'DELETE', value: 'DELETE' },
    { label: 'PATCH', value: 'PATCH' }
  ]
}

/**
 * 获取是否选项（是/否）
 * @param {Function} t - i18n 的 t 函数
 * @returns {Array} 选项数组
 */
export const getYesNoOptions = (t) => {
  return [
    { label: t('common.yes'), value: '1' },
    { label: t('common.no'), value: '0' }
  ]
}

