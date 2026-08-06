import { entityField } from '@/utils/normalize'

export const roleInitialSearchForm = {
  name: '',
  status: '',
}

export type RoleSearchForm = typeof roleInitialSearchForm

export interface RoleRow {
  id: number | string
  name?: string
  slug?: string
  description?: string
  status?: number
  sort?: number
  created_at?: string
}

export function transformRoleRow(row: Record<string, unknown>): RoleRow {
  return {
    id: entityField(row, 'id', '')!,
    name: String(entityField(row, 'name', '') ?? ''),
    slug: String(entityField(row, 'slug', '') ?? ''),
    description: String(entityField(row, 'description', '') ?? ''),
    status: Number(entityField(row, 'status', 0) ?? 0),
    sort: Number(entityField(row, 'sort', 0) ?? 0),
    created_at: String(entityField(row, 'created_at', '') ?? ''),
  }
}
