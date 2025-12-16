<template>
  <div class="notification-page">
    <el-card class="notification-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('notification.center') }}</span>
          <div class="header-actions">
            <el-select
              v-model="filterType"
              style="width: 150px; margin-right: 10px"
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
              style="width: 150px; margin-right: 10px"
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
            <el-button @click="loadData">
              {{ $t('tabs.refresh') }}
            </el-button>
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
          prop="sender"
          :label="$t('notification.table.sender')"
          width="150"
        >
          <template #default="{ row }">
            <template v-if="row.type === 'message' && row.sender_id === userStore.adminInfo?.id">
              <!-- 自己发送的私信，显示接收人 -->
              <span class="text-gray-500">
                {{ $t('notification.sent_to') }}: 
                <span v-if="row.receiver" class="text-blue-600">
                  {{ row.receiver.nickname || row.receiver.username }}
                </span>
                <span v-else class="text-gray-400">-</span>
              </span>
            </template>
            <template v-else>
              <!-- 接收的私信或其他类型通知，显示发送人 -->
              <span v-if="row.sender">
                {{ row.sender.nickname || row.sender.username }}
              </span>
              <span v-else class="text-gray-400">
                {{ $t('notification.system') }}
              </span>
            </template>
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

    <!-- 创建通知对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="$t('notification.create')"
      width="600px"
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
          <el-input
            v-model="formData.content"
            type="textarea"
            :placeholder="$t('notification.content_placeholder')"
            :rows="6"
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
import { useNotificationStore } from '../../store/notification'
import { useUserStore } from '../../store/user'
import { fetchNotifications, createNotification } from '../../api/notification'
import { getOptions } from '../../api/option'

const { t } = useI18n()
const notificationStore = useNotificationStore()
const userStore = useUserStore()

// 权限检查
const canCreate = computed(() => userStore.shouldShowButton('notification.store'))

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
        content: formData.content.trim()
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
      // 重新加载列表
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

.pagination-wrapper {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>

