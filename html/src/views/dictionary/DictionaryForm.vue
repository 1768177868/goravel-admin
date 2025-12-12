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
      <el-form-item :label="$t('dictionary.type')" prop="type">
        <el-input v-model="formData.type" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('dictionary.label')" prop="label">
        <el-input v-model="formData.label" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('dictionary.value')" prop="value">
        <el-input v-model="formData.value" :disabled="loading" />
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
import {
  getDictionaryDetail,
  createDictionary,
  updateDictionary
} from '../../api/dictionary'

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

const dialogTitle = computed(() => formData.id ? t('dictionary.edit_dictionary') : t('dictionary.add_dictionary'))

const formData = reactive({
  id: null,
  type: '',
  label: '',
  value: '',
  status: 1,
  sort: 0
})

const formRules = computed(() => ({
  type: [{ required: true, message: t('dictionary.type_required'), trigger: 'blur' }],
  label: [{ required: true, message: t('dictionary.label_required'), trigger: 'blur' }],
  value: [{ required: true, message: t('dictionary.value_required'), trigger: 'blur' }]
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
    const res = await getDictionaryDetail(id)
    if (res.data && res.data.dictionary) {
      const dict = res.data.dictionary
      // 处理字段映射，支持 PascalCase 和 snake_case
      Object.assign(formData, {
        id: dict.id,
        type: dict.Type || dict.type || '',
        label: dict.Label || dict.label || '',
        value: dict.Value || dict.value || '',
        status: dict.Status !== undefined ? dict.Status : (dict.status !== undefined ? dict.status : 1),
        sort: dict.Sort !== undefined ? dict.Sort : (dict.sort !== undefined ? dict.sort : 0)
      })
    }
  } catch (error) {
    console.error('Load dictionary detail error:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  loading.value = false
  Object.assign(formData, {
    id: null,
    type: '',
    label: '',
    value: '',
    status: 1,
    sort: 0
  })
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (formData.id) {
          await updateDictionary(formData.id, formData)
          ElMessage.success(t('dictionary.update_success'))
        } else {
          await createDictionary(formData)
          ElMessage.success(t('dictionary.create_success'))
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

