import { computed } from 'vue'
import { useUserStore } from '@/store/user'

export function useLayoutAccount() {
  const userStore = useUserStore()

  const userAccountDisplayName = computed(() => {
    const u = userStore.adminInfo
    if (!u) return ''
    return u.nickname || u.username || ''
  })

  const userAccountSubtitle = computed(() => {
    const u = userStore.adminInfo
    if (!u) return ''
    if (u.email) return u.email
    if (u.username) return `@${u.username}`
    if (u.phone) return u.phone
    return ''
  })

  const userAccountDepartment = computed(() => {
    const u = userStore.adminInfo
    if (!u) return ''
    const d = u.department
    return typeof d === 'string' ? d : d?.name || ''
  })

  const userAccountRoleNames = computed(() => {
    const roles = userStore.adminInfo?.roles
    if (!Array.isArray(roles) || roles.length === 0) return []
    return roles
      .map((r) => r.name || r.Name || r.slug || r.Slug || '')
      .filter(Boolean)
  })

  const userAccountRolePreview = computed(() => {
    const names = userAccountRoleNames.value
    return {
      visible: names.slice(0, 2),
      more: Math.max(0, names.length - 2)
    }
  })

  const userAccountShowAllPermissionsHint = computed(
    () => userStore.isSuperAdmin && userAccountRoleNames.value.length === 0
  )

  return {
    userAccountDisplayName,
    userAccountSubtitle,
    userAccountDepartment,
    userAccountRolePreview,
    userAccountShowAllPermissionsHint
  }
}
