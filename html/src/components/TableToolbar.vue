<template>
  <div class="table-toolbar">
    <div class="toolbar-left">
      <slot name="left"></slot>
    </div>
    <div class="toolbar-right">
      <el-button-group>
        <!-- 刷新按钮 -->
        <el-button 
          v-if="showRefresh"
          :icon="RefreshIcon" 
          circle
          :class="{ 'is-refreshing': isRefreshing }"
          @click="handleRefresh"
          :title="$t('common.refresh')"
        />
        <!-- 表格样式设置 -->
        <el-popover
          v-if="showTableStyle"
          trigger="click"
          placement="bottom"
          :width="236"
          popper-class="table-style-popover"
        >
          <template #reference>
            <el-button
              :icon="StyleIcon"
              circle
              :title="$t('common.table_style')"
            />
          </template>
          <div class="style-settings">
            <div class="style-settings__title">{{ $t('common.table_style') }}</div>
            <div class="style-item" :class="{ 'is-active': tableStyle.stripe }">
              <el-checkbox v-model="tableStyle.stripe" @change="handleTableStyleChange" />
              <div class="style-item__text">
                <span class="style-item__label">{{ $t('common.table_zebra') }}</span>
              </div>
            </div>
            <div class="style-item" :class="{ 'is-active': tableStyle.border }">
              <el-checkbox v-model="tableStyle.border" @change="handleTableStyleChange" />
              <div class="style-item__text">
                <span class="style-item__label">{{ $t('common.table_border') }}</span>
              </div>
            </div>
          </div>
        </el-popover>
        <!-- 列设置按钮 -->
        <ColumnSettingDialog
          v-if="showColumnSettingBtn"
          v-model="showColumnSettingDialog"
          :visible-columns="visibleColumns"
          :all-columns="allColumns"
          :default-visible-columns="defaultVisibleColumns"
          :column-order="columnOrder"
          :fixed-columns="fixedColumns"
          @confirm="handleColumnSettingConfirm"
        />
      </el-button-group>
      <slot name="right"></slot>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, markRaw, onMounted } from 'vue'
import { Refresh, SetUp } from '@element-plus/icons-vue'
import ColumnSettingDialog from './ColumnSettingDialog.vue'

const TABLE_STYLE_STORAGE_KEY = 'table_style_preferences'
const DEFAULT_TABLE_STYLE = {
  stripe: false,
  border: true
}

const props = defineProps({
  // 是否显示刷新按钮
  showRefresh: {
    type: Boolean,
    default: true
  },
  // 是否显示列设置按钮
  showColumnSettingBtn: {
    type: Boolean,
    default: true
  },
  // 是否显示表格样式按钮（斑马纹/边框）
  showTableStyle: {
    type: Boolean,
    default: true
  },
  // 刷新回调函数
  onRefresh: {
    type: Function,
    default: null
  },
  // 列设置相关 props
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
  // 列设置确认回调
  onColumnSettingConfirm: {
    type: Function,
    default: null
  }
})

const emit = defineEmits(['refresh'])

// 图标
const RefreshIcon = markRaw(Refresh)
const StyleIcon = markRaw(SetUp)

// 列设置对话框显示状态
const showColumnSettingDialog = ref(false)

// 刷新动画状态
const isRefreshing = ref(false)
// 表格样式设置
const tableStyle = reactive({ ...DEFAULT_TABLE_STYLE })

// 处理刷新
const handleRefresh = async () => {
  if (isRefreshing.value) return
  isRefreshing.value = true
  const startAt = Date.now()

  try {
    if (props.onRefresh) {
      await props.onRefresh()
    }
    emit('refresh')
  } finally {
    // 给动画一个最短展示时间，避免点击后几乎不可见
    const elapsed = Date.now() - startAt
    const minDuration = 450
    const remain = Math.max(0, minDuration - elapsed)
    window.setTimeout(() => {
      isRefreshing.value = false
    }, remain)
  }
}

// 处理列设置确认
const handleColumnSettingConfirm = (result) => {
  if (props.onColumnSettingConfirm) {
    props.onColumnSettingConfirm(result)
  }
}

const loadTableStyle = () => {
  try {
    const raw = window.localStorage.getItem(TABLE_STYLE_STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    tableStyle.stripe = parsed?.stripe === true
    tableStyle.border = parsed?.border !== false
  } catch {
    tableStyle.stripe = DEFAULT_TABLE_STYLE.stripe
    tableStyle.border = DEFAULT_TABLE_STYLE.border
  }
}

const handleTableStyleChange = () => {
  try {
    const payload = {
      stripe: tableStyle.stripe,
      border: tableStyle.border
    }
    window.localStorage.setItem(TABLE_STYLE_STORAGE_KEY, JSON.stringify(payload))
    window.dispatchEvent(new CustomEvent('table-style-change', { detail: payload }))
  } catch {
    // ignore localStorage failure
  }
}

onMounted(() => {
  loadTableStyle()
})
</script>

<style scoped>
.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
  padding: 4px 6px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--card-bg, #fff) 92%, transparent);
  border: 1px solid color-mix(in srgb, var(--border-color-light) 70%, transparent);
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.05);
}

.toolbar-left {
  flex: 1;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-right :deep(.el-button-group) {
  display: inline-flex;
  gap: 6px;
}

.toolbar-right :deep(.el-button-group > .el-button) {
  border-radius: 10px !important;
  border: 1px solid color-mix(in srgb, var(--border-color-light) 72%, transparent);
  background: color-mix(in srgb, var(--card-bg, #fff) 94%, transparent);
  transition: all 0.2s ease;
}

.toolbar-right :deep(.el-button-group > .el-button:hover) {
  color: var(--el-color-primary);
  border-color: color-mix(in srgb, var(--el-color-primary) 30%, transparent);
  background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
}

.toolbar-right :deep(.el-button-group > .el-button.is-refreshing .el-icon) {
  animation: toolbar-refresh-spin 0.8s linear infinite;
}

:deep(.table-style-popover.el-popper) {
  border-radius: 12px;
  border: 1px solid color-mix(in srgb, var(--el-border-color) 70%, transparent);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.12);
  padding: 10px;
}

.style-settings {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  box-sizing: border-box;
}

.style-settings__title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  letter-spacing: 0.2px;
  padding: 2px 6px 8px;
  border-bottom: 1px solid color-mix(in srgb, var(--el-border-color-lighter) 75%, transparent);
  margin-bottom: 2px;
}

.style-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 9px;
  background: transparent;
  border: 1px solid transparent;
  transition: all 0.2s ease;
}

.style-item :deep(.el-checkbox) {
  margin-top: 0;
  margin-right: 0;
  flex: 0 0 auto;
}

.style-item:hover {
  background: color-mix(in srgb, var(--el-fill-color-light) 85%, transparent);
  border-color: color-mix(in srgb, var(--el-border-color) 68%, transparent);
}

.style-item.is-active {
  border-color: color-mix(in srgb, var(--el-color-primary) 35%, transparent);
  background: color-mix(in srgb, var(--el-color-primary) 8%, var(--el-fill-color-light));
}

.style-item__text {
  display: flex;
  flex-direction: column;
  gap: 0;
  min-width: 0;
  flex: 1;
}

.style-item__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  line-height: 1.25;
  word-break: break-word;
}

@keyframes toolbar-refresh-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>