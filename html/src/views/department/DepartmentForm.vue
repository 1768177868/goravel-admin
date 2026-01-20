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
      <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ $t('common.confirm') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import FormField from '../../components/Form/FormField.vue'
import {
  getDepartmentDetail,
  createDepartment,
  updateDepartment
} from '../../api/department'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  editId: { type: [Number, String], default: null },
  departmentOptions: { type: Array, default: () => [] }
})

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const formRef = ref(null)
const submitting = ref(false)
const loading = ref(false)

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const dialogTitle = computed(() => formData.id ? t('department.edit_department') : t('department.add_department'))

const formData = reactive({
  id: null,
  parent_id: 0,
  name: '',
  description: '',
  status: 1,
  sort: 0
})

const formRules = computed(() => ({
  name: [{ required: true, message: t('department.name_required'), trigger: 'blur' }]
}))

const formFields = computed(() => [
  {
    prop: 'parent_id',
    label: t('department.parent_department'),
    type: 'select',
    options: [
      { label: t('department.top_department'), value: 0 },
      ...(props.departmentOptions || []).map(d => ({ label: d.name, value: d.id }))
    ],
    placeholder: t('form.select_parent') + t('department.parent_department'),
    clearable: false,
    disabled: loading.value
  },
  { prop: 'name', label: t('department.name'), type: 'input', disabled: loading.value },
  { prop: 'description', label: t('common.description'), type: 'textarea', disabled: loading.value },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'radio',
    options: [
      { label: t('common.enabled'), value: 1 },
      { label: t('common.disabled'), value: 0 }
    ],
    disabled: loading.value
  },
  { prop: 'sort', label: t('common.sort'), type: 'number', min: 0, disabled: loading.value }
])

watch(() => props.editId, async (newId) => {
  if (newId && dialogVisible.value) await loadDetail(newId)
  else if (!newId && dialogVisible.value) resetForm()
}, { immediate: true })

watch(dialogVisible, (visible) => {
  if (visible) {
    if (props.editId) loadDetail(props.editId)
    else resetForm()
  }
})

const loadDetail = async (id) => {
  loading.value = true
  try {
    const res = await getDepartmentDetail(id)
    if (res.data?.department) {
      const dept = res.data.department
      Object.assign(formData, {
        id: dept.id,
        parent_id: dept.ParentID !== undefined ? dept.ParentID : (dept.parent_id || 0),
        name: dept.Name || dept.name || '',
        description: dept.Remark || dept.remark || dept.description || '',
        status: dept.Status !== undefined ? dept.Status : (dept.status ?? 1),
        sort: dept.Sort !== undefined ? dept.Sort : (dept.sort ?? 0)
      })
    }
  } catch (e) {
    console.error('Load department detail error:', e)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  loading.value = false
  Object.assign(formData, { id: null, parent_id: 0, name: '', description: '', status: 1, sort: 0 })
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const data = {
        name: formData.name,
        remark: formData.description,
        status: formData.status,
        sort: formData.sort,
        parent_id: formData.parent_id === 0 ? null : formData.parent_id
      }
      if (formData.id) {
        await updateDepartment(formData.id, data)
        ElMessage.success(t('department.update_success'))
      } else {
        await createDepartment(data)
        ElMessage.success(t('department.create_success'))
      }
      dialogVisible.value = false
      emit('success')
    } catch (e) {
      console.error('Submit error:', e)
    } finally {
      submitting.value = false
    }
  })
}

const handleCancel = () => { dialogVisible.value = false }
const handleDialogClose = () => { formRef.value?.resetFields() }
</script>
