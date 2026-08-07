import { useCallback, useMemo, useState } from 'react'
import type { ColumnsType } from 'antd/es/table'
import { App } from 'antd'
import { useTranslation } from 'react-i18next'
import Storage from '@/utils/storage'

export type ColumnFixedSide = 'start' | 'end'

export interface ColumnSettingItem {
  key: string
  title: string
  required?: boolean
}

export interface ColumnSettingValue {
  visibleColumns: string[]
  columnOrder: string[]
  fixedColumns: Record<string, ColumnFixedSide>
}

interface RawColumn {
  key?: string | number
  dataIndex?: string | number | readonly (string | number)[]
  title?: unknown
  fixed?: ColumnFixedSide | 'left' | 'right' | boolean
}

function getColumnKey(column: RawColumn): string {
  if (column.key !== undefined && column.key !== null && column.key !== '') return String(column.key)
  const { dataIndex } = column
  if (Array.isArray(dataIndex)) return dataIndex.join('.')
  if (dataIndex !== undefined && dataIndex !== null) return String(dataIndex)
  return ''
}

function isConfigurable(column: RawColumn): boolean {
  const key = getColumnKey(column)
  return !!key && key !== 'operation'
}

/** Ant Design 6 only recognizes start/end; left/right are legacy no-ops. */
export function normalizeColumnFixed(
  fixed?: ColumnFixedSide | 'left' | 'right' | boolean | null,
): ColumnFixedSide | boolean | undefined {
  if (fixed === 'left' || fixed === 'start') return 'start'
  if (fixed === 'right' || fixed === 'end') return 'end'
  if (fixed === true || fixed === false) return fixed
  return undefined
}

function normalizeFixedColumns(
  input: Record<string, string> | null | undefined,
): Record<string, ColumnFixedSide> {
  if (!input || typeof input !== 'object') return {}
  const out: Record<string, ColumnFixedSide> = {}
  Object.entries(input).forEach(([key, value]) => {
    const normalized = normalizeColumnFixed(value as ColumnFixedSide | 'left' | 'right')
    if (normalized === 'start' || normalized === 'end') out[key] = normalized
  })
  return out
}

/** Insert keys missing from `current` at the position they occupy in `reference`, instead of appending to the end. */
function insertMissingNaturally(current: string[], reference: string[]): string[] {
  const missing = reference.filter((key) => !current.includes(key))
  if (!missing.length) return current
  const result = [...current]
  missing.forEach((key) => {
    const targetIdx = reference.indexOf(key)
    let insertAt = result.length
    for (let i = targetIdx - 1; i >= 0; i--) {
      const idx = result.indexOf(reference[i])
      if (idx !== -1) {
        insertAt = idx + 1
        break
      }
    }
    result.splice(insertAt, 0, key)
  })
  return result
}

/**
 * Column visibility/order/fixed-side settings for an antd Table, persisted to localStorage.
 * Port of html/src/composables/useColumnSetting.js adapted for antd ColumnsType.
 */
export function useColumnSetting<T extends object>(
  storageKey: string,
  columns: ColumnsType<T>,
  options: { requiredKeys?: string[] } = {},
) {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const requiredKeys = options.requiredKeys ?? []
  const fullStorageKey = storageKey.includes('_column_setting') ? storageKey : `${storageKey}_column_setting`

  const rawColumns = columns as unknown as RawColumn[]
  const configurableColumns = useMemo(() => rawColumns.filter(isConfigurable), [rawColumns])
  const operationColumn = useMemo(() => rawColumns.find((col) => !isConfigurable(col)), [rawColumns])

  const allColumns = useMemo<ColumnSettingItem[]>(
    () =>
      configurableColumns.map((col) => ({
        key: getColumnKey(col),
        title: typeof col.title === 'string' ? col.title : getColumnKey(col),
        required: requiredKeys.includes(getColumnKey(col)),
      })),
    [configurableColumns, requiredKeys],
  )
  const allKeys = useMemo(() => allColumns.map((c) => c.key), [allColumns])

  const [open, setOpen] = useState(false)
  const [saved, setSaved] = useState<ColumnSettingValue | null>(() => {
    const raw = Storage.getItem<ColumnSettingValue>(fullStorageKey, null)
    if (!raw) return null
    return {
      ...raw,
      fixedColumns: normalizeFixedColumns(raw.fixedColumns as Record<string, string>),
    }
  })

  const visibleColumns = useMemo(() => {
    const list = (saved?.visibleColumns || []).filter((key) => allKeys.includes(key))
    const base = list.length ? list : allKeys
    return Array.from(new Set([...base, ...requiredKeys.filter((key) => allKeys.includes(key))]))
  }, [saved, allKeys, requiredKeys])

  const columnOrder = useMemo(() => {
    const list = (saved?.columnOrder || []).filter((key) => allKeys.includes(key))
    return insertMissingNaturally(list.length ? list : allKeys, allKeys)
  }, [saved, allKeys])

  const fixedColumns = useMemo<Record<string, ColumnFixedSide>>(
    () => normalizeFixedColumns(saved?.fixedColumns as Record<string, string>),
    [saved],
  )

  const filteredColumns = useMemo<ColumnsType<T>>(() => {
    const columnMap = new Map(configurableColumns.map((col) => [getColumnKey(col), col]))
    const result: RawColumn[] = []
    columnOrder.forEach((key) => {
      const col = columnMap.get(key)
      if (!col) return
      if (!visibleColumns.includes(key) && !requiredKeys.includes(key)) return
      const fixed = normalizeColumnFixed(fixedColumns[key] ?? col.fixed)
      result.push(fixed !== undefined ? { ...col, fixed } : { ...col, fixed: undefined })
    })
    if (operationColumn) {
      const fixed = normalizeColumnFixed(fixedColumns.operation ?? operationColumn.fixed) || 'end'
      result.push({ ...operationColumn, fixed })
    }
    return result as unknown as ColumnsType<T>
  }, [configurableColumns, columnOrder, visibleColumns, fixedColumns, operationColumn, requiredKeys])

  const handleConfirm = useCallback(
    (value: ColumnSettingValue) => {
      const next: ColumnSettingValue = {
        ...value,
        fixedColumns: normalizeFixedColumns(value.fixedColumns as Record<string, string>),
      }
      setSaved(next)
      Storage.setItem(fullStorageKey, next)
      setOpen(false)
      message.success(t('common.save_success'))
    },
    [fullStorageKey, message, t],
  )

  const openColumnSetting = useCallback(() => setOpen(true), [])
  const closeColumnSetting = useCallback(() => setOpen(false), [])

  return {
    filteredColumns,
    open,
    openColumnSetting,
    closeColumnSetting,
    allColumns,
    visibleColumns,
    columnOrder,
    fixedColumns,
    handleConfirm,
  }
}
