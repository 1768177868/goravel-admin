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
            {{ getTabTitle(tab) }}
            <el-icon
              v-if="tab.path === activeTab"
              class="refresh-icon"
              @click.stop="handleRefresh(tab.path)"
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

const handleRemove = (path) => {
  tabsStore.removeTab(path)
  if (tabsStore.activeTab === path) {
    if (tabsStore.tabs.length > 0) {
      router.push(tabsStore.tabs[tabsStore.tabs.length - 1].path)
    } else {
      router.push('/dashboard')
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
  // 触发路由刷新
  router.replace({
    path: path + '?refresh=' + Date.now()
  }).then(() => {
    router.replace({ path })
  })
}

const handleContextMenu = (command) => {
  const { action, path } = command
  contextMenuVisible.value = false

  switch (action) {
    case 'refresh':
      handleRefresh(path)
      break
    case 'close':
      handleRemove(path)
      break
    case 'closeOther':
      tabsStore.removeOtherTabs(path)
      router.push(path)
      break
    case 'closeLeft':
      tabsStore.removeLeftTabs(path)
      router.push(path)
      break
    case 'closeRight':
      tabsStore.removeRightTabs(path)
      router.push(path)
      break
    case 'closeAll':
      tabsStore.removeAllTabs()
      router.push('/dashboard')
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
  background: white;
  border-bottom: 1px solid #e4e7ed;
}

.tabs-container {
  margin: 0;
}

.tabs-container :deep(.el-tabs__header) {
  margin: 0;
  border-bottom: none;
}

.tabs-container :deep(.el-tabs__item) {
  height: 40px;
  line-height: 40px;
  padding: 0 15px;
  border: 1px solid #e4e7ed;
  border-bottom: none;
  margin-right: 2px;
  background: #f5f7fa;
  user-select: none;
}

.tabs-container :deep(.el-tabs__item.is-active) {
  background: white;
  border-bottom: 1px solid white;
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.refresh-icon {
  cursor: pointer;
  padding: 2px;
  border-radius: 2px;
  transition: all 0.3s;
}

.refresh-icon:hover {
  background: #e4e7ed;
}

.context-menu {
  position: fixed;
  z-index: 9999;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  min-width: 160px;
  padding: 4px 0;
}

.context-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  transition: background-color 0.3s;
}

.context-menu-item:hover:not(.disabled) {
  background-color: #f5f7fa;
}

.context-menu-item.disabled {
  color: #c0c4cc;
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

