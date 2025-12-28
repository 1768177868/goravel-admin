<template>
  <div class="user-balance-log-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('user.balance_logs') }}</span>
        </div>
      </template>

      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item :label="$t('user.user_id')">
          <el-input v-model="searchForm.user_id" :placeholder="$t('user.user_id_placeholder')" style="width: 200px" />
        </el-form-item>
        <el-form-item :label="$t('user.type')">
          <el-select v-model="searchForm.type" :placeholder="$t('user.type_placeholder')" clearable style="width: 150px">
            <el-option label="收入" value="income" />
            <el-option label="支出" value="expense" />
            <el-option label="退款" value="refund" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('user.source')">
          <el-select v-model="searchForm.source" :placeholder="$t('user.source_placeholder')" clearable style="width: 150px">
            <el-option label="订单" value="order" />
            <el-option label="充值" value="recharge" />
            <el-option label="提现" value="withdraw" />
            <el-option label="手动" value="manual" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">{{ $t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>

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

      <!-- 表格 -->
      <vxe-table
        :data="tableData"
        :loading="loading"
        border
        height="500"
      >
        <vxe-column field="id" :title="$t('table.id')" width="80" />
        <vxe-column field="type" :title="$t('user.type')" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.type === 'income'" type="success">收入</el-tag>
            <el-tag v-else-if="row.type === 'expense'" type="danger">支出</el-tag>
            <el-tag v-else-if="row.type === 'refund'" type="warning">退款</el-tag>
          </template>
        </vxe-column>
        <vxe-column field="amount" :title="$t('user.amount')" width="120">
          <template #default="{ row }">
            <span :style="{ color: row.type === 'expense' ? '#f56c6c' : '#67c23a' }">
              {{ row.type === 'expense' ? '-' : '+' }}¥{{ formatMoney(row.amount) }}
            </span>
          </template>
        </vxe-column>
        <vxe-column field="balance" :title="$t('user.balance')" width="120">
          <template #default="{ row }">
            <span style="color: #409EFF; font-weight: bold;">¥{{ formatMoney(row.balance) }}</span>
          </template>
        </vxe-column>
        <vxe-column field="source" :title="$t('user.source')" width="100" />
        <vxe-column field="description" :title="$t('user.description')" min-width="200" />
        <vxe-column field="created_at" :title="$t('table.created_at')" width="180" />
      </vxe-table>

      <!-- 分页 -->
      <Pagination
        v-model="pagination"
        @page-change="handlePageChange"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import Pagination from '../../components/Pagination.vue'
import { getUserBalanceLogList, getUserBalanceStatistics } from '../../api/userBalanceLog'
import ErrorHandler from '../../utils/errorHandler'

const { t } = useI18n()
const route = useRoute()

const loading = ref(false)
const tableData = ref([])
const statistics = ref(null)

const searchForm = reactive({
  user_id: route.query.user_id || '',
  type: '',
  source: '',
  start_time: '',
  end_time: ''
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const loadData = async () => {
  if (!searchForm.user_id) {
    return
  }

  loading.value = true
  try {
    const params = {
      user_id: searchForm.user_id,
      page: pagination.page,
      page_size: pagination.page_size,
      ...searchForm
    }
    const res = await getUserBalanceLogList(params)
    if (res.code === 0) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    ErrorHandler.handle(error)
  } finally {
    loading.value = false
  }
}

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
    if (res.code === 0) {
      statistics.value = res.data.statistics
    }
  } catch (error) {
    ErrorHandler.handle(error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
  loadStatistics()
}

const handleReset = () => {
  searchForm.type = ''
  searchForm.source = ''
  searchForm.start_time = ''
  searchForm.end_time = ''
  handleSearch()
}

const handlePageChange = () => {
  loadData()
}

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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
  color: #67c23a;
}

.stat-item .value.expense {
  color: #f56c6c;
}

.stat-item .value.refund {
  color: #e6a23c;
}

.stat-item .value.balance {
  color: #409EFF;
}
</style>

