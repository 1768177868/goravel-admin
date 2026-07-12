export const roleInitialSearchForm = {
  name: '',
  status: ''
}

export const protectedRoleSlugs = ['super-admin']

export function createRoleSearchFields(t) {
  return [
    {
      prop: 'name',
      label: t('role.name'),
      type: 'input',
      width: '200px',
      advanced: false
    },
    {
      prop: 'status',
      label: t('table.status'),
      type: 'select',
      width: '150px',
      options: [
        { label: t('common.enabled'), value: '1' },
        { label: t('common.disabled'), value: '0' }
      ],
      advanced: false
    }
  ]
}

export function createRoleTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
    { field: 'name', title: t('role.name'), sortable: false, key: 'name' },
    { field: 'slug', title: t('role.slug'), sortable: false, key: 'slug' },
    { field: 'description', title: t('common.description'), sortable: false, key: 'description' },
    { field: 'status', title: t('table.status'), width: 100, sortable: false, slot: 'status', key: 'status' },
    { field: 'sort', title: t('common.sort'), width: 80, sortable: true, key: 'sort' },
    { field: 'created_at', title: t('table.created_at'), sortable: true, key: 'created_at' },
    { title: t('table.operation'), width: 200, fixed: 'right', slot: 'operation', sortable: false, key: 'operation' }
  ]
}

export function isProtectedRole(row) {
  const slug = row?.slug || ''
  return protectedRoleSlugs.includes(slug)
}
