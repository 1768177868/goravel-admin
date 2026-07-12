import { buildSearchParams } from '@/utils/buildSearchParams'
import { getField } from '@/utils/normalizeFormData'

export const notificationInitialSearchForm = {
  type: '',
  is_read: ''
}

export function transformNotificationRow(notification) {
  return {
    id: getField(notification, 'id'),
    type: getField(notification, 'type', ''),
    title: getField(notification, 'title', ''),
    content: getField(notification, 'content', ''),
    sender: notification.sender || notification.Sender || null,
    sender_id: getField(notification, 'sender_id', null),
    receiver: notification.receiver || notification.Receiver || null,
    receiver_id: getField(notification, 'receiver_id', null),
    is_read: getField(notification, 'is_read', false),
    read_at: getField(notification, 'read_at', null),
    created_at: getField(notification, 'created_at', '')
  }
}

export function createNotificationSearchFields(t) {
  return [
    {
      prop: 'type',
      label: t('notification.table.type'),
      type: 'select',
      options: [
        { label: t('common.all'), value: '' },
        { label: t('notification.types.announcement'), value: 'announcement' },
        { label: t('notification.types.notice'), value: 'notice' },
        { label: t('notification.types.message'), value: 'message' }
      ],
      clearable: true
    },
    {
      prop: 'is_read',
      label: t('notification.table.status'),
      type: 'select',
      options: [
        { label: t('common.all'), value: '' },
        { label: t('notification.unread'), value: 'false' },
        { label: t('notification.read'), value: 'true' }
      ],
      clearable: true
    }
  ]
}

export function createNotificationTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 80 },
    { field: 'title', title: t('notification.table.title'), minWidth: 160 },
    { field: 'content', title: t('notification.table.content'), minWidth: 260, slot: 'content' },
    { field: 'type', title: t('notification.table.type'), width: 150, slot: 'type' },
    { field: 'sender', title: t('notification.table.sender'), width: 160, slot: 'sender' },
    { field: 'is_read', title: t('notification.table.status'), width: 120, slot: 'is_read' },
    {
      field: 'created_at',
      title: t('notification.table.created_at'),
      width: 180,
      sortable: true,
      slot: 'created_at'
    },
    { field: 'operation', title: t('common.operation'), width: 140, fixed: 'right', slot: 'operation' }
  ]
}

export function buildNotificationParams(searchForm, baseParams) {
  const params = buildSearchParams(searchForm, baseParams)

  if (searchForm.is_read !== '' && searchForm.is_read !== null && searchForm.is_read !== undefined) {
    if (searchForm.is_read === 'true') {
      params.is_read = true
    } else if (searchForm.is_read === 'false') {
      params.is_read = false
    }
  }

  return params
}

export function getNotificationTypeLabel(t, type) {
  if (type === 'message') return t('notification.types.message')
  if (type === 'notice') return t('notification.types.notice')
  return t('notification.types.announcement')
}
