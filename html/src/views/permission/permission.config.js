export const permissionInitialSearchForm = {
  name: '',
  slug: '',
  method: '',
  path: '',
  status: '',
  menu_id: ''
}

export function createPermissionSearchFields(t) {
  return [
    { prop: 'name', label: t('permission.name'), type: 'input', width: '200px', advanced: false },
    { prop: 'slug', label: t('permission.slug'), type: 'input', width: '200px', advanced: false },
    {
      prop: 'method',
      label: t('permission.method'),
      type: 'select',
      width: '150px',
      options: [
        { label: 'GET', value: 'GET' },
        { label: 'POST', value: 'POST' },
        { label: 'PUT', value: 'PUT' },
        { label: 'DELETE', value: 'DELETE' },
        { label: 'PATCH', value: 'PATCH' }
      ],
      advanced: false
    },
    { prop: 'path', label: t('permission.path'), type: 'input', width: '200px', advanced: false },
    {
      prop: 'status',
      label: t('table.status'),
      type: 'select',
      width: '120px',
      options: [
        { label: t('common.enabled'), value: '1' },
        { label: t('common.disabled'), value: '0' }
      ],
      advanced: false
    },
    {
      prop: 'menu_id',
      label: t('menu.title'),
      type: 'tree-select',
      width: '200px',
      filterable: true,
      apiUrl: '/options?type=menu',
      treeProps: { label: 'label', value: 'value', children: 'children' },
      advanced: false
    }
  ]
}

export function createPermissionTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
    { field: 'name', title: t('permission.name'), sortable: false, key: 'name' },
    { field: 'slug', title: t('permission.slug'), sortable: false, key: 'slug' },
    { field: 'method', title: t('permission.method'), width: 100, sortable: false, key: 'method' },
    { field: 'path', title: t('permission.path'), sortable: false, key: 'path' },
    { field: 'description', title: t('common.description'), sortable: false, key: 'description' },
    { field: 'menu', title: t('menu.title'), width: 150, slot: 'menu', sortable: false, key: 'menu' },
    { field: 'status', title: t('table.status'), width: 80, sortable: false, slot: 'status', key: 'status' },
    { field: 'sort', title: t('common.sort'), width: 80, sortable: true, key: 'sort' },
    { field: 'created_at', title: t('table.created_at'), sortable: true, key: 'created_at' },
    { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', sortable: false, key: 'operation' }
  ]
}
