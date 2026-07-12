export const userInitialSearchForm = {
  username: '',
  email: '',
  phone: '',
  status: ''
}

export function createUserSearchFields(t, getStatusOptions) {
  return [
    { prop: 'username', label: t('table.username'), type: 'input', width: '200px', advanced: false },
    { prop: 'email', label: t('table.email'), type: 'input', width: '200px', advanced: false },
    { prop: 'phone', label: t('table.phone'), type: 'input', width: '200px', advanced: false },
    {
      prop: 'status',
      label: t('table.status'),
      type: 'select',
      width: '150px',
      options: getStatusOptions(t),
      advanced: false
    }
  ]
}

export function createUserTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
    { field: 'username', title: t('table.username'), sortable: false, key: 'username' },
    { field: 'nickname', title: t('table.nickname'), sortable: false, key: 'nickname' },
    { field: 'email', title: t('table.email'), sortable: false, key: 'email' },
    { field: 'phone', title: t('table.phone'), sortable: false, key: 'phone' },
    { field: 'balance', title: t('user.balance'), width: 120, sortable: true, slot: 'balance', key: 'balance' },
    { field: 'status', title: t('table.status'), width: 100, sortable: false, slot: 'status', key: 'status' },
    { field: 'created_at', title: t('table.created_at'), width: 180, sortable: true, key: 'created_at' },
    { field: 'operation', title: t('table.operation'), width: 220, fixed: 'right', slot: 'operation', key: 'operation' }
  ]
}

export function formatBalance(amount, currency) {
  const symbol = currency?.symbol || '¥'
  const decimalPlaces = currency?.decimal_places ?? 2
  return `${symbol}${Number(amount || 0).toFixed(decimalPlaces)}`
}
