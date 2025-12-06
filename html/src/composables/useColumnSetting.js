import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import Storage from '../utils/storage'

/**
 * 列设置 composable
 * @param {Object} options 配置选项
 * @param {string} options.storageKey localStorage 存储键名
 * @param {Array} options.allColumns 所有列的配置数组，每个列需要包含 key 和 title
 * @param {Array} options.defaultVisibleColumns 默认可见的列 key 数组
 * @param {Array<string>} options.alwaysVisibleKeys 始终显示的列 key 数组（如 'checkbox', 'operation'）
 * @returns {Object} 返回列设置相关的状态和方法
 */
export function useColumnSetting(options) {
  const { t } = useI18n()
  const {
    storageKey,
    allColumns: providedAllColumns,
    defaultVisibleColumns = [],
    alwaysVisibleKeys = ['checkbox', 'operation']
  } = options

  // 列设置对话框显示状态
  const showColumnSetting = ref(false)

  // 从 localStorage 加载或使用默认值
  const visibleColumns = ref(
    Storage.getItem(storageKey, defaultVisibleColumns) || defaultVisibleColumns
  )

  // 所有可配置的列（用于对话框显示，不包括始终显示的列）
  const allColumns = computed(() => {
    return providedAllColumns.value.filter(
      (col) => !alwaysVisibleKeys.includes(col.key)
    )
  })

  // 保存列设置
  const handleSaveColumnSetting = (newVisibleColumns) => {
    // 如果传入了新的列数组，使用新的；否则使用当前的
    const columnsToSave = newVisibleColumns || visibleColumns.value
    visibleColumns.value = columnsToSave
    Storage.setItem(storageKey, columnsToSave)
    showColumnSetting.value = false
    ElMessage.success(t('common.save_success'))
  }

  // 根据 visibleColumns 过滤显示的列
  const getVisibleColumns = (allTableColumns) => {
    return computed(() => {
      return allTableColumns.value.filter((column) => {
        // 始终显示的列
        if (
          column.type === 'checkbox' ||
          alwaysVisibleKeys.includes(column.key)
        ) {
          return true
        }
        // 其他列根据 visibleColumns 决定
        return visibleColumns.value.includes(column.key)
      })
    })
  }

  return {
    showColumnSetting,
    visibleColumns,
    allColumns,
    handleSaveColumnSetting,
    getVisibleColumns
  }
}

