import { ref, reactive } from 'vue'
import { forOwn } from 'lodash-es'
import { buildSearchParams } from '../utils/buildSearchParams'
import { normalizeTreeList } from '../utils/normalize'
import { extractListFromResponse } from '../utils/extractListFromResponse'
import logger from '../utils/logger'

/**
 * Tree list page composable (no pagination).
 * For department/menu-style pages backed by tree APIs.
 */
export function useTreeListPage(options = {}) {
  const {
    fetchApi,
    initialSearchForm = {},
    buildParams = null,
    normalizeRows = true,
    extractList = extractListFromResponse,
    onSearch = null,
    onReset = null,
    onLoadSuccess = null
  } = options

  const searchForm = reactive({ ...initialSearchForm })
  const tableData = ref([])
  const loading = ref(false)

  const loadData = async () => {
    if (!fetchApi) {
      logger.error('useTreeListPage: fetchApi is required')
      return
    }

    loading.value = true
    try {
      const params = buildParams
        ? buildParams(searchForm, {})
        : buildSearchParams(searchForm, {})

      const res = await fetchApi(params)
      let list = extractList(res)

      if (normalizeRows) {
        list = normalizeTreeList(list)
      }

      tableData.value = list

      if (onLoadSuccess) {
        onLoadSuccess(res, tableData.value)
      }
    } catch (error) {
      logger.error('useTreeListPage load error:', error)
    } finally {
      loading.value = false
    }
  }

  const refresh = () => loadData()

  const clearSearchForm = () => {
    forOwn(searchForm, (value, key) => {
      searchForm[key] = initialSearchForm[key] !== undefined ? initialSearchForm[key] : ''
    })
  }

  const handleSearch = () => {
    if (onSearch) {
      onSearch()
    }
    loadData()
  }

  const handleReset = (_formData = null, options = {}) => {
    if (onReset) {
      onReset()
    }
    clearSearchForm()
    if (options.reload !== false) {
      loadData()
    }
  }

  return {
    searchForm,
    tableData,
    loading,
    loadData,
    refresh,
    handleSearch,
    handleReset,
    clearSearchForm
  }
}
