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

      <el-form-item :label="$t('article.title')" prop="title">

        <el-input v-model="form.title" :placeholder="$t('article.title')" />

      </el-form-item>
      <el-form-item :label="$t('article.content')" prop="content">

        <el-input
          v-model="form.content"
          type="textarea"
          :rows="4"
          :placeholder="$t('article.content')" />

      </el-form-item>
      <el-form-item :label="$t('article.status')" prop="status">

        <el-switch v-model="form.status" />

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
import {
  createArticle,
  updateArticle,
  getArticleDetail
} from '../../api/article'
import { getOptions } from '../../api/option'
import request from '../../utils/request'
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






const admin_idOptions = ref([])




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

const loadOptions = async () => {





  try {
    
    // 加载关联数据: admins
    // 假设存在列表接口 /admins (kebab-case)
    const res = await request({
      url: '/admins'.replace(/_/g, '-'), 
      method: 'get',
      params: { page: 1, page_size: 100 }
    })
    if (res.data && res.data.list) {
      admin_idOptions.value = res.data.list.map(item => ({
        label: item.username,
        value: item.id // 假设关联表主键是 id
      }))
    }
    
  } catch (error) {
    console.error('Failed to load admin_id options:', error)
  }



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
    loadOptions()
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