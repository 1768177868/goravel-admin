<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.dictionary') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('dictionary.store').disabled"
            @click="handleAdd"
          >
            <el-icon><Plus /></el-icon>
            {{ $t('dictionary.add_dictionary') }}
          </el-button>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="dictionary"
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

    <DictionaryForm
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import DictionaryForm from './DictionaryForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import {
  getDictionaryList,
  deleteDictionary
} from '../../api/dictionary'

const { t } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)

const {
  dialogVisible,
  editId,
  handleAdd,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: deleteDictionary
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
    field: 'type',
    title: t('dictionary.type'),
    sortable: false,
    formatter: ({ row }) => row.Type || row.type || '-'
  },
  {
    field: 'label',
    title: t('dictionary.label'),
    sortable: false,
    formatter: ({ row }) => row.Label || row.label || '-'
  },
  {
    field: 'value',
    title: t('dictionary.value'),
    sortable: false,
    formatter: ({ row }) => row.Value || row.value || '-'
  },
  {
    field: 'sort',
    title: t('common.sort'),
    width: 80,
    sortable: true,
    formatter: ({ row }) => row.Sort !== undefined ? row.Sort : (row.sort !== undefined ? row.sort : 0)
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 80,
    sortable: false,
    slot: 'status'
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
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
    prop: 'type',
    label: t('dictionary.type'),
    type: 'input',
    width: '200px',
    advanced: false
  }
])

// 初始搜索表单（用于 SearchForm 的 initial-values）
const initialSearchForm = {
  type: ''
}

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
  fetchApi: getDictionaryList,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'id:desc',
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

// 操作按钮配置
const operationActions = computed(() => [
  {
    key: 'edit',
    label: t('common.edit'),
    type: 'primary',
    permission: 'dictionary.update',
    handler: handleEdit
  },
  {
    key: 'delete',
    label: t('common.delete'),
    type: 'danger',
    permission: 'dictionary.destroy',
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

