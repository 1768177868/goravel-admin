<template>
  <div v-if="!isMobile && menuMode === 'top'" class="top-menu-bar">
    <el-menu
      :default-active="activeMenu"
      mode="horizontal"
      class="top-menu"
      popper-class="top-menu-submenu-popper"
      :popper-offset="8"
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
</template>

<script setup>
import { Odometer } from '@element-plus/icons-vue'
import MenuItem from '@/components/MenuItem.vue'

defineProps({
  isMobile: { type: Boolean, default: false },
  menuMode: { type: String, default: 'sidebar' },
  activeMenu: { type: String, default: '' },
  menuTree: { type: Array, default: () => [] }
})

defineEmits(['menu-select'])
</script>
