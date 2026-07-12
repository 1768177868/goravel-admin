import { getField } from './normalizeFormData'

/**
 * Build standard edit/delete action buttons for TableActionButtons.
 */
export function createCrudActions(t, moduleName, handlers) {
  const actions = []

  if (handlers.onEdit) {
    actions.push({
      key: 'edit',
      label: t('common.edit'),
      type: 'primary',
      permission: `${moduleName}.update`,
      handler: handlers.onEdit
    })
  }

  if (handlers.onDelete) {
    actions.push({
      key: 'delete',
      label: t('common.delete'),
      type: 'danger',
      permission: `${moduleName}.destroy`,
      handler: handlers.onDelete
    })
  }

  return actions
}

/**
 * Read normalized status value from a row.
 */
export function rowStatus(row, defaultValue = 1) {
  return Number(getField(row, 'status', defaultValue))
}

/**
 * Read display text from a row field.
 */
export function rowField(row, field, defaultValue = '-') {
  const value = getField(row, field, defaultValue)
  return value === undefined || value === null || value === '' ? defaultValue : value
}
