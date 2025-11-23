<template>
  <div class="admin-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('admin.title') }}</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            {{ $t('admin.add_admin') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item :label="$t('table.username')">
          <el-input v-model="searchForm.username" :placeholder="$t('form.enter_username')" clearable />
        </el-form-item>
        <el-form-item :label="$t('table.status')">
          <el-select v-model="searchForm.status" :placeholder="$t('form.select_status')" clearable>
            <el-option :label="$t('common.enabled')" value="1" />
            <el-option :label="$t('common.disabled')" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">{{ $t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
          <el-button type="success" @click="handleExport">{{ $t('common.export') }}</el-button>
        </el-form-item>
      </el-form>

      <!-- vxe-table -->
      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        resizable
        height="600"
        @page-change="handlePageChange"
      >
        <vxe-column type="seq" width="60" :title="$t('table.seq')" />
        <vxe-column field="id" :title="$t('table.id')" width="80" />
        <vxe-column field="username" :title="$t('table.username')" />
        <vxe-column field="nickname" :title="$t('table.nickname')" />
        <vxe-column field="email" :title="$t('table.email')" />
        <vxe-column field="phone" :title="$t('table.phone')" />
        <vxe-column field="status" :title="$t('table.status')" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="department" :title="$t('table.department')">
          <template #default="{ row }">
            {{ row.department?.name || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="roles" :title="$t('table.roles')">
          <template #default="{ row }">
            <el-tag v-for="role in row.roles" :key="role.id" style="margin-right: 5px;">
              {{ role.name }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="created_at" :title="$t('table.created_at')" />
        <vxe-column :title="$t('table.operation')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
            <el-button type="warning" link @click="handleResetPassword(row)">{{ $t('admin.reset_password') }}</el-button>
            <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </vxe-column>
      </vxe-table>

      <!-- 分页 -->
      <vxe-pager
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        @page-change="handlePageChange"
      />
    </el-card>

    <!-- 添加/编辑对话框 -->
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
        <el-form-item :label="$t('table.username')" prop="username">
          <el-input v-model="formData.username" :disabled="!!formData.id" />
        </el-form-item>
        <el-form-item :label="$t('common.password')" prop="password" v-if="!formData.id">
          <el-input v-model="formData.password" type="password" />
        </el-form-item>
        <el-form-item :label="$t('table.nickname')" prop="nickname">
          <el-input v-model="formData.nickname" />
        </el-form-item>
        <el-form-item :label="$t('table.email')" prop="email">
          <el-input v-model="formData.email" />
        </el-form-item>
        <el-form-item :label="$t('table.phone')" prop="phone">
          <el-input v-model="formData.phone" />
        </el-form-item>
        <el-form-item :label="$t('table.department')" prop="department_id">
          <el-select v-model="formData.department_id" :placeholder="$t('form.select_department')" clearable>
            <el-option
              v-for="dept in departments"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('table.roles')" prop="role_ids">
          <el-select v-model="formData.role_ids" multiple :placeholder="$t('form.select_role')">
            <el-option
              v-for="role in roles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :label="0">{{ $t('common.disabled') }}</el-radio>
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
import {
  getAdminList,
  createAdmin,
  updateAdmin,
  deleteAdmin,
  exportAdmin,
  resetPassword
} from '../../api/admin'
import { getDepartmentList } from '../../api/department'
import { getRoleList } from '../../api/role'

const { t } = useI18n()
const tableRef = ref(null)
const formRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('admin.edit_admin') : t('admin.add_admin'))

const searchForm = reactive({
  username: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])
const departments = ref([])
const roles = ref([])

const formData = reactive({
  id: null,
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  department_id: null,
  role_ids: [],
  status: 1
})

const formRules = computed(() => ({
  username: [{ required: true, message: t('admin.username_required'), trigger: 'blur' }],
  password: [{ required: true, message: t('admin.password_required'), trigger: 'blur' }]
}))

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...searchForm
    }
    const res = await getAdminList(params)
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load admin list error:', error)
  } finally {
    loading.value = false
  }
}

const loadDepartments = async () => {
  try {
    const res = await getDepartmentList()
    if (res.data && res.data.list) {
      departments.value = res.data.list
    }
  } catch (error) {
    console.error('Load departments error:', error)
  }
}

const loadRoles = async () => {
  try {
    const res = await getRoleList()
    if (res.data && res.data.list) {
      roles.value = res.data.list
    }
  } catch (error) {
    console.error('Load roles error:', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.username = ''
  searchForm.status = ''
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
    username: '',
    password: '',
    nickname: '',
    email: '',
    phone: '',
    department_id: null,
    role_ids: [],
    status: 1
  })
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  Object.assign(formData, {
    id: row.id,
    username: row.username,
    password: '',
    nickname: row.nickname || '',
    email: row.email || '',
    phone: row.phone || '',
    department_id: row.department_id || null,
    role_ids: row.roles ? row.roles.map(r => r.id) : [],
    status: row.status
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const data = { ...formData }
        if (formData.id) {
          // 编辑时，如果没有修改密码，不传 password
          if (!data.password) {
            delete data.password
          }
          await updateAdmin(formData.id, data)
          ElMessage.success(t('admin.update_success'))
        } else {
          await createAdmin(data)
          ElMessage.success(t('admin.create_success'))
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
    await ElMessageBox.confirm(t('admin.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteAdmin(row.id)
    ElMessage.success(t('admin.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

const handleResetPassword = async (row) => {
  try {
    const { value: password } = await ElMessageBox.prompt(t('admin.new_password'), t('admin.reset_password'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      inputType: 'password'
    })
    await resetPassword(row.id, { password })
    ElMessage.success(t('admin.reset_password_success'))
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Reset password error:', error)
    }
  }
}

const handleExport = async () => {
  try {
    const res = await exportAdmin(searchForm)
    if (res.data && res.data.file_url) {
      window.open(res.data.file_url, '_blank')
      ElMessage.success(t('admin.export_success'))
    }
  } catch (error) {
    console.error('Export error:', error)
  }
}

onMounted(() => {
  loadData()
  loadDepartments()
  loadRoles()
})
</script>

<style scoped>
.admin-list {
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
}
</style>

