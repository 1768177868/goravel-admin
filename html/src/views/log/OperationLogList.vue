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

      <!-- 搜索表单 -->
      <el-form :model="searchForm" :inline="true" class="search-form">
        <el-form-item :label="$t('log.username')">
          <el-input
            v-model="searchForm.username"
            :placeholder="$t('form.please_enter') + $t('log.username')"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item :label="$t('log.method')">
          <el-select
            v-model="searchForm.method"
            :placeholder="$t('form.please_select') + $t('log.method')"
            clearable
            style="width: 150px"
          >
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('log.path')">
          <el-input
            v-model="searchForm.path"
            :placeholder="$t('form.please_enter') + $t('log.path')"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item :label="$t('log.ip')">
          <el-input
            v-model="searchForm.ip"
            :placeholder="$t('form.please_enter') + $t('log.ip')"
            clearable
            style="width: 150px"
          />
        </el-form-item>
        <el-form-item :label="$t('log.status_code')">
          <el-input
            v-model="searchForm.status"
            :placeholder="$t('form.please_enter') + $t('log.status_code')"
            clearable
            style="width: 120px"
          />
        </el-form-item>
        <el-form-item :label="$t('log.start_time')">
          <el-date-picker
            v-model="searchForm.start_time"
            type="datetime"
            :placeholder="$t('form.please_select') + $t('log.start_time')"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 180px"
            clearable
          />
        </el-form-item>
        <el-form-item :label="$t('log.end_time')">
          <el-date-picker
            v-model="searchForm.end_time"
            type="datetime"
            :placeholder="$t('form.please_select') + $t('log.end_time')"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 180px"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            {{ $t('log.search') }}
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            {{ $t('log.reset') }}
          </el-button>
        </el-form-item>
      </el-form>

      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        resizable
        height="600"
        @checkbox-change="handleSelectionChange"
        @checkbox-all="handleSelectionChange"
      >
        <vxe-column type="checkbox" width="60" />
        <vxe-column type="seq" width="60" :title="$t('table.seq')" />
        <vxe-column field="id" :title="$t('table.id')" width="80" />
        <vxe-column field="admin" :title="$t('log.admin')">
          <template #default="{ row }">
            {{ (row.admin || row.Admin)?.username || (row.admin || row.Admin)?.Username || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="method" :title="$t('log.method')" width="100" />
        <vxe-column field="path" :title="$t('log.path')" />
        <vxe-column field="ip" :title="$t('log.ip')" width="150" />
        <vxe-column field="status_code" :title="$t('log.status_code')" width="100" />
        <vxe-column field="created_at" :title="$t('log.operation_time')" width="180" />
        <vxe-column :title="$t('table.operation')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleView(row)">{{ $t('common.view') }}</el-button>
            <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </vxe-column>
      </vxe-table>

      <vxe-pager
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
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
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Delete } from '@element-plus/icons-vue'
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

const searchForm = reactive({
  username: '',
  method: '',
  path: '',
  ip: '',
  status: '',
  start_time: '',
  end_time: ''
})

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

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...searchForm
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

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  Object.keys(searchForm).forEach(key => {
    searchForm[key] = ''
  })
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

.search-form {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
}

pre {
  margin: 0;
  padding: 10px;
  background: #f5f5f5;
  border-radius: 4px;
  overflow-x: auto;
}
</style>

