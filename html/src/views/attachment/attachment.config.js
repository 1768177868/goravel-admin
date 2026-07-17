import { getField } from '@/utils/normalizeFormData'

export const attachmentInitialSearchForm = {
  filename: '',
  display_name: '',
  category_id: '',
  is_public: '',
  file_type: '',
  extension: '',
  start_time: '',
  end_time: ''
}

export function transformAttachmentRow(item) {
  const category = item.category || item.Category || null
  return {
    id: getField(item, 'id'),
    admin: item.admin || item.Admin || null,
    category_id: getField(item, 'category_id', getField(category, 'id', null)),
    category_name: getField(category, 'name', ''),
    filename: getField(item, 'filename', ''),
    display_name: getField(item, 'display_name', ''),
    file_type: getField(item, 'file_type', 'other'),
    disk: getField(item, 'disk', ''),
    extension: getField(item, 'extension', ''),
    size: getField(item, 'size', 0),
    mime_type: getField(item, 'mime_type', ''),
    is_public: Number(getField(item, 'is_public', 1)),
    created_at: getField(item, 'created_at', ''),
    file_url: getField(item, 'file_url', '')
  }
}

export function createAttachmentSearchFields(t) {
  return [
    { prop: 'filename', label: t('attachment.filename'), type: 'input', width: '200px' },
    { prop: 'display_name', label: t('attachment.display_name'), type: 'input', width: '200px' },
    {
      prop: 'category_id',
      label: t('attachment.category'),
      type: 'select',
      width: '160px',
      apiUrl: '/options?type=attachment_category',
      clearable: true
    },
    {
      prop: 'is_public',
      label: t('attachment.visibility'),
      type: 'select',
      width: '130px',
      options: [
        { label: t('attachment.visibility_public'), value: '1' },
        { label: t('attachment.visibility_private'), value: '0' }
      ],
      clearable: true
    },
    {
      prop: 'file_type',
      label: t('attachment.file_type'),
      type: 'select',
      width: '150px',
      options: [
        { label: t('attachment.file_type_image'), value: 'image' },
        { label: t('attachment.file_type_video'), value: 'video' },
        { label: t('attachment.file_type_document'), value: 'document' },
        { label: t('attachment.file_type_other'), value: 'other' }
      ],
      clearable: true
    },
    { prop: 'extension', label: t('attachment.extension'), type: 'input', width: '150px' },
    {
      prop: 'start_time',
      label: t('log.start_time'),
      type: 'datetime',
      width: '180px',
      valueFormat: 'YYYY-MM-DD HH:mm:ss',
      advanced: true
    },
    {
      prop: 'end_time',
      label: t('log.end_time'),
      type: 'datetime',
      width: '180px',
      valueFormat: 'YYYY-MM-DD HH:mm:ss',
      advanced: true
    }
  ]
}

export function formatAttachmentFileSize(size) {
  if (!size) return '-'
  const numSize = Number(size)
  if (numSize < 1024) return `${numSize} B`
  if (numSize < 1024 * 1024) return `${(numSize / 1024).toFixed(2)} KB`
  if (numSize < 1024 * 1024 * 1024) return `${(numSize / 1024 / 1024).toFixed(2)} MB`
  return `${(numSize / 1024 / 1024 / 1024).toFixed(2)} GB`
}

export function formatAttachmentAdmin({ row }) {
  const admin = row.admin || row.Admin
  if (admin) return getField(admin, 'username', '-')
  return '-'
}

export function createAttachmentTableColumns(t) {
  return [
    { type: 'checkbox', width: 60, key: 'checkbox' },
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
    { field: 'filename', title: t('attachment.filename'), minWidth: 200, slot: 'filename', key: 'filename' },
    { field: 'display_name', title: t('attachment.display_name'), minWidth: 200, slot: 'display_name', key: 'display_name' },
    { field: 'category_id', title: t('attachment.category'), minWidth: 160, slot: 'category', key: 'category' },
    { field: 'is_public', title: t('attachment.visibility'), width: 110, slot: 'is_public', key: 'is_public' },
    { field: 'file_type', title: t('attachment.file_type'), width: 120, slot: 'file_type', key: 'file_type' },
    { field: 'disk', title: t('attachment.disk'), width: 100, slot: 'disk', key: 'disk' },
    { field: 'extension', title: t('attachment.extension'), width: 100, key: 'extension' },
    {
      field: 'size',
      title: t('attachment.size'),
      width: 140,
      formatter: ({ cellValue }) => formatAttachmentFileSize(cellValue),
      key: 'size'
    },
    { field: 'mime_type', title: t('attachment.mime_type'), minWidth: 150, key: 'mime_type' },
    {
      field: 'admin',
      title: t('log.admin'),
      width: 140,
      formatter: formatAttachmentAdmin,
      key: 'admin'
    },
    { field: 'created_at', title: t('table.created_at'), width: 180, sortable: true, key: 'created_at' },
    { title: t('table.operation'), width: 200, fixed: 'right', slot: 'operation', key: 'operation' }
  ]
}

export function getAttachmentFileTypeTagType(fileType) {
  const types = {
    image: 'success',
    video: 'warning',
    document: 'info',
    other: null
  }
  return types[fileType] !== undefined ? types[fileType] : null
}

export function getAttachmentFileTypeLabel(t, fileType) {
  const labels = {
    image: t('attachment.file_type_image'),
    video: t('attachment.file_type_video'),
    document: t('attachment.file_type_document'),
    other: t('attachment.file_type_other')
  }
  return labels[fileType] || fileType
}

export const ATTACHMENT_CHUNK_SIZE = 1 * 1024 * 1024
export const ATTACHMENT_LARGE_FILE_THRESHOLD = 5 * 1024 * 1024
