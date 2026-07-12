<template>
  <div :class="[`${pageClass}-list`]">
    <el-card>
      <template #header>
        <div class="card-header">
          <slot name="header-title">
            <span>{{ title }}</span>
          </slot>
          <div class="header-actions">
            <slot name="header-actions">
              <el-button 
                v-if="showAddButton"
                type="primary" 
                :disabled="addButtonDisabled"
                @click="handleAdd"
              >
                <el-icon><Plus /></el-icon>
                {{ addButtonText }}
              </el-button>
            </slot>
          </div>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchValues"
        :i18n-prefix="i18nPrefix"
        :loading="loading"
        @search="handleSearch"
        @reset="handleReset"
      >
        <template v-for="name in searchFieldSlotNames" #[name]="slotProps">
          <slot :name="name" v-bind="slotProps || {}" />
        </template>
        <template #extra-buttons>
          <slot name="extra-buttons" />
        </template>
      </SearchForm>

      <slot name="before-table" />

      <slot name="toolbar">
        <TableToolbar
          v-if="showToolbar"
          :on-refresh="handleRefresh"
          :show-column-setting-btn="showColumnSetting"
          :visible-columns="visibleColumns"
          :all-columns="allColumns"
          :default-visible-columns="defaultVisibleColumns"
          :column-order="columnOrder"
          :fixed-columns="fixedColumns"
          :on-column-setting-confirm="onColumnSettingConfirm"
        >
          <template #left>
            <slot name="toolbar-left" />
          </template>
          <template #right>
            <slot name="toolbar-right" />
          </template>
        </TableToolbar>
      </slot>

      <slot v-if="$slots.table" name="table" />
      <VxeTable
        v-else
        ref="tableRef"
        :key="tableKey"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="tableHeight"
        @sort-change="handleSortChange"
        @checkbox-change="handleCheckboxChange"
        @checkbox-all="handleCheckboxAll"
        v-bind="$attrs"
      >
        <template v-for="column in slottedColumns" #[column.slot]="slotProps">
          <slot :name="column.slot" v-bind="slotProps || {}" />
        </template>
      </VxeTable>

      <Pagination
        :model-value="pagination"
        :auto-load="autoLoad"
        :hide-total-threshold="hideTotalThreshold"
        :on-page-change="handlePageChange"
        @update:model-value="handlePaginationUpdate"
      />
    </el-card>

    <!-- 表单对话框 -->
    <slot name="form" />
  </div>
</template>

<script setup>
import { ref, computed, useSlots } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from './SearchForm.vue'
import Pagination from './Pagination.vue'
import VxeTable from './VxeTable.vue'
import TableToolbar from './TableToolbar.vue'

const slots = useSlots()

const RESERVED_SLOTS = new Set([
  'header-title',
  'header-actions',
  'extra-buttons',
  'before-table',
  'toolbar',
  'toolbar-left',
  'toolbar-right',
  'table',
  'form'
])

const props = defineProps({
  // 页面类名
  pageClass: {
    type: String,
    default: 'page'
  },
  // 标题
  title: {
    type: String,
    default: ''
  },
  // 是否显示添加按钮
  showAddButton: {
    type: Boolean,
    default: true
  },
  // 添加按钮文本
  addButtonText: {
    type: String,
    default: ''
  },
  // 添加按钮是否禁用
  addButtonDisabled: {
    type: Boolean,
    default: false
  },
  // 搜索表单数据
  searchForm: {
    type: Object,
    required: true
  },
  // 搜索表单字段配置
  searchFields: {
    type: Array,
    required: true
  },
  // 初始搜索值
  initialSearchValues: {
    type: Object,
    default: () => ({})
  },
  // 国际化前缀
  i18nPrefix: {
    type: String,
    default: ''
  },
  // 表格数据
  tableData: {
    type: Array,
    default: () => []
  },
  // 加载状态
  loading: {
    type: Boolean,
    default: false
  },
  // 表格列配置
  tableColumns: {
    type: Array,
    required: true
  },
  // 表格配置
  tableConfig: {
    type: Object,
    default: () => ({ resizable: true })
  },
  // 表格高度
  tableHeight: {
    type: [String, Number],
    default: 600
  },
  // 分页数据
  pagination: {
    type: Object,
    required: true
  },
  // 是否显示工具栏
  showToolbar: {
    type: Boolean,
    default: false
  },
  // 是否显示列设置
  showColumnSetting: {
    type: Boolean,
    default: false
  },
  visibleColumns: {
    type: Array,
    default: () => []
  },
  allColumns: {
    type: Array,
    default: () => []
  },
  defaultVisibleColumns: {
    type: Array,
    default: () => []
  },
  columnOrder: {
    type: Array,
    default: () => []
  },
  fixedColumns: {
    type: Object,
    default: () => ({})
  },
  onColumnSettingConfirm: {
    type: Function,
    default: null
  },
  tableKey: {
    type: String,
    default: ''
  },
  // 分页变化后是否自动加载
  autoLoad: {
    type: Boolean,
    default: true
  },
  // 总数超过阈值时隐藏分页总数（0 表示不隐藏）
  hideTotalThreshold: {
    type: Number,
    default: 0
  },
  // 对话框显示状态
  dialogVisible: {
    type: Boolean,
    default: false
  },
  // 编辑ID
  editId: {
    type: [Number, String],
    default: null
  }
})

const searchFieldSlotNames = computed(() => {
  const columnSlots = new Set(
    (props.tableColumns || [])
      .map((column) => column.slot)
      .filter(Boolean)
  )
  return Object.keys(slots).filter(
    (name) => !RESERVED_SLOTS.has(name) && !columnSlots.has(name)
  )
})

const slottedColumns = computed(() =>
  (props.tableColumns || []).filter((column) => column.slot)
)

const emit = defineEmits([
  'add',
  'search',
  'reset',
  'refresh',
  'update:pagination',
  'page-change',
  'sort-change',
  'selection-change',
  'form-success'
])

const tableRef = ref(null)

const handleRefresh = () => {
  emit('refresh')
}

const handleAdd = () => {
  emit('add')
}

const handleSearch = () => {
  emit('search')
}

const handleReset = () => {
  emit('reset')
}

const handlePaginationUpdate = (value) => {
  emit('update:pagination', value)
}

const handlePageChange = (data) => {
  emit('page-change', data)
}

const handleSortChange = (data) => {
  emit('sort-change', data)
}

const handleCheckboxChange = (payload) => {
  emit('selection-change', payload?.records || [])
}

const handleCheckboxAll = (payload) => {
  emit('selection-change', payload?.records || [])
}

const handleFormSuccess = () => {
  emit('form-success')
}

defineExpose({
  tableRef
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
  letter-spacing: 0.2px;
  color: var(--text-color-primary);
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.header-actions :deep(.el-button--primary) {
  border-radius: 10px;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--el-color-primary) 30%, transparent);
}
</style>

