<template>
  <div class="iframe-container">
    <iframe
      ref="iframeRef"
      :src="iframeUrl"
      frameborder="0"
      class="iframe-content"
      @load="handleLoad"
    ></iframe>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTabsStore } from '../../store/tabs'
import Storage from '../../utils/storage'

const route = useRoute()
const router = useRouter()
const tabsStore = useTabsStore()
const iframeRef = ref(null)

// 从路由参数中获取外部链接URL和标题
const iframeUrl = computed(() => {
  return route.query.url || ''
})

const iframeTitle = computed(() => {
  return route.query.title ? decodeURIComponent(route.query.title) : '外部链接'
})

// 更新标签页标题
const updateTabTitle = () => {
  const tab = tabsStore.tabs.find(t => t.path === route.path)
  if (tab && iframeTitle.value) {
    tab.title = iframeTitle.value
    // 保存到 localStorage
    try {
      const data = { tabs: tabsStore.tabs, activeTab: tabsStore.activeTab }
      Storage.setItem('tabs', data)
    } catch (error) {
      console.error('Failed to save tabs:', error)
    }
  }
}

const handleLoad = () => {
  // iframe 加载完成后的处理
  console.log('Iframe loaded:', iframeUrl.value)
}

watch(() => route.query.title, () => {
  updateTabTitle()
}, { immediate: true })

onMounted(() => {
  if (!iframeUrl.value) {
    // 如果没有URL，返回上一页
    router.back()
  } else {
    updateTabTitle()
  }
})
</script>

<style scoped>
.iframe-container {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
}

.iframe-content {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
}
</style>

