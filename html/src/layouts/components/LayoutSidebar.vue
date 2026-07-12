<template>
  <el-drawer
    :model-value="drawerVisible"
    :with-header="false"
    direction="ltr"
    :size="isMobile ? '80%' : '240px'"
    :modal="true"
    :show-close="false"
    class="mobile-drawer"
    @update:model-value="$emit('update:drawerVisible', $event)"
    @close="$emit('drawer-close')"
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
        @select="$emit('menu-select', $event)"
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

  <el-aside
    v-if="!isMobile && menuMode === 'sidebar'"
    :width="sidebarEffectiveCollapsed ? '64px' : '240px'"
    class="sidebar"
  >
    <div class="logo">
      <div v-if="!sidebarCollapsed" class="logo-brand">
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
      :collapse="sidebarCollapsed"
      :collapse-transition="false"
      @select="$emit('menu-select', $event)"
    >
      <el-menu-item index="/dashboard">
        <el-icon><Odometer /></el-icon>
        <template #title>{{ $t('menu.dashboard') }}</template>
      </el-menu-item>
      <MenuItem
        v-for="menu in menuTree"
        :key="menu.id"
        :menu="menu"
        :popper-class="sidebarCollapsed ? 'sidebar-collapse-submenu-popper' : ''"
      />
    </el-menu>
  </el-aside>
</template>

<script setup>
import { Odometer, Setting } from '@element-plus/icons-vue'
import MenuItem from '@/components/MenuItem.vue'

defineProps({
  isMobile: { type: Boolean, default: false },
  drawerVisible: { type: Boolean, default: false },
  menuMode: { type: String, default: 'sidebar' },
  sidebarCollapsed: { type: Boolean, default: false },
  sidebarEffectiveCollapsed: { type: Boolean, default: false },
  activeMenu: { type: String, default: '' },
  menuTree: { type: Array, default: () => [] },
  systemTitle: { type: String, default: '' },
  websiteLogoUrl: { type: String, default: '' }
})

defineEmits(['update:drawerVisible', 'drawer-close', 'menu-select'])
</script>
