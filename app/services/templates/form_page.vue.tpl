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
<<range .FormFields>>
      <el-form-item :label="$t('<<$.ModuleName>>.<<.Name>>')" prop="<<.Name>>">
<<if eq .FormType "input">>
        <el-input v-model="form.<<.Name>>" :placeholder="$t('<<$.ModuleName>>.<<.Name>>')" />
<<else if eq .FormType "textarea">>
        <el-input
          v-model="form.<<.Name>>"
          type="textarea"
          :rows="4"
          :placeholder="$t('<<$.ModuleName>>.<<.Name>>')" />
<<else if eq .FormType "select">>
        <el-select v-model="form.<<.Name>>" :placeholder="$t('common.select')">
          <el-option label="Option 1" :value="1" />
          <el-option label="Option 2" :value="2" />
        </el-select>
<<else if eq .FormType "switch">>
        <el-switch v-model="form.<<.Name>>" />
<<else if eq .FormType "date-picker">>
        <el-date-picker
          v-model="form.<<.Name>>"
          type="date"
          :placeholder="$t('common.select_date')" style="width: 100%" />
<<else if eq .FormType "datetime-picker">>
        <el-date-picker
          v-model="form.<<.Name>>"
          type="datetime"
          :placeholder="$t('common.select_datetime')" style="width: 100%" />
<<else if eq .FormType "image-upload">>
        <el-upload
          class="image-uploader"
          action="/api/admin/attachments/upload"
          :show-file-list="false"
          :on-success="handleImageSuccess"
        >
          <img v-if="form.<<.Name>>" :src="form.<<.Name>>" class="image-preview" />
          <el-icon v-else class="image-uploader-icon"><Plus /></el-icon>
        </el-upload>
<<else if eq .FormType "file-upload">>
        <el-upload
          action="/api/admin/attachments/upload"
          :show-file-list="true"
          :on-success="handleFileSuccess"
        >
          <el-button type="primary">{{ $t('common.upload') }}</el-button>
        </el-upload>
<<end>>
      </el-form-item>
<<end>>
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
  create<<.ModelName>>,
  update<<.ModelName>>
} from '../../api/<<.ModuleName>>'
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
<<range .FormFields>>
  <<.Name>>: null,
<<end>>
})

const rules = {
<<range .FormFields>>
  <<.Name>>: [
    { required: <<.Required>>, message: t('<<$.ModuleName>>.<<.Name>>_required'), trigger: 'blur' }
  ],
<<end>>
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true

    if (props.editId) {
      await update<<.ModelName>>(props.editId, form.value)
      ElMessage.success(t('common.update_success'))
    } else {
      await create<<.ModelName>>(form.value)
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