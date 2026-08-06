import type { SearchField } from '@/components/SearchForm'
import type { TFunction } from 'i18next'

export interface LoginLogRow {
  id: number | string
  admin?: { username?: string } | null
  ip?: string
  user_agent?: string
  location?: string
  status?: number
  message?: string
  request?: string
  created_at?: string
  name?: string
}

export const loginLogInitialSearchForm = {
  username: '',
  ip: '',
  status: '',
  start_time: '',
  end_time: '',
}

export function transformLoginLogRow(log: Record<string, unknown>): LoginLogRow {
  const adminRaw = log.admin as Record<string, unknown> | undefined
  return {
    id: log.id as number | string,
    admin: adminRaw ? { username: String(adminRaw.username || '') } : null,
    ip: String(log.ip || ''),
    user_agent: String(log.user_agent || ''),
    location: String(log.location || ''),
    status: Number(log.status || 0),
    message: String(log.message || ''),
    request: String(log.request || ''),
    created_at: String(log.created_at || ''),
    name: String(adminRaw?.username || log.username || ''),
  }
}

export function createLoginLogSearchFields(t: TFunction): SearchField[] {
  return [
    { name: 'username', label: t('log.username'), span: 6 },
    { name: 'ip', label: t('log.ip'), span: 6 },
    {
      name: 'status',
      label: t('log.status'),
      type: 'select',
      span: 6,
      options: [
        { label: t('log.success'), value: '1' },
        { label: t('log.failed'), value: '0' },
      ],
    },
    { name: 'start_time', label: t('log.start_time'), type: 'datetime', span: 6, advanced: true },
    { name: 'end_time', label: t('log.end_time'), type: 'datetime', span: 6, advanced: true },
  ]
}

export function formatRequestPreview(request?: string): string {
  if (!request) return '-'
  try {
    return JSON.stringify(JSON.parse(request), null, 2)
  } catch {
    return request
  }
}

export function translateLoginMessage(t: TFunction, messageKey?: string): string {
  if (!messageKey) return '-'
  const translation = t(`log.${messageKey}`, { defaultValue: messageKey })
  return translation !== `log.${messageKey}` ? translation : messageKey
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
