<template>
  <TreeListPage
    ref="treeListPageRef"
    page-class="department"
    :title="$t('menu.department')"
    :add-button-text="$t('department.add_department')"
    :add-button-disabled="getButtonState('department.store').disabled"
    :search-form="searchForm"
    :search-fields="searchFields"
    :initial-search-values="departmentInitialSearchForm"
    i18n-prefix="department"
    :table-data="tableData"
    :loading="loading"
    :table-columns="tableColumns"
    :table-key="`table-${tableColumns.length}-${JSON.stringify(columnOrder)}`"
    :default-expand-all="isExpanded"
    :visible-columns="visibleColumns"
    :all-columns="allColumns"
    :default-visible-columns="defaultVisibleColumns"
    :column-order="columnOrder"
    :fixed-columns="fixedColumns"
    :on-column-setting-confirm="handleColumnSettingConfirm"
    @add="handleAdd"
    @search="handleSearch"
    @reset="handleReset"
    @refresh="loadData"
  >
    <template #header-extra>
      <el-button v-if="!hasSearch" @click="handleToggleExpand">
        <el-icon><component :is="isExpanded ? Fold : Expand" /></el-icon>
        {{ isExpanded ? $t('menu_management.collapse_all') : $t('menu_management.expand_all') }}
      </el-button>
    </template>

    <template #remark="{ row }">
      {{ row.remark || row.description || '-' }}
    </template>

    <template #status="{ row }">
      <el-tag :type="Number(row.status ?? 1) === 1 ? 'success' : 'danger'">
        {{ Number(row.status ?? 1) === 1 ? $t('common.enabled') : $t('common.disabled') }}
      </el-tag>
    </template>

    <template #operation="{ row }">
      <el-button
        type="primary"
        link
        :disabled="getButtonState('department.update').disabled"
        @click="handleEdit(row)"
      >
        {{ $t('common.edit') }}
      </el-button>
      <el-button
        type="danger"
        link
        :disabled="getButtonState('department.destroy').disabled"
        @click="handleDelete(row)"
      >
        {{ $t('common.delete') }}
      </el-button>
    </template>

    <template #form>
      <DepartmentForm
        v-model="dialogVisible"
        :edit-id="editId"
        :department-options="departmentOptions"
        @success="handleFormSuccess"
      />
    </template>
  </TreeListPage>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Fold, Expand } from '@element-plus/icons-vue'
import TreeListPage from '@/components/TreeListPage.vue'
import DepartmentForm from './DepartmentForm.vue'
import { useStandardTreeListPage } from '@/composables/useStandardTreeListPage'
import { useTreeExpand } from '@/composables/useTreeExpand'
import { useColumnSetting } from '@/composables/useColumnSetting'
import { useElTableColumns } from '@/composables/useElTableColumns'
import { getDepartmentList, deleteDepartment } from '@/api/department'
import {
  departmentInitialSearchForm,
  createDepartmentSearchFields,
  createDepartmentTableColumns
} from './department.config'

const { t } = useI18n()
const treeListPageRef = ref(null)

const {
  searchForm,
  tableData,
  loading,
  dialogVisible,
  editId,
  loadData,
  handleSearch,
  handleReset,
  handleAdd,
  handleEdit,
  handleDelete,
  handleFormSuccess,
  getButtonState
} = useStandardTreeListPage({
  fetchApi: getDepartmentList,
  initialSearchForm: departmentInitialSearchForm,
  normalizeRows: false,
  deleteApi: deleteDepartment
})

const { isExpanded, handleToggleExpand } = useTreeExpand(treeListPageRef, tableData)

const searchFields = computed(() => createDepartmentSearchFields(t))
const allTableColumns = computed(() => createDepartmentTableColumns(t))

const {
  tableColumns: tableColumnsConfig,
  visibleColumns,
  allColumns,
  defaultVisibleColumns,
  columnOrder,
  fixedColumns,
  handleColumnSettingConfirm
} = useColumnSetting('department', allTableColumns)

const tableColumns = useElTableColumns(tableColumnsConfig, visibleColumns, columnOrder, fixedColumns)
const hasSearch = computed(() => !!(searchForm.name || searchForm.status))
const departmentOptions = computed(() => tableData.value || [])
</script>
