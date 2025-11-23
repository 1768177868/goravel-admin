<template>
  <div class="role-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            添加角色
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="角色名称">
          <el-input v-model="searchForm.name" placeholder="请输入角色名称" clearable />
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
        </el-form-item>
      </el-form>

      <vxe-table
        :data="tableData"
        :loading="loading"
        border
        resizable
        height="600"
      >
        <vxe-column type="seq" width="60" title="序号" />
        <vxe-column field="id" title="ID" width="80" />
        <vxe-column field="name" title="角色名称" />
        <vxe-column field="slug" title="标识" />
        <vxe-column field="description" title="描述" />
        <vxe-column field="status" title="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="sort" title="排序" width="80" />
        <vxe-column field="created_at" title="创建时间" />
        <vxe-column title="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
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
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item label="标识" prop="slug">
          <el-input v-model="formData.slug" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="formData.description" type="textarea" />
        </el-form-item>
        <el-form-item label="权限">
          <el-tree
            ref="permissionTreeRef"
            :data="permissionTree"
            :props="{ children: 'children', label: 'name' }"
            show-checkbox
            node-key="id"
            :default-checked-keys="formData.permission_ids"
          />
        </el-form-item>
        <el-form-item label="菜单">
          <el-tree
            ref="menuTreeRef"
            :data="menuTree"
            :props="{ children: 'children', label: 'name' }"
            show-checkbox
            node-key="id"
            :default-checked-keys="formData.menu_ids"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="formData.sort" :min="0" />
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
import { getRoleList, getRoleDetail, createRole, updateRole, deleteRole } from '../../api/role'
import { getPermissionList } from '../../api/permission'
import { getMenuList } from '../../api/menu'

const formRef = ref(null)
const permissionTreeRef = ref(null)
const menuTreeRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加角色')

const searchForm = reactive({
  name: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])
const permissionTree = ref([])
const menuTree = ref([])

const formData = reactive({
  id: null,
  name: '',
  slug: '',
  description: '',
  permission_ids: [],
  menu_ids: [],
  status: 1,
  sort: 0
})

const formRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  slug: [{ required: true, message: '请输入标识', trigger: 'blur' }]
}

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...searchForm
    }
    const res = await getRoleList(params)
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load role list error:', error)
  } finally {
    loading.value = false
  }
}

const loadPermissions = async () => {
  try {
    const res = await getPermissionList()
    if (res.data && res.data.list) {
      permissionTree.value = res.data.list
    }
  } catch (error) {
    console.error('Load permissions error:', error)
  }
}

const loadMenus = async () => {
  try {
    const res = await getMenuList()
    if (res.data && res.data.list) {
      menuTree.value = res.data.list
    }
  } catch (error) {
    console.error('Load menus error:', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.name = ''
  searchForm.status = ''
  handleSearch()
}

const handlePageChange = ({ currentPage, pageSize }) => {
  pagination.page = currentPage
  pagination.pageSize = pageSize
  loadData()
}

const handleAdd = () => {
  dialogTitle.value = '添加角色'
  Object.assign(formData, {
    id: null,
    name: '',
    slug: '',
    description: '',
    permission_ids: [],
    menu_ids: [],
    status: 1,
    sort: 0
  })
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  dialogTitle.value = '编辑角色'
  try {
    const res = await getRoleDetail(row.id)
    if (res.data && res.data.role) {
      const role = res.data.role
      Object.assign(formData, {
        id: role.id,
        name: role.name,
        slug: role.slug,
        description: role.description || '',
        permission_ids: role.permissions ? role.permissions.map(p => p.id) : [],
        menu_ids: role.menus ? role.menus.map(m => m.id) : [],
        status: role.status,
        sort: role.sort || 0
      })
      dialogVisible.value = true
      // 等待树组件渲染后设置选中状态
      setTimeout(() => {
        if (permissionTreeRef.value) {
          permissionTreeRef.value.setCheckedKeys(formData.permission_ids)
        }
        if (menuTreeRef.value) {
          menuTreeRef.value.setCheckedKeys(formData.menu_ids)
        }
      }, 100)
    }
  } catch (error) {
    console.error('Load role detail error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const data = {
          ...formData,
          permission_ids: permissionTreeRef.value?.getCheckedKeys() || [],
          menu_ids: menuTreeRef.value?.getCheckedKeys() || []
        }
        if (formData.id) {
          await updateRole(formData.id, data)
          ElMessage.success('更新成功')
        } else {
          await createRole(data)
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
    await ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteRole(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

onMounted(() => {
  loadData()
  loadPermissions()
  loadMenus()
})
</script>

<style scoped>
.role-list {
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

