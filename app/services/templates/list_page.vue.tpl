<template>
  <div class="<<.ModuleName>>-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.<<$.ModuleName>>') }}</span>
          <<if .HasCreate>>
          <el-button 
            type="primary" 
            :disabled="getButtonState('<<.ModuleName>>.store').disabled"
            @click="handleAdd"
          >
            <el-icon><PlusIcon /></el-icon>
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
      >
        <<if .HasExport>>
        <template #extra-buttons>
          <el-button 
            type="success" 
            :disabled="getButtonState('<<.ModuleName>>.export').disabled || isExporting"
            :loading="isExporting"
            @click="handleExport"
          >
            {{ $t('common.export') }}
          </el-button>
        </template>
        <<end>>
      </SearchForm>

      <VxeTable
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="600"
        @sort-change="handleSortChange"
      >
        <<range .ListFields>>
        <<- if and .ShowInList (eq .Name "status") (eq .FormType "switch")>>
        <template #status="{ row }">
          <el-switch
            :model-value="Number(row.status ?? row.Status ?? 1) === 1"
            :disabled="getButtonState('<<$.ModuleName>>.update').disabled"
            @change="(val) => handleStatusChange(row, val)"
          />
        </template>
        <<- else if and .ShowInList (eq .FormType "image-upload")>>
        <template #<<.Name>>="{ row }">
          <el-image
            v-if="row.<<.Name>>"
            :src="row.<<.Name>>"
            :preview-src-list="[row.<<.Name>>]"
            fit="cover"
            style="width: 60px; height: 60px; border-radius: 4px; cursor: pointer;"
            lazy
          />
          <span v-else>-</span>
        </template>
        <<- else if and .ShowInList .Relation>>
        <template #<<.Name>>="{ row }">
          {{ get<<.Relation.JsonName>>DisplayName(row.<<.Relation.JsonName>> || row.<<.Name>>) }}
        </template>
        <<- end>>
        <<- end>>
        <template #operation="{ row }">
          <TableActionButtons
            :row="row"
            :primary-actions="getPrimaryActions(row)"
            :more-actions="getMoreActions(row)"
            :get-button-state="getButtonState"
            @action="handleAction"
          />
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
      />
    </el-card>

    <<print "<">><<.ModelName>>Form
      ref="formRef"
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
<<if .HasExport>>
import { useRouter } from 'vue-router'
<<end>>
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import <<.ModelName>>Form from './<<.ModelName>>Form.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
<<range .ListFields>>
<<- if and .ShowInList (eq .Name "status") (not .Dictionary)>>
import { getStatusOptions } from '../../utils/fieldOptions'
<<- end>>
<<- end>>
import {
  get<<.ModelName>>List,
  delete<<.ModelName>>,
  <<if .HasEdit>>update<<.ModelName>>,<<end>>
  <<if .HasExport>>export<<.ModelName>>,<<end>>
} from '../../api/<<.ModuleName>>'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

const PlusIcon = markRaw(Plus)

// 权限控制
const { getButtonState } = usePermission()

const { t } = useI18n()
<<if .HasExport>>
const router = useRouter()
<<end>>
const tableRef = ref(null)
const formRef = ref(null)
<<if .HasExport>>
const isExporting = ref(false)
<<end>>

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
  handleSortChange,
  initDefaultSort
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
    type: <<if and .ApiUrl (or (and .Relation .Relation.IsTree) .IsTree)>>'tree-select'<<else>>'<<.SearchUIType>>'<<end>>,
    clearable: true,
<<if or (eq .SearchUIType "select") (eq .SearchUIType "radio") (eq .SearchUIType "checkbox") (and .ApiUrl (and .Relation .Relation.IsTree))>>
    <<if .Relation>>
    <<- if .ApiUrl>>
    <<- if .Relation.IsTree>>
    apiUrl: '<<.ApiUrl>>',
    treeProps: { label: '<<.Relation.DisplayField>>', value: 'id', children: 'children' },
    <<- else>>
    apiUrl: '<<.ApiUrl>>',
    optionLabelKey: '<<.Relation.DisplayField>>',
    optionValueKey: 'id',
    <<- end>>
    <<- end>>
    <<else if .ApiUrl>>
    <<- if or (eq .SearchUIType "select") (eq .SearchUIType "radio") (eq .SearchUIType "checkbox")>>
    apiUrl: '<<.ApiUrl>>',
    <<- if .IsTree>>
    treeProps: { label: 'label', value: 'value', children: 'children' },
    <<- end>>
    <<- end>>
    <<end>>
    <<- if and (not .ApiUrl) (not .Relation) .Dictionary>>
    apiUrl: '/options?type=dictionary&dictionary_type=<<.Dictionary>>',
    <<- else if and (eq .Name "status") (not .ApiUrl) (not .Dictionary)>>
    options: getStatusOptions(t),
    <<- end>>
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
<<- if .ShowInList>>
  <<- if and (eq .Name "status") (eq .FormType "switch")>>
  {
    field: 'status',
    title: t('table.status'),
    width: 100,
    sortable: false,
    slot: 'status'
  },
  <<- else if eq .FormType "image-upload">>
  {
    field: '<<.Name>>',
    title: t('<<$.ModuleName>>.<<.Name>>'),
    slot: '<<.Name>>',
    sortable: false,
    width: 120
  },
  <<- else if .Relation>>
  {
    field: '<<.Name>>',
    title: t('<<$.ModuleName>>.<<.Name>>'),
    slot: '<<.Name>>',
    sortable: false
  },
  <<- else>>
  {
    field: '<<.Name>>',
    title: t('<<$.ModuleName>>.<<.Name>>'),
    sortable: <<.Sortable>>
  },
  <<- end>>
<<- end>>
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
    width: 220,
    fixed: 'right',
    slot: 'operation'
  }
])

<<range .ListFields>>
<<- if and .ShowInList .Relation>>
const get<<.Relation.JsonName>>DisplayName = (<<.Name>>) => {
  if (!<<.Name>>) return '-'
  return <<.Name>>.<<.Relation.DisplayField>> || <<.Name>>.<<.Relation.JsonName>> || '-'
}
<<- end>>
<<- end>>

<<if .HasEdit>>
<<range .ListFields>>
<<- if and .ShowInList (eq .Name "status") (eq .FormType "switch")>>
const handleStatusChange = async (row, newStatus) => {
  try {
    const statusValue = newStatus ? 1 : 0
    await update<<$.ModelName>>(row.id, {
      status: statusValue
    })
    ElMessage.success(newStatus ? t('common.enabled') : t('common.disabled'))
    // 更新本地数据
    const item = tableData.value.find(a => a.id === row.id)
    if (item) {
      item.status = statusValue
      item.Status = statusValue
    }
  } catch (error) {
    logger.error('Status change error:', error)
    loadData()
    if (!error.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
      ElMessage.error(errorMessage)
    }
  }
}
<<- end>>
<<- end>>
<<end>>

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleDelete = (row) => handleDeleteCrud(row, loadData)

const handleFormSuccess = () => {
  handleClose()
  loadData()
}

// 获取主要操作按钮配置
const getPrimaryActions = (row) => {
  return [
    <<if .HasEdit>>
    {
      key: 'edit',
      label: t('common.edit'),
      type: 'primary',
      permission: '<<.ModuleName>>.update',
      handler: handleEdit
    },
    <<end>>
    <<if .HasDelete>>
    {
      key: 'delete',
      label: t('common.delete'),
      type: 'danger',
      permission: '<<.ModuleName>>.destroy',
      handler: handleDelete
    }
    <<end>>
  ]
}

// 获取更多操作按钮配置（可根据需要扩展）
const getMoreActions = (row) => {
  return []
}

// 处理操作事件
const handleAction = (command, row) => {
  switch (command) {
    case 'edit':
      handleEdit(row)
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

<<if .HasExport>>
const handleExport = async () => {
  if (isExporting.value) {
    return
  }

  isExporting.value = true

  try {
    await export<<.ModelName>>(searchForm)
    ElMessage.success(t('common.export_task_submitted'))
    router.push('/exports')
  } catch (error) {
    logger.error('Export error:', error)
    if (error.response?.status === 429) {
      ElMessage.warning(t('common.export_in_progress'))
    } else if (!error.__handled) {
      ErrorHandler.handle(error, { silent: true })
    }
  } finally {
    isExporting.value = false
  }
}
<<end>>

onMounted(async () => {
  try {
    initDefaultSort()
    await loadData()
  } catch (error) {
    logger.error('ListPage onMounted error:', error)
    ErrorHandler.handle(error)
  }
})
</script>

<style scoped>
.<<.ModuleName>>-list {
  padding: 20px;
}
</style>
