<template>
  <div class="log-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('log.operation_log') }}</span>
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
        :sort-config="{ multiple: true, trigger: 'default' }"
        @checkbox-change="handleSelectionChange"
        @checkbox-all="handleSelectionChange"
        @sort-change="handleSortChange"
      >
        <vxe-column type="checkbox" width="60" />
        <vxe-column field="id" :title="$t('table.id')" width="80" sortable />
        <vxe-column field="admin" :title="$t('log.admin')">
          <template #default="{ row }">
            {{ (row.admin || row.Admin)?.username || (row.admin || row.Admin)?.Username || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="method" :title="$t('log.method')" width="100" sortable />
        <vxe-column field="path" :title="$t('log.path')" sortable />
        <vxe-column field="ip" :title="$t('log.ip')" width="150" sortable />
        <vxe-column field="status_code" :title="$t('log.status_code')" width="100" sortable />
        <vxe-column field="created_at" :title="$t('log.operation_time')" width="180" sortable />
        <vxe-column :title="$t('table.operation')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleView(row)">{{ $t('common.view') }}</el-button>
            <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </vxe-column>
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
import {
  getOperationLogList,
  getOperationLogDetail,
  deleteOperationLog,
  batchDeleteOperationLogs,
  cleanOperationLogs
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

// 排序状态
const sortConfig = ref({
  multiple: true,
  data: []
})

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

// 搜索表单初始值
const initialSearchForm = {
  username: '',
  method: '',
  path: '',
  ip: '',
  status: '',
  start_time: getSevenDaysAgo(),
  end_time: ''
}

const searchForm = reactive({ ...initialSearchForm })

// 搜索表单字段配置（JSON 方式）
const searchFields = computed(() => [
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
    options: [
      { label: 'GET', value: 'GET' },
      { label: 'POST', value: 'POST' },
      { label: 'PUT', value: 'PUT' },
      { label: 'DELETE', value: 'DELETE' },
      { label: 'PATCH', value: 'PATCH' }
    ],
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
    prop: 'ip',
    label: t('log.ip'),
    type: 'input',
    width: '150px',
    advanced: true
  },
  {
    prop: 'status',
    label: t('log.status_code'),
    type: 'input',
    width: '120px',
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
])

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
    ip: log.IP || log.ip || '',
    status_code: log.Status || log.status || log.StatusCode || log.status_code || 0,
    created_at: log.CreatedAt || log.created_at || '',
    params: params,
    request: log.Request || log.request || null,
    response: log.Response || log.response || null
  }
}

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'method': 'method',
  'path': 'path',
  'ip': 'ip',
  'status_code': 'status', // 前端使用 status_code，数据库字段是 status
  'created_at': 'created_at'
}

// 构建排序参数字符串
const buildOrderBy = () => {
  if (!sortConfig.value.data || sortConfig.value.data.length === 0) {
    return 'id:desc' // 默认按id倒序
  }
  
  return sortConfig.value.data
    .map(sort => {
      const direction = sort.order === 'asc' ? 'asc' : 'desc'
      // 映射字段名到数据库字段名
      const dbField = fieldMapping[sort.field] || sort.field
      return `${dbField}:${direction}`
    })
    .join(',')
}

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...searchForm,
      order_by: buildOrderBy()
    }
    // 移除空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })
    const res = await getOperationLogList(params)
    if (res.data) {
      const logs = res.data.list || []
      tableData.value = logs.map(log => transformOperationLogData(log))
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load operation log list error:', error)
  } finally {
    loading.value = false
  }
}

// 处理排序变化
const handleSortChange = ({ column, property, order, sortBy, sortList }) => {
  // 更新排序配置
  sortConfig.value.data = sortList || []
  // 重新加载数据
  pagination.page = 1
  loadData()
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.username = ''
  searchForm.method = ''
  searchForm.path = ''
  searchForm.ip = ''
  searchForm.status = ''
  searchForm.start_time = getSevenDaysAgo()
  searchForm.end_time = ''
  // 重置排序
  sortConfig.value.data = []
  if (tableRef.value) {
    tableRef.value.clearSort()
  }
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

onMounted(() => {
  // 初始化默认排序（id倒序）
  sortConfig.value.data = [{ field: 'id', order: 'desc' }]
  // 设置表格的默认排序（使用 nextTick 确保表格已渲染）
  nextTick(() => {
    if (tableRef.value) {
      tableRef.value.setSort([
        { field: 'id', order: 'desc' }
      ])
    }
  })
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

