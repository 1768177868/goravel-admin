<template>
  <div class="<<.ModuleName>>-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.<<$.ModuleName>>') }}</span>
          <<if .HasCreate>>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            {{ $t('common.add') }}
          </el-button>
          <<end>>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="<<.ModuleName>>"
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
        <template #operation="{ row }">
          <<if .HasEdit>>
          <el-button type="primary" size="small" @click="handleEdit(row)">
            {{ $t('common.edit') }}
          </el-button>
          <<end>>
          <<if .HasDelete>>
          <el-button type="danger" size="small" @click="handleDelete(row)">
            {{ $t('common.delete') }}
          </el-button>
          <<end>>
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
      />

      <<print "<">><<.ModelName>>Form
        ref="formRef"
        v-model="dialogVisible"
        :edit-id="editId"
        @success="handleFormSuccess"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import <<.ModelName>>Form from './<<.ModelName>>Form.vue'
import { useListPage } from '../../composables/useListPage'
import { useCrud } from '../../composables/useCrud'
import {
  get<<.ModelName>>List,
  delete<<.ModelName>>
} from '../../api/<<.ModuleName>>'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

const { t } = useI18n()
const tableRef = ref(null)
const formRef = ref(null)

const {
  dialogVisible,
  editId,
  handleAdd,
  handleClose,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: delete<<.ModelName>>
})

const initialSearchForm = {
<<range .SearchableFields>>
  <<.Name>>: '',
<<- end>>
}

const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handleSortChange
} = useListPage({
  fetchApi: get<<.ModelName>>List,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef)
})

const searchFields = computed(() => [
<<range .SearchableFields>>
  {
    prop: '<<.Name>>',
    label: t('<<$.ModuleName>>.<<.Name>>'),
    type: '<<.SearchUIType>>',
<<if .ApiUrl>>    apiUrl: '<<.ApiUrl>>',<<end>>
<<if eq .SearchUIType "select">>
    <<if .Dictionary>>
    apiUrl: '/options?type=dictionary&dictionary_type=<<.Dictionary>>',
    <<else if .Relation>>
    apiUrl: '/<<.Relation.Table>>'.replace(/_/g, '-') + '?page=1&page_size=100',
    labelField: '<<.Relation.DisplayField>>',
    valueField: 'id',
    resultField: 'list', // 假设列表接口返回结构 { list: [] }
    <<end>>
<<end>>
    width: '200px',
    advanced: false
  },
<<- end>>
])

const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },
<<range .ListFields>>
  {
    field: '<<.Name>>',
    title: t('<<$.ModuleName>>.<<.Name>>'),
    sortable: <<.Sortable>>
  },
<<if .Relation>>
  {
    field: '<<.Relation.JsonName>>.<<.Relation.DisplayField>>',
    title: t('<<$.ModuleName>>.<<.Name>>'),
    sortable: false
  },
<<end>>
<<- end>>
  {
    field: 'created_at',
    title: t('table.created_at'),
    width: 180,
    sortable: true
  },
  {
    field: 'operation',
    title: t('table.operation'),
    width: 180,
    fixed: 'right',
    slot: 'operation'
  }
])

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  await handleDeleteCrud(row, loadData)
}

const handleFormSuccess = () => {
  handleClose()
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.<<.ModuleName>>-list {
  padding: 20px;
}
</style>