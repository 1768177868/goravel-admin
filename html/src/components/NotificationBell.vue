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
        <el-button size="small" text @click="handleMarkAll" :disabled="currentUnreadCount === 0">
          {{ $t('notification.mark_all') }}
        </el-button>
        <el-button size="small" text @click="goList">
          {{ $t('notification.view_all') }}
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="notification-tabs">
      <el-tab-pane :label="$t('notification.types.announcement')" name="announcement">
        <el-scrollbar
          class="notification-list"
          v-loading="notificationStore.loading"
          height="360px"
        >
          <div v-if="filteredItems.announcement.length === 0" class="notification-empty">
            <el-icon><Bell /></el-icon>
            <p>{{ $t('notification.empty') }}</p>
          </div>
          <div
            v-for="item in filteredItems.announcement"
            :key="item.id"
            class="notification-item"
            :class="{ unread: !item.is_read }"
          >
            <div class="notification-item__head">
              <span class="notification-type">{{ typeLabel(item.type) }}</span>
              <span class="notification-time">{{ formatTime(item.created_at) }}</span>
            </div>
            <div class="notification-item__title">{{ item.title }}</div>
            <div class="notification-item__content markdown-content" v-html="renderMarkdown(item.content)"></div>
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
      </el-tab-pane>
      <el-tab-pane :label="$t('notification.types.notice')" name="notice">
        <el-scrollbar
          class="notification-list"
          v-loading="notificationStore.loading"
          height="360px"
        >
          <div v-if="filteredItems.notice.length === 0" class="notification-empty">
            <el-icon><Bell /></el-icon>
            <p>{{ $t('notification.empty') }}</p>
          </div>
          <div
            v-for="item in filteredItems.notice"
            :key="item.id"
            class="notification-item"
            :class="{ unread: !item.is_read }"
          >
            <div class="notification-item__head">
              <span class="notification-type">{{ typeLabel(item.type) }}</span>
              <span class="notification-time">{{ formatTime(item.created_at) }}</span>
            </div>
            <div class="notification-item__title">{{ item.title }}</div>
            <div class="notification-item__content markdown-content" v-html="renderMarkdown(item.content)"></div>
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
      </el-tab-pane>
      <el-tab-pane :label="$t('notification.types.message')" name="message">
        <el-scrollbar
          class="notification-list"
          v-loading="notificationStore.loading"
          height="360px"
        >
          <div v-if="filteredItems.message.length === 0" class="notification-empty">
            <el-icon><Bell /></el-icon>
            <p>{{ $t('notification.empty') }}</p>
          </div>
          <div
            v-for="item in filteredItems.message"
            :key="item.id"
            class="notification-item"
            :class="{ unread: !item.is_read }"
          >
            <div class="notification-item__head">
              <span class="notification-type">{{ typeLabel(item.type) }}</span>
              <span class="notification-time">{{ formatTime(item.created_at) }}</span>
            </div>
            <div class="notification-item__title">{{ item.title }}</div>
            <div class="notification-item__content markdown-content" v-html="renderMarkdown(item.content)"></div>
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
      </el-tab-pane>
    </el-tabs>
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
import { renderContent } from '../utils/markdown'

dayjs.extend(relativeTime)

const notificationStore = useNotificationStore()
const router = useRouter()
const route = useRoute()
const popoverVisible = ref(false)
const activeTab = ref('announcement')
const badgeValue = computed(() => {
  const count = notificationStore.unreadCount || 0
  return count > 99 ? '99+' : count
})
const { t, locale } = useI18n()

// 按类型过滤通知
const filteredItems = computed(() => {
  const items = notificationStore.items || []
  return {
    announcement: items.filter(item => item.type === 'announcement'),
    notice: items.filter(item => item.type === 'notice'),
    message: items.filter(item => item.type === 'message')
  }
})

// 当前 tab 的未读数量
const currentUnreadCount = computed(() => {
  const items = filteredItems.value[activeTab.value] || []
  return items.filter(item => !item.is_read).length
})

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
  // 只标记当前 tab 的未读通知为已读
  const items = filteredItems.value[activeTab.value] || []
  const unreadItems = items.filter(item => !item.is_read)
  if (unreadItems.length === 0) {
    return
  }
  // 批量标记当前 tab 的未读通知
  Promise.all(unreadItems.map(item => notificationStore.markAsRead(item.id)))
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

// 渲染内容（自动判断 HTML 或 Markdown）
const renderMarkdown = (content) => {
  if (!content) return ''
  return renderContent(content, 'auto')
}

onMounted(() => {
  notificationStore.refresh({ limit: 20 })
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

.notification-tabs {
  margin-top: 8px;
}

.notification-tabs :deep(.el-tabs__header) {
  margin: 0 0 8px 0;
}

.notification-tabs :deep(.el-tabs__nav-wrap) {
  padding: 0;
}

.notification-tabs :deep(.el-tabs__item) {
  padding: 0 12px;
  font-size: 13px;
}

.notification-tabs :deep(.el-tabs__content) {
  padding: 0;
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
  color: var(--el-color-primary);
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
  /* 富文本样式处理 */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.notification-item__content :deep(p) {
  margin: 0;
  display: inline;
}

.notification-item__content :deep(img) {
  display: none; /* 列表预览不显示图片 */
}

.markdown-content :deep(p) {
  margin: 0;
  display: inline;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4),
.markdown-content :deep(h5),
.markdown-content :deep(h6) {
  margin: 0;
  display: inline;
  font-weight: 600;
}

.notification-item__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

/* 暗黑模式样式 */
:deep(.notification-popover) {
  background-color: var(--el-bg-color, #1d1e1f);
}

:deep(.notification-popover .notification-popover__header) {
  color: var(--el-text-color-primary, #e5eaf3);
}

:deep(.notification-popover .notification-empty) {
  color: var(--el-text-color-secondary, #a3a6ad);
}

:deep(.notification-popover .notification-item) {
  background: var(--el-bg-color, #1d1e1f);
  border-color: var(--el-border-color, #3d3e40);
}

:deep(.notification-popover .notification-item.unread) {
  border-color: var(--el-color-warning, #e6a23c);
  background: var(--el-color-warning-light-9, rgba(230, 162, 60, 0.1));
}

:deep(.notification-popover .notification-item__head) {
  color: var(--el-text-color-secondary, #a3a6ad);
}

:deep(.notification-popover .notification-item__title) {
  color: var(--el-text-color-primary, #e5eaf3);
}

:deep(.notification-popover .notification-item__content) {
  color: var(--el-text-color-regular, #cfd3dc);
}

:deep(.notification-popover .notification-tabs) {
  color: var(--el-text-color-primary, #e5eaf3);
}

:deep(.notification-popover .el-tabs__item) {
  color: var(--el-text-color-regular, #cfd3dc);
}

:deep(.notification-popover .el-tabs__item.is-active) {
  color: var(--el-color-primary, #409eff);
}
</style>

