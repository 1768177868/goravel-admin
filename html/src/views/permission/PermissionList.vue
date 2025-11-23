<template>
  <div class="permission-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('permission.title') }}</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            {{ $t('permission.add_permission') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <el-form :model="searchForm" :inline="true" class="search-form">
        <el-form-item :label="$t('permission.name')">
          <el-input
            v-model="searchForm.name"
            :placeholder="$t('form.please_enter') + $t('permission.name')"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item :label="$t('permission.slug')">
          <el-input
            v-model="searchForm.slug"
            :placeholder="$t('form.please_enter') + $t('permission.slug')"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item :label="$t('permission.method')">
          <el-select
            v-model="searchForm.method"
            :placeholder="$t('form.please_select') + $t('permission.method')"
            clearable
            style="width: 150px"
          >
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('permission.path')">
          <el-input
            v-model="searchForm.path"
            :placeholder="$t('form.please_enter') + $t('permission.path')"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item :label="$t('table.status')">
          <el-select
            v-model="searchForm.status"
            :placeholder="$t('form.please_select') + $t('table.status')"
            clearable
            style="width: 120px"
          >
            <el-option :label="$t('common.enabled')" value="1" />
            <el-option :label="$t('common.disabled')" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            {{ $t('log.search') }}
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            {{ $t('log.reset') }}
          </el-button>
        </el-form-item>
      </el-form>

      <vxe-table
        :data="tableData"
        :loading="loading"
        border
        resizable
        height="600"
      >
        <vxe-column type="seq" width="60" :title="$t('table.seq')" />
        <vxe-column field="id" :title="$t('table.id')" width="80" />
        <vxe-column field="name" :title="$t('permission.name')">
          <template #default="{ row }">
            {{ row.Name || row.name || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="slug" :title="$t('permission.slug')">
          <template #default="{ row }">
            {{ row.Slug || row.slug || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="method" :title="$t('permission.method')" width="100">
          <template #default="{ row }">
            {{ row.Method || row.method || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="path" :title="$t('permission.path')">
          <template #default="{ row }">
            {{ row.Path || row.path || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="description" :title="$t('common.description')">
          <template #default="{ row }">
            {{ row.Description || row.description || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="status" :title="$t('table.status')" width="80">
          <template #default="{ row }">
            <el-tag :type="(row.Status !== undefined ? row.Status : (row.status !== undefined ? row.status : 1)) === 1 ? 'success' : 'danger'">
              {{ (row.Status !== undefined ? row.Status : (row.status !== undefined ? row.status : 1)) === 1 ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="sort" :title="$t('common.sort')" width="80">
          <template #default="{ row }">
            {{ row.Sort !== undefined ? row.Sort : (row.sort !== undefined ? row.sort : 0) }}
          </template>
        </vxe-column>
        <vxe-column field="created_at" :title="$t('table.created_at')" />
        <vxe-column :title="$t('table.operation')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
            <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </vxe-column>
      </vxe-table>

      <vxe-pager
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        @page-change="handlePageChange"
      />
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item :label="$t('permission.name')" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item :label="$t('permission.slug')" prop="slug">
          <el-input v-model="formData.slug" />
        </el-form-item>
        <el-form-item :label="$t('permission.method')" prop="method">
          <el-select v-model="formData.method" :placeholder="$t('form.select_method')">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('permission.path')" prop="path">
          <el-input v-model="formData.path" />
        </el-form-item>
        <el-form-item :label="$t('common.description')">
          <el-input v-model="formData.description" type="textarea" />
        </el-form-item>
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :label="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('common.sort')">
          <el-input-number v-model="formData.sort" :min="0" />
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
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import {
  getPermissionList,
  getPermissionDetail,
  createPermission,
  updatePermission,
  deletePermission
} from '../../api/permission'

const { t } = useI18n()
const formRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('permission.edit_permission') : t('permission.add_permission'))

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])

const searchForm = reactive({
  name: '',
  slug: '',
  method: '',
  path: '',
  status: ''
})

const formData = reactive({
  id: null,
  name: '',
  slug: '',
  method: 'GET',
  path: '',
  description: '',
  status: 1,
  sort: 0
})

// 重置表单数据
const resetFormData = () => {
  formData.id = null
  formData.name = ''
  formData.slug = ''
  formData.method = 'GET'
  formData.path = ''
  formData.description = ''
  formData.status = 1
  formData.sort = 0
}

const formRules = computed(() => ({
  name: [{ required: true, message: t('permission.name_required'), trigger: 'blur' }],
  slug: [{ required: true, message: t('permission.slug_required'), trigger: 'blur' }],
  method: [{ required: true, message: t('permission.method_required'), trigger: 'change' }],
  path: [{ required: true, message: t('permission.path_required'), trigger: 'blur' }]
}))

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    // 只添加有值的搜索条件
    if (searchForm.name && searchForm.name.trim()) {
      params.name = searchForm.name.trim()
    }
    if (searchForm.slug && searchForm.slug.trim()) {
      params.slug = searchForm.slug.trim()
    }
    if (searchForm.method) {
      params.method = searchForm.method
    }
    if (searchForm.path && searchForm.path.trim()) {
      params.path = searchForm.path.trim()
    }
    if (searchForm.status) {
      params.status = searchForm.status
    }
    
    console.log('Permission search params:', params)
    const res = await getPermissionList(params)
    console.log('Permission list response:', res)
    
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load permission list error:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  Object.keys(searchForm).forEach(key => {
    searchForm[key] = ''
  })
  pagination.page = 1
  loadData()
}

const handlePageChange = ({ currentPage, pageSize }) => {
  pagination.page = currentPage
  pagination.pageSize = pageSize
  loadData()
}

const handleAdd = () => {
  resetFormData()
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  try {
    console.log('handleEdit - row:', row)
    const res = await getPermissionDetail(row.id)
    console.log('handleEdit - API response:', res)
    
    if (res.data && res.data.permission) {
      const permission = res.data.permission
      console.log('handleEdit - permission data:', permission)
      
      // 处理字段映射，支持 PascalCase 和 snake_case
      const mappedData = {
        id: permission.id || permission.ID,
        name: permission.Name || permission.name || '',
        slug: permission.Slug || permission.slug || '',
        method: permission.Method || permission.method || 'GET',
        path: permission.Path || permission.path || '',
        description: permission.Description || permission.description || '',
        status: permission.Status !== undefined ? permission.Status : (permission.status !== undefined ? permission.status : 1),
        sort: permission.Sort !== undefined ? permission.Sort : (permission.sort !== undefined ? permission.sort : 0)
      }
      
      console.log('handleEdit - mapped formData:', mappedData)
      
      Object.assign(formData, mappedData)
      dialogVisible.value = true
    } else {
      console.error('handleEdit - No permission data in response:', res)
    }
  } catch (error) {
    console.error('Load permission detail error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (formData.id) {
          await updatePermission(formData.id, formData)
          ElMessage.success(t('permission.update_success'))
        } else {
          await createPermission(formData)
          ElMessage.success(t('permission.create_success'))
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
    await ElMessageBox.confirm(t('permission.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deletePermission(row.id)
    ElMessage.success(t('permission.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.permission-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
}
</style>

