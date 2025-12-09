<template>
  <div class="notification-page">
    <el-card class="notification-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('notification.center') }}</span>
          <div class="header-actions">
            <el-select
              v-model="filterType"
              size="small"
              style="width: 120px; margin-right: 10px"
              @change="handleTypeChange"
              clearable
              :placeholder="$t('notification.table.type')"
            >
              <el-option
                :label="$t('common.all')"
                value=""
              />
              <el-option
                :label="$t('notification.types.announcement')"
                value="announcement"
              />
              <el-option
                :label="$t('notification.types.notice')"
                value="notice"
              />
              <el-option
                :label="$t('notification.types.message')"
                value="message"
              />
            </el-select>
            <el-select
              v-model="filterIsRead"
              size="small"
              style="width: 120px; margin-right: 10px"
              @change="handleIsReadChange"
              clearable
              :placeholder="$t('notification.table.status')"
            >
              <el-option
                :label="$t('common.all')"
                value=""
              />
              <el-option
                :label="$t('notification.unread')"
                value="false"
              />
              <el-option
                :label="$t('notification.read')"
                value="true"
              />
            </el-select>
            <el-button size="small" @click="loadData">
              {{ $t('tabs.refresh') }}
            </el-button>
            <el-button
              size="small"
              type="primary"
              plain
              @click="handleMarkAll"
              :disabled="notificationStore.unreadCount === 0"
            >
              {{ $t('notification.mark_all') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="list"
        border
        style="width: 100%"
      >
        <el-table-column
          prop="title"
          :label="$t('notification.table.title')"
          min-width="160"
        />
        <el-table-column
          prop="content"
          :label="$t('notification.table.content')"
          min-width="260"
          show-overflow-tooltip
        />
        <el-table-column
          prop="type"
          :label="$t('notification.table.type')"
          width="120"
        >
          <template #default="{ row }">
            <el-tag size="small">
              {{ typeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="is_read"
          :label="$t('notification.table.status')"
          width="120"
        >
          <template #default="{ row }">
            <el-tag
              size="small"
              :type="row.is_read ? 'info' : 'danger'"
              effect="plain"
            >
              {{ row.is_read ? $t('notification.read') : $t('notification.unread') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="created_at"
          :label="$t('notification.table.created_at')"
          width="180"
        >
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('common.operation')"
          width="140"
        >
          <template #default="{ row }">
            <el-button
              v-if="!row.is_read"
              type="primary"
              link
              @click="handleMarkRead(row)"
            >
              {{ $t('notification.mark_read') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          layout="total, prev, pager, next, sizes"
          :page-sizes="[10, 20, 30, 50]"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useNotificationStore } from '../../store/notification'
import { fetchNotifications } from '../../api/notification'

const { t } = useI18n()
const notificationStore = useNotificationStore()

const list = ref([])
const loading = ref(false)
const filterType = ref('')
const filterIsRead = ref('')
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    // 如果选择了类型筛选，添加到参数中
    if (filterType.value) {
      params.type = filterType.value
    }
    // 如果选择了已读/未读筛选，添加到参数中
    if (filterIsRead.value !== '') {
      params.is_read = filterIsRead.value
    }
    const { data } = await fetchNotifications(params)
    list.value = data.notifications || []
    pagination.total = data.pagination?.total || list.value.length
    notificationStore.unreadCount = data.unread_count || 0
  } catch (error) {
    console.error('Load notifications list failed:', error)
    // 如果错误已经在响应拦截器中处理过，就不再重复显示
    if (!error.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('error.default')
      ElMessage.error(errorMessage)
    }
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page) => {
  pagination.page = page
  loadData()
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  loadData()
}

const handleTypeChange = (value) => {
  // 处理清空筛选器的情况（value 可能为 null）
  filterType.value = value || ''
  pagination.page = 1
  loadData()
}

const handleIsReadChange = (value) => {
  // 处理清空筛选器的情况（value 可能为 null 或 undefined）
  filterIsRead.value = value || ''
  pagination.page = 1
  loadData()
}

const handleMarkRead = async (row) => {
  await notificationStore.markAsRead(row.id)
  row.is_read = true
  row.read_at = new Date().toISOString()
}

const handleMarkAll = async () => {
  await notificationStore.markAllRead()
  list.value = list.value.map(item => ({ ...item, is_read: true }))
}

const formatDate = (value) => {
  if (!value) {
    return ''
  }
  return dayjs(value).format('YYYY-MM-DD HH:mm:ss')
}

const typeLabel = (type) => {
  if (type === 'message') {
    return t('notification.types.message')
  }
  if (type === 'notice') {
    return t('notification.types.notice')
  }
  return t('notification.types.announcement')
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.notification-page {
  padding: 16px;
}

.notification-card {
  width: 100%;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.pagination-wrapper {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>

