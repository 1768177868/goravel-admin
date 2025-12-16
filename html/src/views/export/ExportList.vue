<template>
  <div class="export-list">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('export.title') }}</span>
          <div class="header-actions">
            <el-button 
              type="danger" 
              :disabled="selectedRows.length === 0"
              @click="handleBatchDelete"
            >
              <el-icon><Delete /></el-icon>
              {{ $t('common.delete_selected') }} ({{ selectedRows.length }})
            </el-button>
          </div>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :loading="loading"
        @search="handleSearch"
        @reset="handleReset"
      />

      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        stripe
        height="600"
        :sort-config="{ multiple: false, trigger: 'default' }"
        @sort-change="handleSortChange"
        @checkbox-change="handleSelectionChange"
        @checkbox-all="handleSelectionChange"
      >
        <vxe-column type="checkbox" width="60" />
        <vxe-column field="id" :title="$t('table.id')" width="80" sortable />
        <vxe-column field="filename" :title="$t('export.filename')" min-width="200" />
        <vxe-column field="disk" :title="$t('export.disk')" width="120" />
        <vxe-column field="path" :title="$t('export.path')" min-width="260" />
        <vxe-column field="extension" :title="$t('export.extension')" width="100" />
        <vxe-column field="size" :title="$t('export.size')" width="140" :formatter="formatSize" />
        <vxe-column field="status" :title="$t('log.status')" width="150">
          <template #default="{ row }">
            <div>
              <el-tag :type="getStatusTagType(row.status || row.Status)">
                {{ formatStatus({ row }) }}
              </el-tag>
              <el-progress
                v-if="isExportProcessing(row)"
                :percentage="getExportProgress(row)"
                :status="getExportProgressStatus(row)"
                :stroke-width="4"
                style="margin-top: 4px;"
              />
            </div>
          </template>
        </vxe-column>
        <vxe-column field="admin" :title="$t('log.admin')" width="140" :formatter="formatAdmin" />
        <vxe-column field="created_at" :title="$t('table.created_at')" width="180" sortable />
        <vxe-column :title="$t('table.operation')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button 
              v-if="isExportProcessing(row)"
              type="info" 
              link 
              size="small"
              @click="handleMonitorExport(row)"
            >
              {{ monitoringExports.has(row.id || row.ID) ? ($t('export.stop_monitor') || '停止监控') : ($t('export.monitor_progress') || '监控进度') }}
            </el-button>
            <el-button 
              type="primary" 
              link 
              :disabled="downloadingIds.has(row.id || row.ID) || !isExportCompleted(row)"
              :loading="downloadingIds.has(row.id || row.ID)"
              @click="handleDownload(row)"
            >
              {{ $t('common.view') }}
            </el-button>
            <el-button 
              type="danger" 
              link 
              :disabled="getButtonState('export.destroy').disabled"
              @click="handleDelete(row)"
            >
              {{ $t('common.delete') }}
            </el-button>
          </template>
        </vxe-column>
      </vxe-table>

      <Pagination
        v-model="pagination"
        :show-total="true"
        :show-quick-jumper="true"
        :align="'right'"
        @page-change="handlePageChange"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onActivated, onDeactivated, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { getExportList, deleteExport, batchDeleteExports, createExportProgressSSE } from '../../api/export'
import { createSSEConnection, closeSSEConnection } from '../../utils/sse'
import i18n from '../../i18n'
import Storage from '../../utils/storage'

const { t, locale } = useI18n()
const { getButtonState } = usePermission()

const tableRef = ref(null)
const selectedRows = ref([])
const downloadingIds = ref(new Set()) // 正在下载的文件 ID 集合
const monitoringExports = ref(new Map()) // 正在监控的导出任务 { exportId: eventSource }
const exportProgress = ref(new Map()) // 导出任务进度 { exportId: { status, message, file_url } }

// 数据转换函数
const transformExportData = (item) => {
  return {
    id: item.id || item.ID,
    Admin: item.Admin || item.admin || null,
    filename: item.Filename || item.filename || '',
    disk: item.Disk || item.disk || '',
    path: item.Path || item.path || '',
    extension: item.Extension || item.extension || '',
    size: item.Size || item.size || 0,
    status: item.Status || item.status || 0,
    created_at: item.CreatedAt || item.created_at || '',
    file_url: item.FileURL || item.file_url || ''
  }
}

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'filename': 'filename',
  'disk': 'disk',
  'path': 'path',
  'extension': 'extension',
  'size': 'size',
  'status': 'status',
  'created_at': 'created_at'
}

// 使用列表页面 composable
const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handlePageChange,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getExportList,
  initialSearchForm: {
    filename: '',
    disk: '',
    status: '',
    start_time: '',
    end_time: ''
  },
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  },
  transformData: transformExportData
})

const searchFields = computed(() => [
  {
    prop: 'filename',
    label: t('export.filename'),
    type: 'input',
    width: '200px'
  },
  {
    prop: 'disk',
    label: t('export.disk'),
    type: 'input',
    width: '180px'
  },
  {
    prop: 'status',
    label: t('log.status'),
    type: 'select',
    width: '150px',
    options: [
      { label: t('log.success'), value: '1' },
      { label: t('log.failed'), value: '0' }
    ],
    clearable: true
  },
  {
    prop: 'start_time',
    label: t('log.start_time'),
    type: 'datetime',
    width: '180px',
    valueFormat: 'YYYY-MM-DD HH:mm:ss',
    advanced: true
  },
  {
    prop: 'end_time',
    label: t('log.end_time'),
    type: 'datetime',
    width: '180px',
    valueFormat: 'YYYY-MM-DD HH:mm:ss',
    advanced: true
  }
])

const formatSize = ({ cellValue }) => {
  const size = Number(cellValue || 0)
  if (!size) return '-'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(2)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(2)} GB`
}

const formatAdmin = ({ row }) => {
  if (row.Admin && (row.Admin.Username || row.Admin.username)) {
    return row.Admin.Username || row.Admin.username
  }
  if (row.admin && row.admin.username) {
    return row.admin.username
  }
  return '-'
}

const formatStatus = ({ row }) => {
  const status = row.status || row.Status || 0
  const progress = exportProgress.value.get(row.id || row.ID)
  
  if (progress) {
    return progress.status || (status === 1 ? t('log.success') : t('log.failed'))
  }
  
  return status === 1 ? t('log.success') : status === 0 ? t('log.processing') || '处理中' : t('log.failed')
}

const getStatusTagType = (status) => {
  if (status === 1) return 'success'
  if (status === 0) return 'warning'
  return 'danger'
}

// 判断导出任务是否正在处理中
const isExportProcessing = (row) => {
  const status = row.status || row.Status || 0
  return status === 0 // 0 表示处理中
}

// 判断导出任务是否已完成
const isExportCompleted = (row) => {
  const status = row.status || row.Status || 0
  return status === 1 // 1 表示成功
}

// 获取导出进度百分比
const getExportProgress = (row) => {
  const exportId = row.id || row.ID
  const progress = exportProgress.value.get(exportId)
  if (progress && progress.progress !== undefined) {
    return progress.progress
  }
  return 0
}

// 获取导出进度状态
const getExportProgressStatus = (row) => {
  const exportId = row.id || row.ID
  const progress = exportProgress.value.get(exportId)
  if (progress && progress.status === '失败') {
    return 'exception'
  }
  return null
}

// 监控导出任务进度
const handleMonitorExport = (row) => {
  const exportId = row.id || row.ID
  if (!exportId) return
  
  // 如果已经在监控，先停止
  if (monitoringExports.value.has(exportId)) {
    stopMonitorExport(exportId)
    return
  }
  
  try {
    const url = createExportProgressSSE(exportId, { interval: 1000 })
    const eventSource = createSSEConnection(url, {
      onMessage: (data) => {
        if (data.type === 'progress') {
          exportProgress.value.set(exportId, {
            status: data.status_text || '处理中',
            message: data.message || '',
            progress: data.progress || 0
          })
        } else if (data.type === 'completed') {
          exportProgress.value.set(exportId, {
            status: '成功',
            message: '导出完成',
            file_url: data.file_url,
            filename: data.filename,
            progress: 100
          })
          // 停止监控并刷新列表
          stopMonitorExport(exportId)
          loadData()
          ElMessage.success(t('export.export_completed') || '导出完成')
        } else if (data.type === 'failed') {
          exportProgress.value.set(exportId, {
            status: '失败',
            message: data.message || '导出失败',
            progress: 0
          })
          stopMonitorExport(exportId)
          ElMessage.error(data.message || t('export.export_failed') || '导出失败')
        }
      },
      onError: (error) => {
        console.error('Export progress SSE error:', error)
        stopMonitorExport(exportId)
      }
    })
    
    monitoringExports.value.set(exportId, eventSource)
    ElMessage.info(t('export.monitoring_started') || '开始监控导出进度')
  } catch (error) {
    console.error('Failed to start export progress monitoring:', error)
    if (!error.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('export.monitor_failed') || '启动监控失败'
      ElMessage.error(errorMessage)
    }
  }
}

// 停止监控导出任务
const stopMonitorExport = (exportId) => {
  const eventSource = monitoringExports.value.get(exportId)
  if (eventSource) {
    closeSSEConnection(eventSource)
    monitoringExports.value.delete(exportId)
  }
  // 不清除进度信息，保留显示
}

// loadData, handleSearch, handleReset, handlePageChange 已由 useListPage 提供

const handleDownload = async (row) => {
  // 防止重复点击
  const exportId = row.id || row.ID
  if (downloadingIds.value.has(exportId)) {
    return // 如果正在下载，直接返回
  }
  
  const url = row.file_url || row.FileURL
  if (!url) {
    ElMessage.error(t('export.download_failed') || '无法构造下载链接')
    return
  }
  
  // 标记为正在下载
  downloadingIds.value.add(exportId)
  
  try {
    // 构建完整的下载 URL
    let fullUrl = url
    
    // 如果是相对路径，需要构建完整 URL
    if (url.startsWith('/')) {
      const apiBaseURL = import.meta.env.VITE_API_BASE_URL
      const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
      
      if (apiBaseURL) {
        // 如果配置了完整的基础 URL，使用它
        const base = apiBaseURL.replace(/\/+$/, '')
        // 移除 URL 中可能重复的 /api/admin 前缀
        const cleanUrl = url.replace(/^\/api\/admin/, '')
        fullUrl = `${base}${apiPrefix}${cleanUrl}`
      } else {
        // 如果没有配置基础 URL，直接使用相对路径
        fullUrl = url
      }
    }
    
    // 获取 token
    const token = Storage.getItem('token', '') || ''
    
    // 获取当前语言设置
    const currentLocale = locale.value || i18n.global.locale.value || Storage.getItem('language', 'zh-CN') || 'zh-CN'
    const acceptLanguage = currentLocale === 'en-US' ? 'en-US' : 'zh-CN'
    
    // 使用 fetch 请求下载文件，这样可以携带认证 token
    const response = await fetch(fullUrl, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token.trim()}`,
        'Accept-Language': acceptLanguage
      }
    })
    
    if (!response.ok) {
      if (response.status === 401) {
        ElMessage.error(t('error.unauthorized') || '未授权，请重新登录')
      } else {
        const errorText = await response.text()
        console.error('Download error response:', errorText)
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      return
    }
    
    // 检查响应类型，确保是文件而不是 HTML
    const contentType = response.headers.get('content-type') || ''
    if (contentType.includes('text/html')) {
      const htmlContent = await response.text()
      console.error('Received HTML instead of file:', htmlContent.substring(0, 200))
      ElMessage.error(t('export.download_failed') || '下载失败：服务器返回了错误内容')
      return
    }
    
    // 获取文件内容
    const blob = await response.blob()
    
    // 从响应头获取文件名，如果没有则使用记录中的文件名
    const contentDisposition = response.headers.get('content-disposition') || ''
    let filename = row.filename || row.Filename || 'export.csv'
    
    // 尝试从 Content-Disposition 头中提取文件名
    const filenameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/)
    if (filenameMatch && filenameMatch[1]) {
      filename = filenameMatch[1].replace(/['"]/g, '')
      // 处理 URL 编码的文件名
      try {
        filename = decodeURIComponent(filename)
      } catch (e) {
        // 如果解码失败，使用原始文件名
      }
    }
    
    // 创建下载链接
    const downloadUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = filename
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    // 释放 URL 对象
    window.URL.revokeObjectURL(downloadUrl)
    
    ElMessage.success(t('export.download_success') || '下载成功')
  } catch (error) {
    console.error('Download error:', error)
    if (!error.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('export.download_failed') || '下载失败'
      ElMessage.error(errorMessage)
    }
  } finally {
    // 下载完成或失败后，延迟移除标记（防止短时间内重复点击）
    setTimeout(() => {
      downloadingIds.value.delete(exportId)
    }, 2000) // 2秒内不允许重复点击
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(t('log.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteExport(row.id || row.ID)
    ElMessage.success(t('log.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete export error:', error)
    }
  }
}

const handleSelectionChange = () => {
  selectedRows.value = tableRef.value?.getCheckboxRecords() || []
}

const handleBatchDelete = async () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning(t('common.please_select_items'))
    return
  }

  try {
    await ElMessageBox.confirm(
      t('log.batch_delete_confirm', { count: selectedRows.value.length }), 
      t('form.tip'), 
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    const ids = selectedRows.value.map(row => row.id || row.ID)
    await batchDeleteExports(ids)
    ElMessage.success(t('log.delete_success'))
    selectedRows.value = []
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Batch delete error:', error)
    }
  }
}

onMounted(() => {
  initDefaultSort()
  loadData()
})

// 当组件被激活时（包括从缓存恢复）自动刷新数据
onActivated(() => {
  loadData()
})

// 组件被缓存时清理所有 SSE 连接（keep-alive场景）
onDeactivated(() => {
  monitoringExports.value.forEach((eventSource, exportId) => {
    closeSSEConnection(eventSource)
  })
  monitoringExports.value.clear()
  // 注意：不清除 exportProgress，保留进度信息显示
})

// 组件卸载时清理所有 SSE 连接
onUnmounted(() => {
  monitoringExports.value.forEach((eventSource, exportId) => {
    closeSSEConnection(eventSource)
  })
  monitoringExports.value.clear()
  exportProgress.value.clear()
})
</script>

<style scoped>
.export-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}
</style>


