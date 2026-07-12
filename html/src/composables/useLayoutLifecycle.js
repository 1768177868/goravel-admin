import { watch, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/store/user'
import { useTabsStore } from '@/store/tabs'
import { useAppStore } from '@/store/app'
import request from '@/utils/request'

export function useLayoutLifecycle(options = {}) {
  const { loadWebsiteTitle, cleanupSidebar } = options
  const route = useRoute()
  const userStore = useUserStore()
  const tabsStore = useTabsStore()
  const appStore = useAppStore()

  let heartbeatInterval = null

  const sendHeartbeat = async () => {
    try {
      if (userStore.token) {
        await request.get('/heartbeat')
      }
    } catch (error) {
      console.debug('Heartbeat failed:', error)
    }
  }

  watch(
    () => route.path,
    () => {
      if (route.meta.requiresAuth !== false && route.name !== 'Login') {
        tabsStore.addTab(route)
        if (route.meta?.noCache) {
          tabsStore.refreshTab(route.path)
        }
      }
    },
    { immediate: true }
  )

  onMounted(() => {
    appStore.setLayoutSize(appStore.layoutSize)

    if (route.meta.requiresAuth !== false && route.name !== 'Login') {
      tabsStore.addTab(route)
    }

    loadWebsiteTitle?.()

    appStore.isFullscreen = !!document.fullscreenElement

    const handleFullscreenChange = () => {
      appStore.isFullscreen = !!document.fullscreenElement
    }
    document.addEventListener('fullscreenchange', handleFullscreenChange)

    heartbeatInterval = setInterval(sendHeartbeat, 2 * 60 * 1000)
    sendHeartbeat()

    onUnmounted(() => {
      cleanupSidebar?.()
      document.removeEventListener('fullscreenchange', handleFullscreenChange)
      if (heartbeatInterval) {
        clearInterval(heartbeatInterval)
        heartbeatInterval = null
      }
    })
  })
}
