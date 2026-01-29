import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

/**
 * 将 VxeTable 格式的列配置转换为 el-table 格式
 * 
 * @param {Ref|ComputedRef} tableColumnsConfig - useColumnSetting 返回的 tableColumns
 * @param {Ref|ComputedRef} visibleColumns - 可见列数组
 * @param {Ref|ComputedRef} columnOrder - 列顺序数组
 * @param {Ref|ComputedRef} fixedColumns - 冻结列配置对象
 * @returns {ComputedRef} el-table 格式的列配置数组
 */
export function useElTableColumns(tableColumnsConfig, visibleColumns, columnOrder, fixedColumns) {
  const { t } = useI18n()
  
  const tableColumns = computed(() => {
    const configs = tableColumnsConfig.value || []
    
    // 创建列映射
    const columnMap = {}
    configs.forEach(col => {
      const key = col.key || col.field || col.slot
      if (key) {
        columnMap[key] = col
      }
    })
    
    // 按照 columnOrder 排序（如果存在）
    const orderedKeys = columnOrder.value.length > 0 
      ? columnOrder.value.filter(key => columnMap[key])
      : configs.map(col => col.key || col.field || col.slot).filter(Boolean)
    
    // 构建最终列数组
    const result = []
    
    // 先添加 index 列（如果存在）
    if (columnMap['index']) {
      result.push({
        key: 'index',
        type: 'index',
        width: 60,
        label: t('table.seq')
      })
    }
    
    // 按顺序添加其他列（排除 index 和 operation）
    orderedKeys.forEach(key => {
      if (key === 'index' || key === 'operation') return
      
      const col = columnMap[key]
      if (!col) return
      
      // 检查是否可见
      if (!visibleColumns.value.includes(key)) return
      
      const elColumn = {
        key: key,
        prop: col.field,
        label: col.title,
        width: col.width,
        minWidth: col.minWidth,
        fixed: fixedColumns.value[key] || col.fixed,
        type: col.type
      }
      
      // 如果有 slot，标记需要自定义模板
      if (col.slot) {
        elColumn.slot = col.slot
      }
      
      result.push(elColumn)
    })
    
    // 最后添加 operation 列（如果存在）
    if (columnMap['operation']) {
      result.push({
        key: 'operation',
        label: t('table.operation'),
        width: 150,
        fixed: 'right',
        slot: 'operation'
      })
    }
    
    return result
  })
  
  return tableColumns
}