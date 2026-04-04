<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.position') }}</span>
          <el-button
            type="primary"
            :disabled="getButtonState('position.store').disabled"
            @click="handleAdd"
          >
            <el-icon><Plus /></el-icon>
            {{ $t('position.add_position') }}
          </el-button>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="position"
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
          <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'success' : 'danger'">
            {{ (row.Status ?? row.status ?? 1) === 1 ? $t('common.enabled') : $t('common.disabled') }}
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

    <PositionForm
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import PositionForm from './PositionForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import { Plus } from '@element-plus/icons-vue'
import { getPositionList, deletePosition } from '../../api/position'

const { t } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)

const {
  dialogVisible,
  editId,
  handleAdd,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: deletePosition
})

const tableColumns = computed(() => [
  { field: 'id', title: t('table.id'), width: 80, sortable: true },
  {
    field: 'name',
    title: t('position.name'),
    sortable: false,
    formatter: ({ row }) => row.Name || row.name || '-'
  },
  {
    field: 'code',
    title: t('position.code'),
    sortable: false,
    formatter: ({ row }) => row.Code || row.code || '-'
  },
  {
    field: 'sort',
    title: t('common.sort'),
    width: 80,
    sortable: true,
    formatter: ({ row }) => row.Sort !== undefined ? row.Sort : (row.sort !== undefined ? row.sort : 0)
  },
  { field: 'status', title: t('table.status'), width: 80, sortable: false, slot: 'status' },
  { field: 'created_at', title: t('table.created_at'), sortable: true },
  { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', sortable: false }
])

const searchFields = computed(() => [
  {
    prop: 'name',
    label: t('position.name'),
    type: 'input',
    width: '200px',
    advanced: false
  }
])

const initialSearchForm = { name: '' }

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
  fetchApi: getPositionList,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'sort:asc',
  tableRef: computed(() => tableRef.value?.tableRef)
})

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleFormSuccess = () => {
  loadData()
}

const handleDelete = (row) => handleDeleteCrud(row, loadData)

const operationActions = computed(() => [
  {
    key: 'edit',
    label: t('common.edit'),
    type: 'primary',
    permission: 'position.update',
    handler: handleEdit
  },
  {
    key: 'delete',
    label: t('common.delete'),
    type: 'danger',
    permission: 'position.destroy',
    handler: handleDelete
  }
])

onMounted(() => {
  initDefaultSort()
  loadData()
})
</script>

<style scoped>
</style>
