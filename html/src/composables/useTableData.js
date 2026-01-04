import { ref, reactive } from 'vue'

/**
 * 表格数据管理 composable
 * 自动处理分页、数据加载、total 更新等通用逻辑
 * 
 * @param {Object} options 配置选项
 * @param {Function} options.fetchApi - 获取数据的 API 函数
 * @param {Function} options.buildParams - 构建请求参数的自定义函数（可选）
 * @param {Function} options.transformData - 数据转换函数（可选）
 * @param {Function} options.onLoadSuccess - 加载成功回调（可选）
 * @returns {Object} 返回分页、数据、加载状态和加载函数
 */
export function useTableData(options = {}) {
  const { fetchApi, buildParams = null, transformData = null, onLoadSuccess = null } = options

  // 分页对象（统一格式）
  const pagination = reactive({
    page: 1,
    pageSize: 10,
    total: 0
  })

  // 表格数据
  const tableData = ref([])

  // 加载状态
  const loading = ref(false)

  /**
   * 加载数据
   * @param {Object} extraParams - 完整的请求参数（如果提供，将覆盖基础参数）
   */
  const loadData = async (extraParams = null) => {
    if (!fetchApi) {
      console.error('useTableData: fetchApi is required')
      return
    }

    loading.value = true
    try {
      let params

      if (extraParams) {
        // 如果提供了完整参数，直接使用（但确保包含分页信息）
        params = {
          page: pagination.page,
          page_size: pagination.pageSize,
          ...extraParams
        }
        // 如果 extraParams 中包含了 page 和 page_size，同步到 pagination 对象
        if (extraParams.page !== undefined) {
          pagination.page = Number(extraParams.page) || 1
        }
        if (extraParams.page_size !== undefined) {
          pagination.pageSize = Number(extraParams.page_size) || 10
        }
      } else {
        // 构建基础参数
        const baseParams = {
          page: pagination.page,
          page_size: pagination.pageSize
        }

        // 如果提供了自定义参数构建函数，使用它
        params = buildParams ? buildParams(baseParams) : baseParams
      }

      const res = await fetchApi(params)

      if (res && res.data) {
        // 获取原始数据
        const rawList = res.data.list || res.data.data || []
        
        // 如果提供了数据转换函数，应用它
        if (transformData && typeof transformData === 'function') {
          tableData.value = rawList.map(transformData)
        } else {
          tableData.value = rawList
        }
        
        // 确保 total 是数字类型，并正确更新
        const total = res.data.total
        if (total !== undefined && total !== null) {
          pagination.total = Number(total) || 0
        } else {
          pagination.total = 0
        }
        
        // 如果提供了加载成功回调，调用它
        if (onLoadSuccess && typeof onLoadSuccess === 'function') {
          onLoadSuccess(res, tableData.value)
        }
      }
    } catch (error) {
      console.error('Load table data error:', error)
    } finally {
      loading.value = false
    }
  }

  /**
   * 重置到第一页并加载数据
   */
  const resetAndLoad = (extraParams = {}) => {
    pagination.page = 1
    loadData(extraParams)
  }

  return {
    pagination,
    tableData,
    loading,
    loadData,
    resetAndLoad
  }
}

