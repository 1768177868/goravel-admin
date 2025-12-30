<template>
  <div class="online-admin-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.online_admin') }}</span>
          <div>
            <el-button 
              type="danger" 
              :disabled="!selectedRows || selectedRows.length === 0 || getButtonState('admin.kick_out').disabled"
              @click="handleBatchKickOut"
            >
              <el-icon><Delete /></el-icon>
              {{ $t('online_admin.batch_kick_out') }}
            </el-button>
          </div>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchValues"
        i18n-prefix="online_admin"
        @search="handleSearch"
        @reset="handleReset"
      />

      <!-- 表格工具栏 -->
      <div class="table-toolbar">
        <div class="toolbar-left"></div>
        <div class="toolbar-right">
          <ColumnSettingDialog
            v-model="showColumnSetting"
            :visible-columns="visibleColumns"
            :all-columns="allColumns"
            :default-visible-columns="defaultVisibleColumns"
            @confirm="handleSaveColumnSetting"
          />
        </div>
      </div>

      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        :column-config="{ resizable: true }"
        height="600"
        :sort-config="{ multiple: false, trigger: 'default' }"
        @sort-change="handleSortChange"
        @checkbox-change="handleCheckboxChange"
        @checkbox-all="handleCheckboxAll"
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
          >
            <template v-if="column.slot === 'avatar'" #default="{ row }">
              <el-avatar :size="32" :src="row.avatar">
                {{ row.nickname ? row.nickname.charAt(0) : (row.username ? row.username.charAt(0) : 'U') }}
              </el-avatar>
            </template>
            <template v-else-if="column.slot === 'last_active'" #default="{ row }">
              {{ formatTime(row.last_active) }}
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <el-button 
                type="danger" 
                link 
                :disabled="getButtonState('admin.kick_out').disabled"
                @click="handleKickOut(row)"
              >
                {{ $t('online_admin.kick_out') }}
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

  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, markRaw } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Setting } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

const SettingIcon = markRaw(Setting)
import { getOnlineAdminList, kickOutOnlineAdmin, batchKickOutOnlineAdmins } from '@/api/onlineAdmin'
import SearchForm from '@/components/SearchForm.vue'
import Pagination from '@/components/Pagination.vue'
import ColumnSettingDialog from '@/components/ColumnSettingDialog.vue'
import { useColumnSetting } from '@/composables/useColumnSetting'
import { usePermission } from '@/composables/usePermission'
import { useTableSort } from '@/composables/useTableSort'

const { t } = useI18n()
const { getButtonState } = usePermission()

const tableRef = ref(null)
const loading = ref(false)
const tableData = ref([])
const selectedRows = ref([])

// 初始搜索值（避免每次渲染创建新对象）
const initialSearchValues = {
  username: '',
  ip: '',
  browser: '',
  os: ''
}

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'username': 'username',
  'nickname': 'nickname',
  'browser': 'browser',
  'ip': 'ip',
  'os': 'os',
  'session_id': 'session_id',
  'last_active': 'last_used_at'
}

// 使用排序 composable
const { buildOrderBy, handleSortChange, resetSort, initDefaultSort } = useTableSort({
  tableRef,
  fieldMapping,
  defaultSort: 'last_used_at:desc',
  onSortChange: () => {
    pagination.page = 1
    fetchData()
  }
})

// 所有列的完整配置（必须在 useColumnSetting 之前定义）
const allTableColumns = computed(() => [
  { type: 'checkbox', width: 50, fixed: 'left', key: 'checkbox' },
  {
    field: 'username',
    title: t('online_admin.username'),
    width: 120,
    sortable: false,
    key: 'username'
  },
  {
    field: 'nickname',
    title: t('online_admin.nickname'),
    width: 120,
    sortable: false,
    key: 'nickname'
  },
  {
    field: 'avatar',
    slot: 'avatar',
    title: t('online_admin.avatar'),
    width: 80,
    key: 'avatar'
  },
  {
    field: 'browser',
    title: t('online_admin.browser'),
    width: 150,
    sortable: false,
    key: 'browser'
  },
  {
    field: 'ip',
    title: t('online_admin.ip'),
    width: 150,
    sortable: false,
    key: 'ip'
  },
  {
    field: 'os',
    title: t('online_admin.os'),
    width: 150,
    sortable: false,
    key: 'os'
  },
  {
    field: 'session_id',
    title: t('online_admin.session_id'),
    width: 200,
    key: 'session_id'
  },
  {
    field: 'last_active',
    slot: 'last_active',
    title: t('online_admin.last_active'),
    width: 180,
    sortable: true,
    key: 'last_active'
  },
  {
    slot: 'operation',
    title: t('common.operation'),
    width: 120,
    fixed: 'right',
    key: 'operation'
  }
])

// 使用列设置 composable
const {
  tableColumns,
  showColumnSetting,
  allColumns,
  visibleColumns,
  defaultVisibleColumns,
  handleSaveColumnSetting
} = useColumnSetting('online_admin', allTableColumns)

const searchForm = reactive({
  username: '',
  ip: '',
  browser: '',
  os: ''
})

const searchFields = computed(() => [
  {
    prop: 'username',
    label: t('online_admin.username'),
    type: 'input',
    width: '200px',
    placeholder: t('online_admin.username_placeholder'),
    advanced: false
  },
  {
    prop: 'ip',
    label: t('online_admin.ip'),
    type: 'input',
    width: '200px',
    placeholder: t('online_admin.ip_placeholder'),
    advanced: false
  },
  {
    prop: 'browser',
    label: t('online_admin.browser'),
    type: 'input',
    width: '200px',
    placeholder: t('online_admin.browser_placeholder'),
    advanced: false
  },
  {
    prop: 'os',
    label: t('online_admin.os'),
    type: 'input',
    width: '200px',
    placeholder: t('online_admin.os_placeholder'),
    advanced: false
  }
])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  
  // 如果是字符串，尝试解析
  let date
  if (typeof time === 'string') {
    date = new Date(time)
  } else if (time instanceof Date) {
    date = time
  } else {
    return String(time)
  }
  
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    return String(time)
  }
  
  // 格式化为 YYYY-MM-DD HH:mm:ss
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      order_by: buildOrderBy(),
      ...searchForm
    }
    const res = await getOnlineAdminList(params)
    if (res.code === 200) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    } else {
      ElMessage.error(res.message || t('common.operation_failed'))
    }
  } catch (error) {
    console.error('Fetch online admins error:', error)
    // 如果错误已经在响应拦截器中处理过，就不再重复显示
    if (!error?.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
      ElMessage.error(errorMessage)
    }
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchData()
}

const handleReset = () => {
  Object.assign(searchForm, {
    username: '',
    ip: '',
    browser: '',
    os: ''
  })
  resetSort()
  pagination.page = 1
  fetchData()
}

const handlePageChange = (page, pageSize) => {
  pagination.page = page
  pagination.pageSize = pageSize
  fetchData()
}

const handleCheckboxChange = ({ row, checked }) => {
  if (!selectedRows.value) {
    selectedRows.value = []
  }
  if (checked) {
    if (!selectedRows.value.find(item => item.id === row.id)) {
      selectedRows.value.push(row)
    }
  } else {
    selectedRows.value = selectedRows.value.filter(item => item.id !== row.id)
  }
}

const handleCheckboxAll = ({ checked, records }) => {
  if (!selectedRows.value) {
    selectedRows.value = []
  }
  if (checked) {
    selectedRows.value = Array.isArray(records) ? [...records] : []
  } else {
    selectedRows.value = []
  }
}

const handleKickOut = async (row) => {
  try {
    await ElMessageBox.confirm(
      t('online_admin.kick_out_confirm', { username: row.username }),
      t('common.confirm'),
      {
        type: 'warning'
      }
    )
    const res = await kickOutOnlineAdmin(row.id)
    if (res.code === 200) {
      ElMessage.success(t('online_admin.kick_out_success'))
      fetchData()
    } else {
      ElMessage.error(res.message || t('common.operation_failed'))
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Kick out error:', error)
      // 如果错误已经在响应拦截器中处理过，就不再重复显示
      if (!error?.__handled) {
        const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
        ElMessage.error(errorMessage)
      }
    }
  }
}

const handleBatchKickOut = async () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning(t('online_admin.select_admins_first'))
    return
  }
  try {
    await ElMessageBox.confirm(
      t('online_admin.batch_kick_out_confirm', { count: selectedRows.value.length }),
      t('common.confirm'),
      {
        type: 'warning'
      }
    )
    const tokenIds = selectedRows.value.map(row => row.id)
    const res = await batchKickOutOnlineAdmins(tokenIds)
    if (res.code === 200) {
      ElMessage.success(t('online_admin.batch_kick_out_success'))
      selectedRows.value = []
      fetchData()
    } else {
      ElMessage.error(res.message || t('common.operation_failed'))
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Batch kick out error:', error)
      // 如果错误已经在响应拦截器中处理过，就不再重复显示
      if (!error?.__handled) {
        const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
        ElMessage.error(errorMessage)
      }
    }
  }
}

onMounted(() => {
  initDefaultSort()
  fetchData()
})
</script>

<style scoped>
.online-admin-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding: 0 4px;
}

.toolbar-left {
  flex: 1;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>

