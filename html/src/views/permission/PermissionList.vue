<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.permission') }}</span>
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
        :initial-values="initialSearchForm"
        i18n-prefix="permission"
        @search="handleSearch"
        @reset="handleReset"
      />

      <!-- 表格工具栏 -->
      <TableToolbar
        :on-refresh="handleRefresh"
        fullscreen-target=".list-page"
        :visible-columns="visibleColumns"
        :all-columns="allTableColumns"
        :default-visible-columns="defaultVisibleColumns"
        :column-order="columnOrder"
        :fixed-columns="fixedColumns"
        :on-column-setting-confirm="handleColumnSettingConfirm"
      />

      <VxeTable
        ref="tableRef"
        :key="`table-${tableColumns.length}-${JSON.stringify(tableColumns.map(c => c.field || c.slot || c.key))}`"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="600"
        @sort-change="handleSortChange"
      >
        <template #status="{ row }">
          <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'success' : 'danger'">
            {{ (row.Status ?? row.status ?? 1) === 1 ? $t('common.enabled') : $t('common.disabled') }}
          </el-tag>
        </template>

        <template #menu="{ row }">
          <span>{{ getMenuDisplayTitle(row.Menu || row.menu) }}</span>
        </template>

        <template #operation="{ row }">
          <TableActionButtons
            :row="row"
            :primary-actions="operationActions"
            :get-button-state="getButtonState"
          />
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
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
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import TableToolbar from '../../components/TableToolbar.vue'
import PermissionForm from './PermissionForm.vue'
import { useColumnSetting } from '../../composables/useColumnSetting'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import { getMenuTitle as getMenuTitleUtil } from '../../utils/menuTranslation'
import {
  getPermissionList,
  deletePermission
} from '../../api/permission'
import { getMenuList } from '../../api/menu'

const { t, te } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)

const {
  dialogVisible,
  editId,
  handleAdd,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: deletePermission
})

const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
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
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef)
})

// 初始搜索表单（用于 SearchForm 的 initial-values）
const initialSearchForm = {
  name: '',
  slug: '',
  method: '',
  path: '',
  status: '',
  menu_id: ''
}

// 表格列配置
const allTableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true,
    key: 'id'
  },
  {
    field: 'name',
    title: t('permission.name'),
    sortable: false,
    formatter: ({ row }) => row.Name || row.name || '-',
    key: 'name'
  },
  {
    field: 'slug',
    title: t('permission.slug'),
    sortable: false,
    formatter: ({ row }) => row.Slug || row.slug || '-',
    key: 'slug'
  },
  {
    field: 'method',
    title: t('permission.method'),
    width: 100,
    sortable: false,
    formatter: ({ row }) => row.Method || row.method || '-',
    key: 'method'
  },
  {
    field: 'path',
    title: t('permission.path'),
    sortable: false,
    formatter: ({ row }) => row.Path || row.path || '-',
    key: 'path'
  },
  {
    field: 'description',
    title: t('common.description'),
    sortable: false,
    formatter: ({ row }) => row.Description || row.description || '-',
    key: 'description'
  },
  {
    field: 'menu',
    title: t('menu.title'),
    width: 150,
    slot: 'menu',
    sortable: false,
    key: 'menu'
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 80,
    sortable: false,
    slot: 'status',
    key: 'status'
  },
  {
    field: 'sort',
    title: t('common.sort'),
    width: 80,
    sortable: true,
    formatter: ({ row }) => row.Sort !== undefined ? row.Sort : (row.sort !== undefined ? row.sort : 0),
    key: 'sort'
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    sortable: true,
    formatter: ({ row }) => row.created_at || row.CreatedAt || '-',
    key: 'created_at'
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation',
    sortable: false,
    key: 'operation'
  }
])

// 使用列设置 composable
const {
  tableColumns,
  visibleColumns,
  allColumns,
  defaultVisibleColumns,
  columnOrder,
  fixedColumns,
  handleColumnSettingConfirm
} = useColumnSetting('permission', allTableColumns)

// 处理刷新
const handleRefresh = () => {
  loadData()
}

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

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleFormSuccess = () => {
  loadData()
}

const handleDelete = (row) => handleDeleteCrud(row, loadData)

// 操作按钮配置
const operationActions = computed(() => [
  {
    key: 'edit',
    label: t('common.edit'),
    type: 'primary',
    permission: 'permission.update',
    handler: handleEdit
  },
  {
    key: 'delete',
    label: t('common.delete'),
    type: 'danger',
    permission: 'permission.destroy',
    handler: handleDelete
  }
])

onMounted(() => {
  initDefaultSort()
  loadMenuList()
  loadData()
})
</script>

<style scoped>

</style>

