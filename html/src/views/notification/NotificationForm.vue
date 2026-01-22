<template>
  <el-dialog v-model="dialogVisible" :title="$t('notification.create')" width="800px" @close="handleDialogClose">
    <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
      <FormField v-for="f in basicFormFields" :key="f.prop" :field="f" :model="formData" />

      <el-form-item :label="$t('notification.table.content')" prop="content">
        <WangEditor v-model="formData.content" :placeholder="$t('notification.content_placeholder')" :height="400" />
      </el-form-item>

      <!-- <FormField
        v-for="f in titleFormField"
        :key="f.prop"
        :field="f"
        :model="formData"
      /> -->

    </el-form>
    <template #footer>
      <el-button @click="handleCancel">
        {{ $t('common.cancel') }}
      </el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ $t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import WangEditor from '../../components/WangEditor.vue'
import FormField from '../../components/Form/FormField.vue'
import { createNotification } from '../../api/notification'
import { useNotificationStore } from '../../store/notification'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const notificationStore = useNotificationStore()
const formRef = ref(null)
const submitting = ref(false)

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const formData = reactive({
  type: 'announcement',
  receiver_id: '',
  title: '',
  content: ''
})

const formRules = computed(() => ({
  type: [
    { required: true, message: t('notification.type_required'), trigger: 'change' }
  ],
  receiver_id: [
    {
      validator: (rule, value, callback) => {
        if (formData.type === 'message' && !value) {
          callback(new Error(t('notification.receiver_required')))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ],
  title: [
    { required: true, message: t('notification.title_required'), trigger: 'blur' },
    { max: 150, message: t('notification.title_max_length'), trigger: 'blur' }
  ],
  content: [
    { required: true, message: t('notification.content_required'), trigger: 'blur' }
  ]
}))

const basicFormFields = computed(() => {
  void formData.type // 依赖，使 type 切换时 receiver_id 的 visible 能更新
  return [
    {
      prop: 'type',
      label: t('notification.table.type'),
      type: 'radio',
      options: [
        { label: t('notification.types.announcement'), value: 'announcement' },
        { label: t('notification.types.notice'), value: 'notice' },
        { label: t('notification.types.message'), value: 'message' }
      ]
    },
    {
      prop: 'receiver_id',
      label: t('notification.receiver'),
      type: 'select',
      apiUrl: '/options?type=admin',
      placeholder: t('notification.select_receiver'),
      filterable: true,
      clearable: true,
      visible: () => formData.type === 'message'
    },
    {
      prop: 'title',
      label: t('notification.table.title'),
      type: 'input',
      placeholder: t('notification.title_placeholder'),
      props: { maxlength: 150, showWordLimit: true }
    }
  ]
})

// const titleFormField = computed(() => [
//   {
//     prop: 'title',
//     label: t('notification.table.title'),
//     type: 'input',
//     placeholder: t('notification.title_placeholder'),
//     props: { maxlength: 150, showWordLimit: true }
//   }
// ])

// 监听类型变化，如果不是私信则清空接收者
watch(() => formData.type, (newType) => {
  if (newType !== 'message') {
    formData.receiver_id = ''
    // 清除接收者字段的验证
    if (formRef.value) {
      formRef.value.clearValidate('receiver_id')
    }
  }
})

const resetForm = () => {
  formData.type = 'announcement'
  formData.receiver_id = ''
  formData.title = ''
  formData.content = ''
  formRef.value?.resetFields()
}

const handleDialogClose = () => {
  resetForm()
}

const handleCancel = () => {
  dialogVisible.value = false
}

const handleSubmit = async () => {
  if (!formRef.value) {
    return
  }

  await formRef.value.validate(async (valid) => {
    if (!valid) {
      return false
    }

    submitting.value = true
    try {
      const data = {
        type: formData.type,
        title: formData.title.trim(),
        content: formData.content
      }

      // 如果是私信，必须添加接收者ID
      if (formData.type === 'message') {
        if (!formData.receiver_id) {
          ElMessage.error(t('notification.receiver_required'))
          submitting.value = false
          return
        }
        data.receiver_id = formData.receiver_id
      }
      // 公告和通知不传receiver_id，后端会发送给所有人

      await createNotification(data)
      ElMessage.success(t('notification.create_success'))
      dialogVisible.value = false
      emit('success')

      // 刷新未读数量
      await notificationStore.fetchUnread()
    } catch (error) {
      console.error('Create notification error:', error)
      if (!error.__handled) {
        const errorMessage = error.response?.data?.message || error.message || t('error.default')
        ElMessage.error(errorMessage)
      }
    } finally {
      submitting.value = false
    }
  })
}

// 当对话框打开时重置表单
watch(dialogVisible, (visible) => {
  if (visible) {
    resetForm()
    // 清除验证（延迟执行，确保表单已渲染）
    setTimeout(() => {
      if (formRef.value) {
        formRef.value.clearValidate()
      }
    }, 100)
  }
})
</script>
