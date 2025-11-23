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
        const userStore = useUserStore()
        const tabsStore = useTabsStore()
        
        // 立即清除用户状态
        userStore.logout(true)
        
        // 清除标签页
        tabsStore.removeAllTabs()
        
        // 跳转到登录页
        if (router.currentRoute.value.path !== '/login') {
          router.replace('/login').then(() => {
            ElMessage.error(res.message || '未登录或登录已过期，请重新登录')
          }).catch(() => {
            // 如果路由跳转失败，直接使用 window.location
            window.location.href = '/login'
          })
        }
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
        const userStore = useUserStore()
        const tabsStore = useTabsStore()
        
        // 立即清除用户状态（跳过 API 调用，避免循环）
        userStore.logout(true)
        
        // 清除标签页
        tabsStore.removeAllTabs()
        
        // 跳转到登录页（使用 replace 避免历史记录问题）
        if (router.currentRoute.value.path !== '/login') {
          router.replace('/login').then(() => {
            ElMessage.error('未登录或登录已过期，请重新登录')
          }).catch(() => {
            // 如果路由跳转失败，直接使用 window.location
            window.location.href = '/login'
          })
        }
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

