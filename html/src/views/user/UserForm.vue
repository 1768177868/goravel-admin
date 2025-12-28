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
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model="formData.status" :disabled="loading">
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
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { createUser, updateUser, getUserDetail } from '../../api/user'
import ErrorHandler from '../../utils/errorHandler'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  editId: {
    type: Number,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const formRef = ref(null)
const loading = ref(false)
const submitting = ref(false)

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const dialogTitle = computed(() => {
  return props.editId ? t('user.edit_user') : t('user.add_user')
})

const formData = reactive({
  id: null,
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  status: 0 // 默认禁用状态
})

// 动态验证规则：根据是否为编辑模式调整
const formRules = computed(() => {
  const rules = {
    email: [
      { type: 'email', message: t('form.email_invalid'), trigger: 'blur' }
    ]
  }
  
  // 创建时，username 和 password 是必填的
  if (!props.editId) {
    rules.username = [
      { required: true, message: t('form.username_required'), trigger: 'blur' },
      { min: 3, message: t('form.username_min'), trigger: 'blur' },
      { max: 50, message: t('form.username_max'), trigger: 'blur' }
    ]
    rules.password = [
      { required: true, message: t('form.password_required'), trigger: 'blur' },
      { min: 6, message: t('form.password_min_length'), trigger: 'blur' },
      { max: 50, message: t('form.password_max_length'), trigger: 'blur' }
    ]
  } else {
    // 编辑时，username 和 password 不是必填的（username 不可修改，password 可选）
    rules.password = [
      { min: 6, message: t('form.password_min_length'), trigger: 'blur' },
      { max: 50, message: t('form.password_max_length'), trigger: 'blur' }
    ]
  }
  
  return rules
})

watch(() => props.editId, async (newId) => {
  if (newId && dialogVisible.value) {
    await loadData()
  } else {
    resetForm()
  }
})

watch(dialogVisible, (val) => {
  if (val && props.editId) {
    loadData()
  } else if (!val) {
    resetForm()
  }
})

const loadData = async () => {
  if (!props.editId) {
    resetForm()
    return
  }

  loading.value = true
  try {
    const res = await getUserDetail(props.editId)
    if (res.code === 200 && res.data && res.data.user) {
      const user = res.data.user
      // 兼容大小写字段名，确保 status 是数字类型
      const userStatus = user.status !== undefined ? user.status : (user.Status !== undefined ? user.Status : 1)
      Object.assign(formData, {
        id: user.id || user.ID || null,
        username: user.username || user.Username || '',
        password: '', // 编辑时不填充密码
        nickname: user.nickname || user.Nickname || '',
        email: user.email || user.Email || '',
        phone: user.phone || user.Phone || '',
        status: Number(userStatus) // 确保是数字类型
      })
    }
  } catch (error) {
    ErrorHandler.handle(error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(formData, {
    id: null,
    username: '',
    password: '',
    nickname: '',
    email: '',
    phone: '',
    status: 0 // 默认禁用状态
  })
  formRef.value?.resetFields()
}

const handleDialogClose = () => {
  resetForm()
}

const handleCancel = () => {
  dialogVisible.value = false
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (props.editId) {
        // 编辑时，不提交 username（因为不可修改）
        // 确保 status 是数字类型（0 或 1）
        const updateData = { 
          ...formData,
          status: Number(formData.status) || 0
        }
        delete updateData.username
        delete updateData.id
        // 确保 status 是 0 或 1
        updateData.status = updateData.status === 1 ? 1 : 0
        await updateUser(props.editId, updateData)
        ElMessage.success(t('common.update_success'))
      } else {
        // 创建时，提交所有数据（包括 username）
        // 确保 status 是数字类型（0 或 1）
        const createData = { 
          ...formData,
          status: Number(formData.status) || 0
        }
        delete createData.id
        // 确保 status 是 0 或 1
        createData.status = createData.status === 1 ? 1 : 0
        await createUser(createData)
        ElMessage.success(t('common.create_success'))
      }
      emit('success')
      dialogVisible.value = false
    } catch (error) {
      ErrorHandler.handle(error)
    } finally {
      submitting.value = false
    }
  })
}
</script>

