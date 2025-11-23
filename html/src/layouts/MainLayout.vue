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
        router
        class="sidebar-menu"
        :collapse="appStore.sidebarCollapsed"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <template #title>{{ $t('menu.dashboard') }}</template>
        </el-menu-item>
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
        <el-sub-menu index="logs">
          <template #title>
            <el-icon><Document /></el-icon>
            <span>{{ $t('menu.log_management') }}</span>
          </template>
          <el-menu-item index="/operation-logs">{{ $t('menu.operation_log') }}</el-menu-item>
          <el-menu-item index="/login-logs">{{ $t('menu.login_log') }}</el-menu-item>
          <el-menu-item index="/system-logs">{{ $t('menu.system_log') }}</el-menu-item>
        </el-sub-menu>
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
          <el-dropdown @command="handleLayoutSize" trigger="click">
            <el-button type="text" class="header-btn" :title="$t('layout.title')">
              <el-icon><Grid /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="small" :class="{ 'is-active': appStore.layoutSize === 'small' }">
                  {{ $t('layout.small') }}
                </el-dropdown-item>
                <el-dropdown-item command="default" :class="{ 'is-active': appStore.layoutSize === 'default' }">
                  {{ $t('layout.default') }}
                </el-dropdown-item>
                <el-dropdown-item command="large" :class="{ 'is-active': appStore.layoutSize === 'large' }">
                  {{ $t('layout.large') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <LanguageSwitch />
          <el-dropdown @command="handleCommand" class="user-dropdown">
            <span class="user-info">
              <el-icon><User /></el-icon>
              {{ userStore.adminInfo?.nickname || userStore.adminInfo?.username }}
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
                :key="`${routeItem.path}-${routeItem.query._refresh || ''}`"
              />
            </keep-alive>
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessageBox } from 'element-plus'
import { useUserStore } from '../store/user'
import { useTabsStore } from '../store/tabs'
import { useAppStore } from '../store/app'
import LanguageSwitch from '../components/LanguageSwitch.vue'
import TabsView from '../components/TabsView.vue'
import BreadcrumbView from '../components/BreadcrumbView.vue'
import {
  Fold,
  Expand,
  Setting,
  User,
  UserFilled,
  ArrowDown,
  FullScreen,
  Aim,
  Grid
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const tabsStore = useTabsStore()
const appStore = useAppStore()
const { t } = useI18n()

const activeMenu = computed(() => route.path)


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
  
  // 清理事件监听器
  onUnmounted(() => {
    document.removeEventListener('fullscreenchange', handleFullscreenChange)
  })
})

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

const handleLayoutSize = (size) => {
  appStore.setLayoutSize(size)
}

</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  overflow-y: auto;
  transition: width 0.3s;
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

.header {
  background-color: white;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 60px;
  line-height: 60px;
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
  color: #606266;
}

.size-btn {
  color: #606266;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-btn {
  color: #606266;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.3s;
}

.header-btn:hover {
  background-color: #f5f7fa;
  color: #409EFF;
}

.user-dropdown {
  margin-left: 0;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #606266;
}

.user-info .el-icon {
  margin: 0 5px;
}

.tabs-wrapper {
  background: white;
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
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
