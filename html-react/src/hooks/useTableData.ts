import { useCallback, useState } from 'react'
import type { ApiResponse, ListFetchFn, PaginatedData, PaginationState } from '@/types'
import { normalizeEntity } from '@/utils/normalize'
import logger from '@/utils/logger'

interface UseTableDataOptions<T> {
  fetchApi: ListFetchFn
  transformData?: ((row: Record<string, unknown>) => T) | null
  onLoadSuccess?: ((rows: T[], res: ApiResponse<PaginatedData>) => void) | null
  normalizeRows?: boolean
}

export function useTableData<T = Record<string, unknown>>(options: UseTableDataOptions<T>) {
  const { fetchApi, transformData = null, onLoadSuccess = null, normalizeRows = false } = options

  const [pagination, setPagination] = useState<PaginationState>({
    page: 1,
    pageSize: 10,
    total: 0,
  })
  const [tableData, setTableData] = useState<T[]>([])
  const [loading, setLoading] = useState(false)

  const loadData = useCallback(
    async (params: Record<string, unknown> = {}) => {
      setLoading(true)
      try {
        const res = await fetchApi(params)
        const rawList = (res.data?.list ?? res.data?.data ?? []) as unknown[]
        let rows: T[] = []

        if (Array.isArray(rawList)) {
          rows = rawList.map((item) => {
            let row = item as Record<string, unknown>
            if (normalizeRows) {
              row = normalizeEntity(row) as Record<string, unknown>
            }
            if (transformData) {
              return transformData(row)
            }
            return row as T
          })
        }

        setTableData(rows)
        setPagination((prev) => ({
          ...prev,
          page: Number(params.page ?? prev.page),
          pageSize: Number(params.page_size ?? prev.pageSize),
          total: Number(res.data?.total ?? 0),
        }))

        onLoadSuccess?.(rows, res)
      } catch (error) {
        logger.error('loadData failed:', error)
        throw error
      } finally {
        setLoading(false)
      }
    },
    [fetchApi, transformData, onLoadSuccess, normalizeRows],
  )

  const resetAndLoad = useCallback(
    async (params: Record<string, unknown> = {}) => {
      setPagination((prev) => ({ ...prev, page: 1 }))
      await loadData({ ...params, page: 1, page_size: pagination.pageSize })
    },
    [loadData, pagination.pageSize],
  )

  return {
    pagination,
    setPagination,
    tableData,
    loading,
    loadData,
    resetAndLoad,
  }
}
