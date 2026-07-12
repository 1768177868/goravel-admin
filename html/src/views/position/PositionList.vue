<template>
  <ListPage
    ref="listPageRef"
    page-class="position"
    :title="$t('menu.position')"
    :add-button-text="$t('position.add_position')"
    :add-button-disabled="getButtonState('position.store').disabled"
    :search-form="searchForm"
    :search-fields="searchFields"
    :initial-search-values="positionInitialSearchForm"
    i18n-prefix="position"
    :table-data="tableData"
    :loading="loading"
    :table-columns="tableColumns"
    :pagination="pagination"
    show-toolbar
    @add="handleAdd"
    @search="handleSearch"
    @reset="handleReset"
    @refresh="loadData"
    @page-change="loadData"
    @sort-change="handleSortChange"
  >
    <template #status="{ row }">
      <el-tag :type="rowStatus(row) === 1 ? 'success' : 'danger'">
        {{ rowStatus(row) === 1 ? $t('common.enabled') : $t('common.disabled') }}
      </el-tag>
    </template>

    <template #operation="{ row }">
      <TableActionButtons
        :row="row"
        :primary-actions="operationActions"
        :get-button-state="getButtonState"
      />
    </template>

    <template #form>
      <PositionForm
        v-model="dialogVisible"
        :edit-id="editId"
        @success="handleFormSuccess"
      />
    </template>
  </ListPage>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ListPage from '@/components/ListPage.vue'
import TableActionButtons from '@/components/TableActionButtons.vue'
import PositionForm from './PositionForm.vue'
import { useStandardListPage } from '@/composables/useStandardListPage'
import { createCrudActions, rowStatus } from '@/utils/listPageHelpers'
import { getPositionList, deletePosition } from '@/api/position'
import {
  positionInitialSearchForm,
  createPositionSearchFields,
  createPositionTableColumns
} from './position.config'

const { t } = useI18n()
const listPageRef = ref(null)

const {
  pagination,
  tableData,
  loading,
  searchForm,
  dialogVisible,
  editId,
  loadData,
  handleSearch,
  handleReset,
  handleSortChange,
  handleAdd,
  handleEdit,
  handleFormSuccess,
  handleDelete,
  getButtonState
} = useStandardListPage({
  fetchApi: getPositionList,
  initialSearchForm: positionInitialSearchForm,
  defaultSort: 'sort:asc',
  deleteApi: deletePosition,
  tableRef: computed(() => listPageRef.value?.tableRef?.tableRef)
})

const searchFields = computed(() => createPositionSearchFields(t))
const tableColumns = computed(() => createPositionTableColumns(t))

const operationActions = computed(() =>
  createCrudActions(t, 'position', {
    onEdit: handleEdit,
    onDelete: handleDelete
  })
)
</script>
