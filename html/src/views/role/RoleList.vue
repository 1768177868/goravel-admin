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

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ name: '', status: '' }"
        i18n-prefix="role"
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
            :tree-node="column.treeNode"
          >
            <template v-if="column.slot === 'status'" #default="{ row }">
              <el-tag :type="Number(row.status ?? row.Status ?? 1) === 1 ? 'success' : 'danger'">
                {{ Number(row.status ?? row.Status ?? 1) === 1 ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <el-button type="primary" link @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
              <el-button 
                v-if="!isProtectedRole(row)"
                type="danger" 
                link 
                @click="handleDelete(row)"
              >
                {{ $t('common.delete') }}
              </el-button>
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
      width="900px"
      @close="handleDialogClose"
      @opened="handleDialogOpened"
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
        <el-form-item :label="$t('role.menus_and_permissions')">
          <div class="menu-permission-container">
            <div class="menu-header">
              <div class="header-left">
                <el-icon class="header-icon"><Menu /></el-icon>
                <span class="header-title">{{ $t('role.menus_and_permissions') }}</span>
              </div>
            </div>
            <div class="tree-wrapper">
              <el-tree
                :key="treeKey"
                ref="menuPermissionTreeRef"
                :data="menuPermissionTree"
                :props="{ children: 'children', label: 'label' }"
                show-checkbox
                node-key="id"
                :checked-keys="checkedKeys"
                class="menu-permission-tree"
                :expand-on-click-node="false"
                :default-expand-all="false"
                @check="handleTreeCheck"
              >
                <template #default="{ node, data }">
                  <span v-if="data.isMenu" class="menu-node">
                    <el-icon class="node-icon menu-icon"><FolderOpened /></el-icon>
                    <span class="menu-name">{{ data.name }}</span>
                    <el-tag v-if="data.type" size="small" :type="getMenuTypeTag(data.type)" class="menu-type-tag">
                      {{ getMenuTypeText(data.type) }}
                    </el-tag>
                  </span>
                  <span v-else class="permission-node">
                    <el-icon class="node-icon permission-icon"><Key /></el-icon>
                    <span class="permission-name">{{ data.displayDesc || data.name }}</span>
                    <span v-if="data.method" class="permission-method" :class="`method-${data.method.toLowerCase()}`">
                      {{ data.method }}
                    </span>
                    <el-tooltip v-if="data.path" :content="data.path" placement="top">
                      <el-icon class="permission-path-icon">
                        <InfoFilled />
                      </el-icon>
                    </el-tooltip>
                  </span>
                </template>
              </el-tree>
            </div>
          </div>
        </el-form-item>
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model.number="formData.status">
            <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :label="0" :disabled="isProtectedRole(formData)">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
          <div v-if="isProtectedRole(formData)" class="protected-tip">
            <el-icon><Lock /></el-icon>
            <span>{{ $t('role.protected_cannot_disable') }}</span>
          </div>
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
import { ref, reactive, onMounted, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { InfoFilled, Menu, FolderOpened, Key, Lock } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { useTableSort } from '../../composables/useTableSort'
import { getMenuTranslation } from '../../utils/menuTranslation'
import { getRoleList, getRoleDetail, createRole, updateRole, deleteRole } from '../../api/role'
import { getPermissionList } from '../../api/permission'
import { getMenuList } from '../../api/menu'

const { t, te, tm } = useI18n()
const formRef = ref(null)
const tableRef = ref(null)
const menuPermissionTreeRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('role.edit_role') : t('role.add_role'))

const searchForm = reactive({
  name: '',
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
    field: 'name',
    title: t('role.name'),
    sortable: true,
    formatter: ({ row }) => getDisplayValue(row, 'name')
  },
  {
    field: 'slug',
    title: t('role.slug'),
    sortable: true,
    formatter: ({ row }) => getDisplayValue(row, 'slug')
  },
  {
    field: 'description',
    title: t('common.description'),
    sortable: false,
    formatter: ({ row }) => getDisplayValue(row, 'description')
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 80,
    sortable: true,
    slot: 'status'
  },
  {
    field: 'sort',
    title: t('common.sort'),
    width: 80,
    sortable: true,
    formatter: ({ row }) => row.Sort !== undefined ? row.Sort : (row.sort !== undefined ? row.sort : 0)
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    sortable: true,
    formatter: ({ row }) => getDisplayValue(row, 'created_at')
  },
  {
    title: t('table.operation'),
    width: 200,
    fixed: 'right',
    slot: 'operation',
    sortable: false
  }
])

const getDisplayValue = (row, field) => {
  if (!row || !field) return '-'
  const pascalField = field.charAt(0).toUpperCase() + field.slice(1)
  const pascalValue = row[pascalField]
  if (pascalValue !== undefined && pascalValue !== null && pascalValue !== '') {
    return pascalValue
  }
  const camelValue = row[field]
  return camelValue !== undefined && camelValue !== null && camelValue !== '' ? camelValue : '-'
}

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'name',
    label: t('role.name'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'select',
    width: '150px',
    options: [
      { label: t('common.enabled'), value: '1' },
      { label: t('common.disabled'), value: '0' }
    ],
    advanced: false
  }
])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])
const menuPermissionTree = ref([])
const checkedKeys = ref([])
const treeKey = ref(0)
const protectedRoleSlugs = ref(['super-admin'])

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

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'name': 'name',
  'slug': 'slug',
  'description': 'description',
  'status': 'status',
  'sort': 'sort',
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
      page_size: pagination.pageSize
    }
    if (searchForm.name && searchForm.name.trim()) {
      params.name = searchForm.name.trim()
    }
    if (searchForm.status) {
      params.status = searchForm.status
    }
    
    const res = await getRoleList(params)
    if (res.data) {
      const roles = res.data.list || []
      tableData.value = roles.map(role => ({
        ...role,
        id: role.ID || role.id,
        name: role.Name || role.name,
        slug: role.Slug || role.slug,
        description: role.Description || role.description,
        status: role.Status !== undefined ? Number(role.Status) : (role.status !== undefined ? Number(role.status) : 1),
        sort: role.Sort !== undefined ? role.Sort : (role.sort !== undefined ? role.sort : 0),
        created_at: role.CreatedAt || role.created_at
      }))
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load role list error:', error)
  } finally {
    loading.value = false
  }
}

// 获取菜单标题（优先使用 slug，如果没有则使用 path 和 title 映射，最后使用原始标题）
const getMenuTitle = (menu) => {
  if (!menu || typeof menu !== 'object') {
    return ''
  }
  
  // 优先使用 slug 作为翻译键标识
  const slug = menu.Slug || menu.slug || ''
  if (slug) {
    const translated = getMenuTranslation(t, te, slug)
    if (translated) {
      return translated
    }
  }
  
  // 回退到原始标题
  return menu.Title || menu.title || ''
}

// 获取权限名称（优先使用 slug，如果没有则使用 description 或 name）
const getPermissionName = (permission) => {
  if (!permission || typeof permission !== 'object') {
    return ''
  }
  
  // 优先使用 slug 作为翻译键标识
  // 尝试多种可能的字段名（支持 PascalCase 和 snake_case）
  const slug = permission.Slug || permission.slug || ''
  if (slug) {
    const slugKey = `permission.${slug}`

    if (typeof te === 'function' && te(slugKey)) {
      return t(slugKey)
    }

    const messages = typeof tm === 'function' ? tm('permission') : null
    if (messages && Object.prototype.hasOwnProperty.call(messages, slug)) {
      const value = messages[slug]
      if (typeof value === 'string') {
        return value
      }
    }
  }
  
  // 回退到 description 或 name
  return permission.Description || permission.description || permission.Name || permission.name || ''
}

const getModuleNameFromPath = (path) => {
  if (!path) return t('role.other_module')
  
  let cleanPath = path.split('?')[0].replace(/\*/g, '').replace(/\/$/, '')
  cleanPath = cleanPath.replace(/\/\d+(\/|$)/g, '/')
  cleanPath = cleanPath.replace(/\/$/, '')
  
  const parts = cleanPath.split('/').filter(p => p)
  if (parts.length >= 3) {
    const module = parts[parts.length - 1]
    const singular = module.replace(/s$/, '').replace(/-/g, '_')
    const translationKey = `role.module_${singular}`
    const translated = t(translationKey)
    if (translated !== translationKey) {
      return translated
    }
    return module.charAt(0).toUpperCase() + module.slice(1).replace(/-/g, ' ')
  }
  
  return t('role.other_module')
}

const transformPermissionToTree = (permissions) => {
  if (!permissions || !Array.isArray(permissions)) return []
  
  const moduleGroups = {}
  permissions.forEach(perm => {
    const path = perm.Path || perm.path || '/'
    const method = perm.Method || perm.method || ''
    const name = perm.Name || perm.name || ''
    const slug = perm.Slug || perm.slug || ''
    const description = perm.Description || perm.description || ''
    const id = perm.id || perm.ID
    const moduleName = getModuleNameFromPath(path)
    
    if (!moduleGroups[moduleName]) {
      moduleGroups[moduleName] = {
        id: `module_${moduleName}`,
        name: moduleName,
        label: moduleName,
        children: []
      }
    }
    
    let displayLabel = name
    if (description) {
      displayLabel = description
    }
    
    moduleGroups[moduleName].children.push({
      id: id,
      name: name,
      slug: slug,
      method: method,
      path: path,
      description: description,
      label: displayLabel,
      displayName: name,
      displayDesc: description || name
    })
  })
  
  const tree = Object.values(moduleGroups).sort((a, b) => {
    return a.name.localeCompare(b.name)
  })
  
  tree.forEach(group => {
    group.children.sort((a, b) => {
      const methodOrder = { 'GET': 1, 'POST': 2, 'PUT': 3, 'PATCH': 4, 'DELETE': 5 }
      return (methodOrder[a.method] || 99) - (methodOrder[b.method] || 99)
    })
  })
  
  return tree
}

const transformMenuToTree = (menus) => {
  if (!menus || !Array.isArray(menus)) return []
  
  const convertNode = (node) => {
    const children = node.Children || node.children
    const type = node.Type !== undefined ? node.Type : (node.type !== undefined ? node.type : 1)
    const icon = node.Icon || node.icon || ''
    const path = node.Path || node.path || ''
    const slug = node.Slug || node.slug || ''
    
    // 使用多语言函数获取菜单标题
    const title = getMenuTitle(node)
    
    const result = {
      id: node.id,
      name: title,
      label: title,
      slug: slug,
      type: type,
      icon: icon,
      path: path,
      component: node.Component || node.component || '',
      permission: node.Permission || node.permission || '',
      isMenu: true
    }
    if (children && Array.isArray(children) && children.length > 0) {
      result.children = children.map(child => convertNode(child))
    }
    return result
  }
  
  return menus.map(menu => convertNode(menu))
}

const attachPermissionsToMenus = (menuTree, permissions) => {
  if (!permissions || !Array.isArray(permissions)) return menuTree
  
  const permissionMap = new Map()
  permissions.forEach(perm => {
    const id = perm.id || perm.ID
    const menuId = perm.MenuID || perm.menu_id || 0
    if (!permissionMap.has(menuId)) {
      permissionMap.set(menuId, [])
    }
    permissionMap.get(menuId).push(perm)
  })
  
  const processNode = (node) => {
    const result = { ...node }
    
    if (result.isMenu && result.id) {
      const menuId = result.id
      const matchedPermissions = permissionMap.get(menuId) || []
      
      if (matchedPermissions.length > 0) {
        if (!result.children) {
          result.children = []
        }
        
        matchedPermissions.forEach(perm => {
          const method = perm.Method || perm.method || ''
          const id = perm.id || perm.ID
          const slug = perm.Slug || perm.slug || ''
          
          // 使用多语言函数获取权限名称
          const permissionName = getPermissionName(perm)
          
          result.children.push({
            id: id,
            name: permissionName,
            slug: slug,
            method: method,
            path: perm.Path || perm.path || '',
            description: perm.Description || perm.description || '',
            label: permissionName,
            displayDesc: permissionName,
            isMenu: false,
            isPermission: true
          })
        })
        
        result.children.sort((a, b) => {
          if (a.isMenu !== b.isMenu) {
            return a.isMenu ? -1 : 1
          }
          if (!a.isMenu && !b.isMenu) {
            const methodOrder = { 'GET': 1, 'POST': 2, 'PUT': 3, 'PATCH': 4, 'DELETE': 5 }
            return (methodOrder[a.method] || 99) - (methodOrder[b.method] || 99)
          }
          return 0
        })
      }
    }
    
    if (result.children && Array.isArray(result.children)) {
      result.children = result.children.map(child => processNode(child))
    }
    
    return result
  }
  
  return menuTree.map(node => processNode(node))
}

const buildMenuPermissionTree = (menus, permissions) => {
  const menuTreeData = transformMenuToTree(menus)
  const treeWithPermissions = attachPermissionsToMenus(menuTreeData, permissions)
  
  const matchedPermissionIds = new Set()
  const collectPermissionIds = (nodes) => {
    nodes.forEach(node => {
      if (node.isPermission) {
        matchedPermissionIds.add(node.id)
      }
      if (node.children) {
        collectPermissionIds(node.children)
      }
    })
  }
  collectPermissionIds(treeWithPermissions)
  
  const unmatchedPermissions = permissions.filter(perm => {
    const id = perm.id || perm.ID
    return !matchedPermissionIds.has(id)
  })
  
  if (unmatchedPermissions.length > 0) {
    const otherPermissionsNode = {
      id: 'other_permissions',
      name: t('role.other_permissions'),
      label: t('role.other_permissions'),
      isMenu: true,
      type: 1,
      children: unmatchedPermissions.map(perm => {
        const method = perm.Method || perm.method || ''
        const id = perm.id || perm.ID
        const slug = perm.Slug || perm.slug || ''
        const permissionName = getPermissionName(perm)
        
        return {
          id: id,
          name: permissionName,
          slug: slug,
          method: method,
          path: perm.Path || perm.path || '',
          description: perm.Description || perm.description || '',
          label: permissionName,
          displayDesc: permissionName,
          isMenu: false,
          isPermission: true
        }
      })
    }
    
    treeWithPermissions.push(otherPermissionsNode)
  }
  
  return treeWithPermissions
}

const getMenuTypeTag = (type) => {
  const typeMap = {
    1: 'info',
    2: 'success',
    3: 'warning'
  }
  return typeMap[type] || 'info'
}

const getMenuTypeText = (type) => {
  const typeMap = {
    1: t('menu.type_directory'),
    2: t('menu.type_menu'),
    3: t('menu.type_button')
  }
  return typeMap[type] || ''
}

const loadMenuPermissionTree = async () => {
  try {
    const [menuRes, permissionRes] = await Promise.all([
      getMenuList(),
      getPermissionList({ page_size: 1000 })
    ])
    
    const menus = menuRes.data?.menus || menuRes.data?.list || []
    const permissions = permissionRes.data?.list || []
    
    menuPermissionTree.value = buildMenuPermissionTree(menus, permissions)
  } catch (error) {
    console.error('Load menu permission tree error:', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.name = ''
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
  if (dialogVisible.value) {
    dialogVisible.value = false
    setTimeout(() => {
      initAddForm()
    }, 200)
  } else {
    initAddForm()
  }
}

const initAddForm = () => {
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
  checkedKeys.value = []
  treeKey.value++
  dialogVisible.value = true
  setTimeout(() => {
    if (menuPermissionTreeRef.value) {
      menuPermissionTreeRef.value.setCheckedKeys([], false)
      checkedKeys.value = []
    }
  }, 200)
}

const handleEdit = async (row) => {
  try {
    const res = await getRoleDetail(row.id)
    if (res.data && res.data.role) {
      const role = res.data.role
      const rolePermissions = role.Permissions || role.permissions || []
      const roleMenus = role.Menus || role.menus || []
      
      // 确保 ID 是数字类型，与树节点的 ID 类型一致
      const permissionIds = rolePermissions.map(p => {
        const id = p.id || p.ID
        return id ? Number(id) : null
      }).filter(id => id !== null)
      
      const menuIds = roleMenus.map(m => {
        const id = m.id || m.ID
        return id ? Number(id) : null
      }).filter(id => id !== null)
      
      Object.assign(formData, {
        id: role.id || role.ID,
        name: role.Name || role.name || '',
        slug: role.Slug || role.slug || '',
        description: role.Description || role.description || '',
        permission_ids: permissionIds,
        menu_ids: menuIds,
        status: Number(role.Status !== undefined ? role.Status : (role.status !== undefined ? role.status : 1)),
        sort: role.Sort !== undefined ? role.Sort : (role.sort !== undefined ? role.sort : 0)
      })
      
      // 合并所有选中的 key，确保类型一致
      const allCheckedKeys = [...menuIds, ...permissionIds].map(id => Number(id))
      checkedKeys.value = allCheckedKeys
      
      dialogVisible.value = true
      
      // 等待对话框打开和树组件渲染完成后再设置选中状态
      // 注意：这里不立即设置，而是在 handleDialogOpened 中设置
    }
  } catch (error) {
    console.error('Load role detail error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      if (formData.id && isProtectedRole(formData) && formData.status === 0) {
        ElMessage.warning(t('role.protected_cannot_disable'))
        return
      }
      
      submitting.value = true
      try {
        const allCheckedKeys = menuPermissionTreeRef.value?.getCheckedKeys() || []
        const menuIds = []
        const permissionIds = []
        
        const collectIds = (nodes) => {
          nodes.forEach(node => {
            if (allCheckedKeys.includes(node.id)) {
              if (node.isMenu) {
                menuIds.push(node.id)
              } else if (node.isPermission) {
                permissionIds.push(node.id)
              }
            }
            if (node.children) {
              collectIds(node.children)
            }
          })
        }
        collectIds(menuPermissionTree.value)
        
        const statusValue = formData.status !== undefined && formData.status !== null 
          ? Number(formData.status) 
          : 1
        
        const data = {
          name: formData.name,
          slug: formData.slug,
          description: formData.description || '',
          status: statusValue,
          sort: Number(formData.sort) || 0,
          permission_ids: permissionIds,
          menu_ids: menuIds
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
        if (error.response && error.response.data && error.response.data.message) {
          const errorMsg = error.response.data.message
          if (errorMsg === 'role_protected_cannot_disable') {
            ElMessage.error(t('role.protected_cannot_disable'))
          } else {
            ElMessage.error(error.response.data.message || t('role.update_failed'))
          }
        }
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDialogClose = () => {
  checkedKeys.value = []
  treeKey.value++
  if (menuPermissionTreeRef.value) {
    menuPermissionTreeRef.value.setCheckedKeys([], false)
  }
  formRef.value?.resetFields()
}

const handleDialogOpened = () => {
  if (!formData.id) {
    // 新增角色，清空选中状态
    checkedKeys.value = []
    nextTick(() => {
      setTimeout(() => {
        if (menuPermissionTreeRef.value) {
          menuPermissionTreeRef.value.setCheckedKeys([], false)
          checkedKeys.value = []
        }
      }, 100)
    })
  } else {
    // 编辑角色，确保选中状态正确设置
    const allCheckedKeys = [...formData.menu_ids, ...formData.permission_ids].map(id => Number(id))
    nextTick(() => {
      setTimeout(() => {
        if (menuPermissionTreeRef.value) {
          menuPermissionTreeRef.value.setCheckedKeys(allCheckedKeys, false)
          checkedKeys.value = menuPermissionTreeRef.value.getCheckedKeys() || []
        }
      }, 200)
    })
  }
}

const handleTreeCheck = () => {
  if (menuPermissionTreeRef.value) {
    checkedKeys.value = menuPermissionTreeRef.value.getCheckedKeys() || []
  }
}

const isProtectedRole = (row) => {
  const slug = row.slug || row.Slug || ''
  return protectedRoleSlugs.value.includes(slug)
}

const handleDelete = async (row) => {
  if (isProtectedRole(row)) {
    ElMessage.warning(t('role.protected_cannot_delete'))
    return
  }
  
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
      if (error.response && error.response.data && error.response.data.message) {
        const errorMsg = error.response.data.message
        if (errorMsg === 'role_protected_cannot_delete') {
          ElMessage.error(t('role.protected_cannot_delete'))
        } else {
          ElMessage.error(error.response.data.message || t('role.delete_failed'))
        }
      }
    }
  }
}

onMounted(() => {
  initDefaultSort()
  loadData()
  loadMenuPermissionTree()
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


.menu-permission-container {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fafafa;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.menu-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  margin: 0;
  border-bottom: none;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  font-size: 18px;
}

.header-title {
  margin-right: 10px;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.tree-wrapper {
  max-height: 500px;
  overflow-y: auto;
  padding: 12px;
  background: #fff;
  min-height: 200px;
}

.tree-wrapper::-webkit-scrollbar {
  width: 8px;
}

.tree-wrapper::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.tree-wrapper::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.tree-wrapper::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.menu-permission-tree {
  font-size: 14px;
}

.menu-permission-tree :deep(.el-tree-node) {
  margin-bottom: 2px;
}

.menu-permission-tree :deep(.el-tree-node__content) {
  height: 32px;
  padding: 4px 6px;
  border-radius: 4px;
  margin-bottom: 2px;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.menu-permission-tree :deep(.el-tree-node__content:hover) {
  background-color: #f0f9ff;
  border-color: #b3d8ff;
}

.menu-permission-tree :deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: #e1f3ff;
  border-color: #409eff;
}

.menu-permission-tree :deep(.el-tree-node__expand-icon) {
  color: #909399;
  font-size: 14px;
}

.menu-permission-tree :deep(.el-tree-node__expand-icon:hover) {
  color: #409eff;
}

.menu-permission-tree :deep(.el-checkbox) {
  margin-right: 8px;
}

.permission-node,
.menu-node {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  flex: 1;
  min-width: 0;
}

.node-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.menu-icon {
  color: #409eff;
}

.permission-icon {
  color: #67c23a;
}

.permission-name,
.menu-name {
  font-weight: 500;
  color: #303133;
  flex: 1;
  min-width: 0;
  word-break: break-word;
  line-height: 1.5;
}

.menu-name {
  font-size: 14px;
}

.permission-name {
  font-size: 13px;
}

.menu-type-tag {
  margin-left: 4px;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 10px;
  font-weight: 500;
}

.permission-method {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 600;
  color: #fff;
  letter-spacing: 0.3px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.15);
  flex-shrink: 0;
  line-height: 1.2;
}

.method-get {
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
}

.method-post {
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
}

.method-put {
  background: linear-gradient(135deg, #e6a23c 0%, #ebb563 100%);
}

.method-patch {
  background: linear-gradient(135deg, #f56c6c 0%, #f78989 100%);
}

.method-delete {
  background: linear-gradient(135deg, #f56c6c 0%, #f78989 100%);
}

.permission-path-icon {
  color: #909399;
  font-size: 12px;
  cursor: help;
  margin-left: 2px;
  transition: color 0.2s;
  flex-shrink: 0;
}

.permission-path-icon:hover {
  color: #409eff;
}

.protected-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 8px 12px;
  background: #fff7e6;
  border: 1px solid #ffe58f;
  border-radius: 4px;
  color: #d48806;
  font-size: 12px;
}

.protected-tip .el-icon {
  font-size: 14px;
}
</style>
