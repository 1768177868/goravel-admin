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
      <el-form-item :label="$t('table.username')" prop="username">
        <el-input v-model="formData.username" :disabled="!!formData.id || loading" />
      </el-form-item>
      <el-form-item :label="$t('common.password')" prop="password" v-if="!formData.id">
        <el-input v-model="formData.password" type="password" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('table.nickname')" prop="nickname">
        <el-input v-model="formData.nickname" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('table.email')" prop="email">
        <el-input v-model="formData.email" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('table.phone')" prop="phone">
        <el-input v-model="formData.phone" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('table.department')" prop="department_id">
        <el-popover
          placement="bottom-start"
          :width="300"
          trigger="click"
          v-model="departmentSelectVisible"
        >
          <template #reference>
            <el-input
              :model-value="getDepartmentName(formData.department_id)"
              :placeholder="$t('form.select_department')"
              readonly
              :disabled="loading"
              @click="departmentSelectVisible = !departmentSelectVisible"
              style="cursor: pointer"
            >
              <template #suffix>
                <el-icon class="el-input__icon">
                  <ArrowDownIcon />
                </el-icon>
              </template>
            </el-input>
          </template>
          <el-tree
            :data="departmentTree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            :default-expand-all="false"
            :expand-on-click-node="false"
            :highlight-current="true"
            @node-click="handleDepartmentSelect"
            style="max-height: 300px; overflow-y: auto;"
          >
            <template #default="{ node, data }">
              <span class="custom-tree-node" style="flex: 1; display: flex; align-items: center; justify-content: space-between; font-size: 14px; padding-right: 8px;">
                <span>{{ node.label }}</span>
              </span>
            </template>
          </el-tree>
        </el-popover>
      </el-form-item>
      <el-form-item :label="$t('table.roles')" prop="role_ids">
        <el-select 
          v-model="formData.role_ids" 
          multiple 
          :placeholder="$t('form.select_role')" 
          :disabled="isDefaultAdmin || loading"
          style="width: 100%"
        >
          <el-option
            v-for="role in roles"
            :key="role.id || role.ID"
            :label="role.Name || role.name"
            :value="role.id || role.ID"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('table.status')" prop="status">
        <el-radio-group v-model="formData.status" :disabled="isDefaultAdmin || loading">
          <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
          <el-radio :label="0">{{ $t('common.disabled') }}</el-radio>
        </el-radio-group>
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
import { ref, reactive, computed, watch, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import {
  getAdminDetail,
  createAdmin,
  updateAdmin
} from '../../api/admin'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  editId: {
    type: [Number, String],
    default: null
  },
  departmentTree: {
    type: Array,
    default: () => []
  },
  roles: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const formRef = ref(null)
const submitting = ref(false)
const loading = ref(false)
const departmentSelectVisible = ref(false)

const ArrowDownIcon = markRaw(ArrowDown)

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const dialogTitle = computed(() => formData.id ? t('admin.edit_admin') : t('admin.add_admin'))

const formData = reactive({
  id: null,
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  department_id: null,
  role_ids: [],
  status: 1,
  is_super_admin: false
})

const isDefaultAdmin = computed(() => {
  return formData.is_super_admin === true && formData.id !== null
})

const formRules = computed(() => ({
  username: [{ required: true, message: t('admin.username_required'), trigger: 'blur' }],
  password: [{ required: true, message: t('admin.password_required'), trigger: 'blur' }]
}))

// 获取部门名称
const getDepartmentName = (departmentId) => {
  if (!departmentId) return ''
  const findDept = (depts, id) => {
    for (const dept of depts) {
      if (dept.id === id) {
        return dept.name
      }
      if (dept.children && dept.children.length > 0) {
        const found = findDept(dept.children, id)
        if (found) return found
      }
    }
    return ''
  }
  return findDept(props.departmentTree, departmentId) || ''
}

// 处理部门选择
const handleDepartmentSelect = (data) => {
  if (data && data.id) {
    formData.department_id = data.id
    departmentSelectVisible.value = false
  }
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
    // Admin模块的编辑不需要单独获取详情，数据已经在列表中
    // 这里可以根据需要实现
  } catch (error) {
    logger.error('Load admin detail error:', error)
    ErrorHandler.handle(error, { silent: true })
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  loading.value = false
  Object.assign(formData, {
    id: null,
    username: '',
    password: '',
    nickname: '',
    email: '',
    phone: '',
    department_id: null,
    role_ids: [],
    status: 1,
    is_super_admin: false
  })
  formRef.value?.resetFields()
}

// 设置表单数据（从外部调用）
const setFormData = async (data) => {
  loading.value = true
  try {
    // 使用 nextTick 确保 DOM 更新后再设置数据
    await new Promise(resolve => setTimeout(resolve, 50))
    Object.assign(formData, data)
  } finally {
    loading.value = false
  }
}

// 暴露方法供父组件调用
defineExpose({
  setFormData
})

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const data = { ...formData }
        
        if (data.username) {
          data.username = data.username.trim()
        }
        
        if (formData.id) {
          if (!data.password) {
            delete data.password
          }
          // 如果是默认 admin 用户，不发送 role_ids 字段
          if (isDefaultAdmin.value) {
            delete data.role_ids
          }
          await updateAdmin(formData.id, data)
          ElMessage.success(t('admin.update_success'))
        } else {
          await createAdmin(data)
          ElMessage.success(t('admin.create_success'))
        }
        dialogVisible.value = false
        emit('success')
      } catch (error) {
        logger.error('Submit error:', error)
        // 如果错误已经在响应拦截器中处理过，就不再重复显示
        if (!error.__handled) {
          // 显示更详细的错误信息
          if (error.response && error.response.data && error.response.data.message) {
            ElMessage.error(error.response.data.message)
          } else if (error.message) {
            ElMessage.error(error.message)
          }
        }
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

