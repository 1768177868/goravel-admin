import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { mapTree } from '@/utils/tree'

export const menuInitialSearchForm = {}

export function createMenuSearchFields() {
  return []
}

export function createMenuTableColumns(t) {
  return [
    { type: 'index', width: 60, title: t('table.seq'), key: 'index' },
    { field: 'name', title: t('menu_management.name'), minWidth: 200, key: 'name' },
    { field: 'slug', title: t('menu_management.slug'), minWidth: 150, key: 'slug' },
    { field: 'path', title: t('menu_management.path'), minWidth: 200, key: 'path' },
    { field: 'type', title: t('table.type'), width: 100, slot: 'type', key: 'type' },
    { field: 'link_type', title: t('menu_management.link_type'), width: 120, slot: 'link_type', key: 'link_type' },
    { field: 'open_type', title: t('menu_management.open_type'), width: 140, slot: 'open_type', key: 'open_type' },
    { field: 'no_cache', title: t('menu_management.no_cache'), width: 100, slot: 'no_cache', key: 'no_cache' },
    { field: 'icon', title: t('menu_management.icon'), width: 140, slot: 'icon', key: 'icon' },
    { field: 'sort', title: t('common.sort'), width: 80, key: 'sort' },
    { field: 'status', title: t('table.status'), width: 100, slot: 'status', key: 'status' },
    { field: 'is_hidden', title: t('menu_management.is_hidden'), width: 100, slot: 'is_hidden', key: 'is_hidden' },
    { field: 'created_at', title: t('table.created_at'), width: 180, key: 'created_at' },
    { field: 'updated_at', title: t('table.updated_at'), width: 180, key: 'updated_at' },
    { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', key: 'operation' }
  ]
}

export function normalizeIconName(iconName) {
  if (!iconName) {
    return ''
  }
  const trimmed = iconName.trim()
  if (!trimmed) {
    return ''
  }
  if (ElementPlusIconsVue[trimmed]) {
    return trimmed
  }
  const pascalCase = trimmed.charAt(0).toUpperCase() + trimmed.slice(1)
  if (ElementPlusIconsVue[pascalCase]) {
    return pascalCase
  }
  return ''
}

export function getIconComponent(iconName) {
  const normalized = normalizeIconName(iconName)
  return normalized ? ElementPlusIconsVue[normalized] : null
}

export function mapMenuTreeOptions(tableData) {
  return mapTree(tableData, (menu) => ({
    value: menu.id,
    label: menu.name
  }))
}

export function getMenuTypeTagType(type) {
  if (type === 1) return 'info'
  if (type === 2) return 'primary'
  return 'warning'
}

export function getMenuTypeLabel(t, type) {
  if (type === 1) return t('menu.type_directory')
  if (type === 2) return t('menu.type_menu')
  return t('menu.type_button')
}
