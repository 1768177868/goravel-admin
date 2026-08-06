import type { TFunction } from 'i18next'

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
 * Tries: exact key, kebab→snake, optional `_management` suffix, then fallbacks.
 */
export function resolveMenuTitle(
  t: TFunction,
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

  if (titleKey) {
    pushKey(titleKey)
    pushKey(titleKey.replace(/-/g, '_'))
  }

  if (slug) {
    const normalized = normalizeMenuSlug(slug)
    pushKey(`menu.${normalized}`)
    if (!normalized.endsWith('_management')) {
      pushKey(`menu.${normalized}_management`)
    }
  }

  for (const key of candidates) {
    const translated = t(key)
    if (translated && translated !== key) return translated
  }

  return fallback || titleKey || slug || ''
}
