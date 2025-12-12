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
      <el-form-item :label="$t('department.parent_department')">
        <el-select v-model="formData.parent_id" :placeholder="$t('form.select_parent') + $t('department.parent_department')" clearable :disabled="loading">
          <el-option :label="$t('department.top_department')" :value="0" />
          <el-option
            v-for="dept in departmentOptions"
            :key="dept.id"
            :label="dept.name"
            :value="dept.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('department.name')" prop="name">
        <el-input v-model="formData.name" :disabled="loading" />
      </el-form-item>
      <el-form-item :label="$t('common.description')">
        <el-input v-model="formData.description" type="textarea" :disabled="loading" />
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
  getDepartmentDetail,
  createDepartment,
  updateDepartment
} from '../../api/department'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  editId: {
    type: [Number, String],
    default: null
  },
  departmentOptions: {
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
    const res = await getDepartmentDetail(id)
    if (res.data && res.data.department) {
      const dept = res.data.department
      // 后端返回的是 PascalCase 字段，需要正确映射
      Object.assign(formData, {
        id: dept.id,
        parent_id: dept.ParentID !== undefined ? dept.ParentID : (dept.parent_id || 0),
        name: dept.Name || dept.name || '',
        description: dept.Remark || dept.remark || dept.description || '',
        status: dept.Status !== undefined ? dept.Status : (dept.status !== undefined ? dept.status : 1),
        sort: dept.Sort !== undefined ? dept.Sort : (dept.sort !== undefined ? dept.sort : 0)
      })
    }
  } catch (error) {
    console.error('Load department detail error:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  loading.value = false
  Object.assign(formData, {
    id: null,
    parent_id: 0,
    name: '',
    description: '',
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
        // 转换前端字段名为后端期望的字段名
        const data = {
          name: formData.name,
          remark: formData.description, // description 映射到 remark
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

