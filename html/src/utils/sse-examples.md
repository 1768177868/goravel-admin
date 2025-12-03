# SSE 前端使用示例

## 1. 系统监控实时数据流

在 `Monitor.vue` 组件中使用：

```vue
<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { createSSEConnection, closeSSEConnection } from '@/utils/sse'
import { createSystemInfoSSE } from '@/api/monitor'

const systemInfo = ref(null)
let eventSource = null

onMounted(() => {
  // 创建 SSE 连接
  const url = createSystemInfoSSE({ interval: 2 })
  eventSource = createSSEConnection(url, {
    onMessage: (data) => {
      if (data.type === 'system_info') {
        systemInfo.value = data.data
      }
    },
    onError: (error) => {
      console.error('SSE error:', error)
    },
    onOpen: () => {
      console.log('SSE connected')
    }
  })
})

onUnmounted(() => {
  if (eventSource) {
    closeSSEConnection(eventSource)
  }
})
</script>
```

## 2. 文件上传进度推送

在 `AttachmentList.vue` 组件中使用：

```vue
<script setup>
import { ref } from 'vue'
import { createSSEConnection, closeSSEConnection } from '@/utils/sse'
import { createUploadProgressSSE } from '@/api/attachment'

const uploadProgress = ref(0)
let uploadEventSource = null

// 开始上传时创建 SSE 连接
const startUploadProgress = (chunkID, totalChunks) => {
  const url = createUploadProgressSSE(chunkID, totalChunks, { interval: 500 })
  uploadEventSource = createSSEConnection(url, {
    onMessage: (data) => {
      if (data.type === 'progress') {
        uploadProgress.value = data.progress
      } else if (data.type === 'completed') {
        uploadProgress.value = 100
        closeSSEConnection(uploadEventSource)
        // 显示完成消息和下载链接
        console.log('Upload completed:', data.file_url)
      } else if (data.type === 'error') {
        console.error('Upload error:', data.message)
        closeSSEConnection(uploadEventSource)
      }
    },
    onError: (error) => {
      console.error('Upload progress SSE error:', error)
    }
  })
}

// 上传完成或取消时关闭连接
const stopUploadProgress = () => {
  if (uploadEventSource) {
    closeSSEConnection(uploadEventSource)
    uploadEventSource = null
  }
}
</script>
```

## 3. 导出任务进度推送

在 `ExportList.vue` 组件中使用：

```vue
<script setup>
import { ref } from 'vue'
import { createSSEConnection, closeSSEConnection } from '@/utils/sse'
import { createExportProgressSSE } from '@/api/export'

const exportProgress = ref({})
let exportEventSource = null

// 监控导出任务进度
const monitorExportProgress = (exportID) => {
  const url = createExportProgressSSE(exportID, { interval: 1000 })
  exportEventSource = createSSEConnection(url, {
    onMessage: (data) => {
      if (data.type === 'progress') {
        exportProgress.value = {
          status: data.status_text,
          message: data.message
        }
      } else if (data.type === 'completed') {
        exportProgress.value = {
          status: '成功',
          message: '导出完成',
          file_url: data.file_url,
          filename: data.filename
        }
        closeSSEConnection(exportEventSource)
      } else if (data.type === 'failed') {
        exportProgress.value = {
          status: '失败',
          message: data.message
        }
        closeSSEConnection(exportEventSource)
      }
    },
    onError: (error) => {
      console.error('Export progress SSE error:', error)
    }
  })
}

// 停止监控
const stopExportProgress = () => {
  if (exportEventSource) {
    closeSSEConnection(exportEventSource)
    exportEventSource = null
  }
}
</script>
```

## 4. Dashboard 数据实时更新

在 `Dashboard.vue` 组件中使用：

```vue
<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { createSSEConnection, closeSSEConnection } from '@/utils/sse'
import { createDashboardSSE } from '@/api/dashboard'

const dashboardData = ref({
  count: {},
  user_access_source: [],
  weekly_user_activity: [],
  monthly_sales: [],
  online_user_count: 0
})

let dashboardEventSource = null

onMounted(() => {
  // 创建 SSE 连接
  const url = createDashboardSSE({ interval: 5 })
  dashboardEventSource = createSSEConnection(url, {
    onMessage: (data) => {
      if (data.type === 'dashboard_data') {
        dashboardData.value = data.data
        // 更新图表数据
        updateCharts(data.data)
      }
    },
    onError: (error) => {
      console.error('Dashboard SSE error:', error)
    },
    onOpen: () => {
      console.log('Dashboard SSE connected')
    }
  })
})

onUnmounted(() => {
  if (dashboardEventSource) {
    closeSSEConnection(dashboardEventSource)
  }
})

const updateCharts = (data) => {
  // 更新各种图表
  // updateCountCards(data.count)
  // updateAccessSourceChart(data.user_access_source)
  // updateWeeklyActivityChart(data.weekly_user_activity)
  // updateMonthlySalesChart(data.monthly_sales)
}
</script>
```

## 注意事项

1. **Token 传递**：由于 EventSource API 不支持自定义 headers，token 需要通过 URL 参数传递。后端需要支持从 URL 参数读取 token。

2. **连接管理**：确保在组件卸载时关闭 SSE 连接，避免内存泄漏。

3. **错误处理**：实现完善的错误处理，包括网络错误、认证错误等。

4. **自动重连**：EventSource 会自动重连，但可以在 onError 中实现自定义重连逻辑。

5. **性能优化**：根据实际需求调整推送间隔，避免过于频繁的推送。

