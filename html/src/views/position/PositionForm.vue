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
        <FormField
          v-for="f in formFields"
          :key="f.prop"
          :field="f"
          :model="formData"
        />
      </el-form>
    </div>
    <template #footer>
      <el-button @click="handleCancel">{{ $t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import FormField from '../../components/Form/FormField.vue'
import { getEnableDisableOptions } from '@/utils/options'
import { getPositionDetail, createPosition, updatePosition } from '../../api/position'
import { mapFields } from '../../utils/normalizeFormData'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  editId: { type: [Number, String], default: null }
})

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const formRef = ref(null)
const submitting = ref(false)
const loading = ref(false)

const getFormInitialValue = () => ({
  id: null,
  name: '',
  code: '',
  status: 1,
  sort: 0,
  remark: ''
})

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const dialogTitle = computed(() =>
  formData.id ? t('position.edit_position') : t('position.add_position')
)

const formData = reactive(getFormInitialValue())

const formRules = computed(() => ({
  name: [{ required: true, message: t('position.name_required'), trigger: 'blur' }]
}))

const formFields = computed(() => [
  { prop: 'name', label: t('position.name'), type: 'input', disabled: loading.value },
  { prop: 'code', label: t('position.code'), type: 'input', disabled: loading.value, noValidate: true },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'radio',
    options: getEnableDisableOptions(t),
    disabled: loading.value
  },
  { prop: 'sort', label: t('common.sort'), type: 'number', disabled: loading.value, noValidate: true },
  { prop: 'remark', label: t('position.remark'), type: 'textarea', disabled: loading.value, noValidate: true }
])

watch(
  () => props.editId,
  async (newId) => {
    if (newId && dialogVisible.value) {
      await loadDetail(newId)
    } else if (!newId && dialogVisible.value) {
      resetForm()
    }
  },
  { immediate: true }
)

watch(dialogVisible, (visible) => {
  if (visible) {
    if (props.editId) {
      loadDetail(props.editId)
    } else {
      resetForm()
    }
  }
})

async function loadDetail(id) {
  loading.value = true
  try {
    const res = await getPositionDetail(id)
    const raw = res.data?.position
    if (raw) {
      Object.assign(
        formData,
        mapFields(raw, {
          id: null,
          name: '',
          code: '',
          status: 1,
          sort: 0,
          remark: ''
        })
      )
    }
  } finally {
    loading.value = false
  }
}

function resetForm() {
  Object.assign(formData, getFormInitialValue())
  formRef.value?.resetFields()
}

function handleDialogClose() {
  resetForm()
}

function handleCancel() {
  dialogVisible.value = false
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload = {
        name: formData.name?.trim(),
        code: formData.code?.trim() || '',
        status: formData.status,
        sort: Number(formData.sort) || 0,
        remark: formData.remark || ''
      }
      if (formData.id) {
        await updatePosition(formData.id, payload)
        ElMessage.success(t('position.update_success'))
      } else {
        await createPosition(payload)
        ElMessage.success(t('position.create_success'))
      }
      dialogVisible.value = false
      emit('success')
    } finally {
      submitting.value = false
    }
  })
}
</script>
