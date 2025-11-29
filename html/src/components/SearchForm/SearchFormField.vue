<template>
  <el-form-item
    v-if="(!field.advanced || expanded) && field.prop"
    :label="getFieldLabel(field)"
    :prop="field.prop"
    :style="getFieldStyle(field)"
  >
    <!-- 输入框 -->
    <el-input
      v-if="field.type === 'input' && field.prop"
      v-model="model[field.prop]"
      :placeholder="getFieldPlaceholder(field)"
      :clearable="field.clearable !== false"
      :disabled="field.disabled"
      :style="{ width: field.width || '200px' }"
      v-bind="field.props || {}"
    />
    
    <!-- 文本域 -->
    <el-input
      v-else-if="field.type === 'textarea'"
      v-model="model[field.prop]"
      type="textarea"
      :placeholder="getFieldPlaceholder(field)"
      :clearable="field.clearable !== false"
      :disabled="field.disabled"
      :rows="field.rows || 3"
      :style="{ width: field.width || '200px' }"
      v-bind="field.props || {}"
    />
    
    <!-- 树形选择器 -->
    <TreeSelectField
      v-else-if="field.type === 'tree-select'"
      :field="field"
      :model-value="model[field.prop]"
      :placeholder="getFieldPlaceholder(field)"
      @update:model-value="model[field.prop] = $event"
    />
    
    <!-- 选择器 -->
    <el-select
      v-else-if="field.type === 'select'"
      v-model="model[field.prop]"
      :placeholder="getFieldPlaceholder(field)"
      :clearable="field.clearable !== false"
      :disabled="field.disabled"
      :multiple="field.multiple"
      :filterable="field.filterable"
      :style="{ width: field.width || '200px' }"
      v-bind="field.props || {}"
    >
      <el-option
        v-for="option in getFieldOptions(field)"
        :key="option.value"
        :label="option.label"
        :value="option.value"
        :disabled="option.disabled"
      />
    </el-select>
    
    <!-- 日期选择器 -->
    <el-date-picker
      v-else-if="field.type === 'date' || field.type === 'datetime' || field.type === 'daterange' || field.type === 'datetimerange'"
      v-model="model[field.prop]"
      :type="field.type === 'date' ? 'date' : field.type === 'datetime' ? 'datetime' : field.type === 'daterange' ? 'daterange' : 'datetimerange'"
      :placeholder="getFieldPlaceholder(field)"
      :clearable="field.clearable !== false"
      :disabled="field.disabled"
      :value-format="field.valueFormat || (field.type === 'datetime' || field.type === 'datetimerange' ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD')"
      :style="{ width: field.width || '180px' }"
      v-bind="field.props || {}"
    />
    
    <!-- 数字输入框 -->
    <el-input-number
      v-else-if="field.type === 'number'"
      v-model="model[field.prop]"
      :placeholder="getFieldPlaceholder(field)"
      :disabled="field.disabled"
      :min="field.min"
      :max="field.max"
      :step="field.step"
      :style="{ width: field.width || '150px' }"
      v-bind="field.props || {}"
    />
    
    <!-- 开关 -->
    <el-switch
      v-else-if="field.type === 'switch'"
      v-model="model[field.prop]"
      :disabled="field.disabled"
      v-bind="field.props || {}"
    />
    
    <!-- 自定义插槽 -->
    <slot
      v-else-if="field.type === 'slot'"
      :name="field.slotName || field.prop"
      :field="field"
      :model="model"
    />
  </el-form-item>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import TreeSelectField from './TreeSelectField.vue'
import { useFieldOptions } from './useFieldOptions'

const props = defineProps({
  field: {
    type: Object,
    required: true
  },
  model: {
    type: Object,
    required: true
  },
  expanded: {
    type: Boolean,
    default: false
  },
  i18nPrefix: {
    type: String,
    default: ''
  }
})

const { t } = useI18n()
const { getFieldOptions: getOptions } = useFieldOptions()

// 获取字段标签
const getFieldLabel = (field) => {
  if (!field) return ''
  if (field.label) {
    // 如果 label 是翻译键
    if (typeof field.label === 'string' && (field.label.startsWith('$t(') || (props.i18nPrefix && field.labelKey))) {
      const key = field.labelKey || field.label.replace('$t(', '').replace(')', '')
      return t(props.i18nPrefix ? `${props.i18nPrefix}.${key}` : key)
    }
    return field.label
  }
  // 尝试自动翻译
  if (props.i18nPrefix && field.prop) {
    const key = `${props.i18nPrefix}.${field.prop}`
    const translated = t(key)
    return translated !== key ? translated : field.prop
  }
  return field.prop || ''
}

// 获取字段占位符
const getFieldPlaceholder = (field) => {
  if (field.placeholder) {
    if (field.placeholder.startsWith('$t(') || (props.i18nPrefix && field.placeholderKey)) {
      const key = field.placeholderKey || field.placeholder.replace('$t(', '').replace(')', '')
      return t(props.i18nPrefix ? `${props.i18nPrefix}.${key}` : key)
    }
    return field.placeholder
  }
  // 自动生成占位符
  const label = getFieldLabel(field)
  if (field.type === 'select') {
    return t('form.please_select') + label
  }
  return t('form.please_enter') + label
}

const getFieldStyle = (field) => {
  return field.style || {}
}

const getFieldOptions = (field) => {
  return getOptions(field)
}
</script>

