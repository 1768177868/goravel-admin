import { entityField } from '@/utils/normalize'

export const userInitialSearchForm = {
  username: '',
  status: '',
}

export type UserSearchForm = typeof userInitialSearchForm

export interface UserCurrency {
  symbol?: string
  decimal_places?: number
  code?: string
  name?: string
}

export interface UserRow {
  id: number | string
  username: string
  nickname?: string
  email?: string
  phone?: string
  status?: number
  balance?: number | string
  currency?: UserCurrency | null
  created_at?: string
  name?: string
}

export function transformUserRow(row: Record<string, unknown>): UserRow {
  const currencyRaw = entityField(row, 'currency', null)
  const currency =
    currencyRaw && typeof currencyRaw === 'object' ? (currencyRaw as UserCurrency) : null

  return {
    id: entityField(row, 'id', '')!,
    username: String(entityField(row, 'username', '') ?? ''),
    nickname: String(entityField(row, 'nickname', '') ?? ''),
    email: String(entityField(row, 'email', '') ?? ''),
    phone: String(entityField(row, 'phone', '') ?? ''),
    status: Number(entityField(row, 'status', 0) ?? 0),
    balance: (entityField(row, 'balance', 0) as number | string) ?? 0,
    currency,
    created_at: String(entityField(row, 'created_at', '') ?? ''),
    name: String(entityField(row, 'username', '') ?? ''),
  }
}

export function formatUserBalance(
  balance: number | string | undefined,
  currency?: UserCurrency | null,
) {
  const symbol = currency?.symbol || '¥'
  const decimalPlaces = currency?.decimal_places ?? 2
  return `${symbol}${Number(balance || 0).toFixed(decimalPlaces)}`
}
