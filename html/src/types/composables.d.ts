/**
 * Composables 类型定义
 * 为 JavaScript composables 提供类型提示
 */

import type { Ref, ComputedRef } from 'vue'
import type { Pagination, TableColumn, SearchField } from './index'

// ==================== useCrud ====================

export interface UseCrudOptions {
  /** 单条删除 API 函数 */
  deleteApi?: (id: number | string) => Promise<any>
  /** 批量删除 API 函数（接收 ids 数组） */
  batchDeleteApi?: (ids: (number | string)[]) => Promise<any>
  /** 删除确认提示的 i18n key */
  deleteConfirmKey?: string
  /** 删除成功提示的 i18n key */
  deleteSuccessKey?: string
  /** 批量删除确认提示的 i18n key */
  batchDeleteConfirmKey?: string
  /** 提示框标题的 i18n key */
  tipKey?: string
  /** 删除成功回调 */
  onDeleteSuccess?: (data: any, id?: number | string) => void
  /** 删除失败回调 */
  onDeleteError?: (error: any, data: any) => void
  /** 删除前钩子 */
  beforeDelete?: (data: any) => boolean | Promise<boolean>
  /** 删除后钩子 */
  afterDelete?: (data: any) => void
}

export interface UseCrudReturn {
  /** 对话框可见状态 */
  dialogVisible: Ref<boolean>
  /** 编辑 ID */
  editId: Ref<number | string | null>
  /** 打开添加对话框 */
  handleAdd: () => void
  /** 打开编辑对话框 */
  handleEdit: (row: any) => void
  /** 关闭对话框 */
  handleClose: () => void
  /** 表单提交成功处理 */
  handleFormSuccess: (reloadData?: () => void) => void
  /** 删除单条记录 */
  handleDelete: (row: any, reloadData?: () => void) => Promise<void>
  /** 批量删除记录 */
  handleBatchDelete: (rows: any[], reloadData?: () => void) => Promise<void>
}

export function useCrud(options?: UseCrudOptions): UseCrudReturn

// ==================== useListPage ====================

export interface UseListPageOptions<T = any> {
  /** 获取列表数据的 API 函数 */
  fetchApi: (params: any) => Promise<any>
  /** 初始搜索表单值 */
  initialSearchForm?: Record<string, any>
  /** 排序配置 */
  sortOptions?: {
    tableRef: Ref<any>
    fieldMapping?: Record<string, string>
    defaultSort?: string
  }
  /** 数据转换函数 */
  transformData?: (item: any) => T
  /** 加载成功回调 */
  onLoadSuccess?: (response: any, list: T[]) => void
  /** 加载失败回调 */
  onLoadError?: (error: any) => void
}

export interface UseListPageReturn<T = any> {
  /** 分页状态 */
  pagination: Pagination
  /** 表格数据 */
  tableData: Ref<T[]>
  /** 加载状态 */
  loading: Ref<boolean>
  /** 搜索表单 */
  searchForm: Record<string, any>
  /** 加载数据 */
  loadData: () => Promise<void>
  /** 搜索处理 */
  handleSearch: () => void
  /** 重置处理 */
  handleReset: () => void
  /** 分页变化处理 */
  handlePageChange: (params: { currentPage: number; pageSize: number }) => void
  /** 排序变化处理 */
  handleSortChange: (params: any) => void
  /** 初始化默认排序 */
  initDefaultSort: () => void
  /** 取消请求 */
  cancelRequest: () => void
}

export function useListPage<T = any>(options: UseListPageOptions<T>): UseListPageReturn<T>

// ==================== useTableSort ====================

export interface UseTableSortOptions {
  /** 表格引用 */
  tableRef: Ref<any>
  /** 字段映射 */
  fieldMapping?: Record<string, string>
  /** 默认排序 */
  defaultSort?: string
  /** 排序变化回调 */
  onSortChange?: () => void
}

export interface UseTableSortReturn {
  /** 构建排序参数 */
  buildOrderBy: () => string
  /** 排序变化处理 */
  handleSortChange: (params: any) => void
  /** 重置排序 */
  resetSort: () => void
  /** 初始化默认排序 */
  initDefaultSort: () => void
}

export function useTableSort(options: UseTableSortOptions): UseTableSortReturn

// ==================== useColumnSetting ====================

export interface UseColumnSettingOptions {
  /** 默认显示的列 keys */
  defaultVisibleColumns?: string[]
  /** 始终显示的列 keys */
  alwaysVisibleKeys?: string[]
  /** 排除的字段 */
  excludeFields?: string[]
  /** i18n 前缀 */
  i18nPrefix?: string
  /** 自定义获取列标题函数 */
  getColumnTitle?: (column: TableColumn) => string
}

export interface UseColumnSettingReturn {
  /** 当前可见列 */
  visibleColumns: Ref<string[]>
  /** 所有可配置的列 */
  allColumns: ComputedRef<Array<{ key: string; title: string }>>
  /** 默认可见列 */
  defaultVisibleColumns: string[]
  /** 过滤后的表格列 */
  tableColumns: ComputedRef<TableColumn[]>
  /** 更新可见列 */
  updateVisibleColumns: (columns: string[]) => void
  /** 重置为默认 */
  resetToDefault: () => void
}

export function useColumnSetting(
  storageKey: string,
  allTableColumns: ComputedRef<TableColumn[]>,
  options?: UseColumnSettingOptions
): UseColumnSettingReturn

// ==================== usePermission ====================

export interface ButtonState {
  disabled: boolean
  hidden: boolean
}

export interface UsePermissionReturn {
  /** 获取按钮状态 */
  getButtonState: (permissionSlug: string) => ButtonState
  /** 检查是否有权限 */
  hasPermission: (permissionSlug: string) => boolean
  /** 检查是否有任一权限 */
  hasAnyPermission: (permissionSlugs: string[]) => boolean
  /** 检查是否有所有权限 */
  hasAllPermissions: (permissionSlugs: string[]) => boolean
}

export function usePermission(): UsePermissionReturn

// ==================== useDebounce ====================

export interface UseDebounceReturn<T extends (...args: any[]) => any> {
  /** 防抖后的函数 */
  debouncedFn: T
  /** 取消防抖 */
  cancel: () => void
}

export function useDebounce<T extends (...args: any[]) => any>(
  fn: T,
  delay?: number
): UseDebounceReturn<T>
