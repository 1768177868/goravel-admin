import { ref, computed, watch } from 'vue'
import { getOptions } from '../../api/option'

export function useTreeSelect({ field, modelValue, onUpdate }) {
  const popoverVisible = ref(false)
  const filterText = ref('')
  const treeData = ref([])
  const fieldOptionsCache = ref({})

  // 获取树形选择器显示值
  const getTreeSelectDisplayValue = (fieldObj, value) => {
    if (!value) return ''
    
    const findNode = (data, targetId) => {
      for (const node of data) {
        const nodeId = node[fieldObj.treeProps?.value || 'id']
        if (nodeId == targetId) {
          return node[fieldObj.treeProps?.label || 'name']
        }
        if (node[fieldObj.treeProps?.children || 'children']) {
          const found = findNode(node[fieldObj.treeProps?.children || 'children'], targetId)
          if (found) return found
        }
      }
      return ''
    }
    
    return findNode(fieldObj.treeData || treeData.value || [], value) || ''
  }

  // 计算输入框显示值
  const inputValue = computed(() => {
    const selectedValue = modelValue.value
    if (selectedValue) {
      const displayValue = getTreeSelectDisplayValue(field, selectedValue)
      if (filterText.value && filterText.value !== displayValue) {
        return filterText.value
      }
      return displayValue
    }
    return filterText.value || ''
  })

  // 更新弹窗显示状态
  const updatePopoverVisible = (visible) => {
    popoverVisible.value = visible
    if (!visible && !modelValue.value) {
      filterText.value = ''
    }
  }

  // 切换弹窗显示状态
  const togglePopover = () => {
    popoverVisible.value = !popoverVisible.value
    if (!popoverVisible.value && !modelValue.value) {
      filterText.value = ''
    }
  }

  // 处理节点点击
  const handleNodeClick = (data) => {
    const valueKey = field.treeProps?.value || 'id'
    const value = data[valueKey]
    onUpdate(value)
    filterText.value = ''
  }

  // 处理清除
  const handleClear = () => {
    onUpdate(null)
    filterText.value = ''
  }

  // 处理输入
  const handleInput = (val) => {
    filterText.value = val || ''
    if (val && !popoverVisible.value) {
      togglePopover()
    }
  }

  // 获取过滤后的树形数据
  const getFilteredTreeData = (data) => {
    if (!filterText.value || filterText.value === '') {
      return data || []
    }
    
    const labelKey = field.treeProps?.label || 'name'
    const childrenKey = field.treeProps?.children || 'children'
    
    const filterNode = (node) => {
      const label = node[labelKey] || ''
      const matches = label.toLowerCase().includes(filterText.value.toLowerCase())
      
      if (node[childrenKey] && Array.isArray(node[childrenKey])) {
        const filteredChildren = node[childrenKey].map(child => filterNode(child)).filter(Boolean)
        if (matches || filteredChildren.length > 0) {
          return {
            ...node,
            [childrenKey]: filteredChildren
          }
        }
      } else if (matches) {
        return node
      }
      
      return null
    }
    
    return (data || []).map(node => filterNode(node)).filter(Boolean)
  }

  // 加载树形数据
  const loadData = async () => {
    if (!field.apiUrl) {
      if (field.treeData && Array.isArray(field.treeData)) {
        treeData.value = getFilteredTreeData(field.treeData)
        return
      }
      treeData.value = []
      return
    }
    
    const cacheKey = field.apiUrl
    if (fieldOptionsCache.value[cacheKey] !== undefined) {
      const data = fieldOptionsCache.value[cacheKey]
      if (Array.isArray(data)) {
        treeData.value = getFilteredTreeData(data)
      }
      return
    }
    
    try {
      fieldOptionsCache.value[cacheKey] = null
      
      if (field.apiUrl.startsWith('/options')) {
        const url = new URL(field.apiUrl, window.location.origin)
        const type = url.searchParams.get('type')
        const res = await getOptions(type)
        if (res.data) {
          if (res.data.options && Array.isArray(res.data.options)) {
            fieldOptionsCache.value[cacheKey] = res.data.options
            treeData.value = getFilteredTreeData(res.data.options)
          } else if (res.data.list && Array.isArray(res.data.list)) {
            fieldOptionsCache.value[cacheKey] = res.data.list
            treeData.value = getFilteredTreeData(res.data.list)
          }
        }
      } else {
        const res = await fetch(field.apiUrl, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json'
          }
        })
        if (res.ok) {
          const data = await res.json()
          let options = []
          if (data.data && data.data.options) {
            options = data.data.options
          } else if (data.data && data.data.list) {
            options = data.data.list
          } else if (data.options) {
            options = data.options
          } else if (Array.isArray(data.data)) {
            options = data.data
          } else if (Array.isArray(data)) {
            options = data
          }
          fieldOptionsCache.value[cacheKey] = options
          treeData.value = getFilteredTreeData(options)
        }
      }
    } catch (error) {
      console.error('Load tree select data error:', error)
      fieldOptionsCache.value[cacheKey] = []
      treeData.value = []
    }
  }

  // 监听过滤文本变化，更新树形数据
  watch(filterText, () => {
    const currentTreeData = typeof field.treeData === 'function' ? field.treeData() : field.treeData
    if (currentTreeData && Array.isArray(currentTreeData)) {
      treeData.value = getFilteredTreeData(currentTreeData)
    } else if (field.apiUrl) {
      const cacheKey = field.apiUrl
      if (fieldOptionsCache.value[cacheKey]) {
        treeData.value = getFilteredTreeData(fieldOptionsCache.value[cacheKey])
      }
    }
  })

  // 监听 field.treeData 变化，更新树形数据
  // 使用 computed 来访问 treeData，这样可以正确追踪 getter 的变化
  const treeDataGetter = computed(() => {
    if (typeof field.treeData === 'function') {
      return field.treeData()
    }
    return field.treeData
  })
  
  watch(treeDataGetter, (newTreeData) => {
    if (newTreeData && Array.isArray(newTreeData) && newTreeData.length > 0) {
      treeData.value = getFilteredTreeData(newTreeData)
    } else if (newTreeData && Array.isArray(newTreeData) && newTreeData.length === 0) {
      // 如果数据被清空，也更新
      treeData.value = []
    }
  }, { deep: true, immediate: true })

  // 初始化加载数据
  const initialTreeData = typeof field.treeData === 'function' ? field.treeData() : field.treeData
  if (initialTreeData && Array.isArray(initialTreeData) && initialTreeData.length > 0) {
    treeData.value = getFilteredTreeData(initialTreeData)
  } else if (field.apiUrl) {
    loadData()
  }

  return {
    popoverVisible,
    filterText,
    treeData,
    inputValue,
    updatePopoverVisible,
    togglePopover,
    handleNodeClick,
    handleClear,
    handleInput,
    loadData
  }
}

