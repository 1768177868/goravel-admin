import { useCallback, useEffect, useRef, useState } from 'react'
import { buildSearchParams } from '@/utils/buildSearchParams'
import { useTableData } from './useTableData'
import type { ApiResponse, PaginatedData } from '@/types'

export interface UseListPageOptions<T, S extends Record<string, unknown>> {
  fetchApi: (params: Record<string, unknown>) => Promise<ApiResponse<PaginatedData<T>>>
  initialSearchForm?: S
  fieldMapping?: Record<string, string>
  defaultSort?: string
  transformData?: ((row: T) => T) | null
  onLoadSuccess?: ((rows: T[], res: ApiResponse<PaginatedData<T>>) => void) | null
  buildParams?: ((searchForm: S, baseParams: Record<string, unknown>) => Record<string, unknown>) | null
  onSearch?: (() => void) | null
  onReset?: (() => void) | null
  selectionIdKey?: string
  normalizeRows?: boolean
  autoLoad?: boolean
}

function toOrderBy(
  field: string | undefined,
  order: 'ascend' | 'descend' | null | undefined,
  fieldMapping: Record<string, string>,
  defaultSort: string,
) {
  if (!field || !order) return defaultSort
  const mapped = fieldMapping[field] || field
  return `${mapped}:${order === 'ascend' ? 'asc' : 'desc'}`
}

export function useListPage<T = Record<string, unknown>, S extends Record<string, unknown> = Record<string, unknown>>(
  options: UseListPageOptions<T, S>,
) {
  const {
    fetchApi,
    initialSearchForm = {} as S,
    fieldMapping = {},
    defaultSort = 'id:desc',
    transformData = null,
    onLoadSuccess = null,
    buildParams = null,
    onSearch = null,
    onReset = null,
    selectionIdKey = 'id',
    normalizeRows = false,
    autoLoad = true,
  } = options

  const [searchForm, setSearchForm] = useState<S>({ ...initialSearchForm })
  const [selectedRows, setSelectedRows] = useState<T[]>([])
  const [orderBy, setOrderBy] = useState(defaultSort)

  const searchFormRef = useRef(searchForm)
  const orderByRef = useRef(orderBy)
  searchFormRef.current = searchForm
  orderByRef.current = orderBy

  const { pagination, setPagination, tableData, loading, loadData: baseLoadData } = useTableData<T>({
    fetchApi,
    transformData,
    onLoadSuccess,
    normalizeRows,
  })

  const loadData = useCallback(
    async (
      pageParams: { currentPage?: number; pageSize?: number } | null = null,
      orderOverride?: string,
    ) => {
      const page = pageParams?.currentPage ?? pagination.page
      const pageSize = pageParams?.pageSize ?? pagination.pageSize

      if (pageParams) {
        setPagination((prev) => ({
          ...prev,
          page,
          pageSize,
        }))
      }

      const baseParams = {
        page,
        page_size: pageSize,
        order_by: orderOverride ?? orderByRef.current,
      }

      const params =
        buildParams && typeof buildParams === 'function'
          ? buildParams(searchFormRef.current, baseParams)
          : buildSearchParams(searchFormRef.current, baseParams)

      await baseLoadData(params)
    },
    [pagination.page, pagination.pageSize, setPagination, buildParams, baseLoadData],
  )

  const loadDataRef = useRef(loadData)
  loadDataRef.current = loadData

  const refresh = useCallback(async () => {
    await loadData()
  }, [loadData])

  const handleSearch = useCallback(() => {
    onSearch?.()
    setPagination((prev) => ({ ...prev, page: 1 }))
    void loadData({ currentPage: 1 })
  }, [onSearch, setPagination, loadData])

  const handleReset = useCallback(() => {
    onReset?.()
    setSearchForm({ ...initialSearchForm })
    setOrderBy(defaultSort)
    orderByRef.current = defaultSort
    setPagination((prev) => ({ ...prev, page: 1 }))
    setTimeout(() => {
      void loadDataRef.current({ currentPage: 1 }, defaultSort)
    }, 0)
  }, [onReset, initialSearchForm, defaultSort, setPagination])

  const handleSortChange = useCallback(
    (field?: string, order?: 'ascend' | 'descend' | null) => {
      const next = toOrderBy(field, order, fieldMapping, defaultSort)
      setOrderBy(next)
      orderByRef.current = next
      setPagination((prev) => ({ ...prev, page: 1 }))
      void loadData({ currentPage: 1 }, next)
    },
    [fieldMapping, defaultSort, setPagination, loadData],
  )

  const selectedIds = selectedRows.map((row) => {
    const record = row as Record<string, unknown>
    return record[selectionIdKey] ?? record.ID
  })

  useEffect(() => {
    if (autoLoad) {
      void loadDataRef.current()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- mount-only
  }, [])

  return {
    pagination,
    setPagination,
    tableData,
    loading,
    searchForm,
    setSearchForm,
    selectedRows,
    setSelectedRows,
    selectedIds,
    orderBy,
    loadData,
    refresh,
    handleSearch,
    handleReset,
    handleSortChange,
  }
}
