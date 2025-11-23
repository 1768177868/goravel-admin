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
        <el-form-item :label="$t('role.menus_and_permissions')">
          <div class="menu-permission-container">
            <div class="menu-header">
              <el-button size="small" @click="handleSelectAllMenusAndPermissions">{{ $t('role.select_all') }}</el-button>
              <el-button size="small" @click="handleUnselectAllMenusAndPermissions">{{ $t('role.unselect_all') }}</el-button>
            </div>
            <el-tree
              ref="menuPermissionTreeRef"
              :data="menuPermissionTree"
              :props="{ children: 'children', label: 'label' }"
              show-checkbox
              node-key="id"
              :default-checked-keys="[...formData.menu_ids, ...formData.permission_ids]"
              class="menu-permission-tree"
            >
              <template #default="{ node, data }">
                <!-- 菜单节点 -->
                <span v-if="data.isMenu" class="menu-node">
                  <span class="menu-name">{{ data.name }}</span>
                  <el-tag v-if="data.type" size="small" :type="getMenuTypeTag(data.type)">
                    {{ getMenuTypeText(data.type) }}
                  </el-tag>
                </span>
                <!-- 权限节点 -->
                <span v-else class="permission-node">
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
import { InfoFilled } from '@element-plus/icons-vue'
import { getRoleList, getRoleDetail, createRole, updateRole, deleteRole } from '../../api/role'
import { getPermissionList } from '../../api/permission'
import { getMenuList } from '../../api/menu'

const { t } = useI18n()
const formRef = ref(null)
const permissionTreeRef = ref(null)
const menuTreeRef = ref(null)
const menuPermissionTreeRef = ref(null)
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
const menuPermissionTree = ref([])

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

// 从路径中提取友好的模块名称
const getModuleNameFromPath = (path) => {
  if (!path) return t('role.other_module')
  
  // 移除查询参数和通配符
  let cleanPath = path.split('?')[0].replace(/\*/g, '').replace(/\/$/, '')
  
  // 处理路径中的ID参数（如 /api/admin/admins/123 -> /api/admin/admins）
  cleanPath = cleanPath.replace(/\/\d+(\/|$)/g, '/')
  cleanPath = cleanPath.replace(/\/$/, '')
  
  // 路径到模块名称的映射（支持中英文）
  const pathMap = {
    '/api/admin/admins': t('role.module_admin'),
    '/api/admin/roles': t('role.module_role'),
    '/api/admin/permissions': t('role.module_permission'),
    '/api/admin/menus': t('role.module_menu'),
    '/api/admin/departments': t('role.module_department'),
    '/api/admin/dictionaries': t('role.module_dictionary'),
    '/api/admin/operation-logs': t('role.module_operation_log'),
    '/api/admin/login-logs': t('role.module_login_log'),
    '/api/admin/system-logs': t('role.module_system_log'),
    '/api/admin/auth': t('role.module_auth'),
    '/api/admin/profile': t('role.module_profile')
  }
  
  // 精确匹配
  if (pathMap[cleanPath]) {
    return pathMap[cleanPath]
  }
  
  // 模糊匹配：从路径中提取最后一个部分
  const parts = cleanPath.split('/').filter(p => p)
  if (parts.length >= 3) {
    const module = parts[parts.length - 1]
    // 将复数形式转换为单数
    const singular = module.replace(/s$/, '').replace(/-/g, '_')
    // 尝试从翻译中获取
    const translationKey = `role.module_${singular}`
    const translated = t(translationKey)
    if (translated !== translationKey) {
      return translated
    }
  }
  
  return t('role.other_module')
}

// 转换权限数据为树形结构（按模块分组）
const transformPermissionToTree = (permissions) => {
  if (!permissions || !Array.isArray(permissions)) return []
  
  // 按模块分组
  const moduleGroups = {}
  permissions.forEach(perm => {
    const path = perm.Path || perm.path || '/'
    const method = perm.Method || perm.method || ''
    const name = perm.Name || perm.name || ''
    const slug = perm.Slug || perm.slug || ''
    const description = perm.Description || perm.description || ''
    const id = perm.id || perm.ID
    
    // 获取模块名称
    const moduleName = getModuleNameFromPath(path)
    
    if (!moduleGroups[moduleName]) {
      moduleGroups[moduleName] = {
        id: `module_${moduleName}`,
        name: moduleName,
        label: moduleName,
        children: []
      }
    }
    
    // 构建友好的显示标签
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
  
  // 转换为数组并按模块名称排序
  const tree = Object.values(moduleGroups).sort((a, b) => {
    return a.name.localeCompare(b.name)
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

// 将权限挂载到菜单树中（根据 menu_id 关联）
const attachPermissionsToMenus = (menuTree, permissions) => {
  if (!permissions || !Array.isArray(permissions)) return menuTree
  
  // 创建权限ID到权限的映射
  const permissionMap = new Map()
  permissions.forEach(perm => {
    const id = perm.id || perm.ID
    const menuId = perm.MenuID || perm.menu_id || 0
    if (!permissionMap.has(menuId)) {
      permissionMap.set(menuId, [])
    }
    permissionMap.get(menuId).push(perm)
  })
  
  // 递归处理菜单节点
  const processNode = (node) => {
    const result = { ...node }
    
    // 如果是菜单节点，查找关联的权限
    if (result.isMenu && result.id) {
      const menuId = result.id
      const matchedPermissions = permissionMap.get(menuId) || []
      
      // 将权限转换为树节点格式
      if (matchedPermissions.length > 0) {
        if (!result.children) {
          result.children = []
        }
        
        matchedPermissions.forEach(perm => {
          const method = perm.Method || perm.method || ''
          const name = perm.Name || perm.name || ''
          const description = perm.Description || perm.description || ''
          const id = perm.id || perm.ID
          
          result.children.push({
            id: id,
            name: name,
            slug: perm.Slug || perm.slug || '',
            method: method,
            path: perm.Path || perm.path || '',
            description: description,
            label: description || name,
            displayDesc: description || name,
            isMenu: false,
            isPermission: true
          })
        })
        
        // 对权限按方法排序
        result.children.sort((a, b) => {
          if (a.isMenu !== b.isMenu) {
            return a.isMenu ? -1 : 1 // 菜单在前，权限在后
          }
          if (!a.isMenu && !b.isMenu) {
            const methodOrder = { 'GET': 1, 'POST': 2, 'PUT': 3, 'PATCH': 4, 'DELETE': 5 }
            return (methodOrder[a.method] || 99) - (methodOrder[b.method] || 99)
          }
          return 0
        })
      }
    }
    
    // 处理子节点
    if (result.children && Array.isArray(result.children)) {
      result.children = result.children.map(child => processNode(child))
    }
    
    return result
  }
  
  return menuTree.map(node => processNode(node))
}

// 构建菜单和权限的合并树
const buildMenuPermissionTree = (menus, permissions) => {
  const menuTreeData = transformMenuToTree(menus)
  const treeWithPermissions = attachPermissionsToMenus(menuTreeData, permissions)
  
  // 找出未匹配的权限
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
  
  // 如果有未匹配的权限，创建一个"其他权限"节点
  if (unmatchedPermissions.length > 0) {
    const otherPermissionsNode = {
      id: 'other_permissions',
      name: t('role.other_permissions'),
      label: t('role.other_permissions'),
      isMenu: true,
      type: 1,
      children: unmatchedPermissions.map(perm => {
        const method = perm.Method || perm.method || ''
        const name = perm.Name || perm.name || ''
        const description = perm.Description || perm.description || ''
        const id = perm.id || perm.ID
        
        return {
          id: id,
          name: name,
          slug: perm.Slug || perm.slug || '',
          method: method,
          path: perm.Path || perm.path || '',
          description: description,
          label: description || name,
          displayDesc: description || name,
          isMenu: false,
          isPermission: true
        }
      })
    }
    
    treeWithPermissions.push(otherPermissionsNode)
  }
  
  return treeWithPermissions
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
      // 同时更新合并树
      if (menuTree.value.length > 0) {
        menuPermissionTree.value = buildMenuPermissionTree(menuTree.value, res.data.list)
      }
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
    // 同时更新合并树
    if (permissionTree.value.length > 0 || (res.data?.list && res.data.list.length > 0)) {
      const permissions = permissionTree.value.length > 0 
        ? permissionTree.value.flatMap(g => g.children || [])
        : (await getPermissionList({ page_size: 1000 })).data?.list || []
      menuPermissionTree.value = buildMenuPermissionTree(menus, permissions)
    }
  } catch (error) {
    console.error('Load menus error:', error)
  }
}

// 加载菜单和权限的合并树
const loadMenuPermissionTree = async () => {
  try {
    const [menuRes, permissionRes] = await Promise.all([
      getMenuList(),
      getPermissionList({ page_size: 1000 })
    ])
    
    const menus = menuRes.data?.menus || menuRes.data?.list || []
    const permissions = permissionRes.data?.list || []
    
    // 保存原始数据用于其他用途
    menuTree.value = transformMenuToTree(menus)
    permissionTree.value = transformPermissionToTree(permissions)
    
    // 构建合并树（使用原始权限数据）
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
        if (menuPermissionTreeRef.value) {
          menuPermissionTreeRef.value.setCheckedKeys([...formData.menu_ids, ...formData.permission_ids])
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
        // 从合并树中分离菜单ID和权限ID
        const allCheckedKeys = menuPermissionTreeRef.value?.getCheckedKeys() || []
        const menuIds = []
        const permissionIds = []
        
        // 递归收集所有菜单ID和权限ID
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
        
        const data = {
          ...formData,
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
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

// 全选/取消全选菜单和权限
const handleSelectAllMenusAndPermissions = () => {
  if (menuPermissionTreeRef.value && menuPermissionTree.value) {
    const allKeys = getAllMenuAndPermissionKeys(menuPermissionTree.value)
    menuPermissionTreeRef.value.setCheckedKeys(allKeys)
  }
}

const handleUnselectAllMenusAndPermissions = () => {
  if (menuPermissionTreeRef.value) {
    menuPermissionTreeRef.value.setCheckedKeys([])
  }
}

// 递归获取所有菜单和权限ID
const getAllMenuAndPermissionKeys = (tree) => {
  if (!tree || !Array.isArray(tree)) return []
  const keys = []
  const traverse = (nodes) => {
    if (!nodes || !Array.isArray(nodes)) return
    nodes.forEach(node => {
      if (node.id) {
        keys.push(node.id)
      }
      if (node.children) {
        traverse(node.children)
      }
    })
  }
  traverse(tree)
  return keys
}

// 递归获取所有权限ID（包括子节点）
const getAllPermissionKeys = (tree) => {
  if (!tree || !Array.isArray(tree)) {
    return []
  }
  const keys = []
  const traverse = (nodes) => {
    if (!nodes || !Array.isArray(nodes)) {
      return
    }
    nodes.forEach(node => {
      // 只添加叶子节点（实际权限），不添加分组节点
      if (!node.children || node.children.length === 0) {
        if (node.id) {
          keys.push(node.id)
        }
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
  if (!tree || !Array.isArray(tree)) {
    return []
  }
  const keys = []
  const traverse = (nodes) => {
    if (!nodes || !Array.isArray(nodes)) {
      return
    }
    nodes.forEach(node => {
      if (node.id) {
        keys.push(node.id)
      }
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

.search-form {
  margin-bottom: 20px;
}

.menu-permission-container {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 10px;
  max-height: 500px;
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

.permission-path-icon {
  color: #909399;
  font-size: 14px;
  cursor: help;
  margin-left: 4px;
}

.permission-path-icon:hover {
  color: #409eff;
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

