<template>
  <div :class="[`${pageClass}-list`, 'list-page']">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ title }}</span>
          <div class="header-actions">
            <slot name="header-extra" />
            <el-button
              v-if="showAddButton"
              type="primary"
              :disabled="addButtonDisabled"
              @click="$emit('add')"
            >
              <el-icon><Plus /></el-icon>
              {{ addButtonText }}
            </el-button>
          </div>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchValues"
        :i18n-prefix="i18nPrefix"
        :loading="loading"
        @search="$emit('search')"
        @reset="$emit('reset', $event)"
      >
        <template #extra-buttons>
          <slot name="extra-buttons" />
        </template>
      </SearchForm>

      <TableToolbar
        v-if="showToolbar"
        :on-refresh="handleRefresh"
        :show-column-setting-btn="showColumnSetting"
        :show-table-style="showTableStyle"
        :visible-columns="visibleColumns"
        :all-columns="allColumns"
        :default-visible-columns="defaultVisibleColumns"
        :column-order="columnOrder"
        :fixed-columns="fixedColumns"
        :on-column-setting-confirm="onColumnSettingConfirm"
      />

      <div class="list-table-scroll">
        <el-table
          ref="tableRef"
          :key="tableKey"
          :data="tableData"
          :loading="loading"
          border
          :row-key="rowKey"
          :tree-props="treeProps"
          :default-expand-all="defaultExpandAll"
          style="width: 100%"
          :height="tableHeight"
        >
          <el-table-column
            v-for="column in tableColumns"
            :key="column.key || column.prop || column.type"
            :type="column.type"
            :prop="column.prop"
            :label="column.label"
            :width="column.width"
            :min-width="column.minWidth"
            :fixed="column.fixed"
          >
            <template v-if="column.slot" #default="{ row }">
              <slot :name="column.slot" :row="row" />
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <slot name="form" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from './SearchForm.vue'
import TableToolbar from './TableToolbar.vue'

const emit = defineEmits(['add', 'search', 'reset', 'refresh'])

defineProps({
  pageClass: { type: String, default: 'page' },
  title: { type: String, required: true },
  showAddButton: { type: Boolean, default: true },
  addButtonText: { type: String, default: '' },
  addButtonDisabled: { type: Boolean, default: false },
  searchForm: { type: Object, required: true },
  searchFields: { type: Array, required: true },
  initialSearchValues: { type: Object, default: () => ({}) },
  i18nPrefix: { type: String, default: '' },
  tableData: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  tableColumns: { type: Array, required: true },
  tableKey: { type: String, default: 'tree-table' },
  rowKey: { type: String, default: 'id' },
  treeProps: {
    type: Object,
    default: () => ({ children: 'children', hasChildren: 'hasChildren' })
  },
  defaultExpandAll: { type: Boolean, default: false },
  tableHeight: { type: [String, Number], default: 600 },
  showToolbar: { type: Boolean, default: true },
  showColumnSetting: { type: Boolean, default: false },
  showTableStyle: { type: Boolean, default: false },
  visibleColumns: { type: Array, default: () => [] },
  allColumns: { type: Array, default: () => [] },
  defaultVisibleColumns: { type: Array, default: () => [] },
  columnOrder: { type: Array, default: () => [] },
  fixedColumns: { type: Object, default: () => ({}) },
  onColumnSettingConfirm: { type: Function, default: null }
})

const tableRef = ref(null)

const handleRefresh = () => {
  emit('refresh')
}

defineExpose({
  tableRef,
  toggleRowExpansion: (row, expanded) => tableRef.value?.toggleRowExpansion(row, expanded)
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.card-header > span {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-color-primary);
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
