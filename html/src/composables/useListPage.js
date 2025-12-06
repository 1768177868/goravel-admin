import { ref, reactive, watch } from 'vue'
import { useTableSort } from './useTableSort'
import { useApiRequest } from './useApiRequest'
import logger from '../utils/logger'

/**
 * 列表页面通用逻辑 composable
 * @param {Object} options 配置选项
 * @param {Function} options.fetchApi - 获取列表数据的 API 函数
 * @param {Object} options.initialSearchForm - 初始搜索表单数据
 * @param {Function} options.buildParams - 构建请求参数的自定义函数（可选）
 * @param {Function} options.transformData - 转换数据的自定义函数（可选）
 * @param {Function} options.onLoadSuccess - 加载成功后的回调（可选）
 * @param {Function} options.onLoadError - 加载失败后的回调（可选）
 * @param {Object} options.sortOptions - 排序配置（可选）
 * @returns {Object} 返回列表页面需要的响应式数据和方法
 */
export function useListPage(options = {}) {
  const {
    fetchApi,
    initialSearchForm = {},
    buildParams = null,
    transformData = null,
    onLoadSuccess = null,
    onLoadError = null,
    sortOptions = null
  } = options

  // 分页状态
  const pagination = reactive({
    page: 1,
    pageSize: 10,
    total: 0
  })

  // 表格数据
  const tableData = ref([])
  
  // 使用 API 请求 composable（提供请求取消功能和加载状态）
  const { request: apiRequest, cancel: cancelRequest, loading } = useApiRequest()

  // 搜索表单
  const searchForm = reactive({ ...initialSearchForm })

  // 排序相关（如果提供了排序配置）
  let buildOrderBy = null
  let resetSort = null
  let handleSortChange = null
  let initDefaultSort = null

  if (sortOptions) {
    const sortResult = useTableSort({
      tableRef: sortOptions.tableRef,
      fieldMapping: sortOptions.fieldMapping,
      defaultSort: sortOptions.defaultSort || 'id:desc',
      onSortChange: () => {
        pagination.page = 1
        loadData()
      }
    })
    buildOrderBy = sortResult.buildOrderBy
    resetSort = sortResult.resetSort
    handleSortChange = sortResult.handleSortChange
    initDefaultSort = sortResult.initDefaultSort
  }

  /**
   * 构建请求参数
   */
  const buildRequestParams = () => {
    if (buildParams) {
      // 使用自定义的参数构建函数
      return buildParams(searchForm, pagination, buildOrderBy ? buildOrderBy() : null)
    }

    // 默认参数构建逻辑
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }

    // 添加排序参数
    if (buildOrderBy) {
      const orderBy = buildOrderBy()
      if (orderBy) {
        params.order_by = orderBy
      }
    }

    // 添加搜索条件（只添加有值的字段）
    Object.keys(searchForm).forEach(key => {
      const value = searchForm[key]
      if (value !== '' && value !== null && value !== undefined) {
        // 如果是字符串，去除首尾空格
        if (typeof value === 'string' && value.trim()) {
          params[key] = value.trim()
        } else {
          params[key] = value
        }
      }
    })

    return params
  }

  /**
   * 加载数据
   */
  const loadData = async () => {
    if (!fetchApi) {
      logger.error('useListPage: fetchApi is required')
      return
    }

    try {
      const params = buildRequestParams()
      const res = await apiRequest(() => fetchApi(params))

      if (res && res.data) {
        let list = res.data.list || res.data.data || []
        
        // 如果提供了数据转换函数，使用它
        if (transformData) {
          list = list.map(item => transformData(item))
        }

        tableData.value = list
        pagination.total = res.data.total || res.data.meta?.total || 0

        // 调用成功回调
        if (onLoadSuccess) {
          onLoadSuccess(res, list)
        }
      }
    } catch (error) {
      // 如果是取消错误，不处理
      if (error?.name === 'AbortError' || error?.message === 'canceled') {
        return
      }
      
      logger.error('Load list error:', error)
      
      // 调用错误回调
      if (onLoadError) {
        onLoadError(error)
      }
    }
  }

  /**
   * 搜索处理
   */
  const handleSearch = () => {
    pagination.page = 1
    loadData()
  }

  /**
   * 重置搜索条件
   */
  const handleReset = () => {
    // 重置搜索表单
    Object.keys(searchForm).forEach(key => {
      searchForm[key] = initialSearchForm[key] !== undefined ? initialSearchForm[key] : ''
    })
    
    // 重置排序
    if (resetSort) {
      resetSort()
    }
    
    // 重置分页并加载数据
    pagination.page = 1
    loadData()
  }

  /**
   * 分页变化处理
   */
  const handlePageChange = ({ currentPage, pageSize }) => {
    pagination.page = currentPage
    pagination.pageSize = pageSize
    loadData()
  }

  return {
    // 响应式数据
    pagination,
    tableData,
    loading,
    searchForm,
    
    // 方法
    loadData,
    handleSearch,
    handleReset,
    handlePageChange,
    cancelRequest, // 导出取消请求方法
    
    // 排序相关（如果提供了排序配置）
    buildOrderBy,
    resetSort,
    handleSortChange,
    initDefaultSort
  }
}

