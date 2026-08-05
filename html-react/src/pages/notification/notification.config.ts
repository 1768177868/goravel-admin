import { buildSearchParams } from '@/utils/buildSearchParams'
import { entityField } from '@/utils/normalize'

export const notificationInitialSearchForm = {
  type: '',
  is_read: '',
}

export interface NotificationRow {
  id: number | string
  type?: string
  title?: string
  content?: string
  sender?: { nickname?: string; username?: string } | null
  sender_id?: number | string | null
  receiver?: { nickname?: string; username?: string } | null
  receiver_id?: number | string | null
  is_read?: boolean
  read_at?: string | null
  created_at?: string
  name?: string
}

export function transformNotificationRow(row: Record<string, unknown>): NotificationRow {
  return {
    id: entityField(row, 'id', '')!,
    type: String(entityField(row, 'type', '') ?? ''),
    title: String(entityField(row, 'title', '') ?? ''),
    content: String(entityField(row, 'content', '') ?? ''),
    sender: (entityField(row, 'sender', null) as NotificationRow['sender']) || null,
    sender_id: entityField(row, 'sender_id', null) as number | string | null,
    receiver: (entityField(row, 'receiver', null) as NotificationRow['receiver']) || null,
    receiver_id: entityField(row, 'receiver_id', null) as number | string | null,
    is_read: !!(entityField(row, 'is_read', false) || entityField(row, 'IsRead', false)),
    read_at: String(entityField(row, 'read_at', '') ?? '') || null,
    created_at: String(entityField(row, 'created_at', '') ?? ''),
    name: String(entityField(row, 'title', '') ?? ''),
  }
}

export function buildNotificationParams(
  searchForm: Record<string, unknown>,
  baseParams: Record<string, unknown>,
) {
  const params = buildSearchParams(searchForm, baseParams)
  const isRead = searchForm.is_read
  if (isRead !== '' && isRead !== null && isRead !== undefined) {
    if (isRead === 'true' || isRead === true) params.is_read = true
    else if (isRead === 'false' || isRead === false) params.is_read = false
  }
  return params
}

export function getNotificationTypeLabel(t: (key: string) => string, type?: string) {
  if (type === 'message') return t('notification.types.message')
  if (type === 'notice') return t('notification.types.notice')
  return t('notification.types.announcement')
}
