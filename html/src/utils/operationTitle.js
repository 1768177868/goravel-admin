function convertPluralToSingular(word) {
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

function pluralToSingular(plural) {
  if (!plural) return plural
  const normalized = plural.replace(/-/g, '_')
  const singularMap = {
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
    'user-balance-logs': 'user_balance_log'
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

const LEGACY_TITLE_MAP = {
  'pprof.verify': 'observability.pprof_verify',
  'pprof.cpu_hotspots': 'observability.pprof_cpu_hotspots',
  'pprof.memory_hotspots': 'observability.pprof_memory_hotspots',
  'online-admin.kick-out': 'online_admin.kick_out',
  'online-admin.batch-kick-out': 'online_admin.batch_kick_out'
}

/**
 * Translate operation/permission title keys for logs and dashboard activity.
 */
export function getOperationTitle(t, te, tm, titleKey) {
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
  if (typeof te === 'function' && te(slugKey)) {
    return t(slugKey)
  }

  const messages = typeof tm === 'function' ? tm('permission') : null
  if (messages && Object.prototype.hasOwnProperty.call(messages, slug)) {
    const value = messages[slug]
    if (typeof value === 'string') return value
  }

  if (normalizedTitleKey.startsWith('operation.')) {
    const translated = t(normalizedTitleKey)
    if (translated !== normalizedTitleKey) return translated
  }

  if (slug !== normalizedTitleKey) {
    const originalSlugKey = `permission.${normalizedTitleKey}`
    if (typeof te === 'function' && te(originalSlugKey)) {
      return t(originalSlugKey)
    }
    const originalMessages = typeof tm === 'function' ? tm('permission') : null
    if (originalMessages && Object.prototype.hasOwnProperty.call(originalMessages, normalizedTitleKey)) {
      const value = originalMessages[normalizedTitleKey]
      if (typeof value === 'string') return value
    }
  }

  return normalizedTitleKey
}

export function createOperationTitleTranslator(i18n) {
  const { t, te, tm } = i18n
  return (titleKey) => getOperationTitle(t, te, tm, titleKey)
}
