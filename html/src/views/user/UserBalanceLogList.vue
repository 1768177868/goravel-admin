<template>
  <div class="user-balance-log-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-button 
              type="primary" 
              link 
              :icon="ArrowLeft" 
              @click="handleBack"
              style="margin-right: 10px;"
            >
              {{ $t('user.back') }}
            </el-button>
            <span>{{ $t('user.balance_logs') }}</span>
            <span v-if="searchForm.user_id" class="user-id-display">({{ $t('user.user_id') }}: {{ searchForm.user_id }})</span>
          </div>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchValues"
        i18n-prefix="user"
        :loading="loading"
        @search="handleSearch"
        @reset="handleReset"
      />

      <!-- 统计信息 -->
      <el-card v-if="statistics" class="statistics-card" shadow="never">
        <div class="statistics">
          <div class="stat-item">
            <span class="label">{{ $t('user.total_income') }}:</span>
            <span class="value income">¥{{ formatMoney(statistics.total_income) }}</span>
          </div>
          <div class="stat-item">
            <span class="label">{{ $t('user.total_expense') }}:</span>
            <span class="value expense">¥{{ formatMoney(statistics.total_expense) }}</span>
          </div>
          <div class="stat-item">
            <span class="label">{{ $t('user.total_refund') }}:</span>
            <span class="value refund">¥{{ formatMoney(statistics.total_refund) }}</span>
          </div>
          <div class="stat-item">
            <span class="label">{{ $t('user.current_balance') }}:</span>
            <span class="value balance">¥{{ formatMoney(statistics.current_balance) }}</span>
          </div>
        </div>
      </el-card>

      <TableToolbar
        :on-refresh="loadData"
        :visible-columns="visibleColumns"
        :all-columns="allTableColumns"
        :default-visible-columns="defaultVisibleColumns"
        :column-order="columnOrder"
        :fixed-columns="fixedColumns"
        :on-column-setting-confirm="handleColumnSettingConfirm"
      />

      <VxeTable
        :key="`table-${tableColumns.length}-${JSON.stringify(tableColumns.map(c => c.field || c.slot || c.key))}`"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="500"
      >
        <template #type="{ row }">
          <el-tag v-if="row.type === 'income'" type="success">{{ $t('user.balance_income') }}</el-tag>
          <el-tag v-else-if="row.type === 'expense'" type="danger">{{ $t('user.balance_expense') }}</el-tag>
          <el-tag v-else-if="row.type === 'refund'" type="warning">{{ $t('user.balance_refund') }}</el-tag>
        </template>

        <template #amount="{ row }">
          <span :style="{ color: row.type === 'expense' ? 'var(--el-color-danger)' : 'var(--el-color-success)' }">
            {{ row.type === 'expense' ? '-' : '+' }}¥{{ formatMoney(row.amount) }}
          </span>
        </template>

        <template #balance="{ row }">
          <span style="color: var(--el-color-primary); font-weight: bold;">¥{{ formatMoney(row.balance) }}</span>
        </template>

        <template #source="{ row }">
          <span v-if="row.source === 'order'">{{ $t('user.source_order') }}</span>
          <span v-else-if="row.source === 'recharge'">{{ $t('user.source_recharge') }}</span>
          <span v-else-if="row.source === 'withdraw'">{{ $t('user.source_withdraw') }}</span>
          <span v-else-if="row.source === 'manual'">{{ $t('user.source_manual') }}</span>
          <span v-else>{{ row.source }}</span>
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableToolbar from '../../components/TableToolbar.vue'
import { useListPage } from '../../composables/useListPage'
import { useColumnSetting } from '../../composables/useColumnSetting'
import { getUserBalanceLogList, getUserBalanceStatistics } from '../../api/userBalanceLog'
import ErrorHandler from '../../utils/errorHandler'
import { forOwn } from 'lodash-es'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const userId = route.query.user_id || ''

// 初始搜索值
const initialSearchValues = {
  user_id: userId,
  type: '',
  source: '',
  start_time: '',
  end_time: ''
}

const {
  pagination,
  tableData,
  loading,
  searchForm
} = useListPage({
  fetchApi: getUserBalanceLogList,
  initialSearchForm: initialSearchValues,
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => null)
})

// 重写 loadData 以支持自定义参数构建
const loadData = async () => {
  const params = {
    page: pagination.page,
    page_size: pagination.pageSize,
    user_id: searchForm.user_id
  }
  
  // 只添加有值的搜索条件
  if (searchForm.type) params.type = searchForm.type
  if (searchForm.source) params.source = searchForm.source
  if (searchForm.start_time) params.start_time = searchForm.start_time
  if (searchForm.end_time) params.end_time = searchForm.end_time
  
  loading.value = true
  try {
    const res = await getUserBalanceLogList(params)
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    ErrorHandler.handle(error)
  } finally {
    loading.value = false
  }
}

// 设置初始分页大小
pagination.pageSize = 20

// 统计信息
const statistics = ref(null)

// 所有列配置（用于列设置）
const allTableColumns = computed(() => [
  { field: 'id', title: t('table.id'), width: 180, key: 'id' },
  { field: 'type', title: t('user.type'), width: 100, slot: 'type', key: 'type' },
  { field: 'amount', title: t('user.amount'), width: 120, slot: 'amount', key: 'amount' },
  { field: 'balance', title: t('user.balance'), width: 120, slot: 'balance', key: 'balance' },
  { field: 'source', title: t('user.source'), width: 100, slot: 'source', key: 'source' },
  { field: 'description', title: t('user.description'), minWidth: 200, key: 'description' },
  { field: 'created_at', title: t('table.created_at'), width: 180, key: 'created_at' }
])

const {
  tableColumns,
  visibleColumns,
  defaultVisibleColumns,
  columnOrder,
  fixedColumns,
  handleColumnSettingConfirm
} = useColumnSetting('user_balance_log', allTableColumns)

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'type',
    label: t('user.type'),
    type: 'select',
    width: '150px',
    options: [
      { label: t('user.balance_income'), value: 'income' },
      { label: t('user.balance_expense'), value: 'expense' },
      { label: t('user.balance_refund'), value: 'refund' }
    ],
    clearable: true,
    advanced: false
  },
  {
    prop: 'source',
    label: t('user.source'),
    type: 'select',
    width: '150px',
    options: [
      { label: t('user.source_order'), value: 'order' },
      { label: t('user.source_recharge'), value: 'recharge' },
      { label: t('user.source_withdraw'), value: 'withdraw' },
      { label: t('user.source_manual'), value: 'manual' }
    ],
    clearable: true,
    advanced: false
  }
])

// 加载统计数据
const loadStatistics = async () => {
  if (!searchForm.user_id) {
    return
  }

  try {
    const params = {
      user_id: searchForm.user_id,
      start_time: searchForm.start_time,
      end_time: searchForm.end_time
    }
    const res = await getUserBalanceStatistics(params)
    if (res.code === 200) {
      statistics.value = res.data.statistics
    }
  } catch (error) {
    ErrorHandler.handle(error)
  }
}

// 搜索处理（同时加载列表和统计）
const handleSearch = () => {
  pagination.page = 1
  loadData()
  loadStatistics()
}

// 重置处理（同时加载列表和统计）
const handleReset = () => {
  // 重置搜索表单
  forOwn(searchForm, (value, key) => {
    searchForm[key] = initialSearchValues[key] !== undefined ? initialSearchValues[key] : ''
  })
  pagination.page = 1
  loadData()
  loadStatistics()
}

// 返回按钮处理
const handleBack = () => {
  // 返回上一页，如果没有历史记录则返回用户列表
  if (window.history.length > 1) {
    router.go(-1)
  } else {
    router.push('/users')
  }
}

// 格式化金额
const formatMoney = (amount) => {
  return Number(amount || 0).toFixed(2)
}

onMounted(() => {
  if (searchForm.user_id) {
    loadData()
    loadStatistics()
  }
})
</script>

<style scoped>
.user-balance-log-list {
  padding: 20px;
}

.header-left {
  display: flex;
  align-items: center;
}

.user-id-display {
  margin-left: 10px;
  color: var(--text-color-secondary);
  font-size: 14px;
}

.search-form {
  margin-bottom: 20px;
}

.statistics-card {
  margin-bottom: 20px;
}

.statistics {
  display: flex;
  gap: 30px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.stat-item .label {
  font-weight: 500;
}

.stat-item .value {
  font-size: 18px;
  font-weight: bold;
}

.stat-item .value.income {
  color: var(--el-color-success);
}

.stat-item .value.expense {
  color: var(--el-color-danger);
}

.stat-item .value.refund {
  color: var(--el-color-warning);
}

.stat-item .value.balance {
  color: var(--el-color-primary);
}
</style>

