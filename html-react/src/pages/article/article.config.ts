import { entityField } from '@/utils/normalize'

export const articleInitialSearchForm = {
  admin_id: '',
  title: '',
  content: '',
  status: '',
}

export interface ArticleRow {
  id: number | string
  admin_id?: number | string
  admin?: { nickname?: string; username?: string } | null
  title?: string
  content?: string
  status?: number
  created_at?: string
  updated_at?: string
}

export function transformArticleRow(row: Record<string, unknown>): ArticleRow {
  return {
    id: entityField(row, 'id', '')!,
    admin_id: entityField(row, 'admin_id', '') as number | string,
    admin: (entityField(row, 'admin', null) as ArticleRow['admin']) || null,
    title: String(entityField(row, 'title', '') ?? ''),
    content: String(entityField(row, 'content', '') ?? ''),
    status: Number(entityField(row, 'status', 0) ?? 0),
    created_at: String(entityField(row, 'created_at', '') ?? ''),
    updated_at: String(entityField(row, 'updated_at', '') ?? ''),
  }
}

export function getAdminDisplayName(row: ArticleRow) {
  if (row.admin) return row.admin.nickname || row.admin.username || '-'
  return row.admin_id ? String(row.admin_id) : '-'
}
