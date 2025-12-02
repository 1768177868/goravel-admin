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
        :initial-values="{ name: '', slug: '', method: '', path: '', status: '', menu_id: '' }"
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
import { useListPage } from '../../composables/useListPage'
import { getMenuTitle as getMenuTitleUtil } from '../../utils/menuTranslation'
import {
  getPermissionList,
  getPermissionDetail,
  createPermission,
  updatePermission,
  deletePermission
} from '../../api/permission'
import { getMenuList } from '../../api/menu'

const { t, te } = useI18n()
const formRef = ref(null)
const tableRef = ref(null)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('permission.edit_permission') : t('permission.add_permission'))

// 字段名映射
const fieldMapping = {
  'id': 'id',
  'name': 'name',
  'slug': 'slug',
  'method': 'method',
  'path': 'path',
  'status': 'status',
  'sort': 'sort',
  'created_at': 'created_at'
}

// 使用列表页面 composable
const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handlePageChange,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getPermissionList,
  initialSearchForm: {
    name: '',
    slug: '',
    method: '',
    path: '',
    status: '',
    menu_id: ''
  },
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  }
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
const searchFields = computed(() => {
  const fields = [
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
    },
    {
      prop: 'menu_id',
      label: t('menu.title'),
      type: 'tree-select',
      width: '200px',
      filterable: true,
      apiUrl: '/options?type=menu',
      treeProps: {
        label: 'label',
        value: 'value',
        children: 'children'
      },
      advanced: false
    }
  ]
  return fields
})

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

// 获取菜单标题（使用工具函数，自动从 slug 或路径提取翻译）
const getMenuTitle = (menu) => {
  if (!menu || typeof menu !== 'object') {
    return '-'
  }
  
  const translated = getMenuTitleUtil(t, te, menu)
  return translated || '-'
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

// loadData, handleSearch, handleReset, handlePageChange 已由 useListPage 提供

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

