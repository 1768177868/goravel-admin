import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'
import { useUserStore } from '../store/user'
import { useTabsStore } from '../store/tabs'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_PREFIX || '/api/admin',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
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
    ElMessage.error(message || '未登录或登录已过期，请重新登录')
    
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
    
    // 如果响应头中有新的 token，更新本地存储
    const newToken = response.headers.authorization || response.headers.Authorization
    if (newToken) {
      const token = newToken.replace('Bearer ', '')
      localStorage.setItem('token', token)
    }
    
    // 如果 code 不是 200，说明有错误
    if (res.code !== 200) {
      // 如果是未授权错误，也需要跳转到登录页
      if (res.code === 401 || res.message?.includes('未登录') || res.message?.includes('登录已过期') || res.message?.includes('token') || res.message?.includes('Token')) {
        handle401Error(res.message)
        return Promise.reject(new Error(res.message || '未登录或登录已过期'))
      }
      
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    
    return res
  },
  error => {
    if (error.response) {
      const { status, data } = error.response
      
      if (status === 401) {
        // 未登录或登录已过期，清除所有状态并跳转到登录页
        handle401Error(data?.message)
      } else if (status === 403) {
        ElMessage.error('没有权限访问')
      } else {
        ElMessage.error(data?.message || '请求失败')
      }
    } else {
      ElMessage.error('网络错误，请检查网络连接')
    }
    
    return Promise.reject(error)
  }
)

export default request

