/** @typedef {{ id?: number, is_public?: number, file_url?: string }} AttachmentUploadData */

export const ATTACHMENT_PUBLIC_PATH_RE = /\/api\/admin\/public\/images\/(\d+)/
export const ATTACHMENT_PUBLIC_ALIAS_PATH_RE = /\/api\/public\/files\/(\d+)/
export const ATTACHMENT_PRIVATE_PREVIEW_PATH_RE = /\/api\/admin\/attachments\/(\d+)\/preview/

/**
 * Stable path to store in config / rich text (public attachments only).
 * @param {AttachmentUploadData | null | undefined} data
 */
export function toStablePublicAttachmentUrl(data) {
  if (!data?.id) return ''
  if (Number(data.is_public) === 0) return ''
  return `/api/admin/public/images/${data.id}`
}

/**
 * Prefer stable relative path from upload response; fall back to file_url.
 * @param {AttachmentUploadData | null | undefined} data
 */
export function resolveUploadStorageUrl(data) {
  const stable = toStablePublicAttachmentUrl(data)
  if (stable) return stable
  return data?.file_url || ''
}

export function isPublicAttachmentPath(value) {
  const path = String(value || '')
  return ATTACHMENT_PUBLIC_PATH_RE.test(path) || ATTACHMENT_PUBLIC_ALIAS_PATH_RE.test(path)
}

export function isPrivateAttachmentPreviewPath(value) {
  return ATTACHMENT_PRIVATE_PREVIEW_PATH_RE.test(String(value || ''))
}

export function extractAttachmentIdFromPath(value) {
  const path = String(value || '')
  for (const re of [ATTACHMENT_PUBLIC_PATH_RE, ATTACHMENT_PUBLIC_ALIAS_PATH_RE, ATTACHMENT_PRIVATE_PREVIEW_PATH_RE]) {
    const match = path.match(re)
    if (match?.[1]) return Number(match[1])
  }
  return 0
}
