<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.payment') }}</span>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="payment"
        @search="handleSearch"
        @reset="handleReset"
      >
        <template #extra-buttons>
          <el-button 
            type="success" 
            :disabled="getButtonState('payment.export').disabled || isExporting"
            :loading="isExporting"
            @click="handleExport"
          >
            <el-icon><Download /></el-icon>
            {{ $t('common.export') }}
          </el-button>
        </template>
      </SearchForm>

      <VxeTable
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="600"
        @sort-change="handleSortChange"
      >
        <template #status="{ row }">
          <el-tag
            :type="getStatusType(row.status || row.Status)"
            size="small"
          >
            {{ getStatusText(row.status || row.Status) }}
          </el-tag>
        </template>

        <template #payment_method="{ row }">
          {{ getPaymentMethodName(row.payment_method || row.PaymentMethod) }}
        </template>

        <template #amount="{ row }">
          {{ formatAmount(row.amount || row.Amount) }}
        </template>

        <template #operation="{ row }">
          <el-button
            type="primary"
            size="small"
            :disabled="getButtonState('payment.show').disabled"
            @click="handleView(row)"
          >
            {{ $t('common.view') }}
          </el-button>
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
      />
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="$t('payment.payment_detail')"
      width="800px"
    >
      <div v-loading="detailLoading">
        <el-descriptions :column="2" border v-if="paymentDetail">
          <el-descriptions-item :label="$t('payment.payment_no')">
            {{ paymentDetail.payment_no || paymentDetail.PaymentNo }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('payment.order_no')">
            {{ paymentDetail.order_no || paymentDetail.OrderNo }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('payment.payment_method')">
            {{ getPaymentMethodName(paymentDetail.payment_method || paymentDetail.PaymentMethod) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('payment.user_id')">
            {{ paymentDetail.user_id || paymentDetail.UserID }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('payment.amount')">
            {{ formatAmount(paymentDetail.amount || paymentDetail.Amount) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('payment.status')">
            <el-tag
              :type="getStatusType(paymentDetail.status || paymentDetail.Status)"
              size="small"
            >
              {{ getStatusText(paymentDetail.status || paymentDetail.Status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('payment.third_party_no')">
            {{ paymentDetail.third_party_no || paymentDetail.ThirdPartyNo || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('payment.pay_time')">
            {{ paymentDetail.pay_time || paymentDetail.PayTime || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('payment.fail_reason')" :span="2" v-if="paymentDetail.fail_reason || paymentDetail.FailReason">
            {{ paymentDetail.fail_reason || paymentDetail.FailReason }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('payment.remark')" :span="2" v-if="paymentDetail.remark || paymentDetail.Remark">
            {{ paymentDetail.remark || paymentDetail.Remark }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('table.created_at')">
            {{ formatDateTime(paymentDetail.created_at || paymentDetail.CreatedAt) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('table.updated_at')">
            {{ formatDateTime(paymentDetail.updated_at || paymentDetail.UpdatedAt) }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import { usePermission } from '../../composables/usePermission'
import { useListPage } from '../../composables/useListPage'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import { getPaymentList, getPaymentDetail, exportPayments } from '../../api/payment'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'
import { getSevenDaysAgo } from '../../utils/dateUtils'

// 权限控制
const { getButtonState } = usePermission()

const { t } = useI18n()
const router = useRouter()
const tableRef = ref(null)
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const paymentDetail = ref(null)

// 导出相关状态
const isExporting = ref(false)

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
  fetchApi: getPaymentList,
  initialSearchForm: {
    payment_no: '',
    order_no: '',
    payment_method_id: '',
    user_id: '',
    status: '',
    start_time: getSevenDaysAgo(),
    end_time: ''
  },
  fieldMapping: {},
  defaultSort: 'created_at:desc',
  tableRef: computed(() => tableRef.value?.tableRef)
})

// 初始搜索表单（开始时间默认为一周前）
const initialSearchForm = {
  payment_no: '',
  order_no: '',
  payment_method_id: '',
  user_id: '',
  status: '',
  start_time: getSevenDaysAgo(),
  end_time: ''
}

// 搜索字段配置
const searchFields = computed(() => [
  {
    prop: 'payment_no',
    type: 'input',
    placeholder: t('payment.payment_no_placeholder')
  },
  {
    prop: 'order_no',
    type: 'input',
    placeholder: t('payment.order_no_placeholder')
  },
  {
    prop: 'payment_method_id',
    type: 'select',
    placeholder: t('payment.payment_method_id_placeholder'),
    apiUrl: '/options?type=payment_method',
    filterable: true
  },
  {
    prop: 'user_id',
    type: 'input',
    placeholder: t('payment.user_id_placeholder')
  },
  {
    prop: 'status',
    type: 'select',
    placeholder: t('payment.status_placeholder'),
    options: [
      { label: t('payment.status_pending'), value: 'pending' },
      { label: t('payment.status_paid'), value: 'paid' },
      { label: t('payment.status_failed'), value: 'failed' },
      { label: t('payment.status_cancelled'), value: 'cancelled' }
    ]
  },
  {
    prop: 'start_time',
    type: 'datetime',
    placeholder: t('payment.start_time_placeholder')
  },
  {
    prop: 'end_time',
    type: 'datetime',
    placeholder: t('payment.end_time_placeholder')
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
    field: 'payment_no',
    title: t('payment.payment_no'),
    minWidth: 300,
    sortable: true
  },
  {
    field: 'order_no',
    title: t('payment.order_no'),
    minWidth: 180,
    sortable: true
  },
  {
    field: 'payment_method',
    title: t('payment.payment_method'),
    width: 150,
    slot: 'payment_method'
  },
  {
    field: 'user_id',
    title: t('payment.user_id'),
    width: 100,
    sortable: true
  },
  {
    field: 'amount',
    title: t('payment.amount'),
    width: 120,
    align: 'right',
    slot: 'amount',
    sortable: true
  },
  {
    field: 'status',
    title: t('payment.status'),
    width: 100,
    slot: 'status',
    sortable: true
  },
  {
    field: 'pay_time',
    title: t('payment.pay_time'),
    width: 180,
    sortable: true
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
    width: 100,
    fixed: 'right',
    slot: 'operation',
    sortable: false
  }
])

// 获取状态类型
const getStatusType = (status) => {
  const statusMap = {
    pending: 'warning',
    paid: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return statusMap[status] || undefined
}

// 获取状态文本
const getStatusText = (status) => {
  const statusMap = {
    pending: t('payment.status_pending'),
    paid: t('payment.status_paid'),
    failed: t('payment.status_failed'),
    cancelled: t('payment.status_cancelled')
  }
  return statusMap[status] || status
}

// 获取支付方式名称
const getPaymentMethodName = (paymentMethod) => {
  if (!paymentMethod) return '-'
  if (typeof paymentMethod === 'string') return paymentMethod
  if (typeof paymentMethod === 'object') {
    return paymentMethod.name || paymentMethod.Name || paymentMethod.code || paymentMethod.Code || '-'
  }
  return '-'
}

// 格式化金额
const formatAmount = (amount) => {
  if (amount === null || amount === undefined) return '0.00'
  return Number(amount).toFixed(2)
}

// 格式化日期时间
const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  if (typeof dateTime === 'string') {
    return dateTime.replace('T', ' ').substring(0, 19)
  }
  return dateTime
}

// 查看详情（使用支付单号查询，因为分表后ID可能重复）
const handleView = async (row) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  try {
    const paymentNo = row.payment_no || row.PaymentNo
    const response = await getPaymentDetail(paymentNo)
    paymentDetail.value = response.data?.data || response.data || {}
  } catch (error) {
    logger.error('Load payment detail error:', error)
    ErrorHandler.handle(error)
  } finally {
    detailLoading.value = false
  }
}

// 导出支付记录
const handleExport = async () => {
  if (isExporting.value) {
    return
  }

  isExporting.value = true
  try {
    const params = {
      ...searchForm,
      order_by: 'created_at:desc'
    }
    const response = await exportPayments(params)
    const exportId = response.data?.data?.export_id || response.data?.export_id
    
    if (!exportId) {
      ElMessage.error(t('export.failed'))
      isExporting.value = false
      return
    }

    // 显示提交成功消息
    ElMessage.success(t('export.task_submitted'))
    
    // 立即跳转到导出记录页面
    router.push('/exports')
  } catch (error) {
    logger.error('Export payments error:', error)
    
    // 检查是否是重复提交错误
    if (error.response?.status === 429) {
      ElMessage.warning(t('common.export_in_progress'))
    } else if (!error.__handled) {
      ErrorHandler.handle(error, { silent: true })
    }
  } finally {
    isExporting.value = false
  }
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

