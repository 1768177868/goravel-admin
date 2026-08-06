import { entityField } from '@/utils/normalize'

export const userInitialSearchForm = {
  username: '',
  status: '',
}

export type UserSearchForm = typeof userInitialSearchForm

export interface UserRow {
  id: number | string
  username: string
  nickname?: string
  email?: string
  phone?: string
  status?: number
  balance?: number | string
  currency?: string
  created_at?: string
  name?: string
}

export function transformUserRow(row: Record<string, unknown>): UserRow {
  return {
    id: entityField(row, 'id', '')!,
    username: String(entityField(row, 'username', '') ?? ''),
    nickname: String(entityField(row, 'nickname', '') ?? ''),
    email: String(entityField(row, 'email', '') ?? ''),
    phone: String(entityField(row, 'phone', '') ?? ''),
    status: Number(entityField(row, 'status', 0) ?? 0),
    balance: (entityField(row, 'balance', 0) as number | string) ?? 0,
    currency: String(entityField(row, 'currency', '') ?? ''),
    created_at: String(entityField(row, 'created_at', '') ?? ''),
    name: String(entityField(row, 'username', '') ?? ''),
  }
}

export function formatUserBalance(balance: number | string | undefined, currency?: string) {
  const amount = Number(balance || 0).toFixed(2)
  return currency ? `${amount} ${currency}` : amount
}
