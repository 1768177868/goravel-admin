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
      width="900px"
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
          <div class="permission-container">
            <div class="permission-header">
              <el-button size="small" @click="handleSelectAllPermissions">{{ $t('role.select_all') }}</el-button>
              <el-button size="small" @click="handleUnselectAllPermissions">{{ $t('role.unselect_all') }}</el-button>
            </div>
            <el-tree
              ref="permissionTreeRef"
              :data="permissionTree"
              :props="{ children: 'children', label: 'label' }"
              show-checkbox
              node-key="id"
              :default-checked-keys="formData.permission_ids"
              class="permission-tree"
            >
              <template #default="{ node, data }">
                <span class="permission-node">
                  <span class="permission-name">{{ data.name }}</span>
                  <span v-if="data.method" class="permission-method" :class="`method-${data.method.toLowerCase()}`">
                    {{ data.method }}
                  </span>
                  <span v-if="data.path" class="permission-path">{{ data.path }}</span>
                  <span v-if="data.description" class="permission-desc">{{ data.description }}</span>
                </span>
              </template>
            </el-tree>
          </div>
        </el-form-item>
        <el-form-item :label="$t('role.menus')">
          <div class="menu-container">
            <div class="menu-header">
              <el-button size="small" @click="handleSelectAllMenus">{{ $t('role.select_all') }}</el-button>
              <el-button size="small" @click="handleUnselectAllMenus">{{ $t('role.unselect_all') }}</el-button>
            </div>
            <el-tree
              ref="menuTreeRef"
              :data="menuTree"
              :props="{ children: 'children', label: 'label' }"
              show-checkbox
              node-key="id"
              :default-checked-keys="formData.menu_ids"
              class="menu-tree"
            >
              <template #default="{ node, data }">
                <span class="menu-node">
                  <span class="menu-name">{{ data.name }}</span>
                  <el-tag v-if="data.type" size="small" :type="getMenuTypeTag(data.type)">
                    {{ getMenuTypeText(data.type) }}
                  </el-tag>
                  <span v-if="data.path" class="menu-path">{{ data.path }}</span>
                </span>
              </template>
            </el-tree>
          </div>
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

// 转换权限数据为树形结构（按路径分组）
const transformPermissionToTree = (permissions) => {
  if (!permissions || !Array.isArray(permissions)) return []
  
  // 按路径分组
  const pathGroups = {}
  permissions.forEach(perm => {
    const path = perm.Path || perm.path || '/'
    const method = perm.Method || perm.method || ''
    const name = perm.Name || perm.name || ''
    const slug = perm.Slug || perm.slug || ''
    const description = perm.Description || perm.description || ''
    const id = perm.id || perm.ID
    
    // 提取路径前缀作为分组（例如：/api/admin/users -> /api/admin/users）
    const pathKey = path.split('?')[0] // 移除查询参数
    
    if (!pathGroups[pathKey]) {
      pathGroups[pathKey] = {
        id: `path_${pathKey}`,
        name: pathKey,
        path: pathKey,
        label: pathKey,
        children: []
      }
    }
    
    pathGroups[pathKey].children.push({
      id: id,
      name: name,
      slug: slug,
      method: method,
      path: path,
      description: description,
      label: `${method} ${name}${description ? ` - ${description}` : ''}`
    })
  })
  
  // 转换为数组并按路径排序
  const tree = Object.values(pathGroups).sort((a, b) => {
    return a.path.localeCompare(b.path)
  })
  
  // 对每个分组下的权限按方法排序
  tree.forEach(group => {
    group.children.sort((a, b) => {
      const methodOrder = { 'GET': 1, 'POST': 2, 'PUT': 3, 'PATCH': 4, 'DELETE': 5 }
      return (methodOrder[a.method] || 99) - (methodOrder[b.method] || 99)
    })
  })
  
  return tree
}

// 转换菜单数据为树形结构（支持后端返回的树形结构）
const transformMenuToTree = (menus) => {
  if (!menus || !Array.isArray(menus)) return []
  
  const convertNode = (node) => {
    const children = node.Children || node.children
    const title = node.Title || node.name || ''
    const type = node.Type !== undefined ? node.Type : (node.type !== undefined ? node.type : 1)
    const icon = node.Icon || node.icon || ''
    const path = node.Path || node.path || ''
    
    const result = {
      id: node.id,
      name: title,
      label: title,
      type: type,
      icon: icon,
      path: path,
      component: node.Component || node.component || '',
      permission: node.Permission || node.permission || ''
    }
    if (children && Array.isArray(children) && children.length > 0) {
      result.children = children.map(child => convertNode(child))
    }
    return result
  }
  
  return menus.map(menu => convertNode(menu))
}

// 获取菜单类型标签样式
const getMenuTypeTag = (type) => {
  const typeMap = {
    1: 'info',    // 目录
    2: 'success', // 菜单
    3: 'warning'  // 按钮
  }
  return typeMap[type] || 'info'
}

// 获取菜单类型文本
const getMenuTypeText = (type) => {
  const typeMap = {
    1: t('menu.type_directory'),
    2: t('menu.type_menu'),
    3: t('menu.type_button')
  }
  return typeMap[type] || ''
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

// 全选/取消全选权限
const handleSelectAllPermissions = () => {
  if (permissionTreeRef.value) {
    const allKeys = getAllPermissionKeys(permissionTree)
    permissionTreeRef.value.setCheckedKeys(allKeys)
  }
}

const handleUnselectAllPermissions = () => {
  if (permissionTreeRef.value) {
    permissionTreeRef.value.setCheckedKeys([])
  }
}

// 全选/取消全选菜单
const handleSelectAllMenus = () => {
  if (menuTreeRef.value) {
    const allKeys = getAllMenuKeys(menuTree.value)
    menuTreeRef.value.setCheckedKeys(allKeys)
  }
}

const handleUnselectAllMenus = () => {
  if (menuTreeRef.value) {
    menuTreeRef.value.setCheckedKeys([])
  }
}

// 递归获取所有权限ID（包括子节点）
const getAllPermissionKeys = (tree) => {
  const keys = []
  const traverse = (nodes) => {
    nodes.forEach(node => {
      // 只添加叶子节点（实际权限），不添加分组节点
      if (!node.children || node.children.length === 0) {
        keys.push(node.id)
      } else {
        traverse(node.children)
      }
    })
  }
  traverse(tree)
  return keys
}

// 递归获取所有菜单ID
const getAllMenuKeys = (tree) => {
  const keys = []
  const traverse = (nodes) => {
    nodes.forEach(node => {
      keys.push(node.id)
      if (node.children && node.children.length > 0) {
        traverse(node.children)
      }
    })
  }
  traverse(tree)
  return keys
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

.permission-container,
.menu-container {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 10px;
  max-height: 400px;
  overflow-y: auto;
}

.permission-header,
.menu-header {
  margin-bottom: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
}

.permission-tree,
.menu-tree {
  font-size: 14px;
}

.permission-node {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.permission-name {
  font-weight: 500;
  color: #303133;
}

.permission-method {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 500;
  color: #fff;
}

.method-get {
  background-color: #67c23a;
}

.method-post {
  background-color: #409eff;
}

.method-put {
  background-color: #e6a23c;
}

.method-patch {
  background-color: #f56c6c;
}

.method-delete {
  background-color: #f56c6c;
}

.permission-path {
  color: #909399;
  font-size: 12px;
  font-family: monospace;
}

.permission-desc {
  color: #909399;
  font-size: 12px;
}

.menu-node {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.menu-name {
  font-weight: 500;
  color: #303133;
}

.menu-path {
  color: #909399;
  font-size: 12px;
  font-family: monospace;
}
</style>

