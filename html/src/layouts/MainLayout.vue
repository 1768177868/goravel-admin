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
          <div class="logo-brand">
            <img
              v-if="websiteLogoUrl"
              :src="websiteLogoUrl"
              alt="logo"
              class="logo-image"
            />
            <h3>{{ systemTitle }}</h3>
          </div>
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
      :width="sidebarEffectiveCollapsed ? '64px' : '240px'"
      class="sidebar"
    >
      <div class="logo">
        <div v-if="!appStore.sidebarCollapsed" class="logo-brand">
          <img
            v-if="websiteLogoUrl"
            :src="websiteLogoUrl"
            alt="logo"
            class="logo-image"
          />
          <h3>{{ systemTitle }}</h3>
        </div>
        <img
          v-else-if="websiteLogoUrl"
          :src="websiteLogoUrl"
          alt="logo"
          class="logo-image"
        />
        <el-icon v-else><Setting /></el-icon>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        :collapse="appStore.sidebarCollapsed"
        :collapse-transition="false"
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
          :popper-class="appStore.sidebarCollapsed ? 'sidebar-collapse-submenu-popper' : ''"
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
            @click="handleToggleSidebar"
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
            popper-class="layout-size-popper"
          >
            <el-button type="text" class="header-btn" :title="$t('header.layout_size')">
              <el-icon class="header-icon-fixed"><Operation /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="large" :class="{ 'is-active': appStore.layoutSize === 'large' }">
                  <span class="layout-size-option">
                    <span class="layout-size-option-left">
                      <span class="layout-density layout-density-large" aria-hidden="true">
                        <i></i><i></i><i></i>
                      </span>
                      <span>{{ $t('header.layout_size_large') }}</span>
                    </span>
                    <el-icon v-if="appStore.layoutSize === 'large'" class="layout-size-option-check"><Check /></el-icon>
                  </span>
                </el-dropdown-item>
                <el-dropdown-item command="default" :class="{ 'is-active': appStore.layoutSize === 'default' }">
                  <span class="layout-size-option">
                    <span class="layout-size-option-left">
                      <span class="layout-density layout-density-default" aria-hidden="true">
                        <i></i><i></i><i></i>
                      </span>
                      <span>{{ $t('header.layout_size_default') }}</span>
                    </span>
                    <el-icon v-if="appStore.layoutSize === 'default'" class="layout-size-option-check"><Check /></el-icon>
                  </span>
                </el-dropdown-item>
                <el-dropdown-item command="small" :class="{ 'is-active': appStore.layoutSize === 'small' }">
                  <span class="layout-size-option">
                    <span class="layout-size-option-left">
                      <span class="layout-density layout-density-small" aria-hidden="true">
                        <i></i><i></i><i></i>
                      </span>
                      <span>{{ $t('header.layout_size_small') }}</span>
                    </span>
                    <el-icon v-if="appStore.layoutSize === 'small'" class="layout-size-option-check"><Check /></el-icon>
                  </span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <!-- 设置（导航模式、水印） -->
          <el-popover
            v-if="!isMobile"
            placement="bottom-end"
            :width="300"
            trigger="click"
            popper-class="settings-popover"
          >
            <template #reference>
              <el-button
                type="text"
                class="header-btn"
                :title="$t('header.settings')"
              >
                <el-icon class="header-icon-fixed"><Setting /></el-icon>
              </el-button>
            </template>
            <div class="settings-panel">
              <div class="settings-title">{{ $t('header.settings') }}</div>
              <div class="settings-item settings-item-menu-mode">
                <span class="settings-label">{{ $t('header.menu_mode') }}</span>
                <div class="menu-mode-toggle" role="tablist" :aria-label="$t('header.menu_mode')">
                  <button
                    type="button"
                    class="menu-mode-btn"
                    :class="{ active: appStore.menuMode === 'sidebar' }"
                    @click="appStore.setMenuMode('sidebar')"
                  >
                    <el-icon><Fold /></el-icon>
                    <span>{{ $t('header.menu_mode_sidebar') }}</span>
                  </button>
                  <button
                    type="button"
                    class="menu-mode-btn"
                    :class="{ active: appStore.menuMode === 'top' }"
                    @click="appStore.setMenuMode('top')"
                  >
                    <el-icon><Menu /></el-icon>
                    <span>{{ $t('header.menu_mode_top') }}</span>
                  </button>
                </div>
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
          <el-dropdown
            @command="handleCommand"
            class="user-dropdown"
            popper-class="user-account-popper"
            placement="bottom-end"
            :teleported="true"
          >
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
              <div class="user-account-panel">
                <div class="user-account-header">
                  <el-avatar
                    v-if="userStore.adminInfo?.avatar"
                    :size="48"
                    :src="userStore.adminInfo.avatar"
                    class="user-account-avatar"
                  >
                    <el-icon><User /></el-icon>
                  </el-avatar>
                  <el-avatar v-else :size="48" class="user-account-avatar user-account-avatar--placeholder">
                    <el-icon><User /></el-icon>
                  </el-avatar>
                  <div class="user-account-meta">
                    <div class="user-account-name-row">
                      <span class="user-account-name">{{ userAccountDisplayName }}</span>
                      <el-tag
                        v-if="userStore.isSuperAdmin"
                        size="small"
                        type="warning"
                        effect="plain"
                        class="user-account-badge"
                      >
                        {{ $t('header.super_admin') }}
                      </el-tag>
                    </div>
                    <div v-if="userAccountSubtitle" class="user-account-sub">{{ userAccountSubtitle }}</div>
                    <div v-if="userAccountDepartment" class="user-account-dept">
                      <el-icon class="user-account-dept-icon"><OfficeBuilding /></el-icon>
                      <span>{{ userAccountDepartment }}</span>
                    </div>
                    <div
                      v-if="userAccountRolePreview.visible.length || userAccountShowAllPermissionsHint"
                      class="user-account-roles"
                    >
                      <span class="user-account-roles-label">{{ $t('header.account_roles') }}</span>
                      <div class="user-account-roles-tags">
                        <template v-if="userAccountRolePreview.visible.length">
                          <el-tag
                            v-for="(name, idx) in userAccountRolePreview.visible"
                            :key="`role-${idx}-${name}`"
                            size="small"
                            effect="plain"
                            type="info"
                            class="user-account-role-tag"
                          >
                            {{ name }}
                          </el-tag>
                          <el-tag
                            v-if="userAccountRolePreview.more > 0"
                            size="small"
                            effect="plain"
                            type="info"
                            class="user-account-role-tag user-account-role-tag--more"
                          >
                            +{{ userAccountRolePreview.more }}
                          </el-tag>
                        </template>
                        <el-tag
                          v-else-if="userAccountShowAllPermissionsHint"
                          size="small"
                          effect="plain"
                          type="success"
                          class="user-account-role-tag user-account-role-tag--all"
                        >
                          {{ $t('header.all_permissions_hint') }}
                        </el-tag>
                      </div>
                    </div>
                  </div>
                </div>
                <el-dropdown-menu class="user-account-menu">
                  <el-dropdown-item command="profile" class="user-account-item">
                    <span class="user-account-item-inner">
                      <span class="user-account-item-left">
                        <el-icon class="user-account-item-icon"><User /></el-icon>
                        <span class="user-account-item-text">
                          <span class="user-account-item-title">{{ $t('header.profile') }}</span>
                          <span class="user-account-item-desc">{{ $t('header.profile_desc') }}</span>
                        </span>
                      </span>
                      <el-icon class="user-account-item-chevron"><ArrowRight /></el-icon>
                    </span>
                  </el-dropdown-item>
                  <el-dropdown-item command="logout" class="user-account-item user-account-item--logout">
                    <span class="user-account-item-inner">
                      <span class="user-account-item-left">
                        <el-icon class="user-account-item-icon user-account-item-icon--danger"><SwitchButton /></el-icon>
                        <span class="user-account-item-text">
                          <span class="user-account-item-title">{{ $t('header.logout') }}</span>
                          <span class="user-account-item-desc user-account-item-desc--danger">{{ $t('header.logout_desc') }}</span>
                        </span>
                      </span>
                    </span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </div>
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
          popper-class="top-menu-submenu-popper"
          :popper-offset="8"
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
      
      <el-main
        ref="mainContentRef"
        class="main-content"
        :class="{
          'main-content-iframe': isIframePage,
          'main-content--sidebar-narrowing': sidebarNarrowingLock
        }"
        :style="mainContentInlineStyle"
      >
        <!-- 使用 Element Plus 水印：开启时包裹内容，水印浮在内容之上 -->
        <el-watermark
          v-if="appStore.watermarkEnabled && !sidebarNarrowingLock"
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
import { computed, watch, onMounted, onUnmounted, ref, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessageBox } from 'element-plus'
import { useUserStore } from '../store/user'
import { useTabsStore } from '../store/tabs'
import { useAppStore, THEME_COLORS } from '../store/app'
import request from '../utils/request'
import { getConfigByGroup } from '../api/config'
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
  ArrowDown,
  ArrowRight,
  FullScreen,
  Aim,
  Odometer,
  Menu,
  Operation,
  OfficeBuilding,
  SwitchButton,
  Check,
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
const websiteSiteName = ref('')
const websiteSiteLogo = ref('')

const systemTitle = computed(() => {
  const name = websiteSiteName.value?.trim()
  return name || t('header.system')
})

const websiteLogoUrl = computed(() => {
  const raw = String(websiteSiteLogo.value || '').trim()
  if (!raw) return ''
  if (/^(https?:)?\/\//i.test(raw) || raw.startsWith('data:')) return raw
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL
  const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
  const normalizedPrefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
  if (apiBaseURL) {
    const base = apiBaseURL.replace(/\/+$/, '')
    if (raw.startsWith(normalizedPrefix)) return `${base}${raw}`
    if (raw.startsWith('/')) return `${base}${normalizedPrefix}${raw}`
    return `${base}${normalizedPrefix}/${raw}`
  }
  if (raw.startsWith(normalizedPrefix)) return raw
  if (raw.startsWith('/')) return `${normalizedPrefix}${raw}`
  return `${normalizedPrefix}/${raw}`
})

const userAccountDisplayName = computed(() => {
  const u = userStore.adminInfo
  if (!u) return ''
  return u.nickname || u.username || ''
})

const userAccountSubtitle = computed(() => {
  const u = userStore.adminInfo
  if (!u) return ''
  if (u.email) return u.email
  if (u.username) return `@${u.username}`
  if (u.phone) return u.phone
  return ''
})

const userAccountDepartment = computed(() => {
  const u = userStore.adminInfo
  if (!u) return ''
  const d = u.department
  return typeof d === 'string' ? d : d?.name || ''
})

const userAccountRoleNames = computed(() => {
  const roles = userStore.adminInfo?.roles
  if (!Array.isArray(roles) || roles.length === 0) return []
  return roles
    .map((r) => r.name || r.Name || r.slug || r.Slug || '')
    .filter(Boolean)
})

const userAccountRolePreview = computed(() => {
  const names = userAccountRoleNames.value
  return {
    visible: names.slice(0, 2),
    more: Math.max(0, names.length - 2)
  }
})

/** 超级管理员且未携带角色列表时，展示「全部权限」说明 */
const userAccountShowAllPermissionsHint = computed(
  () => userStore.isSuperAdmin && userAccountRoleNames.value.length === 0
)

const activeMenu = computed(() => route.path)
const isScreenLocked = ref(false)
/** 侧栏变窄瞬间：隔离主内容区布局，避免表格/图表随宽度突变触发巨量同步重排 */
const sidebarNarrowingLock = ref(false)
/** 收起时分离“菜单折叠”和“侧栏宽度收缩”，避免同一帧双重重排 */
const sidebarShrinkDeferred = ref(false)
const sidebarEffectiveCollapsed = computed(() => appStore.sidebarCollapsed && !sidebarShrinkDeferred.value)
/** 主内容区宽度锁定：避免收起过程触发连续 resize 级联重算 */
const mainContentRef = ref(null)
const mainContentWidthLock = ref('')
const mainContentInlineStyle = computed(() => (
  mainContentWidthLock.value
    ? { width: mainContentWidthLock.value, maxWidth: mainContentWidthLock.value }
    : {}
))
let sidebarNarrowingTimer = null
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

const handleToggleSidebar = () => {
  const willCollapse = !appStore.sidebarCollapsed
  const mainEl = mainContentRef.value?.$el || mainContentRef.value
  if (mainEl && typeof mainEl.clientWidth === 'number' && mainEl.clientWidth > 0) {
    mainContentWidthLock.value = `${mainEl.clientWidth}px`
  }

  if (sidebarNarrowingTimer) {
    clearTimeout(sidebarNarrowingTimer)
    sidebarNarrowingTimer = null
  }

  if (!willCollapse) {
    appStore.toggleSidebar()
    nextTick(() => {
      mainContentWidthLock.value = ''
    })
    return
  }

  // 先折叠菜单（文字/层级收起），下一帧再真正收窄侧栏宽度，降低峰值卡顿。
  sidebarShrinkDeferred.value = true
  sidebarNarrowingLock.value = true
  appStore.toggleSidebar()
  nextTick(() => {
    requestAnimationFrame(() => {
      sidebarNarrowingTimer = setTimeout(() => {
        sidebarShrinkDeferred.value = false
        sidebarNarrowingLock.value = false
        mainContentWidthLock.value = ''
        sidebarNarrowingTimer = null
      }, 160)
    })
  })
}

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

const loadWebsiteTitle = async () => {
  try {
    const res = await getConfigByGroup('website')
    const configs = res?.data?.configs
    if (!Array.isArray(configs)) return
    const siteNameConfig = configs.find((config) => {
      const key = config?.Key || config?.key
      return key === 'site_name'
    })
    const value = siteNameConfig?.Value || siteNameConfig?.value || ''
    websiteSiteName.value = typeof value === 'string' ? value : ''
    const siteLogoConfig = configs.find((config) => {
      const key = config?.Key || config?.key
      return key === 'site_logo'
    })
    const logoValue = siteLogoConfig?.Value || siteLogoConfig?.value || ''
    websiteSiteLogo.value = typeof logoValue === 'string' ? logoValue : ''
  } catch (error) {
    // 配置读取失败时回退默认标题，不阻塞页面
    websiteSiteName.value = ''
    websiteSiteLogo.value = ''
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

  loadWebsiteTitle()

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
    if (sidebarNarrowingTimer) {
      clearTimeout(sidebarNarrowingTimer)
      sidebarNarrowingTimer = null
    }
    mainContentWidthLock.value = ''
    sidebarShrinkDeferred.value = false
    sidebarNarrowingLock.value = false
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
  overflow-y: scroll;
  transition: background-color 0.22s ease;
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
  height: 58px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 14px;
  border-bottom: 1px solid color-mix(in srgb, var(--border-color-light) 72%, transparent);
  background: var(--card-bg, #fff);
}

.logo-brand {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  min-width: 0;
}

.logo-image {
  width: 24px;
  height: 24px;
  object-fit: contain;
  flex-shrink: 0;
}

.logo h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.3px;
  white-space: nowrap;
  color: var(--text-color-primary, #383853);
  opacity: 0.96;
  text-overflow: ellipsis;
  overflow: hidden;
  min-width: 0;
  text-align: left;
}

.sidebar.is-collapse .logo {
  padding: 0;
}

.sidebar.is-collapse .logo :deep(.el-icon) {
  font-size: 18px;
  color: var(--el-color-primary);
  opacity: 0.95;
}

.sidebar-menu {
  border-right: none;
  padding: 8px 8px 12px;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 240px;
}

/* 折叠菜单专用：统一一条垂直中心线（不依赖 el-sub-menu 内部结构） */
.sidebar-menu--collapsed :deep(.el-menu-item),
.sidebar-menu--collapsed :deep(.el-sub-menu__title) {
  width: 100%;
  height: 40px;
  min-height: 40px;
  line-height: 40px;
  margin: 3px 0;
  padding: 0 !important;
  text-indent: 0 !important;
  justify-content: center !important;
  align-items: center !important;
  position: relative;
}

.sidebar-menu--collapsed :deep(.el-menu-item > span),
.sidebar-menu--collapsed :deep(.el-sub-menu__title > span),
.sidebar-menu--collapsed :deep(.el-sub-menu__icon-arrow) {
  display: none !important;
}

.sidebar-menu--collapsed :deep(.el-menu-item .el-icon),
.sidebar-menu--collapsed :deep(.el-sub-menu__title .el-icon),
.sidebar-menu--collapsed :deep(.menu-icon) {
  position: absolute;
  left: 50% !important;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 18px;
  min-width: 18px;
  margin: 0 !important;
  text-align: center;
}

.sidebar-menu--collapsed :deep(.el-sub-menu .el-menu-item) {
  margin-left: 0 !important;
  padding-left: 0 !important;
}

.sidebar-menu--collapsed :deep(.el-menu-item.is-active::before),
.sidebar-menu--collapsed :deep(.el-sub-menu__title.is-active::before) {
  display: none !important;
}

/* 侧栏菜单：胶囊化 + 更清晰层次 */
.sidebar-menu :deep(.el-menu-item),
.sidebar-menu :deep(.el-sub-menu__title) {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  height: 38px;
  min-height: 38px;
  line-height: 38px;
  border-radius: 10px;
  margin: 5px 0;
  padding: 0 12px !important;
  overflow: hidden;
  transition: background-color 0.16s ease, color 0.16s ease;
}

.sidebar-menu :deep(.el-menu-item:hover),
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  background: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  color: var(--el-color-primary);
}

.sidebar-menu :deep(.el-menu-item.is-active),
.sidebar-menu :deep(.el-sub-menu__title.is-active) {
  color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 14%, transparent);
  font-weight: 600;
  position: relative;
}

.sidebar-menu :deep(.el-menu-item.is-active::before),
.sidebar-menu :deep(.el-sub-menu__title.is-active::before) {
  content: '';
  position: absolute;
  left: 4px;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 999px;
  background: var(--el-color-primary);
  box-shadow: none;
}

.sidebar-menu :deep(.el-sub-menu .el-menu-item) {
  margin-left: 6px;
}

/* 关闭侧栏子菜单展开过渡，避免展开时上下抖动 */
.sidebar-menu :deep(.el-menu--inline) {
  transition: none !important;
}

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

.sidebar-menu :deep(.el-menu-item .el-icon),
.sidebar-menu :deep(.el-sub-menu__title .el-icon) {
  flex-shrink: 0;
  margin-left: 0;
  margin-right: 10px;
  font-size: 16px;
  opacity: 0.9;
}

.sidebar-menu :deep(.el-sub-menu__icon-arrow) {
  flex-shrink: 0;
  margin-left: auto;
  margin-right: 0;
  width: 16px;
  text-align: right;
  opacity: 0.75;
}

/* 折叠态下图标居中，避免左右跳动 */
.sidebar-menu.el-menu--collapse :deep(.el-menu-item),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__title) {
  display: flex !important;
  width: 100%;
  height: 40px;
  min-height: 40px;
  line-height: 40px;
  margin: 5px 0;
  position: relative;
  justify-content: center !important;
  align-items: center !important;
  padding: 0 !important;
  padding-left: 0 !important;
  padding-right: 0 !important;
  text-indent: 0 !important;
}

.sidebar-menu.el-menu--collapse :deep(.el-menu-item .el-icon),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__title .el-icon) {
  position: absolute !important;
  left: 50% !important;
  top: 50% !important;
  transform: translate(-50%, -50%) !important;
  width: 18px;
  min-width: 18px;
  text-align: center;
  font-size: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  margin: 0 !important;
  margin-left: 0 !important;
  margin-right: 0 !important;
}

/* 折叠后移除子级菜单的层级缩进，确保图标同列对齐 */
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu .el-menu-item) {
  margin-left: 0 !important;
  padding-left: 0 !important;
}

/* 折叠态隐藏文案与下拉箭头占位，避免不同类型菜单图标偏移 */
.sidebar-menu.el-menu--collapse :deep(.el-menu-item > span),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__title > span),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__icon-arrow) {
  display: none !important;
}

/* MenuItem 组件内的 menu-icon 在折叠态也强制居中 */
.sidebar-menu.el-menu--collapse :deep(.menu-icon) {
  margin: 0 !important;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

/* 折叠态移除左侧激活条，避免视觉中心偏左 */
.sidebar-menu.el-menu--collapse :deep(.el-menu-item.is-active::before),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__title.is-active::before) {
  display: none;
}

/* 折叠态禁用菜单项过渡，避免折叠时卡顿 */
.sidebar-menu.el-menu--collapse :deep(.el-menu-item),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__title),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__icon-arrow) {
  transition: none !important;
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
  transition: all 0.25s ease;
}

.collapse-btn:hover {
  color: var(--el-color-primary);
  background-color: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  transform: translateY(-1px);
}

.collapse-btn:active {
  transform: scale(0.96);
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
  transition: all 0.25s ease;
}

.header-icon-fixed {
  font-size: var(--topbar-icon-size, 19px);
  opacity: var(--topbar-icon-opacity, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.25s ease;
}

.header-btn:hover {
  background-color: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  color: var(--el-color-primary);
  transform: translateY(-1px);
}

.header-btn:hover .header-icon-fixed {
  transform: scale(1.08);
}

.header-btn:active {
  transform: scale(0.96);
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
  border-radius: 10px;
  padding: 4px 8px;
  transition: all 0.25s ease;
}

.user-info:hover {
  color: var(--el-color-primary);
  background-color: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
}

.user-info:active {
  transform: scale(0.98);
}

.user-info .el-icon--right {
  transition: transform 0.25s ease;
}

.user-info:hover .el-icon--right {
  transform: translateY(1px);
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

/* 账号下拉：内容区（与设置/布局弹层风格一致） */
.user-account-panel {
  min-width: 268px;
  max-width: 320px;
}

.user-account-header {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 16px 14px 14px;
  border-bottom: 1px solid color-mix(in srgb, var(--border-color-light) 72%, transparent);
  background: linear-gradient(
    165deg,
    color-mix(in srgb, var(--el-color-primary) 10%, transparent) 0%,
    transparent 55%
  );
}

.user-account-avatar {
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.12);
}

.user-account-avatar--placeholder {
  background: color-mix(in srgb, var(--el-color-primary) 16%, var(--card-bg, #fff) 84%);
  color: var(--el-color-primary);
}

.user-account-meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-top: 2px;
}

.user-account-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.user-account-name {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.2px;
  color: var(--text-color-primary);
  line-height: 1.25;
}

.user-account-badge {
  flex-shrink: 0;
  border-radius: 6px;
  font-weight: 600;
}

.user-account-sub {
  font-size: 12px;
  color: var(--text-color-secondary);
  line-height: 1.35;
  word-break: break-all;
}

.user-account-dept {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-top: 2px;
  font-size: 12px;
  color: var(--text-color-regular);
  max-width: 100%;
}

.user-account-dept-icon {
  font-size: 14px;
  flex-shrink: 0;
  opacity: 0.85;
}

.user-account-dept span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-account-roles {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 4px;
}

.user-account-roles-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.4px;
  text-transform: uppercase;
  color: var(--text-color-placeholder);
}

.user-account-roles-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.user-account-role-tag {
  max-width: 100%;
  border-radius: 6px;
  font-weight: 500;
}

.user-account-role-tag :deep(.el-tag__content) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 118px;
}

.user-account-role-tag--more {
  font-weight: 700;
}

.user-account-menu {
  padding: 8px 8px 10px !important;
  border: none !important;
  box-shadow: none !important;
  background: transparent !important;
}

.user-account-menu :deep(.el-dropdown-menu__item) {
  padding: 0 !important;
  margin: 0;
  border-radius: 10px;
  background: transparent !important;
}

.user-account-menu :deep(.el-dropdown-menu__item + .el-dropdown-menu__item) {
  margin-top: 6px;
}

.user-account-item-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid transparent;
  transition: background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease;
}

.user-account-item-left {
  display: inline-flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 0;
}

.user-account-item-icon {
  font-size: 18px;
  margin-top: 1px;
  color: var(--text-color-regular);
  flex-shrink: 0;
}

.user-account-item-icon--danger {
  color: var(--el-color-danger);
}

.user-account-item-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.user-account-item-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
  line-height: 1.3;
}

.user-account-item-desc {
  font-size: 12px;
  color: var(--text-color-secondary);
  line-height: 1.35;
}

.user-account-item-desc--danger {
  color: color-mix(in srgb, var(--el-color-danger) 75%, var(--text-color-secondary) 25%);
}

.user-account-item-chevron {
  font-size: 14px;
  flex-shrink: 0;
  color: var(--text-color-placeholder);
  transition: transform 0.2s ease, color 0.2s ease;
}

.user-account-menu :deep(.el-dropdown-menu__item:not(.is-disabled):hover) .user-account-item-inner {
  background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
  border-color: color-mix(in srgb, var(--el-color-primary) 22%, transparent);
}

.user-account-menu :deep(.el-dropdown-menu__item:not(.is-disabled):hover) .user-account-item-chevron {
  color: var(--el-color-primary);
  transform: translateX(2px);
}

.user-account-menu :deep(.el-dropdown-menu__item.user-account-item--logout:not(.is-disabled):hover) .user-account-item-inner {
  background: color-mix(in srgb, var(--el-color-danger) 10%, transparent);
  border-color: color-mix(in srgb, var(--el-color-danger) 28%, transparent);
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

/* 侧栏收起导致主区域变宽时，先隔离主内容布局，减轻表格/图表等同步重排 */
.main-content.main-content--sidebar-narrowing {
  contain: paint;
  overflow: hidden;
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
  padding: 4px 2px 2px;
}
.settings-title {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.2px;
  color: var(--text-color-primary);
  margin-bottom: 14px;
  padding-bottom: 10px;
  border-bottom: 1px solid color-mix(in srgb, var(--border-color-light) 72%, transparent);
}
.settings-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  gap: 12px;
  padding: 10px 10px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--bg-color-tertiary) 58%, transparent);
  border: 1px solid color-mix(in srgb, var(--border-color-light) 55%, transparent);
  transition: background-color 0.2s ease, border-color 0.2s ease;
}
.settings-item:hover {
  background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
  border-color: color-mix(in srgb, var(--el-color-primary) 30%, transparent);
}
.settings-item:last-child {
  margin-bottom: 0;
}
.settings-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-regular);
  flex-shrink: 0;
}
.menu-mode-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px;
  border-radius: 11px;
  background: color-mix(in srgb, var(--card-bg, #fff) 75%, var(--bg-color-tertiary) 25%);
  border: 1px solid color-mix(in srgb, var(--border-color-light) 72%, transparent);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.35);
}
.menu-mode-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: none;
  background: transparent;
  color: var(--text-color-regular);
  border-radius: 9px;
  height: 32px;
  padding: 0 16px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.2px;
  white-space: nowrap;
  cursor: pointer;
  transition: all 0.22s ease;
}
.menu-mode-btn .el-icon {
  font-size: 14px;
  opacity: 0.85;
  transition: transform 0.2s ease, opacity 0.2s ease;
}
.menu-mode-btn:hover {
  color: var(--text-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
}
.menu-mode-btn:hover .el-icon {
  opacity: 1;
  transform: scale(1.06);
}
.menu-mode-btn.active {
  color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 14%, var(--card-bg, #fff) 86%);
  box-shadow: 0 4px 10px rgba(64, 158, 255, 0.18);
}
.menu-mode-btn.active .el-icon {
  opacity: 1;
}
.menu-mode-btn:active {
  transform: scale(0.97);
}
.settings-item-menu-mode {
  align-items: flex-start;
  flex-wrap: wrap;
}
.settings-item-menu-mode .settings-label {
  width: 100%;
}
.settings-item-menu-mode .menu-mode-toggle {
  width: 100%;
  justify-content: space-between;
}
.settings-item-menu-mode .menu-mode-btn {
  flex: 1;
  min-width: 0;
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
  gap: 10px;
}
.theme-swatch {
  width: 22px;
  height: 22px;
  border-radius: 7px;
  border: 2px solid transparent;
  cursor: pointer;
  padding: 0;
  flex-shrink: 0;
  transition: transform 0.15s ease, border-color 0.15s ease;
  box-shadow: 0 3px 8px rgba(15, 23, 42, 0.2);
}
.theme-swatch:hover {
  transform: translateY(-1px) scale(1.08);
}
.theme-swatch.active {
  border-color: var(--text-color-primary);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--el-color-primary) 22%, transparent);
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

/* 顶部菜单：子菜单标题与下拉箭头留出间距（避免与文字挤在一起） */
.layout-container.layout-top-menu .top-menu :deep(.el-sub-menu__title) {
  display: inline-flex !important;
  align-items: center;
  gap: 10px;
  padding: 0 18px 0 14px !important;
  box-sizing: border-box;
}
.layout-container.layout-top-menu .top-menu :deep(.el-sub-menu__title > .el-tooltip__trigger) {
  min-width: 0;
  flex: 1;
  overflow: hidden;
}
.layout-container.layout-top-menu .top-menu :deep(.el-sub-menu__title .menu-title) {
  margin-right: 2px;
}
.layout-container.layout-top-menu .top-menu :deep(.el-sub-menu__icon-arrow) {
  position: static !important;
  margin-left: 0 !important;
  margin-right: 0 !important;
  right: auto !important;
  top: auto !important;
  flex-shrink: 0;
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
  height: 58px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid color-mix(in srgb, var(--border-color-light) 72%, transparent);
  padding: 0 14px;
  background: var(--card-bg, #fff);
}

.drawer-content .logo h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.3px;
  color: var(--text-color-primary, #383853);
  opacity: 0.96;
  text-overflow: ellipsis;
  overflow: hidden;
  min-width: 0;
  text-align: left;
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
/* 夜间模式：侧栏菜单维持清晰层次（仅侧栏，顶部菜单保持蓝色标题不变） */
html.dark .sidebar-menu .el-menu-item:hover,
html.dark .sidebar-menu .el-sub-menu__title:hover {
  background-color: rgba(255, 255, 255, 0.08) !important;
  color: var(--el-color-primary) !important;
}
html.dark .drawer-content .sidebar-menu .el-menu-item:hover,
html.dark .drawer-content .sidebar-menu .el-sub-menu__title:hover {
  background-color: rgba(255, 255, 255, 0.08) !important;
  color: var(--el-color-primary) !important;
}
html.dark .sidebar-menu .el-menu-item.is-active,
html.dark .sidebar-menu .el-sub-menu__title.is-active {
  background: color-mix(in srgb, var(--el-color-primary) 22%, rgba(255, 255, 255, 0.05)) !important;
}
html.dark .sidebar-menu .el-menu-item.is-active::before,
html.dark .sidebar-menu .el-sub-menu__title.is-active::before {
  box-shadow: none;
}
html.dark .logo,
html.dark .drawer-content .logo {
  border-bottom-color: rgba(255, 255, 255, 0.1) !important;
  background: var(--card-bg, #1d1e1f) !important;
}
/* 布局大小下拉：质感升级 */
.layout-size-popper.el-popper {
  border: 1px solid color-mix(in srgb, var(--border-color-light) 70%, transparent) !important;
  border-radius: 12px !important;
  background: color-mix(in srgb, var(--card-bg, #fff) 96%, transparent) !important;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.14), 0 2px 8px rgba(15, 23, 42, 0.08) !important;
  padding: 6px !important;
  backdrop-filter: blur(8px);
}
.layout-size-popper .el-dropdown-menu {
  border: none !important;
  box-shadow: none !important;
  background: transparent !important;
}
.layout-size-popper .el-dropdown-menu__item {
  border-radius: 9px;
  padding: 9px 12px;
  margin: 2px 0;
  font-size: 13px;
  color: var(--text-color-regular);
  transition: all 0.2s ease;
}
.layout-size-popper .el-dropdown-menu__item:hover {
  background: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  color: var(--el-color-primary);
}
.layout-size-popper .el-dropdown-menu__item.is-active {
  background: color-mix(in srgb, var(--el-color-primary) 14%, var(--card-bg, #fff) 86%);
  color: var(--el-color-primary);
  font-weight: 600;
}
.layout-size-popper .el-dropdown-menu__item .el-icon {
  font-size: 15px;
}
.layout-size-option {
  width: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.layout-size-option-left {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.layout-size-option-check {
  font-size: 14px;
  opacity: 0.95;
}
.layout-density {
  width: 16px;
  height: 14px;
  display: inline-flex;
  flex-direction: column;
  justify-content: center;
  gap: 2px;
  opacity: 0.88;
}
.layout-density i {
  display: block;
  height: 2px;
  border-radius: 10px;
  background: currentColor;
}
.layout-density-large i:nth-child(1) { width: 14px; }
.layout-density-large i:nth-child(2) { width: 11px; }
.layout-density-large i:nth-child(3) { width: 8px; }
.layout-density-default i:nth-child(1) { width: 12px; }
.layout-density-default i:nth-child(2) { width: 9px; }
.layout-density-default i:nth-child(3) { width: 6px; }
.layout-density-small i:nth-child(1) { width: 10px; }
.layout-density-small i:nth-child(2) { width: 7px; }
.layout-density-small i:nth-child(3) { width: 4px; }
/* 设置弹窗：更有质感的卡片层次 */
.settings-popover.el-popover {
  border-radius: 14px !important;
  border: 1px solid color-mix(in srgb, var(--border-color-light) 70%, transparent) !important;
  background: color-mix(in srgb, var(--card-bg, #fff) 96%, transparent) !important;
  box-shadow: 0 14px 34px rgba(15, 23, 42, 0.14), 0 2px 8px rgba(15, 23, 42, 0.08) !important;
  padding: 12px 12px 10px !important;
  backdrop-filter: blur(8px);
}
html.dark .settings-popover.el-popover {
  border-color: rgba(255, 255, 255, 0.12) !important;
  background: color-mix(in srgb, var(--card-bg, #1d1e1f) 92%, transparent) !important;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.5), 0 2px 8px rgba(0, 0, 0, 0.35) !important;
}
html.dark .layout-size-popper.el-popper {
  border-color: rgba(255, 255, 255, 0.12) !important;
  background: color-mix(in srgb, var(--card-bg, #1d1e1f) 92%, transparent) !important;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.5), 0 2px 8px rgba(0, 0, 0, 0.35) !important;
}
html.dark .layout-size-popper .el-dropdown-menu__item:hover {
  background: rgba(255, 255, 255, 0.08);
}
html.dark .layout-size-popper .el-dropdown-menu__item.is-active {
  background: color-mix(in srgb, var(--el-color-primary) 20%, rgba(255, 255, 255, 0.06));
}
html.dark .menu-mode-toggle {
  background: color-mix(in srgb, var(--card-bg, #1d1e1f) 70%, #000 30%);
  border-color: rgba(255, 255, 255, 0.1);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06);
}
html.dark .menu-mode-btn:hover {
  background: rgba(255, 255, 255, 0.08);
}
html.dark .menu-mode-btn.active {
  background: color-mix(in srgb, var(--el-color-primary) 18%, rgba(255, 255, 255, 0.06));
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.35);
}
html.dark .header {
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.28);
}
html.dark .top-menu {
  box-shadow: 0 10px 18px rgba(0, 0, 0, 0.26);
}

/* 顶部横向菜单：子菜单弹出层（卡片质感，与各下拉统一） */
.el-popper.top-menu-submenu-popper {
  padding: 6px !important;
  border-radius: 12px !important;
  border: 1px solid color-mix(in srgb, var(--border-color-light) 70%, transparent) !important;
  background: color-mix(in srgb, var(--card-bg, #fff) 96%, transparent) !important;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.14), 0 2px 8px rgba(15, 23, 42, 0.08) !important;
  overflow: hidden;
  backdrop-filter: blur(8px);
}
.el-popper.top-menu-submenu-popper .el-menu--popup-container {
  background: transparent !important;
  padding: 0 !important;
}
.el-popper.top-menu-submenu-popper .el-menu--popup {
  border: none !important;
  box-shadow: none !important;
  background: transparent !important;
  padding: 4px 2px;
  min-width: 172px;
}
.el-popper.top-menu-submenu-popper .el-menu-item,
.el-popper.top-menu-submenu-popper .el-sub-menu__title {
  border-radius: 8px;
  margin: 2px 6px;
  height: 38px !important;
  line-height: 38px !important;
  padding: 0 12px !important;
  transition: background-color 0.2s ease, color 0.2s ease;
}
.el-popper.top-menu-submenu-popper .el-menu-item:not(.is-disabled):hover,
.el-popper.top-menu-submenu-popper .el-sub-menu__title:hover {
  background-color: color-mix(in srgb, var(--el-color-primary) 10%, transparent) !important;
  color: var(--el-color-primary) !important;
}
.el-popper.top-menu-submenu-popper .el-menu-item.is-active {
  background: color-mix(in srgb, var(--el-color-primary) 14%, transparent) !important;
  color: var(--el-color-primary) !important;
  font-weight: 600;
}
.el-popper.top-menu-submenu-popper .el-icon {
  margin-right: 8px;
}
.el-popper.top-menu-submenu-popper .el-sub-menu__icon-arrow {
  margin-right: 0 !important;
  margin-left: 6px !important;
}
html.dark .el-popper.top-menu-submenu-popper {
  border-color: rgba(255, 255, 255, 0.12) !important;
  background: color-mix(in srgb, var(--card-bg, #1d1e1f) 92%, transparent) !important;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.5), 0 2px 8px rgba(0, 0, 0, 0.35) !important;
}
html.dark .el-popper.top-menu-submenu-popper .el-menu-item:not(.is-disabled):hover,
html.dark .el-popper.top-menu-submenu-popper .el-sub-menu__title:hover {
  background: rgba(255, 255, 255, 0.08) !important;
  color: var(--el-color-primary) !important;
}
html.dark .el-popper.top-menu-submenu-popper .el-menu-item.is-active {
  background: color-mix(in srgb, var(--el-color-primary) 20%, rgba(255, 255, 255, 0.06)) !important;
}

/* 左侧收起态：子菜单弹层风格（与全局卡片质感统一） */
.el-popper.sidebar-collapse-submenu-popper {
  padding: 6px !important;
  border-radius: 12px !important;
  border: none !important;
  background: color-mix(in srgb, var(--card-bg, #fff) 96%, transparent) !important;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.14), 0 2px 8px rgba(15, 23, 42, 0.08) !important;
  overflow: hidden;
  backdrop-filter: blur(8px);
}
.el-popper.sidebar-collapse-submenu-popper .el-menu--popup-container {
  background: transparent !important;
  padding: 0 !important;
}
.el-popper.sidebar-collapse-submenu-popper .el-menu--popup {
  border: none !important;
  box-shadow: none !important;
  background: transparent !important;
  padding: 4px 2px;
  min-width: 172px;
}
.el-popper.sidebar-collapse-submenu-popper .el-menu-item,
.el-popper.sidebar-collapse-submenu-popper .el-sub-menu__title {
  border-radius: 8px;
  margin: 2px 6px;
  height: 38px !important;
  line-height: 38px !important;
  padding: 0 12px !important;
  transition: background-color 0.16s ease, color 0.16s ease;
}
.el-popper.sidebar-collapse-submenu-popper .el-menu-item:not(.is-disabled):hover,
.el-popper.sidebar-collapse-submenu-popper .el-sub-menu__title:hover {
  background-color: color-mix(in srgb, var(--el-color-primary) 10%, transparent) !important;
  color: var(--el-color-primary) !important;
}
.el-popper.sidebar-collapse-submenu-popper .el-menu-item.is-active {
  background: color-mix(in srgb, var(--el-color-primary) 14%, transparent) !important;
  color: var(--el-color-primary) !important;
  font-weight: 600;
}
.el-popper.sidebar-collapse-submenu-popper .el-icon {
  margin-right: 8px;
}
.el-popper.sidebar-collapse-submenu-popper .el-sub-menu__icon-arrow {
  margin-right: 0 !important;
  margin-left: 6px !important;
}
html.dark .el-popper.sidebar-collapse-submenu-popper {
  border: none !important;
  background: color-mix(in srgb, var(--card-bg, #1d1e1f) 92%, transparent) !important;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.5), 0 2px 8px rgba(0, 0, 0, 0.35) !important;
}
html.dark .el-popper.sidebar-collapse-submenu-popper .el-menu-item:not(.is-disabled):hover,
html.dark .el-popper.sidebar-collapse-submenu-popper .el-sub-menu__title:hover {
  background: rgba(255, 255, 255, 0.08) !important;
  color: var(--el-color-primary) !important;
}
html.dark .el-popper.sidebar-collapse-submenu-popper .el-menu-item.is-active {
  background: color-mix(in srgb, var(--el-color-primary) 20%, rgba(255, 255, 255, 0.06)) !important;
}

/* 账号下拉外层 popper（与设置/布局一致） */
.user-account-popper.el-popper {
  padding: 0 !important;
  border-radius: 14px !important;
  border: 1px solid color-mix(in srgb, var(--border-color-light) 70%, transparent) !important;
  background: color-mix(in srgb, var(--card-bg, #fff) 96%, transparent) !important;
  box-shadow: 0 14px 34px rgba(15, 23, 42, 0.14), 0 2px 8px rgba(15, 23, 42, 0.08) !important;
  overflow: hidden;
  backdrop-filter: blur(8px);
}
html.dark .user-account-popper.el-popper {
  border-color: rgba(255, 255, 255, 0.12) !important;
  background: color-mix(in srgb, var(--card-bg, #1d1e1f) 92%, transparent) !important;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.5), 0 2px 8px rgba(0, 0, 0, 0.35) !important;
}
html.dark .user-account-popper .user-account-header {
  border-bottom-color: rgba(255, 255, 255, 0.1);
}
html.dark .user-account-popper .user-account-avatar--placeholder {
  background: color-mix(in srgb, var(--el-color-primary) 22%, rgba(255, 255, 255, 0.06));
}
html.dark .user-account-popper .user-account-roles-label {
  color: var(--el-text-color-placeholder, #6c6e72);
}
</style>
