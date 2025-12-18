<template>
  <div class="role-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('role.title') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('role.store').disabled"
            @click="handleAdd"
          >
            <el-icon><PlusIcon /></el-icon>
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
        :sort-config="{ multiple: false, trigger: 'default' }"
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
              <el-switch
                :model-value="Number(row.status ?? row.Status ?? 1) === 1"
                :disabled="isProtectedRole(row) || getButtonState('role.update').disabled"
                @change="(val) => handleStatusChange(row, val)"
              />
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <el-button 
                type="primary" 
                link 
                :disabled="getButtonState('role.update').disabled"
                @click="handleEdit(row)"
              >
                {{ $t('common.edit') }}
              </el-button>
              <el-button 
                v-if="!isProtectedRole(row)"
                type="danger" 
                link 
                :disabled="getButtonState('role.destroy').disabled"
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
      <div v-loading="formLoading">
        <el-form
          ref="formRef"
          :model="formData"
          :rules="formRules"
          label-width="100px"
        >
        <el-form-item :label="$t('role.name')" prop="name">
          <el-input v-model="formData.name" :disabled="formLoading" />
        </el-form-item>
        <el-form-item :label="$t('role.slug')" prop="slug">
          <el-input 
            v-model="formData.slug" 
            :disabled="isProtectedRole(formData) || formLoading"
            :placeholder="isProtectedRole(formData) ? $t('role.protected_role_slug_disabled') : ''"
          />
        </el-form-item>
        <el-form-item :label="$t('common.description')">
          <el-input v-model="formData.description" type="textarea" :disabled="formLoading" />
        </el-form-item>
        <el-form-item 
          v-if="!isProtectedRole(formData)" 
          :label="$t('role.menus_and_permissions')"
        >
          <div class="menu-permission-container">
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
                :disabled="formLoading"
                @check="handleTreeCheck"
              >
                <template #default="{ node, data }">
                  <span v-if="data.isMenu" class="menu-node">
                    <el-icon class="node-icon menu-icon"><FolderOpenedIcon /></el-icon>
                    <span class="menu-name">{{ data.name }}</span>
                    <el-tag v-if="data.type" size="small" :type="getMenuTypeTag(data.type)" class="menu-type-tag">
                      {{ getMenuTypeText(data.type) }}
                    </el-tag>
                  </span>
                  <span v-else class="permission-node">
                    <el-icon class="node-icon permission-icon"><KeyIcon /></el-icon>
                    <span class="permission-name">{{ data.displayDesc || data.name }}</span>
                    <span v-if="data.method" class="permission-method" :class="`method-${data.method.toLowerCase()}`">
                      {{ data.method }}
                    </span>
                    <el-tooltip v-if="data.path" :content="data.path" placement="top">
                      <el-icon class="permission-path-icon">
                        <InfoFilledIcon />
                      </el-icon>
                    </el-tooltip>
                  </span>
                </template>
              </el-tree>
            </div>
          </div>
        </el-form-item>
        <el-form-item v-if="isProtectedRole(formData)" :label="$t('role.menus_and_permissions')">
          <div class="protected-role-tip">
            <el-icon><LockIcon /></el-icon>
            <span>{{ $t('role.super_admin_has_all_permissions') }}</span>
          </div>
        </el-form-item>
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model.number="formData.status" :disabled="formLoading">
            <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :label="0" :disabled="isProtectedRole(formData)">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
          <div v-if="isProtectedRole(formData)" class="protected-tip">
            <el-icon><LockIcon /></el-icon>
            <span>{{ $t('role.protected_cannot_disable') }}</span>
          </div>
        </el-form-item>
        <el-form-item :label="$t('common.sort')">
          <el-input-number v-model="formData.sort" :min="0" :disabled="formLoading" />
        </el-form-item>
      </el-form>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, nextTick, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, InfoFilled, Menu, FolderOpened, Key, Lock } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { useTableSort } from '../../composables/useTableSort'
import { usePermission } from '../../composables/usePermission'
import { getMenuTranslation } from '../../utils/menuTranslation'
import { getRoleList, getRoleDetail, createRole, updateRole, deleteRole } from '../../api/role'
import { getPermissionList } from '../../api/permission'
import { getMenuList } from '../../api/menu'

// 使用 markRaw 标记图标组件，避免被 Vue 做成响应式对象
const PlusIcon = markRaw(Plus)
const InfoFilledIcon = markRaw(InfoFilled)
const MenuIcon = markRaw(Menu)
const FolderOpenedIcon = markRaw(FolderOpened)
const KeyIcon = markRaw(Key)
const LockIcon = markRaw(Lock)

const { t, te, tm } = useI18n()
const { getButtonState } = usePermission()
const formRef = ref(null)
const tableRef = ref(null)
const menuPermissionTreeRef = ref(null)
const loading = ref(false)
const formLoading = ref(false)
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
    width: 100,
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
      id: `menu_${node.id}`, // 添加前缀避免ID冲突
      rawId: node.id,
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
    
    if (result.isMenu && result.rawId) {
      const menuId = result.rawId
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
            id: `perm_${id}`, // 添加前缀避免ID冲突
            rawId: id,
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
        matchedPermissionIds.add(node.rawId)
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
          id: `perm_${id}`, // 添加前缀
          rawId: id,
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
  formLoading.value = false
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
  formLoading.value = true
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
      
      // 确保菜单权限树数据已加载
      if (menuPermissionTree.value.length === 0) {
        await loadMenuPermissionTree()
      }
      
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
      
      // 注意：不要在这里设置 checkedKeys，因为如果包含菜单ID，会导致菜单下的所有权限被选中
      // 只设置权限ID，让 handleDialogOpened 来处理
      checkedKeys.value = []
      
      // console.log('Role edit - menuIds:', menuIds, 'permissionIds:', permissionIds)
      
      dialogVisible.value = true
      
      // 等待对话框打开和树组件渲染完成后再设置选中状态
      // 注意：这里不立即设置，而是在 handleDialogOpened 中设置
    }
  } catch (error) {
    console.error('Load role detail error:', error)
  } finally {
    formLoading.value = false
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
        const statusValue = formData.status !== undefined && formData.status !== null 
          ? Number(formData.status) 
          : 1
        
        const data = {
          name: formData.name,
          slug: formData.slug,
          description: formData.description || '',
          status: statusValue,
          sort: Number(formData.sort) || 0
        }
        
        // super-admin 角色不需要设置菜单和权限，因为它拥有所有权限
        if (!isProtectedRole(formData)) {
          // 获取所有选中的节点（完全选中的）
          const allCheckedKeys = menuPermissionTreeRef.value?.getCheckedKeys() || []
          
          // 收集菜单ID和权限ID
          const menuIds = []
          const permissionIds = []
          
          // 递归收集所有选中的菜单和权限ID
          const collectIds = (nodes) => {
            nodes.forEach(node => {
              const nodeId = node.id.toString()
              const isChecked = allCheckedKeys.includes(node.id)
              
              // 如果是权限节点且被选中
              if (node.isPermission && isChecked) {
                // 提取权限ID (去除 perm_ 前缀)
                if (nodeId.startsWith('perm_')) {
                  permissionIds.push(Number(nodeId.replace('perm_', '')))
                } else {
                  permissionIds.push(Number(nodeId))
                }
              }
              
              // 如果是菜单节点，检查是否应该保存
              if (node.isMenu && nodeId !== 'other_permissions') {
                // 检查该菜单下是否有权限被选中
                const hasSelectedPermission = checkHasSelectedPermission(node, allCheckedKeys)
                // 如果菜单被完全选中，或者有子权限被选中（半选状态），则保存该菜单
                if (isChecked || hasSelectedPermission) {
                  // 提取菜单ID (去除 menu_ 前缀)
                  if (nodeId.startsWith('menu_')) {
                    menuIds.push(Number(nodeId.replace('menu_', '')))
                  } else {
                    menuIds.push(Number(nodeId))
                  }
                }
              }
              
              // 递归处理子节点
              if (node.children) {
                collectIds(node.children)
              }
            })
          }
          
          // 辅助函数：检查节点下是否有权限被选中
          const checkHasSelectedPermission = (node, checkedKeys) => {
            if (!node.children) return false
            for (const child of node.children) {
              if (child.isPermission && checkedKeys.includes(child.id)) {
                return true
              }
              if (child.children && checkHasSelectedPermission(child, checkedKeys)) {
                return true
              }
            }
            return false
          }
          
          collectIds(menuPermissionTree.value)
          
          // 去重
          data.permission_ids = [...new Set(permissionIds)]
          data.menu_ids = [...new Set(menuIds)]
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
        // 如果错误已经在响应拦截器中处理过，就不再重复显示
        if (!error.__handled) {
        const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
        ElMessage.error(errorMessage)
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

const handleDialogOpened = async () => {
  if (!formData.id) {
    // 新增角色，清空选中状态
    checkedKeys.value = []
    await nextTick()
    setTimeout(() => {
      if (menuPermissionTreeRef.value) {
        menuPermissionTreeRef.value.setCheckedKeys([], false)
        checkedKeys.value = []
      }
    }, 100)
  } else {
    // 编辑角色，确保选中状态正确设置
    // 确保菜单权限树数据已加载
    if (menuPermissionTree.value.length === 0) {
      await loadMenuPermissionTree()
      // 等待数据加载完成
      await nextTick()
    }
    
    // 从 formData 中获取选中的菜单和权限 ID
    const menuIds = (formData.menu_ids || []).map(id => Number(id))
    const permissionIds = (formData.permission_ids || []).map(id => Number(id))
    
    // 只设置权限ID，不设置菜单ID
    // Element Plus 的 el-tree 会自动根据子节点的选中状态来显示父节点的半选状态
    // 如果设置了菜单ID，会导致该菜单下的所有权限都被选中
    const checkedPermissionKeys = permissionIds.map(id => `perm_${id}`)
    
    // 等待 DOM 更新
    await nextTick()
    
    // 多次尝试设置选中状态，确保树组件已完全渲染
    const setCheckedKeysWithRetry = (retries = 5) => {
      if (retries <= 0) {
        console.warn('Failed to set checked keys after retries')
        return
      }
      
      setTimeout(() => {
        if (!menuPermissionTreeRef.value) {
          // 树组件引用不存在，重试
          if (retries > 1) {
            setCheckedKeysWithRetry(retries - 1)
          }
          return
        }
        
        if (menuPermissionTree.value.length === 0) {
          // 树数据还没准备好，重试
          if (retries > 1) {
            setCheckedKeysWithRetry(retries - 1)
          }
          return
        }
        
        try {
          // 验证权限ID是否存在于树中
          const permissionIdSet = new Set()
          const collectPermissionIds = (nodes) => {
            nodes.forEach(node => {
              if (node.isPermission && node.id) {
                permissionIdSet.add(node.id)
              }
              if (node.children) {
                collectPermissionIds(node.children)
              }
            })
          }
          collectPermissionIds(menuPermissionTree.value)
          
          // 过滤出存在于树中的权限ID
          const validPermissionIds = checkedPermissionKeys.filter(id => permissionIdSet.has(id))
          
          if (validPermissionIds.length !== checkedPermissionKeys.length) {
            console.warn('Some permission IDs not found in tree:', {
              requested: checkedPermissionKeys,
              valid: validPermissionIds,
              allInTree: Array.from(permissionIdSet)
            })
          }
          
          // 只设置权限ID，第二个参数为 false 表示不自动选中父节点
          // Element Plus 会自动根据子节点的选中状态来显示父节点的半选状态
          menuPermissionTreeRef.value.setCheckedKeys(validPermissionIds, false)
          
          // 等待一下让树组件更新
          setTimeout(() => {
            // 获取实际选中的 keys
            const actualCheckedKeys = menuPermissionTreeRef.value.getCheckedKeys() || []
            checkedKeys.value = actualCheckedKeys
            
            // 验证是否设置成功
            const allPermissionsSet = validPermissionIds.length === 0 || 
              validPermissionIds.every(id => actualCheckedKeys.includes(id))
            
            // console.log('Set checked keys result:', {
            //   requestedPermissionIds: permissionIds,
            //   validPermissionIds,
            //   actualCheckedKeys,
            //   allPermissionsSet,
            //   checkedKeysValue: checkedKeys.value
            // })
            
            if (!allPermissionsSet && retries > 1) {
              // 如果还有未选中的，重试
              // console.log(`Retrying set checked keys, remaining retries: ${retries - 1}`)
              setCheckedKeysWithRetry(retries - 1)
            } else {
              // console.log('Checked keys set successfully')
            }
          }, 100)
        } catch (error) {
          console.error('Set checked keys error:', error)
          if (retries > 1) {
            setCheckedKeysWithRetry(retries - 1)
          }
        }
      }, 300)
    }
    
    setCheckedKeysWithRetry()
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

const handleStatusChange = async (row, newStatus) => {
  // 检查是否是受保护角色
  if (isProtectedRole(row) && !newStatus) {
    ElMessage.warning(t('role.protected_cannot_disable'))
    // 恢复开关状态
    loadData()
    return
  }

  try {
    const statusValue = newStatus ? 1 : 0
    await updateRole(row.id, {
      status: statusValue
    })
    ElMessage.success(newStatus ? t('role.enable_success') : t('role.disable_success'))
    // 更新本地数据
    const role = tableData.value.find(r => r.id === row.id)
    if (role) {
      role.status = statusValue
      role.Status = statusValue
    }
  } catch (error) {
    console.error('Status change error:', error)
    // 恢复开关状态
    loadData()
    // 如果错误已经在响应拦截器中处理过，就不再重复显示
    if (!error.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
      ElMessage.error(errorMessage)
    }
  }
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
      // 如果错误已经在响应拦截器中处理过，就不再重复显示
      if (!error.__handled) {
        const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
        ElMessage.error(errorMessage)
      }
    }
  }
}

onMounted(async () => {
  try {
    initDefaultSort()
    await Promise.all([
      loadData(),
      loadMenuPermissionTree()
    ])
  } catch (error) {
    console.error('RoleList onMounted error:', error)
    ElMessage.error('页面加载失败，请刷新重试')
  }
})
</script>

<style scoped>
.role-list {
  background: var(--card-bg, white);
  border-radius: 4px;
  transition: background-color 0.3s ease;
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


.tree-wrapper {
  max-height: 500px;
  overflow-y: auto;
  padding: 12px;
  background: var(--card-bg, #fff);
  transition: background-color 0.3s ease;
  min-height: 200px;
}

.tree-wrapper::-webkit-scrollbar {
  width: 8px;
}

.tree-wrapper::-webkit-scrollbar-track {
  background: var(--bg-color-tertiary, #f1f1f1);
  border-radius: 4px;
  transition: background-color 0.3s ease;
}

.tree-wrapper::-webkit-scrollbar-thumb {
  background: var(--border-color-base, #c1c1c1);
  border-radius: 4px;
  transition: background-color 0.3s ease;
}

.tree-wrapper::-webkit-scrollbar-thumb:hover {
  background: var(--text-color-secondary, #a8a8a8);
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
  background-color: var(--bg-color-tertiary, #f0f9ff) !important;
  border-color: var(--border-color-light, #b3d8ff) !important;
  transition: all 0.2s ease;
}

.menu-permission-tree :deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: var(--bg-color-tertiary, #e1f3ff);
  border-color: var(--sidebar-active, #409eff);
}

.menu-permission-tree :deep(.el-tree-node__expand-icon) {
  color: var(--text-color-secondary, #909399);
  font-size: 14px;
  transition: color 0.2s ease;
}

.menu-permission-tree :deep(.el-tree-node__expand-icon:hover) {
  color: var(--sidebar-active, #409eff);
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

.protected-role-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #f0f9ff;
  border: 1px solid #b3d8ff;
  border-radius: 4px;
  color: #409eff;
  font-size: 13px;
}

.protected-role-tip .el-icon {
  font-size: 16px;
  color: #409eff;
}
</style>

<style>
/* 角色管理页面菜单与权限树夜间模式适配 */
.dark-mode .menu-permission-container {
  background-color: var(--card-bg) !important;
  border-color: var(--border-color-light) !important;
}

.dark-mode .menu-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  color: #fff !important;
}

.dark-mode .header-title {
  color: #fff !important;
}

.dark-mode .tree-wrapper {
  background-color: var(--card-bg) !important;
}

.dark-mode .menu-name,
.dark-mode .permission-name {
  color: var(--text-color-primary) !important;
}

.dark-mode .menu-permission-tree :deep(.el-tree-node__label) {
  color: var(--text-color-primary) !important;
}

.dark-mode .menu-permission-tree :deep(.el-tree-node__content) {
  color: var(--text-color-primary) !important;
}

.dark-mode .menu-permission-tree :deep(.el-tree-node__content:hover) {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--border-color-light) !important;
}

.dark-mode .menu-permission-tree :deep(.el-tree-node__expand-icon) {
  color: var(--text-color-secondary) !important;
}

.dark-mode .menu-permission-tree :deep(.el-tree-node__expand-icon:hover) {
  color: var(--sidebar-active) !important;
}

.dark-mode .permission-path-icon {
  color: var(--text-color-secondary) !important;
}

.dark-mode .permission-path-icon:hover {
  color: var(--sidebar-active) !important;
}

.dark-mode .protected-tip {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--border-color-light) !important;
  color: var(--text-color-primary) !important;
}

.dark-mode .protected-role-tip {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--border-color-light) !important;
  color: var(--text-color-primary) !important;
}

.dark-mode .protected-role-tip .el-icon {
  color: var(--sidebar-active) !important;
}

.dark-mode .menu-permission-tree :deep(.el-checkbox__label) {
  color: var(--text-color-primary) !important;
}

.dark-mode .menu-permission-tree :deep(.el-tree-node) {
  color: var(--text-color-primary) !important;
}

.dark-mode .menu-permission-tree :deep(.el-tree-node *) {
  color: var(--text-color-primary) !important;
}
</style>
