<template>
  <div class="permission-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('permission.title') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('permission.store').disabled"
            @click="handleAdd"
          >
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
              <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'success' : 'danger'">
                {{ (row.Status ?? row.status ?? 1) === 1 ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
            </template>
            <template v-else-if="column.slot === 'menu'" #default="{ row }">
              <span>{{ getMenuDisplayTitle(row.Menu || row.menu) }}</span>
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <el-button 
                type="primary" 
                link 
                :disabled="getButtonState('permission.update').disabled"
                @click="handleEdit(row)"
              >
                {{ $t('common.edit') }}
              </el-button>
              <el-button 
                type="danger" 
                link 
                :disabled="getButtonState('permission.destroy').disabled"
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

    <PermissionForm
      v-model="dialogVisible"
      :edit-id="editId"
      :menu-tree-data="menuTreeData"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import PermissionForm from './PermissionForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { getMenuTitle as getMenuTitleUtil } from '../../utils/menuTranslation'
import {
  getPermissionList,
  deletePermission
} from '../../api/permission'
import { getMenuList } from '../../api/menu'

const { t, te } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)
const dialogVisible = ref(false)
const editId = ref(null)

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

const menuTreeData = ref([])

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

// loadData, handleSearch, handleReset, handlePageChange 已由 useListPage 提供

const handleAdd = () => {
  editId.value = null
  dialogVisible.value = true
}

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleFormSuccess = () => {
  loadData()
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
</style>

