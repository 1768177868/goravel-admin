<template>
  <el-popover
    placement="bottom-start"
    :width="field.popoverWidth || 300"
    :visible="popoverVisible"
    trigger="manual"
    @update:visible="(val) => { 
      // 只有在明确设置为 false 时才关闭，避免点击外部时意外关闭
      if (val === false) {
        updatePopoverVisible(false)
      }
    }"
    :popper-options="{ 
      modifiers: [
        { name: 'computeStyles', options: { gpuAcceleration: false } },
        { name: 'preventOverflow', options: { boundary: 'viewport' } }
      ]
    }"
  >
    <template #reference>
      <el-input
        :model-value="displayInputValue"
        :placeholder="placeholder"
        :clearable="field.clearable !== false && (!!modelValue || !!filterText)"
        :disabled="field.disabled"
        :style="{ width: field.width || '200px' }"
        @clear="handleClear"
        @input="handleInput"
        @focus="handleInputFocus"
        @click="handleInputClick"
        style="cursor: text"
      >
        <template #suffix>
          <el-icon 
            v-if="!field.disabled && !modelValue && !inputValue"
            class="el-input__icon"
            :class="{ 'is-reverse': popoverVisible }"
            style="transition: transform 0.3s; pointer-events: none;"
          >
            <ArrowDown />
          </el-icon>
        </template>
      </el-input>
    </template>
    <div @click.stop @mousedown.stop>
      <el-tree
        :data="treeData"
        :props="field.treeProps || { label: 'name', children: 'children' }"
        node-key="id"
        :default-expand-all="field.filterable !== false && filterText ? true : (field.defaultExpandAll || false)"
        :expand-on-click-node="false"
        :highlight-current="true"
        @node-click="handleNodeClick"
        :style="{ maxHeight: field.treeHeight || '300px', overflowY: 'auto' }"
      >
        <template #default="{ node, data }">
          <span class="custom-tree-node" style="flex: 1; display: flex; align-items: center; justify-content: space-between; font-size: 14px; padding-right: 8px;">
            <span>{{ node.label }}</span>
          </span>
        </template>
      </el-tree>
    </div>
  </el-popover>
</template>

<script setup>
import { computed, watch } from 'vue'
import { ArrowDown } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { useTreeSelect } from './useTreeSelect'

const props = defineProps({
  field: {
    type: Object,
    required: true
  },
  modelValue: {
    type: [String, Number],
    default: null
  },
  placeholder: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const { t } = useI18n()

const {
  popoverVisible,
  filterText,
  treeData,
  inputValue,
  updatePopoverVisible,
  togglePopover,
  handleNodeClick: handleTreeSelectNodeClick,
  handleClear: handleTreeSelectClear,
  handleInput: handleTreeSelectInput,
  loadData
} = useTreeSelect({
  field: props.field,
  modelValue: computed(() => props.modelValue),
  onUpdate: (value) => {
    emit('update:modelValue', value)
    emit('change', value)
  }
})

// 计算输入框显示值
const displayInputValue = computed(() => {
  // 如果有输入文本，优先显示输入文本
  if (filterText.value) {
    return filterText.value
  }
  // 如果有选中值，显示选中值的标签（不管弹窗是否打开）
  if (props.modelValue) {
    return inputValue.value
  }
  return ''
})

// 监听 field.apiUrl 变化，重新加载数据
watch(() => props.field.apiUrl, () => {
  if (props.field.apiUrl) {
    loadData()
  }
}, { immediate: true })

// 监听 field.treeData 变化（支持 getter 函数）
watch(() => {
  if (typeof props.field.treeData === 'function') {
    return props.field.treeData()
  }
  return props.field.treeData
}, (newTreeData) => {
  if (newTreeData && Array.isArray(newTreeData) && newTreeData.length > 0) {
    loadData()
  }
}, { deep: true, immediate: true })

const handleNodeClick = (data) => {
  // 获取节点的标签（优先使用 label，否则使用 name）
  const labelKey = props.field.treeProps?.label || 'label'
  const nameKey = props.field.treeProps?.name || 'name'
  const nodeLabel = data[labelKey] || data[nameKey] || ''
  
  // 先更新值
  handleTreeSelectNodeClick(data)
  
  // 清空 filterText，确保显示选中值
  filterText.value = ''
  
  // 然后关闭弹窗（延迟一下，确保值更新完成）
  setTimeout(() => {
    updatePopoverVisible(false)
  }, 10)
}

const handleClear = (e) => {
  e?.stopPropagation()
  handleTreeSelectClear()
}

const handleInput = (val) => {
  const inputVal = val || ''
  
  // 如果输入为空，且有选中值，清空选中值
  if (props.modelValue && !inputVal) {
    handleTreeSelectClear()
    return
  }

  // 如果当前有选中值，且输入内容发生变化
  if (props.modelValue && inputVal) {
    const currentDisplayValue = inputValue.value
    if (inputVal !== currentDisplayValue) {
      handleTreeSelectClear()
      filterText.value = inputVal // 恢复输入内容
    } else {
      filterText.value = inputVal
    }
  } else {
    filterText.value = inputVal
  }
  
  // 输入时自动打开弹窗
  if (inputVal && !popoverVisible.value && !props.field.disabled) {
    togglePopover()
  }
}

const handleInputFocus = (e) => {
  // 获得焦点时，如果有选中值，全选文本方便用户修改
  if (props.modelValue && !filterText.value) {
    // 延迟一下确保 DOM 更新
    setTimeout(() => {
      if (e.target && e.target.select) {
        e.target.select()
      }
    }, 50)
  }
}

const handleInputClick = (e) => {
  // 如果点击的是清除按钮，不打开弹窗
  if (e.target.closest('.el-input__clear')) {
    return
  }
  // 打开弹窗
  if (!popoverVisible.value && !props.field.disabled) {
    togglePopover()
  }
}
</script>

<style scoped>
:deep(.el-input__icon.is-reverse) {
  transform: rotate(180deg);
}
</style>

