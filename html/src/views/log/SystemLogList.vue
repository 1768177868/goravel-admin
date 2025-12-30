<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.system_log') }}</span>
          <div class="header-actions">
            <el-button 
              type="danger" 
              :disabled="selectedRows.length === 0"
              @click="handleBatchDelete"
            >
              <el-icon><Delete /></el-icon>
              {{ $t('common.delete_selected') }} ({{ selectedRows.length }})
            </el-button>
          </div>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ level: '', module: '', trace_id: '', message: '', start_time: '', end_time: '' }"
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
  fetchApi: getSystemLogList,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef),
  transformData: transformSystemLogData,
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
    { label: t('log.module_recover'), value: 'recover' },
    { label: t('log.module_payment'), value: 'payment' },
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

// 转换系统日志数据（PascalCase -> snake_case）
const transformSystemLogData = (log) => {
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
    id: log.ID || log.id,
    level: log.Level || log.level || '',
    trace_id: log.TraceID || log.trace_id || '',
    message: log.Message || log.message || '',
    context: context,
    created_at: log.CreatedAt || log.created_at || ''
  }
}

// 重写 handleSearch 和 handleReset，清除选中状态
const handleSearch = () => {
  selectedRows.value = []
  selectedIds.value.clear()
  baseHandleSearch()
}

const handleReset = () => {
  selectedRows.value = []
  selectedIds.value.clear()
  baseHandleReset()
}

const handleView = async (row) => {
  try {
    const res = await getSystemLogDetail(row.id)
    if (res.data) {
      const log = res.data.system_log || res.data.log || res.data
      logDetail.value = transformSystemLogData(log)
      detailVisible.value = true
    }
  } catch (error) {
    console.error('Load system log detail error:', error)
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

.context-preview {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  color: #409eff;
}
</style>

