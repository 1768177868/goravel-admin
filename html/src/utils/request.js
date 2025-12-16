import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'
import { useUserStore } from '../store/user'
import { useTabsStore } from '../store/tabs'
import { useAppStore } from '../store/app'
import i18n from '../i18n'
import logger from './logger'
import Storage from './storage'

const { t } = i18n.global

// 错误码常量
const ERROR_CODES = {
  // 认证相关
  GOOGLE_CODE_REQUIRED: 'google_code_required',
  GOOGLE_CODE_INVALID: 'google_code_invalid',
  ACCOUNT_DISABLED: 'account_disabled',
  UNAUTHORIZED: 'unauthorized',
  FORBIDDEN: 'forbidden',
  // 验证码相关
  CAPTCHA_REQUIRED: 'captcha_required',
  CAPTCHA_INVALID: 'captcha_invalid',
  CAPTCHA_EXPIRED: 'captcha_expired',
  // 通用错误
  TOO_MANY_REQUESTS: 'too_many_requests',
  NETWORK_ERROR: 'network_error',
  TIMEOUT: 'timeout'
}

// 构建完整的 API baseURL
const getBaseURL = () => {
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL
  const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
  
  if (apiBaseURL) {
    const base = apiBaseURL.replace(/\/+$/, '')
    const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
    return `${base}${prefix}`
  }
  
  return apiPrefix
}

const request = axios.create({
  baseURL: getBaseURL(),
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    const token = Storage.getItem('token', '')
    if (token) {
      config.headers.Authorization = `Bearer ${token.trim()}`
    }
    
    // 设置语言请求头
    const currentLocale = i18n.global.locale.value || Storage.getItem('language', 'zh-CN')
    let acceptLanguage = 'zh-CN'
    if (currentLocale === 'en-US') {
      acceptLanguage = 'en-US'
    } else if (currentLocale === 'zh-CN' || currentLocale === 'cn') {
      acceptLanguage = 'zh-CN'
    }
    config.headers['Accept-Language'] = acceptLanguage
    
    // 设置时区
    const appStore = useAppStore()
    let browserTimezone = 'UTC'
    try {
      browserTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
    } catch {
      browserTimezone = 'UTC'
    }
    const timezone = appStore.timezone || Storage.getItem('timezone', browserTimezone)
    if (timezone) {
      config.headers['X-Timezone'] = timezone
    }
    
    // 支持请求取消（如果传递了 signal）
    // axios 会自动处理 config.signal
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 防止重复跳转的标志
let isRedirecting = false

// 防止重复弹出403错误的标志和定时器
let last403ErrorTime = 0
let isShowing403Error = false
const FORBIDDEN_ERROR_COOLDOWN = 3000 // 3秒内的403错误只提示一次

// 处理401错误的统一函数
const handle401Error = (message) => {
  if (isRedirecting) {
    return
  }
  
  isRedirecting = true
  const userStore = useUserStore()
  const tabsStore = useTabsStore()
  
  userStore.logout(true)
  tabsStore.removeAllTabs()
  
  const currentPath = router.currentRoute.value.path
  if (currentPath !== '/login') {
    ElMessage.error(message || t('error.unauthorized'))
    router.replace('/login').catch(() => {
      window.location.href = '/login'
    })
  }
  
  setTimeout(() => {
    isRedirecting = false
  }, 2000)
}

// 处理403错误的统一函数（带防抖）
const handle403Error = (message) => {
  const now = Date.now()
  
  // 如果正在显示403错误，或者距离上次显示403错误的时间在冷却期内，则跳过
  if (isShowing403Error || (now - last403ErrorTime < FORBIDDEN_ERROR_COOLDOWN)) {
    return
  }
  
  isShowing403Error = true
  last403ErrorTime = now
  
  ElMessage.error(message || t('error.forbidden'))
  
  // 3秒后重置标志，允许再次显示403错误
  setTimeout(() => {
    isShowing403Error = false
  }, FORBIDDEN_ERROR_COOLDOWN)
}

// 处理错误响应，提取错误信息
const extractErrorInfo = (data) => {
  const message = data?.message || data?.data?.message || ''
  const errorCode = data?.error_code || data?.data?.error_code || ''
  const code = data?.code || 0
  
  return {
    message,      // 后端已翻译的消息
    errorCode,    // 错误码（如 'google_code_required'）
    code          // HTTP 状态码或业务状态码
  }
}

// 响应拦截器
request.interceptors.response.use(
  response => {
    const res = response.data
    const url = response.config?.url || ''
    const isAuthEndpoint = url.includes('/login') || url.includes('/logout')
    
    // 更新 token
    const newToken = response.headers.authorization || response.headers.Authorization
    if (newToken) {
      const token = newToken.replace('Bearer ', '').trim()
      if (token) {
        Storage.setItem('token', token)
        const userStore = useUserStore()
        userStore.setToken(token)
      }
    }
    
    if (res.data && res.data.token) {
      const token = res.data.token
      if (token) {
        Storage.setItem('token', token)
        const userStore = useUserStore()
        userStore.setToken(token)
      }
    }
    
    // 处理业务错误（HTTP 200 但 code 不是 200）
    if (res.code !== 200) {
      const { message, errorCode } = extractErrorInfo(res)
      
      if (!isAuthEndpoint) {
        if (res.code === 401) {
          handle401Error(message || t('error.unauthorized'))
        } else if (res.code === 403) {
          handle403Error(message || t('error.forbidden'))
        } else {
          // 显示后端返回的实际错误消息
          ElMessage.error(message || t('error.default'))
        }
      }
      
      // 创建错误对象
      const err = new Error(message || t('error.default'))
      err.code = res.code
      err.errorCode = errorCode
      err.data = res.data
      err.response = response // 确保 response 对象存在，方便组件访问
      err.message = message
      err.translatedMessage = message // 后端已翻译
      
      if (!isAuthEndpoint) {
        err.__handled = true // 标记为已处理，避免组件重复显示
      }
      
      return Promise.reject(err)
    }
    
    return res
  },
  error => {
    if (error.__handled) {
      return Promise.reject(error)
    }

    if (error.response) {
      const { status, data, config } = error.response
      const url = config?.url || ''
      const isAuthEndpoint = url.includes('/login') || url.includes('/logout')
      const { message, errorCode } = extractErrorInfo(data)

      // 根据 HTTP 状态码和错误码处理
      if (status === 429) {
        ElMessage.error(message || t('error.tooManyRequests'))
        error.__handled = true
      } else if (status === 401) {
        if (!isAuthEndpoint) {
          handle401Error(message || t('error.unauthorized'))
        } else {
          // 登录接口错误，不在这里显示，让 Login.vue 处理
          error.errorCode = errorCode
          error.message = message
          error.translatedMessage = message
          error.__handled = false // 明确标记为未处理，让 Login.vue 处理
        }
      } else if (status === 403) {
        if (!isAuthEndpoint) {
          handle403Error(message || t('error.forbidden'))
          error.__handled = true
        } else {
          // 登录接口错误，不在这里显示，让 Login.vue 处理
          error.errorCode = errorCode
          error.message = message
          error.translatedMessage = message
          error.__handled = false // 明确标记为未处理，让 Login.vue 处理
        }
      } else if (isAuthEndpoint && status >= 400) {
        // 登录接口的其他错误（400等），不在这里显示，让 Login.vue 处理
        error.errorCode = errorCode
        error.message = message
        error.translatedMessage = message
        error.code = status
        error.response = error.response // 确保 response 对象存在
        error.__handled = false // 明确标记为未处理，让 Login.vue 处理
      } else if (!isAuthEndpoint) {
        // 非登录接口的错误，直接显示
        ElMessage.error(message || t('error.default'))
        error.__handled = true
      }
    } else {
      // 网络错误
      let errorMessage = t('error.network')
      
      if (error.code === 'ERR_NETWORK' || error.message === 'Network Error') {
        errorMessage = t('error.network') + ' (网络连接失败，请检查 API 地址配置)'
      } else if (error.code === 'ECONNABORTED') {
        errorMessage = t('error.timeout')
      } else if (error.message) {
        errorMessage = error.message
      }
      
      if (!error.config?.silent) {
        ElMessage.error(errorMessage)
        error.__handled = true
      }
    }

    // 如果还没有标记为已处理，且不是登录接口的错误，则标记为已处理
    if (typeof error === 'object' && error.__handled !== false && error.__handled !== true) {
      // 检查是否是登录接口
      const url = error.config?.url || ''
      const isAuthEndpoint = url.includes('/login') || url.includes('/logout')
      if (!isAuthEndpoint) {
        error.__handled = true
      }
    }
    
    return Promise.reject(error)
  }
)

export default request
export { ERROR_CODES }
