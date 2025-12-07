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
    <div 
      class="form-fields-wrapper" 
      :class="{ 'form-fields-collapsed': computedShouldShowExpandButton && !expanded }"
      :style="computedFieldsWrapperStyle"
    >
      <template v-if="fields && fields.length > 0">
        <SearchFormField
          v-for="field in fields"
          :key="field.prop"
          :field="field"
          :model="model"
          :expanded="expanded"
          :i18n-prefix="i18nPrefix"
        >
          <template v-for="(_, slotName) in $slots" #[slotName]="slotProps">
            <slot :name="slotName" v-bind="slotProps" />
          </template>
        </SearchFormField>
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
    </div>
    
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
      <!-- 展开/收起按钮（移到重置按钮后面，根据表单高度自动判断显示） -->
      <el-button
        v-if="computedShouldShowExpandButton"
        :type="expandButtonType"
        :plain="expandButtonPlain"
        :size="buttonSize"
        @click="toggleExpand"
      >
        <el-icon><component :is="expanded ? ArrowUp : ArrowDown" /></el-icon>
        {{ expanded ? collapseText : expandText }}
      </el-button>
      <slot name="extra-buttons" />
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref, useSlots, computed, watch, onMounted, onUnmounted } from 'vue'
import { Search, Refresh, ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import SearchFormField from './SearchForm/SearchFormField.vue'
import { useFormHeight } from './SearchForm/useFormHeight'
import { useFieldOptions } from './SearchForm/useFieldOptions'

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
const { loadFieldOptions } = useFieldOptions()
let resizeObserver = null

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

// 使用表单高度检测 composable
const { singleLineHeight, shouldShowExpandButton, checkFormHeight } = useFormHeight(
  formRef,
  expanded,
  hasAdvancedFields,
  computed(() => props.defaultExpanded)
)

// 计算是否应该显示展开按钮（根据表单高度自动判断）
const computedShouldShowExpandButton = computed(() => {
  // 如果手动设置了 showExpandButton 为 false，则不显示
  if (props.showExpandButton === false) {
    return false
  }
  // 如果有高级搜索字段，使用原来的逻辑
  if (hasAdvancedFields.value) {
    return props.showExpandButton !== false
  }
  // 否则根据表单高度自动判断
  return shouldShowExpandButton.value
})

// 计算表单字段容器的样式（动态设置 max-height）
const computedFieldsWrapperStyle = computed(() => {
  if (computedShouldShowExpandButton.value && !expanded.value) {
    // 收起状态：尽可能多地显示表单项，但不超过一行的高度
    if (singleLineHeight.value > 0) {
      return {
        maxHeight: `${singleLineHeight.value}px`
      }
    }
  }
  return {}
})

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
  
  if (expanded.value) {
    expanded.value = false
    emit('expand-change', false)
  }
  
  emit('reset', props.model)
}

watch(() => props.initialValues, (newVal) => {
  if (newVal && Object.keys(newVal).length > 0) {
    Object.keys(newVal).forEach((key) => {
      if (props.model.hasOwnProperty(key)) {
        props.model[key] = newVal[key]
      }
    })
  }
}, { deep: true, immediate: true })

watch(() => expanded.value, () => {
  setTimeout(() => {
    checkFormHeight()
  }, 300)
})

watch(() => props.fields, (newFields) => {
  checkFormHeight()
  
  if (newFields && Array.isArray(newFields)) {
    newFields.forEach(field => {
      if (field.apiUrl) {
        if (field.type === 'tree-select') {
          // TreeSelectField 会自动处理
        } else if (field.type === 'select') {
          loadFieldOptions(field)
        }
      }
    })
  }
}, { deep: true })

onMounted(() => {
  checkFormHeight()
  
  if (formRef.value && formRef.value.$el) {
    resizeObserver = new ResizeObserver(() => {
      checkFormHeight()
    })
    resizeObserver.observe(formRef.value.$el)
  }
  
  if (props.fields && Array.isArray(props.fields)) {
    props.fields.forEach(field => {
      if (field.apiUrl && field.type === 'select') {
        loadFieldOptions(field)
      }
    })
  }
  
  setTimeout(() => {
    checkFormHeight()
  }, 100)
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
})

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
  background: var(--bg-color-tertiary, #f5f7fa);
  border-radius: 4px;
  transition: all 0.3s ease;

  &.search-form-compact {
    padding: 15px;
    margin-bottom: 15px;
  }

  :deep(.el-form-item) {
    margin-bottom: 18px;
    margin-right: 10px; // 添加右边距，确保表单项之间有间距

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

  // 表单字段容器
  .form-fields-wrapper {
    transition: max-height 0.3s ease, margin-bottom 0.3s ease;
    overflow: hidden;
    margin-bottom: 0;
    
    // 收起状态：尽可能多地显示表单项（max-height 通过 computedFieldsWrapperStyle 动态设置）
    &.form-fields-collapsed {
      // 确保表单项对齐，使用 flex 布局
      display: flex;
      flex-wrap: wrap;
      align-items: flex-start;
    }
    
    // 展开状态：显示所有内容，并添加底部间距
    &:not(.form-fields-collapsed) {
      max-height: none;
      margin-bottom: 18px; // 展开后添加底部间距，避免贴着按钮
      display: flex;
      flex-wrap: wrap;
      align-items: flex-start;
    }
  }

  // 操作按钮区域，确保有合适的间距
  .action-item {
    margin-top: 0;
    margin-bottom: 0;
  }
}
</style>
