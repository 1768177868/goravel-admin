<template>
  <div class="log-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('log.operation_log') }}</span>
          <div class="header-actions">
            <el-button 
              type="danger" 
              :disabled="selectedRows.length === 0 || getButtonState('operation_log.destroy').disabled"
              @click="handleBatchDelete"
            >
              <el-icon><Delete /></el-icon>
              {{ $t('common.delete_selected') }} ({{ selectedRows.length }})
            </el-button>
          </div>
        </div>
      </template>

      <!-- 搜索表单（使用 JSON 配置方式） -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        :loading="loading"
        i18n-prefix="log"
        @search="handleSearch"
        @reset="handleReset"
      />

      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        :column-config="{ resizable: true }"
        height="600"
        :sort-config="{ multiple: false, trigger: 'default' }"
        @checkbox-change="handleSelectionChange"
        @checkbox-all="handleSelectionChange"
        @sort-change="handleSortChange"
      >
        <template v-for="column in tableColumns" :key="column.field || column.title || column.type">
          <vxe-column
            v-if="column.type === 'checkbox'"
            type="checkbox"
            :width="column.width"
            :fixed="column.fixed"
          />
          <vxe-column
            v-else
            :field="column.field"
            :title="column.title"
            :width="column.width"
            :sortable="column.sortable"
            :fixed="column.fixed"
            :formatter="column.formatter"
            :tree-node="column.treeNode"
          >
            <template v-if="column.slot === 'admin'" #default="{ row }">
              {{ (row.admin || row.Admin)?.username || (row.admin || row.Admin)?.Username || '-' }}
            </template>
            <template v-else-if="column.slot === 'title'" #default="{ row }">
              {{ getOperationTitle(row.title || row.Title) }}
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <el-button 
                type="primary" 
                link 
                :disabled="getButtonState('operation_log.show').disabled"
                @click="handleView(row)"
              >
                {{ $t('common.view') }}
              </el-button>
              <el-button 
                type="danger" 
                link 
                :disabled="getButtonState('operation_log.destroy').disabled"
                @click="handleDelete(row)"
              >
                {{ $t('common.delete') }}
              </el-button>
            </template>
          </vxe-column>
        </template>
      </vxe-table>

      <Pagination
        v-model="pagination"
        :show-total="true"
        :show-quick-jumper="true"
        :align="'right'"
        @page-change="handlePageChange"
      />
    </el-card>

    <el-dialog v-model="detailVisible" :title="$t('log.detail')" width="800px">
      <el-descriptions :column="2" border v-if="logDetail">
        <el-descriptions-item :label="$t('table.id')">{{ logDetail.id }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.admin')">{{ logDetail.admin?.username }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.method')">{{ logDetail.method }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.path')">{{ logDetail.path }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.ip')">{{ logDetail.ip }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.status_code')">{{ logDetail.status_code }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.operation_time')" :span="2">{{ logDetail.created_at }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.request_params')" :span="2">
          <pre>{{ JSON.stringify(logDetail.params || logDetail.request || {}, null, 2) }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { getMethodOptions } from '../../utils/fieldOptions'
import {
  getOperationLogList,
  getOperationLogDetail,
  deleteOperationLog,
  batchDeleteOperationLogs,
  cleanOperationLogs,
  getOperationLogTitleOptions
} from '../../api/log'

const { t, te, tm } = useI18n()
const { getButtonState } = usePermission()

const tableRef = ref(null)
const detailVisible = ref(false)
const logDetail = ref(null)
const selectedRows = ref([])

// 预置的操作标题（权限标识），用于下拉选项，即使还没有对应的操作日志也能选择
// 对应多语言中的 permission.* 配置
const defaultTitleSlugs = [
  // 管理员
  'admin.store',
  'admin.update',
  'admin.destroy',
  'admin.export',
  'admin.password',
  'admin.kick_out',
  'admin.unbind_google_auth',
  // 角色
  'role.store',
  'role.update',
  'role.destroy',
  // 权限
  'permission.store',
  'permission.update',
  'permission.destroy',
  // 菜单
  'menu.store',
  'menu.update',
  'menu.destroy',
  // 部门
  'department.store',
  'department.update',
  'department.destroy',
  // 字典
  'dictionary.store',
  'dictionary.update',
  'dictionary.destroy',
  'dictionary.type',
  // 操作日志
  'operation_log.destroy',
  'operation_log.batch_delete',
  'operation_log.clean',
  // 登录日志
  'login_log.destroy',
  'login_log.batch_delete',
  'login_log.clean',
  // 系统日志
  'system_log.destroy',
  'system_log.batch_delete',
  'system_log.clean',
  // 个人中心
  'profile.update',
  'password.update'
]

const titleOptions = ref([])

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'method': 'method',
  'path': 'path',
  'ip': 'ip',
  'status_code': 'status', // 前端使用 status_code，数据库字段是 status
  'created_at': 'created_at'
}

// 格式化日期为 YYYY-MM-DD HH:mm:ss
const formatDateTime = (date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 获取7天前的日期时间
const getSevenDaysAgo = () => {
  const date = new Date()
  date.setDate(date.getDate() - 7)
  date.setHours(0, 0, 0, 0) // 设置为当天的00:00:00
  return formatDateTime(date)
}

// 初始搜索表单数据
const initialSearchForm = {
  username: '',
  method: '',
  path: '',
  title: '',
  ip: '',
  status: '',
  start_time: getSevenDaysAgo(),
  end_time: ''
}

// 转换操作日志数据（PascalCase -> snake_case）
const transformOperationLogData = (log) => {
  let params = null
  try {
    if (log.Request) {
      params = typeof log.Request === 'string' ? JSON.parse(log.Request) : log.Request
    } else if (log.request) {
      params = typeof log.request === 'string' ? JSON.parse(log.request) : log.request
    } else if (log.Params) {
      params = typeof log.Params === 'string' ? JSON.parse(log.Params) : log.Params
    } else if (log.params) {
      params = typeof log.params === 'string' ? JSON.parse(log.params) : log.params
    }
  } catch (e) {
    params = log.Request || log.request || log.Params || log.params || null
  }
  
  return {
    id: log.ID || log.id,
    admin: log.Admin ? {
      username: log.Admin.Username || log.Admin.username || ''
    } : (log.admin ? {
      username: log.admin.username || ''
    } : null),
    method: log.Method || log.method || '',
    path: log.Path || log.path || '',
    title: log.Title || log.title || '',
    ip: log.IP || log.ip || '',
    status_code: log.Status || log.status || log.StatusCode || log.status_code || 0,
    created_at: log.CreatedAt || log.created_at || '',
    params: params,
    request: log.Request || log.request || null,
    response: log.Response || log.response || null
  }
}

// 使用列表页面 composable
const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handlePageChange,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getOperationLogList,
  initialSearchForm,
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  },
  transformData: transformOperationLogData
})

// 将复数形式转换为单数形式，以匹配权限配置中的 slug
// 支持处理：复数形式（roles）、连字符形式（operation-logs）、下划线形式（operation_logs）
const pluralToSingular = (plural) => {
  if (!plural) return plural

  // 先处理连字符形式，转换为下划线形式统一处理
  let normalized = plural.replace(/-/g, '_')

  // 常见的复数到单数映射（使用下划线形式）
  const singularMap = {
    'roles': 'role',
    'permissions': 'permission',
    'menus': 'menu',
    'departments': 'department',
    'dictionaries': 'dictionary',
    'blacklists': 'blacklist',
    'admins': 'admin',
    'operation_logs': 'operation_log',
    'login_logs': 'login_log',
    'system_logs': 'system_log',
    'online_users': 'online-user',
    // 连字符形式也支持
    'operation-logs': 'operation_log',
    'login-logs': 'login_log',
    'system-logs': 'system_log',
    'online-users': 'online-user'
  }

  // 先检查完整匹配
  if (singularMap[plural]) {
    return singularMap[plural]
  }
  if (singularMap[normalized]) {
    return singularMap[normalized]
  }

  // 处理复合词（包含下划线的情况，如 operation_logs）
  if (normalized.includes('_')) {
    const parts = normalized.split('_')
    const lastPart = parts[parts.length - 1]
    // 只转换最后一个部分
    const singularLastPart = convertPluralToSingular(lastPart)
    if (singularLastPart !== lastPart) {
      return parts.slice(0, -1).join('_') + '_' + singularLastPart
    }
  }

  // 如果没有找到映射，尝试常见的复数规则
  return convertPluralToSingular(normalized)
}

// 基础的复数转单数转换函数
const convertPluralToSingular = (word) => {
  if (!word || word.length <= 1) return word

  // 以 -s 结尾的单词，去掉 s
  if (word.endsWith('s')) {
    // 特殊情况：-ies 结尾的单词（如 dictionaries -> dictionary）
    if (word.endsWith('ies') && word.length > 3) {
      return word.slice(0, -3) + 'y'
    }
    // 特殊情况：-es 结尾的单词（如 permissions -> permission）
    if (word.endsWith('es') && word.length > 2) {
      const beforeEs = word.slice(0, -2)
      // 如果去掉 es 后以 ch, sh, x, s, z 结尾，保留 e
      const lastChar = beforeEs[beforeEs.length - 1]
      if (['c', 's', 'x', 'z'].includes(lastChar)) {
        return beforeEs
      }
      return beforeEs
    }
    return word.slice(0, -1)
  }

  // 默认返回原值
  return word
}

// 获取操作标题的翻译
// 标题可能存储为：
// 1. 权限标识（单数形式）：admin.update, role.update（优先使用，由权限中间件设置）
// 2. 路径生成的标题（可能是复数形式）：admins.update, roles.update（当没有权限标识时）
// 3. 连字符或下划线形式：operation-logs.update 或 operation_logs.update
// 前端统一处理：将复数形式转换为单数形式，然后查找翻译
const getOperationTitle = (titleKey) => {
  if (!titleKey) return '-'

  // 0. 先尝试将复数形式转换为单数形式
  // 如果包含点号，说明是 module.action 格式，需要转换 module 部分
  let slug = titleKey
  if (slug.includes('.')) {
    const parts = slug.split('.')
    if (parts.length >= 2) {
      const module = pluralToSingular(parts[0])
      slug = module + '.' + parts.slice(1).join('.')
    }
  } else {
    // 如果没有点号，直接转换整个字符串
    slug = pluralToSingular(slug)
  }

  // 1. 作为权限标识翻译：permission.admin.update 这种形式
  const slugKey = `permission.${slug}`

  // 1.1 使用 te 检测路径是否存在（兼容嵌套路径）
  if (typeof te === 'function' && te(slugKey)) {
    return t(slugKey)
  }

  // 1.2 直接从 permission 命名空间对象里取（兼容平铺的 "admin.update" 键）
  const messages = typeof tm === 'function' ? tm('permission') : null
  if (messages && Object.prototype.hasOwnProperty.call(messages, slug)) {
    const value = messages[slug]
    if (typeof value === 'string') {
      return value
    }
  }

  // 2. 兼容旧的 operation.xxx key（如果还有残留数据）
  if (titleKey.startsWith('operation.')) {
    const translated = t(titleKey)
    if (translated !== titleKey) {
      return translated
    }
  }

  // 3. 如果转换后的 slug 和原始 titleKey 不同，再尝试用原始值查找一次（兼容旧数据）
  if (slug !== titleKey) {
    const originalSlugKey = `permission.${titleKey}`
    if (typeof te === 'function' && te(originalSlugKey)) {
      return t(originalSlugKey)
    }
    const originalMessages = typeof tm === 'function' ? tm('permission') : null
    if (originalMessages && Object.prototype.hasOwnProperty.call(originalMessages, titleKey)) {
      const value = originalMessages[titleKey]
      if (typeof value === 'string') {
        return value
      }
    }
  }

  // 4. 找不到翻译就原样返回
  return titleKey
}

// 表格列配置（使用 vxe-table columns）
const tableColumns = computed(() => [
  {
    type: 'checkbox',
    width: 60
  },
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },
  {
    field: 'admin',
    title: t('log.admin'),
    slot: 'admin',
    sortable: false
  },
  {
    field: 'title',
    title: t('log.title'),
    slot: 'title',
    sortable: true,
    width: 200
  },
  {
    field: 'method',
    title: t('log.method'),
    width: 100,
    sortable: true
  },
  {
    field: 'path',
    title: t('log.path'),
    sortable: true
  },
  {
    field: 'ip',
    title: t('log.ip'),
    width: 150,
    sortable: true
  },
  {
    field: 'status_code',
    title: t('log.status'),
    width: 100,
    sortable: true,
    formatter: ({ row }) => {
      const v = row.status_code
      if (v === 1 || v === '1') {
        return t('log.success')
      }
      if (v === 0 || v === '0') {
        return t('log.failed')
      }
      return v ?? '-'
    }
  },
  {
    field: 'created_at',
    title: t('log.operation_time'),
    width: 180,
    sortable: true
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation'
  }
])

// 搜索表单字段配置（JSON 方式）
const searchFields = computed(() => {
  // 构建标题选项，添加空选项
  const titleSelectOptions = [
    {
      label: t('common.all'),
      value: ''
    },
    ...titleOptions.value.map(title => ({
      label: getOperationTitle(title),
      value: title
    }))
  ]

  return [
    {
      prop: 'username',
      label: t('log.username'),
      type: 'input',
      width: '200px',
      advanced: false
    },
    {
      prop: 'method',
      label: t('log.method'),
      type: 'select',
      width: '150px',
      options: getMethodOptions().filter(opt => String(opt.value).toUpperCase() !== 'GET'),
      advanced: false
    },
    {
      prop: 'path',
      label: t('log.path'),
      type: 'input',
      width: '200px',
      advanced: false
    },
    {
      prop: 'title',
      label: t('log.title'),
      type: 'select',
      width: '200px',
      options: titleSelectOptions,
      filterable: true,
      clearable: true,
      advanced: false
    },
    {
      prop: 'ip',
      label: t('log.ip'),
      type: 'input',
      width: '150px',
      advanced: true
    },
    {
      prop: 'status',
      label: t('log.status'),
      type: 'select',
      width: '150px',
      options: [
        { label: t('log.success'), value: '1' },
        { label: t('log.failed'), value: '0' }
      ],
      clearable: true,
      advanced: true
    },
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
})

// loadData, handleSearch, handleReset, handlePageChange 已由 useListPage 提供

const handleView = async (row) => {
  try {
    const res = await getOperationLogDetail(row.id)
    if (res.data) {
      const log = res.data.operation_log || res.data.log || res.data
      logDetail.value = transformOperationLogData(log)
      detailVisible.value = true
    }
  } catch (error) {
    console.error('Load operation log detail error:', error)
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(t('log.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteOperationLog(row.id)
    ElMessage.success(t('log.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

const handleSelectionChange = () => {
  selectedRows.value = tableRef.value?.getCheckboxRecords() || []
}

const handleBatchDelete = async () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning(t('common.please_select_items'))
    return
  }

  try {
    await ElMessageBox.confirm(t('log.batch_delete_confirm', { count: selectedRows.value.length }), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    const ids = selectedRows.value.map(row => row.id)
    await batchDeleteOperationLogs(ids)
    ElMessage.success(t('log.delete_success'))
    selectedRows.value = []
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Batch delete error:', error)
    }
  }
}

const handleClean = async () => {
  try {
    await ElMessageBox.confirm(t('log.clean_confirm'), t('form.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await cleanOperationLogs()
    ElMessage.success(t('log.clean_success'))
    selectedRows.value = []
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Clean error:', error)
    }
  }
}

// 加载标题选项
const loadTitleOptions = async () => {
  try {
    const res = await getOperationLogTitleOptions()
    // 使用 Set 合并后端返回的标题和预置的权限标识
    const mergedSet = new Set(defaultTitleSlugs)

    if (res.data && res.data.titles && Array.isArray(res.data.titles)) {
      res.data.titles.forEach(title => {
        if (title && typeof title === 'string') {
          const trimmed = title.trim()
          if (trimmed && trimmed !== 'operation.unknown' && !trimmed.startsWith('operation.')) {
            mergedSet.add(trimmed)
          }
        }
      })
    }

    // 转成数组并排序（按翻译后的文本）
    const uniqueTitles = Array.from(mergedSet)
    uniqueTitles.sort((a, b) => {
      const labelA = getOperationTitle(a)
      const labelB = getOperationTitle(b)
      const locale = t('common.locale') || navigator.language || 'zh-CN'
      return labelA.localeCompare(labelB, locale)
    })

    titleOptions.value = uniqueTitles
  } catch (error) {
    console.error('Load title options error:', error)
    // 如果加载失败，至少使用预置的权限标识
    const uniqueTitles = Array.from(new Set(defaultTitleSlugs))
    uniqueTitles.sort((a, b) => {
      const labelA = getOperationTitle(a)
      const labelB = getOperationTitle(b)
      const locale = t('common.locale') || navigator.language || 'zh-CN'
      return labelA.localeCompare(labelB, locale)
    })
    titleOptions.value = uniqueTitles
  }
}

onMounted(() => {
  // 初始化默认排序
  initDefaultSort()
  // 加载标题选项
  loadTitleOptions()
  loadData()
})
</script>

<style scoped>
.log-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}


pre {
  margin: 0;
  padding: 10px;
  background: #f5f5f5;
  border-radius: 4px;
  overflow-x: auto;
}
</style>

