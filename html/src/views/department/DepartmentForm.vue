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
        <!-- 父部门树形选择插槽 -->
        <el-form-item :label="$t('department.parent_department')">
          <el-tree-select
            v-model="formData.parent_id"
            :data="treeSelectData"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            :placeholder="$t('form.select_parent') + $t('department.parent_department')"
            clearable
            check-strictly
            :render-after-expand="false"
            :disabled="loading"
            style="width: 100%"
          />
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
import FormField from '../../components/Form/FormField.vue'
import { getEnableDisableOptions } from '@/utils/options'
import { excludeNodeAndChildren } from '../../utils/tree'
import {
  getDepartmentDetail,
  createDepartment,
  updateDepartment
} from '../../api/department'
import { mapFields } from '../../utils/normalizeFormData'

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

// 树形选择数据，排除当前部门及其子部门
const treeSelectData = computed(() => {
  if (!props.departmentOptions || props.departmentOptions.length === 0) {
    return [{ id: 0, name: t('department.top_department') }]
  }
  
  // 如果有编辑ID，需要排除当前部门及其所有子部门
  if (formData.id) {
    // 使用工具函数排除当前部门及其子部门
    const filtered = excludeNodeAndChildren(props.departmentOptions, formData.id, 'id', 'children')
    return [{ id: 0, name: t('department.top_department') }, ...filtered]
  }
  
  // 新增时，直接使用树形结构
  return [{ id: 0, name: t('department.top_department') }, ...props.departmentOptions]
})

const formFields = computed(() => [
  { prop: 'name', label: t('department.name'), type: 'input', disabled: loading.value },
  { prop: 'description', label: t('common.description'), type: 'textarea', disabled: loading.value },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'radio',
    options: getEnableDisableOptions(t),
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
      // 使用工具函数映射字段，自动处理 snake_case 和 PascalCase
      const mapped = mapFields(dept, {
        id: null,
        parent_id: 0,
        name: '',
        status: 1,
        sort: 0
      })
      // 处理 description 字段的特殊映射（Remark -> description）
      mapped.description = dept.Remark ?? dept.remark ?? dept.description ?? ''
      Object.assign(formData, mapped)
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
      // 只处理需要转换的字段
      const data = {
        ...formData,
        remark: formData.description, // description -> remark
        parent_id: formData.parent_id === 0 ? null : formData.parent_id // 0 -> null
      }
      // 删除前端使用的 description 字段
      delete data.description
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
