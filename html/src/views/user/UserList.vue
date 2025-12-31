<template>
  <div class="user-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.user') }}</span>
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
        :initial-values="initialSearchForm"
        i18n-prefix="user"
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
      >
        <template #status="{ row }">
          <el-switch
            :model-value="Number(row.status ?? row.Status ?? 1) === 1"
            :disabled="getButtonState('user.update').disabled"
            @change="(val) => handleStatusChange(row, val)"
          />
        </template>

        <template #balance="{ row }">
          <span style="color: #409EFF; font-weight: bold;">
            {{ formatBalance(row.balance || 0, row.currency) }}
          </span>
        </template>

        <template #operation="{ row }">
          <TableActionButtons
            :row="row"
            :primary-actions="getPrimaryActions(row)"
            :more-actions="getMoreActions(row)"
            :get-button-state="getButtonState"
            @action="handleAction"
          />
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
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
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import UserForm from './UserForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import { getStatusOptions } from '../../utils/fieldOptions'
import {
  getUserList,
  deleteUser,
  updateUser,
  resetPassword,
  updateBalance
} from '../../api/user'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

const PlusIcon = markRaw(Plus)

const { getButtonState } = usePermission()
const { t } = useI18n()
const router = useRouter()
const tableRef = ref(null)
const userFormRef = ref(null)

const {
  dialogVisible,
  editId,
  handleAdd,
  handleClose,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: deleteUser
})

// 转换用户数据（确保字段名统一）
const transformUserData = (user) => {
  return {
    id: user.id || user.ID,
    username: user.username || user.Username || '',
    nickname: user.nickname || user.Nickname || '',
    email: user.email || user.Email || '',
    phone: user.phone || user.Phone || '',
    balance: user.balance || user.Balance || 0,
    currency: user.currency || user.Currency || null,
    status: user.status !== undefined ? user.status : (user.Status !== undefined ? user.Status : 1),
    created_at: user.created_at || user.CreatedAt || ''
  }
}

// 初始搜索表单
const initialSearchForm = {
  username: '',
  email: '',
  phone: '',
  status: ''
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
  fetchApi: getUserList,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef),
  transformData: transformUserData
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
    sortable: true,
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
    width: 220,
    fixed: 'right',
    slot: 'operation'
  }
])

const getPrimaryActions = (row) => {
  return [
    {
      key: 'edit',
      label: t('common.edit'),
      permission: 'user.update'
    },
    {
      key: 'updateBalance',
      label: t('user.update_balance'),
      permission: 'user.update',
      type: 'warning'
    }
  ]
}

const getMoreActions = (row) => {
  return [
    {
      key: 'resetPassword',
      command: 'resetPassword',
      label: t('user.reset_password'),
      permission: 'user.password'
    },
    {
      key: 'balanceLogs',
      command: 'balanceLogs',
      label: t('user.balance_logs'),
      permission: 'user.balance_logs',
      divided: true
    },
    {
      key: 'delete',
      command: 'delete',
      label: t('common.delete'),
      permission: 'user.destroy',
      type: 'danger'
    }
  ]
}

const handleAction = async (command, row) => {
  switch (command) {
    case 'edit':
      handleEdit(row)
      break
    case 'updateBalance':
      handleUpdateBalance(row)
      break
    case 'resetPassword':
      handleResetPassword(row)
      break
    case 'balanceLogs':
      handleBalanceLogs(row)
      break
    case 'delete':
      handleDelete(row)
      break
    default:
      logger.warn('Unknown action command:', command)
  }
}

const handleEdit = (row) => {
  editId.value = row.id || row.ID
  dialogVisible.value = true
}

const handleUpdateBalance = async (row) => {
  try {
    // 第一步：输入金额
    const { value: amountStr } = await ElMessageBox.prompt(
      t('user.update_balance_prompt'),
      t('user.update_balance'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        inputType: 'number',
        inputPlaceholder: t('user.amount_placeholder'),
        inputValidator: (value) => {
          if (!value || Number(value) === 0) {
            return t('user.amount_required')
          }
          return true
        }
      }
    )

    const amount = Number(amountStr)
    if (amount === 0) {
      ElMessage.error(t('user.amount_cannot_be_zero'))
      return
    }

    // 第二步：选择变动类型（使用 prompt 输入编号）
    const typeOptions = [
      { value: 'income', label: t('user.balance_income') },
      { value: 'expense', label: t('user.balance_expense') },
      { value: 'refund', label: t('user.balance_refund') }
    ]

    const typeMessage = `${t('user.select_change_type_prompt')}\n\n${typeOptions.map((opt, idx) => `${idx + 1}. ${opt.label}`).join('\n')}`
    
    const { value: typeIndex } = await ElMessageBox.prompt(
      typeMessage,
      t('user.select_change_type'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        inputType: 'number',
        inputPlaceholder: '1-3',
        inputValidator: (value) => {
          const num = Number(value)
          if (!value || num < 1 || num > 3) {
            return t('user.invalid_selection')
          }
          return true
        }
      }
    )

    const selectedIndex = Number(typeIndex) - 1
    if (selectedIndex < 0 || selectedIndex >= typeOptions.length) {
      return
    }

    const logType = typeOptions[selectedIndex].value

    // 第三步：确认操作
    await ElMessageBox.confirm(
      `${t('user.balance_update_confirm')}\n${t('user.amount')}: ${Math.abs(amount).toFixed(2)}\n${t('user.change_type')}: ${typeOptions[selectedIndex].label}`,
      t('common.confirm'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )

    // 执行更新
    const userId = row.id || row.ID
    await updateBalance(userId, {
      amount: Math.abs(amount),
      type: logType,
      source: 'manual',
      description: t('user.manual_balance_adjustment')
    })
    ElMessage.success(t('user.balance_update_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ErrorHandler.handle(error)
    }
  }
}

const handleResetPassword = async (row) => {
  try {
    const { value: password } = await ElMessageBox.prompt(t('user.new_password'), t('user.reset_password'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      inputType: 'password',
      inputValidator: (value) => {
        if (!value || value.length < 6) {
          return t('user.password_min_length')
        }
        return true
      }
    })
    const userId = row.id || row.ID
    await resetPassword(userId, { password })
    ElMessage.success(t('user.reset_password_success'))
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      logger.error('Reset password error:', error)
      ErrorHandler.handle(error, { silent: true })
    }
  }
}

const handleBalanceLogs = (row) => {
  // 跳转到余额变动记录页面
  const userId = row.id || row.ID
  router.push({
    path: '/user-balance-logs',
    query: { user_id: userId }
  })
}

const handleDelete = async (row) => {
  await handleDeleteCrud(row,loadData)
}

const handleStatusChange = async (row, val) => {
  try {
    const userId = row.id || row.ID
    await updateUser(userId, { status: val ? 1 : 0 })
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

const formatBalance = (amount, currency) => {
  const symbol = currency?.symbol || '¥'
  const decimalPlaces = currency?.decimal_places ?? 2
  const formatted = Number(amount).toFixed(decimalPlaces)
  return `${symbol}${formatted}`
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

</style>

