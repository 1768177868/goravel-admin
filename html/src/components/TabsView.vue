<template>
  <div class="tabs-view" v-if="tabsStore.hasTabs">
    <el-tabs
      v-model="activeTab"
      type="card"
      closable
      @tab-remove="handleRemove"
      @tab-click="handleClick"
      class="tabs-container"
    >
      <el-tab-pane
        v-for="tab in tabsStore.tabs"
        :key="tab.path"
        :label="getTabTitle(tab)"
        :name="tab.path"
      >
        <template #label>
          <span
            class="tab-label"
            @contextmenu.prevent="showContextMenu($event, tab.path)"
          >
            <span class="tab-title">{{ getTabTitle(tab) }}</span>
            <el-icon
              v-if="tab.path === activeTab"
              class="refresh-icon"
              @click.stop="handleRefresh(tab.path)"
              title="刷新"
            >
              <Refresh />
            </el-icon>
          </span>
        </template>
      </el-tab-pane>
    </el-tabs>

    <!-- 右键菜单 -->
    <teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="context-menu-overlay"
        @click="contextMenuVisible = false"
      ></div>
      <div
        v-if="contextMenuVisible"
        class="context-menu"
        :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
        @click.stop
      >
        <div
          class="context-menu-item"
          @click="handleContextMenu({ action: 'refresh', path: contextMenuPath })"
        >
          <el-icon><Refresh /></el-icon>
          <span>{{ $t('tabs.refresh') }}</span>
        </div>
        <div
          class="context-menu-item"
          @click="handleContextMenu({ action: 'close', path: contextMenuPath })"
        >
          <el-icon><Close /></el-icon>
          <span>{{ $t('tabs.close') }}</span>
        </div>
        <div
          class="context-menu-item"
          :class="{ disabled: tabsStore.tabs.length <= 1 }"
          @click="tabsStore.tabs.length > 1 && handleContextMenu({ action: 'closeOther', path: contextMenuPath })"
        >
          <el-icon><CircleClose /></el-icon>
          <span>{{ $t('tabs.closeOther') }}</span>
        </div>
        <div
          class="context-menu-item"
          :class="{ disabled: !canCloseLeft }"
          @click="canCloseLeft && handleContextMenu({ action: 'closeLeft', path: contextMenuPath })"
        >
          <el-icon><ArrowLeft /></el-icon>
          <span>{{ $t('tabs.closeLeft') }}</span>
        </div>
        <div
          class="context-menu-item"
          :class="{ disabled: !canCloseRight }"
          @click="canCloseRight && handleContextMenu({ action: 'closeRight', path: contextMenuPath })"
        >
          <el-icon><ArrowRight /></el-icon>
          <span>{{ $t('tabs.closeRight') }}</span>
        </div>
        <div
          class="context-menu-item"
          :class="{ disabled: tabsStore.tabs.length === 0 }"
          @click="tabsStore.tabs.length > 0 && handleContextMenu({ action: 'closeAll', path: contextMenuPath })"
        >
          <el-icon><Delete /></el-icon>
          <span>{{ $t('tabs.closeAll') }}</span>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useTabsStore } from '../store/tabs'
import {
  Refresh,
  Close,
  CircleClose,
  ArrowLeft,
  ArrowRight,
  Delete
} from '@element-plus/icons-vue'

const router = useRouter()
const { t } = useI18n()
const tabsStore = useTabsStore()

const activeTab = computed({
  get: () => tabsStore.activeTab,
  set: (val) => tabsStore.setActiveTab(val)
})

const contextMenuRef = ref(null)
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuPath = ref('')

const canCloseLeft = computed(() => {
  if (!contextMenuPath.value) return false
  const index = tabsStore.tabs.findIndex(t => t.path === contextMenuPath.value)
  return index > 0
})

const canCloseRight = computed(() => {
  if (!contextMenuPath.value) return false
  const index = tabsStore.tabs.findIndex(t => t.path === contextMenuPath.value)
  return index < tabsStore.tabs.length - 1
})

const getTabTitle = (tab) => {
  if (tab.titleKey) {
    return t(tab.titleKey)
  }
  return tab.title || tab.name
}

const handleRemove = async (path) => {
  const isCurrentTab = tabsStore.activeTab === path
  const currentIndex = tabsStore.tabs.findIndex(t => t.path === path)
  
  // 先移除标签
  tabsStore.removeTab(path)
  
  // 如果关闭的是当前激活的标签，需要跳转到其他标签
  if (isCurrentTab) {
    if (tabsStore.tabs.length > 0) {
      // 优先跳转到右侧的标签，如果没有则跳转到左侧的标签
      let nextTab
      if (currentIndex < tabsStore.tabs.length) {
        // 右侧还有标签，跳转到右侧第一个
        nextTab = tabsStore.tabs[currentIndex]
      } else if (currentIndex > 0) {
        // 左侧还有标签，跳转到左侧最后一个
        nextTab = tabsStore.tabs[currentIndex - 1]
      } else {
        // 没有其他标签了，跳转到最后一个
        nextTab = tabsStore.tabs[tabsStore.tabs.length - 1]
      }
      
      if (nextTab) {
        tabsStore.setActiveTab(nextTab.path)
        await router.push(nextTab.path)
      }
    } else {
      // 如果没有标签了，跳转到首页并添加标签
      tabsStore.setActiveTab('/dashboard')
      await router.push('/dashboard')
    }
  }
}

const handleClick = (tab) => {
  const path = tab.paneName
  tabsStore.setActiveTab(path)
  router.push(path)
}

const handleRefresh = (path) => {
  tabsStore.refreshTab(path)
}

const handleContextMenu = async (command) => {
  const { action, path } = command
  contextMenuVisible.value = false

  switch (action) {
    case 'refresh':
      handleRefresh(path)
      break
    case 'close':
      await handleRemove(path)
      break
    case 'closeOther':
      const isCurrentTab = tabsStore.activeTab === path
      tabsStore.removeOtherTabs(path)
      if (isCurrentTab || tabsStore.activeTab !== path) {
        await router.push(path)
        tabsStore.setActiveTab(path)
      }
      break
    case 'closeLeft':
      const isCurrentTabLeft = tabsStore.activeTab === path
      tabsStore.removeLeftTabs(path)
      if (isCurrentTabLeft || tabsStore.activeTab !== path) {
        await router.push(path)
        tabsStore.setActiveTab(path)
      }
      break
    case 'closeRight':
      const isCurrentTabRight = tabsStore.activeTab === path
      tabsStore.removeRightTabs(path)
      if (isCurrentTabRight || tabsStore.activeTab !== path) {
        await router.push(path)
        tabsStore.setActiveTab(path)
      }
      break
    case 'closeAll':
      tabsStore.removeAllTabs()
      await router.push('/dashboard')
      tabsStore.setActiveTab('/dashboard')
      break
  }
}

const showContextMenu = (e, path) => {
  e.preventDefault()
  e.stopPropagation()
  contextMenuX.value = e.clientX
  contextMenuY.value = e.clientY
  contextMenuPath.value = path
  contextMenuVisible.value = true
}

onMounted(() => {
  // 点击其他地方关闭右键菜单
  const handleClick = () => {
    contextMenuVisible.value = false
  }
  document.addEventListener('click', handleClick)

  onUnmounted(() => {
    document.removeEventListener('click', handleClick)
  })
})
</script>

<style scoped>
.tabs-view {
  background: var(--header-bg);
  border-bottom: 1px solid var(--border-color-light);
  padding: 0;
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.tabs-container {
  margin: 0;
}

.tabs-container :deep(.el-tabs__header) {
  margin: 0;
  border-bottom: 1px solid var(--border-color-light);
  background: var(--header-bg);
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.tabs-container :deep(.el-tabs__nav-wrap) {
  margin-bottom: 0;
  padding: 0 12px;
}

.tabs-container :deep(.el-tabs__nav) {
  border: none;
}

.tabs-container :deep(.el-tabs__item) {
  height: 36px;
  line-height: 36px;
  padding: 0 14px;
  margin-right: 4px;
  margin-top: 4px;
  background: var(--bg-color-tertiary);
  border: 1px solid var(--border-color-light);
  border-bottom: none;
  border-radius: 4px 4px 0 0;
  color: var(--text-color-regular);
  font-size: 12px;
  user-select: none;
  transition: all 0.2s;
  position: relative;
}

.tabs-container :deep(.el-tabs__item:hover) {
  color: #409EFF;
  background: var(--bg-color-tertiary);
}

.tabs-container :deep(.el-tabs__item.is-active) {
  background: var(--header-bg);
  border: 1px solid #409EFF;
  border-bottom: 2px solid #409EFF;
  color: #409EFF;
  font-weight: 500;
  margin-bottom: -1px;
}

.tabs-container :deep(.el-tabs__item .el-icon-close) {
  width: 14px;
  height: 14px;
  font-size: 12px;
  margin-left: 8px;
  border-radius: 50%;
  transition: all 0.2s;
  color: var(--text-color-secondary);
}

.tabs-container :deep(.el-tabs__item .el-icon-close:hover) {
  background: #f56c6c;
  color: #fff;
}

.tabs-container :deep(.el-tabs__item.is-active .el-icon-close) {
  color: #409EFF;
}

.tabs-container :deep(.el-tabs__item.is-active .el-icon-close:hover) {
  background: #f56c6c;
  color: #fff;
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tab-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.refresh-icon {
  cursor: pointer;
  padding: 3px;
  border-radius: 4px;
  transition: all 0.2s;
  color: #409EFF;
  font-size: 14px;
  width: 20px;
  height: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-left: 4px;
}

.refresh-icon:hover {
  background: var(--bg-color-tertiary);
  color: #66B1FF;
  transform: rotate(180deg);
}

.context-menu {
  position: fixed;
  z-index: 9999;
  background: var(--card-bg);
  border: 1px solid var(--border-color-light);
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 var(--shadow-base);
  min-width: 160px;
  padding: 4px 0;
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.context-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-color-regular);
  transition: background-color 0.3s;
}

.context-menu-item:hover:not(.disabled) {
  background-color: var(--bg-color-tertiary);
}

.context-menu-item.disabled {
  color: var(--text-color-placeholder);
  cursor: not-allowed;
}

.context-menu-item .el-icon {
  font-size: 16px;
}

.context-menu-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9998;
  background: transparent;
}
</style>

