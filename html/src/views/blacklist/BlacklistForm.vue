<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="700px"
    @close="handleDialogClose"
  >
    <div v-loading="loading">
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
      <el-form-item :label="$t('blacklist.ip')" prop="ip">
        <el-input
          v-model="formData.ip"
          type="textarea"
          :rows="4"
          :placeholder="$t('blacklist.ip_placeholder')"
          :disabled="loading"
        />
        <div style="margin-top: 8px; color: #909399; font-size: 12px;">
          {{ $t('blacklist.ip_tip') }}
        </div>
      </el-form-item>
      <el-form-item :label="$t('blacklist.remark')" prop="remark">
        <el-input
          v-model="formData.remark"
          type="textarea"
          :rows="3"
          :placeholder="$t('blacklist.remark_placeholder')"
          :disabled="loading"
        />
      </el-form-item>
      <el-form-item :label="$t('table.status')" prop="status">
        <el-radio-group v-model="formData.status" :disabled="loading">
          <el-radio :label="1">{{ $t('blacklist.enabled') }}</el-radio>
          <el-radio :label="0">{{ $t('blacklist.disabled') }}</el-radio>
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
import {
  getBlacklistDetail,
  createBlacklist,
  updateBlacklist
} from '../../api/blacklist'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  editId: {
    type: [Number, String],
    default: null
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

const dialogTitle = computed(() => formData.id ? t('blacklist.edit_blacklist') : t('blacklist.add_blacklist'))

const formData = reactive({
  id: null,
  ip: '',
  remark: '',
  status: 1
})

const formRules = computed(() => ({
  ip: [
    { required: true, message: t('blacklist.ip_required'), trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value || value.trim() === '') {
          callback(new Error(t('blacklist.ip_required')))
          return
        }
        // 前端简单验证，后端会做详细验证
        const ipList = value.split(',')
        for (const ip of ipList) {
          const trimmedIP = ip.trim()
          if (trimmedIP === '') continue
          // 简单检查：至少包含点或斜杠或横线
          if (!trimmedIP.includes('.') && !trimmedIP.includes('/') && !trimmedIP.includes('-')) {
            callback(new Error(t('blacklist.ip_format_error')))
            return
          }
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}))

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
    const res = await getBlacklistDetail(id)
    if (res.data && res.data.blacklist) {
      const blacklist = res.data.blacklist
      Object.assign(formData, {
        id: blacklist.id,
        ip: blacklist.IP || blacklist.ip || '',
        remark: blacklist.Remark || blacklist.remark || '',
        status: blacklist.Status !== undefined ? blacklist.Status : (blacklist.status !== undefined ? blacklist.status : 1)
      })
    }
  } catch (error) {
    console.error('Load blacklist detail error:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  loading.value = false
  Object.assign(formData, {
    id: null,
    ip: '',
    remark: '',
    status: 1
  })
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const submitData = {
          ip: formData.ip.trim(),
          remark: formData.remark.trim(),
          status: formData.status
        }
        if (formData.id) {
          await updateBlacklist(formData.id, submitData)
          ElMessage.success(t('blacklist.update_success'))
        } else {
          await createBlacklist(submitData)
          ElMessage.success(t('blacklist.create_success'))
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

