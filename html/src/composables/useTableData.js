import { ref, reactive } from 'vue'

/**
 * 表格数据管理 composable
 * 自动处理分页、数据加载、total 更新等通用逻辑
 * 
 * @param {Object} options 配置选项
 * @param {Function} options.fetchApi - 获取数据的 API 函数
 * @param {Function} options.buildParams - 构建请求参数的自定义函数（可选）
 * @returns {Object} 返回分页、数据、加载状态和加载函数
 */
export function useTableData(options = {}) {
  const { fetchApi, buildParams = null } = options

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
        // 自动更新表格数据和总数
        tableData.value = res.data.list || res.data.data || []
        pagination.total = res.data.total || 0
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

