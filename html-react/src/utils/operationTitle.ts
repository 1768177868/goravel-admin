import type { TFunction } from 'i18next'

function convertPluralToSingular(word: string): string {
  if (!word || word.length <= 1) return word
  if (word.endsWith('s')) {
    if (word.endsWith('ies') && word.length > 3) {
      return `${word.slice(0, -3)}y`
    }
    if (word.endsWith('es') && word.length > 2) {
      return word.slice(0, -2)
    }
    return word.slice(0, -1)
  }
  return word
}

function pluralToSingular(plural: string): string {
  if (!plural) return plural
  const normalized = plural.replace(/-/g, '_')
  const singularMap: Record<string, string> = {
    roles: 'role',
    permissions: 'permission',
    menus: 'menu',
    departments: 'department',
    dictionaries: 'dictionary',
    blacklists: 'blacklist',
    admins: 'admin',
    users: 'user',
    notifications: 'notification',
    operation_logs: 'operation_log',
    login_logs: 'login_log',
    system_logs: 'system_log',
    online_admins: 'online-admin',
    'operation-logs': 'operation_log',
    'login-logs': 'login_log',
    'system-logs': 'system_log',
    'online-admins': 'online-admin',
    'user-balance-logs': 'user_balance_log',
  }
  if (singularMap[plural]) return singularMap[plural]
  if (singularMap[normalized]) return singularMap[normalized]
  if (normalized.includes('_')) {
    const parts = normalized.split('_')
    const lastPart = parts[parts.length - 1]
    const singularLastPart = convertPluralToSingular(lastPart)
    if (singularLastPart !== lastPart) {
      return `${parts.slice(0, -1).join('_')}_${singularLastPart}`
    }
  }
  return convertPluralToSingular(normalized)
}

const LEGACY_TITLE_MAP: Record<string, string> = {
  'pprof.verify': 'observability.pprof_verify',
  'pprof.cpu_hotspots': 'observability.pprof_cpu_hotspots',
  'pprof.memory_hotspots': 'observability.pprof_memory_hotspots',
  'online-admin.kick-out': 'online_admin.kick_out',
  'online-admin.batch-kick-out': 'online_admin.batch_kick_out',
}

export function getOperationTitle(t: TFunction, titleKey?: string | null): string {
  if (!titleKey) return '-'

  const normalizedTitleKey = LEGACY_TITLE_MAP[titleKey] || titleKey

  let slug = normalizedTitleKey
  if (slug.includes('.')) {
    const parts = slug.split('.')
    if (parts.length >= 2) {
      const module = pluralToSingular(parts[0])
      slug = `${module}.${parts.slice(1).join('.')}`
    }
  } else {
    slug = pluralToSingular(slug)
  }

  const slugKey = `permission.${slug}`
  const translated = t(slugKey, { defaultValue: '__missing__' })
  if (translated !== '__missing__' && translated !== slugKey) {
    return translated
  }

  if (normalizedTitleKey.startsWith('operation.')) {
    const translated = t(normalizedTitleKey)
    if (translated !== normalizedTitleKey) return translated
  }

  if (slug !== normalizedTitleKey) {
    const originalSlugKey = `permission.${normalizedTitleKey}`
    const originalTranslated = t(originalSlugKey, { defaultValue: '__missing__' })
    if (originalTranslated !== '__missing__' && originalTranslated !== originalSlugKey) {
      return originalTranslated
    }
  }

  return normalizedTitleKey
}
