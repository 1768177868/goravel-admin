<template>
  <el-container class="layout-container" :class="`layout-${appStore.layoutSize}`">
    <el-aside
      :width="appStore.sidebarCollapsed ? '64px' : '200px'"
      class="sidebar"
      :class="{ 'is-collapse': appStore.sidebarCollapsed }"
    >
      <div class="logo">
        <h3 v-if="!appStore.sidebarCollapsed">{{ $t('header.system_management') }}</h3>
        <el-icon v-else><Setting /></el-icon>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        :collapse="appStore.sidebarCollapsed"
        background-color="#1f2937"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
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
              <span>{{ $t('menu.system_management') }}</span>
            </template>
            <el-menu-item index="/admins">
              <el-icon><User /></el-icon>
              <template #title>{{ $t('menu.admin_management') }}</template>
            </el-menu-item>
            <el-menu-item index="/roles">
              <el-icon><Avatar /></el-icon>
              <template #title>{{ $t('menu.role_management') }}</template>
            </el-menu-item>
            <el-menu-item index="/permissions">
              <el-icon><Key /></el-icon>
              <template #title>{{ $t('menu.permission_management') }}</template>
            </el-menu-item>
            <el-menu-item index="/menus">
              <el-icon><Menu /></el-icon>
              <template #title>{{ $t('menu.menu_management') }}</template>
            </el-menu-item>
            <el-menu-item index="/departments">
              <el-icon><OfficeBuilding /></el-icon>
              <template #title>{{ $t('menu.department_management') }}</template>
            </el-menu-item>
            <el-menu-item index="/dictionaries">
              <el-icon><Document /></el-icon>
              <template #title>{{ $t('menu.dictionary_management') }}</template>
            </el-menu-item>
            <el-menu-item index="/configs">
              <el-icon><Setting /></el-icon>
              <template #title>{{ $t('menu.config_management') }}</template>
            </el-menu-item>
            <el-menu-item index="/blacklists">
              <el-icon><Warning /></el-icon>
              <template #title>{{ $t('menu.blacklist_management') }}</template>
            </el-menu-item>
            <el-menu-item index="/online-users">
              <el-icon><User /></el-icon>
              <template #title>{{ $t('menu.online_user_management') }}</template>
            </el-menu-item>
            <el-menu-item index="/exports">
              <el-icon><Document /></el-icon>
              <template #title>{{ $t('menu.export_management') }}</template>
            </el-menu-item>
          </el-sub-menu>
          <el-sub-menu index="logs">
            <template #title>
              <el-icon><Document /></el-icon>
              <span>{{ $t('menu.log_management') }}</span>
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
          <el-button
            type="text"
            class="collapse-btn"
            @click="appStore.toggleSidebar"
          >
            <el-icon><Fold v-if="!appStore.sidebarCollapsed" /><Expand v-else /></el-icon>
          </el-button>
          <BreadcrumbView />
        </div>
        <div class="header-right">
          <el-button
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
          <TimezoneSwitch />
          <LanguageSwitch />
          <el-dropdown @command="handleCommand" class="user-dropdown">
            <span class="user-info">
              <el-avatar 
                v-if="userStore.adminInfo?.avatar" 
                :size="32" 
                :src="userStore.adminInfo.avatar"
                class="user-avatar"
              >
                <el-icon><User /></el-icon>
              </el-avatar>
              <el-icon v-else class="user-icon"><User /></el-icon>
              <span class="user-name">{{ userStore.adminInfo?.nickname || userStore.adminInfo?.username }}</span>
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
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

      <div class="tabs-wrapper">
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
import TabsView from '../components/TabsView.vue'
import BreadcrumbView from '../components/BreadcrumbView.vue'
import MenuItem from '../components/MenuItem.vue'
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

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const tabsStore = useTabsStore()
const appStore = useAppStore()
const { t } = useI18n()

const activeMenu = computed(() => route.path)

// 根据主题获取 sidebar 背景色

// 转换菜单数据格式并构建树形结构
const menuTree = computed(() => {
  const menus = userStore.menus || []
  
  if (menus.length === 0) {
    // console.warn('No menus found in userStore.menus, userStore:', userStore)
    return []
  }
  
  // 转换数据格式（后端返回的是扁平数组，需要自己构建树形结构）
  const transformMenu = (menu) => {
    // 直接使用后端返回的路径，前后端路径已统一
    const originalPath = menu.Path || menu.path || ''
    
    return {
      id: menu.id,
      parent_id: menu.ParentID || menu.parent_id || 0,
      title: menu.Title || menu.title || '',
      slug: menu.Slug || menu.slug || '',
      path: originalPath,
      icon: menu.Icon || menu.icon || '',
      type: menu.Type !== undefined ? menu.Type : (menu.type !== undefined ? menu.type : 1),
      status: menu.Status !== undefined ? menu.Status : (menu.status !== undefined ? menu.status : 1),
      sort: menu.Sort !== undefined ? menu.Sort : (menu.sort !== undefined ? menu.sort : 0),
      is_hidden: menu.IsHidden !== undefined ? menu.IsHidden : (menu.is_hidden !== undefined ? menu.is_hidden : 0),
      link_type: menu.LinkType !== undefined ? menu.LinkType : (menu.link_type !== undefined ? menu.link_type : 1),
      open_type: menu.OpenType !== undefined ? menu.OpenType : (menu.open_type !== undefined ? menu.open_type : 1)
    }
  }
  
  // 转换所有菜单（扁平数组）
  const transformedMenus = menus.map(menu => transformMenu(menu))
  
  // 构建树形结构（只返回顶级菜单，子菜单在children中）
  const buildTree = (menus, parentId = 0) => {
    const result = menus
      .filter(menu => menu.parent_id === parentId && menu.is_hidden === 0 && menu.status === 1)
      .map(menu => ({
        ...menu,
        children: buildTree(menus, menu.id)
      }))
      .sort((a, b) => a.sort - b.sort)
    
    return result
  }
  
  const tree = buildTree(transformedMenus)
  return tree
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
    }
  }
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
  background-color: var(--sidebar-bg);
  overflow-y: auto;
  transition: width 0.3s;
}

/* 自定义滚动条样式 - 更细更美观 */
.sidebar::-webkit-scrollbar {
  width: 6px;
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
  border-bottom: 1px solid #434a55;
}

.logo h3 {
  margin: 0;
  font-size: 18px;
  white-space: nowrap;
}

.sidebar-menu {
  border-right: none;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 200px;
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
</style>
