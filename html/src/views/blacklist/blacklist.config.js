export const blacklistInitialSearchForm = {
  ip: '',
  status: ''
}

export function createBlacklistSearchFields(t) {
  return [
    {
      prop: 'ip',
      label: t('blacklist.ip'),
      type: 'input',
      width: '200px',
      advanced: false
    },
    {
      prop: 'status',
      label: t('table.status'),
      type: 'select',
      width: '150px',
      advanced: false,
      options: [
        { label: t('blacklist.enabled'), value: '1' },
        { label: t('blacklist.disabled'), value: '0' }
      ]
    }
  ]
}

export function createBlacklistTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
    { field: 'ip', title: t('blacklist.ip'), sortable: false, slot: 'ip', key: 'ip' },
    { field: 'remark', title: t('blacklist.remark'), sortable: false, key: 'remark' },
    { field: 'status', title: t('table.status'), width: 100, sortable: false, slot: 'status', key: 'status' },
    { field: 'created_at', title: t('table.created_at'), width: 180, sortable: true, key: 'created_at' },
    { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', sortable: false, key: 'operation' }
  ]
}

export function formatBlacklistIP(ip) {
  if (!ip) return '-'
  if (ip.includes('-')) {
    const parts = ip.split('-')
    if (parts.length === 2) {
      return `${parts[0].trim()} ~ ${parts[1].trim()}`
    }
  }
  return ip
}
