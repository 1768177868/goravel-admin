import axios from 'axios'
import Storage from '@/utils/storage'
import { getApiPrefix, resolvePublicAssetUrl } from '@/utils/env'

export const PUBLIC_IMAGE_PATH_RE = /\/api\/admin\/public\/images\/(\d+)/

/**
 * Build a fetchable URL for a stored image path (public preview / relative / absolute).
 */
export function buildImageFetchUrl(raw) {
  const value = String(raw || '').trim()
  if (!value) return ''
  if (value.startsWith('data:') || value.startsWith('blob:')) return value
  if (/^https?:\/\//i.test(value)) return value

  const publicPath = resolvePublicAssetUrl(value)
  const path = publicPath.startsWith('/') ? publicPath : value
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL
  if (apiBaseURL) {
    return `${String(apiBaseURL).replace(/\/+$/, '')}${path.startsWith('/') ? path : `/${path}`}`
  }
  const prefix = getApiPrefix().startsWith('/') ? getApiPrefix() : `/${getApiPrefix()}`
  if (path.startsWith(prefix) || path.startsWith('/')) return path
  return `${prefix}/${path}`
}

/**
 * Resolve a displayable image URL for <img>/el-image.
 * Public API paths are fetched as blob (same as attachment list) so they work in dev with VITE_API_BASE_URL.
 * @returns {Promise<{ url: string, revoke?: () => void }>}
 */
export async function resolveImageDisplayUrl(raw) {
  const value = String(raw || '').trim()
  if (!value) return { url: '' }

  if (value.startsWith('data:') || value.startsWith('blob:')) {
    return { url: value }
  }

  const resolvedPublic = resolvePublicAssetUrl(value) || ''
  const isPublicImage = PUBLIC_IMAGE_PATH_RE.test(value) || PUBLIC_IMAGE_PATH_RE.test(resolvedPublic)

  // Permanent external URLs (CDN etc.)
  if (/^https?:\/\//i.test(value) && !isPublicImage) {
    return { url: value }
  }

  const fetchUrl = buildImageFetchUrl(value)
  if (!fetchUrl) return { url: '' }

  try {
    const token = Storage.getItem('token', '') || ''
    const response = await axios.get(fetchUrl, {
      responseType: 'blob',
      headers: {
        Authorization: `Bearer ${typeof token === 'string' ? token.trim() : ''}`
      }
    })
    const blobUrl = URL.createObjectURL(new Blob([response.data]))
    return {
      url: blobUrl,
      revoke: () => URL.revokeObjectURL(blobUrl)
    }
  } catch {
    return { url: resolvedPublic || fetchUrl }
  }
}

export const WEBSITE_CONFIG_UPDATED_EVENT = 'website-config-updated'

export function notifyWebsiteConfigUpdated() {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent(WEBSITE_CONFIG_UPDATED_EVENT))
  }
}
