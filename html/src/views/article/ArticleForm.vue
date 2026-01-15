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

      <el-form-item :label="$t('article.name')" prop="name">

        <el-input v-model="form.name" :placeholder="$t('article.name')" />

      </el-form-item>
      <el-form-item :label="$t('article.status')" prop="status">

        <el-input v-model="form.status" :placeholder="$t('article.status')" />

      </el-form-item>
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
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  createArticle,
  updateArticle,
  getArticleDetail
} from '../../api/article'
import { getOptions } from '../../api/option'
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





const visible = ref(props.modelValue)
watch(() => props.modelValue, (val) => {
  visible.value = val
})
watch(visible, (val) => {
  emit('update:modelValue', val)
})

const form = ref({

  name: null,
  status: null,
})

const rules = {

  name: [
    { required: true, message: t('article.name_required'), trigger: 'blur' }
  ],
  status: [
    { required: false, message: t('article.status_required'), trigger: 'blur' }
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

const handleImageSuccess = (response) => {
  form.value.image = response.data.url
}

const handleFileSuccess = (response) => {
  form.value.file = response.data.url
}

const loadOptions = async () => {



}

const loadData = async () => {
  if (!props.editId) return
  
  try {
    const res = await getArticleDetail(props.editId)
    if (res.data && res.data.article) {
      const data = res.data.article
      form.value = {
        name: data.name,
        status: data.status,
      }
    }
  } catch (error) {
    ErrorHandler.handle(error)
  }
}

watch(visible, (val) => {
  if (val) {
    loadOptions()
    if (props.editId) {
      loadData()
    } else {
      formRef.value?.resetFields()
      form.value = {
        name: null,
        status: null,
      }
    }
  }
})
</script>

<style scoped>
.image-uploader {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  width: 178px;
  height: 178px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-uploader:hover {
  border-color: #409EFF;
}

.image-preview {
  width: 178px;
  height: 178px;
  object-fit: cover;
}

.image-uploader-icon {
  font-size: 28px;
  color: #8c939d;
}
</style>