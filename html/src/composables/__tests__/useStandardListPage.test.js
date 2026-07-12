import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'

const mockLoadData = vi.fn()
const mockInitDefaultSort = vi.fn()
const mockHandleFormSuccess = vi.fn()
const mockHandleDelete = vi.fn()
const mockGetButtonState = vi.fn(() => ({ disabled: false }))

vi.mock('../useListPage', () => ({
  useListPage: vi.fn(() => ({
    pagination: ref({ page: 1, pageSize: 10, total: 0 }),
    tableData: ref([]),
    loading: ref(false),
    searchForm: {},
    loadData: mockLoadData,
    handleSearch: vi.fn(),
    handleReset: vi.fn(),
    handleSortChange: vi.fn(),
    initDefaultSort: mockInitDefaultSort
  }))
}))

vi.mock('../useCrud', () => ({
  useCrud: vi.fn(() => ({
    dialogVisible: ref(false),
    editId: ref(null),
    handleAdd: vi.fn(),
    handleEdit: vi.fn(),
    handleClose: vi.fn(),
    handleBatchDelete: vi.fn(),
    handleFormSuccess: mockHandleFormSuccess,
    handleDelete: mockHandleDelete
  }))
}))

vi.mock('../usePermission', () => ({
  usePermission: vi.fn(() => ({
    getButtonState: mockGetButtonState
  }))
}))

vi.mock('vue', async () => {
  const actual = await vi.importActual('vue')
  return {
    ...actual,
    onMounted: (fn) => fn()
  }
})

import { useListPage } from '../useListPage'
import { useCrud } from '../useCrud'
import { useStandardListPage } from '../useStandardListPage'

describe('useStandardListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('forwards list options to useListPage', () => {
    const transformData = (row) => row
    const buildParams = vi.fn()

    useStandardListPage({
      fetchApi: vi.fn(),
      initialSearchForm: { name: '' },
      defaultSort: 'id:desc',
      normalizeRows: false,
      transformData,
      buildParams,
      onSearch: vi.fn()
    })

    expect(useListPage).toHaveBeenCalledWith(
      expect.objectContaining({
        normalizeRows: false,
        transformData,
        buildParams
      })
    )
  })

  it('passes deleteApi and batchDeleteApi to useCrud', () => {
    const deleteApi = vi.fn()
    const batchDeleteApi = vi.fn()
    const beforeDelete = vi.fn()

    useStandardListPage({
      fetchApi: vi.fn(),
      deleteApi,
      batchDeleteApi,
      beforeDelete
    })

    expect(useCrud).toHaveBeenCalledWith(
      expect.objectContaining({
        deleteApi,
        batchDeleteApi,
        beforeDelete
      })
    )
  })

  it('loads data on mount when immediate is true', () => {
    useStandardListPage({
      fetchApi: vi.fn(),
      immediate: true
    })

    expect(mockInitDefaultSort).toHaveBeenCalled()
    expect(mockLoadData).toHaveBeenCalled()
  })

  it('handleFormSuccess reloads table data', () => {
    const { handleFormSuccess } = useStandardListPage({
      fetchApi: vi.fn(),
      immediate: false
    })

    handleFormSuccess()

    expect(mockHandleFormSuccess).toHaveBeenCalledWith(mockLoadData)
  })

  it('handleDelete delegates to crud delete', () => {
    const row = { id: 1 }
    const { handleDelete } = useStandardListPage({
      fetchApi: vi.fn(),
      deleteApi: vi.fn(),
      immediate: false
    })

    handleDelete(row)

    expect(mockHandleDelete).toHaveBeenCalledWith(row, mockLoadData)
  })

  it('exposes getButtonState from usePermission', () => {
    const { getButtonState } = useStandardListPage({
      fetchApi: vi.fn(),
      immediate: false
    })

    expect(getButtonState).toBe(mockGetButtonState)
  })
})
