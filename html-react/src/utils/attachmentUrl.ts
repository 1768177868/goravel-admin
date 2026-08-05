export const ATTACHMENT_PUBLIC_PATH_RE = /\/api\/admin\/public\/images\/(\d+)/
export const ATTACHMENT_PUBLIC_ALIAS_PATH_RE = /\/api\/public\/files\/(\d+)/
export const ATTACHMENT_PRIVATE_PREVIEW_PATH_RE = /\/api\/admin\/attachments\/(\d+)\/preview/

export function isPublicAttachmentPath(value: unknown) {
  const path = String(value || '')
  return ATTACHMENT_PUBLIC_PATH_RE.test(path) || ATTACHMENT_PUBLIC_ALIAS_PATH_RE.test(path)
}

export function isPrivateAttachmentPreviewPath(value: unknown) {
  return ATTACHMENT_PRIVATE_PREVIEW_PATH_RE.test(String(value || ''))
}
