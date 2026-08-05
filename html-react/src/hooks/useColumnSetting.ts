import { useCallback, useMemo, useState } from 'react'
import type { ColumnsType } from 'antd/es/table'
import { App } from 'antd'
import { useTranslation } from 'react-i18next'
import Storage from '@/utils/storage'

export interface ColumnSettingItem {
  key: string
  title: string
  required?: boolean
}

export interface ColumnSettingValue {
  visibleColumns: string[]
  columnOrder: string[]
  fixedColumns: Record<string, 'left' | 'right'>
}

interface RawColumn {
  key?: string | number
  dataIndex?: string | number | readonly (string | number)[]
  title?: unknown
  fixed?: 'left' | 'right' | boolean
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
  const [saved, setSaved] = useState<ColumnSettingValue | null>(() =>
    Storage.getItem<ColumnSettingValue>(fullStorageKey, null),
  )

  const visibleColumns = useMemo(() => {
    const list = (saved?.visibleColumns || []).filter((key) => allKeys.includes(key))
    const base = list.length ? list : allKeys
    return Array.from(new Set([...base, ...requiredKeys.filter((key) => allKeys.includes(key))]))
  }, [saved, allKeys, requiredKeys])

  const columnOrder = useMemo(() => {
    const list = (saved?.columnOrder || []).filter((key) => allKeys.includes(key))
    return insertMissingNaturally(list.length ? list : allKeys, allKeys)
  }, [saved, allKeys])

  const fixedColumns = useMemo<Record<string, 'left' | 'right'>>(
    () => (saved?.fixedColumns && typeof saved.fixedColumns === 'object' ? saved.fixedColumns : {}),
    [saved],
  )

  const filteredColumns = useMemo<ColumnsType<T>>(() => {
    const columnMap = new Map(configurableColumns.map((col) => [getColumnKey(col), col]))
    const result: RawColumn[] = []
    columnOrder.forEach((key) => {
      const col = columnMap.get(key)
      if (!col) return
      if (!visibleColumns.includes(key) && !requiredKeys.includes(key)) return
      const fixed = fixedColumns[key]
      result.push(fixed ? { ...col, fixed } : col)
    })
    if (operationColumn) {
      const fixed = fixedColumns.operation
      result.push(fixed ? { ...operationColumn, fixed } : operationColumn)
    }
    return result as unknown as ColumnsType<T>
  }, [configurableColumns, columnOrder, visibleColumns, fixedColumns, operationColumn, requiredKeys])

  const handleConfirm = useCallback(
    (value: ColumnSettingValue) => {
      setSaved(value)
      Storage.setItem(fullStorageKey, value)
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
