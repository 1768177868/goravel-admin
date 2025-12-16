<template>
  <div class="attachment-list">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('attachment.title') }}</span>
          <div class="header-actions">
            <el-upload
              ref="uploadRef"
              :action="uploadAction"
              :headers="uploadHeaders"
              :data="uploadData"
              :before-upload="beforeUpload"
              :on-success="handleUploadSuccess"
              :on-error="handleUploadError"
              :on-progress="handleUploadProgress"
              :show-file-list="false"
              :multiple="false"
            >
              <template #trigger>
                <el-button 
                  type="primary"
                  :disabled="getButtonState('attachment.store').disabled"
                >
                  <el-icon><UploadIcon /></el-icon>
                  {{ $t('attachment.upload') }}
                </el-button>
              </template>
            </el-upload>
            <el-button 
              type="success"
              :disabled="getButtonState('attachment.store').disabled"
              @click="handleLargeFileUpload"
            >
              <el-icon><UploadIcon /></el-icon>
              {{ $t('attachment.large_file_upload') }}
            </el-button>
            <el-button 
              type="danger" 
              :disabled="selectedRows.length === 0 || getButtonState('attachment.destroy').disabled"
              @click="handleBatchDelete"
            >
              <el-icon><DeleteIcon /></el-icon>
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
        <vxe-column field="filename" :title="$t('attachment.filename')" min-width="200">
          <template #default="{ row }">
            <div class="filename-cell">
              <el-image
                v-if="row.file_type === 'image'"
                :src="getImageUrl(row)"
                :preview-src-list="[getImageUrl(row)]"
                fit="cover"
                class="filename-thumbnail"
                :preview-teleported="true"
              />
              <span class="filename-text">{{ row.filename || row.Filename }}</span>
            </div>
          </template>
        </vxe-column>
        <vxe-column field="display_name" :title="$t('attachment.display_name')" min-width="200">
          <template #default="{ row }">
            <el-input
              v-model="row.display_name"
              :placeholder="$t('attachment.display_name_placeholder')"
              size="small"
              @blur="handleUpdateDisplayName(row)"
              @keyup.enter="handleUpdateDisplayName(row)"
            />
          </template>
        </vxe-column>
        <vxe-column field="file_type" :title="$t('attachment.file_type')" width="120">
          <template #default="{ row }">
            <el-tag :type="getFileTypeTagType(row.file_type)">
              {{ getFileTypeLabel(row.file_type) }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="disk" :title="$t('attachment.disk')" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">
              {{ row.disk || row.Disk || '-' }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="extension" :title="$t('attachment.extension')" width="100" />
        <vxe-column field="size" :title="$t('attachment.size')" width="140" :formatter="formatSize" />
        <vxe-column field="mime_type" :title="$t('attachment.mime_type')" min-width="150" />
        <vxe-column field="admin" :title="$t('log.admin')" width="140" :formatter="formatAdmin" />
        <vxe-column field="created_at" :title="$t('table.created_at')" width="180" sortable />
        <vxe-column :title="$t('table.operation')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button 
              type="success" 
              link 
              :disabled="downloadingIds.has(row.id || row.ID)"
              :loading="downloadingIds.has(row.id || row.ID)"
              @click="handleDownload(row)"
            >
              {{ $t('common.download') }}
            </el-button>
            <el-button 
              type="danger" 
              link 
              :disabled="getButtonState('attachment.destroy').disabled"
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

    <!-- 大文件上传对话框 -->
    <el-dialog
      v-model="chunkUploadVisible"
      :title="$t('attachment.chunk_upload')"
      width="600px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <div v-if="chunkUploadFile" class="chunk-upload-container">
        <div class="upload-info">
          <p><strong>{{ $t('attachment.filename') }}:</strong> {{ chunkUploadFile.name }}</p>
          <p><strong>{{ $t('attachment.size') }}:</strong> {{ formatFileSize(chunkUploadFile.size) }}</p>
        </div>
        <el-progress 
          :percentage="Math.round(chunkUploadProgress)" 
          :status="chunkUploadStatus"
          :stroke-width="20"
        />
        <div class="upload-status">
          <p v-if="chunkUploadStatus === 'success'">{{ $t('attachment.upload_success') }}</p>
          <p v-else-if="chunkUploadStatus === 'exception'">{{ $t('attachment.upload_failed') }}</p>
          <p v-else>{{ $t('attachment.uploading') }}: {{ Math.round(chunkUploadProgress) }}%</p>
        </div>
        <div class="upload-actions" style="margin-top: 20px; text-align: right;">
          <el-button 
            v-if="chunkUploadStatus !== 'success' && chunkUploadStatus !== 'exception'"
            @click="handleCancelChunkUpload"
          >
            {{ $t('common.cancel') }}
          </el-button>
          <el-button 
            v-if="chunkUploadStatus === 'success'"
            type="primary"
            @click="handleChunkUploadClose"
          >
            {{ $t('common.confirm') }}
          </el-button>
          <el-button 
            v-if="chunkUploadStatus === 'exception'"
            type="primary"
            @click="handleRetryChunkUpload"
          >
            {{ $t('common.retry') }}
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onActivated, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, Delete } from '@element-plus/icons-vue'

// 使用 markRaw 标记图标组件，避免被 Vue 做成响应式对象
const UploadIcon = markRaw(Upload)
const DeleteIcon = markRaw(Delete)
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import axios from 'axios'
import { 
  getAttachmentList, 
  uploadFile, 
  deleteAttachment, 
  batchDeleteAttachments,
  initChunkUpload,
  uploadChunk,
  mergeChunks,
  getChunkProgress,
  updateDisplayName
} from '../../api/attachment'
import i18n from '../../i18n'
import Storage from '../../utils/storage'

const { t, locale } = useI18n()
const { getButtonState } = usePermission()

const tableRef = ref(null)
const uploadRef = ref(null)
const selectedRows = ref([])
const downloadingIds = ref(new Set()) // 正在下载的文件 ID 集合
const chunkUploadVisible = ref(false)
const chunkUploadFile = ref(null)
const chunkUploadProgress = ref(0)
const chunkUploadStatus = ref('')
const chunkUploadChunkID = ref('')
const chunkUploadChunks = ref([])
const chunkUploadCancelled = ref(false) // 标记是否已取消上传
// 图片URL缓存（key: attachment_id, value: blob_url 或 直接URL）
const imageUrlMap = ref(new Map())

// 大文件阈值（5MB），超过此大小使用分片上传
const CHUNK_SIZE = 2 * 1024 * 1024 // 2MB per chunk
const LARGE_FILE_THRESHOLD = 5 * 1024 * 1024 // 5MB

// 数据转换函数
const transformAttachmentData = (item) => {
  return {
    id: item.id || item.ID,
    Admin: item.Admin || item.admin || null,
    filename: item.Filename || item.filename || '',
    display_name: item.DisplayName || item.display_name || '',
    file_type: item.FileType || item.file_type || 'other',
    disk: item.Disk || item.disk || '',
    extension: item.Extension || item.extension || '',
    size: item.Size || item.size || 0,
    mime_type: item.MimeType || item.mime_type || '',
    created_at: item.CreatedAt || item.created_at || '',
    file_url: item.FileURL || item.file_url || ''
  }
}

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'filename': 'filename',
  'display_name': 'display_name',
  'file_type': 'file_type',
  'disk': 'disk',
  'extension': 'extension',
  'size': 'size',
  'mime_type': 'mime_type',
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
  fetchApi: getAttachmentList,
  initialSearchForm: {
    filename: '',
    display_name: '',
    file_type: '',
    extension: '',
    start_time: '',
    end_time: ''
  },
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  },
  transformData: transformAttachmentData,
  onLoadSuccess: (res, list) => {
    // 加载所有图片的blob URL
    list.forEach(row => {
      if (row.file_type === 'image') {
        loadImageAsBlob(row)
      }
    })
  }
})

const searchFields = computed(() => [
  {
    prop: 'filename',
    label: t('attachment.filename'),
    type: 'input',
    width: '200px'
  },
  {
    prop: 'display_name',
    label: t('attachment.display_name'),
    type: 'input',
    width: '200px'
  },
  {
    prop: 'file_type',
    label: t('attachment.file_type'),
    type: 'select',
    width: '150px',
    options: [
      { label: t('attachment.file_type_image'), value: 'image' },
      { label: t('attachment.file_type_video'), value: 'video' },
      { label: t('attachment.file_type_document'), value: 'document' },
      { label: t('attachment.file_type_other'), value: 'other' }
    ],
    clearable: true
  },
  {
    prop: 'extension',
    label: t('attachment.extension'),
    type: 'input',
    width: '150px'
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

const uploadAction = computed(() => {
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL
  const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
  if (apiBaseURL) {
    const base = apiBaseURL.replace(/\/+$/, '')
    const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
    return `${base}${prefix}/attachments/upload`
  }
  return `${apiPrefix}/attachments/upload`
})

const uploadHeaders = computed(() => {
  const token = Storage.getItem('token', '') || ''
  return {
    'Authorization': `Bearer ${typeof token === 'string' ? token.trim() : ''}`
  }
})

const uploadData = computed(() => {
  const currentLocale = locale.value || Storage.getItem('language', 'zh-CN') || 'zh-CN'
  const acceptLanguage = currentLocale === 'en-US' ? 'en-US' : 'zh-CN'
  return {
    'Accept-Language': acceptLanguage
  }
})

// 加载图片并转换为blob URL
const loadImageAsBlob = async (row) => {
  if (!row || row.file_type !== 'image') return
  
  const attachmentId = row.id || row.ID
  if (!attachmentId) return
  
  // 如果已经加载过，直接返回
  if (imageUrlMap.value.has(attachmentId)) {
    return
  }
  
  const fileUrl = row.file_url || row.FileURL
  if (!fileUrl) return
  
  // 如果是外部URL（http/https），直接使用
  // 对于云存储（S3、OSS、COS、MinIO等），GetFileURL 会返回：
  // 1. 临时URL（带签名，可直接访问，如：https://bucket.s3.amazonaws.com/path?signature=...）
  // 2. 配置的基础URL + 文件路径（如：https://cdn.example.com/path/file.jpg）
  // 这些URL都可以直接在浏览器中访问，不需要JWT认证
  if (fileUrl.startsWith('http')) {
    imageUrlMap.value.set(attachmentId, fileUrl)
    return
  }
  
  // 构建完整的URL
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL
  const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
  let fullUrl = fileUrl
  if (apiBaseURL) {
    const base = apiBaseURL.replace(/\/+$/, '')
    fullUrl = `${base}${fileUrl}`
  } else {
    fullUrl = `${apiPrefix}${fileUrl.startsWith('/') ? '' : '/'}${fileUrl}`
  }
  
  // 通过axios获取图片并转换为blob URL
  // 因为预览接口需要JWT认证，直接使用src无法携带认证头
  try {
    const token = Storage.getItem('token', '') || ''
    const tokenStr = typeof token === 'string' ? token.trim() : ''
    const response = await axios.get(fullUrl, {
      responseType: 'blob',
      headers: {
        'Authorization': `Bearer ${tokenStr}`
      }
    })
    const blob = new Blob([response.data])
    const blobUrl = URL.createObjectURL(blob)
    imageUrlMap.value.set(attachmentId, blobUrl)
  } catch (error) {
    console.error('Failed to load image:', error)
    // 加载失败时设置为空，避免重复请求
    imageUrlMap.value.set(attachmentId, '')
  }
}

// 获取图片URL（用于缩略图和预览）
const getImageUrl = (row) => {
  if (!row) return ''
  const attachmentId = row.id || row.ID
  if (!attachmentId) return ''
  
  // 从缓存中获取
  return imageUrlMap.value.get(attachmentId) || ''
}


const formatSize = ({ cellValue }) => {
  return formatFileSize(cellValue)
}

const formatFileSize = (size) => {
  if (!size) return '-'
  const numSize = Number(size)
  if (numSize < 1024) return `${numSize} B`
  if (numSize < 1024 * 1024) return `${(numSize / 1024).toFixed(2)} KB`
  if (numSize < 1024 * 1024 * 1024) return `${(numSize / 1024 / 1024).toFixed(2)} MB`
  return `${(numSize / 1024 / 1024 / 1024).toFixed(2)} GB`
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

const getFileTypeTagType = (fileType) => {
  const types = {
    'image': 'success',
    'video': 'warning',
    'document': 'info',
    'other': null
  }
  const result = types[fileType]
  // 返回 null 而不是 undefined，这样 ElTag 的 type 属性会是 undefined（不传），而不是空字符串
  return result !== undefined ? result : null
}

const getFileTypeLabel = (fileType) => {
  const labels = {
    'image': t('attachment.file_type_image'),
    'video': t('attachment.file_type_video'),
    'document': t('attachment.file_type_document'),
    'other': t('attachment.file_type_other')
  }
  return labels[fileType] || fileType
}

// loadData, handleSearch, handleReset, handlePageChange 已由 useListPage 提供

// 大文件上传按钮处理
const handleLargeFileUpload = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.onchange = (e) => {
    const file = e.target.files[0]
    if (file) {
      handleChunkUpload(file, true) // 第二个参数表示是大文件上传按钮触发的
    }
  }
  input.click()
}

const beforeUpload = (file) => {
  // 检查文件大小（限制100MB）
  const maxSize = 100 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.error(t('attachment.file_too_large'))
    return false
  }

  // 如果是大文件，使用分片上传
  if (file.size > LARGE_FILE_THRESHOLD) {
    handleChunkUpload(file, false) // 普通上传按钮触发的
    return false // 阻止默认上传
  }

  return true // 小文件使用普通上传
}

const handleUploadProgress = (event, file) => {
  // 普通上传的进度处理
}

const handleUploadSuccess = (response, file) => {
  ElMessage.success(t('attachment.upload_success'))
  loadData()
}

const handleUploadError = (error, file) => {
  ElMessage.error(t('attachment.upload_failed'))
}

// 分片上传处理（支持断点续传）
const handleChunkUpload = async (file, isLargeFileButton = false, useExistingChunkID = false) => {
  chunkUploadFile.value = file
  chunkUploadVisible.value = true
  chunkUploadProgress.value = 0
  chunkUploadStatus.value = ''
  chunkUploadCancelled.value = false // 重置取消标志

  try {
    // 计算分片信息
    const totalSize = file.size
    if (!totalSize || totalSize <= 0) {
      ElMessage.error(t('attachment.invalid_file_size'))
      chunkUploadVisible.value = false
      chunkUploadFile.value = null
      return
    }
    
    const totalChunks = Math.ceil(totalSize / CHUNK_SIZE)
    if (!totalChunks || totalChunks <= 0 || !isFinite(totalChunks)) {
      console.error('Invalid totalChunks calculated:', { totalSize, CHUNK_SIZE, totalChunks })
      ElMessage.error(t('attachment.invalid_chunk_calculation'))
      chunkUploadVisible.value = false
      chunkUploadFile.value = null
      return
    }

    // 如果使用已存在的 chunk_id（重试场景），跳过初始化
    if (!useExistingChunkID || !chunkUploadChunkID.value) {
      // 初始化分片上传
      try {
        const initRes = await initChunkUpload(
          file.name,
          totalSize,
          CHUNK_SIZE,
          totalChunks
        )
        chunkUploadChunkID.value = initRes.data.chunk_id
        
        // 保存分片信息到 localStorage（用于断点续传）
        try {
          Storage.setItem(`chunk_${chunkUploadChunkID.value}`, {
            filename: file.name,
            total_size: totalSize,
            chunk_size: CHUNK_SIZE,
            total_chunks: totalChunks,
            created_at: Date.now()
          })
        } catch (e) {
          console.warn('Failed to save chunk info to localStorage:', e)
        }
      } catch (error) {
        console.error('Init chunk upload error:', error)
        // 如果错误已经在响应拦截器中处理过，就不再重复显示
        if (!error.__handled) {
          // 检查错误码来决定是否需要关闭对话框
          const errorCode = error.errorCode || error.response?.data?.error_code || ''
          const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
          
          // 如果是存储驱动不支持的错误，需要关闭对话框
          if (errorCode === 'chunk_upload_only_local_storage') {
            chunkUploadVisible.value = false
            chunkUploadFile.value = null
          }
          
          ElMessage.error(errorMessage)
        } else {
          // 即使错误已处理，也可能需要关闭对话框
          const errorCode = error.errorCode || error.response?.data?.error_code || ''
          if (errorCode === 'chunk_upload_only_local_storage') {
            chunkUploadVisible.value = false
            chunkUploadFile.value = null
          }
        }
        throw error // 重新抛出其他错误
      }
    }

    // 检查已上传的分片（断点续传）
    let uploadedChunksSet = new Set()
    // 只有在 chunkID 存在且未取消时才获取进度
    if (chunkUploadChunkID.value && !chunkUploadCancelled.value) {
      try {
        const progressRes = await getChunkProgress(chunkUploadChunkID.value, totalChunks)
        if (progressRes.data && progressRes.data.uploaded_chunks) {
          // 后端返回已上传的分片索引数组
          const uploadedIndices = progressRes.data.uploaded_chunks || []
          uploadedChunksSet = new Set(uploadedIndices)
          if (uploadedChunksSet.size > 0) {
            ElMessage.info(t('attachment.resume_upload', { count: uploadedChunksSet.size, total: totalChunks }))
          }
        }
      } catch (error) {
        // 如果获取进度失败，继续正常上传（但不显示错误，因为可能是已取消）
        if (!chunkUploadCancelled.value) {
          console.warn('Failed to get chunk progress, starting fresh upload:', error)
        }
      }
    }

    // 不再使用SSE，改用基于分片上传进度的方式

    // 准备所有分片
    const chunks = []
    for (let i = 0; i < totalChunks; i++) {
      const start = i * CHUNK_SIZE
      const end = Math.min(start + CHUNK_SIZE, totalSize)
      const chunk = file.slice(start, end)
      chunks.push({ index: i, chunk, uploaded: uploadedChunksSet.has(i) })
    }

    chunkUploadChunks.value = chunks

    // 过滤出未上传的分片
    const pendingChunks = chunks.filter(chunk => !chunk.uploaded)
    const alreadyUploadedCount = chunks.length - pendingChunks.length

    // 更新初始进度
    if (alreadyUploadedCount > 0) {
      chunkUploadProgress.value = Math.round((alreadyUploadedCount / totalChunks) * 100)
    }

    // 并发上传分片（限制并发数为3）
    const concurrency = 3
    let uploadedCount = alreadyUploadedCount
    // 记录每个分片的上传进度（0-1），用于计算总进度
    const chunkProgressMap = new Map()
    // 初始化所有分片的进度
    for (let i = 0; i < totalChunks; i++) {
      if (uploadedChunksSet.has(i)) {
        chunkProgressMap.set(i, 1) // 已上传的分片进度为1
      } else {
        chunkProgressMap.set(i, 0) // 未上传的分片进度为0
      }
    }

    // 更新总进度的函数
    const updateTotalProgress = () => {
      if (chunkUploadCancelled.value) return
      let totalProgress = 0
      for (let i = 0; i < totalChunks; i++) {
        totalProgress += chunkProgressMap.get(i) || 0
      }
      const percent = Math.min(Math.round((totalProgress / totalChunks) * 100), 99) // 最多显示99%，等合并完成再显示100%
      chunkUploadProgress.value = percent
    }

    const uploadChunkWithProgress = async (chunkData) => {
      // 如果已取消，停止上传
      if (chunkUploadCancelled.value) {
        return
      }
      
      try {
        // 如果已上传，跳过
        if (chunkData.uploaded) {
          return
        }

        await uploadChunk(
          chunkUploadChunkID.value,
          chunkData.index,
          chunkData.chunk,
          (progress) => {
            // 单个分片的上传进度（0-100），转换为0-1
            if (!chunkUploadCancelled.value) {
              chunkProgressMap.set(chunkData.index, progress / 100)
              updateTotalProgress()
            }
          }
        )
        
        // 再次检查是否已取消
        if (chunkUploadCancelled.value) {
          return
        }
        
        uploadedCount++
        chunkProgressMap.set(chunkData.index, 1) // 标记该分片已完成
        // 更新总进度
        updateTotalProgress()
      } catch (error) {
        // 如果已取消，不抛出错误
        if (!chunkUploadCancelled.value) {
          throw error
        }
      }
    }

    // 分批上传未完成的分片
    for (let i = 0; i < pendingChunks.length; i += concurrency) {
      const batch = pendingChunks.slice(i, i + concurrency)
      await Promise.all(batch.map(uploadChunkWithProgress))
    }

    // 检查是否已取消
    if (chunkUploadCancelled.value) {
      return
    }

    // 所有分片上传完成，合并
    const mimeType = file.type || 'application/octet-stream'
    const mergeRes = await mergeChunks(
      chunkUploadChunkID.value,
      file.name,
      mimeType,
      totalChunks
    )

    // 再次检查是否已取消
    if (chunkUploadCancelled.value) {
      return
    }

    // 上传完成
    chunkUploadStatus.value = 'success'
    chunkUploadProgress.value = 100
    ElMessage.success(t('attachment.upload_success'))
    
    // 清理 localStorage 中的分片信息
    try {
      if (chunkUploadChunkID.value) {
        Storage.removeItem(`chunk_${chunkUploadChunkID.value}`)
      }
    } catch (e) {
      console.warn('Failed to remove chunk info from storage:', e)
    }
    
    // 刷新列表
    loadData()
  } catch (error) {
    if (chunkUploadCancelled.value) {
      return
    }
    console.error('Chunk upload error:', error)
    chunkUploadStatus.value = 'exception'
    if (!error.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('attachment.upload_failed')
      ElMessage.error(errorMessage)
    }
  }
}

const handleCancelChunkUpload = () => {
  // 标记为已取消，停止所有上传操作
  chunkUploadCancelled.value = true
  chunkUploadVisible.value = false
  chunkUploadFile.value = null
  chunkUploadChunkID.value = ''
  chunkUploadChunks.value = []
}

const handleChunkUploadClose = () => {
  handleCancelChunkUpload()
  loadData()
}

const handleRetryChunkUpload = () => {
  if (chunkUploadFile.value && chunkUploadChunkID.value) {
    // 重试时使用已存在的 chunk_id，实现断点续传
    handleChunkUpload(chunkUploadFile.value, false, true)
  } else if (chunkUploadFile.value) {
    // 如果没有 chunk_id，重新开始上传
    handleChunkUpload(chunkUploadFile.value)
  }
}

const handleUpdateDisplayName = async (row) => {
  const attachmentId = row.id || row.ID
  if (!attachmentId) return
  
  const displayName = row.display_name || ''
  
  try {
    await updateDisplayName(attachmentId, displayName)
    ElMessage.success(t('attachment.update_success'))
  } catch (error) {
    console.error('Update display name error:', error)
    if (!error.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('attachment.update_failed')
      ElMessage.error(errorMessage)
    }
    // 重新加载数据以恢复原值
    loadData()
  }
}

const handleDownload = async (row) => {
  // 防止重复点击
  const attachmentId = row.id || row.ID
  if (downloadingIds.value.has(attachmentId)) {
    return // 如果正在下载，直接返回
  }
  
  // 标记为正在下载
  downloadingIds.value.add(attachmentId)
  
  try {
    // 构建下载 URL
    const apiBaseURL = import.meta.env.VITE_API_BASE_URL
    const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
    
    let downloadUrl = `${apiPrefix}/attachments/${attachmentId}/download`
    
    if (apiBaseURL) {
      const base = apiBaseURL.replace(/\/+$/, '')
      downloadUrl = `${base}${downloadUrl}`
    }
    
    // 获取 token
    const token = Storage.getItem('token', '') || ''
    const tokenStr = typeof token === 'string' ? token.trim() : ''
    
    // 获取当前语言设置
    const currentLocale = locale.value || i18n.global.locale.value || Storage.getItem('language', 'zh-CN') || 'zh-CN'
    const acceptLanguage = currentLocale === 'en-US' ? 'en-US' : 'zh-CN'
    
    // 使用 fetch 请求下载文件，这样可以携带认证 token
    const response = await fetch(downloadUrl, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${tokenStr}`,
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
      ElMessage.error(t('attachment.download_failed') || '下载失败：服务器返回了错误内容')
      return
    }
    
    // 获取文件内容
    const blob = await response.blob()
    
    // 从响应头获取文件名，如果没有则使用记录中的文件名
    const contentDisposition = response.headers.get('content-disposition') || ''
    let filename = row.filename || row.Filename || 'attachment'
    
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
    const downloadUrlObj = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrlObj
    link.download = filename
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    // 释放 URL 对象
    window.URL.revokeObjectURL(downloadUrlObj)
    
    ElMessage.success(t('attachment.download_success') || '下载成功')
  } catch (error) {
    console.error('Download error:', error)
    if (!error.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('attachment.download_failed') || '下载失败'
      ElMessage.error(errorMessage)
    }
  } finally {
    // 下载完成或失败后，延迟移除标记（防止短时间内重复点击）
    setTimeout(() => {
      downloadingIds.value.delete(attachmentId)
    }, 2000) // 2秒内不允许重复点击
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(t('attachment.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteAttachment(row.id || row.ID)
    ElMessage.success(t('attachment.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete attachment error:', error)
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
      t('attachment.batch_delete_confirm', { count: selectedRows.value.length }), 
      t('form.tip'), 
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    const ids = selectedRows.value.map(row => row.id || row.ID)
    await batchDeleteAttachments(ids)
    ElMessage.success(t('attachment.delete_success'))
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

onActivated(() => {
  loadData()
})
</script>

<style scoped>
.attachment-list {
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

.preview-container {
  text-align: center;
}

.preview-image img {
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.preview-video video {
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.preview-document iframe {
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.chunk-upload-container {
  padding: 20px;
}

.upload-info {
  margin-bottom: 20px;
}

.upload-info p {
  margin: 8px 0;
}

.upload-status {
  margin-top: 10px;
  text-align: center;
  color: #666;
}

.upload-actions {
  margin-top: 20px;
}

.filename-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filename-thumbnail {
  width: 50px;
  height: 50px;
  cursor: pointer;
  border-radius: 4px;
  flex-shrink: 0;
  border: 1px solid #e4e7ed;
}

.filename-text {
  flex: 1;
  word-break: break-all;
}
</style>

