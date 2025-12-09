import { ref, nextTick } from 'vue'

/**
 * 表格排序 Composable
 * @param {Object} options 配置选项
 * @param {Object} options.tableRef 表格引用
 * @param {Object} options.fieldMapping 字段映射（前端字段名 -> 数据库字段名）
 * @param {String} options.defaultSort 默认排序，格式：'field:direction' 或 'field1:direction1,field2:direction2'
 * @param {Function} options.onSortChange 排序变化回调函数
 * @returns {Object} 排序相关的状态和方法
 */
export function useTableSort(options = {}) {
  const {
    tableRef = null,
    fieldMapping = {},
    defaultSort = 'id:desc',
    onSortChange = null
  } = options

  // 排序配置（单字段排序）
  const sortConfig = ref({
    multiple: false,
    data: []
  })

  // 解析默认排序
  const parseDefaultSort = (sortStr) => {
    if (!sortStr) return []
    return sortStr.split(',').map(item => {
      const [field, order = 'desc'] = item.trim().split(':')
      return { field: field.trim(), order: order.trim() }
    })
  }

  // 初始化默认排序
  const initDefaultSort = () => {
    const defaultSorts = parseDefaultSort(defaultSort)
    // 单字段排序：只取第一个排序字段
    if (defaultSorts.length > 0) {
      sortConfig.value.data = [defaultSorts[0]]
    } else {
      sortConfig.value.data = []
    }
    
    // 设置表格的默认排序
    if (tableRef?.value) {
      nextTick(() => {
        if (tableRef.value) {
          tableRef.value.setSort(sortConfig.value.data)
        }
      })
    }
  }

  // 构建排序参数字符串（单字段排序）
  const buildOrderBy = () => {
    if (!sortConfig.value.data || sortConfig.value.data.length === 0) {
      // 如果没有排序，返回默认排序的第一个字段
      const defaultSorts = parseDefaultSort(defaultSort)
      if (defaultSorts.length > 0) {
        const sort = defaultSorts[0]
        const direction = sort.order === 'asc' ? 'asc' : 'desc'
        const dbField = fieldMapping[sort.field] || sort.field
        return `${dbField}:${direction}`
      }
      return defaultSort || ''
    }
    
    // 单字段排序：只取第一个排序字段
    const sort = sortConfig.value.data[0]
    const direction = sort.order === 'asc' ? 'asc' : 'desc'
    const dbField = fieldMapping[sort.field || sort.property] || (sort.field || sort.property)
    return `${dbField}:${direction}`
  }

  // 处理排序变化（单字段排序）
  const handleSortChange = ({ column, property, order, sortBy, sortList }) => {
    // 获取当前点击的字段
    const clickedField = property || column?.field || column?.property
    
    // 更新排序配置（优先使用 vxe-table 返回的 sortList）
    if (sortList && Array.isArray(sortList)) {
      // 单字段排序：只保留最后一个排序字段
      if (sortList.length > 0) {
        // 取最后一个（最新点击的）
        sortConfig.value.data = [sortList[sortList.length - 1]]
      } else {
        // 取消排序
        sortConfig.value.data = []
      }
    } else if (clickedField) {
      // 如果没有 sortList，使用当前列的信息更新
      if (order && (order === 'asc' || order === 'desc')) {
        // 单字段排序：清除之前的排序，只保留当前字段
        sortConfig.value.data = [{ field: clickedField, order }]
      } else {
        // 取消排序
        sortConfig.value.data = []
      }
    }
    
    // 调用回调函数
    if (onSortChange && typeof onSortChange === 'function') {
      onSortChange(sortConfig.value.data)
    }
  }

  // 重置排序
  const resetSort = () => {
    sortConfig.value.data = []
    if (tableRef?.value) {
      tableRef.value.clearSort()
    }
  }

  // 设置排序
  const setSort = (sorts) => {
    if (Array.isArray(sorts)) {
      sortConfig.value.data = sorts
      if (tableRef?.value) {
        tableRef.value.setSort(sorts)
      }
    }
  }

  return {
    sortConfig,
    buildOrderBy,
    handleSortChange,
    resetSort,
    setSort,
    initDefaultSort
  }
}

