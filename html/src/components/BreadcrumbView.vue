<template>
  <el-breadcrumb separator="/" class="breadcrumb">
    <el-breadcrumb-item :to="{ path: '/' }">
      <span class="breadcrumb-item-inner">
        <el-icon><HomeFilled /></el-icon>
        <span>{{ $t('breadcrumb.home') }}</span>
      </span>
    </el-breadcrumb-item>
    <el-breadcrumb-item
      v-for="(item, index) in breadcrumbList"
      :key="index"
      :to="item.path ? { path: item.path } : undefined"
    >
      <span class="breadcrumb-item-inner">
        <span>{{ item.title }}</span>
      </span>
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { HomeFilled } from '@element-plus/icons-vue'

const route = useRoute()
const { t } = useI18n()

const breadcrumbList = computed(() => {
  const matched = route.matched.filter(item => item.meta && item.meta.titleKey)
  return matched.map(item => ({
    title: item.meta.titleKey ? t(item.meta.titleKey) : item.meta.title || item.name,
    path: item.path !== route.path ? item.path : undefined
  }))
})
</script>

<style scoped>
.breadcrumb {
  margin: 0;
  line-height: 1;
}

.breadcrumb-item-inner {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  vertical-align: middle;
}

.breadcrumb :deep(.el-breadcrumb__inner) {
  display: inline-flex;
  align-items: center;
  vertical-align: middle;
}

.breadcrumb :deep(.el-breadcrumb__inner.is-link) {
  color: #606266;
  font-weight: normal;
}

.breadcrumb :deep(.el-breadcrumb__inner.is-link:hover) {
  color: #409EFF;
}

.breadcrumb :deep(.el-breadcrumb__separator) {
  margin: 0 8px;
  color: #c0c4cc;
  vertical-align: middle;
}

.breadcrumb :deep(.el-icon) {
  font-size: 16px;
  vertical-align: middle;
}
</style>

