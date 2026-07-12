/**
 * Department list page configuration.
 * Used by DepartmentList.vue and code generator tree templates.
 */

export const departmentInitialSearchForm = {
  name: '',
  status: ''
}

export function createDepartmentSearchFields(t) {
  return [
    {
      prop: 'name',
      label: t('department.name'),
      type: 'input',
      width: '200px',
      advanced: false
    },
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
    }
  ]
}

export function createDepartmentTableColumns(t) {
  return [
    { type: 'index', width: 60, title: t('table.seq'), key: 'index' },
    { field: 'name', title: t('department.name'), minWidth: 150, key: 'name' },
    { field: 'remark', title: t('common.description'), minWidth: 200, slot: 'remark', key: 'remark' },
    { field: 'sort', title: t('common.sort'), width: 100, key: 'sort' },
    { field: 'status', title: t('table.status'), width: 200, slot: 'status', key: 'status' },
    { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', key: 'operation' }
  ]
}
