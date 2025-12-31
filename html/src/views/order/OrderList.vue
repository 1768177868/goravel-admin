<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.order') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('order.store').disabled"
            @click="handleAdd"
          >
            <el-icon><Plus /></el-icon>
            {{ $t('order.add_order') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="order"
        @search="handleSearch"
        @reset="handleReset"
      >
        <template #extra-buttons>
          <el-button 
            type="success" 
            :disabled="getButtonState('order.export').disabled"
            @click="handleExport"
          >
            {{ $t('common.export') }}
          </el-button>
        </template>
      </SearchForm>

      <!-- vxe-table with expand row -->
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
        <!-- 展开行：显示订单详情 -->
        <vxe-column type="expand" width="60">
          <template #content="{ row }">
            <div class="order-details-expand">
              <h4 style="margin: 10px 0; font-weight: bold;">{{ t('order.order_details') }}</h4>
              <vxe-table
                :data="row.details || row.Details || []"
                border
                size="small"
                :show-header="true"
              >
                <vxe-column field="product_name" :title="t('order.product_name')" width="200">
                  <template #default="{ row: detail }">
                    {{ detail.product_name || detail.ProductName || '-' }}
                  </template>
                </vxe-column>
                <vxe-column field="price" :title="t('order.price')" width="120">
                  <template #default="{ row: detail }">
                    {{ formatAmount(detail.price || detail.Price) }}
                  </template>
                </vxe-column>
                <vxe-column field="quantity" :title="t('order.quantity')" width="100">
                  <template #default="{ row: detail }">
                    {{ detail.quantity || detail.Quantity || 0 }}
                  </template>
                </vxe-column>
                <vxe-column field="subtotal" :title="t('order.subtotal')" width="120">
                  <template #default="{ row: detail }">
                    {{ formatAmount(detail.subtotal || detail.Subtotal) }}
                  </template>
                </vxe-column>
              </vxe-table>
            </div>
          </template>
        </vxe-column>
        <template v-for="(column, index) in tableColumns" :key="column.field || column.slot || index">
          <vxe-column
            v-if="column.type"
            :type="column.type"
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
              <el-tag :type="getStatusTagType(row.status || row.Status)">
                {{ getStatusText(row.status || row.Status) }}
              </el-tag>
            </template>
            <template v-else-if="column.slot === 'amount'" #default="{ row }">
              {{ formatAmount(row.amount || row.Amount) }}
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <TableActionButtons
                :row="row"
                :primary-actions="getPrimaryActions(row)"
                :get-button-state="getButtonState"
                @action="handleAction"
              />
            </template>
          </vxe-column>
        </template>
      </vxe-table>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
      />
    </el-card>

    <!-- 创建订单对话框 -->
    <OrderForm
      v-model="dialogVisible"
      @success="handleFormSuccess"
    />

    <!-- 编辑订单对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      :title="$t('order.edit_order')"
      width="600px"
      :close-on-click-modal="false"
      @close="handleEditDialogClose"
    >
      <el-form
        ref="editFormRef"
        :model="editFormData"
        :rules="editFormRules"
        label-width="120px"
      >
        <el-form-item :label="$t('order.status')" prop="status">
          <el-select
            v-model="editFormData.status"
            :placeholder="$t('order.update_status_tip')"
            style="width: 100%"
          >
            <el-option
              :label="$t('order.status_pending')"
              value="pending"
            />
            <el-option
              :label="$t('order.status_paid')"
              value="paid"
            />
            <el-option
              :label="$t('order.status_cancelled')"
              value="cancelled"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('order.remark')" prop="remark">
          <el-input
            v-model="editFormData.remark"
            type="textarea"
            :rows="4"
            :placeholder="$t('order.remark_placeholder')"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleEditDialogClose">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleEditSubmit" :loading="editSubmitting">
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 订单详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="$t('order.detail')"
      width="80%"
      :close-on-click-modal="false"
    >
      <div v-if="orderDetail" class="order-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item :label="$t('order.order_no')">
            {{ orderDetail.order?.order_no || orderDetail.order?.OrderNo }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.user_id')">
            {{ orderDetail.order?.user_id || orderDetail.order?.UserID }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.amount')">
            {{ formatAmount(orderDetail.order?.amount || orderDetail.order?.Amount) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.status')">
            <el-tag :type="getStatusTagType(orderDetail.order?.status || orderDetail.order?.Status)">
              {{ getStatusText(orderDetail.order?.status || orderDetail.order?.Status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.created_at')">
            {{ formatTime(orderDetail.order?.created_at || orderDetail.order?.CreatedAt) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.remark')">
            {{ orderDetail.order?.remark || orderDetail.order?.Remark || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider>{{ $t('order.details') }}</el-divider>

        <vxe-table
          :data="orderDetail.details || []"
          border
        >
          <vxe-column field="product_name" :title="$t('order.product_name')" />
          <vxe-column field="price" :title="$t('order.price')" :formatter="({ row }) => formatAmount(row.price || row.Price)" />
          <vxe-column field="quantity" :title="$t('order.quantity')" />
          <vxe-column field="subtotal" :title="$t('order.subtotal')" :formatter="({ row }) => formatAmount(row.subtotal || row.Subtotal)" />
        </vxe-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import OrderForm from './OrderForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import {
  getOrderList,
  getOrderDetail,
  updateOrder,
  deleteOrder,
  exportOrder,
  getExportStatus
} from '../../api/order'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'
import { getSevenDaysAgo } from '../../utils/dateUtils'
import { validateTimeRange, ORDER_MAX_TIME_RANGE_MONTHS } from '../../utils/timeRangeValidator'

// 权限控制
const { getButtonState } = usePermission()

const { t } = useI18n()
const router = useRouter()
const tableRef = ref(null)

// 使用 CRUD composable（只用于添加功能）
const {
  dialogVisible,
  editId,
  handleFormSuccess: handleFormSuccessCrud
} = useCrud()

// 订单详情对话框
const detailDialogVisible = ref(false)
const orderDetail = ref(null)

// 编辑订单对话框
const editDialogVisible = ref(false)
const editFormRef = ref(null)
const editSubmitting = ref(false)
const currentEditOrder = ref(null)
const editFormData = reactive({
  status: '',
  remark: ''
})

const editFormRules = computed(() => ({
  status: [
    { required: true, message: t('order.status_required'), trigger: 'change' }
  ]
}))

// 初始搜索表单数据（开始时间默认为一周前）
const initialSearchForm = {
  user_id: '',
  order_no: '',
  status: '',
  min_amount: null,
  max_amount: null,
  start_time: getSevenDaysAgo(),
  end_time: ''
}

// 使用列表页面 composable
const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch: handleSearchBase,
  handleReset,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getOrderList,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'created_at:desc',
  tableRef: computed(() => tableRef.value)
})

// 获取当前时间的字符串格式（YYYY-MM-DD HH:mm:ss）
const getCurrentTimeString = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  const seconds = String(now.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 验证时间范围的函数
const validateTimeRangeForSearch = () => {
  // 如果只填了开始时间，结束时间默认为当前时间
  if (searchForm.start_time) {
    const endTime = searchForm.end_time || getCurrentTimeString()
    const validation = validateTimeRange(searchForm.start_time, endTime, ORDER_MAX_TIME_RANGE_MONTHS)
    if (!validation.valid) {
      // 优先使用翻译键
      let errorMessage = validation.error
      if (validation.errorKey) {
        const translationKey = `order.${validation.errorKey}`
        if (validation.errorParams) {
          errorMessage = t(translationKey, validation.errorParams)
        } else {
          errorMessage = t(translationKey)
        }
      }
      ElMessage.warning(errorMessage)
      return false
    }
  }
  return true
}

// 监听开始时间和结束时间的变化，实时验证
const isInitialized = ref(false)
watch(
  () => [searchForm.start_time, searchForm.end_time],
  ([newStartTime, newEndTime], [oldStartTime, oldEndTime]) => {
    // 跳过初始化时的触发
    if (!isInitialized.value) {
      isInitialized.value = true
      return
    }
    
    // 只在开始时间或结束时间发生变化时验证
    if (newStartTime !== oldStartTime || newEndTime !== oldEndTime) {
      // 如果只填了开始时间，结束时间默认为当前时间
      if (newStartTime) {
        const endTime = newEndTime || getCurrentTimeString()
        const validation = validateTimeRange(newStartTime, endTime, ORDER_MAX_TIME_RANGE_MONTHS)
        if (!validation.valid) {
          // 优先使用翻译键
          let errorMessage = validation.error
          if (validation.errorKey) {
            const translationKey = `order.${validation.errorKey}`
            if (validation.errorParams) {
              errorMessage = t(translationKey, validation.errorParams)
            } else {
              errorMessage = t(translationKey)
            }
          }
          ElMessage.warning(errorMessage)
        }
      }
    }
  },
  { immediate: false }
)

// 包装搜索处理，添加时间范围验证
const handleSearch = () => {
  // 验证时间范围
  if (!validateTimeRangeForSearch()) {
    return
  }
  
  handleSearchBase()
}

// 表格列配置
const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },
  {
    field: 'order_no',
    title: t('order.order_no'),
    sortable: false
  },
  {
    field: 'user_id',
    title: t('order.user_id'),
    width: 100,
    sortable: false
  },
  {
    field: 'amount',
    title: t('order.amount'),
    width: 120,
    sortable: true,
    slot: 'amount'
  },
  {
    field: 'status',
    title: t('order.status'),
    width: 100,
    sortable: false,
    slot: 'status'
  },
  {
    field: 'created_at',
    title: t('order.created_at'),
    width: 180,
    sortable: true
  },
  {
    field: 'remark',
    title: t('order.remark'),
    sortable: false,
    width: 200,
    formatter: ({ cellValue }) => {
      return cellValue || '-'
    }
  },
  {
    title: t('table.operation'),
    width: 220,
    fixed: 'right',
    slot: 'operation'
  }
])

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'user_id',
    label: t('order.user_id'),
    type: 'input',
    width: '150px',
    advanced: false
  },
  {
    prop: 'order_no',
    label: t('order.order_no'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('order.status'),
    type: 'select',
    width: '150px',
    options: [
      { label: t('order.status_pending'), value: 'pending' },
      { label: t('order.status_paid'), value: 'paid' },
      { label: t('order.status_cancelled'), value: 'cancelled' }
    ],
    advanced: false
  },
  {
    prop: 'min_amount',
    label: t('order.min_amount'),
    type: 'number',
    width: '150px',
    advanced: true,
    min: 0,
    step: 0.01
  },
  {
    prop: 'max_amount',
    label: t('order.max_amount'),
    type: 'number',
    width: '150px',
    advanced: true,
    min: 0,
    step: 0.01
  },
  {
    prop: 'start_time',
    label: t('order.start_time'),
    type: 'datetime',
    width: '200px',
    advanced: true
  },
  {
    prop: 'end_time',
    label: t('order.end_time'),
    type: 'datetime',
    width: '200px',
    advanced: true
  }
])

// 格式化金额
const formatAmount = (amount) => {
  if (amount === null || amount === undefined) return '-'
  return `¥${Number(amount).toFixed(2)}`
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  return typeof time === 'string' ? time : new Date(time).toLocaleString('zh-CN')
}

// 获取状态文本
const getStatusText = (status) => {
  const statusMap = {
    'pending': t('order.status_pending'),
    'paid': t('order.status_paid'),
    'cancelled': t('order.status_cancelled')
  }
  return statusMap[status] || status || '-'
}

// 获取状态标签类型
const getStatusTagType = (status) => {
  const typeMap = {
    'pending': 'warning',
    'paid': 'success',
    'cancelled': 'danger'
  }
  return typeMap[status] || 'info'
}

// 查看详情
const handleView = async (row) => {
  try {

    const res = await getOrderDetail(row.id)
    
    if (res.data) {
      orderDetail.value = res.data
      detailDialogVisible.value = true
    }
  } catch (error) {
    logger.error('View order detail error:', error)
    ErrorHandler.handle(error, { silent: true })
  }
}

// 编辑订单（打开编辑对话框）
const handleEdit = async (row) => {
  try {
    // 获取订单详情
    const res = await getOrderDetail(row.id)
    if (res.data && res.data.order) {
      currentEditOrder.value = res.data.order
      // 填充表单数据
      editFormData.status = res.data.order.status || res.data.order.Status || 'pending'
      editFormData.remark = res.data.order.remark || res.data.order.Remark || ''
      editDialogVisible.value = true
    }
  } catch (error) {
    logger.error('Get order detail for edit error:', error)
    ErrorHandler.handle(error, { silent: true })
  }
}

// 编辑对话框关闭
const handleEditDialogClose = () => {
  editDialogVisible.value = false
  currentEditOrder.value = null
  editFormData.status = ''
  editFormData.remark = ''
  editFormRef.value?.resetFields()
  editFormRef.value?.clearValidate()
}

// 提交编辑
const handleEditSubmit = async () => {
  if (!editFormRef.value) return
  
  try {
    await editFormRef.value.validate()
  } catch (error) {
    return
  }
  
  if (!currentEditOrder.value) return
  
  editSubmitting.value = true
  try {
    await updateOrder(currentEditOrder.value.id || currentEditOrder.value.ID, {
      status: editFormData.status,
      remark: editFormData.remark || ''
    })
    
    ElMessage.success(t('common.update_success'))
    handleEditDialogClose()
    loadData()
  } catch (error) {
    logger.error('Update order error:', error)
    ErrorHandler.handle(error, { silent: true })
  } finally {
    editSubmitting.value = false
  }
}

// 删除订单
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      t('order.delete_confirm'),
      t('form.tip'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    
    await deleteOrder(row.id)
    
    ElMessage.success(t('common.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      logger.error('Delete order error:', error)
      ErrorHandler.handle(error, { silent: true })
    }
  }
}

// 获取主要操作按钮配置
const getPrimaryActions = (row) => {
  return [
    {
      key: 'view',
      label: t('common.view'),
      type: 'info',
      permission: 'order.show',
      handler: handleView
    },
    {
      key: 'edit',
      label: t('common.edit'),
      type: 'primary',
      permission: 'order.update',
      handler: handleEdit
    },
    {
      key: 'delete',
      label: t('common.delete'),
      type: 'danger',
      permission: 'order.destroy',
      handler: handleDelete
    }
  ]
}

// 处理操作事件
const handleAction = (command, row) => {
  switch (command) {
    case 'view':
      handleView(row)
      break
    case 'edit':
      handleEdit(row)
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

// 导出订单（异步）
const handleExport = async () => {
  try {
    // 提交导出任务
    const response = await exportOrder(searchForm)
    const exportId = response.data?.export_id || response.data?.data?.export_id
    
    if (!exportId) {
      ElMessage.error(t('order.export_failed') || '导出失败')
      return
    }

    // 显示提交成功消息
    ElMessage.success(t('order.export_task_submitted') || response.data?.message || '导出任务已提交，请稍后查看导出记录')
    
    // 立即跳转到导出记录页面
    router.push('/exports')
    
    // 开始轮询查询导出状态（在后台进行）
    const pollInterval = 3000 // 每3秒查询一次
    const maxPollTime = 300000 // 最多轮询5分钟
    const startTime = Date.now()
    
    const pollExportStatus = async () => {
      try {
        const statusResponse = await getExportStatus(exportId)
        const statusData = statusResponse.data?.data || statusResponse.data
        const status = statusData?.status
        
        if (status === 1) {
          // 导出成功
          ElMessage.success(t('order.export_success') || '导出成功')
          // 跳转到导出记录页面
          router.push('/exports')
          return
        } else if (status === 2) {
          // 导出失败
          const errorMsg = statusData?.error_msg || t('order.export_failed') || '导出失败'
          ElMessage.error(errorMsg)
          return
        } else if (status === 0) {
          // 处理中，继续轮询
          if (Date.now() - startTime < maxPollTime) {
            setTimeout(pollExportStatus, pollInterval)
          } else {
            // 超时，提示用户去导出记录页面查看
            ElMessage.warning(t('order.export_timeout') || '导出任务处理时间较长，请前往导出记录页面查看进度')
            router.push('/exports')
          }
        }
      } catch (error) {
        logger.error('Poll export status error:', error)
        // 如果查询失败，提示用户去导出记录页面查看
        ElMessage.warning(t('order.export_check_manually') || '请前往导出记录页面查看导出进度')
        router.push('/exports')
      }
    }
    
    // 延迟一下再开始轮询，给服务器一点处理时间
    setTimeout(pollExportStatus, 1000)
    
  } catch (error) {
    logger.error('Export order error:', error)
    
    // 检查是否是重复提交错误
    if (error.response?.status === 429) {
      ElMessage.warning(t('order.export_in_progress') || '导出任务正在处理中，请勿重复提交')
    } else {
      ErrorHandler.handle(error, { silent: true })
    }
  }
}

// 创建订单（打开创建对话框）
const handleAdd = () => {
  dialogVisible.value = true
}

// 表单提交成功回调
const handleFormSuccess = () => {
  handleFormSuccessCrud(loadData)
}

onMounted(async () => {
  try {
    await loadData()
    // 等待表格渲染完成后再初始化排序
    await nextTick()
    initDefaultSort()
  } catch (error) {
    logger.error('OrderList onMounted error:', error)
    ErrorHandler.handle(error)
  }
})
</script>

<style scoped>
.order-details-expand {
  padding: 15px;
  background-color: #f5f7fa;
}

.order-details-expand h4 {
  margin: 10px 0;
  font-weight: bold;
  color: #303133;
}


.order-detail {
  padding: 20px 0;
}
</style>

