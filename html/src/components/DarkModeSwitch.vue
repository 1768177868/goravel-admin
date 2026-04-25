<template>
  <el-button
    type="text"
    class="dark-mode-switch topbar-icon-btn"
    @click="toggleDarkMode"
    :title="isDarkMode ? $t('header.switch_to_light') : $t('header.switch_to_dark')"
  >
    <el-icon class="dark-mode-icon topbar-icon">
      <Sunny v-if="isDarkMode" />
      <Moon v-else />
    </el-icon>
  </el-button>
</template>

<script setup>
import { computed } from 'vue'
import { useAppStore } from '../store/app'
import { Sunny, Moon } from '@element-plus/icons-vue'

const appStore = useAppStore()

const isDarkMode = computed(() => appStore.darkMode)

const toggleDarkMode = () => {
  appStore.toggleDarkMode()
}
</script>

<style scoped>
.dark-mode-switch {
  color: var(--text-color-regular, #606266);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  margin: 0;
  min-width: 44px;
  min-height: 44px;
}

.dark-mode-icon {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 暗黑模式下的样式 */
:deep(html.dark) .dark-mode-switch {
  color: var(--el-text-color-regular);
}

:deep(html.dark) .dark-mode-switch:hover {
  background-color: var(--el-bg-color-page);
  color: var(--el-color-primary);
}
</style>

