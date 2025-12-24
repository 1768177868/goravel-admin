import { computed, ref } from 'vue'

/**
 * 表格性能优化 composable
 * 提供虚拟滚动、列渲染优化等功能
 * 
 * @param {Object} options 配置选项
 * @param {Number} options.enableVirtualScroll - 启用虚拟滚动的数据量阈值（默认 100）
 * @param {Number} options.virtualScrollItemHeight - 虚拟滚动每行高度（默认 46）
 * @param {Number} options.tableHeight - 表格高度（默认 600）
 * @param {Array} options.tableColumns - 表格列配置
 * @param {Array} options.tableData - 表格数据
 * @returns {Object} 性能优化相关的配置和方法
 */
export function useTablePerformance(options = {}) {
  const {
    enableVirtualScroll = 100, // 超过 100 条数据时启用虚拟滚动
    virtualScrollItemHeight = 46, // 每行高度（px）
    tableHeight = 600,
    tableColumns = [],
    tableData = []
  } = options

  // 计算是否启用虚拟滚动
  const shouldEnableVirtualScroll = computed(() => {
    return tableData.length >= enableVirtualScroll
  })

  // 虚拟滚动配置
  const scrollYConfig = computed(() => {
    if (shouldEnableVirtualScroll.value) {
      return {
        enabled: true,
        gt: enableVirtualScroll, // 超过这个数量才启用虚拟滚动
        oSize: virtualScrollItemHeight, // 每行高度
        rSize: virtualScrollItemHeight, // 渲染行高度
        mode: 'wheel' // 滚动模式：wheel（鼠标滚轮）或 scroll（滚动条）
      }
    }
    return { enabled: false }
  })

  // 优化列渲染：使用 computed 缓存列配置
  const optimizedColumns = computed(() => {
    return tableColumns.map(column => {
      // 如果 formatter 是函数，确保它被正确缓存
      if (column.formatter && typeof column.formatter === 'function') {
        // 保持 formatter 函数引用，避免重复创建
        return {
          ...column,
          formatter: column.formatter
        }
      }
      return column
    })
  })

  // 表格配置优化
  const optimizedTableConfig = computed(() => {
    return {
      resizable: true,
      // 优化渲染性能
      showOverflow: 'tooltip', // 超出内容显示 tooltip，而不是换行
      showHeaderOverflow: 'tooltip'
    }
  })

  return {
    // 配置
    scrollYConfig,
    optimizedColumns,
    optimizedTableConfig,
    shouldEnableVirtualScroll,
    
    // 方法
    enableVirtualScroll,
    virtualScrollItemHeight
  }
}

