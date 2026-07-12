import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'
import { filterAndSortTree } from '@/utils/tree'

export function useLayoutMenu(options = {}) {
  const { isMobile, onMobileMenuSelect } = options
  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()

  const activeMenu = computed(() => route.path)

  const menuTree = computed(() => {
    const menus = userStore.menus || []
    if (menus.length === 0) return []
    return filterAndSortTree(
      menus,
      (menu) => menu.is_hidden === 0 && menu.status === 1,
      (a, b) => a.sort - b.sort
    )
  })

  const handleMenuSelect = (index) => {
    if (index && typeof index === 'string' && !index.startsWith('external-')) {
      if (!index.startsWith('http://') && !index.startsWith('https://')) {
        router.push(index)
        if (isMobile?.value && onMobileMenuSelect) {
          onMobileMenuSelect()
        }
      }
    }
  }

  return {
    activeMenu,
    menuTree,
    handleMenuSelect
  }
}
