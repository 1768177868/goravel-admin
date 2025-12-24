<template>
  <div class="online-user-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('online_user.title') }}</span>
          <div>
            <el-button type="info" @click="showColumnSetting = true">
              <el-icon><Setting /></el-icon>
              {{ $t('common.column_setting') }}
            </el-button>
            <el-button 
              type="danger" 
              :disabled="selectedRows.length === 0 || getButtonState('admin.kick_out').disabled"
              @click="handleBatchKickOut"
            >
              <el-icon><Delete /></el-icon>
              {{ $t('online_user.batch_kick_out') }}
            </el-button>
          </div>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ username: '', ip: '', browser: '', os: '' }"
        i18n-prefix="online_user"
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
                {{ $t('online_user.kick_out') }}
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

    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      v-model="showColumnSetting"
      :visible-columns="visibleColumns"
      :all-columns="allColumns"
      :default-visible-columns="['username', 'nickname', 'avatar', 'browser', 'ip', 'os', 'session_id', 'last_active']"
      @confirm="handleSaveColumnSetting"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Setting } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { getOnlineUserList, kickOutOnlineUser, batchKickOutOnlineUsers } from '@/api/onlineUser'
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

// 所有可配置的列（不包括checkbox和operation，它们始终显示）
const allColumnsConfig = computed(() => [
  { key: 'username', title: t('online_user.username'), required: false },
  { key: 'nickname', title: t('online_user.nickname'), required: false },
  { key: 'avatar', title: t('online_user.avatar'), required: false },
  { key: 'browser', title: t('online_user.browser'), required: false },
  { key: 'ip', title: t('online_user.ip'), required: false },
  { key: 'os', title: t('online_user.os'), required: false },
  { key: 'session_id', title: t('online_user.session_id'), required: false },
  { key: 'last_active', title: t('online_user.last_active'), required: false }
])

// 使用列设置 composable
const {
  showColumnSetting,
  visibleColumns,
  allColumns,
  handleSaveColumnSetting,
  getVisibleColumns
} = useColumnSetting({
  storageKey: 'online_user_column_setting',
  allColumns: allColumnsConfig,
  defaultVisibleColumns: ['username', 'nickname', 'avatar', 'browser', 'ip', 'os', 'session_id', 'last_active'],
  alwaysVisibleKeys: ['checkbox', 'operation']
})

const searchForm = reactive({
  username: '',
  ip: '',
  browser: '',
  os: ''
})

const searchFields = computed(() => [
  {
    prop: 'username',
    label: t('online_user.username'),
    type: 'input',
    width: '200px',
    placeholder: t('online_user.username_placeholder'),
    advanced: false
  },
  {
    prop: 'ip',
    label: t('online_user.ip'),
    type: 'input',
    width: '200px',
    placeholder: t('online_user.ip_placeholder'),
    advanced: false
  },
  {
    prop: 'browser',
    label: t('online_user.browser'),
    type: 'input',
    width: '200px',
    placeholder: t('online_user.browser_placeholder'),
    advanced: false
  },
  {
    prop: 'os',
    label: t('online_user.os'),
    type: 'input',
    width: '200px',
    placeholder: t('online_user.os_placeholder'),
    advanced: false
  }
])

// 所有列的完整配置
const allTableColumns = computed(() => [
  { type: 'checkbox', width: 50, fixed: 'left', key: 'checkbox' },
  {
    field: 'username',
    title: t('online_user.username'),
    width: 120,
    sortable: false,
    key: 'username'
  },
  {
    field: 'nickname',
    title: t('online_user.nickname'),
    width: 120,
    sortable: false,
    key: 'nickname'
  },
  {
    slot: 'avatar',
    title: t('online_user.avatar'),
    width: 80,
    key: 'avatar'
  },
  {
    field: 'browser',
    title: t('online_user.browser'),
    width: 150,
    sortable: false,
    key: 'browser'
  },
  {
    field: 'ip',
    title: t('online_user.ip'),
    width: 150,
    sortable: false,
    key: 'ip'
  },
  {
    field: 'os',
    title: t('online_user.os'),
    width: 150,
    sortable: false,
    key: 'os'
  },
  {
    field: 'session_id',
    title: t('online_user.session_id'),
    width: 200,
    key: 'session_id'
  },
  {
    slot: 'last_active',
    title: t('online_user.last_active'),
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

// 根据visibleColumns过滤显示的列
const tableColumns = getVisibleColumns(allTableColumns)

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
    const res = await getOnlineUserList(params)
    if (res.code === 200) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    } else {
      ElMessage.error(res.message || t('common.operation_failed'))
    }
  } catch (error) {
    console.error('Fetch online users error:', error)
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
      t('online_user.kick_out_confirm', { username: row.username }),
      t('common.confirm'),
      {
        type: 'warning'
      }
    )
    const res = await kickOutOnlineUser(row.id)
    if (res.code === 200) {
      ElMessage.success(t('online_user.kick_out_success'))
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
    ElMessage.warning(t('online_user.select_users_first'))
    return
  }
  try {
    await ElMessageBox.confirm(
      t('online_user.batch_kick_out_confirm', { count: selectedRows.value.length }),
      t('common.confirm'),
      {
        type: 'warning'
      }
    )
    const tokenIds = selectedRows.value.map(row => row.id)
    const res = await batchKickOutOnlineUsers(tokenIds)
    if (res.code === 200) {
      ElMessage.success(t('online_user.batch_kick_out_success'))
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
.online-user-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

