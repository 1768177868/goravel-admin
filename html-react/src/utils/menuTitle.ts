import type { i18n as I18nInstance } from 'i18next'

/** Normalize menu slug for i18n keys: operation-log -> operation_log */
export function normalizeMenuSlug(slug: string): string {
  return String(slug || '')
    .replace(/^\//, '')
    .replace(/-/g, '_')
}

/** Build preferred i18n key for a menu slug/path segment. */
export function menuTitleKeyFromSlug(slug: string): string {
  return `menu.${normalizeMenuSlug(slug)}`
}

/**
 * Resolve a display title for menu / tab / breadcrumb.
 * Tries: exact key, kebab/snake variants, optional `_management` suffix, then fallbacks.
 * Locale keys may use either `code-generator` or `code_generator` under `menu.*`.
 */
export function resolveMenuTitle(
  t: (key: string) => string,
  options: {
    titleKey?: string | null
    slug?: string | null
    fallback?: string | null
  },
): string {
  const { titleKey, slug, fallback } = options
  const candidates: string[] = []

  const pushKey = (key?: string | null) => {
    if (!key) return
    if (!candidates.includes(key)) candidates.push(key)
  }

  const pushMenuVariants = (raw?: string | null) => {
    if (!raw) return
    const clean = String(raw).replace(/^\//, '')
    if (!clean) return
    const variants = [
      clean,
      clean.replace(/-/g, '_'),
      clean.replace(/_/g, '-'),
    ]
    for (const variant of [...new Set(variants)]) {
      pushKey(`menu.${variant}`)
      if (!variant.endsWith('_management') && !variant.endsWith('-management')) {
        pushKey(`menu.${variant.replace(/-/g, '_')}_management`)
      }
    }
  }

  if (titleKey) {
    pushKey(titleKey)
    pushKey(titleKey.replace(/-/g, '_'))
    pushKey(titleKey.replace(/_/g, '-'))
    if (titleKey.startsWith('menu.')) {
      pushMenuVariants(titleKey.slice('menu.'.length))
    }
  }

  pushMenuVariants(slug)

  for (const key of candidates) {
    const translated = t(key)
    if (translated && translated !== key) return translated
  }

  return fallback || titleKey || slug || ''
}

/**
 * Resolve permission display title by slug.
 * Keys like `admin.index` live as dotted flat keys under `permission`, so we read the
 * resource bundle directly (i18next would treat dots as nested path separators).
 */
export function resolvePermissionTitle(
  i18n: I18nInstance,
  options: {
    slug?: string | null
    fallback?: string | null
  },
): string {
  const { slug, fallback } = options
  if (slug) {
    const bundle = i18n.getResourceBundle(i18n.language, 'translation') as
      | { permission?: Record<string, unknown> }
      | undefined
    const value = bundle?.permission?.[slug]
    if (typeof value === 'string' && value) return value

    const fallbackLng = i18n.options.fallbackLng
    const fallbacks = Array.isArray(fallbackLng)
      ? fallbackLng
      : typeof fallbackLng === 'string'
        ? [fallbackLng]
        : []
    for (const lng of fallbacks) {
      if (lng === i18n.language) continue
      const fbBundle = i18n.getResourceBundle(lng, 'translation') as
        | { permission?: Record<string, unknown> }
        | undefined
      const fbValue = fbBundle?.permission?.[slug]
      if (typeof fbValue === 'string' && fbValue) return fbValue
    }
  }

  return fallback || slug || ''
}
