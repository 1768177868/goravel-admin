<template>
  <el-breadcrumb separator="›" class="breadcrumb">
    <el-breadcrumb-item :to="{ path: '/' }">
      <span class="breadcrumb-item-inner">
        <el-icon><House /></el-icon>
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
import { House } from '@element-plus/icons-vue'
import { getMenuTranslation } from '../utils/menuTranslation'

const route = useRoute()
const { t, te } = useI18n()

const breadcrumbList = computed(() => {
  const matched = route.matched.filter(item => item.meta && item.meta.titleKey)
  return matched.map(item => {
    let title = item.meta.title || item.name
    
    if (item.meta.titleKey) {
      // 优先使用 menuSlug（如果存在）进行智能翻译
      if (item.meta.menuSlug) {
        const translated = getMenuTranslation(t, te, item.meta.menuSlug)
        if (translated) {
          title = translated
        } else {
          // 如果智能翻译失败，尝试直接翻译 titleKey
          title = te(item.meta.titleKey) ? t(item.meta.titleKey) : item.meta.titleKey
        }
      } else {
        // 如果没有 menuSlug，从 titleKey 中提取 slug
        const keyParts = item.meta.titleKey.split('.')
        if (keyParts.length >= 2 && keyParts[0] === 'menu') {
          const slug = keyParts.slice(1).join('.')
          const translated = getMenuTranslation(t, te, slug)
          if (translated) {
            title = translated
          } else {
            // 如果智能翻译失败，尝试直接翻译
            title = te(item.meta.titleKey) ? t(item.meta.titleKey) : item.meta.titleKey
          }
        } else {
          // 如果不是 menu.xxx 格式，直接翻译
          title = te(item.meta.titleKey) ? t(item.meta.titleKey) : item.meta.titleKey
        }
      }
    }
    
    return {
      title,
      path: item.path !== route.path ? item.path : undefined
    }
  })
})
</script>

<style scoped>
.breadcrumb {
  margin: 0;
  line-height: 1;
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 12px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--card-bg, #fff) 88%, transparent);
  border: 1px solid color-mix(in srgb, var(--border-color-light) 68%, transparent);
  box-shadow: 0 2px 10px rgba(15, 23, 42, 0.05);
}

.breadcrumb-item-inner {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  vertical-align: middle;
}

.breadcrumb :deep(.el-breadcrumb__inner) {
  display: inline-flex;
  align-items: center;
  vertical-align: middle;
  font-size: 13px;
  line-height: 1;
  transition: color 0.2s ease;
}

.breadcrumb :deep(.el-breadcrumb__inner.is-link) {
  color: var(--text-color-regular, #606266);
  font-weight: 500;
}

.breadcrumb :deep(.el-breadcrumb__inner.is-link:hover) {
  color: var(--el-color-primary);
}

.breadcrumb :deep(.el-breadcrumb__separator) {
  margin: 0 7px;
  color: color-mix(in srgb, var(--text-color-placeholder, #909399) 88%, transparent);
  font-size: 12px;
  font-weight: 600;
  vertical-align: middle;
}

.breadcrumb :deep(.el-icon) {
  font-size: 14px;
  vertical-align: middle;
}

/* 当前页（最后一项）更突出 */
.breadcrumb :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: var(--text-color-primary, #303133);
  font-weight: 700;
}

/* 暗色模式下保持层次 */
:deep(html.dark) .breadcrumb {
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.25);
}
</style>

