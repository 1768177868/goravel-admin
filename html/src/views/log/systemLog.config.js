export const systemLogInitialSearchForm = {
  level: '',
  module: '',
  trace_id: '',
  message: '',
  start_time: '',
  end_time: ''
}

export function transformSystemLogRow(log) {
  if (!log) {
    return {
      id: 0,
      level: '',
      module: '',
      trace_id: '',
      message: '',
      context: null,
      created_at: ''
    }
  }

  let context = null
  try {
    if (log.context) {
      context = typeof log.context === 'string' ? JSON.parse(log.context) : log.context
    }
  } catch {
    context = log.context || null
  }

  return {
    id: log.id || 0,
    level: log.level || '',
    module: log.module || '',
    trace_id: log.trace_id || '',
    message: log.message || '',
    context,
    created_at: log.created_at || ''
  }
}

export function createSystemLogSearchFields(t, moduleOptions = []) {
  return [
    {
      prop: 'level',
      label: t('log.level'),
      type: 'select',
      width: '120px',
      options: [
        { label: t('log.level_error'), value: 'error' },
        { label: t('log.level_warning'), value: 'warning' },
        { label: t('log.level_info'), value: 'info' },
        { label: t('log.level_debug'), value: 'debug' }
      ],
      advanced: false
    },
    {
      prop: 'module',
      label: t('log.module'),
      type: 'select',
      width: '150px',
      options: moduleOptions,
      filterable: true,
      clearable: true,
      allowCreate: true,
      advanced: false
    },
    { prop: 'trace_id', label: t('log.trace_id'), type: 'input', width: '200px', advanced: false },
    { prop: 'message', label: t('log.message'), type: 'input', width: '200px', advanced: false },
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

export function createSystemLogTableColumns(t, getModuleLabel) {
  return [
    { type: 'checkbox', width: 60 },
    { field: 'id', title: t('table.id'), width: 80, sortable: true },
    { field: 'level', title: t('log.level'), width: 100, sortable: false, slot: 'level' },
    {
      field: 'module',
      title: t('log.module'),
      width: 120,
      sortable: false,
      formatter: ({ row }) => getModuleLabel(row.module)
    },
    {
      field: 'trace_id',
      title: t('log.trace_id'),
      width: 220,
      sortable: false,
      formatter: ({ row }) => row.trace_id || '-'
    },
    { field: 'message', title: t('log.message'), sortable: false },
    { field: 'context', title: t('log.context'), width: 200, slot: 'context', sortable: false },
    { field: 'created_at', title: t('log.time'), width: 180, sortable: true },
    { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', sortable: false }
  ]
}

export function getSystemLogLevelType(level) {
  const levelMap = {
    error: 'danger',
    warning: 'warning',
    info: 'success',
    debug: 'info'
  }
  return levelMap[level?.toLowerCase()] || 'info'
}

export function getSystemLogLevelLabel(t, level) {
  if (!level) return '-'
  const levelMap = {
    error: t('log.level_error'),
    warning: t('log.level_warning'),
    info: t('log.level_info'),
    debug: t('log.level_debug')
  }
  return levelMap[level.toLowerCase()] || level
}

export function getSystemLogModuleLabel(t, te, module) {
  if (!module) return '-'
  const normalized = String(module).trim()
  if (!normalized) return '-'

  const compatibilityMap = {
    'operation-log': 'operation_log',
    'login-log': 'login_log',
    'system-log': 'system_log',
    'online-admin': 'online_admin',
    'background-task': 'module_background_task'
  }
  const mapped = compatibilityMap[normalized] || normalized
  const snake = mapped.replace(/-/g, '_')

  const menuKey = `menu.${snake}`
  if (typeof te === 'function' && te(menuKey)) {
    return t(menuKey)
  }

  const logKey = snake.startsWith('module_') ? `log.${snake}` : `log.module_${snake}`
  if (typeof te === 'function' && te(logKey)) {
    return t(logKey)
  }

  return normalized
}

export function formatSystemLogContext(context) {
  if (!context) return '-'
  try {
    if (typeof context === 'string') {
      return JSON.stringify(JSON.parse(context), null, 2)
    }
    return JSON.stringify(context, null, 2)
  } catch {
    return String(context)
  }
}

export function formatSystemLogContextPreview(context) {
  if (!context) return '-'
  try {
    let obj = context
    if (typeof context === 'string') {
      obj = JSON.parse(context)
    }
    if (typeof obj === 'object' && obj !== null) {
      const keys = Object.keys(obj)
      if (keys.length === 0) return '{}'
      const preview = keys.slice(0, 2).map((key) => {
        const value = obj[key]
        const valueStr = typeof value === 'object' ? JSON.stringify(value) : String(value)
        return `${key}: ${valueStr.length > 20 ? `${valueStr.substring(0, 20)}...` : valueStr}`
      }).join(', ')
      return keys.length > 2 ? `${preview}...` : preview
    }
    return String(obj)
  } catch {
    return String(context)
  }
}
