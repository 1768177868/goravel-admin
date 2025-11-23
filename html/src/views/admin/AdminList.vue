<template>
  <div class="admin-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>管理员列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            添加管理员
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="用户名">
          <el-input v-model="searchForm.username" placeholder="请输入用户名" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="请选择状态" clearable>
            <el-option label="启用" value="1" />
            <el-option label="禁用" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleExport">导出</el-button>
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
        <vxe-column type="seq" width="60" title="序号" />
        <vxe-column field="id" title="ID" width="80" />
        <vxe-column field="username" title="用户名" />
        <vxe-column field="nickname" title="昵称" />
        <vxe-column field="email" title="邮箱" />
        <vxe-column field="phone" title="手机号" />
        <vxe-column field="status" title="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="department" title="部门">
          <template #default="{ row }">
            {{ row.department?.name || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="roles" title="角色">
          <template #default="{ row }">
            <el-tag v-for="role in row.roles" :key="role.id" style="margin-right: 5px;">
              {{ role.name }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="created_at" title="创建时间" />
        <vxe-column title="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="warning" link @click="handleResetPassword(row)">重置密码</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
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
        <el-form-item label="用户名" prop="username">
          <el-input v-model="formData.username" :disabled="!!formData.id" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!formData.id">
          <el-input v-model="formData.password" type="password" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="formData.nickname" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="formData.email" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="formData.phone" />
        </el-form-item>
        <el-form-item label="部门" prop="department_id">
          <el-select v-model="formData.department_id" placeholder="请选择部门" clearable>
            <el-option
              v-for="dept in departments"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="角色" prop="role_ids">
          <el-select v-model="formData.role_ids" multiple placeholder="请选择角色">
            <el-option
              v-for="role in roles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
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

const tableRef = ref(null)
const formRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加管理员')

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

const formRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

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
  dialogTitle.value = '添加管理员'
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
  dialogTitle.value = '编辑管理员'
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
          ElMessage.success('更新成功')
        } else {
          await createAdmin(data)
          ElMessage.success('创建成功')
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
    await ElMessageBox.confirm('确定要删除该管理员吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteAdmin(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

const handleResetPassword = async (row) => {
  try {
    const { value: password } = await ElMessageBox.prompt('请输入新密码', '重置密码', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputType: 'password'
    })
    await resetPassword(row.id, { password })
    ElMessage.success('密码重置成功')
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
      ElMessage.success('导出成功')
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

