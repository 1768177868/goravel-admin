<template>
  <div class="vxe-table-wrapper" :class="{ 'mobile-view': isMobile }">
    <!-- 移动端卡片视图 -->
    <div v-if="isMobile && data.length > 0" class="mobile-card-list">
      <div
        v-for="(row, index) in data"
        :key="index"
        class="mobile-card-item"
        @click="handleCardClick(row)"
      >
        <div
          v-for="column in mobileColumns"
          :key="column.field || column.slot"
          class="card-field"
        >
          <div class="card-label">{{ column.title }}</div>
          <div class="card-value">
            <slot v-if="column.slot" :name="column.slot" :row="row" />
            <span v-else>{{ getFieldValue(row, column.field) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 桌面端表格视图 -->
    <vxe-table
      v-else
      ref="tableRef"
      :data="data"
      :loading="loading"
      border
      :column-config="{ resizable: true }"
      :height="height"
      :tree-config="treeConfig"
      :sort-config="{ multiple: false, trigger: 'default' }"
      :scroll-x="{ enabled: true }"
      :scroll-y="{ enabled: true }"
      @sort-change="handleSortChange"
      @checkbox-change="handleCheckboxChange"
      @checkbox-all="handleCheckboxAll"
      class="desktop-table"
    >
      <template v-for="(column, index) in columns" :key="column.field || column.slot || index">
        <vxe-column
          :type="column.type"
          :field="column.field"
          :title="column.title"
          :width="column.width"
          :min-width="column.minWidth || (isMobile ? 100 : 120)"
          :sortable="column.sortable"
          :fixed="column.fixed"
          :formatter="column.formatter"
          :tree-node="column.treeNode"
        >
          <template v-if="column.slot" #default="{ row }">
            <slot :name="column.slot" :row="row" />
          </template>
        </vxe-column>
      </template>
    </vxe-table>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useResponsive } from '../composables/useResponsive'

const { isMobile } = useResponsive()

const props = defineProps({
  data: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  columns: {
    type: Array,
    required: true
  },
  height: {
    type: [String, Number],
    default: 600
  },
  treeConfig: {
    type: [Object, Boolean],
    default: null
  }
})

const emit = defineEmits(['sort-change', 'checkbox-change', 'checkbox-all', 'row-click'])

const tableRef = ref(null)

// 移动端显示的列（排除操作列，操作列在卡片底部显示）
const mobileColumns = computed(() => {
  return props.columns.filter(col => 
    col.field && 
    col.field !== 'operation' && 
    col.type !== 'checkbox' &&
    col.type !== 'seq'
  ).slice(0, 4) // 移动端最多显示4个字段
})

// 获取字段值
const getFieldValue = (row, field) => {
  if (!field) return ''
  const keys = field.split('.')
  let value = row
  for (const key of keys) {
    value = value?.[key]
    if (value === undefined || value === null) return ''
  }
  return value
}

const handleSortChange = (params) => {
  emit('sort-change', params)
}

const handleCheckboxChange = (params) => {
  emit('checkbox-change', params)
}

const handleCheckboxAll = (params) => {
  emit('checkbox-all', params)
}

const handleCardClick = (row) => {
  emit('row-click', row)
}

defineExpose({
  get tableRef() {
    return tableRef.value
  }
})
</script>

<style scoped lang="scss">
.vxe-table-wrapper {
  width: 100%;
}

/* 移动端卡片视图 */
.mobile-card-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 8px 0;
}

.mobile-card-item {
  background: var(--card-bg, #fff);
  border: 1px solid var(--border-color-light, #e4e7ed);
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: all 0.3s;
  cursor: pointer;

  &:active {
    transform: scale(0.98);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  }
}

.card-field {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;

  &:last-child {
    margin-bottom: 0;
  }
}

.card-label {
  font-size: 12px;
  color: var(--text-color-secondary, #909399);
  min-width: 80px;
  flex-shrink: 0;
  margin-right: 12px;
}

.card-value {
  flex: 1;
  font-size: 14px;
  color: var(--text-color-primary, #303133);
  text-align: right;
  word-break: break-word;
}

/* 桌面端表格 */
.desktop-table {
  width: 100%;
}

/* 移动端表格优化 */
@media (max-width: 768px) {
  .vxe-table-wrapper.mobile-view {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .desktop-table {
    min-width: 600px; // 确保表格可以横向滚动
  }

  :deep(.vxe-table) {
    font-size: 12px;
  }

  :deep(.vxe-cell) {
    padding: 8px 4px;
  }

  :deep(.vxe-header--column) {
    padding: 8px 4px;
    font-size: 12px;
  }
}

/* 平板优化 */
@media (min-width: 769px) and (max-width: 991px) {
  :deep(.vxe-table) {
    font-size: 13px;
  }

  :deep(.vxe-cell) {
    padding: 10px 6px;
  }
}
</style>

