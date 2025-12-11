<template>
  <div class="dictionary-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('dictionary.title') }}</span>
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
        :initial-values="{ type: '' }"
        i18n-prefix="dictionary"
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
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <TableActionButtons
                :row="row"
                :primary-actions="[
                  {
                    key: 'edit',
                    label: $t('common.edit'),
                    type: 'primary',
                    permission: 'dictionary.update',
                    handler: handleEdit
                  },
                  {
                    key: 'delete',
                    label: $t('common.delete'),
                    type: 'danger',
                    permission: 'dictionary.destroy',
                    handler: handleDelete
                  }
                ]"
                :get-button-state="getButtonState"
              />
            </template>
          </vxe-column>
        </template>
      </vxe-table>

      <Pagination
        v-model="pagination"
        @page-change="handlePageChange"
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
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import DictionaryForm from './DictionaryForm.vue'
import { useTableSort } from '../../composables/useTableSort'
import { usePermission } from '../../composables/usePermission'
import {
  getDictionaryList,
  deleteDictionary
} from '../../api/dictionary'

const { t } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)
const loading = ref(false)
const dialogVisible = ref(false)
const editId = ref(null)

const searchForm = reactive({
  type: ''
})

// 表格列配置（使用 vxe-table columns）
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
    sortable: true,
    formatter: ({ row }) => row.Type || row.type || '-'
  },
  {
    field: 'label',
    title: t('dictionary.label'),
    sortable: true,
    formatter: ({ row }) => row.Label || row.label || '-'
  },
  {
    field: 'value',
    title: t('dictionary.value'),
    sortable: true,
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
    sortable: true,
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

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'type': 'type',
  'label': 'label',
  'value': 'value',
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
    if (searchForm.type && searchForm.type.trim()) {
      params.type = searchForm.type.trim()
    }
    
    const res = await getDictionaryList(params)
    
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load dictionary list error:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.type = ''
  resetSort()
  handleSearch()
}

const handlePageChange = ({ currentPage, pageSize }) => {
  pagination.page = currentPage
  pagination.pageSize = pageSize
  loadData()
}

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
    await ElMessageBox.confirm(t('dictionary.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteDictionary(row.id)
    ElMessage.success(t('dictionary.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

onMounted(() => {
  initDefaultSort()
  loadData()
})
</script>

<style scoped>
.dictionary-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

</style>

