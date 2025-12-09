<template>
  <ErrorBoundary>
    <el-config-provider :locale="elementLocale">
      <router-view />
    </el-config-provider>
  </ErrorBoundary>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import en from 'element-plus/dist/locale/en.mjs'
import ErrorBoundary from './components/ErrorBoundary.vue'

const { locale } = useI18n()

// 根据当前语言动态返回 Element Plus 的语言配置
const elementLocale = computed(() => {
  return locale.value === 'zh-CN' ? zhCn : en
})
</script>

