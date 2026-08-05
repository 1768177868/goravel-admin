import { entityField } from '@/utils/normalize'

export const ATTACHMENT_CHUNK_SIZE = 1 * 1024 * 1024
export const ATTACHMENT_LARGE_FILE_THRESHOLD = 5 * 1024 * 1024

export const attachmentInitialSearchForm = {
  filename: '',
  display_name: '',
  category_id: '',
  is_public: '',
  file_type: '',
  extension: '',
  start_time: '',
  end_time: '',
}

export interface AttachmentRow {
  id: number | string
  admin?: Record<string, unknown> | null
  category_id?: number | string | null
  category_name?: string
  filename?: string
  display_name?: string
  file_type?: string
  disk?: string
  extension?: string
  size?: number
  mime_type?: string
  is_public?: number
  created_at?: string
  file_url?: string
  name?: string
}

export function transformAttachmentRow(item: Record<string, unknown>): AttachmentRow {
  const category = (item.category || item.Category || null) as Record<string, unknown> | null
  return {
    id: entityField(item, 'id', '')!,
    admin: (item.admin || item.Admin || null) as Record<string, unknown> | null,
    category_id: entityField(item, 'category_id', entityField(category, 'id', null)) as
      | number
      | string
      | null,
    category_name: String(entityField(category, 'name', '') ?? ''),
    filename: String(entityField(item, 'filename', '') ?? ''),
    display_name: String(entityField(item, 'display_name', '') ?? ''),
    file_type: String(entityField(item, 'file_type', 'other') ?? 'other'),
    disk: String(entityField(item, 'disk', '') ?? ''),
    extension: String(entityField(item, 'extension', '') ?? ''),
    size: Number(entityField(item, 'size', 0) ?? 0),
    mime_type: String(entityField(item, 'mime_type', '') ?? ''),
    is_public: Number(entityField(item, 'is_public', 1) ?? 1),
    created_at: String(entityField(item, 'created_at', '') ?? ''),
    file_url: String(entityField(item, 'file_url', '') ?? ''),
    name: String(entityField(item, 'filename', '') ?? ''),
  }
}

export function formatAttachmentFileSize(size?: number) {
  if (!size) return '-'
  const numSize = Number(size)
  if (numSize < 1024) return `${numSize} B`
  if (numSize < 1024 * 1024) return `${(numSize / 1024).toFixed(2)} KB`
  if (numSize < 1024 * 1024 * 1024) return `${(numSize / 1024 / 1024).toFixed(2)} MB`
  return `${(numSize / 1024 / 1024 / 1024).toFixed(2)} GB`
}

export function getAttachmentFileTypeColor(fileType?: string) {
  const colors: Record<string, string> = {
    image: 'success',
    video: 'warning',
    document: 'processing',
    other: 'default',
  }
  return colors[fileType || 'other'] || 'default'
}

export function getAttachmentFileTypeLabel(t: (key: string) => string, fileType?: string) {
  const labels: Record<string, string> = {
    image: t('attachment.file_type_image'),
    video: t('attachment.file_type_video'),
    document: t('attachment.file_type_document'),
    other: t('attachment.file_type_other'),
  }
  return labels[fileType || ''] || fileType || '-'
}

export function formatAttachmentAdmin(row: AttachmentRow) {
  const admin = row.admin
  if (!admin) return '-'
  return String(entityField(admin, 'username', '-') ?? '-')
}
