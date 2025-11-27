<template>
  <el-form
    ref="formRef"
    :model="model"
    :inline="inline"
    :label-width="labelWidth"
    :label-position="labelPosition"
    :rules="rules"
    :class="['search-form', { 'search-form-compact': compact }]"
    :style="formStyle"
  >
    <!-- 通过配置生成表单 -->
    <template v-if="fields && fields.length > 0">
      <template v-for="field in fields" :key="field.prop">
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
          
          <!-- 选择器 -->
          <el-select
            v-else-if="field.type === 'select'"
            v-model="model[field.prop]"
            :placeholder="getFieldPlaceholder(field)"
            :clearable="field.clearable !== false"
            :disabled="field.disabled"
            :multiple="field.multiple"
            :filterable="field.filterable"
            :style="{ width: field.width || '150px' }"
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
    </template>
    
    <!-- 插槽方式（向后兼容） -->
    <template v-else>
      <!-- 基础搜索项（始终显示） -->
      <slot />
      <!-- 高级搜索项（可展开/收起） -->
      <template v-if="hasAdvancedSlot">
        <slot name="advanced" :expanded="expanded" />
      </template>
    </template>
    
    <!-- 展开/收起按钮（当有高级搜索项时显示） -->
    <el-form-item v-if="hasAdvancedFields && showExpandButton" class="expand-item">
      <el-button
        :type="expandButtonType"
        :plain="expandButtonPlain"
        :size="buttonSize"
        @click="toggleExpand"
      >
        <el-icon><component :is="expanded ? ArrowUp : ArrowDown" /></el-icon>
        {{ expanded ? collapseText : expandText }}
      </el-button>
    </el-form-item>
    
    <!-- 操作按钮 -->
    <el-form-item class="action-item">
      <el-button
        type="primary"
        :size="buttonSize"
        :loading="loading"
        :icon="searchIcon"
        @click="handleSearch"
      >
        {{ searchText }}
      </el-button>
      <el-button
        :size="buttonSize"
        :icon="resetIcon"
        @click="handleReset"
      >
        {{ resetText }}
      </el-button>
      <slot name="extra-buttons" />
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref, useSlots, computed, watch } from 'vue'
import { Search, Refresh, ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

// 防抖函数
const debounce = (fn, delay) => {
  let timer = null
  return function(...args) {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      fn.apply(this, args)
    }, delay)
  }
}

const props = defineProps({
  // 表单数据模型
  model: {
    type: Object,
    required: true
  },
  // 字段配置（JSON 配置方式）
  fields: {
    type: Array,
    default: () => []
  },
  // 表单验证规则
  rules: {
    type: Object,
    default: () => ({})
  },
  // 是否行内表单
  inline: {
    type: Boolean,
    default: true
  },
  // 标签宽度
  labelWidth: {
    type: [String, Number],
    default: ''
  },
  // 标签位置
  labelPosition: {
    type: String,
    default: 'right',
    validator: (value) => ['left', 'right', 'top'].includes(value)
  },
  // 默认展开状态
  defaultExpanded: {
    type: Boolean,
    default: false
  },
  // 是否显示展开按钮
  showExpandButton: {
    type: Boolean,
    default: true
  },
  // 展开按钮类型
  expandButtonType: {
    type: String,
    default: 'default'
  },
  // 展开按钮是否朴素按钮
  expandButtonPlain: {
    type: Boolean,
    default: false
  },
  // 按钮尺寸
  buttonSize: {
    type: String,
    default: 'default',
    validator: (value) => ['large', 'default', 'small'].includes(value)
  },
  // 搜索按钮文本
  searchText: {
    type: String,
    default: ''
  },
  // 重置按钮文本
  resetText: {
    type: String,
    default: ''
  },
  // 展开按钮文本
  expandText: {
    type: String,
    default: ''
  },
  // 收起按钮文本
  collapseText: {
    type: String,
    default: ''
  },
  // 是否紧凑模式
  compact: {
    type: Boolean,
    default: false
  },
  // 加载状态
  loading: {
    type: Boolean,
    default: false
  },
  // 是否启用搜索防抖
  debounce: {
    type: Boolean,
    default: false
  },
  // 防抖延迟时间（毫秒）
  debounceDelay: {
    type: Number,
    default: 300
  },
  // 自定义样式
  formStyle: {
    type: Object,
    default: () => ({})
  },
  // 初始值（用于重置）
  initialValues: {
    type: Object,
    default: () => ({})
  },
  // 国际化前缀（用于自动翻译 label 和 placeholder）
  i18nPrefix: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['search', 'reset', 'expand-change', 'validate'])

const slots = useSlots()
const formRef = ref(null)
const expanded = ref(props.defaultExpanded)

// 国际化文本
const { t } = useI18n()

// 检查是否有高级搜索项插槽
const hasAdvancedSlot = computed(() => {
  return !!slots.advanced
})

// 检查是否有高级搜索字段
const hasAdvancedFields = computed(() => {
  if (props.fields && props.fields.length > 0) {
    return props.fields.some(field => field.advanced === true)
  }
  return hasAdvancedSlot.value
})

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

// 获取字段选项
const getFieldOptions = (field) => {
  if (!field) return []
  if (field.options && Array.isArray(field.options)) {
    return field.options
  }
  if (field.optionsFn && typeof field.optionsFn === 'function') {
    try {
      return field.optionsFn()
    } catch (e) {
      console.warn('Error in optionsFn:', e)
      return []
    }
  }
  return []
}

// 获取字段样式
const getFieldStyle = (field) => {
  return field.style || {}
}

const searchText = computed(() => {
  return props.searchText || t('log.search') || '搜索'
})

const resetText = computed(() => {
  return props.resetText || t('log.reset') || '重置'
})

const expandText = computed(() => {
  return props.expandText || t('log.expand') || '展开'
})

const collapseText = computed(() => {
  return props.collapseText || t('log.collapse') || '收起'
})

const searchIcon = computed(() => {
  return props.searchText ? undefined : Search
})

const resetIcon = computed(() => {
  return props.resetText ? undefined : Refresh
})

// 切换展开/收起
const toggleExpand = () => {
  expanded.value = !expanded.value
  emit('expand-change', expanded.value)
}

// 搜索处理（支持防抖）
const doSearch = () => {
  if (formRef.value && Object.keys(props.rules).length > 0) {
    formRef.value.validate((valid) => {
      if (valid) {
        emit('search', props.model)
      } else {
        emit('validate', false)
      }
    })
  } else {
    emit('search', props.model)
  }
}

const handleSearch = props.debounce
  ? debounce(doSearch, props.debounceDelay)
  : doSearch

// 重置处理
const handleReset = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  
  // 重置为初始值
  if (Object.keys(props.initialValues).length > 0) {
    Object.keys(props.model).forEach((key) => {
      if (props.initialValues.hasOwnProperty(key)) {
        props.model[key] = props.initialValues[key]
      } else {
        // 根据类型设置默认值
        const value = props.model[key]
        if (Array.isArray(value)) {
          props.model[key] = []
        } else if (typeof value === 'number') {
          props.model[key] = 0
        } else if (typeof value === 'boolean') {
          props.model[key] = false
        } else {
          props.model[key] = ''
        }
      }
    })
  } else {
    // 如果没有初始值，清空所有字段
    Object.keys(props.model).forEach((key) => {
      const value = props.model[key]
      if (Array.isArray(value)) {
        props.model[key] = []
      } else if (typeof value === 'number') {
        props.model[key] = 0
      } else if (typeof value === 'boolean') {
        props.model[key] = false
      } else {
        props.model[key] = ''
      }
    })
  }
  
  // 收起高级搜索
  if (expanded.value) {
    expanded.value = false
    emit('expand-change', false)
  }
  
  emit('reset', props.model)
}

// 监听初始值变化
watch(() => props.initialValues, (newVal) => {
  if (newVal && Object.keys(newVal).length > 0) {
    Object.keys(newVal).forEach((key) => {
      if (props.model.hasOwnProperty(key)) {
        props.model[key] = newVal[key]
      }
    })
  }
}, { deep: true, immediate: true })

// 暴露方法
defineExpose({
  formRef,
  expanded,
  validate: () => formRef.value?.validate(),
  resetFields: () => {
    formRef.value?.resetFields()
    handleReset()
  },
  clearValidate: () => formRef.value?.clearValidate(),
  toggleExpand
})
</script>

<style scoped lang="scss">
.search-form {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
  transition: all 0.3s ease;

  &.search-form-compact {
    padding: 15px;
    margin-bottom: 15px;
  }

  :deep(.el-form-item) {
    margin-bottom: 18px;

    &:last-child {
      margin-bottom: 0;
    }
  }

  .expand-item {
    margin-left: 10px;
  }

  .action-item {
    margin-left: 10px;
    flex: 1;
    display: flex;
    justify-content: flex-end;
  }

  // 响应式布局
  @media (max-width: 768px) {
    padding: 15px;

    :deep(.el-form-item) {
      width: 100%;
      margin-right: 0;
    }

    .action-item {
      width: 100%;
      justify-content: flex-start;
      margin-left: 0;
      margin-top: 10px;
    }
  }
}
</style>
