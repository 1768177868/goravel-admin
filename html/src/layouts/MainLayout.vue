<template>
  <el-container class="layout-container" :class="`layout-${appStore.layoutSize}`">
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
          <template v-if="menuTree.length > 0">
            <MenuItem
              v-for="menu in menuTree"
              :key="menu.id"
              :menu="menu"
            />
          </template>
          <template v-else>
            <!-- 默认菜单 -->
            <el-sub-menu index="system">
              <template #title>
                <el-icon><Setting /></el-icon>
                <span>{{ $t('menu.system') }}</span>
              </template>
              <el-menu-item index="/admins">
                <el-icon><User /></el-icon>
                <template #title>{{ $t('menu.admin') }}</template>
              </el-menu-item>
              <el-menu-item index="/roles">
                <el-icon><Avatar /></el-icon>
                <template #title>{{ $t('menu.role') }}</template>
              </el-menu-item>
              <el-menu-item index="/permissions">
                <el-icon><Key /></el-icon>
                <template #title>{{ $t('menu.permission') }}</template>
              </el-menu-item>
              <el-menu-item index="/menus">
                <el-icon><Menu /></el-icon>
                <template #title>{{ $t('menu.menu') }}</template>
              </el-menu-item>
              <el-menu-item index="/departments">
                <el-icon><OfficeBuilding /></el-icon>
                <template #title>{{ $t('menu.department') }}</template>
              </el-menu-item>
              <el-menu-item index="/dictionaries">
                <el-icon><Document /></el-icon>
                <template #title>{{ $t('menu.dictionary') }}</template>
              </el-menu-item>
              <el-menu-item index="/configs">
                <el-icon><Setting /></el-icon>
                <template #title>{{ $t('menu.config') }}</template>
              </el-menu-item>
              <el-menu-item index="/blacklists">
                <el-icon><Warning /></el-icon>
                <template #title>{{ $t('menu.blacklist') }}</template>
              </el-menu-item>
              <el-menu-item index="/online-admins">
                <el-icon><User /></el-icon>
                <template #title>{{ $t('menu.online_admin') }}</template>
              </el-menu-item>
              <el-menu-item index="/exports">
                <el-icon><Document /></el-icon>
                <template #title>{{ $t('menu.export') }}</template>
              </el-menu-item>
            </el-sub-menu>
            <el-sub-menu index="logs">
              <template #title>
                <el-icon><Document /></el-icon>
                <span>{{ $t('menu.log') }}</span>
              </template>
              <el-menu-item index="/operation-logs">{{ $t('menu.operation_log') }}</el-menu-item>
              <el-menu-item index="/login-logs">{{ $t('menu.login_log') }}</el-menu-item>
              <el-menu-item index="/system-logs">{{ $t('menu.system_log') }}</el-menu-item>
            </el-sub-menu>
            <el-menu-item index="/notifications">
              <el-icon><Bell /></el-icon>
              <template #title>{{ $t('menu.notification_center') }}</template>
            </el-menu-item>
            <el-menu-item index="/monitor">
              <el-icon><Monitor /></el-icon>
              <template #title>{{ $t('menu.service_monitor') }}</template>
            </el-menu-item>
          </template>
        </el-menu>
      </div>
    </el-drawer>

    <!-- 桌面端固定侧边栏 -->
    <el-aside
      v-if="!isMobile"
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
        <template v-if="menuTree.length > 0">
          <MenuItem
            v-for="menu in menuTree"
            :key="menu.id"
            :menu="menu"
          />
        </template>
        <template v-else>
          <!-- 如果菜单数据为空，显示默认菜单 -->
          <el-sub-menu index="system">
            <template #title>
              <el-icon><Setting /></el-icon>
              <span>{{ $t('menu.system') }}</span>
            </template>
            <el-menu-item index="/admins">
              <el-icon><User /></el-icon>
              <template #title>{{ $t('menu.admin') }}</template>
            </el-menu-item>
            <el-menu-item index="/roles">
              <el-icon><Avatar /></el-icon>
              <template #title>{{ $t('menu.role') }}</template>
            </el-menu-item>
            <el-menu-item index="/permissions">
              <el-icon><Key /></el-icon>
              <template #title>{{ $t('menu.permission') }}</template>
            </el-menu-item>
            <el-menu-item index="/menus">
              <el-icon><Menu /></el-icon>
              <template #title>{{ $t('menu.menu') }}</template>
            </el-menu-item>
            <el-menu-item index="/departments">
              <el-icon><OfficeBuilding /></el-icon>
              <template #title>{{ $t('menu.department') }}</template>
            </el-menu-item>
            <el-menu-item index="/dictionaries">
              <el-icon><Document /></el-icon>
              <template #title>{{ $t('menu.dictionary') }}</template>
            </el-menu-item>
            <el-menu-item index="/configs">
              <el-icon><Setting /></el-icon>
              <template #title>{{ $t('menu.config') }}</template>
            </el-menu-item>
            <el-menu-item index="/blacklists">
              <el-icon><Warning /></el-icon>
              <template #title>{{ $t('menu.blacklist') }}</template>
            </el-menu-item>
            <el-menu-item index="/online-admins">
              <el-icon><User /></el-icon>
              <template #title>{{ $t('menu.online_admin') }}</template>
            </el-menu-item>
            <el-menu-item index="/exports">
              <el-icon><Document /></el-icon>
              <template #title>{{ $t('menu.export') }}</template>
            </el-menu-item>
          </el-sub-menu>
          <el-sub-menu index="logs">
            <template #title>
              <el-icon><Document /></el-icon>
              <span>{{ $t('menu.log') }}</span>
            </template>
            <el-menu-item index="/operation-logs">{{ $t('menu.operation_log') }}</el-menu-item>
            <el-menu-item index="/login-logs">{{ $t('menu.login_log') }}</el-menu-item>
            <el-menu-item index="/system-logs">{{ $t('menu.system_log') }}</el-menu-item>
          </el-sub-menu>
          <el-menu-item index="/notifications">
            <el-icon><Bell /></el-icon>
            <template #title>{{ $t('menu.notification_center') }}</template>
          </el-menu-item>
          <el-menu-item index="/monitor">
            <el-icon><Monitor /></el-icon>
            <template #title>{{ $t('menu.service_monitor') }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>
    
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <!-- 移动端显示菜单按钮，桌面端显示折叠按钮 -->
          <el-button
            v-if="isMobile"
            type="text"
            class="collapse-btn mobile-menu-btn"
            @click="drawerVisible = true"
          >
            <el-icon><Menu /></el-icon>
          </el-button>
          <el-button
            v-else
            type="text"
            class="collapse-btn"
            @click="appStore.toggleSidebar"
          >
            <el-icon><Fold v-if="!appStore.sidebarCollapsed" /><Expand v-else /></el-icon>
          </el-button>
          <BreadcrumbView :class="{ 'mobile-hidden': isXs }" />
        </div>
        <div class="header-right">
          <!-- 移动端隐藏全屏按钮 -->
          <el-button
            v-if="!isMobile"
            type="text"
            class="header-btn"
            @click="appStore.toggleFullscreen"
            :title="$t('header.fullscreen')"
          >
            <el-icon>
              <FullScreen v-if="!appStore.isFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-button>
          <NotificationBell />
          <!-- <DarkModeSwitch /> -->
          <!-- 移动端隐藏时区和语言切换 -->
          <TimezoneSwitch :class="{ 'mobile-hidden': isMobile }" />
          <LanguageSwitch :class="{ 'mobile-hidden': isXs }" />
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

      <div class="tabs-wrapper" :class="{ 'mobile-hidden': isMobile }">
        <TabsView />
      </div>
      
      <el-main class="main-content">
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
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, watch, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessageBox } from 'element-plus'
import { useUserStore } from '../store/user'
import { useTabsStore } from '../store/tabs'
import { useAppStore } from '../store/app'
import request from '../utils/request'
import LanguageSwitch from '../components/LanguageSwitch.vue'
import TimezoneSwitch from '../components/TimezoneSwitch.vue'
import NotificationBell from '../components/NotificationBell.vue'
import DarkModeSwitch from '../components/DarkModeSwitch.vue'
import TabsView from '../components/TabsView.vue'
import BreadcrumbView from '../components/BreadcrumbView.vue'
import MenuItem from '../components/MenuItem.vue'
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
  Avatar,
  Key,
  Menu,
  OfficeBuilding,
  Document,
  Bell,
  Monitor,
  Warning
} from '@element-plus/icons-vue'

// 响应式检测
const { isMobile, isTablet, isXs } = useResponsive()

// 移动端抽屉控制
const drawerVisible = ref(false)

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const tabsStore = useTabsStore()
const appStore = useAppStore()
const { t } = useI18n()

const activeMenu = computed(() => route.path)

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

</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  /* background-color: var(--sidebar-bg); */
  background-color:#fff;
  overflow-y: auto;
  transition: width 0.3s;
  border-right: 1px solid #00000014;
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
  color:#383853;
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
  padding: 0 20px;
  height: 60px;
  line-height: 60px;
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
  transition: color 0.3s ease;
}

.size-btn {
  color: var(--text-color-regular);
  transition: color 0.3s ease;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-btn {
  color: var(--text-color-regular);
  padding: 8px;
  border-radius: 4px;
  transition: all 0.3s;
}

.header-btn:hover {
  background-color: var(--bg-color-tertiary);
  color: #409EFF;
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
  transition: color 0.3s ease;
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
  background: var(--header-bg);
  border-bottom: 1px solid var(--border-color-light);
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.main-content {
  background-color: var(--bg-color-secondary);
  padding: 20px;
  overflow-y: auto;
  transition: background-color 0.3s ease;
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
  color: #409EFF;
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

  .sidebar {
    width: 200px !important;
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
