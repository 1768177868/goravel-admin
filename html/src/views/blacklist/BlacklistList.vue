<template>
  <ListPage
    ref="listPageRef"
    page-class="blacklist"
    :title="$t('blacklist.title')"
    :show-add-button="true"
    :add-button-text="$t('blacklist.add_blacklist')"
    :add-button-disabled="getButtonState('blacklist.store').disabled"
    :search-form="searchForm"
    :search-fields="searchFields"
    :initial-search-values="{ ip: '', status: '' }"
    i18n-prefix="blacklist"
    :table-data="tableData"
    :loading="loading"
    :table-columns="tableColumns"
    :pagination="pagination"
    :dialog-visible="dialogVisible"
    :edit-id="editId"
    @add="handleAdd"
    @search="handleSearch"
    @reset="handleReset"
    @update:pagination="(val) => Object.assign(pagination, val)"
    @page-change="handlePageChange"
    @sort-change="handleSortChange"
    @form-success="handleFormSuccess"
  >
    <!-- 自定义表格列插槽 -->
    <template #status="{ row }">
      <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'danger' : 'info'">
        {{ (row.Status ?? row.status ?? 1) === 1 ? $t('blacklist.enabled') : $t('blacklist.disabled') }}
      </el-tag>
    </template>

    <template #ip="{ row }">
      <div style="word-break: break-all;">
        {{ formatIP(row.IP || row.ip || '') }}
      </div>
    </template>

    <template #operation="{ row }">
      <el-button 
        type="primary" 
        link 
        :disabled="getButtonState('blacklist.update').disabled"
        @click="handleEdit(row)"
      >
        {{ $t('common.edit') }}
      </el-button>
      <el-button 
        type="danger" 
        link 
        :disabled="getButtonState('blacklist.destroy').disabled"
        @click="handleDelete(row)"
      >
        {{ $t('common.delete') }}
      </el-button>
    </template>

    <!-- 表单对话框 -->
    <template #form>
      <BlacklistForm
        v-model="dialogVisible"
        :edit-id="editId"
        @success="handleFormSuccess"
      />
    </template>
  </ListPage>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus } from '@element-plus/icons-vue'
import ListPage from '../../components/ListPage.vue'
import BlacklistForm from './BlacklistForm.vue'
import { useListPage } from '../../composables/useListPage'
import { useCrud } from '../../composables/useCrud'
import { usePermission } from '../../composables/usePermission'
import {
  getBlacklistList,
  deleteBlacklist
} from '../../api/blacklist'

const { t } = useI18n()
const { getButtonState } = usePermission()
const listPageRef = ref(null)
const tableRef = computed(() => listPageRef.value?.tableRef)

// 字段名映射
const fieldMapping = {
  'id': 'id',
  'ip': 'ip',
  'remark': 'remark',
  'status': 'status',
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
  fetchApi: getBlacklistList,
  initialSearchForm: {
    ip: '',
    status: ''
  },
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  }
})

// 使用 CRUD composable
const {
  dialogVisible,
  editId,
  handleAdd,
  handleEdit,
  handleFormSuccess: handleFormSuccessCrud,
  handleDelete: handleDeleteCrud
} = useCrud({ deleteApi: deleteBlacklist })


const handleFormSuccess = () => {
  handleFormSuccessCrud(loadData)
}

// 表格列配置
const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },
  {
    field: 'ip',
    title: t('blacklist.ip'),
    sortable: false,
    slot: 'ip'
  },
  {
    field: 'remark',
    title: t('blacklist.remark'),
    sortable: false,
    formatter: ({ row }) => row.Remark || row.remark || '-'
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 100,
    sortable: false,
    slot: 'status'
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    width: 180,
    sortable: true
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
    prop: 'ip',
    label: t('blacklist.ip'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'select',
    width: '150px',
    advanced: false,
    options: [
      { label: t('blacklist.enabled'), value: '1' },
      { label: t('blacklist.disabled'), value: '0' }
    ]
  }
])

// 格式化IP显示
const formatIP = (ip) => {
  if (!ip) return '-'
  // 如果是IP范围，格式化显示
  if (ip.includes('-')) {
    const parts = ip.split('-')
    if (parts.length === 2) {
      return `${parts[0].trim()} ~ ${parts[1].trim()}`
    }
  }
  return ip
}

// 删除处理（包装 useCrud 的 handleDelete，传入 reloadData）
const handleDelete = (row) => {
  handleDeleteCrud(row, loadData)
}

onMounted(() => {
  initDefaultSort()
  loadData()
})
</script>

<style scoped>
.blacklist-list {
  background: white;
  border-radius: 4px;
}
</style>
