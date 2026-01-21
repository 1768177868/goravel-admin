<template>
  <el-form-item
    v-if="field.prop && resolveVisible(field)"
    :label="getFieldLabel(field)"
    :prop="field.prop"
    :style="field.style || {}"
  >
    <!-- 输入框（inputType: text|password|number 等，透传至 el-input type） -->
    <el-input
      v-if="field.type === 'input' || field.type === 'password'"
      v-model="model[field.prop]"
      :type="field.type === 'password' ? 'password' : (field.inputType || 'text')"
      :placeholder="getPlaceholder(field)"
      :clearable="field.clearable !== false"
      :show-password="field.type === 'password' && (field.showPassword !== false)"
      :disabled="field.disabled"
      v-bind="field.props || {}"
    />

    <!-- 文本域 -->
    <el-input
      v-else-if="field.type === 'textarea'"
      v-model="model[field.prop]"
      type="textarea"
      :placeholder="getPlaceholder(field)"
      :clearable="field.clearable !== false"
      :disabled="field.disabled"
      :rows="field.rows || 3"
      v-bind="field.props || {}"
    />

    <!-- 树形选择（支持 apiUrl / treeData / options，与 SearchForm 一致） -->
    <TreeSelectField
      v-else-if="field.type === 'tree-select'"
      :field="field"
      :model-value="model[field.prop]"
      :placeholder="getPlaceholder(field)"
      @update:model-value="model[field.prop] = $event"
    />

    <!-- 下拉选择（支持 apiUrl / options / optionsFn） -->
    <el-select
      v-else-if="field.type === 'select'"
      v-model="model[field.prop]"
      :placeholder="getPlaceholder(field)"
      :clearable="field.clearable !== false"
      :disabled="field.disabled"
      :multiple="field.multiple"
      :filterable="field.filterable"
      v-bind="field.props || {}"
    >
      <el-option
        v-for="opt in getFieldOptions(field)"
        :key="String(opt.value)"
        :label="opt.label"
        :value="opt.value"
        :disabled="opt.disabled"
      />
    </el-select>

    <!-- 单选组（支持 apiUrl / options / optionsFn） -->
    <el-radio-group
      v-else-if="field.type === 'radio'"
      v-model="model[field.prop]"
      :disabled="field.disabled"
      v-bind="field.props || {}"
    >
      <el-radio
        v-for="opt in getFieldOptions(field)"
        :key="String(opt.value)"
        :label="opt.value"
        :disabled="opt.disabled"
      >
        {{ opt.label }}
      </el-radio>
    </el-radio-group>

    <!-- 多选/复选框组（支持 apiUrl / options / optionsFn） -->
    <el-checkbox-group
      v-else-if="field.type === 'checkbox' || field.type === 'checkbox-group'"
      v-model="model[field.prop]"
      :disabled="field.disabled"
      v-bind="field.props || {}"
    >
      <el-checkbox
        v-for="opt in getFieldOptions(field)"
        :key="String(opt.value)"
        :label="opt.value"
        :disabled="opt.disabled"
      >
        {{ opt.label }}
      </el-checkbox>
    </el-checkbox-group>

    <!-- 日期 -->
    <el-date-picker
      v-else-if="field.type === 'date' || field.type === 'datetime' || field.type === 'daterange' || field.type === 'datetimerange'"
      v-model="model[field.prop]"
      :type="field.type === 'date' ? 'date' : field.type === 'datetime' ? 'datetime' : field.type === 'daterange' ? 'daterange' : 'datetimerange'"
      :placeholder="getPlaceholder(field)"
      :clearable="field.clearable !== false"
      :disabled="field.disabled"
      :value-format="field.valueFormat || (field.type === 'datetime' || field.type === 'datetimerange' ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD')"
      v-bind="field.props || {}"
    />

    <!-- 数字输入框：新增支持 input-number 类型（与 number 等价） -->
    <el-input-number
      v-else-if="field.type === 'number' || field.type === 'input-number'"
      v-model="model[field.prop]"
      :placeholder="getPlaceholder(field)"
      :disabled="field.disabled"
      :min="field.min"
      :max="field.max"
      :step="field.step"
      v-bind="field.props || {}"
    />

    <!-- 开关 -->
    <el-switch
      v-else-if="field.type === 'switch'"
      v-model="model[field.prop]"
      :disabled="field.disabled"
      v-bind="field.props || {}"
    />

    <!-- 级联选择（支持 apiUrl / options / optionsFn，options 为树形 {value,label,children}） -->
    <el-cascader
      v-else-if="field.type === 'cascader'"
      v-model="model[field.prop]"
      :options="getFieldOptions(field)"
      :props="field.cascaderProps || { value: 'value', label: 'label', children: 'children' }"
      :placeholder="getPlaceholder(field)"
      :clearable="field.clearable !== false"
      :disabled="field.disabled"
      :show-all-levels="field.showAllLevels !== false"
      v-bind="field.props || {}"
    />

    <!-- 评分 -->
    <el-rate
      v-else-if="field.type === 'rate'"
      v-model="model[field.prop]"
      :max="field.max ?? 5"
      :disabled="field.disabled"
      v-bind="field.props || {}"
    />

    <!-- 滑动条 -->
    <el-slider
      v-else-if="field.type === 'slider'"
      v-model="model[field.prop]"
      :min="field.min"
      :max="field.max"
      :step="field.step"
      :disabled="field.disabled"
      v-bind="field.props || {}"
    />

    <!-- 颜色选择 -->
    <el-color-picker
      v-else-if="field.type === 'color'"
      v-model="model[field.prop]"
      :disabled="field.disabled"
      :show-alpha="field.showAlpha"
      v-bind="field.props || {}"
    />

    <!-- 默认插槽：自定义表单项（如富文本、上传、图标选择等） -->
    <slot v-else />
  </el-form-item>
</template>

<script setup>
/**
 * FormField 支持的 type：
 * - input, password, textarea
 * - select, radio, checkbox, checkbox-group（支持 apiUrl / options / optionsFn）
 * - tree-select（支持 apiUrl / treeData，与 SearchForm 一致）
 * - cascader（支持 apiUrl / options，树形 {value,label,children}）
 * - date, datetime, daterange, datetimerange
 * - number, switch, rate, slider, color
 * - 其它 type 走默认插槽，可自定义富文本、上传等
 */
import { onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import TreeSelectField from '../SearchForm/TreeSelectField.vue'
import { useFieldOptions } from '../SearchForm/useFieldOptions'

const props = defineProps({
  field: {
    type: Object,
    required: true
  },
  model: {
    type: Object,
    required: true
  },
  i18nPrefix: {
    type: String,
    default: ''
  }
})

const { t } = useI18n()
const { loadFieldOptions, getFieldOptions } = useFieldOptions()

function getFieldLabel(f) {
  if (f?.label) {
    if (typeof f.label === 'string' && (f.label.startsWith('$t(') || (props.i18nPrefix && f.labelKey))) {
      const key = f.labelKey || f.label.replace(/\$t\(|\)/g, '')
      return t(props.i18nPrefix ? `${props.i18nPrefix}.${key}` : key)
    }
    return f.label
  }
  if (f?.labelKey && props.i18nPrefix) return t(`${props.i18nPrefix}.${f.labelKey}`)
  return f?.prop || ''
}

// function getPlaceholder(f) {
//   if (f?.placeholder) return f.placeholder
//   const label = getFieldLabel(f)
//   const selectLike = ['select', 'radio', 'checkbox', 'checkbox-group', 'cascader', 'tree-select']
//   if (selectLike.includes(f?.type)) return (t('form.please_select') || '请选择') + label
//   return (t('form.please_enter') || '请输入') + label
// }

function getPlaceholder(f) {
  // 1. 如果是函数，执行并获取返回值
  if (typeof f?.placeholder === 'function') {
    return f.placeholder()
  }
  
  // 2. 原有逻辑：字符串类型直接使用
  if (f?.placeholder) return f.placeholder
  
  // 3. 自动生成 placeholder 的逻辑
  const label = getFieldLabel(f)
  const selectLike = ['select', 'radio', 'checkbox', 'checkbox-group', 'cascader', 'tree-select']
  if (selectLike.includes(f?.type)) return (t('form.please_select') || '请选择') + label
  return (t('form.please_enter') || '请输入') + label
}

// visible: 布尔或 () => boolean，为 false 时不渲染
function resolveVisible(f) {
  if (f?.visible === false) return false
  if (typeof f?.visible === 'function') return f.visible()
  return true
}



const OPTION_TYPES = ['select', 'radio', 'checkbox', 'checkbox-group', 'cascader']

// function ensureOptionsLoaded() {
//   const f = props.field
//   if (f?.apiUrl && OPTION_TYPES.includes(f.type)) loadFieldOptions(f)
// }

function ensureOptionsLoaded() {
  const f = props.field
  if (!f?.apiUrl || !OPTION_TYPES.includes(f.type)) return

  // 注册 options 加载完成后的回显钩子
  f.__onOptionsLoaded = f.__onOptionsLoaded || []
  f.__onOptionsLoaded.push(async () => {
    const val = props.model[f.prop]
    if (val !== undefined && val !== null) {
      await nextTick()
      // 重新赋值，触发 el-select label 映射
      props.model[f.prop] = Array.isArray(val) ? [...val] : val
    }
  })

  loadFieldOptions(f)
}


onMounted(ensureOptionsLoaded)
watch(() => [props.field?.apiUrl, props.field?.type], ensureOptionsLoaded)
</script>
