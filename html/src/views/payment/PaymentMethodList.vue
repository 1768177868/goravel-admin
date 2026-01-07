<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.payment_method') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('payment_method.store').disabled"
            @click="handleAdd"
          >
            <el-icon><PlusIcon /></el-icon>
            {{ $t('payment_method.add_payment_method') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="payment_method"
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
        <template #is_active="{ row }">
          <!-- <el-switch
            :model-value="row.is_active || row.IsActive"
            :disabled="getButtonState('payment_method.update').disabled"
            @change="(val) => handleStatusChange(row, val)"
          /> -->
          <el-tag :type="row.is_active || row.IsActive ? 'success' : 'danger'">{{ row.is_active || row.IsActive ? t('common.enabled') : t('common.disabled') }}</el-tag>
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

    <!-- 添加/编辑对话框 -->
    <PaymentMethodForm
      ref="paymentMethodFormRef"
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, computed, markRaw, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import PaymentMethodForm from './PaymentMethodForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import {
  getPaymentMethodList,
  deletePaymentMethod,
  updatePaymentMethod
} from '../../api/paymentMethod'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

// 使用 markRaw 标记图标组件
const PlusIcon = markRaw(Plus)

// 权限控制
const { getButtonState } = usePermission()

const { t } = useI18n()
const tableRef = ref(null)
const paymentMethodFormRef = ref(null)

const {
  dialogVisible,
  editId,
  handleAdd,
  handleClose,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: deletePaymentMethod
})

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
  fetchApi: getPaymentMethodList,
  initialSearchForm: {
    name: '',
    code: '',
    type: '',
    is_active: '',
    description: ''
  },
  fieldMapping: {},
  defaultSort: 'sort:asc,id:desc',
  tableRef: computed(() => tableRef.value?.tableRef)
})

// 初始搜索表单
const initialSearchForm = {
  name: '',
  code: '',
  type: '',
  is_active: '',
  description: ''
}

// 搜索字段配置
const searchFields = computed(() => [
  {
    prop: 'name',
    type: 'input',
    placeholder: t('payment_method.name_placeholder')
  },
  {
    prop: 'code',
    type: 'input',
    placeholder: t('payment_method.code_placeholder')
  },
  {
    prop: 'type',
    type: 'select',
    placeholder: t('payment_method.type_placeholder'),
    options: [
      { label: t('payment_method.type_wechat'), value: 'wechat' },
      { label: t('payment_method.type_alipay'), value: 'alipay' },
      { label: t('payment_method.type_qq'), value: 'qq' },
      { label: t('payment_method.type_allinpay'), value: 'allinpay' },
      { label: t('payment_method.type_lakala'), value: 'lakala' },
      { label: t('payment_method.type_paypal'), value: 'paypal' },
      { label: t('payment_method.type_apple'), value: 'apple' },
      { label: t('payment_method.type_saobei'), value: 'saobei' }
    ]
  },
  {
    prop: 'is_active',
    type: 'select',
    placeholder: t('payment_method.is_active_placeholder'),
    options: [
      { label: t('common.enabled'), value: '1' },
      { label: t('common.disabled'), value: '0' }
    ]
  },
  {
    prop: 'description',
    type: 'input',
    placeholder: t('payment_method.description_placeholder')
  }
])

// 表格列配置
const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },
  {
    field: 'name',
    title: t('payment_method.name'),
    minWidth: 150,
    sortable: true
  },
  {
    field: 'code',
    title: t('payment_method.code'),
    width: 120,
    sortable: true
  },
  {
    field: 'type',
    title: t('payment_method.type'),
    width: 120,
    sortable: true
  },
  {
    field: 'is_active',
    title: t('table.status'),
    width: 100,
    slot: 'is_active',
  },
  {
    field: 'sort',
    title: t('table.sort'),
    width: 100,
    sortable: true
  },
  {
    field: 'description',
    title: t('table.description'),
    minWidth: 200
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
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

// 操作按钮配置
const operationActions = computed(() => [
  {
    key: 'edit',
    label: t('common.edit'),
    type: 'primary',
    permission: 'payment_method.update',
    handler: handleEdit
  },
  {
    key: 'delete',
    label: t('common.delete'),
    type: 'danger',
    permission: 'payment_method.destroy',
    handler: handleDelete
  }
])

// 编辑
const handleEdit = (row) => {
  editId.value = row.id || row.ID
  dialogVisible.value = true
}

// 删除
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      t('payment_method.delete_confirm'),
      t('form.warning'),
      {
        type: 'warning',
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel')
      }
    )
    await handleDeleteCrud(row.id || row.ID)
    ElMessage.success(t('common.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      ErrorHandler.handle(error)
    }
  }
}

// 状态变更
const handleStatusChange = async (row, val) => {
  try {
    const id = row.id || row.ID
    await updatePaymentMethod(id, {
      name: row.name || row.Name,
      is_active: val ? 1 : 0,
      sort: row.sort || row.Sort || 0,
      description: row.description || row.Description || ''
    })
    ElMessage.success(t('common.update_success'))
    loadData()
  } catch (error) {
    ErrorHandler.handle(error)
    loadData() // 重新加载以恢复原状态
  }
}

// 表单成功回调
const handleFormSuccess = () => {
  handleClose()
  loadData()
}

// 初始化
onMounted(() => {
  initDefaultSort()
  loadData()
})
</script>

<style scoped>
.list-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

