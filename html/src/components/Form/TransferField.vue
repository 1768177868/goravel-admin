<template>
  <el-transfer
    :model-value="normalizedValue"
    :data="transferOptions"
    :titles="titles"
    :filterable="field.filterable !== false"
    :filter-placeholder="field.filterPlaceholder || ''"
    :props="transferProps"
    :disabled="field.disabled"
    v-bind="field.props || {}"
    @update:model-value="emit('update:modelValue', $event)"
  />
</template>

<script setup>
import { computed, onMounted, watch } from 'vue'
import { useFieldOptions } from '../SearchForm/useFieldOptions'

const props = defineProps({
  field: {
    type: Object,
    required: true
  },
  modelValue: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue'])

const { loadFieldOptions, getFieldOptions } = useFieldOptions()

const transferProps = computed(() => ({
  key: props.field.optionValueKey || 'value',
  label: props.field.optionLabelKey || 'label',
  disabled: 'disabled',
  ...(props.field.transferProps || {})
}))

const titles = computed(() => props.field.titles || ['Source', 'Target'])

const normalizedValue = computed(() => {
  return Array.isArray(props.modelValue) ? props.modelValue : []
})

const transferOptions = computed(() => {
  const options = getFieldOptions(props.field)
  if (!Array.isArray(options)) return []
  return options
})

function ensureOptionsLoaded() {
  const f = props.field
  if (!f?.apiUrl) return
  loadFieldOptions(f)
}

onMounted(ensureOptionsLoaded)
watch(() => [props.field?.apiUrl, props.field?.type], ensureOptionsLoaded)
</script>
