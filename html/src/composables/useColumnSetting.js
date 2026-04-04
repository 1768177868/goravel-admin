import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import Storage from '../utils/storage'

/**
 * 将缺失列按 tableColumns 定义顺序插入到当前顺序中（而非一律追加到末尾）
 */
function insertMissingKeysNaturally(middleKeys, allMiddleKeys) {
  const missing = allMiddleKeys.filter((key) => !middleKeys.includes(key))
  const result = [...middleKeys]
  for (const key of missing) {
    const targetIdx = allMiddleKeys.indexOf(key)
    let insertAt = result.length
    for (let i = targetIdx - 1; i >= 0; i--) {
      const prev = allMiddleKeys[i]
      const idx = result.indexOf(prev)
      if (idx !== -1) {
        insertAt = idx + 1
        break
      }
    }
    result.splice(insertAt, 0, key)
  }
  return result
}

/**
 * 强制 right 列紧跟在 left 列之后（用于修正本地缓存列顺序）
 */
function applyAdjacentPairs(middleKeys, pairs) {
  if (!pairs || !pairs.length) return middleKeys
  let result = [...middleKeys]
  for (const [left, right] of pairs) {
    const li = result.indexOf(left)
    const ri = result.indexOf(right)
    if (li === -1 || ri === -1) continue
    if (ri === li + 1) continue
    result = result.filter((k) => k !== right)
    const newLi = result.indexOf(left)
    result.splice(newLi + 1, 0, right)
  }
  return result
}

/**
 * 解析中间列顺序：去掉已删除列、插入新列、可选地应用相邻约束
 */
function resolveMiddleKeys(savedMiddle, allMiddleKeys, adjacentPairs) {
  let middleKeys = savedMiddle.filter((k) => allMiddleKeys.includes(k))
  if (middleKeys.length === 0) {
    middleKeys = [...allMiddleKeys]
  }
  middleKeys = insertMissingKeysNaturally(middleKeys, allMiddleKeys)
  middleKeys = applyAdjacentPairs(middleKeys, adjacentPairs || [])
  return middleKeys
}

/**
 * 列设置 composable（简化版）
 * 
 * 使用示例：
 * const { tableColumns, showColumnSetting, allColumns, visibleColumns, handleSaveColumnSetting } = useColumnSetting('page_name', allTableColumns)
 * 
 * @param {string} storageKey localStorage 存储键名（会自动加上 '_column_setting' 后缀）
 * @param {ComputedRef|Array} tableColumns 表格列配置数组
 * @param {Object} [options]
 * @param {Array<[string, string]>} [options.adjacentPairs] 初始化时强制列顺序，如 [['department','position']]；仅首次加载持久化时应用，之后可由用户在列设置中调整
 * @returns {Object} 返回列设置相关的状态和方法
 */
export function useColumnSetting(storageKey, tableColumns, options = {}) {
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
        // 排除 checkbox 类型列
        if (column.type === 'checkbox') return false
        
        const key = getColumnKey(column)
        
        // 排除没有 key 的列
        if (!key) return false
        
        // 排除 operation 列（通过 key、slot 或 title 判断）
        if (key === 'operation' || 
            column.slot === 'operation' || 
            column.field === 'operation' ||
            (column.title && (column.title.includes('操作') || column.title.includes('operation')) && !column.field && !column.slot)) {
          return false
        }
        
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

  // 从 localStorage 加载列设置
  const savedSettings = Storage.getItem(fullStorageKey) || {}
  
  // 可见列
  const visibleColumns = ref(
    Array.isArray(savedSettings.visibleColumns) ? [...savedSettings.visibleColumns] : []
  )
  
  // 列顺序
  const columnOrder = ref(
    Array.isArray(savedSettings.columnOrder) ? [...savedSettings.columnOrder] : []
  )
  
  // 冻结列配置
  const fixedColumns = ref(
    savedSettings.fixedColumns && typeof savedSettings.fixedColumns === 'object' 
      ? { ...savedSettings.fixedColumns } 
      : {}
  )
  
  // 如果没有存储值，使用默认值
  if (!visibleColumns.value.length && defaultVisibleColumns.value.length) {
    visibleColumns.value = [...defaultVisibleColumns.value]
  }
  
  // 初始化列顺序（如果没有保存的顺序）
  if (!columnOrder.value.length) {
    const columns = Array.isArray(tableColumns) ? tableColumns : (tableColumns?.value || [])
    columnOrder.value = columns.map(col => getColumnKey(col)).filter(Boolean)
  }

  // 一次性规范化：去掉无效 key、按定义顺序插入新列、可选相邻约束（修正旧缓存，如岗位被追加在末尾）
  const adjacentPairsOnInit = options.adjacentPairs || []
  {
    const columnsForInit = Array.isArray(tableColumns) ? tableColumns : (tableColumns?.value || [])
    const allMiddleKeysInit = columnsForInit
      .map((col) => getColumnKey(col))
      .filter((key) => key && key !== 'checkbox' && key !== 'operation')
    const rawOrder = [...columnOrder.value]
    const savedMiddle = rawOrder.filter((key) => key && key !== 'checkbox' && key !== 'operation')
    const prefix = rawOrder.filter((k) => k === 'checkbox')
    const suffix = rawOrder.filter((k) => k === 'operation')
    const normalizedMiddle = resolveMiddleKeys(savedMiddle, allMiddleKeysInit, adjacentPairsOnInit)
    const newFullOrder = [...prefix, ...normalizedMiddle, ...suffix]
    if (newFullOrder.join(',') !== rawOrder.join(',')) {
      columnOrder.value = newFullOrder
      Storage.setItem(fullStorageKey, {
        visibleColumns: visibleColumns.value,
        columnOrder: columnOrder.value,
        fixedColumns: fixedColumns.value
      })
    }
  }

  // 保存列设置
  const handleSaveColumnSetting = (settings) => {
    // 支持新格式（对象）和旧格式（数组）
    if (settings && typeof settings === 'object' && !Array.isArray(settings)) {
      // 新格式：包含 visibleColumns, fixedColumns, columnOrder
      if (Array.isArray(settings.visibleColumns)) {
        visibleColumns.value = [...settings.visibleColumns]
      }
      // 强制更新 columnOrder（创建新数组引用以触发响应式更新）
      if (Array.isArray(settings.columnOrder) && settings.columnOrder.length > 0) {
        // 始终创建新数组，确保 Vue 能检测到变化
        // 过滤掉 checkbox 和 operation，只保留可配置的列
        const filteredOrder = settings.columnOrder.filter(
          key => key && key !== 'checkbox' && key !== 'operation'
        )
        
        // 构建完整的顺序：checkbox + 可配置列 + operation
        const newOrder = []
        if (settings.columnOrder.includes('checkbox')) {
          newOrder.push('checkbox')
        }
        newOrder.push(...filteredOrder)
        if (settings.columnOrder.includes('operation')) {
          newOrder.push('operation')
        }
        
        columnOrder.value = newOrder.length > 0 ? [...newOrder] : [...settings.columnOrder]
      } else if (Array.isArray(settings.columnOrder) && settings.columnOrder.length === 0) {
        // 如果传入空数组，重置为默认顺序
        const cols = Array.isArray(tableColumns) ? tableColumns : (tableColumns?.value || [])
        const configurableKeys = cols
          .map(col => getColumnKey(col))
          .filter(key => key && key !== 'checkbox' && key !== 'operation')
        
        const newOrder = []
        if (cols.some(col => getColumnKey(col) === 'checkbox')) {
          newOrder.push('checkbox')
        }
        newOrder.push(...configurableKeys)
        if (cols.some(col => {
          const key = getColumnKey(col)
          return !key || key === 'operation'
        })) {
          newOrder.push('operation')
        }
        
        columnOrder.value = newOrder
      }
      if (settings.fixedColumns && typeof settings.fixedColumns === 'object') {
        fixedColumns.value = { ...settings.fixedColumns }
      }
      
      // 保存到 localStorage
      const saveData = {
        visibleColumns: visibleColumns.value,
        columnOrder: columnOrder.value,
        fixedColumns: fixedColumns.value
      }
      Storage.setItem(fullStorageKey, saveData)
    } else {
      // 旧格式：直接是 visibleColumns 数组
      visibleColumns.value = Array.isArray(settings) ? [...settings] : []
      Storage.setItem(fullStorageKey, {
        visibleColumns: visibleColumns.value,
        columnOrder: columnOrder.value,
        fixedColumns: fixedColumns.value
      })
    }
    
    showColumnSetting.value = false
    ElMessage.success(t('common.save_success'))
  }

  // 根据 visibleColumns、columnOrder 和 fixedColumns 过滤和排序显示的列
  const filteredColumns = computed(() => {
    // 直接使用 ref.value 以确保响应式追踪
    // 注意：这里不能先赋值给变量，必须直接使用 ref.value
    const currentColumnOrder = columnOrder.value
    const currentVisibleColumns = visibleColumns.value
    const currentFixedColumns = fixedColumns.value
    
    const columns = Array.isArray(tableColumns) ? tableColumns : (tableColumns?.value || [])
    
    // 创建列映射
    const columnMap = {}
    columns.forEach(col => {
      const key = getColumnKey(col)
      if (key) {
        columnMap[key] = col
      }
    })
    
    // 按照 columnOrder 排序中间列
    // columnOrder 只包含中间列的 key（不包括 checkbox 和 operation）
    // checkbox 固定在最前，operation 固定在最后
    const allMiddleKeys = columns
      .map(col => getColumnKey(col))
      .filter(key => key && key !== 'checkbox' && key !== 'operation')
    
    const savedMiddle = currentColumnOrder.filter(
      (key) => key && key !== 'checkbox' && key !== 'operation'
    )
    // 与初始化一致：去无效列、新列按定义顺序插入；不在此重复 adjacentPairs，避免覆盖用户在列设置中的调整
    const middleKeys = resolveMiddleKeys(savedMiddle, allMiddleKeys, [])
    
    // 构建最终的 orderedKeys：checkbox + 中间列 + operation
    // 注意：这个 orderedKeys 只用于调试，实际排序使用 middleKeys
    const orderedKeys = []
    if (columns.some(col => col.type === 'checkbox')) {
      orderedKeys.push('checkbox')
    }
    orderedKeys.push(...middleKeys)
    if (columns.some(col => {
      const key = getColumnKey(col)
      return !key || key === 'operation'
    })) {
      orderedKeys.push('operation')
    }
    
    // 过滤和排序列
    const result = []
    
    // 先添加 checkbox（如果存在）
    const checkboxCol = columns.find(col => col.type === 'checkbox')
    if (checkboxCol) {
      result.push(checkboxCol)
    }
    
    // 按顺序添加可见列（使用 middleKeys，排除 checkbox 和 operation）
    middleKeys.forEach(key => {
      if (!key) return
      
      const column = columnMap[key]
      if (column) {
        const colKey = getColumnKey(column)
        // 只添加可见的列（required 列始终显示，避免旧本地缓存缺少新列）
        if (currentVisibleColumns.includes(colKey) || column.required) {
          // 应用冻结设置
          const fixed = currentFixedColumns[colKey]
          if (fixed) {
            result.push({ ...column, fixed })
          } else {
            // 保留原有的 fixed 属性（如果存在）
            result.push(column)
          }
        }
      }
    })
    
    // 最后添加 operation 列（操作列始终显示，不受列设置控制）
    const operationCol = columns.find(col => {
      const key = getColumnKey(col)
      return !key || key === 'operation'
    })
    if (operationCol) {
      const opKey = getColumnKey(operationCol)
      if (!opKey || opKey === 'operation') {
        const fixed = currentFixedColumns['operation']
        if (fixed) {
          result.push({ ...operationCol, fixed })
        } else {
          result.push(operationCol)
        }
      }
    }
    
    return result
  })

  // 处理列设置确认（统一处理逻辑，避免每个页面重复）
  const handleColumnSettingConfirm = (result) => {
    if (result && typeof result === 'object' && !Array.isArray(result)) {
      // 新格式：包含 visibleColumns, fixedColumns, columnOrder
      handleSaveColumnSetting(result)
    } else {
      // 兼容旧格式：直接是 visibleColumns 数组
      handleSaveColumnSetting(result)
    }
  }

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
    // 列顺序
    columnOrder,
    // 冻结列配置
    fixedColumns,
    // 保存列设置
    handleSaveColumnSetting,
    // 处理列设置确认（统一方法）
    handleColumnSettingConfirm
  }
}
