<template>
  <div class="vxe-table-wrapper" :class="{ 'mobile-view': isMobile }">
    <!-- 移动端卡片视图 -->
    <div v-if="isMobile && data.length > 0" class="mobile-card-list">
      <div
        v-for="(row, index) in data"
        :key="row.id ?? row.ID ?? index"
        class="mobile-card-item"
        @click="handleCardClick(row)"
      >
        <div
          v-for="column in mobileDataColumns"
          :key="column.field || column.slot || column.key"
          class="card-field"
        >
          <div class="card-label">{{ column.title }}</div>
          <div class="card-value">
            <slot v-if="column.slot" :name="column.slot" :row="row" />
            <span v-else>{{ getFieldValue(row, column.field) }}</span>
          </div>
        </div>
        <div
          v-if="slots.operation"
          class="card-actions"
          @click.stop
        >
          <slot name="operation" :row="row" />
        </div>
      </div>
    </div>

    <!-- 桌面端表格视图 -->
    <vxe-table
      v-else
      ref="tableRef"
      :data="data"
      :loading="loading"
      :border="tableStyle.border"
      :stripe="tableStyle.stripe"
      :size="vxeSize"
      :column-config="{ resizable: true }"
      :height="height"
      :tree-config="treeConfig"
      :sort-config="{ multiple: false, trigger: 'default' }"
      :scroll-x="{ enabled: true }"
      :scroll-y="{ enabled: true }"
      class="desktop-table"
      @sort-change="handleSortChange"
      @checkbox-change="handleCheckboxChange"
      @checkbox-all="handleCheckboxAll"
    >
      <template v-for="(column, index) in columns" :key="`${index}-${column.field || column.slot || column.key || column.type}`">
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
import { ref, computed, useSlots, onMounted, onBeforeUnmount } from 'vue'
import { useResponsive } from '../composables/useResponsive'
import { useVxeTableSize } from '../composables/useVxeTableSize'

const slots = useSlots()

const { isMobile } = useResponsive()
const { vxeSize } = useVxeTableSize()
const TABLE_STYLE_STORAGE_KEY = 'table_style_preferences'
const DEFAULT_TABLE_STYLE = {
  stripe: false,
  border: true
}

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
const tableStyle = ref({ ...DEFAULT_TABLE_STYLE })

const applyTableStyle = (style) => {
  tableStyle.value = {
    stripe: style?.stripe === true,
    border: style?.border !== false
  }
}

const loadTableStyle = () => {
  try {
    const raw = window.localStorage.getItem(TABLE_STYLE_STORAGE_KEY)
    if (!raw) {
      applyTableStyle(DEFAULT_TABLE_STYLE)
      return
    }
    applyTableStyle(JSON.parse(raw))
  } catch {
    applyTableStyle(DEFAULT_TABLE_STYLE)
  }
}

const handleTableStyleChange = (event) => {
  if (event?.detail) {
    applyTableStyle(event.detail)
    return
  }
  loadTableStyle()
}

function isOperationColumn(col) {
  return (
    col.slot === 'operation' ||
    col.key === 'operation' ||
    col.field === 'operation'
  )
}

function isSpecialColumnType(col) {
  const t = col.type
  return (
    t === 'checkbox' ||
    t === 'seq' ||
    t === 'radio' ||
    t === 'expand' ||
    t === 'index'
  )
}

// 移动端卡片：展示全部数据列（操作列由底部 #operation 插槽渲染）
const mobileDataColumns = computed(() => {
  return props.columns.filter((col) => {
    if (isOperationColumn(col)) return false
    if (isSpecialColumnType(col)) return false
    if (!col.field && !col.slot) return false
    return true
  })
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

onMounted(() => {
  loadTableStyle()
  window.addEventListener('table-style-change', handleTableStyleChange)
})

onBeforeUnmount(() => {
  window.removeEventListener('table-style-change', handleTableStyleChange)
})

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

.card-actions {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color-lighter, #ebeef5);
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
  align-items: center;
}

.card-actions :deep(.el-button) {
  min-height: 36px;
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

/* 注意：暗黑模式通过全局 CSS 变量映射自动适配（参考 vue3-element-admin）
   详见 style.css 中的 --vxe-* 变量映射到 Element Plus 变量 */
</style>

