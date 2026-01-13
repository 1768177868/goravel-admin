<template>
  <!-- <el-button
    type="text"
    class="dark-mode-switch"
    @click="toggleDarkMode"
    :title="isDarkMode ? $t('header.switch_to_light') : $t('header.switch_to_dark')"
  >
    <el-icon class="dark-mode-icon">
      <Sunny v-if="isDarkMode" />
      <Moon v-else />
    </el-icon>
  </el-button> -->
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
  color: #606266;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.3s;
  margin: 0;
}

.dark-mode-switch:hover {
  background-color: #f5f7fa;
  color: #409EFF;
}

.dark-mode-icon {
  font-size: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>

