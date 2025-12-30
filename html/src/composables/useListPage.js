import { ref, reactive, computed } from 'vue'
import { useTableData } from './useTableData'
import { useTableSort } from './useTableSort'
import { buildSearchParams } from '../utils/buildSearchParams'

/**
 * 列表页面通用 composable
 * 整合表格数据、排序、搜索等功能，简化列表页代码
 * 
 * @param {Object} options 配置选项
 * @param {Function} options.fetchApi - 获取数据的 API 函数
 * @param {Object} options.initialSearchForm - 初始搜索表单数据
 * @param {Object} options.fieldMapping - 字段映射（可选）
 * @param {String} options.defaultSort - 默认排序（可选）
 * @param {Object} options.tableRef - 表格引用（可选）
 * @param {Function} options.transformData - 数据转换函数（可选）
 * @param {Function} options.onLoadSuccess - 加载成功回调（可选）
 * @param {Function} options.buildParams - 自定义参数构建函数（可选），接收 (searchForm, baseParams) 参数，返回构建后的参数对象
 * @returns {Object} 返回列表页面需要的所有状态和方法
 */
export function useListPage(options = {}) {
  const {
    fetchApi,
    initialSearchForm = {},
    fieldMapping = {},
    defaultSort = 'id:desc',
    tableRef = null,
    transformData = null,
    onLoadSuccess = null,
    buildParams = null
  } = options

  // 搜索表单
  const searchForm = reactive({ ...initialSearchForm })

  // 使用表格数据 composable
  const { pagination, tableData, loading, loadData: baseLoadData, resetAndLoad: baseResetAndLoad } = useTableData({
    fetchApi,
    transformData,
    onLoadSuccess
  })

  // 使用排序 composable
  const { buildOrderBy, handleSortChange, resetSort, initDefaultSort } = useTableSort({
    tableRef,
    fieldMapping,
    defaultSort,
    onSortChange: () => {
      pagination.page = 1
      enhancedLoadData()
    }
  })

  /**
   * 增强的加载数据函数（自动添加排序和搜索参数）
   * @param {Object} pageParams - 可选的分页参数 { currentPage, pageSize }
   */
  const enhancedLoadData = async (pageParams = null) => {
    // 如果提供了分页参数，使用它们；否则使用 pagination 的值
    const page = pageParams?.currentPage ?? pagination.page
    const pageSize = pageParams?.pageSize ?? pagination.pageSize
    
    // 更新 pagination（确保同步）
    if (pageParams) {
      pagination.page = page
      pagination.pageSize = pageSize
    }
    
    // 基础参数
    const baseParams = {
      page,
      page_size: pageSize,
      order_by: buildOrderBy()
    }
    
    // 如果提供了自定义参数构建函数，使用它；否则使用默认的 buildSearchParams
    let params
    if (buildParams && typeof buildParams === 'function') {
      params = buildParams(searchForm, baseParams)
    } else {
      params = buildSearchParams(searchForm, baseParams)
    }
    
    await baseLoadData(params)
  }

  /**
   * 增强的重置并加载函数（自动重置到第一页）
   */
  const enhancedResetAndLoad = async () => {
    pagination.page = 1
    await enhancedLoadData()
  }

  /**
   * 搜索处理（自动重置到第一页并加载）
   */
  const handleSearch = () => {
    enhancedResetAndLoad()
  }

  /**
   * 清空搜索表单（不刷新数据）
   */
  const clearSearchForm = () => {
    // 重置搜索表单
    Object.keys(searchForm).forEach(key => {
      searchForm[key] = initialSearchForm[key] !== undefined ? initialSearchForm[key] : ''
    })
    // 重置排序
    resetSort()
  }

  /**
   * 重置搜索条件
   * @param {Object} formData - 重置后的表单数据（由 SearchForm 组件传递，可选）
   * @param {Object} options - 选项对象，包含 reload 属性（是否刷新数据，默认为 true）
   */
  const handleReset = (formData = null, options = {}) => {
    // 清空搜索表单
    clearSearchForm()
    // 如果需要刷新，则重置并加载（默认为 true，保持向后兼容）
    const shouldReload = options.reload !== false
    if (shouldReload) {
      enhancedResetAndLoad()
    }
  }

  return {
    // 响应式数据
    pagination,
    tableData,
    loading,
    searchForm,
    
    // 方法
    loadData: enhancedLoadData,
    resetAndLoad: enhancedResetAndLoad,
    handleSearch,
    handleReset,
    clearSearchForm, // 只清空表单，不刷新数据
    
    // 排序相关
    buildOrderBy,
    handleSortChange,
    resetSort,
    initDefaultSort
  }
}
