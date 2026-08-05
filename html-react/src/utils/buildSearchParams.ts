import { forOwn } from 'lodash-es'

/**
 * Build list query params: merge pagination/sort extras with non-empty search fields.
 */
export function buildSearchParams(
  searchForm: Record<string, unknown> = {},
  extraParams: Record<string, unknown> = {},
): Record<string, unknown> {
  const params: Record<string, unknown> = { ...extraParams }

  forOwn(searchForm, (value, key) => {
    if (value === '' || value === null || value === undefined) return

    if (typeof value === 'string') {
      const trimmed = value.trim()
      if (trimmed) params[key] = trimmed
    } else {
      params[key] = value
    }
  })

  return params
}
