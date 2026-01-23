<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="1000px"
    @close="handleDialogClose"
  >
    <div v-loading="loading">
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
        <<range .FormFields>>
        <<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
        <<- if .ShowInForm>>
        <<- if or (eq .FormType "image-upload") (eq .FormType "file-upload") (eq .FormType "editor") (eq .FormType "markdown")>>
        <el-form-item :label="$t('<<$.ModuleName>>.<<.Name>>')" prop="<<.Name>>">
          <<- if eq .FormType "image-upload">>
          <!-- 图片上传组件，请根据实际需求实现 -->
          <el-upload
            v-model="formData.<<.Name>>"
            action="/api/upload/image"
            :show-file-list="false"
          >
            <el-button type="primary">上传图片</el-button>
          </el-upload>
          <<- else if eq .FormType "file-upload">>
          <!-- 文件上传组件，请根据实际需求实现 -->
          <el-upload
            v-model="formData.<<.Name>>"
            action="/api/upload/file"
            :show-file-list="false"
          >
            <el-button type="primary">上传文件</el-button>
          </el-upload>
          <<- else if eq .FormType "editor">>
          <WangEditor
            v-model="formData.<<.Name>>"
            :placeholder="$t('<<$.ModuleName>>.<<.Name>>_placeholder')"
            :height="400"
          />
          <<- else if eq .FormType "markdown">>
          <MarkdownEditor
            v-model="formData.<<.Name>>"
            :placeholder="$t('<<$.ModuleName>>.<<.Name>>_placeholder')"
            :height="400"
          />
          <<- end>>
        </el-form-item>
        <<- end>>
        <<- end>>
        <<- end>>
        <<- end>>
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
<<if .HasEditor>>
import WangEditor from '../../components/WangEditor.vue'
<<end>>
<<if .HasMarkdown>>
import MarkdownEditor from '../../components/MarkdownEditor.vue'
<<end>>
<<range .FormFields>>
<<- if and (eq .FormType "select") (or .Relation .ApiUrl)>>
<<- if .Relation>>
// 关联字段: <<.Name>> -> <<.Relation.Table>>
<<- end>>
<<- end>>
<<- end>>
import {
  <<if .HasCreate>>create<<.ModelName>>,<<end>>
  <<if .HasEdit>>update<<.ModelName>>,<<end>>
  get<<.ModelName>>Detail
} from '../../api/<<.ModuleName>>'
import { mapFields, normalizeFormData } from '../../utils/normalizeFormData'
import ErrorHandler from '../../utils/errorHandler'

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
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
  <<.Name>>: <<if eq .FormType "switch">>0<<else if eq .FormType "number">>0<<else if eq .FormType "date-picker">>null<<else if eq .FormType "datetime-picker">>null<<else if .Relation>>null<<else if or (eq .FormType "select") (eq .FormType "radio")>>null<<else if eq .FormType "checkbox">>[]<<else>>''<<end>>,
<<- end>>
<<- end>>
})

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const dialogTitle = computed(() => {
  return formData.id ? t('<<.ModuleName>>.edit_<<.ModuleName>>') : t('<<.ModuleName>>.add_<<.ModuleName>>')
})

const formData = reactive(getFormInitialValue())

const formRules = computed(() => {
  const rules = {}
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
  <<- if .Required>>
  rules['<<.Name>>'] = [
    { required: true, message: t('<<$.ModuleName>>.<<.Name>>_required'), trigger: 'blur' }
  ]
  <<- end>>
<<- end>>
<<- end>>
  return rules
})

// 配置式表单字段
const formFields = computed(() => {
  const fields = []
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
  <<- if and .ShowInForm (ne .FormType "image-upload") (ne .FormType "file-upload") (ne .FormType "editor") (ne .FormType "markdown")>>
  fields.push({
    prop: '<<.Name>>',
    label: t('<<$.ModuleName>>.<<.Name>>'),
    type: <<if eq .FormType "input">>'input'<<else if eq .FormType "textarea">>'textarea'<<else if eq .FormType "select">><<- if and .ApiUrl (or (and .Relation .Relation.IsTree) .IsTree)>>'tree-select'<<else>>'select'<<- end>><<else if eq .FormType "radio">>'radio'<<else if eq .FormType "checkbox">>'checkbox'<<else if eq .FormType "switch">>'switch'<<else if eq .FormType "date-picker">>'date'<<else if eq .FormType "datetime-picker">>'datetime'<<else if eq .FormType "number">>'number'<<else>>'input'<<end>>,
    disabled: loading.value,
    <<- if eq .FormType "textarea">>
    rows: 4,
    <<- end>>
    <<- if .Relation>>
    <<- if .ApiUrl>>
    <<- if .Relation.IsTree>>
    apiUrl: '<<.ApiUrl>>',
    treeProps: { label: '<<.Relation.DisplayField>>', value: 'id', children: 'children' },
    clearable: true,
    <<- else>>
    apiUrl: '<<.ApiUrl>>',
    clearable: true,
    <<- if .Relation.DisplayField>>
    optionLabelKey: '<<.Relation.DisplayField>>',
    <<- end>>
    optionValueKey: 'id',
    <<- end>>
    <<- end>>
    <<- else if .ApiUrl>>
    <<- if or (eq .FormType "select") (eq .FormType "radio") (eq .FormType "checkbox")>>
    apiUrl: '<<.ApiUrl>>',
    <<- if .IsTree>>
    treeProps: { label: 'label', value: 'value', children: 'children' },
    <<- end>>
    clearable: true,
    <<- end>>
    <<- else if or (eq .FormType "select") (eq .FormType "radio") (eq .FormType "checkbox")>>
    <<- if .Dictionary>>
    apiUrl: '/options?type=dictionary&dictionary_type=<<.Dictionary>>',
    clearable: true,
    <<- end>>
    <<- end>>
    <<- if or (eq .FormType "date-picker") (eq .FormType "datetime-picker")>>
    clearable: true,
    <<- end>>
    <<- if eq .FormType "number">>
    min: 0,
    <<- end>>
    <<- if eq .FormType "switch">>
    props: {
      activeValue: 1,
      inactiveValue: 0
    },
    <<- end>>
  })
  <<- end>>
<<- end>>
<<- end>>
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
    const res = await get<<.ModelName>>Detail(props.editId)
    if (res.data && res.data.<<.ModuleName>>) {
      const data = res.data.<<.ModuleName>>
      // 使用工具函数映射字段，自动处理 snake_case 和 PascalCase
      const mapped = mapFields(data, getFormInitialValue())
      // 对于使用字典的字段（radio、select、checkbox），需要将值转换为字符串以匹配选项值
      // 因为字典选项的值都是字符串类型
      const normalizeRules = {}
<<range .FormFields>>
<<- if and (or (eq .FormType "radio") (eq .FormType "select") (eq .FormType "checkbox")) .Dictionary>>
      normalizeRules['<<.Name>>'] = 'string'
<<- end>>
<<- end>>
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
      <<if .HasEdit>>
      if (props.editId) {
        await update<<.ModelName>>(props.editId, data)
        ElMessage.success(t('common.update_success'))
      } else {
        <<if .HasCreate>>
        await create<<.ModelName>>(data)
        ElMessage.success(t('common.create_success'))
        <<end>>
      }
      <<else>>
      <<if .HasCreate>>
      await create<<.ModelName>>(data)
      ElMessage.success(t('common.create_success'))
      <<end>>
      <<end>>
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

<style scoped>
</style>
