<template>
  <div class="blacklist-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('blacklist.title') }}</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            {{ $t('blacklist.add_blacklist') }}
          </el-button>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ ip: '', status: '' }"
        i18n-prefix="blacklist"
        @search="handleSearch"
        @reset="handleReset"
      />

      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        :column-config="{ resizable: true }"
        height="600"
        :sort-config="{ multiple: true, trigger: 'default' }"
        @sort-change="handleSortChange"
      >
        <template v-for="column in tableColumns" :key="column.field || column.title || column.type">
          <vxe-column
            v-if="column.type === 'checkbox'"
            type="checkbox"
            :width="column.width"
            :fixed="column.fixed"
          />
          <vxe-column
            v-else
            :field="column.field"
            :title="column.title"
            :width="column.width"
            :sortable="column.sortable"
            :fixed="column.fixed"
            :formatter="column.formatter"
          >
            <template v-if="column.slot === 'status'" #default="{ row }">
              <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'danger' : 'info'">
                {{ (row.Status ?? row.status ?? 1) === 1 ? $t('blacklist.enabled') : $t('blacklist.disabled') }}
              </el-tag>
            </template>
            <template v-else-if="column.slot === 'ip'" #default="{ row }">
              <div style="word-break: break-all;">
                {{ formatIP(row.IP || row.ip || '') }}
              </div>
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <el-button type="primary" link @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
              <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
            </template>
          </vxe-column>
        </template>
      </vxe-table>

      <Pagination
        v-model="pagination"
        @page-change="handlePageChange"
      />
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="700px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item :label="$t('blacklist.ip')" prop="ip">
          <el-input
            v-model="formData.ip"
            type="textarea"
            :rows="4"
            :placeholder="$t('blacklist.ip_placeholder')"
          />
          <div style="margin-top: 8px; color: #909399; font-size: 12px;">
            {{ $t('blacklist.ip_tip') }}
          </div>
        </el-form-item>
        <el-form-item :label="$t('blacklist.remark')" prop="remark">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="3"
            :placeholder="$t('blacklist.remark_placeholder')"
          />
        </el-form-item>
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">{{ $t('blacklist.enabled') }}</el-radio>
            <el-radio :label="0">{{ $t('blacklist.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { useTableSort } from '../../composables/useTableSort'
import {
  getBlacklistList,
  getBlacklistDetail,
  createBlacklist,
  updateBlacklist,
  deleteBlacklist
} from '../../api/blacklist'

const { t } = useI18n()
const formRef = ref(null)
const tableRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('blacklist.edit_blacklist') : t('blacklist.add_blacklist'))

const searchForm = reactive({
  ip: '',
  status: ''
})

// 表格列配置
const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },
  {
    field: 'ip',
    title: t('blacklist.ip'),
    sortable: true,
    slot: 'ip'
  },
  {
    field: 'remark',
    title: t('blacklist.remark'),
    sortable: false,
    formatter: ({ row }) => row.Remark || row.remark || '-'
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 100,
    sortable: true,
    slot: 'status'
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    width: 180,
    sortable: true
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation',
    sortable: false
  }
])

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'ip',
    label: t('blacklist.ip'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'select',
    width: '150px',
    advanced: false,
    options: [
      { label: t('blacklist.enabled'), value: '1' },
      { label: t('blacklist.disabled'), value: '0' }
    ]
  }
])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])

const formData = reactive({
  id: null,
  ip: '',
  remark: '',
  status: 1
})

const formRules = computed(() => ({
  ip: [
    { required: true, message: t('blacklist.ip_required'), trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value || value.trim() === '') {
          callback(new Error(t('blacklist.ip_required')))
          return
        }
        // 前端简单验证，后端会做详细验证
        const ipList = value.split(',')
        for (const ip of ipList) {
          const trimmedIP = ip.trim()
          if (trimmedIP === '') continue
          // 简单检查：至少包含点或斜杠或横线
          if (!trimmedIP.includes('.') && !trimmedIP.includes('/') && !trimmedIP.includes('-')) {
            callback(new Error(t('blacklist.ip_format_error')))
            return
          }
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}))

// 格式化IP显示
const formatIP = (ip) => {
  if (!ip) return '-'
  // 如果是IP范围，格式化显示
  if (ip.includes('-')) {
    const parts = ip.split('-')
    if (parts.length === 2) {
      return `${parts[0].trim()} ~ ${parts[1].trim()}`
    }
  }
  return ip
}

// 字段名映射
const fieldMapping = {
  'id': 'id',
  'ip': 'ip',
  'remark': 'remark',
  'status': 'status',
  'created_at': 'created_at'
}

// 使用排序 composable
const { buildOrderBy, handleSortChange, resetSort, initDefaultSort } = useTableSort({
  tableRef,
  fieldMapping,
  defaultSort: 'id:desc',
  onSortChange: () => {
    pagination.page = 1
    loadData()
  }
})

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      order_by: buildOrderBy()
    }
    if (searchForm.ip && searchForm.ip.trim()) {
      params.ip = searchForm.ip.trim()
    }
    if (searchForm.status) {
      params.status = searchForm.status
    }
    
    const res = await getBlacklistList(params)
    
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load blacklist list error:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.ip = ''
  searchForm.status = ''
  resetSort()
  handleSearch()
}

const handlePageChange = ({ currentPage, pageSize }) => {
  pagination.page = currentPage
  pagination.pageSize = pageSize
  loadData()
}

const handleAdd = () => {
  Object.assign(formData, {
    id: null,
    ip: '',
    remark: '',
    status: 1
  })
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  try {
    const res = await getBlacklistDetail(row.id)
    if (res.data && res.data.blacklist) {
      const blacklist = res.data.blacklist
      Object.assign(formData, {
        id: blacklist.id,
        ip: blacklist.IP || blacklist.ip || '',
        remark: blacklist.Remark || blacklist.remark || '',
        status: blacklist.Status !== undefined ? blacklist.Status : (blacklist.status !== undefined ? blacklist.status : 1)
      })
      dialogVisible.value = true
    }
  } catch (error) {
    console.error('Load blacklist detail error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const submitData = {
          ip: formData.ip.trim(),
          remark: formData.remark.trim(),
          status: formData.status
        }
        if (formData.id) {
          await updateBlacklist(formData.id, submitData)
          ElMessage.success(t('blacklist.update_success'))
        } else {
          await createBlacklist(submitData)
          ElMessage.success(t('blacklist.create_success'))
        }
        dialogVisible.value = false
        loadData()
      } catch (error) {
        console.error('Submit error:', error)
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(t('blacklist.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteBlacklist(row.id)
    ElMessage.success(t('blacklist.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

onMounted(() => {
  initDefaultSort()
  loadData()
})
</script>

<style scoped>
.blacklist-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

