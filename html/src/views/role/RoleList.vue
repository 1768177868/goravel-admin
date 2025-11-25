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
            <el-tag :type="Number(row.status) === 1 ? 'success' : 'danger'">
              {{ Number(row.status) === 1 ? $t('common.enabled') : $t('common.disabled') }}
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
              <div class="header-left">
                <el-icon class="header-icon"><Menu /></el-icon>
                <span class="header-title">{{ $t('role.menus_and_permissions') }}</span>
              </div>
              <div class="header-actions">
                <el-button size="small" type="primary" plain @click="handleSelectAllMenusAndPermissions">
                  <el-icon><Check /></el-icon>
                  {{ isAllSelected ? $t('role.unselect_all') : $t('role.select_all') }}
                </el-button>
                <el-button size="small" plain @click="handleUnselectAllMenusAndPermissions">
                  <el-icon><Close /></el-icon>
                  {{ $t('role.unselect_all') }}
                </el-button>
              </div>
            </div>
            <div class="tree-wrapper">
              <el-tree
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
                check-strictly
              >
                <template #default="{ node, data }">
                  <!-- 菜单节点 -->
                  <span v-if="data.isMenu" class="menu-node">
                    <el-icon class="node-icon menu-icon"><FolderOpened /></el-icon>
                    <span class="menu-name">{{ data.name }}</span>
                    <el-tag v-if="data.type" size="small" :type="getMenuTypeTag(data.type)" class="menu-type-tag">
                      {{ getMenuTypeText(data.type) }}
                    </el-tag>
                  </span>
                  <!-- 权限节点 -->
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
import { InfoFilled, Menu, Check, Close, FolderOpened, Key } from '@element-plus/icons-vue'
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
const isAllSelected = ref(false)
const checkedKeys = ref([])

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
      const roles = res.data.list || []
      // 转换数据格式，确保状态值正确
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
  // 清空选中状态
  checkedKeys.value = []
  isAllSelected.value = false
  dialogVisible.value = true
  // 等待树组件渲染后确保清空选中状态
  setTimeout(() => {
    if (menuPermissionTreeRef.value) {
      menuPermissionTreeRef.value.setCheckedKeys([])
      checkedKeys.value = []
      isAllSelected.value = false
    }
  }, 100)
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
        status: Number(role.Status !== undefined ? role.Status : (role.status !== undefined ? role.status : 1)),
        sort: role.Sort !== undefined ? role.Sort : (role.sort !== undefined ? role.sort : 0)
      })
      // 设置选中状态
      const allCheckedKeys = [...formData.menu_ids, ...formData.permission_ids]
      checkedKeys.value = allCheckedKeys
      dialogVisible.value = true
      // 等待树组件渲染后设置选中状态
      setTimeout(() => {
        if (menuPermissionTreeRef.value) {
          menuPermissionTreeRef.value.setCheckedKeys(allCheckedKeys, false)
          // 延迟检查全选状态，确保树组件状态已更新
          setTimeout(() => {
            checkAllSelected()
          }, 50)
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
        
        // 确保状态值正确：0 表示禁用，1 表示启用
        const statusValue = formData.status !== undefined && formData.status !== null 
          ? Number(formData.status) 
          : 1 // 默认启用
        
        console.log('Form data status:', formData.status, 'Status value:', statusValue)
        
        const data = {
          name: formData.name,
          slug: formData.slug,
          description: formData.description || '',
          status: statusValue,
          sort: Number(formData.sort) || 0,
          permission_ids: permissionIds,
          menu_ids: menuIds
        }
        
        console.log('Submit data:', JSON.stringify(data, null, 2))
        
        if (formData.id) {
          await updateRole(formData.id, data)
          ElMessage.success(t('role.update_success'))
        } else {
          const res = await createRole(data)
          console.log('Create role response:', res)
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
  // 清空树形选择的选中状态
  checkedKeys.value = []
  if (menuPermissionTreeRef.value) {
    menuPermissionTreeRef.value.setCheckedKeys([])
  }
  isAllSelected.value = false
  formRef.value?.resetFields()
}

// 检查是否全选
const checkAllSelected = () => {
  if (menuPermissionTreeRef.value && menuPermissionTree.value) {
    const allKeys = getAllMenuAndPermissionKeys(menuPermissionTree.value)
    const currentCheckedKeys = menuPermissionTreeRef.value.getCheckedKeys() || []
    isAllSelected.value = allKeys.length > 0 && 
      allKeys.every(key => currentCheckedKeys.includes(key)) &&
      currentCheckedKeys.length === allKeys.length
  } else {
    isAllSelected.value = false
  }
}

// 全选/取消全选菜单和权限
const handleSelectAllMenusAndPermissions = () => {
  if (menuPermissionTreeRef.value && menuPermissionTree.value) {
    const allKeys = getAllMenuAndPermissionKeys(menuPermissionTree.value)
    console.log('全选 - 所有节点ID:', allKeys, '数量:', allKeys.length)
    
    if (allKeys.length === 0) {
      ElMessage.warning('没有可选的菜单或权限')
      return
    }
    
    const currentCheckedKeys = menuPermissionTreeRef.value.getCheckedKeys() || []
    console.log('全选 - 当前已选中:', currentCheckedKeys, '数量:', currentCheckedKeys.length)
    
    // 检查是否已经全选（需要精确匹配，包括数量）
    const allSelected = allKeys.length > 0 && 
      allKeys.length === currentCheckedKeys.length &&
      allKeys.every(key => currentCheckedKeys.includes(key))
    
    if (allSelected) {
      // 如果已全选，则取消全选
      checkedKeys.value = []
      menuPermissionTreeRef.value.setCheckedKeys([], false)
      isAllSelected.value = false
    } else {
      // 否则全选所有（包括所有节点，无论之前是否选中）
      checkedKeys.value = allKeys
      
      // 使用 check-strictly 模式，直接设置所有节点（包括深层节点）
      // 先设置一次
      menuPermissionTreeRef.value.setCheckedKeys(allKeys, false)
      
      // 延迟多次检查并补全，确保所有深层节点都被选中
      let retryCount = 0
      const maxRetries = 5
      const checkAndRetry = () => {
        setTimeout(() => {
          const afterCheckKeys = menuPermissionTreeRef.value.getCheckedKeys() || []
          const missingKeys = allKeys.filter(key => !afterCheckKeys.includes(key))
          
          console.log(`全选 - 第${retryCount + 1}次检查: 已选中${afterCheckKeys.length}个, 缺失${missingKeys.length}个`, missingKeys)
          
          if (missingKeys.length > 0 && retryCount < maxRetries) {
            // 如果有缺失的节点（特别是深层权限节点），再次设置
            const finalKeys = [...new Set([...afterCheckKeys, ...missingKeys])]
            menuPermissionTreeRef.value.setCheckedKeys(finalKeys, false)
            checkedKeys.value = finalKeys
            retryCount++
            // 继续重试
            checkAndRetry()
          } else {
            // 最终检查并更新状态
            const finalChecked = menuPermissionTreeRef.value.getCheckedKeys() || []
            console.log('全选 - 最终结果:', finalChecked.length, '个节点被选中')
            checkAllSelected()
          }
        }, 150)
      }
      checkAndRetry()
      isAllSelected.value = true
    }
  }
}

const handleUnselectAllMenusAndPermissions = () => {
  if (menuPermissionTreeRef.value) {
    // 强制清空所有选中项（包括半选状态）
    checkedKeys.value = []
    menuPermissionTreeRef.value.setCheckedKeys([], false)
    isAllSelected.value = false
  }
}

// 处理树节点勾选变化
const handleTreeCheck = (data, checkedInfo) => {
  // 更新选中键数组（只获取完全选中的节点，不包括半选状态）
  if (menuPermissionTreeRef.value) {
    const checkedKeysList = menuPermissionTreeRef.value.getCheckedKeys() || []
    // 只使用完全选中的节点
    checkedKeys.value = checkedKeysList
  }
  // 延迟检查，确保状态已更新
  setTimeout(() => {
    checkAllSelected()
  }, 50)
}

// 递归获取所有菜单和权限ID（包括所有层级的节点，特别是深层权限节点）
const getAllMenuAndPermissionKeys = (tree) => {
  if (!tree || !Array.isArray(tree)) return []
  const keys = []
  const visited = new Set() // 防止重复添加
  
  const traverse = (nodes, depth = 0) => {
    if (!nodes || !Array.isArray(nodes)) return
    nodes.forEach(node => {
      // 确保节点有 id，并且不是分组节点（如 'other_permissions' 或 'module_xxx'）
      if (node.id && !visited.has(node.id)) {
        // 如果是数字类型，直接添加
        if (typeof node.id === 'number') {
          keys.push(node.id)
          visited.add(node.id)
        } 
        // 如果是字符串类型，排除分组节点
        else if (typeof node.id === 'string' && 
                 !node.id.startsWith('module_') && 
                 node.id !== 'other_permissions') {
          keys.push(node.id)
          visited.add(node.id)
        }
      }
      // 递归处理子节点（确保能获取到最深层的权限节点）
      if (node.children && Array.isArray(node.children) && node.children.length > 0) {
        traverse(node.children, depth + 1)
      }
    })
  }
  traverse(tree)
  console.log('获取到的所有节点ID:', keys, '总数:', keys.length)
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

.header-actions {
  display: flex;
  gap: 6px;
}

.header-actions .el-button {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.3);
  color: #fff;
  backdrop-filter: blur(10px);
}

.header-actions .el-button:hover {
  background: rgba(255, 255, 255, 0.3);
  border-color: rgba(255, 255, 255, 0.4);
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

.menu-path {
  color: #909399;
  font-size: 12px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
}
</style>


