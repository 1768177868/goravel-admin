import { onMounted } from 'vue'
import { useTreeListPage } from './useTreeListPage'
import { useCrud } from './useCrud'
import { usePermission } from './usePermission'

/**
 * Standard tree list page composable (department/menu style).
 */
export function useStandardTreeListPage(options = {}) {
  const {
    fetchApi,
    initialSearchForm = {},
    buildParams = null,
    normalizeRows = true,
    extractList = null,
    deleteApi = null,
    immediate = true,
    onLoadSuccess = null
  } = options

  const { getButtonState } = usePermission()

  const crud = useCrud({
    deleteApi: deleteApi || undefined
  })

  const tree = useTreeListPage({
    fetchApi,
    initialSearchForm,
    buildParams,
    normalizeRows,
    extractList: extractList || undefined,
    onLoadSuccess
  })

  const handleFormSuccess = () => {
    crud.handleFormSuccess(tree.loadData)
  }

  const handleDelete = (row) => {
    if (deleteApi) {
      crud.handleDelete(row, tree.loadData)
    }
  }

  if (immediate) {
    onMounted(() => {
      tree.loadData()
    })
  }

  return {
    ...tree,
    dialogVisible: crud.dialogVisible,
    editId: crud.editId,
    handleAdd: crud.handleAdd,
    handleEdit: crud.handleEdit,
    handleClose: crud.handleClose,
    getButtonState,
    handleFormSuccess,
    handleDelete
  }
}
