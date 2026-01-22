<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.role') }}</span>
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
        :initial-values="initialSearchForm"
        i18n-prefix="role"
        @search="handleSearch"
        @reset="handleReset"
      />

      <VxeTable
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="600"
        @sort-change="handleSortChange"
      >
        <template #status="{ row }">
          <el-switch
            :model-value="Number(row.status ?? row.Status ?? 1) === 1"
            :disabled="isProtectedRole(row) || getButtonState('role.update').disabled"
            @change="(val) => handleStatusChange(row, val)"
          />
        </template>

        <template #operation="{ row }">
          <TableActionButtons
            :row="row"
            :primary-actions="getPrimaryActions(row)"
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

    <!-- 添加/编辑对话框 -->
    <RoleForm
      ref="roleFormRef"
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import RoleForm from './RoleForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import { getRoleList, deleteRole, updateRole } from '../../api/role'

// 使用 markRaw 标记图标组件，避免被 Vue 做成响应式对象
const PlusIcon = markRaw(Plus)

const { t } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)
const roleFormRef = ref(null)

// 使用 CRUD composable（只用删除功能，因为角色有自定义对话框）
const { handleDelete: handleDeleteCrud } = useCrud({
  deleteApi: deleteRole
})

const dialogVisible = ref(false)
const editId = ref(null)

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
    sortable: false,
    formatter: ({ row }) => getDisplayValue(row, 'name')
  },
  {
    field: 'slug',
    title: t('role.slug'),
    sortable: false,
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
    sortable: false,
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

const protectedRoleSlugs = ref(['super-admin'])

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
  fetchApi: getRoleList,
  initialSearchForm: {
    name: '',
    status: ''
  },
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef)
})

// 初始搜索表单（用于 SearchForm 的 initial-values）
const initialSearchForm = {
  name: '',
  status: ''
}

const handleAdd = () => {
  if (dialogVisible.value) {
    dialogVisible.value = false
    setTimeout(() => {
      editId.value = null
      dialogVisible.value = true
    }, 200)
  } else {
    editId.value = null
    dialogVisible.value = true
  }
}

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleFormSuccess = () => {
  loadData()
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

const handleDelete = (row) => {
  if (isProtectedRole(row)) {
    ElMessage.warning(t('role.protected_cannot_delete'))
    return
  }
  handleDeleteCrud(row, loadData)
}

// 获取主要操作按钮配置
const getPrimaryActions = (row) => {
  return [
    {
      key: 'edit',
      label: t('common.edit'),
      type: 'primary',
      permission: 'role.update',
      handler: handleEdit
    },
    {
      key: 'delete',
      label: t('common.delete'),
      type: 'danger',
      permission: 'role.destroy',
      show: () => !isProtectedRole(row),
      handler: handleDelete
    }
  ]
}

onMounted(async () => {
  try {
    initDefaultSort()
    await loadData()
  } catch (error) {
    console.error('RoleList onMounted error:', error)
    ElMessage.error(t('error.page_load_failed'))
  }
})
</script>

<style scoped>
</style>