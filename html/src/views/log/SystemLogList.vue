<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.system_log') }}</span>
          <div class="header-actions">
            <el-button 
              type="danger" 
              :disabled="!selectedRows || selectedRows.length === 0 || getButtonState('system_log.batch_delete').disabled"
              @click="handleBatchDelete"
            >
              <el-icon><Delete /></el-icon>
              {{ $t('common.delete_selected') }} ({{ selectedRows?.length || 0 }})
            </el-button>
          </div>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="log"
        @search="handleSearch"
        @reset="handleReset"
      />

      <VxeTable
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="600"
        @sort-change="handleSortChange"
        @checkbox-change="handleSelectionChange"
        @checkbox-all="handleSelectionChange"
      >
        <template #level="{ row }">
          <el-tag :type="getLevelType(row.level)">
            {{ getLevelLabel(row.level) }}
          </el-tag>
        </template>

        <template #context="{ row }">
          <el-tooltip
            v-if="row.context"
            :content="formatContext(row.context)"
            placement="top"
            effect="dark"
          >
            <div class="context-preview">
              {{ formatContextPreview(row.context) }}
            </div>
          </el-tooltip>
          <span v-else>-</span>
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
      />
    </el-card>

    <el-dialog v-model="detailVisible" :title="$t('log.detail')" width="800px">
      <el-descriptions :column="2" border v-if="logDetail">
        <el-descriptions-item :label="$t('table.id')">{{ logDetail.id }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.level')">
          <el-tag :type="getLevelType(logDetail.level)">
            {{ getLevelLabel(logDetail.level) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('log.module')">
          {{ getModuleLabel(logDetail.module) }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('log.trace_id')" :span="2">
          {{ logDetail.trace_id || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('log.message')" :span="2">{{ logDetail.message }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.context')" :span="2">
          <pre v-if="logDetail.context">{{ formatContext(logDetail.context) }}</pre>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('log.time')" :span="2">{{ logDetail.created_at }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import {
  getSystemLogList,
  getSystemLogDetail,
  deleteSystemLog,
  batchDeleteSystemLogs,
  cleanSystemLogs
} from '../../api/log'

const { t } = useI18n()
const { getButtonState } = usePermission()

// 使用 CRUD composable（删除和批量删除）
const { handleDelete: handleDeleteCrud, handleBatchDelete: handleBatchDeleteCrud } = useCrud({
  deleteApi: deleteSystemLog,
  batchDeleteApi: batchDeleteSystemLogs
})

const tableRef = ref(null)
const detailVisible = ref(false)
const logDetail = ref(null)
const selectedRows = ref([])
// 维护跨页选中的ID集合
const selectedIds = ref(new Set())

// 初始搜索表单
const initialSearchForm = {
  level: '',
  module: '',
  trace_id: '',
  message: '',
  start_time: '',
  end_time: ''
}

// 转换系统日志数据（PascalCase -> snake_case）
// 必须在 useListPage 之前定义，因为 useListPage 会使用它
const transformSystemLogData = (log) => {
  if (!log) {
    return {
      id: 0,
      level: '',
      trace_id: '',
      message: '',
      context: null,
      created_at: ''
    }
  }
  
  let context = null
  try {
    if (log.Context) {
      context = typeof log.Context === 'string' ? JSON.parse(log.Context) : log.Context
    } else if (log.context) {
      context = typeof log.context === 'string' ? JSON.parse(log.context) : log.context
    }
  } catch (e) {
    context = log.Context || log.context || null
  }
  
  return {
    id: log.ID || log.id || 0,
    level: log.Level || log.level || '',
    module: log.Module || log.module || '',
    trace_id: log.TraceID || log.trace_id || '',
    message: log.Message || log.message || '',
    context: context,
    created_at: log.CreatedAt || log.created_at || ''
  }
}

const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getSystemLogList,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef),
  transformData: transformSystemLogData,
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
    field: 'level',
    title: t('log.level'),
    width: 100,
    sortable: false,
    slot: 'level'
  },
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
    formatter: ({ row }) => row.TraceID || row.trace_id || '-'
  },
  {
    field: 'message',
    title: t('log.message'),
    sortable: false
  },
  {
    field: 'context',
    title: t('log.context'),
    width: 200,
    slot: 'context',
    sortable: false
  },
  {
    field: 'created_at',
    title: t('log.time'),
    width: 180,
    sortable: true
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation',
    sortable: false
  }
])

// 搜索表单字段配置
const searchFields = computed(() => [
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
    options: getModuleOptions(t),
    filterable: true,
    clearable: true,
    allowCreate: true,
    advanced: false
  },
  {
    prop: 'trace_id',
    label: t('log.trace_id'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'message',
    label: t('log.message'),
    type: 'input',
    width: '200px',
    advanced: false
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
])

// 获取模块选项（带多语言）
const getModuleOptions = (t) => {
  return [
    { label: t('menu.system_log'), value: 'system-log' },
    { label: t('menu.attachment'), value: 'attachment' },
    { label: t('log.module_auth'), value: 'auth' },
    { label: t('menu.monitor'), value: 'monitor' },
    { label: t('menu.operation_log'), value: 'operation-log' },
    { label: t('menu.login_log'), value: 'login-log' },
    { label: t('log.module_recover'), value: 'recover' },
    { label: t('log.module_payment'), value: 'payment' },
    { label: t('menu.payment_method'), value: 'payment_method' },
    { label: t('menu.order'), value: 'order' },
    { label: t('menu.export'), value: 'export' },
    { label: t('menu.user'), value: 'user' },
    { label: t('menu.admin'), value: 'admin' },
    { label: t('menu.role'), value: 'role' },
    { label: t('menu.permission'), value: 'permission' },
    { label: t('menu.menu'), value: 'menu' },
    { label: t('menu.department'), value: 'department' },
    { label: t('menu.dictionary'), value: 'dictionary' },
    { label: t('menu.config'), value: 'config' },
    { label: t('menu.blacklist'), value: 'blacklist' },
    { label: t('menu.online_admin'), value: 'online-admin' },
    { label: t('log.module_background_task'), value: 'background-task' }
  ]
}

const getLevelType = (level) => {
  const levelMap = {
    'error': 'danger',
    'warning': 'warning',
    'info': 'success',
    'debug': 'info'
  }
  return levelMap[level?.toLowerCase()] || 'info'
}

// 获取级别的多语言标签
const getLevelLabel = (level) => {
  if (!level) return '-'
  const levelLower = level.toLowerCase()
  const levelMap = {
    'error': t('log.level_error'),
    'warning': t('log.level_warning'),
    'info': t('log.level_info'),
    'debug': t('log.level_debug')
  }
  return levelMap[levelLower] || level
}

// 获取模块的多语言标签
const getModuleLabel = (module) => {
  if (!module) return '-'
  const moduleMap = {
    'system-log': t('menu.system_log'),
    'attachment': t('menu.attachment'),
    'auth': t('log.module_auth'),
    'monitor': t('menu.monitor'),
    'operation-log': t('menu.operation_log'),
    'login-log': t('menu.login_log'),
    'recover': t('log.module_recover'),
    'payment': t('log.module_payment'),
    'payment_method': t('menu.payment_method'),
    'order': t('menu.order'),
    'export': t('menu.export'),
    'user': t('menu.user'),
    'admin': t('menu.admin'),
    'role': t('menu.role'),
    'permission': t('menu.permission'),
    'menu': t('menu.menu'),
    'department': t('menu.department'),
    'dictionary': t('menu.dictionary'),
    'config': t('menu.config'),
    'blacklist': t('menu.blacklist'),
    'online-admin': t('menu.online_admin'),
    'background-task': t('log.module_background_task')
  }
  return moduleMap[module] || module
}

// 格式化上下文为可读字符串（用于tooltip）
const formatContext = (context) => {
  if (!context) return '-'
  try {
    if (typeof context === 'string') {
      const parsed = JSON.parse(context)
      return JSON.stringify(parsed, null, 2)
    }
    return JSON.stringify(context, null, 2)
  } catch (e) {
    return String(context)
  }
}

// 格式化上下文预览（用于列表显示）
const formatContextPreview = (context) => {
  if (!context) return '-'
  try {
    let obj = context
    if (typeof context === 'string') {
      obj = JSON.parse(context)
    }
    // 如果是对象，显示前几个键值对
    if (typeof obj === 'object' && obj !== null) {
      const keys = Object.keys(obj)
      if (keys.length === 0) return '{}'
      // 只显示前2个键值对
      const preview = keys.slice(0, 2).map(key => {
        const value = obj[key]
        const valueStr = typeof value === 'object' ? JSON.stringify(value) : String(value)
        return `${key}: ${valueStr.length > 20 ? valueStr.substring(0, 20) + '...' : valueStr}`
      }).join(', ')
      return keys.length > 2 ? `${preview}...` : preview
    }
    return String(obj)
  } catch (e) {
    return String(context)
  }
}

// 使用回调清除选中状态，无需重写方法

const handleView = async (row) => {
  try {
    const res = await getSystemLogDetail(row.id)
    if (res && res.data) {
      const log = res.data.system_log || res.data.log || res.data
      if (log) {
        logDetail.value = transformSystemLogData(log)
        detailVisible.value = true
      }
    }
  } catch (error) {
    console.error('Load system log detail error:', error)
    ElMessage.error(error.response?.data?.message || error.message || t('error.default'))
  }
}

const handleDelete = (row) => handleDeleteCrud(row, loadData)

// 操作按钮配置
const operationActions = computed(() => [
  {
    key: 'view',
    label: t('common.view'),
    type: 'primary',
    permission: 'system_log.show',
    handler: handleView
  },
  {
    key: 'delete',
    label: t('common.delete'),
    type: 'danger',
    permission: 'system_log.destroy',
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

onMounted(() => {
  initDefaultSort()
  loadData()
})
</script>

<style scoped>
  
pre {
  margin: 0;
  padding: 10px;
  background: #f5f5f5;
  border-radius: 4px;
  overflow-x: auto;
  max-height: 400px;
  overflow-y: auto;
}

/* 暗黑模式样式 */
html.dark pre {
  background: var(--el-bg-color) !important;
  color: var(--el-text-color-regular) !important;
  border: 1px solid var(--el-border-color);
}

.context-preview {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  color: var(--el-color-primary);
}
</style>

