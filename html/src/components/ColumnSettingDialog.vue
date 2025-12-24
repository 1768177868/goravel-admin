<template>
  <el-popover
    v-model:visible="visible"
    placement="bottom-end"
    :width="popoverWidth"
    trigger="click"
    popper-class="column-setting-popover"
    @hide="handleClose"
  >
    <template #reference>
      <slot name="reference">
        <el-button 
          type="info" 
          size="small"
          :icon="SettingIcon"
          circle
          :title="$t('common.column_setting')"
        />
      </slot>
    </template>
    
    <div class="popover-content">
      <div class="popover-header">
        <span class="popover-title">{{ $t('common.column_setting') }}</span>
      </div>
      <div class="column-list">
        <el-checkbox-group v-model="localVisibleColumns" class="checkbox-group">
          <div
            v-for="column in allColumns"
            :key="column.key"
            class="column-item"
            :class="{ 'column-item-disabled': column.required }"
          >
            <el-checkbox
              :label="column.key"
              :disabled="column.required"
              class="column-checkbox"
            >
              <span class="column-title">{{ column.title }}</span>
            </el-checkbox>
            <el-tag
              v-if="column.required"
              size="small"
              type="info"
              class="required-tag"
            >
              {{ $t('common.required') }}
            </el-tag>
          </div>
        </el-checkbox-group>
      </div>
      <div class="popover-tips">
        <el-icon class="tips-icon"><InfoFilled /></el-icon>
        <span>{{ $t('common.column_setting_tip') }}</span>
      </div>
      <div class="popover-footer">
        <el-button size="small" @click="handleReset">{{ $t('common.reset') }}</el-button>
        <el-button size="small" @click="handleClose">{{ $t('common.cancel') }}</el-button>
        <el-button size="small" type="primary" @click="handleConfirm">{{ $t('common.confirm') }}</el-button>
      </div>
    </div>
  </el-popover>
</template>

<script setup>
import { ref, watch, markRaw } from 'vue'
import { InfoFilled, Setting } from '@element-plus/icons-vue'

const SettingIcon = markRaw(Setting)

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  visibleColumns: {
    type: Array,
    required: true
  },
  allColumns: {
    type: Array,
    required: true
  },
  defaultVisibleColumns: {
    type: Array,
    default: () => []
  },
  popoverWidth: {
    type: [String, Number],
    default: 280
  }
})

const emit = defineEmits(['update:modelValue', 'confirm'])

const visible = ref(props.modelValue)
const localVisibleColumns = ref([...props.visibleColumns])

// 监听外部 visible 变化
watch(() => props.modelValue, (newVal) => {
  visible.value = newVal
  if (newVal) {
    // 打开对话框时，重置为当前的 visibleColumns
    localVisibleColumns.value = [...props.visibleColumns]
  }
})

// 监听内部 visible 变化
watch(visible, (newVal) => {
  emit('update:modelValue', newVal)
})

// 监听 visibleColumns 变化
watch(() => props.visibleColumns, (newVal) => {
  if (!visible.value) {
    localVisibleColumns.value = [...newVal]
  }
}, { deep: true })

// 重置
const handleReset = () => {
  localVisibleColumns.value = [...props.defaultVisibleColumns]
}

const handleClose = () => {
  visible.value = false
  // 关闭时恢复为原始值
  localVisibleColumns.value = [...props.visibleColumns]
}

const handleConfirm = () => {
  emit('confirm', [...localVisibleColumns.value])
  visible.value = false
}
</script>

<style scoped>
.popover-content {
  min-width: 260px;
}

.popover-header {
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e4e7ed;
}

.popover-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.column-list {
  max-height: 300px;
  overflow-y: auto;
  padding: 4px 0;
  margin-bottom: 12px;
}

.column-list::-webkit-scrollbar {
  width: 6px;
}

.column-list::-webkit-scrollbar-track {
  background: #f5f5f5;
  border-radius: 3px;
}

.column-list::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.column-list::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.column-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  margin-bottom: 4px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 4px;
  transition: all 0.3s ease;
  cursor: pointer;
}

.column-item:hover {
  background: #e9ecef;
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.column-item-disabled {
  background: #f5f7fa;
  cursor: not-allowed;
  opacity: 0.7;
}

.column-item-disabled:hover {
  background: #f5f7fa;
  border-color: #e9ecef;
  box-shadow: none;
}

.column-checkbox {
  flex: 1;
  margin: 0;
}

.column-checkbox :deep(.el-checkbox__label) {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
  padding-left: 8px;
}

.column-title {
  display: inline-block;
  line-height: 1.5;
}

.required-tag {
  margin-left: 8px;
  font-size: 11px;
}

.popover-tips {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  padding: 8px 12px;
  background: #ecf5ff;
  border: 1px solid #b3d8ff;
  border-radius: 4px;
  color: #409eff;
  font-size: 12px;
}

.tips-icon {
  margin-right: 6px;
  font-size: 14px;
}

.popover-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px solid #e4e7ed;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .checkbox-group {
    grid-template-columns: 1fr;
  }
}
</style>

<style>
/* 列设置弹窗样式 */
.column-setting-popover {
  padding: 12px;
}

/* 列设置弹窗夜间模式适配 - 需要非 scoped 样式来覆盖组件内部样式 */
.dark-mode .column-setting-popover {
  background-color: var(--bg-color) !important;
  border-color: var(--border-color-base) !important;
}

.dark-mode .column-item {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--border-color-light) !important;
  color: var(--text-color-primary) !important;
}

.dark-mode .column-item:hover {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--sidebar-active) !important;
}

.dark-mode .column-item-disabled {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--border-color-light) !important;
  opacity: 0.7;
}

.dark-mode .column-item-disabled:hover {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--border-color-light) !important;
}

.dark-mode .column-checkbox :deep(.el-checkbox__label) {
  color: var(--text-color-primary) !important;
}

.dark-mode .column-title {
  color: var(--text-color-primary) !important;
}

.dark-mode .column-item * {
  color: var(--text-color-primary) !important;
}

.dark-mode .column-item-disabled * {
  color: var(--text-color-regular) !important;
}

.dark-mode .popover-header {
  border-bottom-color: var(--border-color-base) !important;
}

.dark-mode .popover-title {
  color: var(--text-color-primary) !important;
}

.dark-mode .popover-tips {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--border-color-light) !important;
  color: var(--sidebar-active) !important;
}

.dark-mode .popover-tips * {
  color: var(--sidebar-active) !important;
}

.dark-mode .popover-footer {
  border-top-color: var(--border-color-base) !important;
}

.dark-mode .column-list::-webkit-scrollbar-track {
  background: var(--bg-color-tertiary) !important;
}

.dark-mode .column-list::-webkit-scrollbar-thumb {
  background: var(--border-color-base) !important;
}

.dark-mode .column-list::-webkit-scrollbar-thumb:hover {
  background: var(--text-color-secondary) !important;
}
</style>

