import type { SearchField } from '@/components/SearchForm'
import type { TFunction } from 'i18next'

export interface SystemLogRow {
  id: number | string
  level?: string
  module?: string
  trace_id?: string
  message?: string
  context?: unknown
  created_at?: string
  name?: string
}

export const systemLogInitialSearchForm = {
  level: '',
  module: '',
  trace_id: '',
  message: '',
  start_time: '',
  end_time: '',
}

export function transformSystemLogRow(log: Record<string, unknown>): SystemLogRow {
  let context: unknown = null
  try {
    if (log.context) {
      context = typeof log.context === 'string' ? JSON.parse(log.context) : log.context
    }
  } catch {
    context = log.context || null
  }

  return {
    id: (log.id as number | string) || 0,
    level: String(log.level || ''),
    module: String(log.module || ''),
    trace_id: String(log.trace_id || ''),
    message: String(log.message || ''),
    context,
    created_at: String(log.created_at || ''),
    name: String(log.message || ''),
  }
}

export function createSystemLogSearchFields(
  t: TFunction,
  moduleOptions: Array<{ label: string; value: string }>,
): SearchField[] {
  return [
    {
      name: 'level',
      label: t('log.level'),
      type: 'select',
      span: 6,
      options: [
        { label: t('log.level_error'), value: 'error' },
        { label: t('log.level_warning'), value: 'warning' },
        { label: t('log.level_info'), value: 'info' },
        { label: t('log.level_debug'), value: 'debug' },
      ],
    },
    {
      name: 'module',
      label: t('log.module'),
      type: 'select',
      span: 6,
      options: moduleOptions,
    },
    { name: 'trace_id', label: t('log.trace_id'), span: 6 },
    { name: 'message', label: t('log.message'), span: 6, advanced: true },
    { name: 'start_time', label: t('log.start_time'), type: 'datetime', span: 6, advanced: true },
    { name: 'end_time', label: t('log.end_time'), type: 'datetime', span: 6, advanced: true },
  ]
}

export function getSystemLogLevelColor(level?: string): string {
  const levelMap: Record<string, string> = {
    error: 'error',
    warning: 'warning',
    info: 'success',
    debug: 'processing',
  }
  return levelMap[level?.toLowerCase() || ''] || 'default'
}

export function getSystemLogLevelLabel(t: TFunction, level?: string): string {
  if (!level) return '-'
  const levelMap: Record<string, string> = {
    error: t('log.level_error'),
    warning: t('log.level_warning'),
    info: t('log.level_info'),
    debug: t('log.level_debug'),
  }
  return levelMap[level.toLowerCase()] || level
}

export function getSystemLogModuleLabel(t: TFunction, module?: string): string {
  if (!module) return '-'
  const normalized = String(module).trim()
  if (!normalized) return '-'

  const compatibilityMap: Record<string, string> = {
    'operation-log': 'operation_log',
    'login-log': 'login_log',
    'system-log': 'system_log',
    'online-admin': 'online_admin',
    'background-task': 'module_background_task',
  }
  const mapped = compatibilityMap[normalized] || normalized
  const snake = mapped.replace(/-/g, '_')

  const menuKey = `menu.${snake}`
  const menuLabel = t(menuKey, { defaultValue: '' })
  if (menuLabel && menuLabel !== menuKey) return menuLabel

  const logKey = snake.startsWith('module_') ? `log.${snake}` : `log.module_${snake}`
  const logLabel = t(logKey, { defaultValue: '' })
  if (logLabel && logLabel !== logKey) return logLabel

  return normalized
}

export function formatSystemLogContext(context: unknown): string {
  if (!context) return '-'
  try {
    if (typeof context === 'string') {
      return JSON.stringify(JSON.parse(context), null, 2)
    }
    return JSON.stringify(context, null, 2)
  } catch {
    return String(context)
  }
}

export function formatSystemLogContextPreview(context: unknown): string {
  if (!context) return '-'
  try {
    let obj: unknown = context
    if (typeof context === 'string') {
      obj = JSON.parse(context)
    }
    if (typeof obj === 'object' && obj !== null) {
      const keys = Object.keys(obj as object)
      if (keys.length === 0) return '{}'
      const preview = keys
        .slice(0, 2)
        .map((key) => {
          const value = (obj as Record<string, unknown>)[key]
          const valueStr = typeof value === 'object' ? JSON.stringify(value) : String(value)
          return `${key}: ${valueStr.length > 20 ? `${valueStr.substring(0, 20)}...` : valueStr}`
        })
        .join(', ')
      return keys.length > 2 ? `${preview}...` : preview
    }
    return String(obj)
  } catch {
    return String(context)
  }
}

export function extractLogDetail<T>(data: Record<string, unknown> | undefined | null, keys: string[]): T | null {
  if (!data) return null
  for (const key of keys) {
    const value = data[key]
    if (value && typeof value === 'object') {
      return value as T
    }
  }
  return data as T
}
