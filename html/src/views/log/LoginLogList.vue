<template>
  <div class="log-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('log.login_log') }}</span>
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
        :initial-values="{ username: '', ip: '', status: '', start_time: '', end_time: '' }"
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
            <template v-else-if="column.slot === 'status'" #default="{ row }">
              <el-tag :type="(row.status ?? row.Status ?? 1) === 1 ? 'success' : 'danger'">
                {{ (row.status ?? row.Status ?? 1) === 1 ? $t('log.success') : $t('log.failed') }}
              </el-tag>
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <el-button 
                type="primary" 
                link 
                :disabled="getButtonState('login_log.show').disabled"
                @click="handleView(row)"
              >
                {{ $t('common.view') }}
              </el-button>
              <el-button 
                type="danger" 
                link 
                :disabled="getButtonState('login_log.destroy').disabled"
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
        @page-change="handlePageChange"
      />
    </el-card>

    <el-dialog v-model="detailVisible" :title="$t('log.detail')" width="800px">
      <el-descriptions :column="2" border v-if="logDetail">
        <el-descriptions-item :label="$t('table.id')">{{ logDetail.id }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.admin')">{{ logDetail.admin?.username }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.ip')">{{ logDetail.ip }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.location')">{{ logDetail.location || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('table.status')">
          <el-tag :type="logDetail.status === 1 ? 'success' : 'danger'">
            {{ logDetail.status === 1 ? $t('log.success') : $t('log.failed') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('log.user_agent')">{{ logDetail.user_agent }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.login_time')">{{ logDetail.created_at }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.message')" :span="2">{{ logDetail.message }}</el-descriptions-item>
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
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import {
  getLoginLogList,
  getLoginLogDetail,
  deleteLoginLog,
  batchDeleteLoginLogs,
  cleanLoginLogs
} from '../../api/log'

const { t } = useI18n()
const { getButtonState } = usePermission()

const tableRef = ref(null)
const detailVisible = ref(false)
const logDetail = ref(null)
const selectedRows = ref([])

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'admin': 'admin_id',
  'ip': 'ip',
  'location': 'location',
  'user_agent': 'user_agent',
  'status': 'status',
  'message': 'message',
  'created_at': 'created_at'
}

// 转换登录日志数据（PascalCase -> snake_case）
const transformLoginLogData = (log) => {
  return {
    id: log.ID || log.id,
    admin: log.Admin ? {
      username: log.Admin.Username || log.Admin.username || ''
    } : (log.admin ? {
      username: log.admin.username || ''
    } : null),
    ip: log.IP || log.ip || '',
    user_agent: log.UserAgent || log.user_agent || '',
    location: log.Location || log.location || '',
    status: log.Status || log.status || 0,
    message: log.Message || log.message || '',
    created_at: log.CreatedAt || log.created_at || ''
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
  fetchApi: getLoginLogList,
  initialSearchForm: {
    username: '',
    ip: '',
    status: '',
    start_time: '',
    end_time: ''
  },
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  },
  transformData: transformLoginLogData
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
    field: 'admin',
    title: t('log.admin'),
    slot: 'admin',
    sortable: false
  },
  {
    field: 'ip',
    title: t('log.ip'),
    width: 150,
    sortable: true
  },
  {
    field: 'location',
    title: t('log.location'),
    width: 200,
    sortable: false
  },
  {
    field: 'user_agent',
    title: t('log.user_agent'),
    sortable: false
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 100,
    sortable: true,
    slot: 'status'
  },
  {
    field: 'message',
    title: t('log.message'),
    sortable: false
  },
  {
    field: 'created_at',
    title: t('log.login_time'),
    width: 180,
    sortable: true
  },
  {
    title: t('table.operation'),
    width: 100,
    fixed: 'right',
    slot: 'operation',
    sortable: false
  }
])

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'username',
    label: t('log.username'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'ip',
    label: t('log.ip'),
    type: 'input',
    width: '150px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('log.status'),
    type: 'select',
    width: '120px',
    options: [
      { label: t('log.success'), value: '1' },
      { label: t('log.failed'), value: '0' }
    ],
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

// loadData, handleSearch, handleReset, handlePageChange 已由 useListPage 提供

const handleView = async (row) => {
  try {
    const res = await getLoginLogDetail(row.id)
    if (res.data) {
      const log = res.data.login_log || res.data.log || res.data
      logDetail.value = transformLoginLogData(log)
      detailVisible.value = true
    }
  } catch (error) {
    console.error('Load login log detail error:', error)
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(t('log.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteLoginLog(row.id)
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
    await batchDeleteLoginLogs(ids)
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
    await cleanLoginLogs()
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

</style>

