<template>
  <div :class="[`${pageClass}-list`]">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ title }}</span>
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
        <template v-for="(_, name) in $slots" #[name]="slotData" :key="name">
          <slot :name="name" v-bind="slotData" />
        </template>
      </SearchForm>

      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        :column-config="optimizedTableConfig"
        :height="tableHeight"
        :scroll-y="scrollYConfig"
        :sort-config="{ multiple: false, trigger: 'default' }"
        @sort-change="handleSortChange"
        v-bind="$attrs"
        v-on="$attrs"
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
            <!-- 默认 slot 处理 -->
            <template v-if="column.slot" #default="slotProps">
              <slot 
                :name="column.slot" 
                :row="slotProps.row" 
                :column="slotProps.column"
                :rowIndex="slotProps.rowIndex"
                :columnIndex="slotProps.columnIndex"
              />
            </template>
          </vxe-column>
        </template>
        
        <!-- 额外的表格插槽 -->
        <template v-for="(_, name) in $slots" #[name]="slotData" :key="name">
          <slot :name="name" v-bind="slotData" />
        </template>
      </vxe-table>

      <Pagination
        :model-value="pagination"
        @update:model-value="handlePaginationUpdate"
        @page-change="handlePageChange"
      />
    </el-card>

    <!-- 表单对话框 -->
    <slot name="form" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from './SearchForm.vue'
import Pagination from './Pagination.vue'
import { useTablePerformance } from '../composables/useTablePerformance'

const props = defineProps({
  // 页面类名
  pageClass: {
    type: String,
    default: 'page'
  },
  // 标题
  title: {
    type: String,
    required: true
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

const emit = defineEmits([
  'add',
  'search',
  'reset',
  'update:pagination',
  'page-change',
  'sort-change',
  'form-success'
])

const tableRef = ref(null)

// 性能优化：虚拟滚动和列渲染优化
const {
  scrollYConfig,
  optimizedTableConfig
} = useTablePerformance({
  tableColumns: computed(() => props.tableColumns),
  tableData: computed(() => props.tableData),
  tableHeight: props.tableHeight
})

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
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>

