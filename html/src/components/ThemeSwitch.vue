<template>
  <el-button
    type="text"
    class="theme-switch"
    @click="handleToggle"
    :title="currentTheme === 'light' ? $t('header.switch_to_dark') : $t('header.switch_to_light')"
  >
    <el-icon class="theme-icon">
      <Sunny v-if="currentTheme === 'dark'" />
      <Moon v-else />
    </el-icon>
  </el-button>
</template>

<script setup>
import { computed } from 'vue'
import { useAppStore } from '../store/app'
import { Sunny, Moon } from '@element-plus/icons-vue'

const appStore = useAppStore()

const currentTheme = computed(() => appStore.theme)

const handleToggle = () => {
  appStore.toggleTheme()
}
</script>

<style scoped>
.theme-switch {
  color: #606266;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.3s;
  margin: 0;
}

.theme-switch:hover {
  background-color: #f5f7fa;
  color: #409EFF;
}

.theme-icon {
  font-size: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 夜间模式样式 */
.dark-mode .theme-switch {
  color: #e5eaf3;
}

.dark-mode .theme-switch:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: #409EFF;
}
</style>

