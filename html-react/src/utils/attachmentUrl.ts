export const ATTACHMENT_PUBLIC_PATH_RE = /\/api\/admin\/public\/images\/(\d+)/
export const ATTACHMENT_PUBLIC_ALIAS_PATH_RE = /\/api\/public\/files\/(\d+)/
export const ATTACHMENT_PRIVATE_PREVIEW_PATH_RE = /\/api\/admin\/attachments\/(\d+)\/preview/

interface AttachmentUploadData {
  id?: number
  is_public?: number
  file_url?: string
}

export function toStablePublicAttachmentUrl(data: AttachmentUploadData | null | undefined) {
  if (!data?.id) return ''
  if (Number(data.is_public) === 0) return ''
  return `/api/admin/public/images/${data.id}`
}

export function resolveUploadStorageUrl(data: AttachmentUploadData | null | undefined) {
  const stable = toStablePublicAttachmentUrl(data)
  if (stable) return stable
  return data?.file_url || ''
}

export function isPublicAttachmentPath(value: unknown) {
  const path = String(value || '')
  return ATTACHMENT_PUBLIC_PATH_RE.test(path) || ATTACHMENT_PUBLIC_ALIAS_PATH_RE.test(path)
}

export function isPrivateAttachmentPreviewPath(value: unknown) {
  return ATTACHMENT_PRIVATE_PREVIEW_PATH_RE.test(String(value || ''))
}
