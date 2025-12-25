import { ref, nextTick, type Ref } from 'vue'

/**
 * 排序项
 */
export interface SortItem {
  field: string
  order: 'asc' | 'desc'
  property?: string
}

/**
 * 排序配置
 */
export interface SortConfig {
  multiple: boolean
  data: SortItem[]
}

/**
 * 表格排序选项
 */
export interface UseTableSortOptions {
  /** 表格引用 */
  tableRef?: Ref<any>
  /** 字段映射（前端字段名 -> 数据库字段名） */
  fieldMapping?: Record<string, string>
  /** 默认排序，格式：'field:direction' 或 'field1:direction1,field2:direction2' */
  defaultSort?: string
  /** 排序变化回调函数 */
  onSortChange?: (sortData: SortItem[]) => void
}

/**
 * 表格排序返回值
 */
export interface UseTableSortReturn {
  sortConfig: Ref<SortConfig>
  buildOrderBy: () => string
  handleSortChange: (params: {
    column?: any
    property?: string
    order?: 'asc' | 'desc' | null
    sortBy?: any
    sortList?: SortItem[]
  }) => void
  resetSort: () => void
  setSort: (sorts: SortItem[]) => void
  initDefaultSort: () => void
}

/**
 * 表格排序 Composable
 * @param options 配置选项
 * @returns 排序相关的状态和方法
 *
 * @example
 * const { sortConfig, buildOrderBy, handleSortChange, initDefaultSort } = useTableSort({
 *   tableRef,
 *   fieldMapping: { created_at: 'created_at' },
 *   defaultSort: 'id:desc',
 *   onSortChange: (sortData) => loadData()
 * })
 */
export function useTableSort(options: UseTableSortOptions = {}): UseTableSortReturn {
  const {
    tableRef = null,
    fieldMapping = {},
    defaultSort = 'id:desc',
    onSortChange = null
  } = options

  // 排序配置（单字段排序）
  const sortConfig = ref<SortConfig>({
    multiple: false,
    data: []
  })

  // 解析默认排序
  const parseDefaultSort = (sortStr: string): SortItem[] => {
    if (!sortStr) return []
    return sortStr.split(',').map((item) => {
      const [field, order = 'desc'] = item.trim().split(':')
      return { field: field.trim(), order: order.trim() as 'asc' | 'desc' }
    })
  }

  // 初始化默认排序
  const initDefaultSort = (): void => {
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
  const buildOrderBy = (): string => {
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
    const field = sort.field || sort.property || ''
    const dbField = fieldMapping[field] || field
    return `${dbField}:${direction}`
  }

  // 处理排序变化（单字段排序）
  const handleSortChange = ({
    column,
    property,
    order,
    sortList
  }: {
    column?: any
    property?: string
    order?: 'asc' | 'desc' | null
    sortBy?: any
    sortList?: SortItem[]
  }): void => {
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
  const resetSort = (): void => {
    sortConfig.value.data = []
    if (tableRef?.value) {
      tableRef.value.clearSort()
    }
  }

  // 设置排序
  const setSort = (sorts: SortItem[]): void => {
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

