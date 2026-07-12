import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getConfigByGroup } from '@/api/config'
import { resolvePublicAssetUrl } from '@/utils/env'

export function useLayoutWebsite() {
  const { t } = useI18n()
  const websiteSiteName = ref('')
  const websiteSiteLogo = ref('')
  const websiteConfigLoaded = ref(false)

  const systemTitle = computed(() => {
    if (!websiteConfigLoaded.value) return ''
    const name = websiteSiteName.value?.trim()
    return name || t('header.system')
  })

  const websiteLogoUrl = computed(() => {
    const raw = String(websiteSiteLogo.value || '').trim()
    if (!raw) return ''
    if (raw.startsWith('data:')) return raw

    const publicUrl = resolvePublicAssetUrl(raw)
    if (publicUrl.startsWith('/')) return publicUrl

    if (/^(https?:)?\/\//i.test(raw)) return raw

    const apiBaseURL = import.meta.env.VITE_API_BASE_URL
    const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
    const normalizedPrefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
    if (apiBaseURL) {
      const base = apiBaseURL.replace(/\/+$/, '')
      if (raw.startsWith(normalizedPrefix)) return `${base}${raw}`
      if (raw.startsWith('/')) return `${base}${normalizedPrefix}${raw}`
      return `${base}${normalizedPrefix}/${raw}`
    }
    if (raw.startsWith(normalizedPrefix)) return raw
    if (raw.startsWith('/')) return `${normalizedPrefix}${raw}`
    return `${normalizedPrefix}/${raw}`
  })

  const loadWebsiteTitle = async () => {
    try {
      const res = await getConfigByGroup('website')
      const configs = res?.data?.configs
      if (Array.isArray(configs)) {
        const siteNameConfig = configs.find((config) => {
          const key = config?.Key || config?.key
          return key === 'site_name'
        })
        const value = siteNameConfig?.Value || siteNameConfig?.value || ''
        websiteSiteName.value = typeof value === 'string' ? value : ''
        const siteLogoConfig = configs.find((config) => {
          const key = config?.Key || config?.key
          return key === 'site_logo'
        })
        const logoValue = siteLogoConfig?.Value || siteLogoConfig?.value || ''
        websiteSiteLogo.value = typeof logoValue === 'string' ? logoValue : ''
      } else {
        websiteSiteName.value = ''
        websiteSiteLogo.value = ''
      }
    } catch {
      websiteSiteName.value = ''
      websiteSiteLogo.value = ''
    } finally {
      websiteConfigLoaded.value = true
    }
  }

  return {
    systemTitle,
    websiteLogoUrl,
    loadWebsiteTitle
  }
}
