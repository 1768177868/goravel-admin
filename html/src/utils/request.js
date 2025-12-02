import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'
import { useUserStore } from '../store/user'
import { useTabsStore } from '../store/tabs'
import { useAppStore } from '../store/app'
import i18n from '../i18n'

const { t } = i18n.global

// 构建完整的 API baseURL
// 如果配置了 VITE_API_BASE_URL，使用它 + VITE_API_PREFIX，否则使用相对路径
const getBaseURL = () => {
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL
  const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
  
  // 如果配置了完整的基础 URL，使用它
  if (apiBaseURL) {
    // 确保 URL 格式正确（移除末尾的 /，确保前缀以 / 开头）
    const base = apiBaseURL.replace(/\/+$/, '')
    const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
    return `${base}${prefix}`
  }
  
  // 如果没有配置基础 URL，使用相对路径
  return apiPrefix
}

const request = axios.create({
  baseURL: getBaseURL(),
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      // 确保 token 没有多余的空格
      const cleanToken = token.trim()
      config.headers.Authorization = `Bearer ${cleanToken}`
    }
    
    // 设置语言请求头，将前端的语言代码转换为后端期望的格式
    const currentLocale = i18n.global.locale.value || localStorage.getItem('language') || 'zh-CN'
    // 前端使用 zh-CN/en-US，后端期望 cn/en
    let acceptLanguage = 'zh-CN'
    if (currentLocale === 'en-US') {
      acceptLanguage = 'en-US'
    } else if (currentLocale === 'zh-CN' || currentLocale === 'cn') {
      acceptLanguage = 'zh-CN'
    }
    config.headers['Accept-Language'] = acceptLanguage
    
    const appStore = useAppStore()
    let browserTimezone = 'UTC'
    try {
      browserTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
    } catch {
      browserTimezone = 'UTC'
    }
    const timezone = appStore.timezone || localStorage.getItem('timezone') || browserTimezone
    if (timezone) {
      config.headers['X-Timezone'] = timezone
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 防止重复跳转的标志
let isRedirecting = false

// 处理401错误的统一函数
const handle401Error = (message) => {
  // 如果正在跳转，直接返回
  if (isRedirecting) {
    return
  }
  
  isRedirecting = true
  const userStore = useUserStore()
  const tabsStore = useTabsStore()
  
  // 立即清除用户状态（同步执行）
  userStore.logout(true)
  
  // 清除标签页
  tabsStore.removeAllTabs()
  
  // 如果当前不在登录页，立即跳转
  const currentPath = router.currentRoute.value.path
  if (currentPath !== '/login') {
      // 显示错误消息
      ElMessage.error(message || t('error.unauthorized'))
    
    // 立即使用 router.replace 跳转
    router.replace('/login').catch(() => {
      // 如果路由跳转失败，直接使用 window.location 强制跳转
      window.location.href = '/login'
    })
  }
  
  // 清除跳转标志（延迟清除，避免立即重复）
  setTimeout(() => {
    isRedirecting = false
  }, 2000)
}

// 响应拦截器
request.interceptors.response.use(
  response => {
    const res = response.data
    const url = response.config?.url || ''
    
    // 排除登录和退出接口，这些接口返回 401 是正常的业务错误，不应该触发自动跳转
    const isAuthEndpoint = url.includes('/login') || url.includes('/logout')
    
    // 如果响应头中有新的 token，更新本地存储
    const newToken = response.headers.authorization || response.headers.Authorization
    if (newToken) {
      const token = newToken.replace('Bearer ', '').trim()
      if (token) {
        localStorage.setItem('token', token)
        // 同时更新 userStore 中的 token
        const userStore = useUserStore()
        userStore.setToken(token)
      }
    }
    
    // 如果响应数据中有 token（登录接口会在 data.token 中返回），也更新本地存储
    if (res.data && res.data.token) {
      const token = res.data.token
      if (token) {
        localStorage.setItem('token', token)
        const userStore = useUserStore()
        userStore.setToken(token)
      }
    }
    
    // 如果 code 不是 200，说明有错误
    // 注意：HTTP 403 状态码会进入错误回调，不会进入这里
    // 这里只处理 HTTP 200 但 code 不是 200 的情况
    if (res.code !== 200) {
      if (!isAuthEndpoint) {
        const message = res.message || t('error.default')
        if (res.code === 401) {
          handle401Error(message || t('error.unauthorized'))
        } else if (res.code === 403) {
          // 403 错误会在错误回调中处理，这里不重复显示
          // 但如果 HTTP 状态码是 200，说明是业务逻辑错误，需要显示
          ElMessage.error(message || t('error.forbidden'))
        } else {
          ElMessage.error(message)
        }
      }
      // 创建错误对象，包含更多信息
      const errorMessage = res.message || t('error.default') || '请求失败'
      const err = new Error(errorMessage)
      err.code = res.code
      err.data = res.data
      err.__handled = true
      return Promise.reject(err)
    }
    
    return res
  },
  error => {
    // 如果错误已经被标记为已处理，不再重复处理
    if (error.__handled) {
      return Promise.reject(error)
    }

    if (error.response) {
      const { status, data, config } = error.response
      const url = config?.url || ''

      const isAuthEndpoint = url.includes('/login') || url.includes('/logout')

      if (status === 429) {
        const message = data?.message || data?.data?.message || t('error.tooManyRequests')
        ElMessage.error(message)
      } else if (status === 401 && !isAuthEndpoint) {
        const message = data?.message || data?.data?.message || t('error.unauthorized')
        handle401Error(message)
      } else if (status === 401 && isAuthEndpoint) {
        // 登录/退出接口 401 只提示一次
        const message = data?.message || data?.data?.message || t('login.login_failed')
        ElMessage.error(message)
      } else if (status === 403 && isAuthEndpoint) {
        // 登录/退出接口 403（账号被禁用等）只提示一次
        const message = data?.message || data?.data?.message || t('error.forbidden')
        // 如果是 account_disabled，使用登录相关的翻译
        if (message === 'account_disabled') {
          ElMessage.error(t('login.account_disabled'))
        } else {
          ElMessage.error(message)
        }
      } else if (!isAuthEndpoint) {
        if (status === 403) {
          // 使用后端返回的错误消息，如果没有则使用默认翻译
          const message = data?.message || data?.data?.message || t('error.forbidden')
          ElMessage.error(message)
        } else {
          const errorMessage = data?.message || data?.data?.message || t('error.default')
          ElMessage.error(errorMessage)
        }
      }
    } else {
      // 网络错误或 CORS 错误
      let errorMessage = t('error.network')
      
      if (error.code === 'ERR_NETWORK' || error.message === 'Network Error') {
        errorMessage = t('error.network') + ' (网络连接失败，请检查 API 地址配置)'
      } else if (error.code === 'ECONNABORTED') {
        errorMessage = t('error.timeout') || '请求超时'
      } else if (error.message) {
        errorMessage = error.message
      }
      
      // 只在非静默错误时显示消息
      if (!error.config?.silent) {
        ElMessage.error(errorMessage)
      }
    }

    if (typeof error === 'object') {
      error.__handled = true
    }
    
    return Promise.reject(error)
  }
)

export default request
