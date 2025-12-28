<template>
  <div class="user-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('user.title') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('user.store').disabled"
            @click="handleAdd"
          >
            <el-icon><PlusIcon /></el-icon>
            {{ $t('user.add_user') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchValues"
        i18n-prefix="user"
        @search="handleSearch"
        @reset="handleReset"
      />

      <!-- vxe-table -->
      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        :column-config="{ resizable: true }"
        height="600"
        :sort-config="{ multiple: false, trigger: 'default' }"
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
          >
            <template v-if="column.slot === 'status'" #default="{ row }">
              <el-switch
                :model-value="Number(row.status ?? row.Status ?? 1) === 1"
                :disabled="getButtonState('user.update').disabled"
                @change="(val) => handleStatusChange(row, val)"
              />
            </template>
            <template v-else-if="column.slot === 'balance'" #default="{ row }">
              <span style="color: #409EFF; font-weight: bold;">¥{{ formatMoney(row.balance || 0) }}</span>
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <TableActionButtons
                :row="row"
                :primary-actions="getPrimaryActions(row)"
                :more-actions="getMoreActions(row)"
                :get-button-state="getButtonState"
                @action="handleAction"
              />
            </template>
          </vxe-column>
        </template>
      </vxe-table>

      <!-- 分页 -->
      <Pagination
        v-model="pagination"
        @page-change="handlePageChange"
      />
    </el-card>

    <!-- 添加/编辑对话框 -->
    <UserForm
      ref="userFormRef"
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import UserForm from './UserForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import { getStatusOptions } from '../../utils/fieldOptions'
import {
  getUserList,
  deleteUser,
  updateUser
} from '../../api/user'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

const PlusIcon = markRaw(Plus)

const { getButtonState } = usePermission()
const { t } = useI18n()
const tableRef = ref(null)
const userFormRef = ref(null)

// 初始搜索值（避免每次渲染创建新对象）
const initialSearchValues = {
  username: '',
  email: '',
  phone: '',
  status: ''
}

const {
  dialogVisible,
  editId,
  handleAdd,
  handleClose,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: deleteUser
})

const fieldMapping = {
  'id': 'id',
  'username': 'username',
  'email': 'email',
  'phone': 'phone',
  'status': 'status',
  'created_at': 'created_at'
}

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
  fetchApi: getUserList,
  initialSearchForm: {
    username: '',
    email: '',
    phone: '',
    status: ''
  },
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  }
})

const searchFields = computed(() => [
  {
    prop: 'username',
    label: t('table.username'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'email',
    label: t('table.email'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'phone',
    label: t('table.phone'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'select',
    width: '150px',
    options: getStatusOptions(t),
    advanced: false
  }
])

const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },
  {
    field: 'username',
    title: t('table.username'),
    sortable: false
  },
  {
    field: 'nickname',
    title: t('table.nickname'),
    sortable: false
  },
  {
    field: 'email',
    title: t('table.email'),
    sortable: false
  },
  {
    field: 'phone',
    title: t('table.phone'),
    sortable: false
  },
  {
    field: 'balance',
    title: t('user.balance'),
    width: 120,
    sortable: false,
    slot: 'balance'
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 100,
    sortable: false,
    slot: 'status'
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    width: 180,
    sortable: true
  },
  {
    field: 'operation',
    title: t('table.operation'),
    width: 200,
    fixed: 'right',
    slot: 'operation'
  }
])

const getPrimaryActions = (row) => {
  return [
    {
      label: t('common.edit'),
      action: 'edit',
      permission: 'user.update'
    },
    {
      label: t('user.update_balance'),
      action: 'updateBalance',
      permission: 'user.update',
      type: 'warning'
    }
  ]
}

const getMoreActions = (row) => {
  return [
    {
      label: t('user.balance_logs'),
      action: 'balanceLogs',
      permission: 'user.balance_logs'
    },
    {
      label: t('common.delete'),
      action: 'delete',
      permission: 'user.destroy',
      type: 'danger'
    }
  ]
}

const handleAction = async (action, row) => {
  switch (action) {
    case 'edit':
      handleEdit(row)
      break
    case 'updateBalance':
      handleUpdateBalance(row)
      break
    case 'balanceLogs':
      handleBalanceLogs(row)
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleUpdateBalance = (row) => {
  // 跳转到余额更新页面或打开对话框
  // 这里可以打开一个对话框来更新余额
  ElMessage.info('余额更新功能待实现')
}

const handleBalanceLogs = (row) => {
  // 跳转到余额变动记录页面
  window.location.href = `/user-balance-logs?user_id=${row.id}`
}

const handleDelete = async (row) => {
  await handleDeleteCrud(row)
}

const handleStatusChange = async (row, val) => {
  try {
    await updateUser(row.id, { status: val ? 1 : 0 })
    ElMessage.success(t('common.update_success'))
    loadData()
  } catch (error) {
    ErrorHandler.handle(error)
  }
}

const handleFormSuccess = () => {
  handleClose()
  loadData()
}

const formatMoney = (amount) => {
  return Number(amount).toFixed(2)
}

onMounted(() => {
  initDefaultSort()
  loadData()
})
</script>

<style scoped>
.user-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

