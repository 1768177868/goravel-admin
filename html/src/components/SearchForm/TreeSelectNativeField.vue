<template>
  <el-tree-select
    v-model="innerValue"
    :data="treeData"
    :props="treeProps"
    :node-key="nodeKey"
    :placeholder="placeholder"
    :clearable="field.clearable !== false"
    :disabled="field.disabled"
    :style="{ width: field.width || '200px' }"
    :filterable="field.filterable !== false"
    :check-strictly="field.checkStrictly !== false"
    :render-after-expand="field.renderAfterExpand !== false"
    :default-expand-all="field.defaultExpandAll || false"
    :show-checkbox="field.showCheckbox || false"
    :multiple="field.multiple || false"
    :popper-class="field.popperClass || 'el-tree-select__popper'"
    v-bind="field.props || {}"
  />
</template>

<script setup>
import { computed, watch } from 'vue'
import { useTreeSelect } from './useTreeSelect'

const props = defineProps({
  field: {
    type: Object,
    required: true
  },
  modelValue: {
    type: [String, Number, Array],
    default: null
  },
  placeholder: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const innerValue = computed({
  get: () => props.modelValue,
  set: (value) => {
    emit('update:modelValue', value)
    emit('change', value)
  }
})

const treeProps = computed(() => props.field.treeProps || { label: 'name', value: 'id', children: 'children' })
const nodeKey = computed(() => treeProps.value.value || 'id')

const { treeData, loadData } = useTreeSelect({
  field: props.field,
  modelValue: computed(() => props.modelValue),
  onUpdate: (value) => {
    emit('update:modelValue', value)
    emit('change', value)
  }
})

watch(() => props.field.apiUrl, () => {
  if (props.field.apiUrl) loadData()
}, { immediate: true })

watch(() => {
  if (typeof props.field.treeData === 'function') return props.field.treeData()
  return props.field.treeData
}, () => {
  if (props.field.apiUrl) loadData()
}, { deep: true, immediate: true })
</script>
