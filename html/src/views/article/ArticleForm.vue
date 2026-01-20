<template>
  <el-dialog
    v-model="visible"
    :title="editId ? $t('common.edit') : $t('common.add')"
    width="600px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
    >
      <FormField
        v-for="f in formFields"
        :key="f.prop"
        :field="f"
        :model="form"
        i18n-prefix="article"
      />
    </el-form>

    <template #footer>
      <el-button @click="handleClose">{{ $t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ $t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import FormField from '../../components/Form/FormField.vue'
import {
  createArticle,
  updateArticle,
  getArticleDetail
} from '../../api/article'
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
const submitting = ref(false)

// 表单项配置：下拉/单选/多选支持 apiUrl，与搜索一致；也可用 options、optionsFn
const formFields = computed(() => [
  { prop: 'title', label: t('article.title'), type: 'input' },
  { prop: 'content', label: t('article.content'), type: 'textarea', rows: 4 },
  {
    prop: 'status',
    label: t('article.status'),
    type: 'select',
    apiUrl: '/options?type=dictionary&dictionary_type=status'
    // 也可: apiUrl: '/options?type=status'
  },
  {
    prop: 'admin_id',
    label: t('article.admin_id'),
    type: 'select',
    apiUrl: '/admins',
    apiParams: { page: 1, page_size: 100 },
    optionLabelKey: 'username',
    optionValueKey: 'id'
  }
])




const visible = ref(props.modelValue)
watch(() => props.modelValue, (val) => {
  visible.value = val
})
watch(visible, (val) => {
  emit('update:modelValue', val)
})

const form = ref({

  title: null,
  content: null,
  status: null,
  admin_id: null,
})

const rules = {

  title: [
    { required: true, message: t('article.title_required'), trigger: 'blur' }
  ],
  content: [
    { required: false, message: t('article.content_required'), trigger: 'blur' }
  ],
  status: [
    { required: true, message: t('article.status_required'), trigger: 'blur' }
  ],
  admin_id: [
    { required: true, message: t('article.admin_id_required'), trigger: 'blur' }
  ],
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true

    if (props.editId) {
      
      await updateArticle(props.editId, form.value)
      ElMessage.success(t('common.update_success'))
      
    } else {
      
      await createArticle(form.value)
      ElMessage.success(t('common.create_success'))
      
    }

    emit('success')
    handleClose()
  } catch (error) {
    ErrorHandler.handle(error)
  } finally {
    submitting.value = false
  }
}

const handleClose = () => {
  visible.value = false
  formRef.value?.resetFields()
}

const loadData = async () => {
  if (!props.editId) return
  
  try {
    const res = await getArticleDetail(props.editId)
    if (res.data && res.data.article) {
      const data = res.data.article
      form.value = {

        title: data.title,
        content: data.content,
        status: data.status,
        admin_id: data.admin_id,
      }
    }
  } catch (error) {
    ErrorHandler.handle(error)
  }
}

watch(visible, (val) => {
  if (val) {
    if (props.editId) {
      loadData()
    } else {
      formRef.value?.resetFields()
      form.value = {

        title: null,
        content: null,
        status: null,
        admin_id: null,
      }
    }
  }
})
</script>

<style scoped>

</style>