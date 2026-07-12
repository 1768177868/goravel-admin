export const onlineAdminInitialSearchForm = {
  username: '',
  ip: '',
  browser: '',
  os: ''
}

export function createOnlineAdminSearchFields(t) {
  return [
    { prop: 'username', label: t('online_admin.username'), type: 'input', width: '200px' },
    { prop: 'ip', label: t('online_admin.ip'), type: 'input', width: '200px' },
    { prop: 'browser', label: t('online_admin.browser'), type: 'input', width: '200px' },
    { prop: 'os', label: t('online_admin.os'), type: 'input', width: '200px' }
  ]
}

export function createOnlineAdminTableColumns(t) {
  return [
    { type: 'checkbox', width: 50, fixed: 'left', key: 'checkbox' },
    { field: 'username', title: t('online_admin.username'), width: 120, sortable: false, key: 'username' },
    { field: 'nickname', title: t('online_admin.nickname'), width: 120, sortable: false, key: 'nickname' },
    { field: 'avatar', title: t('online_admin.avatar'), width: 80, slot: 'avatar', key: 'avatar' },
    { field: 'browser', title: t('online_admin.browser'), width: 150, sortable: false, key: 'browser' },
    { field: 'ip', title: t('online_admin.ip'), width: 150, sortable: false, key: 'ip' },
    { field: 'os', title: t('online_admin.os'), width: 150, sortable: false, key: 'os' },
    { field: 'session_id', title: t('online_admin.session_id'), width: 200, key: 'session_id' },
    { field: 'last_active', title: t('online_admin.last_active'), width: 180, sortable: true, slot: 'last_active', key: 'last_active' },
    { title: t('common.operation'), width: 120, fixed: 'right', slot: 'operation', key: 'operation' }
  ]
}

export function formatOnlineTime(time) {
  if (!time) return '-'
  const date = new Date(time)
  if (Number.isNaN(date.getTime())) return time
  return date.toLocaleString().replace(/\//g, '-')
}
