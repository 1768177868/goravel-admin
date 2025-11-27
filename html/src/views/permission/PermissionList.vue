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
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ name: '', slug: '', method: '', path: '', status: '' }"
        i18n-prefix="permission"
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
              <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'success' : 'danger'">
                {{ (row.Status ?? row.status ?? 1) === 1 ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
            </template>
            <template v-else-if="column.slot === 'menu'" #default="{ row }">
              <span>{{ getMenuDisplayTitle(row.Menu || row.menu) }}</span>
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <el-button type="primary" link @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
              <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
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
        <el-form-item :label="$t('menu.title')" prop="menu_id">
          <el-popover
            v-model:visible="menuSelectVisible"
            placement="bottom-start"
            :width="300"
            trigger="click"
          >
            <template #reference>
              <el-input
                :model-value="getSelectedMenuLabel()"
                :placeholder="$t('form.please_select') + $t('menu.title')"
                readonly
                clearable
                @clear="formData.menu_id = null"
                style="cursor: pointer"
              >
                <template #suffix>
                  <el-icon class="el-input__icon"><ArrowDown /></el-icon>
                </template>
              </el-input>
            </template>
            <el-tree
              :data="menuTreeData"
              :props="{ label: 'label', children: 'children' }"
              :default-expand-all="false"
              node-key="value"
              highlight-current
              :current-node-key="formData.menu_id"
              @node-click="handleMenuSelect"
            >
              <template #default="{ node, data }">
                <span class="tree-node-label">{{ data.label }}</span>
              </template>
            </el-tree>
          </el-popover>
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
import { Plus, ArrowDown } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { useTableSort } from '../../composables/useTableSort'
import {
  getPermissionList,
  getPermissionDetail,
  createPermission,
  updatePermission,
  deletePermission
} from '../../api/permission'
import { getMenuList } from '../../api/menu'

const { t } = useI18n()
const formRef = ref(null)
const tableRef = ref(null)
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
    title: t('permission.name'),
    sortable: true,
    formatter: ({ row }) => row.Name || row.name || '-'
  },
  {
    field: 'slug',
    title: t('permission.slug'),
    sortable: true,
    formatter: ({ row }) => row.Slug || row.slug || '-'
  },
  {
    field: 'method',
    title: t('permission.method'),
    width: 100,
    sortable: true,
    formatter: ({ row }) => row.Method || row.method || '-'
  },
  {
    field: 'path',
    title: t('permission.path'),
    sortable: true,
    formatter: ({ row }) => row.Path || row.path || '-'
  },
  {
    field: 'description',
    title: t('common.description'),
    sortable: false,
    formatter: ({ row }) => row.Description || row.description || '-'
  },
  {
    field: 'menu',
    title: t('menu.title'),
    width: 150,
    slot: 'menu',
    sortable: false
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
    formatter: ({ row }) => row.created_at || row.CreatedAt || '-'
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation',
    sortable: false
  }
])

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'name',
    label: t('permission.name'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'slug',
    label: t('permission.slug'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'method',
    label: t('permission.method'),
    type: 'select',
    width: '150px',
    options: [
      { label: 'GET', value: 'GET' },
      { label: 'POST', value: 'POST' },
      { label: 'PUT', value: 'PUT' },
      { label: 'DELETE', value: 'DELETE' },
      { label: 'PATCH', value: 'PATCH' }
    ],
    advanced: false
  },
  {
    prop: 'path',
    label: t('permission.path'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'select',
    width: '120px',
    options: [
      { label: t('common.enabled'), value: '1' },
      { label: t('common.disabled'), value: '0' }
    ],
    advanced: false
  }
])

const formData = reactive({
  id: null,
  name: '',
  slug: '',
  method: 'GET',
  path: '',
  description: '',
  menu_id: null,
  status: 1,
  sort: 0
})

const menuTreeData = ref([])
const menuSelectVisible = ref(false)

// 路径到翻译键的映射
const pathToTranslationKey = {
  '/system': 'menu.system_management',
  '/system/admin': 'menu.admin_management',
  '/system/role': 'menu.role_management',
  '/admins': 'menu.admin_management',
  '/roles': 'menu.role_management',
  '/permissions': 'menu.permission_management',
  '/menus': 'menu.menu_management',
  '/departments': 'menu.department_management',
  '/dictionaries': 'menu.dictionary_management',
  '/logs': 'menu.log_management',
  '/operation-logs': 'menu.operation_log',
  '/login-logs': 'menu.login_log',
  '/system-logs': 'menu.system_log',
  '/monitor': 'menu.service_monitor',
  '/profile': 'menu.profile',
  '/notifications': 'menu.notification_center'
}

// 标题到翻译键的映射（根据后端返回的中文标题）
const titleToTranslationKey = {
  '系统管理': 'menu.system_management',
  '管理员管理': 'menu.admin_management',
  '角色管理': 'menu.role_management',
  '权限管理': 'menu.permission_management',
  '菜单管理': 'menu.menu_management',
  '部门管理': 'menu.department_management',
  '字典管理': 'menu.dictionary_management',
  '日志管理': 'menu.log_management',
  '操作日志': 'menu.operation_log',
  '登录日志': 'menu.login_log',
  '系统日志': 'menu.system_log',
  '个人中心': 'menu.profile',
  '服务监控': 'menu.service_monitor',
  '通知中心': 'menu.notification_center'
}

// 获取菜单标题（优先使用 slug，如果没有则使用 path 和 title 映射，最后使用原始标题）
const getMenuTitle = (menu) => {
  if (!menu || typeof menu !== 'object') {
    return '-'
  }
  
  // 优先使用 slug 作为翻译键标识
  const slug = menu.Slug || menu.slug || ''
  if (slug) {
    const slugKey = `menu.${slug}`
    // 尝试使用 slug 查找翻译键，如果存在则使用
    try {
      const translated = t(slugKey)
      // 如果翻译结果不等于键名本身，说明找到了翻译
      if (translated !== slugKey) {
        return translated
      }
    } catch (e) {
      // 翻译键不存在，继续尝试其他方式
    }
  }
  
  // 尝试多种可能的字段名（支持 PascalCase 和 snake_case）
  const path = menu.Path || menu.path || ''
  const title = menu.Title || menu.title || ''
  
  // 回退到路径映射（向后兼容）
  if (path) {
    const pathKey = pathToTranslationKey[path]
    if (pathKey) {
      return t(pathKey)
    }
  }
  
  // 回退到标题映射（向后兼容）
  if (title) {
    const titleKey = titleToTranslationKey[title]
    if (titleKey) {
      return t(titleKey)
    }
    // 如果没有匹配的翻译键，返回原始标题
    return title
  }
  
  return '-'
}

// 转换菜单数据为树形选择器格式
const convertMenuToTreeData = (menus) => {
  return menus.map(menu => {
    const menuId = menu.id || menu.ID
    const title = getMenuTitle(menu)
    const path = menu.Path || menu.path || ''
    const label = path ? `${title} (${path})` : title
    
    const node = {
      value: menuId,
      label: label,
      title: title,
      path: path
    }
    
    // 递归处理子菜单
    const children = menu.Children || menu.children || []
    if (children.length > 0) {
      node.children = convertMenuToTreeData(children)
    }
    
    return node
  })
}

// 获取菜单列表
const loadMenuList = async () => {
  try {
    const { data } = await getMenuList()
    // 菜单返回的是树形结构，直接转换为树形选择器格式
    menuTreeData.value = convertMenuToTreeData(data.menus || [])
  } catch (error) {
    console.error('Load menu list failed:', error)
  }
}

// 获取菜单显示标题（用于表格显示）
const getMenuDisplayTitle = (menu) => {
  if (!menu) return '-'
  
  // 尝试多种可能的字段名
  const menuObj = menu.Menu || menu.menu || menu
  
  if (!menuObj || (typeof menuObj !== 'object')) {
    return '-'
  }
  
  return getMenuTitle(menuObj)
}

// 获取选中的菜单标签
const getSelectedMenuLabel = () => {
  if (!formData.menu_id) return ''
  const findMenu = (menus, id) => {
    for (const menu of menus) {
      if (menu.value === id) {
        return menu.label
      }
      if (menu.children && menu.children.length > 0) {
        const found = findMenu(menu.children, id)
        if (found) return found
      }
    }
    return ''
  }
  return findMenu(menuTreeData.value, formData.menu_id) || ''
}

// 处理菜单选择
const handleMenuSelect = (data) => {
  formData.menu_id = data.value
  menuSelectVisible.value = false
}

// 重置表单数据
const resetFormData = () => {
  formData.id = null
  formData.menu_id = null
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

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'name': 'name',
  'slug': 'slug',
  'method': 'method',
  'path': 'path',
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
      page_size: pagination.pageSize,
      order_by: buildOrderBy()
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
    
    const res = await getPermissionList(params)
    
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
  resetSort()
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
    const res = await getPermissionDetail(row.id)
    
    if (res.data && res.data.permission) {
      const permission = res.data.permission
      
      const mappedData = {
        id: permission.id || permission.ID,
        name: permission.Name || permission.name || '',
        slug: permission.Slug || permission.slug || '',
        method: permission.Method || permission.method || 'GET',
        path: permission.Path || permission.path || '',
        description: permission.Description || permission.description || '',
        menu_id: permission.MenuID !== undefined ? permission.MenuID : (permission.menu_id !== undefined ? permission.menu_id : null),
        status: permission.Status !== undefined ? permission.Status : (permission.status !== undefined ? permission.status : 1),
        sort: permission.Sort !== undefined ? permission.Sort : (permission.sort !== undefined ? permission.sort : 0)
      }
      
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
        // 准备提交数据，将 null 转换为 0
        const submitData = {
          ...formData,
          menu_id: formData.menu_id || 0
        }
        if (formData.id) {
          await updatePermission(formData.id, submitData)
          ElMessage.success(t('permission.update_success'))
        } else {
          await createPermission(submitData)
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
  initDefaultSort()
  loadMenuList()
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


.tree-node-label {
  font-size: 14px;
}
</style>

