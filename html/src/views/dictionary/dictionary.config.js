export const dictionaryInitialSearchForm = {
  type: ''
}

export function createDictionarySearchFields(t) {
  return [
    {
      prop: 'type',
      label: t('dictionary.type'),
      type: 'input',
      width: '200px',
      advanced: false
    }
  ]
}

export function createDictionaryTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
    { field: 'type', title: t('dictionary.type'), sortable: false, key: 'type' },
    { field: 'label', title: t('dictionary.label'), sortable: false, key: 'label' },
    { field: 'value', title: t('dictionary.value'), sortable: false, key: 'value' },
    { field: 'translation_key', title: t('dictionary.translation_key'), sortable: false, key: 'translation_key' },
    { field: 'sort', title: t('common.sort'), width: 80, sortable: true, key: 'sort' },
    { field: 'status', title: t('table.status'), width: 80, sortable: false, slot: 'status', key: 'status' },
    { field: 'created_at', title: t('table.created_at'), sortable: true, key: 'created_at' },
    { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', sortable: false, key: 'operation' }
  ]
}
