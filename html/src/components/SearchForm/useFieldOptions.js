import { ref } from 'vue'
import Storage from '../../utils/storage'
import { getOptions } from '../../api/option'

export function useFieldOptions() {
  const fieldOptionsCache = ref({})

  const loadFieldOptions = async (field) => {
    if (!field.apiUrl) return []
    
    const cacheKey = field.apiUrl
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
        const token = Storage.getItem('token', '') || ''
        const res = await fetch(field.apiUrl, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${typeof token === 'string' ? token.trim() : ''}`,
            'Content-Type': 'application/json'
          }
        })
        if (res.ok) {
          const data = await res.json()
          let options = []
          if (data.data && data.data.options) {
            options = data.data.options
          } else if (data.options) {
            options = data.options
          } else if (Array.isArray(data.data)) {
            options = data.data.map(item => ({
              label: item.name || item.Name || item.label || String(item.id || item.ID),
              value: String(item.id || item.ID || item.value)
            }))
          } else if (Array.isArray(data)) {
            options = data.map(item => ({
              label: item.name || item.Name || item.label || String(item.id || item.ID),
              value: String(item.id || item.ID || item.value)
            }))
          }
          fieldOptionsCache.value[cacheKey] = options
          return options
        }
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
      const cacheKey = field.apiUrl
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

