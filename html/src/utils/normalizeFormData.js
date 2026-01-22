/**
 * 统一规范化表单数据
 * - 多选 select / checkbox → string[]
 * - 单选 select → string
 */
export function normalizeFormData(data, rules = {}) {
    const result = { ...data }
  
    Object.keys(rules).forEach((key) => {
      const type = rules[key]
  
      if (type === 'string-array') {
        result[key] = Array.isArray(data[key])
          ? data[key].map(String)
          : []
      }
  
      if (type === 'string') {
        result[key] = data[key] !== undefined && data[key] !== null
          ? String(data[key])
          : null
      }
    })
  
    return result
  }

/**
 * 从数据对象中获取字段值，自动处理 snake_case 和 PascalCase
 * @param {Object} data - 数据对象
 * @param {string} fieldName - 字段名（snake_case）
 * @param {*} defaultValue - 默认值，如果字段不存在则返回此值
 * @returns {*} 字段值或默认值
 * 
 * @example
 * const data = { name: 'test', Name: 'Test', ID: 1 }
 * getField(data, 'name') // 'test' (优先使用 snake_case)
 * getField(data, 'id') // 1 (自动查找 ID)
 * getField(data, 'code', '') // '' (不存在时返回默认值)
 */
export function getField(data, fieldName, defaultValue = undefined) {
  if (!data || typeof data !== 'object') {
    return defaultValue
  }
  
  // 优先使用原始字段名（snake_case）
  if (fieldName in data && data[fieldName] !== undefined && data[fieldName] !== null) {
    return data[fieldName]
  }
  
  // 转换为 PascalCase 查找
  const pascalCase = fieldName
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join('')
  
  if (pascalCase in data && data[pascalCase] !== undefined && data[pascalCase] !== null) {
    return data[pascalCase]
  }
  
  return defaultValue
}

/**
 * 批量映射字段，自动处理 snake_case 和 PascalCase
 * @param {Object} data - 源数据对象
 * @param {Object} fieldMap - 字段映射配置，key 为目标字段名，value 为配置对象或默认值
 * @param {Object} options - 选项
 * @param {boolean} options.strict - 严格模式，如果字段不存在且无默认值则抛出错误（默认 false）
 * @returns {Object} 映射后的对象
 * 
 * @example
 * const data = { Name: 'test', ID: 1, Code: 'ABC' }
 * mapFields(data, {
 *   name: '', // 简单默认值
 *   id: 0,
 *   code: { default: '', transform: (v) => v.trim() } // 带转换函数
 * })
 * // { name: 'test', id: 1, code: 'ABC' }
 * 
 * @example
 * // 使用配置对象
 * mapFields(data, {
 *   name: { default: '' },
 *   is_active: { default: true, transform: Boolean }
 * })
 */
export function mapFields(data, fieldMap, options = {}) {
  if (!data || typeof data !== 'object') {
    return {}
  }
  
  const result = {}
  const { strict = false } = options
  
  Object.keys(fieldMap).forEach(targetField => {
    const config = fieldMap[targetField]
    let defaultValue = undefined
    let transform = null
    
    // 如果配置是对象
    if (config && typeof config === 'object' && !Array.isArray(config)) {
      defaultValue = config.default
      transform = config.transform
    } else {
      // 如果配置是简单值，作为默认值
      defaultValue = config
    }
    
    // 获取字段值
    let value = getField(data, targetField, defaultValue)
    
    // 如果严格模式且值为 undefined，抛出错误
    if (strict && value === undefined) {
      throw new Error(`Field '${targetField}' is required but not found in data`)
    }
    
    // 应用转换函数
    if (transform && typeof transform === 'function') {
      value = transform(value)
    }
    
    result[targetField] = value
  })
  
  return result
}
  