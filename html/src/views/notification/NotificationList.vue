<template>
  <div class="notification-page">
    <el-card class="notification-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('notification.center') }}</span>
          <div class="header-actions">
            <el-button
              type="primary"
              :disabled="!canCreate"
              @click="handleAdd"
            >
              <el-icon><Plus /></el-icon>
              {{ $t('notification.create') }}
            </el-button>
            <el-button
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

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        :loading="loading"
        i18n-prefix="notification"
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
        <template #content="{ row }">
          <div class="text-truncate" :title="extractTextFromMarkdown(row.content)">
            {{ extractTextFromMarkdown(row.content) }}
          </div>
        </template>

        <template #type="{ row }">
          <el-tag size="small">
            {{ typeLabel(row.type) }}
          </el-tag>
        </template>

        <template #sender="{ row }">
          <template v-if="row.type === 'message' && row.sender_id === userStore.adminInfo?.id">
            <span class="text-gray-500">
              {{ $t('notification.sent_to') }}: 
              <span v-if="row.receiver" class="text-blue-600">
                {{ row.receiver.nickname || row.receiver.username }}
              </span>
              <span v-else class="text-gray-400">-</span>
            </span>
          </template>
          <template v-else>
            <span v-if="row.sender">
              {{ row.sender.nickname || row.sender.username }}
            </span>
            <span v-else class="text-gray-400">
              {{ $t('notification.system') }}
            </span>
          </template>
        </template>

        <template #is_read="{ row }">
          <el-tag
            size="small"
            :type="row.is_read ? 'info' : 'danger'"
            effect="plain"
          >
            {{ row.is_read ? $t('notification.read') : $t('notification.unread') }}
          </el-tag>
        </template>

        <template #created_at="{ row }">
          {{ formatDate(row.created_at) }}
        </template>

        <template #operation="{ row }">
          <el-button
            type="primary"
            link
            @click="handleView(row)"
          >
            {{ $t('common.view') }}
          </el-button>
          <el-button
            v-if="!row.is_read"
            type="primary"
            link
            @click="handleMarkRead(row)"
          >
            {{ $t('notification.mark_read') }}
          </el-button>
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
      />
    </el-card>

    <!-- 创建通知对话框 -->
    <NotificationForm
      v-model="dialogVisible"
      @success="handleFormSuccess"
    />

    <!-- 查看通知详情对话框 -->
    <el-dialog
      v-model="viewDialogVisible"
      :title="$t('notification.detail')"
      width="800px"
    >
      <div v-if="currentNotification">
        <h3 class="text-lg font-bold mb-4" style="font-size: 1.125rem; font-weight: 700; margin-bottom: 1rem;">{{ currentNotification.title }}</h3>
        <div class="flex items-center text-gray-500 mb-4 text-sm" style="display: flex; align-items: center; color: #6b7280; margin-bottom: 1rem; font-size: 0.875rem;">
          <el-tag size="small" class="mr-4" style="margin-right: 1rem;">{{ typeLabel(currentNotification.type) }}</el-tag>
          <span class="mr-4" style="margin-right: 1rem;">{{ formatDate(currentNotification.created_at) }}</span>
          <span>{{ $t('notification.table.sender') }}: {{ currentNotification.sender?.nickname || currentNotification.sender?.username || $t('notification.system') }}</span>
        </div>
        <div class="rich-text-content-view markdown-content" v-html="renderMarkdown(currentNotification.content)"></div>
      </div>
      <template #footer>
        <el-button @click="viewDialogVisible = false">
          {{ $t('common.close') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import dayjs from 'dayjs'
import { Plus } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import NotificationForm from './NotificationForm.vue'
import { useListPage } from '../../composables/useListPage'
import { buildSearchParams } from '../../utils/buildSearchParams'
import { useNotificationStore } from '../../store/notification'
import { useUserStore } from '../../store/user'
import { fetchNotifications } from '../../api/notification'
import { getApiBaseURL } from '../../utils/env'
import { renderContent, extractTextFromMarkdown } from '../../utils/markdown'

const { t } = useI18n()
const notificationStore = useNotificationStore()
const userStore = useUserStore()

// 权限检查
const canCreate = computed(() => userStore.shouldShowButton('notification.store'))

const viewDialogVisible = ref(false)
const currentNotification = ref(null)

const handleView = (row) => {
  // 深拷贝 row，避免直接修改原始数据
  const notification = { ...row }

  currentNotification.value = notification
  viewDialogVisible.value = true
  
  // 如果未读，标记为已读
  if (!row.is_read) {
    handleMarkRead(row)
  }
}

// 渲染内容（自动判断 HTML 或 Markdown）
const renderMarkdown = (content) => {
  if (!content) return ''
  return renderContent(content, 'auto')
}

const tableRef = ref(null)
const dialogVisible = ref(false)

// 初始搜索表单数据
const initialSearchForm = {
  type: '',
  is_read: ''
}

// 搜索字段配置
const searchFields = computed(() => [
  {
    prop: 'type',
    label: t('notification.table.type'),
    type: 'select',
    options: [
      { label: t('common.all'), value: '' },
      { label: t('notification.types.announcement'), value: 'announcement' },
      { label: t('notification.types.notice'), value: 'notice' },
      { label: t('notification.types.message'), value: 'message' }
    ],
    clearable: true
  },
  {
    prop: 'is_read',
    label: t('notification.table.status'),
    type: 'select',
    options: [
      { label: t('common.all'), value: '' },
      { label: t('notification.unread'), value: 'false' },
      { label: t('notification.read'), value: 'true' }
    ],
    clearable: true
  }
])

// 表格列配置
const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    // sortable: true
  },
  {
    field: 'title',
    title: t('notification.table.title'),
    minWidth: 160
  },
  {
    field: 'content',
    title: t('notification.table.content'),
    minWidth: 260,
    slot: 'content'
  },
  {
    field: 'type',
    title: t('notification.table.type'),
    width: 150,
    slot: 'type'
  },
  {
    field: 'sender',
    title: t('notification.table.sender'),
    width: 160,
    slot: 'sender'
  },
  {
    field: 'is_read',
    title: t('notification.table.status'),
    width: 120,
    slot: 'is_read'
  },
  {
    field: 'created_at',
    title: t('notification.table.created_at'),
    width: 180,
    sortable: true,
    slot: 'created_at'
  },
  {
    field: 'operation',
    title: t('common.operation'),
    width: 140,
    fixed: 'right',
    slot: 'operation'
  }
])

// 自定义 fetchApi 包装函数（适配 notifications API 的返回格式）
const fetchNotificationsWrapper = async (params) => {
  const res = await fetchNotifications(params)
  // 将 notifications API 的返回格式转换为 useListPage 期望的格式
  if (res && res.data) {
    return {
      ...res,
      data: {
        list: res.data.notifications || [],
        total: res.data.pagination?.total || 0,
        unread_count: res.data.unread_count || 0,
        pagination: res.data.pagination
      }
    }
  }
  return res
}

// 数据转换函数（适配 API 返回格式）
const transformNotificationData = (notification) => {
  return {
    id: notification.id || notification.ID,
    type: notification.type || notification.Type || '',
    title: notification.title || notification.Title || '',
    content: notification.content || notification.Content || '',
    sender: notification.sender || notification.Sender || null,
    sender_id: notification.sender_id || notification.SenderID || null,
    receiver: notification.receiver || notification.Receiver || null,
    receiver_id: notification.receiver_id || notification.ReceiverID || null,
    is_read: notification.is_read || notification.IsRead || false,
    read_at: notification.read_at || notification.ReadAt || null,
    created_at: notification.created_at || notification.CreatedAt || ''
  }
}

// 自定义参数构建函数（只处理 is_read 的特殊转换，其他字段由 buildSearchParams 统一处理）
const buildNotificationParams = (searchForm, baseParams) => {
  // 先使用 buildSearchParams 处理所有字段（包括 trim 等）
  const params = buildSearchParams(searchForm, baseParams)
  
  // 特殊处理 is_read：字符串 'true'/'false' 转换为布尔值
  if (searchForm.is_read !== '' && searchForm.is_read !== null && searchForm.is_read !== undefined) {
    if (searchForm.is_read === 'true') {
      params.is_read = true
    } else if (searchForm.is_read === 'false') {
      params.is_read = false
    }
    // 如果已经是布尔值或其他类型，保持原样（buildSearchParams 已处理）
  }
  
  return params
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
  fetchApi: fetchNotificationsWrapper,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef),
  transformData: transformNotificationData,
  buildParams: buildNotificationParams, // 使用自定义参数构建函数
  onLoadSuccess: (res) => {
    // 更新未读数量
    if (res && res.data && res.data.unread_count !== undefined) {
      notificationStore.unreadCount = res.data.unread_count || 0
    }
  }
})

const handleMarkRead = async (row) => {
  await notificationStore.markAsRead(row.id)
  row.is_read = true
  row.read_at = new Date().toISOString()
}

const handleMarkAll = async () => {
  await notificationStore.markAllRead()
  // 重新加载数据
  await loadData()
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

const handleAdd = () => {
  dialogVisible.value = true
}

const handleFormSuccess = async () => {
  // 重置到第一页并重新加载列表
  pagination.page = 1
  await loadData()
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

.text-truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  width: 100%;
  color: #606266;
}

/* 暗黑模式样式 */
html.dark .text-truncate {
  color: var(--el-text-color-regular, #cfd3dc);
}

.rich-text-content-view {
  min-height: 100px;
  max-height: 60vh;
  overflow-y: auto;
  white-space: normal;
  padding: 10px;
  border: 1px solid #eee;
  border-radius: 4px;
}

/* 暗黑模式样式 */
html.dark .rich-text-content-view {
  background-color: var(--el-bg-color, #1d1e1f);
  border-color: var(--el-border-color, #3d3e40);
  color: var(--el-text-color-regular, #cfd3dc);
}

.rich-text-content-view :deep(p) {
  margin: 0 0 10px;
}

.rich-text-content-view :deep(img) {
  max-width: 100%;
  height: auto;
  display: block;
  margin: 10px 0;
}

.markdown-content :deep(p) {
  margin: 0 0 10px;
  line-height: 1.6;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4),
.markdown-content :deep(h5),
.markdown-content :deep(h6) {
  margin: 16px 0 8px;
  font-weight: 600;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 10px 0;
  padding-left: 30px;
}

.markdown-content :deep(li) {
  margin: 4px 0;
}

.markdown-content :deep(blockquote) {
  margin: 10px 0;
  padding: 10px 15px;
  border-left: 4px solid var(--el-color-primary);
  background-color: #f5f7fa;
  color: #606266;
}

.markdown-content :deep(code) {
  background-color: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 0.9em;
}

.markdown-content :deep(pre) {
  background-color: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 10px 0;
}

.markdown-content :deep(pre code) {
  background-color: transparent;
  padding: 0;
}

.markdown-content :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 10px 0;
}

.markdown-content :deep(th),
.markdown-content :deep(td) {
  border: 1px solid #ebeef5;
  padding: 8px 12px;
  text-align: left;
}

.markdown-content :deep(th) {
  background-color: #f5f7fa;
  font-weight: 600;
}
</style>
