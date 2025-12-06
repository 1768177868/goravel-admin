# 工具函数使用说明

## 1. Logger（日志工具）

统一日志输出，开发环境输出到控制台，生产环境可发送到日志服务。

```javascript
import logger from '@/utils/logger'

logger.log('普通日志')
logger.info('信息日志')
logger.warn('警告日志')
logger.error('错误日志')
logger.debug('调试日志')
```

## 2. ErrorHandler（错误处理）

统一错误处理，自动显示错误消息。

```javascript
import ErrorHandler from '@/utils/errorHandler'

try {
  await someApi()
} catch (error) {
  // 基本使用
  ErrorHandler.handle(error)
  
  // 使用通知而不是消息
  ErrorHandler.handle(error, { showNotification: true })
  
  // 静默处理（不显示消息）
  ErrorHandler.handle(error, { silent: true })
  
  // 自定义消息
  ErrorHandler.handle(error, { customMessage: '自定义错误消息' })
  
  // 处理 API 错误
  ErrorHandler.handleApiError(error)
}
```

## 3. Storage（存储工具）

安全的 localStorage 操作，包含错误处理和配额管理。

```javascript
import Storage from '@/utils/storage'

// 设置值
Storage.setItem('key', { data: 'value' })

// 获取值
const value = Storage.getItem('key', defaultValue)

// 删除值
Storage.removeItem('key')

// 清空所有
Storage.clear()

// 清理旧数据
Storage.clearOldData(['token', 'adminInfo'])

// 检查是否可用
if (Storage.isAvailable()) {
  // 使用存储
}

// 获取使用情况
const usage = Storage.getUsage()
```

## 4. Validation（输入验证）

提供常用的表单验证规则。

```javascript
import { validators } from '@/utils/validation'

// 在 Element Plus 表单中使用
const rules = {
  username: [
    { validator: validators.required },
    { validator: validators.minLength(3) },
    { validator: validators.maxLength(20) }
  ],
  email: [
    { validator: validators.required },
    { validator: validators.email }
  ],
  phone: [
    { validator: validators.required },
    { validator: validators.phone }
  ],
  password: [
    { validator: validators.required },
    { validator: validators.password(8) }
  ]
}
```

## 5. XSS（XSS 防护）

提供 HTML 转义和清理功能。

```javascript
import { escapeHtml, sanitizeHtml, isSafeUrl, sanitizeInput } from '@/utils/xss'

// HTML 转义
const safe = escapeHtml('<script>alert("xss")</script>')

// 清理 HTML
const clean = sanitizeHtml('<div onclick="alert(1)">content</div>')

// 验证 URL 是否安全
if (isSafeUrl(userInput)) {
  // 使用 URL
}

// 清理用户输入
const sanitized = sanitizeInput(userInput)
```

## 6. useApiRequest（请求取消）

提供请求取消机制和加载状态管理。

```javascript
import { useApiRequest } from '@/composables/useApiRequest'

const { request, cancel, loading, error } = useApiRequest()

// 执行请求（自动取消之前的请求）
const result = await request(() => getAdminList(params))

// 手动取消请求
cancel()
```

## 7. API Factory（API 工厂）

减少 API 代码重复。

```javascript
import { createCRUDApi, extendApi } from '@/utils/apiFactory'

// 创建基础 CRUD API
const baseApi = createCRUDApi('admins')

// 扩展 API，添加自定义方法
const adminApi = extendApi(baseApi, {
  resetPassword: (id, data) => {
    return request({
      url: `/admins/${id}/password`,
      method: 'put',
      data
    })
  }
})

// 导出
export const {
  list: getAdminList,
  detail: getAdminDetail,
  create: createAdmin,
  update: updateAdmin,
  delete: deleteAdmin,
  resetPassword
} = adminApi
```

## 8. Env（环境变量）

环境变量验证和获取。

```javascript
import { getEnv, getApiBaseURL, getApiPrefix, isDev, isProd } from '@/utils/env'

// 获取环境变量
const apiUrl = getEnv('VITE_API_BASE_URL', 'http://localhost:3000')

// 获取 API 基础 URL
const baseURL = getApiBaseURL()

// 检查环境
if (isDev()) {
  // 开发环境代码
}
```

