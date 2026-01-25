<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    :width="isMobile ? '95%' : '1000px'"
    :fullscreen="isMobile"
    :close-on-click-modal="false"
    @close="handleDialogClose"
    class="article-form-dialog"
  >
    <div v-loading="loading">
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
        
        <el-form-item :label="$t('article.content')" prop="content">
          <MarkdownEditor
            v-model="formData.content"
            :height="isMobile ? 300 : 400"
          />
        </el-form-item>
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
import ImageUpload from '../../components/ImageUpload.vue'
import MarkdownEditor from '../../components/MarkdownEditor.vue'
import {
  createArticle,
  updateArticle,
  getArticleDetail
} from '../../api/article'
import { mapFields, normalizeFormData } from '../../utils/normalizeFormData'
import ErrorHandler from '../../utils/errorHandler'
import { useResponsive } from '../../composables/useResponsive'

const { isMobile } = useResponsive()

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

// 定义表单初始值的复用函数
const getFormInitialValue = () => ({

  title: '',
  content: '',
  status: null,
  admin_id: null,
})

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const dialogTitle = computed(() => {
  return formData.id ? t('article.edit_article') : t('article.add_article')
})

const formData = reactive(getFormInitialValue())

const formRules = computed(() => {
  const rules = {}

  rules['title'] = [
    { required: true, message: t('article.title_required'), trigger: 'blur' }
  ]
  rules['status'] = [
    { required: true, message: t('article.status_required'), trigger: 'blur' }
  ]
  rules['admin_id'] = [
    { required: true, message: t('article.admin_id_required'), trigger: 'blur' }
  ]
  return rules
})

// 配置式表单字段
const formFields = computed(() => {
  const fields = []

  fields.push({
    prop: 'title',
    label: t('article.title'),
    type: 'input',
    disabled: loading.value,
  })
  fields.push({
    prop: 'status',
    label: t('article.status'),
    type: 'radio',
    disabled: loading.value,
    apiUrl: '/options?type=dictionary&dictionary_type=status',
    clearable: true,
  })
  fields.push({
    prop: 'admin_id',
    label: t('article.admin_id'),
    type: 'input',
    disabled: loading.value,
  })
  return fields
})

watch(() => props.editId, async (newId) => {
  if (newId && dialogVisible.value) {
    await loadData()
  } else if (!newId && dialogVisible.value) {
    resetForm()
  }
}, { immediate: true })

watch(dialogVisible, (visible) => {
  if (visible) {
    if (props.editId) {
      loadData()
    } else {
      resetForm()
    }
  }
})

const loadData = async () => {
  if (!props.editId) {
    resetForm()
    return
  }

  loading.value = true
  try {
    const res = await getArticleDetail(props.editId)
    if (res.data && res.data.article) {
      const data = res.data.article
      // 使用工具函数映射字段，自动处理 snake_case 和 PascalCase
      const mapped = mapFields(data, getFormInitialValue())
      // 对于使用字典的字段（radio、select、checkbox），需要将值转换为字符串以匹配选项值
      // 因为字典选项的值都是字符串类型
      const normalizeRules = {}

      normalizeRules['status'] = 'string'
      const normalized = normalizeFormData(mapped, normalizeRules)
      Object.assign(formData, normalized)
    }
  } catch (error) {
    ErrorHandler.handle(error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(formData, getFormInitialValue())
  formRef.value?.resetFields()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
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
      const data = { ...formData }
      delete data.id
      
      if (props.editId) {
        await updateArticle(props.editId, data)
        ElMessage.success(t('common.update_success'))
      } else {
        
        await createArticle(data)
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

<style scoped lang="scss">
.article-form-dialog {
  :deep(.el-dialog__body) {
    padding: 20px;
    max-height: calc(100vh - 200px);
    overflow-y: auto;
  }

  :deep(.el-form-item__label) {
    font-weight: 500;
  }
}

/* 移动端优化 */
@media (max-width: 768px) {
  .article-form-dialog {
    :deep(.el-dialog) {
      margin: 0;
      height: 100vh;
      border-radius: 0;
    }

    :deep(.el-dialog__body) {
      padding: 16px;
      max-height: calc(100vh - 120px);
    }

    :deep(.el-form-item) {
      margin-bottom: 20px;
    }

    :deep(.el-form-item__label) {
      width: 100% !important;
      text-align: left;
      margin-bottom: 8px;
      padding: 0;
    }
  }
}
</style>
