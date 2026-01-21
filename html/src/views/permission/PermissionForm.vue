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
          <!-- 菜单选择的自定义插槽 -->
          <template v-if="f.prop === 'menu_id'">
            <el-popover
              v-model:visible="menuSelectVisible"
              placement="bottom-start"
              :width="300"
              trigger="click"
              :popper-options="{
                modifiers: [
                  {
                    name: 'computeStyles',
                    options: {
                      gpuAcceleration: false,
                    },
                  },
                  {
                    name: 'flip',
                    options: {
                      fallbackPlacements: ['top-start', 'bottom-start'],
                    },
                  },
                ],
              }"
              popper-class="menu-select-popover"
            >
              <template #reference>
                <el-input
                  :model-value="getSelectedMenuLabel()"
                  :placeholder="$t('form.please_select') + $t('menu.title')"
                  readonly
                  clearable
                  :disabled="loading"
                  @clear="formData.menu_id = null"
                  style="cursor: pointer"
                >
                  <template #suffix>
                    <el-icon class="el-input__icon"><ArrowDown /></el-icon>
                  </template>
                </el-input>
              </template>
              <el-tree
                :data="menuTreeData"
                :props="{ label: 'label', children: 'children' }"
                :default-expand-all="false"
                node-key="value"
                highlight-current
                :current-node-key="formData.menu_id"
                @node-click="handleMenuSelect"
                class="menu-select-tree"
              >
                <template #default="{ node, data }">
                  <span class="tree-node-label">{{ data.label }}</span>
                </template>
              </el-tree>
            </el-popover>
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
import { ArrowDown } from '@element-plus/icons-vue'
// 引入配置式表单组件
import FormField from '../../components/Form/FormField.vue'

import {
  getPermissionDetail,
  createPermission,
  updatePermission
} from '../../api/permission'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  editId: {
    type: [Number, String],
    default: null
  },
  menuTreeData: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const formRef = ref(null)
const submitting = ref(false)
const loading = ref(false)
const menuSelectVisible = ref(false)

// 对话框显隐状态
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 对话框标题
const dialogTitle = computed(() => formData.id ? t('permission.edit_permission') : t('permission.add_permission'))

// 表单数据
const formData = reactive({
  id: null,
  name: '',
  slug: '',
  method: 'GET',
  path: '',
  description: '',
  menu_id: null,
  status: 1,
  sort: 0
})

// 表单验证规则
const formRules = computed(() => ({
  name: [{ required: true, message: t('permission.name_required'), trigger: 'blur' }],
  slug: [{ required: true, message: t('permission.slug_required'), trigger: 'blur' }],
  method: [{ required: true, message: t('permission.method_required'), trigger: 'change' }],
  path: [{ required: true, message: t('permission.path_required'), trigger: 'blur' }]
}))

// 配置式表单字段
const formFields = computed(() => {
  const fields = [
    {
      prop: 'name',
      label: t('permission.name'),
      type: 'input',
      disabled: loading.value
    },
    {
      prop: 'slug',
      label: t('permission.slug'),
      type: 'input',
      disabled: loading.value
    },
    {
      prop: 'method',
      label: t('permission.method'),
      type: 'select',
      placeholder: t('form.select_method'),
      disabled: loading.value,
      // 配置select的固定选项
      options: [
        { label: 'GET', value: 'GET' },
        { label: 'POST', value: 'POST' },
        { label: 'PUT', value: 'PUT' },
        { label: 'DELETE', value: 'DELETE' }
      ]
    },
    {
      prop: 'path',
      label: t('permission.path'),
      type: 'input',
      disabled: loading.value
    },
    {
      prop: 'description',
      label: t('common.description'),
      type: 'textarea',
      disabled: loading.value,
      // 非必选字段，无需验证
      noValidate: true
    },
    {
      prop: 'menu_id',
      label: t('menu.title'),
      type: 'custom', // 自定义类型（通过插槽渲染）
      disabled: loading.value,
      noValidate: false // 需要验证（保留prop在formRules中）
    },
    {
      prop: 'status',
      label: t('table.status'),
      type: 'radio',
      disabled: loading.value,
      options: [
        { label: t('common.enabled'), value: 1 },
        { label: t('common.disabled'), value: 0 }
      ]
    },
    {
      prop: 'sort',
      label: t('common.sort'),
      type: 'input-number',
      disabled: loading.value,
      min: 0, // 最小值限制
      noValidate: true // 非必选字段
    }
  ]
  return fields
})

// 获取选中的菜单标签
const getSelectedMenuLabel = () => {
  if (!formData.menu_id) return ''
  const findMenu = (menus, id) => {
    const found = menus.find(menu => menu.value === id)
    if (found) return found.label
    
    for (const menu of menus) {
      if (menu.children && menu.children.length > 0) {
        const found = findMenu(menu.children, id)
        if (found) return found
      }
    }
    return ''
  }
  return findMenu(props.menuTreeData, formData.menu_id) || ''
}

// 处理菜单选择
const handleMenuSelect = (data) => {
  formData.menu_id = data.value
  menuSelectVisible.value = false
}

// 监听editId变化，加载详情（
watch(() => props.editId, async (newId) => {
  if (newId && dialogVisible.value) {
    await loadDetail(newId)
  } else if (!newId && dialogVisible.value) {
    resetForm()
  }
}, { immediate: true })

// 监听对话框显隐
watch(dialogVisible, (visible) => {
  if (visible) {
    if (props.editId) {
      loadDetail(props.editId)
    } else {
      resetForm()
    }
  }
})

// 加载权限详情
const loadDetail = async (id) => {
  loading.value = true
  try {
    const res = await getPermissionDetail(id)
    
    if (res.data && res.data.permission) {
      const permission = res.data.permission
      
      const mappedData = {
        id: permission.id || permission.ID,
        name: permission.Name || permission.name || '',
        slug: permission.Slug || permission.slug || '',
        method: permission.Method || permission.method || 'GET',
        path: permission.Path || permission.path || '',
        description: permission.Description || permission.description || '',
        menu_id: permission.MenuID !== undefined ? permission.MenuID : (permission.menu_id !== undefined ? permission.menu_id : null),
        status: permission.Status !== undefined ? permission.Status : (permission.status !== undefined ? permission.status : 1),
        sort: permission.Sort !== undefined ? permission.Sort : (permission.sort !== undefined ? permission.sort : 0)
      }
      
      Object.assign(formData, mappedData)
    }
  } catch (error) {
    console.error('Load permission detail error:', error)
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  loading.value = false
  formData.id = null
  formData.menu_id = null
  formData.name = ''
  formData.slug = ''
  formData.method = 'GET'
  formData.path = ''
  formData.description = ''
  formData.status = 1
  formData.sort = 0
  formRef.value?.resetFields()
}

// 提交表单（保留原有逻辑）
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        // 准备提交数据，将 null 转换为 0
        const submitData = {
          ...formData,
          menu_id: formData.menu_id || 0
        }
        if (formData.id) {
          await updatePermission(formData.id, submitData)
          ElMessage.success(t('permission.update_success'))
        } else {
          await createPermission(submitData)
          ElMessage.success(t('permission.create_success'))
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
}

// 对话框关闭时重置表单
const handleDialogClose = () => {
  formRef.value?.resetFields()
}
</script>

<style scoped>
.tree-node-label {
  font-size: 14px;
}
</style>

<style>
.menu-select-popover {
  max-height: 400px;
  overflow: hidden;
}

.menu-select-tree {
  max-height: 400px;
  overflow-y: auto;
}
</style>