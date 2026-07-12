import { ref } from 'vue'

/**
 * Toggle expand/collapse for el-table tree rows.
 */
export function useTreeExpand(treeListPageRef, tableData) {
  const isExpanded = ref(false)

  const handleToggleExpand = () => {
    isExpanded.value = !isExpanded.value
    const elTable = treeListPageRef.value?.tableRef
    if (!elTable) {
      return
    }

    const toggleNode = (rows) => {
      if (!Array.isArray(rows)) {
        return
      }
      rows.forEach((row) => {
        elTable.toggleRowExpansion(row, isExpanded.value)
        if (row.children?.length) {
          toggleNode(row.children)
        }
      })
    }

    toggleNode(tableData.value)
  }

  return {
    isExpanded,
    handleToggleExpand
  }
}
