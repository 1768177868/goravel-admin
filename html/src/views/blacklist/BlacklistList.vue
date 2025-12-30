<template>
  <div class="blacklist-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.blacklist') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('blacklist.store').disabled"
            @click="handleAdd"
          >
            <el-icon><Plus /></el-icon>
            {{ $t('blacklist.add_blacklist') }}
          </el-button>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ ip: '', status: '' }"
        i18n-prefix="blacklist"
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
        <template #ip="{ row }">
          <div style="word-break: break-all;">
            {{ formatIP(row.IP || row.ip || '') }}
          </div>
        </template>

        <template #status="{ row }">
          <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'danger' : 'info'">
            {{ (row.Status ?? row.status ?? 1) === 1 ? $t('blacklist.enabled') : $t('blacklist.disabled') }}
          </el-tag>
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

    <BlacklistForm
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onActivated } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
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
const tableRef = ref(null)

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
  fetchApi: getBlacklistList,
  initialSearchForm: {
    ip: '',
    status: ''
  },
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef)
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

const handleDelete = (row) => {
  handleDeleteCrud(row, loadData)
}

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

// 操作按钮配置
const operationActions = computed(() => [
  {
    key: 'edit',
    label: t('common.edit'),
    type: 'primary',
    permission: 'blacklist.update',
    handler: handleEdit
  },
  {
    key: 'delete',
    label: t('common.delete'),
    type: 'danger',
    permission: 'blacklist.destroy',
    handler: handleDelete
  }
])

onMounted(() => {
  initDefaultSort()
  loadData()
})

onActivated(() => {
  loadData()
})
</script>

<style scoped>
.blacklist-list {
  background: white;
  border-radius: 4px;
}
</style>
