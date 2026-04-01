<template>
  <el-form-item
    v-if="field.prop && resolveVisible(field)"
    :label="getFieldLabel(field)"
    :prop="field.prop"
    :style="field.style || {}"
    :label-width="isMobile ? '100%' : undefined"
    :class="{ 'mobile-form-item': isMobile }"
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
      :autocomplete="field.autocomplete || (field.type === 'password' ? 'new-password' : undefined)"
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

    <!-- 数字输入（原生 number 输入） -->
    <el-input
      v-else-if="field.type === 'number'"
      v-model="model[field.prop]"
      :type="'number'"
      :placeholder="getPlaceholder(field)"
      :clearable="field.clearable !== false"
      :disabled="field.disabled"
      :min="field.min"
      :max="field.max"
      :step="field.step"
      v-bind="field.props || {}"
    />

    <!-- 数字步进器（支持单位显示） -->
    <el-input-number
      v-else-if="field.type === 'input-number'"
      v-model="model[field.prop]"
      :placeholder="getPlaceholder(field)"
      :disabled="field.disabled"
      :min="field.min"
      :max="field.max"
      :step="field.step"
      v-bind="field.props || {}"
    >
      <template v-if="field.prefix" #prefix>
        <span>{{ field.prefix }}</span>
      </template>
      <template v-if="field.suffix" #suffix>
        <span>{{ field.suffix }}</span>
      </template>
    </el-input-number>
    <span v-if="field.type === 'input-number' && field.unit" class="input-number-unit">{{ field.unit }}</span>

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

    <!-- 图片上传（复用封装组件） -->
    <ImageUpload
      v-else-if="field.type === 'image-upload'"
      v-model="model[field.prop]"
      :upload-mode="field.uploadMode || 'both'"
      :width="field.cropWidth || 400"
      :height="field.cropHeight || 400"
      :aspect-ratio="field.aspectRatio ?? null"
      v-bind="field.props || {}"
    />

    <ImageUploadMultiple
      v-else-if="field.type === 'image-upload-multiple'"
      v-model="model[field.prop]"
      :limit="field.limit ?? 9"
      :min-count="field.minCount ?? 0"
      :max-size-m-b="field.maxSizeMB ?? 10"
      v-bind="field.props || {}"
    />

     <!-- 新增：图标选择器 -->
     <div v-else-if="field.type === 'icon'" class="icon-picker">
      <el-input
        v-model="model[field.prop]"
        :placeholder="getPlaceholder(field)"
        :clearable="field.clearable !== false"
        :disabled="field.disabled"
        @clear="clearIcon(model, field.prop)"
        v-bind="field.props || {}"
      >
        <template #prefix>
          <el-icon v-if="getIconComponent(model[field.prop])" class="selected-icon">
            <component :is="getIconComponent(model[field.prop])" />
          </el-icon>
        </template>
      </el-input>
      <el-popover
        placement="bottom"
        trigger="click"
        width="420"
        v-model:visible="iconPickerVisible"
        popper-class="icon-picker-popover"
      >
        <div class="icon-picker-content">
          <el-input
            v-model="iconSearch"
            :placeholder="field.iconSearchPlaceholder || '搜索图标'"
            size="small"
            clearable
          />
          <div class="icon-grid">
            <el-tooltip
              v-for="icon in iconOptions"
              :key="icon"
              :content="icon"
              placement="top"
            >
              <el-button circle @click="selectIcon(icon, model, field.prop)">
                <el-icon><component :is="iconComponents[icon]" /></el-icon>
              </el-button>
            </el-tooltip>
          </div>
        </div>
        <template #reference>
          <el-button link type="primary">{{ field.selectText || '选择图标' }}</el-button>
        </template>
      </el-popover>
    </div>

    <!-- 穿梭框（支持 apiUrl / options / optionsFn） -->
    <TransferField
      v-else-if="field.type === 'transfer'"
      :field="field"
      :model-value="model[field.prop]"
      @update:model-value="model[field.prop] = $event"
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
 * - number, switch, rate, slider, color, transfer
 * - 其它 type 走默认插槽，可自定义富文本、上传等
 */
import { computed, onMounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import TreeSelectField from '../SearchForm/TreeSelectField.vue'
import TransferField from './TransferField.vue'
import ImageUpload from '../ImageUpload.vue'
import ImageUploadMultiple from '../ImageUploadMultiple.vue'
import { useFieldOptions } from '../SearchForm/useFieldOptions'
import { useIconPicker } from '../SearchForm/useIconPicker'
import { useResponsive } from '../../composables/useResponsive'

const { isMobile } = useResponsive()

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

// 复用图标选择器逻辑
const {
  iconComponents,
  iconPickerVisible,
  iconSearch,
  normalizeIconName,
  getIconComponent,
  filteredIcons: defaultFilteredIcons,
  selectIcon,
  clearIcon
} = useIconPicker()

const iconOptions = computed(() => {
  const dynamicOptions = getFieldOptions(props.field)
  const hasDynamicSource = !!(
    props.field?.apiUrl ||
    (Array.isArray(props.field?.options) && props.field.options.length > 0) ||
    typeof props.field?.optionsFn === 'function'
  )

  if (!hasDynamicSource) {
    return defaultFilteredIcons.value
  }

  const keyword = iconSearch.value.trim().toLowerCase()
  const normalizedIcons = (Array.isArray(dynamicOptions) ? dynamicOptions : [])
    .map((item) => {
      if (typeof item === 'string') return item
      if (item && typeof item === 'object') {
        return item.value ?? item.name ?? item.icon ?? item.label ?? ''
      }
      return ''
    })
    .map((name) => normalizeIconName(String(name)))
    .filter(Boolean)

  const uniqueIcons = [...new Set(normalizedIcons)]
  const source = uniqueIcons.length > 0 ? uniqueIcons : defaultFilteredIcons.value

  if (!keyword) return source
  return source.filter((name) => name.toLowerCase().includes(keyword))
})

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



const OPTION_TYPES = ['select', 'radio', 'checkbox', 'checkbox-group', 'cascader', 'icon']

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

<style scoped>
  /* 图标选择器样式 */
  .icon-picker {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  
  .icon-picker-content {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  
  .icon-grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 8px;
    max-height: 220px;
    overflow-y: auto;
  }
  
  .icon-grid .el-button {
    width: 36px;
    height: 36px;
    padding: 0;
  }
  
  .selected-icon {
    margin-right: 6px;
  }
  
  .form-item-tip {
    font-size: 12px;
    color: #909399;
    margin-top: 4px;
    line-height: 1.4;
  }

  .input-number-unit {
    margin-left: 8px;
    color: var(--el-text-color-regular);
    white-space: nowrap;
  }

  /* 移动端优化 */
  .mobile-form-item {
    margin-bottom: 20px;
  }

  .mobile-form-item :deep(.el-form-item__label) {
    width: 100% !important;
    text-align: left;
    margin-bottom: 8px;
    padding: 0;
    font-weight: 500;
  }

  .mobile-form-item :deep(.el-form-item__content) {
    width: 100%;
    margin-left: 0 !important;
  }

  @media (max-width: 768px) {
    .icon-grid {
      grid-template-columns: repeat(6, 1fr);
      gap: 6px;
      max-height: 250px;
    }

    .icon-grid .el-button {
      width: 32px;
      height: 32px;
    }
  }
</style>
