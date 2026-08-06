import type { TablePaginationConfig } from 'antd/es/table'
import type { SorterResult } from 'antd/es/table/interface'

export function getTableSortField<T>(
  sorter: SorterResult<T> | SorterResult<T>[] | undefined,
): { field: string; order: 'ascend' | 'descend' } | null {
  const single = Array.isArray(sorter) ? sorter[0] : sorter
  if (!single?.field || !single.order) return null
  const field = Array.isArray(single.field) ? String(single.field[0]) : String(single.field)
  return { field, order: single.order }
}

export function handlePaginatedTableChange<T>(options: {
  pager: TablePaginationConfig
  sorter: SorterResult<T> | SorterResult<T>[] | undefined
  pagination: { pageSize: number }
  loadData: (params: { currentPage?: number; pageSize?: number }) => void | Promise<void>
  handleSortChange: (field: string, order: 'ascend' | 'descend' | null | undefined) => void
}) {
  const sort = getTableSortField(options.sorter)
  if (sort) {
    options.handleSortChange(sort.field, sort.order)
    return
  }
  void options.loadData({
    currentPage: options.pager.current || 1,
    pageSize: options.pager.pageSize || options.pagination.pageSize,
  })
}
