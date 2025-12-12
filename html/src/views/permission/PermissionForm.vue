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
      <el-form-item :label="$t('permission.name')" prop="name">
        <el-input v-model="formData.name" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('permission.slug')" prop="slug">
        <el-input v-model="formData.slug" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('permission.method')" prop="method">
        <el-select v-model="formData.method" :placeholder="$t('form.select_method')" :disabled="loading">
          <el-option label="GET" value="GET" />
          <el-option label="POST" value="POST" />
          <el-option label="PUT" value="PUT" />
          <el-option label="DELETE" value="DELETE" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('permission.path')" prop="path">
        <el-input v-model="formData.path" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('common.description')">
        <el-input v-model="formData.description" type="textarea" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('menu.title')" prop="menu_id">
        <el-popover
          v-model:visible="menuSelectVisible"
          placement="bottom-start"
          :width="300"
          trigger="click"
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
          >
            <template #default="{ node, data }">
              <span class="tree-node-label">{{ data.label }}</span>
            </template>
          </el-tree>
        </el-popover>
      </el-form-item>
      <el-form-item :label="$t('table.status')" prop="status">
        <el-radio-group v-model="formData.status" :disabled="loading">
          <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
          <el-radio :label="0">{{ $t('common.disabled') }}</el-radio>
        </el-radio-group>
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
import { ArrowDown } from '@element-plus/icons-vue'
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

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const dialogTitle = computed(() => formData.id ? t('permission.edit_permission') : t('permission.add_permission'))

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

const formRules = computed(() => ({
  name: [{ required: true, message: t('permission.name_required'), trigger: 'blur' }],
  slug: [{ required: true, message: t('permission.slug_required'), trigger: 'blur' }],
  method: [{ required: true, message: t('permission.method_required'), trigger: 'change' }],
  path: [{ required: true, message: t('permission.path_required'), trigger: 'blur' }]
}))

// 获取选中的菜单标签
const getSelectedMenuLabel = () => {
  if (!formData.menu_id) return ''
  const findMenu = (menus, id) => {
    for (const menu of menus) {
      if (menu.value === id) {
        return menu.label
      }
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

// 监听 editId 变化，加载详情
watch(() => props.editId, async (newId) => {
  if (newId && dialogVisible.value) {
    await loadDetail(newId)
  } else if (!newId && dialogVisible.value) {
    // 新增模式，重置表单
    resetForm()
  }
}, { immediate: true })

// 监听 dialogVisible 变化
watch(dialogVisible, (visible) => {
  if (visible) {
    if (props.editId) {
      loadDetail(props.editId)
    } else {
      resetForm()
    }
  }
})

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

const handleCancel = () => {
  dialogVisible.value = false
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}
</script>

<style scoped>
.tree-node-label {
  font-size: 14px;
}
</style>

