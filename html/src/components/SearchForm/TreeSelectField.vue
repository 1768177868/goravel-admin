<template>
  <el-popover
    placement="bottom-start"
    :width="field.popoverWidth || 300"
    :visible="popoverVisible"
    @update:visible="(val) => { if (val !== undefined) updatePopoverVisible(val) }"
    :popper-options="{ 
      modifiers: [
        { name: 'computeStyles', options: { gpuAcceleration: false } },
        { name: 'preventOverflow', options: { boundary: 'viewport' } }
      ]
    }"
  >
    <template #reference>
      <el-input
        :model-value="inputValue"
        :placeholder="placeholder"
        :clearable="field.clearable !== false"
        :disabled="field.disabled"
        :style="{ width: field.width || '200px' }"
        @clear="(e) => { e?.stopPropagation(); handleClear() }"
        @input="handleInput"
        @focus="!field.disabled && !popoverVisible && togglePopover()"
        @click.stop.prevent="!field.disabled && togglePopover()"
        @mousedown.stop.prevent
      >
        <template #suffix>
          <el-icon 
            v-if="!field.disabled"
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
      <el-input
        v-if="field.filterable !== false"
        v-model="filterText"
        :placeholder="t('common.search') || '搜索'"
        clearable
        size="small"
        style="margin-bottom: 8px;"
        @input="() => {}"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
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
import { ArrowDown, Search } from '@element-plus/icons-vue'
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

// 监听 field 变化，重新加载数据
watch(() => props.field.apiUrl, () => {
  if (props.field.apiUrl) {
    loadData()
  }
}, { immediate: true })

const handleNodeClick = (data) => {
  handleTreeSelectNodeClick(data)
  updatePopoverVisible(false)
}

const handleClear = () => {
  handleTreeSelectClear()
}

const handleInput = (val) => {
  handleTreeSelectInput(val)
}

// 监听 field 变化，重新加载数据
watch(() => props.field.apiUrl, () => {
  if (props.field.apiUrl) {
    loadData()
  }
}, { immediate: true })
</script>

<style scoped>
:deep(.el-input__icon.is-reverse) {
  transform: rotate(180deg);
}
</style>

