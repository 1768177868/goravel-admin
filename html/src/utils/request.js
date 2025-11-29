import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'
import { useUserStore } from '../store/user'
import { useTabsStore } from '../store/tabs'
import { useAppStore } from '../store/app'
import i18n from '../i18n'

const { t } = i18n.global

const request = axios.create({
  baseURL: import.meta.env.VITE_API_PREFIX || '/api/admin',
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
    if (res.code !== 200) {
      if (!isAuthEndpoint) {
        const message = res.message || t('error.default')
        if (res.code === 401) {
          handle401Error(message || t('error.unauthorized'))
        } else if (res.code === 403) {
          ElMessage.error(message || t('error.forbidden'))
        } else {
          ElMessage.error(message)
        }
      }
      const err = new Error(res.message || t('error.default'))
      err.__handled = true
      return Promise.reject(err)
    }
    
    return res
  },
  error => {
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
          ElMessage.error(t('error.forbidden'))
        } else {
          const errorMessage = data?.message || data?.data?.message || t('error.default')
          ElMessage.error(errorMessage)
        }
      }
    } else {
      ElMessage.error(t('error.network'))
    }

    if (typeof error === 'object') {
      error.__handled = true
    }
    
    return Promise.reject(error)
  }
)

export default request

