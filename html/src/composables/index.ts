/**
 * Composables 导出入口
 *
 * @example
 * import { useCrud, useListPage, useTreeListPage, usePermission } from '@/composables'
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

// JavaScript composables
export { useListPage } from './useListPage'
export { useStandardListPage } from './useStandardListPage'
export { useTreeListPage } from './useTreeListPage'
export { useStandardTreeListPage } from './useStandardTreeListPage'
export { useTreeExpand } from './useTreeExpand'
export { useQueuedExport } from './useQueuedExport'
export { useCsvImport } from './useCsvImport'
export { useAttachmentImagePreview } from './useAttachmentImagePreview'
export { useAttachmentChunkUpload, useAttachmentUploadConfig } from './useAttachmentChunkUpload'
export { useTableData } from './useTableData'
export { useColumnSetting } from './useColumnSetting'
