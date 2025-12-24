import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import Storage from '../utils/storage'

/**
 * 列设置 composable（简化版）
 * 
 * 使用示例：
 * const { tableColumns, showColumnSetting, allColumns, visibleColumns, handleSaveColumnSetting } = useColumnSetting('page_name', allTableColumns)
 * 
 * @param {string} storageKey localStorage 存储键名（会自动加上 '_column_setting' 后缀）
 * @param {ComputedRef|Array} tableColumns 表格列配置数组
 * @returns {Object} 返回列设置相关的状态和方法
 */
export function useColumnSetting(storageKey, tableColumns) {
  const { t } = useI18n()
  
  // 处理存储键
  const fullStorageKey = storageKey.includes('_column_setting') ? storageKey : `${storageKey}_column_setting`
  
  // 列设置对话框显示状态
  const showColumnSetting = ref(false)

  // 获取列的唯一标识
  const getColumnKey = (column) => column.field || column.slot || column.key || ''

  // 从 tableColumns 自动提取所有可配置的列（排除 checkbox 和 operation）
  const allColumns = computed(() => {
    const columns = Array.isArray(tableColumns) ? tableColumns : (tableColumns?.value || [])
    
    return columns
      .filter((column) => {
        if (column.type === 'checkbox') return false
        const key = getColumnKey(column)
        if (!key || key === 'operation') return false
        return true
      })
      .map((column) => ({
        key: getColumnKey(column),
        title: column.title || getColumnKey(column),
        required: column.required || false
      }))
  })

  // 默认显示所有列
  const defaultVisibleColumns = computed(() => allColumns.value.map(col => col.key))

  // 从 localStorage 加载或使用默认值
  const visibleColumns = ref(
    Storage.getItem(fullStorageKey) || []
  )
  
  // 如果没有存储值，使用默认值
  if (!visibleColumns.value.length && defaultVisibleColumns.value.length) {
    visibleColumns.value = [...defaultVisibleColumns.value]
  }

  // 保存列设置
  const handleSaveColumnSetting = (newVisibleColumns) => {
    visibleColumns.value = Array.isArray(newVisibleColumns) ? newVisibleColumns : []
    Storage.setItem(fullStorageKey, visibleColumns.value)
    showColumnSetting.value = false
    ElMessage.success(t('common.save_success'))
  }

  // 根据 visibleColumns 过滤显示的列
  const filteredColumns = computed(() => {
    const columns = Array.isArray(tableColumns) ? tableColumns : (tableColumns?.value || [])
    return columns.filter((column) => {
      // checkbox 和 operation 始终显示
      if (column.type === 'checkbox') return true
      const key = getColumnKey(column)
      if (!key || key === 'operation') return true
      // 其他列根据 visibleColumns 决定
      return visibleColumns.value.includes(key)
    })
  })

  return {
    // 过滤后的表格列（直接用于 vxe-table）
    tableColumns: filteredColumns,
    // 列设置对话框显示状态
    showColumnSetting,
    // 所有可配置的列（用于 ColumnSettingDialog）
    allColumns,
    // 当前可见的列 key 数组
    visibleColumns,
    // 默认可见列
    defaultVisibleColumns,
    // 保存列设置
    handleSaveColumnSetting
  }
}
