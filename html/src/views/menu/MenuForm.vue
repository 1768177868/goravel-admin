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
      <el-form-item :label="$t('menu_management.parent_menu')">
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
      </el-form-item>
      <el-form-item :label="$t('table.type')" prop="type">
        <el-radio-group v-model="formData.type" :disabled="loading">
          <el-radio :label="1">{{ $t('menu.type_directory') }}</el-radio>
          <el-radio :label="2">{{ $t('menu.type_menu') }}</el-radio>
          <el-radio :label="3">{{ $t('menu.type_button') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item :label="$t('menu_management.name')" prop="name">
        <el-input v-model="formData.name" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('menu_management.slug')" prop="slug">
        <el-input v-model="formData.slug" :placeholder="$t('menu_management.slug_placeholder')" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('menu_management.link_type')" prop="link_type">
        <el-radio-group v-model="formData.link_type" :disabled="loading">
          <el-radio :label="1">{{ $t('menu_management.link_type_internal') }}</el-radio>
          <el-radio :label="2">{{ $t('menu_management.link_type_external') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item 
        :label="$t('menu_management.path')" 
        prop="path"
        v-if="formData.link_type === 1"
      >
        <el-input 
          v-model="formData.path" 
          :placeholder="$t('menu_management.path_placeholder_internal')"
          :disabled="loading"
        />
      </el-form-item>
      <el-form-item 
        :label="$t('menu_management.path')" 
        prop="path"
        v-else
      >
        <el-input 
          v-model="formData.path" 
          :placeholder="$t('menu_management.path_placeholder_external')"
          :disabled="loading"
        />
      </el-form-item>
      <el-form-item 
        :label="$t('menu_management.component')" 
        v-if="formData.link_type === 1"
      >
        <el-input 
          v-model="formData.component" 
          :placeholder="$t('menu_management.component_placeholder')"
          :disabled="loading"
        />
        <div class="form-item-tip">{{ $t('menu_management.component_tip') }}</div>
      </el-form-item>
      <el-form-item 
        :label="$t('menu_management.open_type')" 
        prop="open_type"
        v-if="formData.link_type === 2"
      >
        <el-radio-group v-model="formData.open_type" :disabled="loading">
          <el-radio :label="1">{{ $t('menu_management.open_type_iframe') }}</el-radio>
          <el-radio :label="2">{{ $t('menu_management.open_type_new_window') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item :label="$t('menu_management.icon')">
        <div class="icon-picker">
          <el-input
            v-model="formData.icon"
            :placeholder="$t('menu_management.icon_placeholder')"
            clearable
            :disabled="loading"
            @clear="clearIcon"
          >
            <template #prefix>
              <el-icon v-if="getIconComponent(formData.icon)" class="selected-icon">
                <component :is="getIconComponent(formData.icon)" />
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
                :placeholder="$t('menu_management.icon_search')"
                size="small"
                clearable
              />
              <div class="icon-grid">
                <el-tooltip
                  v-for="icon in filteredIcons"
                  :key="icon"
                  :content="icon"
                  placement="top"
                >
                  <el-button circle @click="selectIcon(icon)">
                    <el-icon><component :is="iconComponents[icon]" /></el-icon>
                  </el-button>
                </el-tooltip>
              </div>
            </div>
            <template #reference>
              <el-button link type="primary">{{ $t('menu_management.select_icon') }}</el-button>
            </template>
          </el-popover>
        </div>
      </el-form-item>
      <el-form-item :label="$t('table.status')" prop="status">
        <el-radio-group v-model="formData.status" :disabled="loading">
          <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
          <el-radio :label="0">{{ $t('common.disabled') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item :label="$t('menu_management.is_hidden')">
        <el-radio-group v-model="formData.is_hidden" :disabled="loading">
          <el-radio :label="0">{{ $t('menu_management.is_hidden_show') }}</el-radio>
          <el-radio :label="1">{{ $t('menu_management.is_hidden_hide') }}</el-radio>
        </el-radio-group>
        <!-- <div class="form-item-tip">{{ $t('menu_management.is_hidden_tip') }}</div> -->
      </el-form-item>
      <el-form-item :label="$t('common.sort')">
        <el-input-number v-model="formData.sort" :min="0" :disabled="loading" />
      </el-form-item>
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

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const dialogTitle = computed(() => formData.id ? t('menu_management.edit_menu') : t('menu_management.add_menu'))

// 树形选择数据，包含顶级菜单选项
const treeSelectData = computed(() => {
  return [
    { value: 0, label: t('menu_management.top_menu') },
    ...props.menuOptions
  ]
})

const formData = reactive({
  id: null,
  parent_id: 0,
  type: 2, // 默认为菜单
  name: '',
  slug: '',
  path: '',
  component: '',
  icon: '',
  status: 1,
  sort: 0,
  is_hidden: 0, // 0: 显示, 1: 隐藏
  link_type: 1, // 1: 内部页面, 2: 外部链接
  open_type: 1 // 1: iframe嵌套, 2: 新窗口打开
})

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

const formRules = computed(() => ({
  name: [{ required: true, message: t('menu_management.name_required'), trigger: 'blur' }],
  slug: [{ required: true, message: t('menu_management.slug_required'), trigger: 'blur' }],
  path: [
    { required: true, message: t('menu_management.path_required'), trigger: 'blur' },
    { validator: validateUrl, trigger: ['blur', 'change'] }
  ]
}))

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
      // 新增模式，重置表单
      resetForm()
    }
  } else {
    // 对话框关闭时，重置表单
    resetForm()
  }
})

// 监听 editId 变化，加载详情
watch(() => props.editId, async (newId) => {
  if (dialogVisible.value) {
    if (newId) {
      await loadDetail(newId)
    } else {
      // 新增模式，重置表单
      resetForm()
    }
  }
})

const loadDetail = async (id) => {
  loading.value = true
  try {
    const res = await getMenuDetail(id)
    if (res.data && res.data.menu) {
      const menu = res.data.menu
      // 后端返回的是 PascalCase 字段，需要正确映射
      Object.assign(formData, {
        id: menu.id,
        parent_id: menu.ParentID !== undefined ? menu.ParentID : (menu.parent_id || 0),
        type: menu.Type !== undefined ? menu.Type : (menu.type !== undefined ? menu.type : 2),
        name: menu.Title || menu.name || '',
        slug: menu.Slug || menu.slug || '',
        path: menu.Path || menu.path || '',
        component: menu.Component || menu.component || '',
        icon: menu.Icon || menu.icon || '',
        status: menu.Status !== undefined ? menu.Status : (menu.status !== undefined ? menu.status : 1),
        sort: menu.Sort !== undefined ? menu.Sort : (menu.sort !== undefined ? menu.sort : 0),
        is_hidden: menu.IsHidden !== undefined ? menu.IsHidden : (menu.is_hidden !== undefined ? menu.is_hidden : 0),
        link_type: menu.LinkType !== undefined ? menu.LinkType : (menu.link_type !== undefined ? menu.link_type : 1),
        open_type: menu.OpenType !== undefined ? menu.OpenType : (menu.open_type !== undefined ? menu.open_type : 1)
      })
    }
  } catch (error) {
    console.error('Load menu detail error:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  loading.value = false
  Object.assign(formData, {
    id: null,
    parent_id: 0,
    type: 2,
    name: '',
    slug: '',
    path: '',
    component: '',
    icon: '',
    status: 1,
    sort: 0,
    is_hidden: 0,
    link_type: 1,
    open_type: 1
  })
  formRef.value?.resetFields()
}

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

const handleCancel = () => {
  dialogVisible.value = false
  resetForm()
}

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

