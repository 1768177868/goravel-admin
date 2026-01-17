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
          <div v-html="row.content" class="rich-text-content"></div>
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
    <el-dialog
      v-model="dialogVisible"
      :title="$t('notification.create')"
      width="800px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item
          :label="$t('notification.table.type')"
          prop="type"
        >
          <el-radio-group v-model="formData.type">
            <el-radio value="announcement">
              {{ $t('notification.types.announcement') }}
            </el-radio>
            <el-radio value="notice">
              {{ $t('notification.types.notice') }}
            </el-radio>
            <el-radio value="message">
              {{ $t('notification.types.message') }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          v-if="formData.type === 'message'"
          :label="$t('notification.receiver')"
          prop="receiver_id"
        >
          <el-select
            v-model="formData.receiver_id"
            :placeholder="$t('notification.select_receiver')"
            filterable
            style="width: 100%"
            clearable
          >
            <el-option
              v-for="admin in adminOptions"
              :key="admin.value"
              :label="admin.label"
              :value="admin.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          :label="$t('notification.table.title')"
          prop="title"
        >
          <el-input
            v-model="formData.title"
            :placeholder="$t('notification.title_placeholder')"
            maxlength="150"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('notification.table.content')"
          prop="content"
        >
          <WangEditor
            v-model="formData.content"
            :placeholder="$t('notification.content_placeholder')"
            :height="300"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ $t('common.cancel') }}
        </el-button>
        <el-button
          type="primary"
          :loading="submitting"
          @click="handleSubmit"
        >
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, computed } from 'vue'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import WangEditor from '../../components/WangEditor.vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import { useListPage } from '../../composables/useListPage'
import { buildSearchParams } from '../../utils/buildSearchParams'
import { useNotificationStore } from '../../store/notification'
import { useUserStore } from '../../store/user'
import { fetchNotifications, createNotification } from '../../api/notification'
import { getOptions } from '../../api/option'

const { t } = useI18n()
const notificationStore = useNotificationStore()
const userStore = useUserStore()

// 权限检查
const canCreate = computed(() => userStore.shouldShowButton('notification.store'))

const tableRef = ref(null)
const formRef = ref(null)
const dialogVisible = ref(false)
const submitting = ref(false)
const adminOptions = ref([])

const formData = reactive({
  type: 'announcement',
  receiver_id: '',
  title: '',
  content: ''
})

const formRules = {
  type: [
    { required: true, message: t('notification.type_required'), trigger: 'change' }
  ],
  receiver_id: [
    {
      validator: (rule, value, callback) => {
        if (formData.type === 'message' && !value) {
          callback(new Error(t('notification.receiver_required')))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ],
  title: [
    { required: true, message: t('notification.title_required'), trigger: 'blur' },
    { max: 150, message: t('notification.title_max_length'), trigger: 'blur' }
  ],
  content: [
    { required: true, message: t('notification.content_required'), trigger: 'blur' }
  ]
}

// 字段名映射：前端字段名 -> 数据库字段名（只包含不同的字段）
const fieldMapping = {} // 所有字段名都相同，无需映射

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

const loadAdminOptions = async () => {
  try {
    const res = await getOptions('admin')
    if (res.data && res.data.options) {
      adminOptions.value = res.data.options
    }
  } catch (error) {
    console.error('Load admin options error:', error)
  }
}

// 监听类型变化，如果不是私信则清空接收者
watch(() => formData.type, (newType) => {
  if (newType !== 'message') {
    formData.receiver_id = ''
    // 清除接收者字段的验证
    if (formRef.value) {
      formRef.value.clearValidate('receiver_id')
    }
  }
})

const handleAdd = () => {
  dialogVisible.value = true
  // 重置表单
  formData.type = 'announcement'
  formData.receiver_id = ''
  formData.title = ''
  formData.content = ''

  // 清除验证（延迟执行，确保表单已渲染）
  setTimeout(() => {
    if (formRef.value) {
      formRef.value.clearValidate()
    }
  }, 100)
}

const handleDialogClose = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

const handleSubmit = async () => {
  if (!formRef.value) {
    return
  }
  
  await formRef.value.validate(async (valid) => {
    if (!valid) {
      return false
    }
    
    submitting.value = true
    try {
      const data = {
        type: formData.type,
        title: formData.title.trim(),
        content: formData.content
      }
      
      // 如果是私信，必须添加接收者ID
      if (formData.type === 'message') {
        if (!formData.receiver_id) {
          ElMessage.error(t('notification.receiver_required'))
          submitting.value = false
          return
        }
        data.receiver_id = formData.receiver_id
      }
      // 公告和通知不传receiver_id，后端会发送给所有人
      
      await createNotification(data)
      ElMessage.success(t('notification.create_success'))
      dialogVisible.value = false
      
      // 重置到第一页并重新加载列表
      pagination.page = 1
      await loadData()
      
      // 刷新未读数量
      await notificationStore.fetchUnread()
    } catch (error) {
      console.error('Create notification error:', error)
      if (!error.__handled) {
        const errorMessage = error.response?.data?.message || error.message || t('error.default')
        ElMessage.error(errorMessage)
      }
    } finally {
      submitting.value = false
    }
  })
}

onMounted(() => {
  loadData()
  loadAdminOptions()
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

.rich-text-content {
  max-height: 100px;
  overflow-y: auto;
  white-space: normal;
}

.rich-text-content :deep(p) {
  margin: 0 0 10px;
}

.rich-text-content :deep(img) {
  max-width: 100%;
  height: auto;
}
</style>
