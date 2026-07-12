export function createUserBalanceLogInitialSearch(userId = '') {
  return {
    user_id: userId,
    type: '',
    source: '',
    start_time: '',
    end_time: ''
  }
}

export function createUserBalanceLogSearchFields(t) {
  return [
    {
      prop: 'type',
      label: t('user.type'),
      type: 'select',
      width: '150px',
      options: [
        { label: t('user.balance_income'), value: 'income' },
        { label: t('user.balance_expense'), value: 'expense' },
        { label: t('user.balance_refund'), value: 'refund' }
      ],
      clearable: true,
      advanced: false
    },
    {
      prop: 'source',
      label: t('user.source'),
      type: 'select',
      width: '150px',
      options: [
        { label: t('user.source_order'), value: 'order' },
        { label: t('user.source_recharge'), value: 'recharge' },
        { label: t('user.source_withdraw'), value: 'withdraw' },
        { label: t('user.source_manual'), value: 'manual' }
      ],
      clearable: true,
      advanced: false
    }
  ]
}

export function createUserBalanceLogTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 180, key: 'id' },
    { field: 'type', title: t('user.type'), width: 100, slot: 'type', key: 'type' },
    { field: 'amount', title: t('user.amount'), width: 120, slot: 'amount', key: 'amount' },
    { field: 'balance', title: t('user.balance'), width: 120, slot: 'balance', key: 'balance' },
    { field: 'source', title: t('user.source'), width: 100, slot: 'source', key: 'source' },
    { field: 'description', title: t('user.description'), minWidth: 200, key: 'description' },
    { field: 'created_at', title: t('table.created_at'), width: 180, key: 'created_at' }
  ]
}

export function formatMoney(amount) {
  return Number(amount || 0).toFixed(2)
}

export function buildBalanceLogParams(searchForm, baseParams) {
  const params = {
    ...baseParams,
    user_id: searchForm.user_id
  }
  if (searchForm.type) params.type = searchForm.type
  if (searchForm.source) params.source = searchForm.source
  if (searchForm.start_time) params.start_time = searchForm.start_time
  if (searchForm.end_time) params.end_time = searchForm.end_time
  return params
}
