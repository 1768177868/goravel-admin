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

  // 排序配置
  const sortConfig = ref({
    multiple: true,
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
    sortConfig.value.data = defaultSorts
    
    // 设置表格的默认排序
    if (tableRef?.value) {
      nextTick(() => {
        if (tableRef.value) {
          tableRef.value.setSort(defaultSorts)
        }
      })
    }
  }

  // 构建排序参数字符串
  const buildOrderBy = () => {
    if (!sortConfig.value.data || sortConfig.value.data.length === 0) {
      return defaultSort || ''
    }
    
    return sortConfig.value.data
      .map(sort => {
        const direction = sort.order === 'asc' ? 'asc' : 'desc'
        // 映射字段名到数据库字段名
        const dbField = fieldMapping[sort.field] || sort.field
        return `${dbField}:${direction}`
      })
      .join(',')
  }

  // 处理排序变化
  const handleSortChange = ({ column, property, order, sortBy, sortList }) => {
    // 更新排序配置
    sortConfig.value.data = sortList || []
    
    // 调用回调函数
    if (onSortChange && typeof onSortChange === 'function') {
      onSortChange(sortList || [])
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

