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
  