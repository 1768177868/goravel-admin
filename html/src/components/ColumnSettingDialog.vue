<template>
  <el-dialog
    v-model="visible"
    :title="$t('common.column_setting')"
    width="600px"
    @close="handleClose"
    class="column-setting-dialog"
  >
    <div class="dialog-content">
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
      <div class="dialog-tips">
        <el-icon class="tips-icon"><InfoFilled /></el-icon>
        <span>{{ $t('common.column_setting_tip') }}</span>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
        <el-button @click="handleClose">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleConfirm">{{ $t('common.confirm') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { InfoFilled } from '@element-plus/icons-vue'

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
.column-setting-dialog :deep(.el-dialog__body) {
  padding: 20px 24px;
}

.dialog-content {
  min-height: 200px;
}

.column-list {
  max-height: 400px;
  overflow-y: auto;
  padding: 8px 0;
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
  padding: 12px 16px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 6px;
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

.dialog-tips {
  display: flex;
  align-items: center;
  margin-top: 16px;
  padding: 12px 16px;
  background: #ecf5ff;
  border: 1px solid #b3d8ff;
  border-radius: 6px;
  color: #409eff;
  font-size: 13px;
}

.tips-icon {
  margin-right: 8px;
  font-size: 16px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .checkbox-group {
    grid-template-columns: 1fr;
  }
  
  .column-setting-dialog :deep(.el-dialog) {
    width: 90% !important;
  }
}
</style>

<style>
/* 列设置弹窗夜间模式适配 - 需要非 scoped 样式来覆盖组件内部样式 */
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

.dark-mode .dialog-tips {
  background-color: var(--bg-color-tertiary) !important;
  border-color: var(--border-color-light) !important;
  color: var(--sidebar-active) !important;
}

.dark-mode .dialog-tips * {
  color: var(--sidebar-active) !important;
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

