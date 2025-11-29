<template>
  <div class="export-list">
    <el-card class="box-card">
      <div class="card-header">
        <span>{{ $t('export.title') }}</span>
      </div>

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
      >
        <vxe-column type="checkbox" width="60" />
        <vxe-column field="id" :title="$t('table.id')" width="80" sortable />
        <vxe-column field="filename" :title="$t('export.filename')" min-width="200" />
        <vxe-column field="disk" :title="$t('export.disk')" width="120" />
        <vxe-column field="path" :title="$t('export.path')" min-width="260" />
        <vxe-column field="extension" :title="$t('export.extension')" width="100" />
        <vxe-column field="size" :title="$t('export.size')" width="140" :formatter="formatSize" />
        <vxe-column field="admin" :title="$t('log.admin')" width="140" :formatter="formatAdmin" />
        <vxe-column field="created_at" :title="$t('table.created_at')" width="180" sortable />
        <vxe-column :title="$t('table.operation')" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleDownload(row)">{{ $t('common.view') }}</el-button>
            <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
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
import { ref, reactive, computed, onMounted, onActivated } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { getExportList, deleteExport } from '../../api/export'

const { t } = useI18n()

const tableRef = ref(null)
const loading = ref(false)
const tableData = ref([])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const searchForm = reactive({
  filename: '',
  disk: '',
  status: '',
  start_time: '',
  end_time: ''
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

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...searchForm
    }
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })
    const res = await getExportList(params)
    if (res.data) {
      const list = res.data.list || res.data.data || []
      // 兼容后端返回的 PascalCase / snake_case 字段，并加入 file_url
      tableData.value = list.map(item => ({
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
      }))
      pagination.total = res.data.total || res.data.meta?.total || 0
    }
  } catch (error) {
    console.error('Load export list error:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.filename = ''
  searchForm.disk = ''
  searchForm.status = ''
  searchForm.start_time = ''
  searchForm.end_time = ''
  pagination.page = 1
  loadData()
}

const handlePageChange = ({ currentPage, pageSize }) => {
  pagination.page = currentPage
  pagination.pageSize = pageSize
  loadData()
}

const handleDownload = async (row) => {
  const url = row.file_url || row.FileURL
  if (!url) {
    ElMessage.error(t('export.download_failed') || '无法构造下载链接')
    return
  }
  
  try {
    // 获取完整的 URL（如果是相对路径，加上 baseURL）
    let fullUrl = url
    if (url.startsWith('/')) {
      const baseURL = import.meta.env.VITE_API_PREFIX || '/api/admin'
      fullUrl = baseURL + url.replace('/api/admin', '')
    }
    
    // 获取 token
    const token = localStorage.getItem('token') || ''
    
    // 使用 fetch 请求下载文件，这样可以携带认证 token
    const response = await fetch(fullUrl, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token.trim()}`
      }
    })
    
    if (!response.ok) {
      if (response.status === 401) {
        ElMessage.error(t('error.unauthorized') || '未授权，请重新登录')
      } else {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      return
    }
    
    // 获取文件内容
    const blob = await response.blob()
    
    // 创建下载链接
    const downloadUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = row.filename || row.Filename || ''
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    // 释放 URL 对象
    window.URL.revokeObjectURL(downloadUrl)
  } catch (error) {
    console.error('Download error:', error)
    ElMessage.error(t('export.download_failed') || '下载失败')
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

onMounted(() => {
  loadData()
})

// 当组件被激活时（包括从缓存恢复）自动刷新数据
onActivated(() => {
  loadData()
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
  margin-bottom: 10px;
}
</style>


