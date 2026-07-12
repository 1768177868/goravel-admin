<template>
  <ListPage
    ref="listPageRef"
    page-class="dictionary"
    :title="$t('menu.dictionary')"
    :add-button-text="$t('dictionary.add_dictionary')"
    :add-button-disabled="getButtonState('dictionary.store').disabled"
    :search-form="searchForm"
    :search-fields="searchFields"
    :initial-search-values="dictionaryInitialSearchForm"
    i18n-prefix="dictionary"
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
      <DictionaryForm
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
import DictionaryForm from './DictionaryForm.vue'
import { useStandardListPage } from '@/composables/useStandardListPage'
import { createCrudActions, rowStatus } from '@/utils/listPageHelpers'
import { getDictionaryList, deleteDictionary } from '@/api/dictionary'
import {
  dictionaryInitialSearchForm,
  createDictionarySearchFields,
  createDictionaryTableColumns
} from './dictionary.config'

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
  fetchApi: getDictionaryList,
  initialSearchForm: dictionaryInitialSearchForm,
  defaultSort: 'id:desc',
  deleteApi: deleteDictionary,
  tableRef: computed(() => listPageRef.value?.tableRef?.tableRef)
})

const searchFields = computed(() => createDictionarySearchFields(t))
const tableColumns = computed(() => createDictionaryTableColumns(t))

const operationActions = computed(() =>
  createCrudActions(t, 'dictionary', {
    onEdit: handleEdit,
    onDelete: handleDelete
  })
)
</script>
