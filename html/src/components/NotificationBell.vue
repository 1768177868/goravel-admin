<template>
  <el-popover
    v-model:visible="popoverVisible"
    placement="bottom-end"
    width="420"
    trigger="click"
    popper-class="notification-popover"
  >
    <template #reference>
      <el-badge
        :value="badgeValue"
        :hidden="notificationStore.unreadCount === 0"
        :offset="[-6, 10]"
      >
        <el-button type="text" class="header-btn bell-btn">
          <el-icon><Bell /></el-icon>
        </el-button>
      </el-badge>
    </template>

    <div class="notification-popover__header">
      <span>{{ $t('notification.center') }}</span>
      <div class="header-actions">
        <el-button size="small" text @click="handleMarkAll" :disabled="notificationStore.unreadCount === 0">
          {{ $t('notification.mark_all') }}
        </el-button>
        <el-button size="small" text @click="goList">
          {{ $t('notification.view_all') }}
        </el-button>
      </div>
    </div>

    <el-scrollbar
      class="notification-list"
      v-loading="notificationStore.loading"
      height="360px"
    >
      <div v-if="notificationStore.items.length === 0" class="notification-empty">
        <el-icon><Bell /></el-icon>
        <p>{{ $t('notification.empty') }}</p>
      </div>
      <div
        v-for="item in notificationStore.items"
        :key="item.id"
        class="notification-item"
        :class="{ unread: !item.is_read }"
      >
        <div class="notification-item__head">
          <span class="notification-type">{{ typeLabel(item.type) }}</span>
          <span class="notification-time">{{ formatTime(item.created_at) }}</span>
        </div>
        <div class="notification-item__title">{{ item.title }}</div>
        <div class="notification-item__content">{{ item.content }}</div>
        <div class="notification-item__actions">
          <el-tag v-if="!item.is_read" size="small" type="danger" effect="plain">
            {{ $t('notification.unread') }}
          </el-tag>
          <el-button
            v-if="!item.is_read"
            size="small"
            text
            @click="notificationStore.markAsRead(item.id)"
          >
            {{ $t('notification.mark_read') }}
          </el-button>
        </div>
      </div>
    </el-scrollbar>
  </el-popover>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'
import { Bell } from '@element-plus/icons-vue'
import { useNotificationStore } from '../store/notification'
import { useI18n } from 'vue-i18n'

dayjs.extend(relativeTime)

const notificationStore = useNotificationStore()
const router = useRouter()
const route = useRoute()
const popoverVisible = ref(false)
const badgeValue = computed(() => {
  const count = notificationStore.unreadCount || 0
  return count > 99 ? '99+' : count
})
const { t, locale } = useI18n()

const typeLabel = (type) => {
  if (type === 'message') {
    return t('notification.types.message')
  }
  if (type === 'notice') {
    return t('notification.types.notice')
  }
  return t('notification.types.announcement')
}

const handleMarkAll = () => {
  notificationStore.markAllRead()
}

const goList = () => {
  popoverVisible.value = false
  if (route.path !== '/notifications') {
    router.push('/notifications')
  }
}

const formatTime = (value) => {
  if (!value) return ''
  const currentLocale = locale.value === 'zh-CN' ? 'zh-cn' : 'en'
  return dayjs(value).locale(currentLocale).fromNow()
}

onMounted(() => {
  notificationStore.refresh({ limit: 7 })
  notificationStore.connect()
})

onBeforeUnmount(() => {
  notificationStore.disconnect()
})
</script>

<style scoped>
.bell-btn {
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bell-btn .el-icon {
  font-size: 20px;
}

/* 确保 el-badge 不会影响按钮大小 */
:deep(.el-badge) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

:deep(.el-badge__content) {
  font-size: 12px;
  height: 18px;
  line-height: 18px;
  padding: 0 6px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.notification-popover__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  font-weight: 600;
}

.notification-list {
  height: 360px;
}

.notification-popover :deep(.el-scrollbar__wrap) {
  max-height: 360px;
}

.notification-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px 0;
  color: #909399;
  gap: 8px;
}

.notification-item {
  padding: 12px;
  border-radius: 8px;
  border: 1px solid #ebeef5;
  margin-bottom: 10px;
  background: #fff;
}

.notification-item.unread {
  border-color: #ffd04b;
  background: #fffdf5;
}

.notification-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.notification-type {
  font-weight: 600;
  color: #409eff;
}

.notification-item__title {
  font-weight: 600;
  margin-bottom: 4px;
  color: #303133;
}

.notification-item__content {
  font-size: 13px;
  color: #606266;
  line-height: 1.4;
}

.notification-item__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}
</style>

<style>
/* 通知弹窗夜间模式适配 - 需要非 scoped 样式来覆盖组件内部样式 */
.dark-mode .notification-popover {
  background-color: var(--card-bg) !important;
  border-color: var(--card-border) !important;
}

.dark-mode .notification-popover .notification-popover__header {
  color: var(--text-color-primary) !important;
}

.dark-mode .notification-item {
  background-color: var(--card-bg) !important;
  border-color: var(--border-color-light) !important;
  color: var(--text-color-primary) !important;
}

.dark-mode .notification-item.unread {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--sidebar-active) !important;
}

.dark-mode .notification-item__head {
  color: var(--text-color-secondary) !important;
}

.dark-mode .notification-type {
  color: var(--sidebar-active) !important;
}

.dark-mode .notification-item__title {
  color: var(--text-color-primary) !important;
}

.dark-mode .notification-item__content {
  color: var(--text-color-regular) !important;
}

.dark-mode .notification-empty {
  color: var(--text-color-secondary) !important;
}
</style>


