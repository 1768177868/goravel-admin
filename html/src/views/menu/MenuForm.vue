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
          <template v-if="f.prop === 'parent_id'">
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
          </template>

         
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
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
// 引入配置式表单组件
import FormField from '../../components/Form/FormField.vue'
import { normalizeFormData } from '../../utils/normalizeFormData'
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

// 对话框显隐状态
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 对话框标题
const dialogTitle = computed(() => formData.id ? t('menu_management.edit_menu') : t('menu_management.add_menu'))

// 树形选择数据，包含顶级菜单选项
const treeSelectData = computed(() => {
  return [
    { value: 0, label: t('menu_management.top_menu') },
    ...props.menuOptions
  ]
})

// 表单数据
const formData = reactive({
  id: null,
  parent_id: 0,
  type: "2", // 默认为菜单
  name: '',
  slug: '',
  path: '',
  component: '',
  icon: '',
  status: "1",
  sort: 0,
  is_hidden: 0, // 0: 显示, 1: 隐藏
  link_type: 1, // 1: 内部页面, 2: 外部链接
  open_type: 1 // 1: iframe嵌套, 2: 新窗口打开
})

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
    {
      prop: 'parent_id',
      label: t('menu_management.parent_menu'),
      type: 'custom', // 自定义树形选择
      disabled: loading.value,
      noValidate: true
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
      options: [
        { label: t('menu_management.link_type_internal'), value: 1 },
        { label: t('menu_management.link_type_external'), value: 2 }
      ]
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
      options: [
        { label: t('menu_management.open_type_iframe'), value: 1 },
        { label: t('menu_management.open_type_new_window'), value: 2 }
      ],
      // 仅外部链接时显示
      visible: () => formData.link_type === 2
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
      options: [
        { label: t('menu_management.is_hidden_show'), value: 0 },
        { label: t('menu_management.is_hidden_hide'), value: 1 }
      ],
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
      // 后端返回的是 PascalCase 字段，需要正确映射
      const normalized = normalizeFormData({
        id: menu.id,
        parent_id: menu.ParentID ?? menu.parent_id ?? 0,
        type: menu.Type ?? menu.type ?? 2,
        name: menu.Title ?? menu.name ?? '',
        slug: menu.Slug ?? menu.slug ?? '',
        path: menu.Path ?? menu.path ?? '',
        component: menu.Component ?? menu.component ?? '',
        icon: menu.Icon ?? menu.icon ?? '',
        status: menu.Status ?? menu.status ?? 1,
        sort: menu.Sort ?? menu.sort ?? 0,
        is_hidden: menu.IsHidden ?? menu.is_hidden ?? 0,
        link_type: menu.LinkType ?? menu.link_type ?? 1,
        open_type: menu.OpenType ?? menu.open_type ?? 1
      }, {
        status: 'string',
        type: 'string',
      })

      Object.assign(formData, normalized)
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
  Object.assign(formData, {
    id: null,
    parent_id: 0,
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
    open_type: 1
  })
  formRef.value?.resetFields()
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        // 转换前端字段名为后端期望的字段名
        const data = {
          type: formData.type,
          title: formData.name,
          slug: formData.slug,
          path: formData.path,
          component: formData.link_type === 1 ? formData.component : '',
          icon: formData.icon,
          status: formData.status,
          sort: formData.sort,
          is_hidden: formData.is_hidden,
          parent_id: formData.parent_id === 0 ? null : formData.parent_id,
          link_type: formData.link_type,
          open_type: formData.open_type
        }
        
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
</style>