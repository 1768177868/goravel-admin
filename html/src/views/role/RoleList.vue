<template>
  <div class="role-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('role.title') }}</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            {{ $t('role.add_role') }}
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item :label="$t('role.name')">
          <el-input v-model="searchForm.name" :placeholder="$t('form.please_enter') + $t('role.name')" clearable />
        </el-form-item>
        <el-form-item :label="$t('table.status')">
          <el-select v-model="searchForm.status" :placeholder="$t('form.select_status')" clearable style="width: 150px">
            <el-option :label="$t('common.enabled')" value="1" />
            <el-option :label="$t('common.disabled')" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">{{ $t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
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
        <vxe-column field="name" :title="$t('role.name')">
          <template #default="{ row }">
            {{ row.Name || row.name || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="slug" :title="$t('role.slug')">
          <template #default="{ row }">
            {{ row.Slug || row.slug || '-' }}
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
        <vxe-column :title="$t('table.operation')" width="200" fixed="right">
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
        <el-form-item :label="$t('role.name')" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item :label="$t('role.slug')" prop="slug">
          <el-input v-model="formData.slug" />
        </el-form-item>
        <el-form-item :label="$t('common.description')">
          <el-input v-model="formData.description" type="textarea" />
        </el-form-item>
        <el-form-item :label="$t('role.permissions')">
          <el-tree
            ref="permissionTreeRef"
            :data="permissionTree"
            :props="{ children: 'children', label: 'name' }"
            show-checkbox
            node-key="id"
            :default-checked-keys="formData.permission_ids"
          />
        </el-form-item>
        <el-form-item :label="$t('role.menus')">
          <el-tree
            ref="menuTreeRef"
            :data="menuTree"
            :props="{ children: 'children', label: 'name' }"
            show-checkbox
            node-key="id"
            :default-checked-keys="formData.menu_ids"
          />
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
import { getRoleList, getRoleDetail, createRole, updateRole, deleteRole } from '../../api/role'
import { getPermissionList } from '../../api/permission'
import { getMenuList } from '../../api/menu'

const { t } = useI18n()
const formRef = ref(null)
const permissionTreeRef = ref(null)
const menuTreeRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('role.edit_role') : t('role.add_role'))

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

const formRules = computed(() => ({
  name: [{ required: true, message: t('role.name_required'), trigger: 'blur' }],
  slug: [{ required: true, message: t('role.slug_required'), trigger: 'blur' }]
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
    if (searchForm.status) {
      params.status = searchForm.status
    }
    
    console.log('Role search params:', params)
    const res = await getRoleList(params)
    console.log('Role list response:', res)
    
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

// 转换权限数据为树形结构（权限是扁平结构，需要转换为树形）
const transformPermissionToTree = (permissions) => {
  if (!permissions || !Array.isArray(permissions)) return []
  
  return permissions.map(perm => ({
    id: perm.id || perm.ID,
    name: perm.Name || perm.name || '',
    slug: perm.Slug || perm.slug || ''
  }))
}

// 转换菜单数据为树形结构（支持后端返回的树形结构）
const transformMenuToTree = (menus) => {
  if (!menus || !Array.isArray(menus)) return []
  
  const convertNode = (node) => {
    const children = node.Children || node.children
    const result = {
      id: node.id,
      name: node.Title || node.name || '',
    }
    if (children && Array.isArray(children) && children.length > 0) {
      result.children = children.map(child => convertNode(child))
    }
    return result
  }
  
  return menus.map(menu => convertNode(menu))
}

const loadPermissions = async () => {
  try {
    const res = await getPermissionList({ page_size: 1000 }) // 获取所有权限
    if (res.data && res.data.list) {
      permissionTree.value = transformPermissionToTree(res.data.list)
    }
  } catch (error) {
    console.error('Load permissions error:', error)
  }
}

const loadMenus = async () => {
  try {
    const res = await getMenuList()
    // 菜单返回的是 menus 字段，不是 list
    const menus = res.data?.menus || res.data?.list || []
    menuTree.value = transformMenuToTree(menus)
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
  try {
    const res = await getRoleDetail(row.id)
    if (res.data && res.data.role) {
      const role = res.data.role
      // 处理字段映射，支持 PascalCase 和 snake_case
      const rolePermissions = role.Permissions || role.permissions
      const roleMenus = role.Menus || role.menus
      
      Object.assign(formData, {
        id: role.id,
        name: role.Name || role.name || '',
        slug: role.Slug || role.slug || '',
        description: role.Description || role.description || '',
        permission_ids: rolePermissions ? rolePermissions.map(p => p.id || p.ID).filter(id => id) : [],
        menu_ids: roleMenus ? roleMenus.map(m => m.id || m.ID).filter(id => id) : [],
        status: role.Status !== undefined ? role.Status : (role.status !== undefined ? role.status : 1),
        sort: role.Sort !== undefined ? role.Sort : (role.sort !== undefined ? role.sort : 0)
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
          ElMessage.success(t('role.update_success'))
        } else {
          await createRole(data)
          ElMessage.success(t('role.create_success'))
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
    await ElMessageBox.confirm(t('role.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteRole(row.id)
    ElMessage.success(t('role.delete_success'))
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

