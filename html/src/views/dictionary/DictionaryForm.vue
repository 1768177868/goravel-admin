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
// 引入配置式表单组件
import FormField from '../../components/Form/FormField.vue'
import { getEnableDisableOptions } from '@/utils/options'

import {
  getDictionaryDetail,
  createDictionary,
  updateDictionary
} from '../../api/dictionary'
import { mapFields } from '../../utils/normalizeFormData'

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

// 对话框显隐状态
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 对话框标题
const dialogTitle = computed(() => formData.id ? t('dictionary.edit_dictionary') : t('dictionary.add_dictionary'))

// 表单数据
const formData = reactive({
  id: null,
  type: '',
  label: '',
  value: '',
  translation_key: '',
  status: 1,
  sort: 0
})

// 表单验证规则
const formRules = computed(() => ({
  type: [{ required: true, message: t('dictionary.type_required'), trigger: 'blur' }],
  label: [{ required: true, message: t('dictionary.label_required'), trigger: 'blur' }],
  value: [{ required: true, message: t('dictionary.value_required'), trigger: 'blur' }]
}))

// 配置式表单字段
const formFields = computed(() => {
  const fields = [
    {
      prop: 'type',
      label: t('dictionary.type'),
      type: 'input',
      disabled: loading.value
    },
    {
      prop: 'label',
      label: t('dictionary.label'),
      type: 'input',
      disabled: loading.value
    },
    {
      prop: 'value',
      label: t('dictionary.value'),
      type: 'input',
      disabled: loading.value
    },
    {
      prop: 'translation_key',
      label: t('dictionary.translation_key'),
      type: 'input',
      disabled: loading.value,
      noValidate: true // 无验证规则，无需校验
    },
    {
      prop: 'status',
      label: t('table.status'),
      type: 'radio',
      disabled: loading.value,
      // 配置radio选项
      options: getEnableDisableOptions(t),
    },
    {
      prop: 'sort',
      label: t('common.sort'),
      type: 'number', // 兼容 input-number，FormField 已支持两种类型
      disabled: loading.value,
      min: 0, // 最小值限制，透传给 el-input-number
      noValidate: true // 原代码无prop，无需校验
    }
  ]
  return fields
})

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

// 加载字典详情
const loadDetail = async (id) => {
  loading.value = true
  try {
    const res = await getDictionaryDetail(id)
    if (res.data && res.data.dictionary) {
      const dict = res.data.dictionary
      // 使用工具函数映射字段，自动处理 snake_case 和 PascalCase
      const mapped = mapFields(dict, {
        id: null,
        type: '',
        label: '',
        value: '',
        translation_key: '',
        status: 1,
        sort: 0
      })
      Object.assign(formData, mapped)
    }
  } catch (error) {
    console.error('Load dictionary detail error:', error)
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  loading.value = false
  Object.assign(formData, {
    id: null,
    type: '',
    label: '',
    value: '',
    translation_key: '',
    status: 1,
    sort: 0
  })
  formRef.value?.resetFields()
}

// 提交表单
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

// 取消按钮
const handleCancel = () => {
  dialogVisible.value = false
}

// 对话框关闭时重置表单（
const handleDialogClose = () => {
  formRef.value?.resetFields()
}
</script>