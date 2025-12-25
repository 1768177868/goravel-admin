/**
 * Composables 导出入口
 *
 * TypeScript 版本的 composables 可以直接从这里导入：
 * import { useCrud, useDebounce, useTableSort, usePermission } from '@/composables'
 *
 * 注意：部分 composables 仍为 JavaScript 版本，可以直接从各自文件导入
 */

// TypeScript composables
export { useCrud, type UseCrudOptions, type UseCrudReturn } from './useCrud'
export {
  useDebounce,
  useDebouncedRef,
  type DebouncedFunction
} from './useDebounce'
export {
  useTableSort,
  type SortItem,
  type SortConfig,
  type UseTableSortOptions,
  type UseTableSortReturn
} from './useTableSort'
export {
  usePermission,
  type ButtonState,
  type UsePermissionReturn
} from './usePermission'

// JavaScript composables（仍可使用，但无完整类型支持）
// export { useListPage } from './useListPage'
// export { useColumnSetting } from './useColumnSetting'
// export { useTablePerformance } from './useTablePerformance'
// export { useApiRequest } from './useApiRequest'

