import { ref, computed, nextTick } from 'vue'
import { useAppStore } from '@/store/app'

export function useLayoutSidebar(mainContentRef) {
  const appStore = useAppStore()
  const sidebarNarrowingLock = ref(false)
  const sidebarShrinkDeferred = ref(false)
  const mainContentWidthLock = ref('')
  let sidebarNarrowingTimer = null

  const sidebarEffectiveCollapsed = computed(
    () => appStore.sidebarCollapsed && !sidebarShrinkDeferred.value
  )

  const mainContentInlineStyle = computed(() => (
    mainContentWidthLock.value
      ? { width: mainContentWidthLock.value, maxWidth: mainContentWidthLock.value }
      : {}
  ))

  const handleToggleSidebar = () => {
    const willCollapse = !appStore.sidebarCollapsed
    const mainEl = mainContentRef.value?.$el || mainContentRef.value
    if (mainEl && typeof mainEl.clientWidth === 'number' && mainEl.clientWidth > 0) {
      mainContentWidthLock.value = `${mainEl.clientWidth}px`
    }

    if (sidebarNarrowingTimer) {
      clearTimeout(sidebarNarrowingTimer)
      sidebarNarrowingTimer = null
    }

    if (!willCollapse) {
      appStore.toggleSidebar()
      nextTick(() => {
        mainContentWidthLock.value = ''
      })
      return
    }

    sidebarShrinkDeferred.value = true
    sidebarNarrowingLock.value = true
    appStore.toggleSidebar()
    nextTick(() => {
      requestAnimationFrame(() => {
        sidebarNarrowingTimer = setTimeout(() => {
          sidebarShrinkDeferred.value = false
          sidebarNarrowingLock.value = false
          mainContentWidthLock.value = ''
          sidebarNarrowingTimer = null
        }, 160)
      })
    })
  }

  const cleanupSidebar = () => {
    if (sidebarNarrowingTimer) {
      clearTimeout(sidebarNarrowingTimer)
      sidebarNarrowingTimer = null
    }
    mainContentWidthLock.value = ''
    sidebarShrinkDeferred.value = false
    sidebarNarrowingLock.value = false
  }

  return {
    sidebarNarrowingLock,
    sidebarEffectiveCollapsed,
    mainContentInlineStyle,
    handleToggleSidebar,
    cleanupSidebar
  }
}
