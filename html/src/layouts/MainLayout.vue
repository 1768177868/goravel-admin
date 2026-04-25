<template>
  <el-container class="layout-container" :class="[`layout-${appStore.layoutSize}`, { 'layout-top-menu': appStore.menuMode === 'top' && !isMobile }]">
    <!-- 移动端抽屉式侧边栏 -->
    <el-drawer
      v-model="drawerVisible"
      :with-header="false"
      direction="ltr"
      :size="isMobile ? '80%' : '240px'"
      :modal="true"
      :show-close="false"
      class="mobile-drawer"
      @close="handleDrawerClose"
    >
      <div class="drawer-content">
        <div class="logo">
          <h3>{{ $t('header.system') }}</h3>
        </div>
        <el-menu
          :default-active="activeMenu"
          class="sidebar-menu"
          @select="handleMenuSelect"
        >
          <el-menu-item index="/dashboard">
            <el-icon><Odometer /></el-icon>
            <template #title>{{ $t('menu.dashboard') }}</template>
          </el-menu-item>
          <MenuItem
            v-for="menu in menuTree"
            :key="menu.id"
            :menu="menu"
          />
        </el-menu>
      </div>
    </el-drawer>

    <!-- 桌面端固定侧边栏（仅左侧菜单模式显示） -->
    <el-aside
      v-if="!isMobile && appStore.menuMode === 'sidebar'"
      :width="appStore.sidebarCollapsed ? '64px' : '240px'"
      class="sidebar"
      :class="{ 'is-collapse': appStore.sidebarCollapsed }"
    >
      <div class="logo">
        <h3 v-if="!appStore.sidebarCollapsed">{{ $t('header.system') }}</h3>
        <el-icon v-else><Setting /></el-icon>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        :collapse="appStore.sidebarCollapsed"
        @select="handleMenuSelect"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <template #title>{{ $t('menu.dashboard') }}</template>
        </el-menu-item>
        <MenuItem
          v-for="menu in menuTree"
          :key="menu.id"
          :menu="menu"
        />
      </el-menu>
    </el-aside>
    
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <!-- 移动端显示菜单按钮；桌面端左侧菜单模式显示折叠按钮；顶部菜单模式不显示 -->
          <el-button
            v-if="isMobile"
            type="text"
            class="collapse-btn mobile-menu-btn"
            @click="drawerVisible = true"
          >
            <el-icon><Menu /></el-icon>
          </el-button>
          <el-button
            v-else-if="appStore.menuMode === 'sidebar'"
            type="text"
            class="collapse-btn"
            @click="appStore.toggleSidebar"
          >
            <el-icon><Fold v-if="!appStore.sidebarCollapsed" /><Expand v-else /></el-icon>
          </el-button>
          <BreadcrumbView :class="{ 'mobile-hidden': isXs }" />
        </div>
        <div class="header-right">
          <!-- 菜单搜索 -->
          <MenuSearch v-if="!isMobile" :menus="menuTree" />
          <!-- 移动端隐藏全屏按钮 -->
          <el-button
            v-if="!isMobile"
            type="text"
            class="header-btn"
            @click="appStore.toggleFullscreen"
            :title="$t('header.fullscreen')"
          >
            <el-icon class="header-icon-fixed">
              <FullScreen v-if="!appStore.isFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-button>
          <!-- 布局大小设置 -->
          <el-dropdown
            v-if="!isMobile"
            @command="handleLayoutSizeChange"
            class="layout-size-dropdown"
          >
            <el-button type="text" class="header-btn" :title="$t('header.layout_size')">
              <el-icon class="header-icon-fixed"><Grid /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="large" :class="{ 'is-active': appStore.layoutSize === 'large' }">
                  <span style="display: flex; align-items: center; gap: 8px;">
                    <el-icon v-if="appStore.layoutSize === 'large'" style="font-size: 16px;"><Check /></el-icon>
                    <span v-else style="width: 16px;"></span>
                    {{ $t('header.layout_size_large') }}
                  </span>
                </el-dropdown-item>
                <el-dropdown-item command="default" :class="{ 'is-active': appStore.layoutSize === 'default' }">
                  <span style="display: flex; align-items: center; gap: 8px;">
                    <el-icon v-if="appStore.layoutSize === 'default'" style="font-size: 16px;"><Check /></el-icon>
                    <span v-else style="width: 16px;"></span>
                    {{ $t('header.layout_size_default') }}
                  </span>
                </el-dropdown-item>
                <el-dropdown-item command="small" :class="{ 'is-active': appStore.layoutSize === 'small' }">
                  <span style="display: flex; align-items: center; gap: 8px;">
                    <el-icon v-if="appStore.layoutSize === 'small'" style="font-size: 16px;"><Check /></el-icon>
                    <span v-else style="width: 16px;"></span>
                    {{ $t('header.layout_size_small') }}
                  </span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <!-- 设置（导航模式、水印） -->
          <el-popover
            v-if="!isMobile"
            placement="bottom-end"
            :width="280"
            trigger="click"
          >
            <template #reference>
              <el-button
                type="text"
                class="header-btn"
                :title="$t('header.settings')"
              >
                <el-icon class="header-icon-fixed"><Tools /></el-icon>
              </el-button>
            </template>
            <div class="settings-panel">
              <div class="settings-title">{{ $t('header.settings') }}</div>
              <div class="settings-item">
                <span class="settings-label">{{ $t('header.menu_mode') }}</span>
                <el-radio-group :model-value="appStore.menuMode" size="small" @update:model-value="appStore.setMenuMode">
                  <el-radio-button label="sidebar">{{ $t('header.menu_mode_sidebar') }}</el-radio-button>
                  <el-radio-button label="top">{{ $t('header.menu_mode_top') }}</el-radio-button>
                </el-radio-group>
              </div>
              <div class="settings-item">
                <span class="settings-label">{{ $t('header.watermark') }}</span>
                <el-switch v-model="appStore.watermarkEnabled" @change="appStore.setWatermarkEnabled(appStore.watermarkEnabled)" />
              </div>
              <div class="settings-item settings-item-theme">
                <span class="settings-label">{{ $t('header.theme_color') }}</span>
                <div class="theme-color-swatches">
                  <button
                    v-for="t in themeColorOptions"
                    :key="t.key"
                    type="button"
                    class="theme-swatch"
                    :class="{ active: appStore.themeColor === t.key }"
                    :style="{ backgroundColor: t.color }"
                    :title="t.key"
                    @click="appStore.setThemeColor(t.key)"
                  />
                </div>
              </div>
            </div>
          </el-popover>
          <NotificationBell />
          <DarkModeSwitch />
          <LanguageSwitch :class="{ 'mobile-hidden': isXs }" />
          <el-button
            type="text"
            class="header-btn"
            :class="{ 'mobile-hidden': isXs }"
            :title="$t('header.lock_screen')"
            @click="handleLockScreen"
          >
            <el-icon class="header-icon-fixed"><Lock /></el-icon>
          </el-button>
          <!-- 移动端隐藏时区切换 -->
          <TimezoneSwitch :class="{ 'mobile-hidden': isMobile }" />
          <el-dropdown @command="handleCommand" class="user-dropdown">
            <span class="user-info">
              <el-avatar 
                v-if="userStore.adminInfo?.avatar" 
                :size="isMobile ? 28 : 32" 
                :src="userStore.adminInfo.avatar"
                class="user-avatar"
              >
                <el-icon><User /></el-icon>
              </el-avatar>
              <el-icon v-else class="user-icon"><User /></el-icon>
              <span class="user-name" :class="{ 'mobile-hidden': isXs }">
                {{ userStore.adminInfo?.nickname || userStore.adminInfo?.username }}
              </span>
              <el-icon class="el-icon--right" :class="{ 'mobile-hidden': isMobile }">
                <ArrowDown />
              </el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><UserFilled /></el-icon>
                  {{ $t('header.profile') }}
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">{{ $t('header.logout') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 顶部菜单模式：水平菜单栏 -->
      <div v-if="!isMobile && appStore.menuMode === 'top'" class="top-menu-bar">
        <el-menu
          :default-active="activeMenu"
          mode="horizontal"
          class="top-menu"
          @select="handleMenuSelect"
        >
          <el-menu-item index="/dashboard">
            <el-icon><Odometer /></el-icon>
            <template #title>{{ $t('menu.dashboard') }}</template>
          </el-menu-item>
          <MenuItem
            v-for="menu in menuTree"
            :key="menu.id"
            :menu="menu"
          />
        </el-menu>
      </div>

      <div class="tabs-wrapper" :class="{ 'mobile-hidden': isMobile }">
        <TabsView />
      </div>
      
      <el-main class="main-content" :class="{ 'main-content-iframe': isIframePage }">
        <!-- 使用 Element Plus 水印：开启时包裹内容，水印浮在内容之上 -->
        <el-watermark
          v-if="appStore.watermarkEnabled"
          :content="watermarkText"
          :font="watermarkFont"
          :width="120"
          :height="48"
          :z-index="9"
          :gap="[80, 80]"
          :rotate="-22"
          class="main-watermark"
        >
          <div class="main-content-inner">
            <router-view v-slot="{ Component, route: routeItem }">
              <transition name="fade-transform" mode="out-in">
                <keep-alive>
                  <component
                    :is="Component"
                    :key="`${routeItem.path}-${tabsStore.getRefreshKey(routeItem.path)}`"
                  />
                </keep-alive>
              </transition>
            </router-view>
          </div>
        </el-watermark>
        <div v-else class="main-content-inner">
          <router-view v-slot="{ Component, route: routeItem }">
            <transition name="fade-transform" mode="out-in">
              <keep-alive>
                <component
                  :is="Component"
                  :key="`${routeItem.path}-${tabsStore.getRefreshKey(routeItem.path)}`"
                />
              </keep-alive>
            </transition>
          </router-view>
        </div>
      </el-main>
    </el-container>

    <el-dialog
      v-model="lockDialogVisible"
      :title="$t('header.lock_screen')"
      width="420px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      append-to-body
    >
      <el-input
        v-model="pendingLockPassword"
        type="password"
        show-password
        :name="lockDialogInputName"
        autocomplete="new-password"
        autocorrect="off"
        spellcheck="false"
        :placeholder="$t('header.lock_password_placeholder')"
        @keyup.enter="confirmLockScreen"
      />
      <div v-if="lockDialogError" class="lock-screen-error">{{ lockDialogError }}</div>
      <template #footer>
        <el-button @click="lockDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmLockScreen">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <div v-if="isScreenLocked" class="lock-screen-overlay">
      <div class="lock-screen-card">
        <div class="lock-screen-avatar-wrap">
          <el-avatar
            v-if="userStore.adminInfo?.avatar"
            :size="68"
            :src="userStore.adminInfo.avatar"
          />
          <el-avatar v-else :size="68">
            <el-icon><User /></el-icon>
          </el-avatar>
        </div>
        <div class="lock-screen-title">{{ $t('header.lock_screen_title') }}</div>
        <div class="lock-screen-user">{{ userStore.adminInfo?.nickname || userStore.adminInfo?.username }}</div>
        <el-input
          v-model="unlockPassword"
          type="password"
          show-password
          class="lock-screen-input"
          autocomplete="new-password"
          :name="lockInputName"
          autocorrect="off"
          spellcheck="false"
          :placeholder="$t('header.lock_password_placeholder')"
          @input="handleUnlockInput"
          @keyup.enter="handleUnlockScreen"
        />
        <div v-if="unlockError" class="lock-screen-error">{{ unlockError }}</div>
        <div class="lock-screen-actions">
          <el-button type="primary" @click="handleUnlockScreen">{{ $t('header.unlock') }}</el-button>
          <el-button @click="goToLogin">{{ $t('header.back_to_login') }}</el-button>
        </div>
      </div>
    </div>
  </el-container>
</template>

<script setup>
import { computed, watch, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessageBox } from 'element-plus'
import { useUserStore } from '../store/user'
import { useTabsStore } from '../store/tabs'
import { useAppStore, THEME_COLORS } from '../store/app'
import request from '../utils/request'
import LanguageSwitch from '../components/LanguageSwitch.vue'
import TimezoneSwitch from '../components/TimezoneSwitch.vue'
import NotificationBell from '../components/NotificationBell.vue'
import DarkModeSwitch from '../components/DarkModeSwitch.vue'
import TabsView from '../components/TabsView.vue'
import BreadcrumbView from '../components/BreadcrumbView.vue'
import MenuItem from '../components/MenuItem.vue'
import MenuSearch from '../components/MenuSearch.vue'
import { filterAndSortTree } from '../utils/tree'
import { useResponsive } from '../composables/useResponsive'
import {
  Fold,
  Expand,
  Setting,
  User,
  UserFilled,
  ArrowDown,
  FullScreen,
  Aim,
  Odometer,
  Menu,
  Grid,
  Check,
  Tools,
  Lock
} from '@element-plus/icons-vue'

// 主题色选项（与设置面板色块一致）
const themeColorOptions = THEME_COLORS

// 响应式检测
const { isMobile, isTablet, isXs } = useResponsive()

// 移动端抽屉控制
const drawerVisible = ref(false)

// 水印文字（当前用户）
const watermarkText = computed(() => {
  const name = userStore.adminInfo?.nickname || userStore.adminInfo?.username || 'Admin'
  return name
})
// Element Plus 水印字体：小字号、半透明，暗色模式适配
const watermarkFont = computed(() => ({
  fontSize: 14,
  color: appStore.darkMode ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.12)',
  fontWeight: 'normal'
}))

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const tabsStore = useTabsStore()
const appStore = useAppStore()
const { t } = useI18n()

const activeMenu = computed(() => route.path)
const isScreenLocked = ref(false)
const lockPassword = ref('')
const unlockPassword = ref('')
const unlockError = ref('')
const lockInputName = `lock-screen-password-${Math.random().toString(36).slice(2)}`
const lockDialogInputName = `set-lock-password-${Math.random().toString(36).slice(2)}`
const lockDialogVisible = ref(false)
const pendingLockPassword = ref('')
const lockDialogError = ref('')

// 是否为 iframe 外部链接页面（用于占满主内容区高度）
const isIframePage = computed(() => route.path === '/iframe')

// 过滤菜单树形结构（后端已返回树形结构，这里只需要过滤）
const menuTree = computed(() => {
  const menus = userStore.menus || []
  
  if (menus.length === 0) {
    return []
  }
  
  // 后端已返回树形结构，只需要过滤掉隐藏和禁用的菜单，然后排序
  return filterAndSortTree(
    menus,
    menu => menu.is_hidden === 0 && menu.status === 1,
    (a, b) => a.sort - b.sort
  )
})

// 监听路由变化，自动添加标签页
watch(
  () => route.path,
  (newPath) => {
    if (route.meta.requiresAuth !== false && route.name !== 'Login') {
      tabsStore.addTab(route)
      // 菜单设置为不缓存时，每次进入页面刷新（更新 key 使组件重新挂载并请求接口）
      if (route.meta?.noCache) {
        tabsStore.refreshTab(route.path)
      }
    }
  },
  { immediate: true }
)

// 心跳机制：每2分钟发送一次心跳请求，更新用户的最后活跃时间
let heartbeatInterval = null

const sendHeartbeat = async () => {
  try {
    // 只有在已登录状态下才发送心跳
    if (userStore.token) {
      await request.get('/heartbeat')
    }
  } catch (error) {
    // 心跳失败不显示错误，静默处理
    console.debug('Heartbeat failed:', error)
  }
}

// 监听全屏事件
onMounted(() => {
  // 初始化布局大小
  appStore.setLayoutSize(appStore.layoutSize)
  
  // 如果当前路由需要标签页，添加它
  if (route.meta.requiresAuth !== false && route.name !== 'Login') {
    tabsStore.addTab(route)
  }

  // 初始化全屏状态
  appStore.isFullscreen = !!document.fullscreenElement

  // 监听全屏状态变化
  const handleFullscreenChange = () => {
    appStore.isFullscreen = !!document.fullscreenElement
  }
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  
  // 启动心跳机制：每2分钟发送一次
  heartbeatInterval = setInterval(sendHeartbeat, 2 * 60 * 1000)
  // 立即发送一次心跳
  sendHeartbeat()
  
  // 清理事件监听器和心跳定时器
  onUnmounted(() => {
    document.removeEventListener('fullscreenchange', handleFullscreenChange)
    if (heartbeatInterval) {
      clearInterval(heartbeatInterval)
      heartbeatInterval = null
    }
  })
})

const handleMenuSelect = (index) => {
  // 处理静态菜单项的导航（如 dashboard）
  // MenuItem 组件已经处理了动态菜单的点击，所以这里主要处理静态菜单
  // 外部链接的 index 以 'external-' 开头，不应该在这里处理
  if (index && typeof index === 'string' && !index.startsWith('external-')) {
    // 检查是否是有效的内部路由路径（不以 http:// 或 https:// 开头）
    if (!index.startsWith('http://') && !index.startsWith('https://')) {
      router.push(index)
      // 移动端选择菜单后自动关闭抽屉
      if (isMobile.value) {
        drawerVisible.value = false
      }
    }
  }
}

// 处理抽屉关闭
const handleDrawerClose = () => {
  drawerVisible.value = false
}

const handleCommand = async (command) => {
  if (command === 'profile') {
    router.push('/profile')
  } else if (command === 'logout') {
    try {
      await ElMessageBox.confirm(t('header.logout_confirm'), t('common.confirm'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      })
      await userStore.logout()
      tabsStore.removeAllTabs()
      router.push('/login')
    } catch (error) {
      // 用户取消
    }
  }
}

const handleLayoutSizeChange = (size) => {
  appStore.setLayoutSize(size)
}

const handleLockScreen = async () => {
  pendingLockPassword.value = ''
  lockDialogError.value = ''
  lockDialogVisible.value = true
}

const confirmLockScreen = () => {
  if (!pendingLockPassword.value || !pendingLockPassword.value.trim()) {
    lockDialogError.value = t('header.lock_password_required')
    return
  }

  lockPassword.value = pendingLockPassword.value.trim()
  unlockPassword.value = ''
  unlockError.value = ''
  pendingLockPassword.value = ''
  lockDialogError.value = ''
  lockDialogVisible.value = false
  isScreenLocked.value = true
}

const handleUnlockInput = () => {
  if (unlockError.value) {
    unlockError.value = ''
  }
}

const handleUnlockScreen = () => {
  if (!unlockPassword.value.trim()) {
    unlockError.value = t('header.lock_password_required')
    return
  }

  if (unlockPassword.value !== lockPassword.value) {
    unlockError.value = t('header.lock_password_invalid')
    return
  }

  isScreenLocked.value = false
  lockPassword.value = ''
  unlockPassword.value = ''
  unlockError.value = ''
}

const goToLogin = async () => {
  try {
    await userStore.logout()
  } finally {
    tabsStore.removeAllTabs()
    isScreenLocked.value = false
    lockPassword.value = ''
    unlockPassword.value = ''
    unlockError.value = ''
    pendingLockPassword.value = ''
    lockDialogError.value = ''
    lockDialogVisible.value = false
    router.push('/login')
  }
}

</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  /* background-color: var(--sidebar-bg); */
  background-color: var(--card-bg, #fff);
  overflow-y: auto;
  transition: width 0.3s, background-color 0.3s ease;
  border-right: 1px solid var(--border-color-light, #00000014);
}

/* 自定义滚动条样式 - 更细更美观 */
.sidebar::-webkit-scrollbar {
  width: 4px;
}

.sidebar::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
  transition: background-color 0.3s;
}

.sidebar::-webkit-scrollbar-thumb:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

/* 兼容 Firefox */
.sidebar {
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.2) transparent;
}

.sidebar.is-collapse {
  width: 64px;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  /* border-bottom: 1px solid #434a55; */
}

.logo h3 {
  margin: 0;
  font-size: 18px;
  white-space: nowrap;
  color: var(--text-color-primary, #383853);
  opacity: 1;
}

.sidebar-menu {
  border-right: none;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 240px;
}

/* 菜单项文字溢出处理 */
.sidebar-menu :deep(.el-menu-item),
.sidebar-menu :deep(.el-sub-menu__title) {
  display: flex;
  align-items: center;
  overflow: hidden;
}

/* 菜单项标题容器 */
.sidebar-menu :deep(.el-menu-item > span),
.sidebar-menu :deep(.el-sub-menu__title > span) {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  display: flex;
  align-items: center;
}

/* 确保下拉箭头不被遮挡，固定在右侧 */
.sidebar-menu :deep(.el-sub-menu__icon-arrow) {
  flex-shrink: 0;
  margin-left: auto;
  margin-right: 0;
  width: 16px;
  text-align: right;
}

/* 菜单项图标样式 */
.sidebar-menu :deep(.el-menu-item .el-icon),
.sidebar-menu :deep(.el-sub-menu__title .el-icon) {
  flex-shrink: 0;
  margin-right: 8px;
}

/* 菜单项文字溢出处理 */
.sidebar-menu :deep(.el-menu-item),
.sidebar-menu :deep(.el-sub-menu__title) {
  display: flex;
  align-items: center;
  overflow: hidden;
}

.sidebar-menu :deep(.el-menu-item > span),
.sidebar-menu :deep(.el-sub-menu__title > span) {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

/* 确保下拉箭头不被遮挡 */
.sidebar-menu :deep(.el-sub-menu__icon-arrow) {
  flex-shrink: 0;
  margin-left: auto;
  margin-right: 0;
}

.header {
  background-color: var(--header-bg);
  border-bottom: 1px solid var(--border-color-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 18px;
  height: 62px;
  line-height: 62px;
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.04);
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.header-left {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 15px;
  height: 100%;
}

.collapse-btn {
  font-size: 18px;
  color: var(--text-color-regular);
  border-radius: 10px;
  transition: color 0.25s ease, background-color 0.25s ease, transform 0.2s ease;
}

.collapse-btn :deep(.el-icon) {
  transition: transform 0.25s ease;
}

.collapse-btn:hover {
  background-color: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  color: var(--el-color-primary);
  transform: translateY(-1px);
}

.collapse-btn:hover :deep(.el-icon) {
  transform: scale(1.08);
}

.collapse-btn:active {
  transform: translateY(0) scale(0.98);
}

.size-btn {
  color: var(--text-color-regular);
  transition: color 0.3s ease;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-btn {
  color: var(--text-color-regular);
  padding: 8px;
  border-radius: 10px;
  position: relative;
  overflow: hidden;
  transform: translateZ(0);
  transition: color 0.25s ease, background-color 0.25s ease, transform 0.2s ease, box-shadow 0.25s ease;
}

.header-icon-fixed {
  font-size: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.25s ease, color 0.25s ease;
}

.header-btn:hover {
  background-color: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  color: var(--el-color-primary);
  transform: translateY(-1px);
  box-shadow: 0 6px 12px color-mix(in srgb, var(--el-color-primary) 18%, transparent);
}

.header-btn:hover .header-icon-fixed {
  transform: scale(1.1);
}

.header-btn:active {
  transform: translateY(0) scale(0.98);
  box-shadow: none;
}

.header-btn::after {
  content: '';
  position: absolute;
  width: 90px;
  height: 90px;
  left: -120%;
  top: 50%;
  transform: translateY(-50%) rotate(25deg);
  background: linear-gradient(
    90deg,
    transparent 0%,
    color-mix(in srgb, var(--el-color-primary) 22%, transparent) 48%,
    transparent 100%
  );
  opacity: 0;
  transition: left 0.55s ease, opacity 0.35s ease;
  pointer-events: none;
}

.header-btn:hover::after {
  left: 120%;
  opacity: 0.55;
}

.layout-size-dropdown {
  margin-right: 0;
}

.layout-size-dropdown :deep(.el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: 8px;
}

.layout-size-dropdown :deep(.el-dropdown-menu__item .el-icon) {
  margin-right: 0;
  font-size: 16px;
}

.user-dropdown {
  margin-left: 0;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: var(--text-color-regular);
  gap: 8px;
  padding: 4px 8px;
  border-radius: 10px;
  transition: color 0.25s ease, background-color 0.25s ease, transform 0.2s ease;
}

.user-info:hover {
  color: var(--el-color-primary);
  background-color: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  transform: translateY(-1px);
}

.user-info:active {
  transform: translateY(0) scale(0.99);
}

.user-avatar {
  flex-shrink: 0;
}

.user-icon {
  flex-shrink: 0;
}

.user-name {
  white-space: nowrap;
}

.tabs-wrapper {
  background: linear-gradient(180deg, color-mix(in srgb, var(--header-bg) 92%, var(--el-color-primary) 8%) 0%, var(--header-bg) 100%);
  border-bottom: 1px solid color-mix(in srgb, var(--border-color-light) 70%, transparent);
  padding: 8px 12px 10px;
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.main-content {
  position: relative;
  background-color: var(--bg-color-secondary);
  padding: 20px;
  overflow-y: auto;
  transition: background-color 0.3s ease;
}

/* iframe 外部链接页面：去掉内边距并让内容区占满高度 */
.main-content-iframe.main-content {
  padding: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.main-content-iframe .main-content-inner,
.main-content-iframe .main-content-inner > *,
.main-content-iframe .main-content-inner > * > *,
.main-content-iframe .main-content-inner > * > * > * {
  height: 100%;
  min-height: 0;
}
.main-content-iframe .main-content-inner {
  display: flex;
  flex-direction: column;
}
.main-content-iframe .main-content-inner > * {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.main-content-iframe .main-content-inner > * > *,
.main-content-iframe .main-content-inner > * > * > * {
  flex: 1;
  min-height: 0;
}

/* 设置面板 */
.settings-panel {
  padding: 4px 0;
}
.settings-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color-light);
}
.settings-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  gap: 12px;
}
.settings-item:last-child {
  margin-bottom: 0;
}
.settings-label {
  font-size: 13px;
  color: var(--text-color-regular);
  flex-shrink: 0;
}
.settings-item .el-radio-group {
  flex-wrap: wrap;
}
.settings-item .el-radio-button {
  margin-bottom: 4px;
}
.settings-item-theme {
  flex-wrap: wrap;
  align-items: flex-start;
}
.settings-item-theme .settings-label {
  width: 100%;
  margin-bottom: 8px;
}
.theme-color-swatches {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.theme-swatch {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  border: 2px solid transparent;
  cursor: pointer;
  padding: 0;
  flex-shrink: 0;
  transition: transform 0.15s ease, border-color 0.15s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
.theme-swatch:hover {
  transform: scale(1.1);
}
.theme-swatch.active {
  border-color: var(--text-color-primary);
  box-shadow: 0 0 0 1px var(--bg-color);
}

/* 顶部菜单栏 */
.top-menu-bar {
  background: linear-gradient(180deg, color-mix(in srgb, var(--header-bg) 94%, var(--el-color-primary) 6%) 0%, var(--header-bg) 100%);
  border-bottom: 1px solid color-mix(in srgb, var(--border-color-light) 70%, transparent);
  padding: 8px 12px 10px;
  flex-shrink: 0;
}
.top-menu {
  border-bottom: none !important;
  background: color-mix(in srgb, var(--card-bg, #fff) 92%, transparent) !important;
  border-radius: 12px;
  padding: 0 8px;
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.06);
}
.top-menu :deep(.el-menu-item),
.top-menu :deep(.el-sub-menu__title) {
  height: 48px;
  line-height: 48px;
  border-bottom: 2px solid transparent;
  border-radius: 10px;
  margin: 6px 4px;
  padding: 0 14px !important;
  transition: all 0.2s ease;
}
.top-menu :deep(.el-menu-item:hover),
.top-menu :deep(.el-sub-menu__title:hover) {
  background-color: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
}
.top-menu :deep(.el-menu-item.is-active),
.top-menu :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
  border-bottom-color: transparent;
  color: var(--el-menu-active-color, var(--el-color-primary));
  background-color: color-mix(in srgb, var(--el-color-primary) 14%, transparent);
  font-weight: 600;
}
.top-menu :deep(.el-sub-menu .el-menu-item) {
  min-width: 120px;
}

/* Element Plus 水印容器：占满主内容区 */
.main-watermark {
  display: block;
  min-height: 100%;
  width: 100%;
}
.main-watermark :deep(.el-watermark) {
  min-height: 100%;
}
.main-content-inner {
  position: relative;
  min-height: 100%;
}

.lock-screen-overlay {
  position: fixed;
  inset: 0;
  z-index: 4000;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.lock-screen-card {
  width: 360px;
  max-width: 100%;
  background: var(--card-bg, #fff);
  border-radius: 12px;
  padding: 24px 20px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.18);
  text-align: center;
}

.lock-screen-avatar-wrap {
  display: flex;
  justify-content: center;
  margin-bottom: 12px;
}

.lock-screen-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.lock-screen-user {
  margin-top: 6px;
  margin-bottom: 16px;
  color: var(--text-color-regular);
}

.lock-screen-input {
  margin-bottom: 8px;
}

.lock-screen-error {
  min-height: 20px;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--el-color-danger);
  text-align: left;
}

.lock-screen-actions {
  display: flex;
  justify-content: center;
  gap: 10px;
}

/* 布局大小样式 */
.layout-small .main-content {
  padding: 10px;
}

.layout-large .main-content {
  padding: 30px;
}

/* 过渡动画 */
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.is-active {
  color: var(--el-color-primary);
  font-weight: bold;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .layout-container {
    position: relative;
  }

  .sidebar {
    display: none;
  }

  .header {
    padding: 0 12px;
    height: 50px;
    line-height: 50px;
    box-shadow: none;
  }

  .header-left {
    gap: 8px;
  }

  .header-right {
    gap: 6px;
  }

  .mobile-menu-btn {
    font-size: 20px;
    padding: 8px;
    min-width: 44px;
    min-height: 44px;
  }

  .collapse-btn {
    font-size: 20px;
    padding: 8px;
    min-width: 44px;
    min-height: 44px;
  }

  .header-btn {
    padding: 6px;
    min-width: 40px;
    min-height: 40px;
  }

  .user-info {
    gap: 4px;
  }

  .user-name {
    font-size: 14px;
  }

  .main-content {
    padding: 12px;
  }

  .mobile-hidden {
    display: none !important;
  }
}

/* 平板适配 */
@media (min-width: 769px) and (max-width: 991px) {
  .header {
    padding: 0 16px;
  }

  .main-content {
    padding: 16px;
  }

  .sidebar.is-collapse {
    width: 64px;
  }
}

/* 移动端抽屉样式 */
.mobile-drawer {
  z-index: 2000;
}

.mobile-drawer :deep(.el-drawer__body) {
  padding: 0;
}

.drawer-content {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.drawer-content .logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid var(--border-color-light);
  padding: 0 20px;
}

.drawer-content .logo h3 {
  margin: 0;
  font-size: 18px;
  color: #383853;
}

.drawer-content .sidebar-menu {
  flex: 1;
  overflow-y: auto;
  border-right: none;
}

/* 触摸优化 - 增大点击区域 */
@media (max-width: 768px) {
  .el-button {
    min-height: 44px;
    padding: 10px 16px;
  }

  .el-menu-item {
    min-height: 48px;
    line-height: 48px;
  }

  .el-dropdown-menu__item {
    min-height: 44px;
    line-height: 44px;
  }
}
</style>

<style>
/* 夜间模式：侧边栏菜单悬停时使用深色背景，避免浅字+浅底看不清（仅侧栏，顶部菜单保持蓝色标题不变） */
html.dark .sidebar-menu .el-menu-item:hover,
html.dark .sidebar-menu .el-sub-menu__title:hover {
  background-color: rgba(255, 255, 255, 0.06) !important;
}
html.dark .drawer-content .sidebar-menu .el-menu-item:hover,
html.dark .drawer-content .sidebar-menu .el-sub-menu__title:hover {
  background-color: rgba(255, 255, 255, 0.06) !important;
}
html.dark .header {
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.28);
}
html.dark .top-menu {
  box-shadow: 0 10px 18px rgba(0, 0, 0, 0.26);
}
html.dark .header-btn:hover {
  box-shadow: 0 8px 14px rgba(0, 0, 0, 0.24);
}
</style>
