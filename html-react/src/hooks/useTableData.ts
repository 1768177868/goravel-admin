import { useCallback, useState } from 'react'
import type { ApiResponse, PaginatedData, PaginationState } from '@/types'
import { normalizeEntity } from '@/utils/normalize'
import logger from '@/utils/logger'

interface UseTableDataOptions<T> {
  fetchApi: (params: Record<string, unknown>) => Promise<ApiResponse<PaginatedData<T>>>
  transformData?: ((row: T) => T) | null
  onLoadSuccess?: ((rows: T[], res: ApiResponse<PaginatedData<T>>) => void) | null
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
        const rawList = (res.data?.list ?? res.data?.data ?? []) as T[]
        let rows = Array.isArray(rawList) ? rawList : []

        if (normalizeRows) {
          rows = rows.map((row) => normalizeEntity(row))
        }
        if (transformData) {
          rows = rows.map((row) => transformData(row))
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
