import { ref } from 'vue'
import { getOptions } from '../../api/option'
import request from '../../utils/request'

/**
 * 映射接口数据为 { label, value }，支持 optionLabelKey / optionValueKey
 */
function mapItemToOption(item, field) {
  const label = field.optionLabelKey
    ? (item[field.optionLabelKey] ?? item.label ?? item.name ?? item.title ?? String(item.id ?? item.ID))
    : (item.label ?? item.name ?? item.Name ?? item.title ?? String(item.id ?? item.ID))
  const value = field.optionValueKey
    ? (item[field.optionValueKey] ?? item.value ?? item.id ?? item.ID)
    : (item.value ?? item.id ?? item.ID)
  return { label, value: String(value ?? ''), disabled: item.disabled }
}

export function useFieldOptions() {
  const fieldOptionsCache = ref({})

  const loadFieldOptions = async (field) => {
    if (!field.apiUrl) return []
    
    const cacheKey = field.apiUrl + (field.apiParams ? JSON.stringify(field.apiParams) : '')
    if (fieldOptionsCache.value[cacheKey] !== undefined) {
      return fieldOptionsCache.value[cacheKey] || []
    }
    
    try {
      fieldOptionsCache.value[cacheKey] = null
      
      if (field.apiUrl.startsWith('/options')) {
        const url = new URL(field.apiUrl, window.location.origin)
        const type = url.searchParams.get('type')
        const params = {}
        for (const [key, value] of url.searchParams) {
          if (key !== 'type') {
            params[key] = value
          }
        }
        const res = await getOptions(type, params)
        if (res.data) {
          // 适配后端返回结构：有的返回 { data: { options: [] } }，有的直接返回 { data: [] }
          let options = []
          if (res.data.options) {
            options = res.data.options
          } else if (Array.isArray(res.data)) {
            options = res.data
          }
          fieldOptionsCache.value[cacheKey] = options
          return options
        }
      } else {
        // 通用接口：支持 apiUrl（可含查询串）或 apiUrl + apiParams
        const config = { url: field.apiUrl, method: 'get' }
        if (field.apiParams && typeof field.apiParams === 'object') config.params = field.apiParams
        const res = await request(config)
        const data = res.data || {}
        let options = []
        if (data.data && data.data.options) {
          options = data.data.options
        } else if (data.options) {
          options = data.options
        } else if (data.data?.list && Array.isArray(data.data.list)) {
          options = data.data.list.map(item => mapItemToOption(item, field))
        } else if (Array.isArray(data.data)) {
          options = data.data.map(item => mapItemToOption(item, field))
        } else if (Array.isArray(data)) {
          options = data.map(item => mapItemToOption(item, field))
        }
        fieldOptionsCache.value[cacheKey] = options

        // 调用 onOptionsLoaded 回调
        if (Array.isArray(field.__onOptionsLoaded)) {
          field.__onOptionsLoaded.forEach(fn => fn(options))
        }

        return options
      }
    } catch (error) {
      console.error('Load field options error:', error)
      fieldOptionsCache.value[cacheKey] = []
      return []
    }
    
    return []
  }

  const getFieldOptions = (field) => {
    if (!field) return []
    
    if (field.options && Array.isArray(field.options)) {
      return field.options
    }
    
    if (field.apiUrl) {
      const cacheKey = field.apiUrl + (field.apiParams ? JSON.stringify(field.apiParams) : '')
      if (fieldOptionsCache.value[cacheKey]) {
        return fieldOptionsCache.value[cacheKey]
      }
      loadFieldOptions(field)
      return []
    }
    
    if (field.optionsFn && typeof field.optionsFn === 'function') {
      try {
        return field.optionsFn()
      } catch (e) {
        return []
      }
    }
    
    return []
  }

  return {
    fieldOptionsCache,
    loadFieldOptions,
    getFieldOptions
  }
}

