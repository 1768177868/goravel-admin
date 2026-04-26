<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.operation_log') }}</span>
          <div class="header-actions">
            <el-button 
              type="danger" 
              :disabled="selectedRows.length === 0 || getButtonState('operation_log.batch_delete').disabled"
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

      <!-- 表格工具栏 -->
      <TableToolbar
        :on-refresh="handleRefresh"
        fullscreen-target=".list-page"
        :visible-columns="visibleColumns"
        :all-columns="allTableColumns"
        :default-visible-columns="defaultVisibleColumns"
        :column-order="columnOrder"
        :fixed-columns="fixedColumns"
        :on-column-setting-confirm="handleColumnSettingConfirm"
      />

      <VxeTable
        ref="tableRef"
        :key="`table-${tableColumns.length}-${JSON.stringify(columnOrder)}`"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="600"
        @sort-change="handleSortChange"
        @checkbox-change="handleSelectionChange"
        @checkbox-all="handleSelectionChange"
      >
        <template #admin="{ row }">
          {{ row.admin?.username || '-' }}
        </template>

        <template #title="{ row }">
          {{ getOperationTitle(row.title || row.Title) }}
        </template>

        <template #request="{ row }">
          <div class="request-params-cell">
            <el-popover
              v-if="hasRequestParams(row.request || row.Request)"
              placement="top"
              :width="600"
              trigger="hover"
              :title="$t('log.request_params')"
            >
              <template #reference>
                <div class="request-preview">
                  <span class="preview-text">{{ getRequestPreview(row.request || row.Request) }}</span>
                </div>
              </template>
              <div class="request-params-popover">
                <div class="request-params-header">
                  <el-button 
                    type="primary" 
                    link 
                    size="small"
                    @click="copyRequestParamsToClipboard(row.request || row.Request)"
                  >
                    <el-icon><DocumentCopy /></el-icon>
                    {{ $t('common.copy') }}
                  </el-button>
                </div>
                <pre class="request-params-content">{{ formatRequestParamsFull(row.request || row.Request) }}</pre>
              </div>
            </el-popover>
            <span v-else class="text-muted">-</span>
          </div>
        </template>

        <template #operation="{ row }">
          <TableActionButtons
            :row="row"
            :primary-actions="operationActions"
            :get-button-state="getButtonState"
          />
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
        :show-total="true"
        :show-quick-jumper="true"
        :align="'right'"
      />
    </el-card>

    <el-dialog v-model="detailVisible" :title="$t('log.detail')" width="1100px">
      <el-descriptions :column="2" border v-if="logDetail">
        <el-descriptions-item :label="$t('table.id')">{{ logDetail.id }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.admin')">{{ logDetail.admin?.username }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.method')">{{ logDetail.method }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.path')">{{ logDetail.path }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.ip')">{{ logDetail.ip }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.status_code')">{{ logDetail.status_code }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.operation_time')" :span="2">{{ logDetail.created_at }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.request_params')" :span="2">
          <div class="request-params-detail">
            <div class="request-params-header">
              <el-button 
                type="primary" 
                link 
                size="small"
                @click="copyRequestParams"
              >
                <el-icon><DocumentCopy /></el-icon>
                {{ $t('common.copy') }}
              </el-button>
            </div>
            <pre ref="requestParamsPre" class="request-params-content">{{ formatRequestParamsFull(logDetail.params || logDetail.request || {}) }}</pre>
          </div>
        </el-descriptions-item>
        <el-descriptions-item v-if="logDetail.changes && logDetail.changes.length > 0" :label="$t('log.changes')" :span="2">
          <div class="changes-detail">
            <el-table :data="logDetail.changes" border size="small" class="changes-table">
              <el-table-column :label="$t('log.changes_field')" prop="field" width="180" />
              <el-table-column :label="$t('log.changes_old')">
                <template #default="{ row }">
                  <span class="changes-value changes-old">{{ formatChangeValue(row.old) }}</span>
                </template>
              </el-table-column>
              <el-table-column :label="$t('log.changes_new')">
                <template #default="{ row }">
                  <span class="changes-value changes-new">{{ formatChangeValue(row.new) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, DocumentCopy } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import TableToolbar from '../../components/TableToolbar.vue'
import { useListPage } from '../../composables/useListPage'
import { useCrud } from '../../composables/useCrud'
import { usePermission } from '../../composables/usePermission'
import { useColumnSetting } from '../../composables/useColumnSetting'
import { getMethodOptions } from '../../utils/fieldOptions'
import { getSevenDaysAgo } from '../../utils/dateUtils'
import { validateTimeRange, OPERATION_LOG_MAX_TIME_RANGE_MONTHS } from '../../utils/timeRangeValidator'
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
// 维护跨页选中的ID集合
const selectedIds = ref(new Set())

// 使用 CRUD composable（删除和批量删除）
const { handleDelete: handleDeleteCrud, handleBatchDelete: handleBatchDeleteCrud } = useCrud({
  deleteApi: deleteOperationLog,
  batchDeleteApi: batchDeleteOperationLogs
})

const titleOptions = ref([])

// 字段名映射：前端字段名 -> 数据库字段名（只包含不同的字段）
const fieldMapping = {
  'status_code': 'status' // 前端使用 status_code，数据库字段是 status
}

// 使用共用的日期工具函数（已从 utils/dateUtils 导入）

// 初始搜索表单数据
const initialSearchForm = {
  username: '',
  method: '',
  path: '',
  title: '',
  ip: '',
  status: '',
  request: '',
  start_time: getSevenDaysAgo(),
  end_time: ''
}

// 转换操作日志数据（以 snake_case 为主）
const transformOperationLogData = (log) => {
  let params = null
  try {
    if (log.request) {
      params = typeof log.request === 'string' ? JSON.parse(log.request) : log.request
    } else if (log.params) {
      params = typeof log.params === 'string' ? JSON.parse(log.params) : log.params
    }
  } catch (e) {
    params = log.request || log.params || null
  }
  
  let changes = null
  try {
    const raw = log.changes
    if (raw) {
      changes = typeof raw === 'string' ? JSON.parse(raw) : raw
    }
  } catch (e) {
    changes = null
  }

  return {
    id: log.id,
    admin: log.admin ? {
      username: log.admin.username || ''
    } : null,
    method: log.method || '',
    path: log.path || '',
    title: log.title || '',
    ip: log.ip || '',
    status_code: log.status_code ?? log.status ?? 0,
    created_at: log.created_at || '',
    params: params,
    request: log.request || null,
    response: log.response || null,
    changes: changes
  }
}

const requestParamsPre = ref(null)

// 检查是否有请求参数
const hasRequestParams = (request) => {
  if (!request) return false
  try {
    const parsed = typeof request === 'string' ? JSON.parse(request) : request
    if (typeof parsed === 'object' && parsed !== null) {
      return Object.keys(parsed).length > 0
    }
    return String(parsed).trim().length > 0
  } catch (e) {
    return String(request).trim().length > 0
  }
}

// 获取请求参数的简洁预览（用于表格显示）
const getRequestPreview = (request) => {
  if (!request) return '-'
  
  try {
    const parsed = typeof request === 'string' ? JSON.parse(request) : request
    if (typeof parsed === 'object' && parsed !== null) {
      const keys = Object.keys(parsed)
      if (keys.length === 0) return '-'
      // 显示前3个字段名
      const previewKeys = keys.slice(0, 3)
      const preview = previewKeys.map(key => {
        const value = parsed[key]
        if (typeof value === 'object' && value !== null) {
          return `${key}: {...}`
        }
        const valueStr = String(value)
        return `${key}: ${valueStr.length > 20 ? valueStr.substring(0, 20) + '...' : valueStr}`
      }).join(', ')
      return keys.length > 3 ? `${preview} ... (${keys.length} fields)` : preview
    }
    const str = String(parsed)
    return str.length > 50 ? str.substring(0, 50) + '...' : str
  } catch (e) {
    const str = String(request)
    return str.length > 50 ? str.substring(0, 50) + '...' : str
  }
}

// 复制请求参数到剪贴板
const copyRequestParamsToClipboard = async (request) => {
  try {
    const content = formatRequestParamsFull(request)
    await navigator.clipboard.writeText(content)
    ElMessage.success(t('common.copy_success') || '复制成功')
  } catch (err) {
    console.error('Failed to copy: ', err)
    ElMessage.error(t('common.copy_failed') || '复制失败')
  }
}

// 格式化请求参数完整内容（用于详情显示）
const formatRequestParamsFull = (request) => {
  if (!request) return '-'
  
  try {
    const parsed = typeof request === 'string' ? JSON.parse(request) : request
    if (typeof parsed === 'object' && parsed !== null) {
      return JSON.stringify(parsed, null, 2)
    }
    return String(parsed)
  } catch (e) {
    return String(request)
  }
}

// 格式化变更值（用于 diff 表格显示）
const formatChangeValue = (value) => {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'object') return JSON.stringify(value, null, 2)
  return String(value)
}

// 复制请求参数
const copyRequestParams = async () => {
  if (!requestParamsPre.value) return
  
  const text = requestParamsPre.value.textContent
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(t('common.copy_success') || '复制成功')
  } catch (err) {
    // 降级方案
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    try {
      document.execCommand('copy')
      ElMessage.success(t('common.copy_success') || '复制成功')
    } catch (e) {
      ElMessage.error(t('common.copy_failed') || '复制失败')
    }
    document.body.removeChild(textarea)
  }
}

// 使用列表页面 composable
const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch: baseHandleSearch,
  handleReset: baseHandleReset,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getOperationLogList,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef),
  transformData: transformOperationLogData,
  onSearch: () => {
    // 搜索前清除选中状态
    selectedRows.value = []
    selectedIds.value.clear()
  },
  onReset: () => {
    // 重置前清除选中状态
    selectedRows.value = []
    selectedIds.value.clear()
  },
  onLoadSuccess: () => {
    // 数据加载后，恢复选中状态
    nextTick(() => {
      if (tableRef.value?.tableRef && selectedIds.value.size > 0) {
        const rowsToSelect = tableData.value.filter(row => selectedIds.value.has(row.id))
        rowsToSelect.forEach(row => {
          tableRef.value.tableRef.setCheckboxRow(row, true)
        })
        // 更新 selectedRows
        selectedRows.value = tableRef.value.tableRef.getCheckboxRecords() || []
      }
    })
  }
})

// 获取当前时间的字符串格式（YYYY-MM-DD HH:mm:ss）
const getCurrentTimeString = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  const seconds = String(now.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 验证时间范围的函数
const validateTimeRangeForSearch = () => {
  // 如果只填了开始时间，结束时间默认为当前时间
  if (searchForm.start_time) {
    const endTime = searchForm.end_time || getCurrentTimeString()
    const validation = validateTimeRange(searchForm.start_time, endTime, OPERATION_LOG_MAX_TIME_RANGE_MONTHS)
    if (!validation.valid) {
      // 优先使用翻译键
      let errorMessage = validation.error
      if (validation.errorKey) {
        const translationKey = `log.${validation.errorKey}`
        if (validation.errorParams) {
          errorMessage = t(translationKey, validation.errorParams)
        } else {
          errorMessage = t(translationKey)
        }
      }
      ElMessage.warning(errorMessage)
      return false
    }
  }
  return true
}

// 监听开始时间和结束时间的变化，实时验证
const isInitialized = ref(false)
watch(
  () => [searchForm.start_time, searchForm.end_time],
  ([newStartTime, newEndTime], [oldStartTime, oldEndTime]) => {
    // 跳过初始化时的触发
    if (!isInitialized.value) {
      isInitialized.value = true
      return
    }
    
    // 只在开始时间或结束时间发生变化时验证
    if (newStartTime !== oldStartTime || newEndTime !== oldEndTime) {
      // 如果只填了开始时间，结束时间默认为当前时间
      if (newStartTime) {
        const endTime = newEndTime || getCurrentTimeString()
        const validation = validateTimeRange(newStartTime, endTime, OPERATION_LOG_MAX_TIME_RANGE_MONTHS)
        if (!validation.valid) {
          // 优先使用翻译键
          let errorMessage = validation.error
          if (validation.errorKey) {
            const translationKey = `log.${validation.errorKey}`
            if (validation.errorParams) {
              errorMessage = t(translationKey, validation.errorParams)
            } else {
              errorMessage = t(translationKey)
            }
          }
          ElMessage.warning(errorMessage)
        }
      }
    }
  },
  { immediate: false }
)

// 包装搜索处理，添加时间范围验证
const handleSearch = () => {
  // 验证时间范围
  if (!validateTimeRangeForSearch()) {
    return
  }
  // 调用 baseHandleSearch（onSearch 回调已处理清除选中状态）
  baseHandleSearch()
}

// 使用 baseHandleReset（onReset 回调已处理清除选中状态）
const handleReset = baseHandleReset

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
    'users': 'user',
    'operation_logs': 'operation_log',
    'login_logs': 'login_log',
    'system_logs': 'system_log',
    'online_admins': 'online-admin',
    // 连字符形式也支持
    'operation-logs': 'operation_log',
    'login-logs': 'login_log',
    'system-logs': 'system_log',
    'online-admins': 'online-admin',
    'user-balance-logs': 'user_balance_log'
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

  // 兼容历史 pprof.* 标题，统一映射到 permission 命名空间里的 observability.* 键
  const legacyTitleMap = {
    'pprof.verify': 'observability.pprof_verify',
    'pprof.cpu_hotspots': 'observability.pprof_cpu_hotspots',
    'pprof.memory_hotspots': 'observability.pprof_memory_hotspots',
    'online-admin.kick-out': 'online_admin.kick_out',
    'online-admin.batch-kick-out': 'online_admin.batch_kick_out'
  }
  const normalizedTitleKey = legacyTitleMap[titleKey] || titleKey

  // 0. 先尝试将复数形式转换为单数形式
  // 如果包含点号，说明是 module.action 格式，需要转换 module 部分
  let slug = normalizedTitleKey
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
  if (normalizedTitleKey.startsWith('operation.')) {
    const translated = t(normalizedTitleKey)
    if (translated !== normalizedTitleKey) {
      return translated
    }
  }

  // 3. 如果转换后的 slug 和原始 titleKey 不同，再尝试用原始值查找一次（兼容旧数据）
  if (slug !== normalizedTitleKey) {
    const originalSlugKey = `permission.${normalizedTitleKey}`
    if (typeof te === 'function' && te(originalSlugKey)) {
      return t(originalSlugKey)
    }
    const originalMessages = typeof tm === 'function' ? tm('permission') : null
    if (originalMessages && Object.prototype.hasOwnProperty.call(originalMessages, normalizedTitleKey)) {
      const value = originalMessages[normalizedTitleKey]
      if (typeof value === 'string') {
        return value
      }
    }
  }

  // 4. 找不到翻译就原样返回
  return normalizedTitleKey
}

// 表格列配置（使用 vxe-table columns）
const allTableColumns = computed(() => [
  {
    type: 'checkbox',
    width: 60,
    key: 'checkbox'
  },
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true,
    key: 'id'
  },
  {
    field: 'admin',
    title: t('log.admin'),
    slot: 'admin',
    sortable: false,
    key: 'admin',
    width: 120,
  },
  {
    field: 'title',
    title: t('log.title'),
    slot: 'title',
    sortable: false,
    width: 200,
    key: 'title'
  },
  {
    field: 'method',
    title: t('log.method'),
    width: 100,
    sortable: false,
    key: 'method'
  },
  {
    field: 'path',
    title: t('log.path'),
    sortable: false,
    key: 'path'
  },
  {
    field: 'ip',
    title: t('log.ip'),
    width: 150,
    sortable: false,
    key: 'ip'
  },
  // {
  //   field: 'status_code',
  //   title: t('log.status'),
  //   width: 100,
  //   sortable: false,
  //   formatter: ({ row }) => {
  //     const v = row.status_code
  //     if (v === 1 || v === '1') {
  //       return t('log.success')
  //     }
  //     if (v === 0 || v === '0') {
  //       return t('log.failed')
  //     }
  //     return v ?? '-'
  //   },
  //   key: 'status_code'
  // },
  {
    field: 'request',
    title: t('log.request_params'),
    slot: 'request',
    sortable: false,
    width: 250,
    showOverflow: 'tooltip',
    key: 'request'
  },
  {
    field: 'created_at',
    title: t('log.operation_time'),
    width: 180,
    sortable: true,
    key: 'created_at'
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation',
    key: 'operation'
  }
])

// 使用列设置 composable
const {
  tableColumns,
  visibleColumns,
  allColumns,
  defaultVisibleColumns,
  columnOrder,
  fixedColumns,
  handleColumnSettingConfirm
} = useColumnSetting('operation_log', allTableColumns)

// 处理刷新
const handleRefresh = () => {
  loadData()
}

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
      prop: 'request',
      label: t('log.request_params'),
      type: 'input',
      width: '200px',
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

const handleDelete = (row) => handleDeleteCrud(row, loadData)

// 操作按钮配置
const operationActions = computed(() => [
  {
    key: 'view',
    label: t('common.view'),
    type: 'primary',
    permission: 'operation_log.show',
    handler: handleView
  },
  {
    key: 'delete',
    label: t('common.delete'),
    type: 'danger',
    permission: 'operation_log.destroy',
    handler: handleDelete
  }
])

const handleSelectionChange = () => {
  // 使用 vxe-table 的 getCheckboxRecords 方法获取选中的行
  if (tableRef.value?.tableRef) {
    const currentSelected = tableRef.value.tableRef.getCheckboxRecords() || []
    selectedRows.value = currentSelected
    
    // 更新选中ID集合：先移除当前页的所有ID，再添加当前选中的ID
    tableData.value.forEach(row => {
      selectedIds.value.delete(row.id)
    })
    currentSelected.forEach(row => {
      selectedIds.value.add(row.id)
    })
  }
}

const handleBatchDelete = () => {
  handleBatchDeleteCrud(selectedRows.value, () => {
    // 清除选中状态
    selectedRows.value = []
    selectedIds.value.clear()
    loadData()
  })
}

// 加载标题选项
const loadTitleOptions = async () => {
  try {
    const res = await getOperationLogTitleOptions()
    const mergedSet = new Set()
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
    titleOptions.value = []
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

pre {
  margin: 0;
  padding: var(--space-sm);
  background: var(--bg-color-tertiary);
  border-radius: var(--border-radius-sm);
  overflow-x: auto;
}

.request-params-cell {
  display: flex;
  align-items: center;
  width: 100%;
}

.request-preview {
  width: 100%;
  cursor: pointer;
}

.preview-text {
  font-size: 12px;
  color: var(--text-color-regular);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro', monospace;
  display: block;
}

.text-muted {
  color: var(--text-color-secondary);
}

.request-params-popover {
  max-height: 500px;
  overflow: hidden;
}

.request-params-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: var(--space-xs);
  padding-bottom: var(--space-xs);
  border-bottom: 1px solid var(--border-color-lighter);
}

.request-params-content {
  margin: 0;
  padding: var(--space-sm);
  background: var(--bg-color-tertiary);
  border-radius: var(--border-radius-sm);
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.6;
  max-height: 400px;
  overflow-y: auto;
  word-break: break-all;
  white-space: pre-wrap;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro', monospace;
}

.request-params-detail {
  width: 100%;
}

.request-params-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: var(--space-xs);
}

.request-params-content {
  margin: 0;
  padding: var(--space-sm);
  background: var(--bg-color-tertiary);
  border-radius: var(--border-radius-sm);
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.6;
  max-height: 400px;
  overflow-y: auto;
  word-break: break-all;
  white-space: pre-wrap;
}

.changes-detail {
  width: 100%;
}

.changes-table {
  width: 100%;
}

.changes-value {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro', monospace;
  font-size: 12px;
  word-break: break-all;
  white-space: pre-wrap;
}

.changes-old {
  color: var(--el-color-danger);
}

.changes-new {
  color: var(--el-color-success);
}

/* 暗黑模式样式 */
html.dark pre {
  background: var(--el-bg-color) !important;
  color: var(--el-text-color-regular) !important;
  border: 1px solid var(--el-border-color);
}

html.dark .request-params-content {
  background: var(--el-bg-color) !important;
  color: var(--el-text-color-regular) !important;
  border: 1px solid var(--el-border-color);
}

html.dark .request-params-header {
  border-bottom-color: var(--el-border-color) !important;
}

html.dark .preview-text {
  color: var(--el-text-color-regular) !important;
}

html.dark .text-muted {
  color: var(--el-text-color-secondary) !important;
}
</style>

