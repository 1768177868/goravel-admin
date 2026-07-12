import { uniqBy } from 'lodash-es'
import { getField } from '@/utils/normalizeFormData'

export const adminInitialSearchForm = {
  username: '',
  status: '',
  role_id: '',
  department_id: '',
  position_id: '',
  is_2fa_bound: ''
}

export const protectedAdminIds = [1, 2]

export function isProtectedAdmin(adminId) {
  return protectedAdminIds.includes(adminId)
}

export function createAdminSearchFields(t, getStatusOptions) {
  return [
    { prop: 'username', label: t('table.username'), type: 'input', width: '200px', advanced: false },
    { prop: 'status', label: t('table.status'), type: 'select', width: '150px', options: getStatusOptions(t), advanced: false },
    {
      prop: 'is_2fa_bound',
      label: t('admin.google_auth_status'),
      type: 'select',
      width: '150px',
      options: [
        { label: t('admin.google_auth_bound'), value: '1' },
        { label: t('admin.google_auth_not_bound'), value: '0' }
      ],
      advanced: false
    },
    { prop: 'role_id', label: t('menu.role'), type: 'select', width: '150px', filterable: true, apiUrl: '/options?type=role', advanced: false },
    {
      prop: 'department_id',
      label: t('table.department'),
      type: 'tree-select',
      width: '200px',
      filterable: true,
      apiUrl: '/options?type=department',
      treeProps: { label: 'name', children: 'children', value: 'id' },
      advanced: false
    },
    { prop: 'position_id', label: t('table.position'), type: 'select', width: '180px', filterable: true, apiUrl: '/options?type=position', advanced: false }
  ]
}

export function createAdminTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
    { field: 'username', title: t('table.username'), sortable: false, key: 'username' },
    { field: 'nickname', title: t('table.nickname'), sortable: false, key: 'nickname' },
    { field: 'email', title: t('table.email'), sortable: false, key: 'email' },
    { field: 'phone', title: t('table.phone'), sortable: false, key: 'phone' },
    { field: 'status', title: t('table.status'), width: 100, sortable: false, slot: 'status', key: 'status' },
    { field: 'is_2fa_bound', title: t('admin.google_auth_status'), width: 120, sortable: false, slot: 'is_2fa_bound', key: 'is_2fa_bound' },
    { field: 'department', title: t('table.department'), slot: 'department', sortable: false, key: 'department' },
    { field: 'position', title: t('table.position'), slot: 'position', sortable: false, key: 'position', required: true },
    { field: 'roles', title: t('table.roles'), slot: 'roles', sortable: false, key: 'roles' },
    { field: 'created_at', title: t('table.created_at'), sortable: true, key: 'created_at', width: 160 },
    { title: t('table.operation'), width: 220, fixed: 'right', slot: 'operation', key: 'operation' }
  ]
}

export function getUniqueRoles(roles) {
  if (!Array.isArray(roles)) return []
  return uniqBy(
    roles.filter((role) => role?.id),
    (role) => role.id
  )
}

export function getDepartmentDisplayName(department) {
  if (!department) return '-'
  return getField(department, 'name', '-')
}

export function getPositionDisplayName(position) {
  if (!position) return '-'
  const name = getField(position, 'name', '')
  const id = getField(position, 'id', 0)
  if (!name && (!id || id === 0)) return '-'
  return name || '-'
}
