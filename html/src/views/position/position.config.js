export const positionInitialSearchForm = {
  name: ''
}

export function createPositionSearchFields(t) {
  return [
    {
      prop: 'name',
      label: t('position.name'),
      type: 'input',
      width: '200px',
      advanced: false
    }
  ]
}

export function createPositionTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
    { field: 'name', title: t('position.name'), sortable: false, key: 'name' },
    { field: 'code', title: t('position.code'), sortable: false, key: 'code' },
    { field: 'sort', title: t('common.sort'), width: 80, sortable: true, key: 'sort' },
    { field: 'status', title: t('table.status'), width: 80, sortable: false, slot: 'status', key: 'status' },
    { field: 'created_at', title: t('table.created_at'), sortable: true, key: 'created_at' },
    { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', sortable: false, key: 'operation' }
  ]
}
