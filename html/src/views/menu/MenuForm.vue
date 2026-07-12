<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="600px"
    @close="handleDialogClose"
  >
    <div v-loading="loading">
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <!-- 配置式渲染表单字段 -->
        <FormField
          v-for="f in formFields"
          :key="f.prop"
          :field="f"
          :model="formData"
        >
          <!-- 父菜单树形选择插槽 -->
          <!-- <template v-if="f.prop === 'parent_id'">
            <el-tree-select
              v-model="formData.parent_id"
              :data="treeSelectData"
              :placeholder="$t('form.select_parent') + $t('menu_management.parent_menu')"
              :props="{ label: 'label', value: 'value', children: 'children' }"
              clearable
              check-strictly
              :render-after-expand="false"
              :disabled="loading"
              style="width: 100%"
            />
          </template> -->

          <!-- 组件路径字段的提示文字插槽 -->
          <template v-if="f.prop === 'component'">
            <el-input 
              v-model="formData.component" 
              :placeholder="$t('menu_management.component_placeholder')"
              :disabled="loading"
            />
            <div class="form-item-tip">{{ $t('menu_management.component_tip') }}</div>
          </template>
        </FormField>
      </el-form>
    </div>
    <template #footer>
      <el-button @click="handleCancel">{{ $t('common.cancel') }}</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ $t('common.confirm') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch ,nextTick} from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
// 引入配置式表单组件
import FormField from '../../components/Form/FormField.vue'
import { normalizeFormData, mapFields } from '../../utils/normalizeFormData'
import { getShowHideOptions, getOpenTypeOptions, getMenuLinkTypeOptions } from '@/utils/options'
import { excludeNodeAndChildren, mapTree } from '../../utils/tree'

import {
  getMenuDetail,
  createMenu,
  updateMenu
} from '../../api/menu'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  editId: {
    type: [Number, String],
    default: null
  },
  menuOptions: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const formRef = ref(null)
const submitting = ref(false)
const loading = ref(false)

// 定义表单初始值的复用函数（返回新对象，避免引用问题）
// 新增时 parent_id 为 null，不高亮任何节点；用户选择「顶级菜单」后再设为 0
const getFormInitialValue = () => ({
  id: null,
  parent_id: null,
  type: "2",
  name: '',
  slug: '',
  path: '',
  component: '',
  icon: '',
  status: "1",
  sort: 0,
  is_hidden: 0,
  link_type: 1,
  open_type: 1,
  no_cache: 0
})

// 对话框显隐状态
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 对话框标题
const dialogTitle = computed(() => formData.id ? t('menu_management.edit_menu') : t('menu_management.add_menu'))

// 树形选择数据，排除当前菜单及其子菜单
const treeSelectData = computed(() => {
  if (!props.menuOptions || props.menuOptions.length === 0) {
    return [{ value: 0, label: t('menu_management.top_menu') }]
  }
  
  // 如果有编辑ID，需要排除当前菜单及其所有子菜单
  if (formData.id) {
    // 将菜单选项转换为标准格式（value/label -> id/name），用于排除函数
    const standardTree = mapTree(props.menuOptions, node => ({
      id: node.value,
      name: node.label,
      children: node.children
    }), 'children')
    
    // 使用工具函数排除当前菜单及其子菜单
    const filtered = excludeNodeAndChildren(standardTree, formData.id, 'id', 'children')
    
    // 转换回 el-tree-select 需要的格式（value/label）
    const result = mapTree(filtered, node => ({
      value: node.id,
      label: node.name
    }), 'children')
    
    return [{ value: 0, label: t('menu_management.top_menu') }, ...result]
  }
  
  // 新增时，直接使用菜单选项
  return [
    { value: 0, label: t('menu_management.top_menu') },
    ...props.menuOptions
  ]
})

// 表单数据
const formData = reactive(getFormInitialValue())

// 图标选择相关逻辑
const iconComponents = ElementPlusIconsVue
const iconPickerVisible = ref(false)
const iconSearch = ref('')
const iconList = Object.keys(ElementPlusIconsVue).filter(name => /^[A-Z]/.test(name)).sort()

const normalizeIconName = (iconName) => {
  if (!iconName) {
    return ''
  }
  const trimmed = iconName.trim()
  if (!trimmed) {
    return ''
  }
  if (iconComponents[trimmed]) {
    return trimmed
  }
  const pascalCase = trimmed.charAt(0).toUpperCase() + trimmed.slice(1)
  if (iconComponents[pascalCase]) {
    return pascalCase
  }
  return ''
}

const getIconComponent = (iconName) => {
  const normalized = normalizeIconName(iconName)
  return normalized ? iconComponents[normalized] : null
}

const filteredIcons = computed(() => {
  const keyword = iconSearch.value.trim().toLowerCase()
  if (!keyword) {
    return iconList
  }
  return iconList.filter(name => name.toLowerCase().includes(keyword))
})

const selectIcon = (icon) => {
  formData.icon = icon
  iconPickerVisible.value = false
}

const clearIcon = () => {
  formData.icon = ''
}

// URL验证函数
const validateUrl = (rule, value, callback) => {
  if (!value) {
    callback()
    return
  }
  
  // 如果是外部链接，验证URL格式
  if (formData.link_type === 2) {
    try {
      new URL(value)
      callback()
    } catch {
      callback(new Error(t('menu_management.path_url_invalid')))
    }
  } else {
    callback()
  }
}

// 表单验证规则
const formRules = computed(() => ({
  name: [{ required: true, message: t('menu_management.name_required'), trigger: 'blur' }],
  slug: [{ required: true, message: t('menu_management.slug_required'), trigger: 'blur' }],
  path: [
    { required: true, message: t('menu_management.path_required'), trigger: 'blur' },
    { validator: validateUrl, trigger: ['blur', 'change'] }
  ]
}))

// 配置式表单字段（
const formFields = computed(() => {
  const fields = [
    // {
    //   prop: 'parent_id',
    //   label: t('menu_management.parent_menu'),
    //   type: 'custom', // 自定义树形选择
    //   disabled: loading.value,
    //   noValidate: true
    // },
    {
      prop: 'parent_id',
      label: t('menu_management.parent_menu'),
      type: 'tree-select',
      apiUrl: '/options?type=menu',
      treeProps: { label: 'name', value: 'id', children: 'children' },
      clearable: true,
      disabled: loading.value,
      topNodeLabel: t('menu_management.top_menu'),
    },
    {
      prop: 'type',
      label: t('table.type'),
      type: 'radio',
      disabled: loading.value,
      apiUrl: '/options/?type=dictionary&dictionary_type=menu_type',
      // options: [
      //   { label: t('menu.type_directory'), value: 1 },
      //   { label: t('menu.type_menu'), value: 2 },
      //   { label: t('menu.type_button'), value: 3 }
      // ]
    },
    {
      prop: 'name',
      label: t('menu_management.name'),
      type: 'input',
      disabled: loading.value
    },
    {
      prop: 'slug',
      label: t('menu_management.slug'),
      type: 'input',
      disabled: loading.value,
      placeholder: t('menu_management.slug_placeholder')
    },
    {
      prop: 'link_type',
      label: t('menu_management.link_type'),
      type: 'radio',
      disabled: loading.value,
      options: getMenuLinkTypeOptions(t),
    },
    {
      prop: 'path',
      label: t('menu_management.path'),
      type: 'input',
      disabled: loading.value,
      // 修复：使用函数替代 computed
      placeholder: () => formData.link_type === 1 
        ? t('menu_management.path_placeholder_internal') 
        : t('menu_management.path_placeholder_external')
    },
    {
      prop: 'component',
      label: t('menu_management.component'),
      type: 'custom', // 自定义带提示文字的输入框
      disabled: loading.value,
      noValidate: true,
      // 仅内部链接时显示
      visible: () => formData.link_type === 1
    },
    {
      prop: 'open_type',
      label: t('menu_management.open_type'),
      type: 'radio',
      disabled: loading.value,
      options: getOpenTypeOptions(t),
      // 仅外部链接时显示
      visible: () => formData.link_type === 2
    },
    {
      prop: 'no_cache',
      label: t('menu_management.no_cache'),
      type: 'radio',
      disabled: loading.value,
      options: [
        { label: t('menu_management.no_cache_yes'), value: 0 },
        { label: t('menu_management.no_cache_no'), value: 1 }
      ],
      noValidate: true,
      // 仅内部页面时显示
      visible: () => formData.link_type === 1
    },
    {
      prop: 'icon',
      label: t('menu_management.icon'),
      type: 'icon', // 原生支持的 icon 类型
      disabled: loading.value,
      noValidate: true,
      // 可选自定义配置
      placeholder: t('menu_management.icon_placeholder'),
      iconSearchPlaceholder: t('menu_management.icon_search'),
      selectText: t('menu_management.select_icon')
    },
    {
      prop: 'status',
      label: t('table.status'),
      type: 'radio',
      apiUrl: '/options?type=dictionary&dictionary_type=status',
      disabled: loading.value,
      // options: [
      //   { label: t('common.enabled'), value: 1 },
      //   { label: t('common.disabled'), value: 0 }
      // ]
    },
    {
      prop: 'is_hidden',
      label: t('menu_management.is_hidden'),
      type: 'radio',
      disabled: loading.value,
      options:getShowHideOptions(t),
      noValidate: true
    },
    {
      prop: 'sort',
      label: t('common.sort'),
      type: 'number', // 兼容 input-number
      disabled: loading.value,
      min: 0,
      noValidate: true
    }
  ]
  return fields
})

// 监听 link_type 变化，重新验证 path
watch(() => formData.link_type, () => {
  if (formRef.value) {
    formRef.value.validateField('path')
  }
})

// 监听 dialogVisible 变化
watch(dialogVisible, (visible) => {
  if (visible) {
    if (props.editId) {
      loadDetail(props.editId)
    } else {
      resetForm()
    }
  } else {
    resetForm()
  }
})

// 监听 editId 变化，加载详情
watch(() => props.editId, async (newId) => {
  if (dialogVisible.value) {
    if (newId) {
      await loadDetail(newId)
    } else {
      resetForm()
    }
  }
})

// 加载菜单详情
const loadDetail = async (id) => {
  loading.value = true
  try {
    const res = await getMenuDetail(id)
    if (res.data && res.data.menu) {
      const menu = res.data.menu
      // 使用工具函数映射字段，自动处理 snake_case 和 PascalCase
      const mapped = mapFields(menu, getFormInitialValue())
      // 处理 name 字段的特殊映射（Title -> name）
      mapped.name = menu.Title ?? menu.name ?? ''
      // 手动处理 parent_id（后端返回 ParentID）
      mapped.parent_id = menu.ParentID ?? menu.parent_id ?? 0

      // 规范化表单数据（类型转换）
      const normalized = normalizeFormData(mapped, {
        parent_id: 'number',
        status: 'string',
        type: 'string',
        link_type: 'number',
        open_type: 'number',
        no_cache: 'number',
        is_hidden: 'number',
        sort: 'number'
      })

      // 部署/历史菜单可能 link_type、open_type 为 0，与单选项 1/2 不匹配
      if (!normalized.link_type) {
        normalized.link_type = 1
      }
      if (!normalized.open_type) {
        normalized.open_type = 1
      }
      
      Object.assign(formData, normalized)
      
      // formData.parent_id = normalized.parent_id
    }
  } catch (error) {
    console.error('Load menu detail error:', error)
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  loading.value = false
  Object.assign(formData, getFormInitialValue())
  formRef.value?.resetFields()
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        // 转换前端字段名为后端期望的字段名（只处理需要转换的字段）
        const data = {
          ...formData,
          title: formData.name, // name -> title
          parent_id: (formData.parent_id === 0 || formData.parent_id == null) ? null : formData.parent_id, // 0/null -> null
          component: formData.link_type === 1 ? formData.component : '' // 外部链接时清空
        }
        // 删除前端使用的 name 字段，避免后端混淆
        delete data.name
        
        if (formData.id) {
          await updateMenu(formData.id, data)
          ElMessage.success(t('menu_management.update_success'))
        } else {
          await createMenu(data)
          ElMessage.success(t('menu_management.create_success'))
        }
        dialogVisible.value = false
        emit('success')
      } catch (error) {
        console.error('Submit error:', error)
      } finally {
        submitting.value = false
      }
    }
  })
}

// 取消按钮
const handleCancel = () => {
  dialogVisible.value = false
  resetForm()
}

// 对话框关闭时重置表单
const handleDialogClose = () => {
  resetForm()
}
</script>

<style scoped>
.icon-picker {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.icon-picker-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: var(--space-xs);
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
  color: var(--text-color-secondary);
  margin-top: 4px;
  line-height: 1.4;
}
</style>