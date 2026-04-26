<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.login_log') }}</span>
          <div class="header-actions">
            <!-- 清理日志功能已禁用（后端权限被注释） -->
            <!-- <el-button 
              type="warning" 
              :disabled="getButtonState('login_log.clean').disabled"
              @click="handleClean"
            >
              <el-icon><Delete /></el-icon>
              {{ $t('log.clean') || '清空日志' }}
            </el-button> -->
            <el-button 
              type="danger" 
              :disabled="selectedRows.length === 0 || getButtonState('login_log.batch_delete').disabled"
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
        :initial-values="initialSearchValues"
        i18n-prefix="log"
        @search="handleSearch"
        @reset="handleReset"
      />

      <!-- 表格工具栏 -->
      <TableToolbar
        :on-refresh="handleRefresh"
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

        <template #status="{ row }">
          <el-tag :type="(row.status ?? 1) === 1 ? 'success' : 'danger'">
            {{ (row.status ?? 1) === 1 ? $t('log.success') : $t('log.failed') }}
          </el-tag>
        </template>

        <template #message="{ row }">
          {{ translateLoginMessage(row.message || '') }}
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

    <el-dialog v-model="detailVisible" :title="$t('log.detail')" width="1100px">
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
        <el-descriptions-item :label="$t('log.message')" :span="2">{{ translateLoginMessage(logDetail.message || '') }}</el-descriptions-item>
        <el-descriptions-item :label="$t('log.request')" :span="2">
          <pre v-if="logDetail.request" class="request-preview-content">{{ formatRequest(logDetail.request) }}</pre>
          <span v-else>-</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, View } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import TableToolbar from '../../components/TableToolbar.vue'
import { useListPage } from '../../composables/useListPage'
import { buildSearchParams } from '../../utils/buildSearchParams'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import { useColumnSetting } from '../../composables/useColumnSetting'
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
// 维护跨页选中的ID集合
const selectedIds = ref(new Set())

// 初始搜索值（避免每次渲染创建新对象）
const initialSearchValues = {
  username: '',
  ip: '',
  status: '',
  start_time: '',
  end_time: ''
}
const logDetail = ref(null)
const selectedRows = ref([])

// 使用 CRUD composable（删除和批量删除）
const { handleDelete: handleDeleteCrud, handleBatchDelete: handleBatchDeleteCrud } = useCrud({
  deleteApi: deleteLoginLog,
  batchDeleteApi: batchDeleteLoginLogs
})

// 字段名映射：前端字段名 -> 数据库字段名（只包含不同的字段）
const fieldMapping = {
  'admin': 'admin_id' // 前端使用 admin，数据库字段是 admin_id
}

// 转换登录日志数据（以 snake_case 为主）
const transformLoginLogData = (log) => {
  return {
    id: log.id,
    admin: log.admin ? {
      username: log.admin.username || ''
    } : null,
    ip: log.ip || '',
    user_agent: log.user_agent || '',
    location: log.location || '',
    status: log.status || 0,
    message: log.message || '',
    request: log.request || '',
    created_at: log.created_at || ''
  }
}

// 初始搜索表单
const initialSearchForm = {
  username: '',
  ip: '',
  status: '',
  start_time: '',
  end_time: ''
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
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getLoginLogList,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef),
  transformData: transformLoginLogData,
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

// 使用回调清除选中状态，无需重写方法

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
    key: 'admin'
  },
  {
    field: 'ip',
    title: t('log.ip'),
    width: 150,
    sortable: false,
    key: 'ip'
  },
  {
    field: 'location',
    title: t('log.location'),
    width: 200,
    sortable: false,
    key: 'location'
  },
  {
    field: 'user_agent',
    title: t('log.user_agent'),
    sortable: false,
    key: 'user_agent'
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 100,
    sortable: false,
    slot: 'status',
    key: 'status'
  },
  {
    field: 'message',
    title: t('log.message'),
    sortable: false,
    slot: 'message',
    key: 'message'
  },
  {
    field: 'created_at',
    title: t('log.login_time'),
    width: 180,
    sortable: true,
    key: 'created_at'
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation',
    sortable: false,
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
} = useColumnSetting('login_log', allTableColumns)

// 处理刷新
const handleRefresh = () => {
  loadData()
}

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

const handleDelete = (row) => handleDeleteCrud(row, loadData)

// 操作按钮配置
const operationActions = computed(() => [
  {
    key: 'view',
    label: t('common.view'),
    type: 'primary',
    permission: 'login_log.show',
    icon: View,
    handler: handleView
  },
  {
    key: 'delete',
    label: t('common.delete'),
    type: 'danger',
    permission: 'login_log.destroy',
    icon: Delete,
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

// 翻译登录消息
const translateLoginMessage = (messageKey) => {
  if (!messageKey) return '-'
  
  // 尝试翻译消息键，如果翻译不存在则返回原值
  const translation = t(`log.${messageKey}`, null)
  return translation !== `log.${messageKey}` ? translation : messageKey
}

// 格式化请求数据（如果是 JSON 字符串，则格式化显示）
const formatRequest = (request) => {
  if (!request) return '-'
  
  try {
    // 尝试解析为 JSON
    const parsed = JSON.parse(request)
    // 格式化 JSON 字符串，缩进 2 个空格
    return JSON.stringify(parsed, null, 2)
  } catch (e) {
    // 如果不是有效的 JSON，直接返回原字符串
    return request
  }
}

const handleClean = async () => {
  try {
    // 构建搜索参数
    const searchParams = buildSearchParams(searchForm, {})
    
    // 如果有搜索条件，提示用户将清空符合条件的日志
    const hasSearchConditions = Object.keys(searchParams).some(key => {
      const value = searchParams[key]
      return value !== '' && value !== null && value !== undefined
    })
    
    const confirmMessage = hasSearchConditions 
      ? t('log.clean_filtered_confirm')
      : t('log.clean_confirm')
    
    await ElMessageBox.confirm(confirmMessage, t('form.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    
    // 传递搜索条件给清空 API
    await cleanLoginLogs(searchParams)
    ElMessage.success(t('log.clean_success'))
    selectedRows.value = []
    selectedIds.value.clear()
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
.request-preview-content {
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 300px;
  overflow-y: auto;
  background: var(--bg-color-tertiary);
  padding: var(--space-sm);
  border-radius: var(--border-radius-sm);
  margin: 0;
}

/* 暗黑模式样式 */
html.dark .request-preview-content {
  background: var(--el-bg-color) !important;
  color: var(--el-text-color-regular) !important;
  border: 1px solid var(--el-border-color);
}
</style>

