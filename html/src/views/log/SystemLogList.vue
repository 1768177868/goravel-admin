<template>
  <div class="log-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('log.system_log') }}</span>
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
        :initial-values="{ level: '', module: '', message: '', start_time: '', end_time: '' }"
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
        :sort-config="{ multiple: true, trigger: 'default' }"
        @checkbox-change="handleSelectionChange"
        @checkbox-all="handleSelectionChange"
        @sort-change="handleSortChange"
      >
        <vxe-column type="checkbox" width="60" />
        <template v-for="column in tableColumns" :key="column.field || column.type">
          <vxe-column
            v-if="column.type !== 'operation'"
            :field="column.field"
            :title="column.title"
            :width="column.width"
            :sortable="column.sortable"
            :fixed="column.fixed"
          >
            <template #default="{ row }">
              <!-- 文本类型 -->
              <template v-if="!column.type || column.type === 'text'">
                {{ getFieldValue(row, column.field, column.formatter) || '-' }}
              </template>
              <!-- 标签类型 -->
              <template v-else-if="column.type === 'tag'">
                <el-tag :type="getTagType(row, column)">
                  {{ getTagText(row, column) }}
                </el-tag>
              </template>
              <!-- 自定义格式化 -->
              <template v-else-if="column.type === 'custom' && column.formatter">
                {{ column.formatter(row) }}
              </template>
              <!-- 自定义插槽 - context -->
              <template v-else-if="column.type === 'custom' && column.slotName === 'context'">
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
              <!-- 其他自定义插槽 -->
              <template v-else-if="column.type === 'custom' && column.slotName">
                <slot :name="column.slotName" :row="row" :column="column" />
              </template>
            </template>
          </vxe-column>
          <!-- 操作列 -->
          <vxe-column
            v-else
            :title="column.title"
            :width="column.width"
            :fixed="column.fixed"
          >
            <template #default="{ row }">
              <slot name="operation" :row="row">
                <el-button type="primary" link @click="handleView(row)">{{ $t('common.view') }}</el-button>
                <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
              </slot>
            </template>
          </vxe-column>
        </template>
      </vxe-table>

      <Pagination
        v-model="pagination"
        @page-change="handlePageChange"
      />
    </el-card>

    <el-dialog v-model="detailVisible" :title="$t('log.detail')" width="800px">
      <el-descriptions :column="2" border v-if="logDetail">
        <el-descriptions-item :label="$t('table.id')">{{ logDetail.id }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.level')">
          <el-tag :type="getLevelType(logDetail.level)">
            {{ logDetail.level }}
          </el-tag>
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
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { useTableSort } from '../../composables/useTableSort'
import {
  getSystemLogList,
  getSystemLogDetail,
  deleteSystemLog,
  batchDeleteSystemLogs,
  cleanSystemLogs
} from '../../api/log'

const { t } = useI18n()

const tableRef = ref(null)
const loading = ref(false)
const detailVisible = ref(false)
const logDetail = ref(null)
const selectedRows = ref([])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])

const searchForm = reactive({
  level: '',
  module: '',
  message: '',
  start_time: '',
  end_time: ''
})

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'level': 'level',
  'module': 'module',
  'message': 'message',
  'context': 'context',
  'created_at': 'created_at'
}

// 使用排序 composable
const { buildOrderBy, handleSortChange, resetSort, initDefaultSort } = useTableSort({
  tableRef,
  fieldMapping,
  defaultSort: 'id:desc',
  onSortChange: () => {
    pagination.page = 1
    loadData()
  }
})

// 表格列配置
const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true,
    type: 'text'
  },
  {
    field: 'level',
    title: t('log.level'),
    width: 100,
    sortable: true,
    type: 'tag',
    tagConfig: {
      value: (row) => row.level,
      type: (val) => getLevelType(val),
      text: (val) => val || '-'
    }
  },
  {
    field: 'message',
    title: t('log.message'),
    sortable: true,
    type: 'text'
  },
  {
    field: 'context',
    title: t('log.context'),
    width: 200,
    type: 'custom',
    slotName: 'context'
  },
  {
    field: 'created_at',
    title: t('log.time'),
    width: 180,
    sortable: true,
    type: 'text'
  },
  {
    type: 'operation',
    title: t('table.operation'),
    width: 100,
    fixed: 'right'
  }
])

// 获取字段值（支持 PascalCase 和 snake_case，以及格式化函数）
const getFieldValue = (row, field, formatter) => {
  if (formatter && typeof formatter === 'function') {
    return formatter(row)
  }
  if (!field) return ''
  const pascalField = field.charAt(0).toUpperCase() + field.slice(1)
  return row[pascalField] !== undefined ? row[pascalField] : (row[field] !== undefined ? row[field] : '')
}

// 获取标签类型
const getTagType = (row, column) => {
  if (column.tagConfig && column.tagConfig.type) {
    const value = column.tagConfig.value(row)
    return column.tagConfig.type(value)
  }
  return 'info'
}

// 获取标签文本
const getTagText = (row, column) => {
  if (column.tagConfig && column.tagConfig.text) {
    const value = column.tagConfig.value(row)
    return column.tagConfig.text(value)
  }
  return getFieldValue(row, column.field) || '-'
}

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'level',
    label: t('log.level'),
    type: 'select',
    width: '120px',
    options: [
      { label: 'error', value: 'error' },
      { label: 'warning', value: 'warning' },
      { label: 'info', value: 'info' },
      { label: 'debug', value: 'debug' }
    ],
    advanced: false
  },
  {
    prop: 'module',
    label: t('log.module'),
    type: 'input',
    width: '150px',
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

const getLevelType = (level) => {
  const levelMap = {
    'error': 'danger',
    'warning': 'warning',
    'info': 'success',
    'debug': 'info'
  }
  return levelMap[level?.toLowerCase()] || 'info'
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
    message: log.Message || log.message || '',
    context: context,
    created_at: log.CreatedAt || log.created_at || ''
  }
}

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      order_by: buildOrderBy(),
      ...searchForm
    }
    // 移除空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })
    const res = await getSystemLogList(params)
    if (res.data) {
      const logs = res.data.list || []
      tableData.value = logs.map(log => transformSystemLogData(log))
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load system log list error:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  Object.keys(searchForm).forEach(key => {
    searchForm[key] = ''
  })
  resetSort()
  pagination.page = 1
  loadData()
}

const handlePageChange = ({ currentPage, pageSize }) => {
  pagination.page = currentPage
  pagination.pageSize = pageSize
  loadData()
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

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(t('log.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteSystemLog(row.id)
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
    await batchDeleteSystemLogs(ids)
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
    await cleanSystemLogs()
    ElMessage.success(t('log.clean_success'))
    selectedRows.value = []
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Clean error:', error)
    }
  }
}

onMounted(() => {
  initDefaultSort()
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

