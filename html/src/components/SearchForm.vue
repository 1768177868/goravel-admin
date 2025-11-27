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
          
          <!-- 树形选择器（使用 el-popover + el-tree） -->
          <el-popover
            v-else-if="field.type === 'tree-select'"
            placement="bottom-start"
            :width="field.popoverWidth || 300"
            :visible="getTreeSelectPopoverVisible(field)"
            @update:visible="(val) => { if (val !== undefined) setTreeSelectPopoverVisible(field, val) }"
            :popper-options="{ 
              modifiers: [
                { name: 'computeStyles', options: { gpuAcceleration: false } },
                { name: 'preventOverflow', options: { boundary: 'viewport' } }
              ]
            }"
          >
            <template #reference>
              <el-input
                :model-value="getTreeSelectInputValue(field)"
                :placeholder="getFieldPlaceholder(field)"
                :clearable="field.clearable !== false"
                :disabled="field.disabled"
                :style="{ width: field.width || '200px' }"
                @clear="(e) => { e?.stopPropagation(); handleTreeSelectClear(field) }"
                @input="(val) => handleTreeSelectInput(field, val)"
                @focus="!field.disabled && !getTreeSelectPopoverVisible(field) && toggleTreeSelectPopover(field)"
                @click.stop.prevent="!field.disabled && toggleTreeSelectPopover(field)"
                @mousedown.stop.prevent
              >
                <template #suffix>
                  <el-icon 
                    v-if="!field.disabled"
                    class="el-input__icon"
                    :class="{ 'is-reverse': getTreeSelectPopoverVisible(field) }"
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
                v-model="treeSelectFilterText[field.prop]"
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
                ref="(el) => setTreeRef(field, el)"
                :data="getTreeSelectData(field)"
                :props="field.treeProps || { label: 'name', children: 'children' }"
                node-key="id"
                :default-expand-all="field.filterable !== false && treeSelectFilterText[field.prop] ? true : (field.defaultExpandAll || false)"
                :expand-on-click-node="false"
                :highlight-current="true"
                @node-click="(data) => handleTreeSelectNodeClick(field, data)"
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
import { ref, useSlots, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { Search, Refresh, ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { getOptions } from '../api/option'

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
const formHeight = ref(0)
const singleLineHeight = ref(0)
const shouldShowExpandButton = ref(false)
const treeSelectPopovers = ref({}) // 存储树形选择器的弹窗状态
const treeSelectFilterText = ref({}) // 存储树形选择器的搜索文本
const fieldOptionsCache = ref({}) // 缓存通过 API 获取的选项数据
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
    // 计算可以在一行内显示的字段数量
    if (singleLineHeight.value > 0) {
      return {
        maxHeight: `${singleLineHeight.value}px`
      }
    }
  }
  return {}
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

const getFieldOptions = (field) => {
  if (!field) return []
  
  if (field.options && Array.isArray(field.options)) {
    return field.options
  }
  
  if (field.apiUrl) {
    const cacheKey = field.apiUrl
    if (fieldOptionsCache.value[cacheKey]) {
      return fieldOptionsCache.value[cacheKey]
    }
    loadFieldOptions(field)
    return []
  }
  
  if (field.optionsFn && typeof field.optionsFn === 'function') {
    try {
      return field.optionsFn()
    } catch (e) {
      return []
    }
  }
  
  return []
}

const loadFieldOptions = async (field) => {
  if (!field.apiUrl) return
  
  const cacheKey = field.apiUrl
  if (fieldOptionsCache.value[cacheKey] !== undefined) {
    return
  }
  
  try {
    fieldOptionsCache.value[cacheKey] = null
    
    if (field.apiUrl.startsWith('/options')) {
      const url = new URL(field.apiUrl, window.location.origin)
      const type = url.searchParams.get('type')
      const res = await getOptions(type)
      if (res.data && res.data.options) {
        fieldOptionsCache.value[cacheKey] = res.data.options
      }
    } else {
      const res = await fetch(field.apiUrl, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json'
        }
      })
      if (res.ok) {
        const data = await res.json()
        if (data.data && data.data.options) {
          fieldOptionsCache.value[cacheKey] = data.data.options
        } else if (data.options) {
          fieldOptionsCache.value[cacheKey] = data.options
        } else if (Array.isArray(data.data)) {
          fieldOptionsCache.value[cacheKey] = data.data.map(item => ({
            label: item.name || item.Name || item.label || String(item.id || item.ID),
            value: String(item.id || item.ID || item.value)
          }))
        } else if (Array.isArray(data)) {
          fieldOptionsCache.value[cacheKey] = data.map(item => ({
            label: item.name || item.Name || item.label || String(item.id || item.ID),
            value: String(item.id || item.ID || item.value)
          }))
        }
      }
    }
  } catch (error) {
    console.error('Load field options error:', error)
    fieldOptionsCache.value[cacheKey] = []
  }
}

const getFieldStyle = (field) => {
  return field.style || {}
}

const getTreeSelectDisplayValue = (field) => {
  const value = props.model[field.prop]
  if (!value) return ''
  
  const findNode = (data, targetId) => {
    for (const node of data) {
      const nodeId = node[field.treeProps?.value || 'id']
      if (nodeId == value) {
        return node[field.treeProps?.label || 'name']
      }
      if (node[field.treeProps?.children || 'children']) {
        const found = findNode(node[field.treeProps?.children || 'children'], targetId)
        if (found) return found
      }
    }
    return ''
  }
  
  return findNode(field.treeData || [], value) || ''
}

const getTreeSelectInputValue = (field) => {
  const selectedValue = props.model[field.prop]
  if (selectedValue) {
    const displayValue = getTreeSelectDisplayValue(field)
    const filterText = treeSelectFilterText.value[field.prop]
    if (filterText && filterText !== displayValue) {
      return filterText
    }
    return displayValue
  }
  const filterText = treeSelectFilterText.value[field.prop]
  return filterText || ''
}

const getTreeSelectPopoverVisible = (field) => {
  if (!treeSelectPopovers.value[field.prop]) {
    treeSelectPopovers.value[field.prop] = false
  }
  return treeSelectPopovers.value[field.prop]
}

const setTreeSelectPopoverVisible = (field, visible) => {
  treeSelectPopovers.value[field.prop] = visible
}

const toggleTreeSelectPopover = (field) => {
  if (!treeSelectPopovers.value.hasOwnProperty(field.prop)) {
    treeSelectPopovers.value[field.prop] = false
  }
  const newState = !treeSelectPopovers.value[field.prop]
  treeSelectPopovers.value[field.prop] = newState
  if (!newState && !props.model[field.prop]) {
    treeSelectFilterText.value[field.prop] = ''
  }
}

const handleTreeSelectNodeClick = (field, data) => {
  const valueKey = field.treeProps?.value || 'id'
  props.model[field.prop] = data[valueKey]
  treeSelectFilterText.value[field.prop] = ''
  treeSelectPopovers.value[field.prop] = false
}

const handleTreeSelectClear = (field) => {
  props.model[field.prop] = ''
  treeSelectFilterText.value[field.prop] = ''
}

const handleTreeSelectInput = (field, val) => {
  if (!treeSelectFilterText.value.hasOwnProperty(field.prop)) {
    treeSelectFilterText.value[field.prop] = ''
  }
  treeSelectFilterText.value[field.prop] = val
  if (val && !getTreeSelectPopoverVisible(field)) {
    toggleTreeSelectPopover(field)
  }
}

const getTreeSelectData = (field) => {
  if (field.treeData && Array.isArray(field.treeData)) {
    return getFilteredTreeData(field, field.treeData)
  }
  
  if (field.apiUrl) {
    const cacheKey = field.apiUrl
    if (fieldOptionsCache.value[cacheKey]) {
      const data = fieldOptionsCache.value[cacheKey]
      if (Array.isArray(data) && data.length > 0 && data[0].children !== undefined) {
        return getFilteredTreeData(field, data)
      } else if (Array.isArray(data)) {
        return getFilteredTreeData(field, data)
      }
    }
    loadTreeSelectData(field)
    return []
  }
  
  return getFilteredTreeData(field, [])
}

const loadTreeSelectData = async (field) => {
  if (!field.apiUrl) return
  
  const cacheKey = field.apiUrl
  if (fieldOptionsCache.value[cacheKey] !== undefined) {
    return
  }
  
  try {
    fieldOptionsCache.value[cacheKey] = null
    
    if (field.apiUrl.startsWith('/options')) {
      const url = new URL(field.apiUrl, window.location.origin)
      const type = url.searchParams.get('type')
      const res = await getOptions(type)
      if (res.data) {
        if (res.data.options && Array.isArray(res.data.options)) {
          fieldOptionsCache.value[cacheKey] = res.data.options
        } else if (res.data.list && Array.isArray(res.data.list)) {
          fieldOptionsCache.value[cacheKey] = res.data.list
        }
      }
    } else {
      const res = await fetch(field.apiUrl, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json'
        }
      })
      if (res.ok) {
        const data = await res.json()
        if (data.data && data.data.options) {
          fieldOptionsCache.value[cacheKey] = data.data.options
        } else if (data.data && data.data.list) {
          fieldOptionsCache.value[cacheKey] = data.data.list
        } else if (data.options) {
          fieldOptionsCache.value[cacheKey] = data.options
        } else if (Array.isArray(data.data)) {
          fieldOptionsCache.value[cacheKey] = data.data
        } else if (Array.isArray(data)) {
          fieldOptionsCache.value[cacheKey] = data
        }
      }
    }
  } catch (error) {
    console.error('Load tree select data error:', error)
    fieldOptionsCache.value[cacheKey] = []
  }
}

const getFilteredTreeData = (field, treeData) => {
  const filterText = treeSelectFilterText.value[field.prop]
  if (!filterText || filterText === '') {
    return treeData || []
  }
  
  const labelKey = field.treeProps?.label || 'name'
  const childrenKey = field.treeProps?.children || 'children'
  
  const filterNode = (node) => {
    const label = node[labelKey] || ''
    const matches = label.toLowerCase().includes(filterText.toLowerCase())
    
    if (node[childrenKey] && Array.isArray(node[childrenKey])) {
      const filteredChildren = node[childrenKey].map(child => filterNode(child)).filter(Boolean)
      if (matches || filteredChildren.length > 0) {
        return {
          ...node,
          [childrenKey]: filteredChildren
        }
      }
    } else if (matches) {
      return node
    }
    
    return null
  }
  
  return (treeData || []).map(node => filterNode(node)).filter(Boolean)
}

// 存储树形组件的引用
const treeRefs = ref({})
const setTreeRef = (field, el) => {
  if (el) {
    treeRefs.value[field.prop] = el
  }
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

// 检测表单高度，判断是否需要显示展开按钮
const checkFormHeight = () => {
  nextTick(() => {
    if (!formRef.value || !formRef.value.$el) return
    
    const formEl = formRef.value.$el
    const formItems = Array.from(formEl.querySelectorAll('.el-form-item:not(.action-item)'))
    
    if (formItems.length === 0) return
    
    // 获取第一个表单项的高度（作为单行高度参考）
    const firstItem = formItems[0]
    if (firstItem) {
      const firstItemRect = firstItem.getBoundingClientRect()
      const computedStyle = window.getComputedStyle(firstItem)
      const marginBottom = parseFloat(computedStyle.marginBottom) || 18
      // 单行高度 = 表单项高度 + 底部间距
      singleLineHeight.value = firstItemRect.height + marginBottom
    }
    
    // 获取表单字段容器的高度
    const fieldsWrapper = formEl.querySelector('.form-fields-wrapper')
    if (fieldsWrapper) {
      const wrapperRect = fieldsWrapper.getBoundingClientRect()
      formHeight.value = wrapperRect.height
      
      // 计算可以显示多少行（根据容器宽度和表单项宽度）
      // 获取容器宽度（减去 padding）
      const containerPadding = 40 // 左右 padding 各 20px
      const containerWidth = wrapperRect.width - containerPadding
      let currentRowWidth = 0
      let rowCount = 1
      const rowGap = 10 // 表单项之间的间距
      let firstRowItems = [] // 第一行能显示的字段
      
      formItems.forEach((item) => {
        const itemRect = item.getBoundingClientRect()
        const itemWidth = itemRect.width
        const itemMarginRight = parseFloat(window.getComputedStyle(item).marginRight) || 10
        
        // 如果当前行放不下这个表单项，换行
        if (currentRowWidth + itemWidth + itemMarginRight > containerWidth && currentRowWidth > 0) {
          rowCount++
          currentRowWidth = itemWidth + itemMarginRight
        } else {
          if (rowCount === 1) {
            firstRowItems.push(item)
          }
          currentRowWidth += itemWidth + itemMarginRight + rowGap
        }
      })
      
      // 如果超过一行，需要显示展开按钮
      if (rowCount > 1) {
        shouldShowExpandButton.value = true
        // 计算收起状态应该显示的高度（尽可能多地显示第一行的字段）
        // 如果第一行能显示所有基础字段，则高度就是单行高度
        // 否则需要计算第一行实际占用的高度
        if (firstRowItems.length > 0) {
          // 使用第一行最后一个元素的位置来计算高度
          const lastFirstRowItem = firstRowItems[firstRowItems.length - 1]
          const lastItemRect = lastFirstRowItem.getBoundingClientRect()
          const firstItemRect = firstRowItems[0].getBoundingClientRect()
          // 第一行的实际高度 = 最后一个元素底部 - 第一个元素顶部 + 底部间距
          const firstRowHeight = lastItemRect.bottom - firstItemRect.top + 18
          singleLineHeight.value = Math.max(singleLineHeight.value, firstRowHeight)
        }
        
        // 如果默认是收起状态，且表单高度超过单行，则默认收起
        if (!props.defaultExpanded && expanded.value === props.defaultExpanded) {
          expanded.value = false
        }
      } else {
        shouldShowExpandButton.value = false
        // 如果不需要展开按钮，则始终展开
        if (!hasAdvancedFields.value) {
          expanded.value = true
        }
      }
    }
  })
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
          loadTreeSelectData(field)
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
      if (field.apiUrl) {
        if (field.type === 'tree-select') {
          loadTreeSelectData(field)
        } else if (field.type === 'select') {
          loadFieldOptions(field)
        }
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
  background: #f5f7fa;
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

  // 树形选择器箭头图标旋转样式
  :deep(.el-input__icon.is-reverse) {
    transform: rotate(180deg);
  }
}
</style>
