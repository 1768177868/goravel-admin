/**
 * 通用类型定义
 * 这些类型定义可以在 .js 和 .ts 文件中使用
 * 使用 JSDoc 注释在 JS 文件中引用这些类型
 */

// ==================== 基础类型 ====================

/** 分页参数 */
export interface Pagination {
  page: number
  pageSize: number
  total: number
}

/** 分页响应 */
export interface PaginatedResponse<T> {
  list: T[]
  total: number
}

/** API 响应基础结构 */
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

/** 表格列配置 */
export interface TableColumn {
  field?: string
  title?: string
  type?: 'checkbox' | 'seq' | 'radio'
  width?: number | string
  minWidth?: number | string
  sortable?: boolean
  fixed?: 'left' | 'right'
  slot?: string
  formatter?: (params: { row: any; cellValue: any }) => string
  treeNode?: boolean
}

/** 搜索字段配置 */
export interface SearchField {
  prop: string
  label: string
  type: 'input' | 'select' | 'datetime' | 'date' | 'tree-select' | 'cascader'
  width?: string
  options?: Array<{ label: string; value: string | number }>
  apiUrl?: string
  treeProps?: Record<string, string>
  filterable?: boolean
  clearable?: boolean
  allowCreate?: boolean
  advanced?: boolean
  valueFormat?: string
}

// ==================== 业务实体类型 ====================

/** 管理员 */
export interface Admin {
  id: number
  username: string
  nickname?: string
  email?: string
  phone?: string
  status: number
  department_id?: number
  is_super_admin?: boolean
  is_2fa_bound?: boolean
  created_at?: string
  updated_at?: string
  Department?: Department
  Roles?: Role[]
}

/** 角色 */
export interface Role {
  id: number
  name: string
  slug: string
  description?: string
  status: number
  sort?: number
  created_at?: string
}

/** 部门 */
export interface Department {
  id: number
  parent_id?: number
  name: string
  remark?: string
  status: number
  sort?: number
  children?: Department[]
  created_at?: string
}

/** 菜单 */
export interface Menu {
  id: number
  parent_id?: number
  title: string
  slug?: string
  path?: string
  icon?: string
  link_type?: number
  open_type?: number
  status: number
  sort?: number
  children?: Menu[]
  created_at?: string
}

/** 权限 */
export interface Permission {
  id: number
  name: string
  slug: string
  method?: string
  path?: string
  description?: string
  menu_id?: number
  status: number
  sort?: number
  created_at?: string
}

/** 字典 */
export interface Dictionary {
  id: number
  type: string
  label: string
  value: string
  sort?: number
  status: number
  created_at?: string
}

/** 附件 */
export interface Attachment {
  id: number
  filename: string
  display_name?: string
  file_type: string
  disk: string
  extension: string
  size: number
  mime_type: string
  file_url?: string
  created_at?: string
  Admin?: Admin
}

/** 操作日志 */
export interface OperationLog {
  id: number
  admin_id: number
  title: string
  method: string
  path: string
  ip?: string
  user_agent?: string
  request_body?: string
  response_body?: string
  status: number
  duration?: number
  created_at?: string
  Admin?: Admin
}

/** 登录日志 */
export interface LoginLog {
  id: number
  admin_id: number
  ip?: string
  location?: string
  user_agent?: string
  status: number
  message?: string
  request?: string
  created_at?: string
  Admin?: Admin
}

/** 系统日志 */
export interface SystemLog {
  id: number
  level: string
  trace_id?: string
  message: string
  context?: any
  created_at?: string
}

/** 黑名单 */
export interface Blacklist {
  id: number
  type: string
  value: string
  reason?: string
  expire_at?: string
  created_at?: string
}

/** 在线管理员 */
export interface OnlineAdmin {
  id: number
  admin_id: number
  token: string
  ip?: string
  user_agent?: string
  last_activity?: string
  Admin?: Admin
}

/** 通知 */
export interface Notification {
  id: number
  admin_id: number
  title: string
  content: string
  type: string
  is_read: boolean
  read_at?: string
  created_at?: string
}

/** 导出记录 */
export interface ExportRecord {
  id: number
  admin_id: number
  filename: string
  type: string
  status: number
  file_path?: string
  error?: string
  created_at?: string
  Admin?: Admin
}

