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
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
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
          <el-option
            v-for="item in <<.Name>>Options"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
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
        <el-input v-model="form.<<.Name>>" :placeholder="$t('<<$.ModuleName>>.<<.Name>>')" />
<<else if eq .FormType "file-upload">>
        <el-input v-model="form.<<.Name>>" :placeholder="$t('<<$.ModuleName>>.<<.Name>>')" />
<<end>>
      </el-form-item>
<<- end>>
<<- end>>
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
  <<if .HasCreate>>create<<.ModelName>>,<<end>>
  <<if .HasEdit>>update<<.ModelName>>,<<end>>
  get<<.ModelName>>Detail
} from '../../api/<<.ModuleName>>'
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

<<range .FormFields>>
<<if eq .FormType "select">>
const <<.Name>>Options = ref([])
<<end>>
<<- end>>

const visible = ref(props.modelValue)
watch(() => props.modelValue, (val) => {
  visible.value = val
})
watch(visible, (val) => {
  emit('update:modelValue', val)
})

const form = ref({
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
  <<.Name>>: null,
<<- end>>
<<- end>>
})

const rules = {
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
  <<.Name>>: [
    { required: <<.Required>>, message: t('<<$.ModuleName>>.<<.Name>>_required'), trigger: 'blur' }
  ],
<<- end>>
<<- end>>
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true

    if (props.editId) {
      <<if .HasEdit>>
      await update<<.ModelName>>(props.editId, form.value)
      ElMessage.success(t('common.update_success'))
      <<else>>
      ElMessage.error(t('common.operation_failed'))
      <<end>>
    } else {
      <<if .HasCreate>>
      await create<<.ModelName>>(form.value)
      ElMessage.success(t('common.create_success'))
      <<else>>
      ElMessage.error(t('common.operation_failed'))
      <<end>>
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
<<range .FormFields>>
<<if eq .FormType "select">>
  try {
    <<if .ApiUrl>>
    const res = await request({
      url: '<<.ApiUrl>>',
      method: 'get'
    })
    if (res.data) {
       // 适配不同的返回格式，假设返回 list 或 直接是数组
       const list = res.data.list || res.data || []
       <<.Name>>Options.value = list.map(item => ({
         label: item.label || item.name || item.title,
         value: item.value || item.id
       }))
    }
    <<else if .Dictionary>>
    const res = await getOptions('dictionary', { dictionary_type: '<<.Dictionary>>' })
    if (res.data) {
      <<.Name>>Options.value = res.data
    }
    <<else if .Relation>>
    // 加载关联数据: <<.Relation.Table>>
    // 假设存在列表接口 /<<.Relation.Table>> (kebab-case)
    const res = await request({
      url: '/<<.Relation.Table>>'.replace(/_/g, '-'), 
      method: 'get',
      params: { page: 1, page_size: 100 }
    })
    if (res.data && res.data.list) {
      <<.Name>>Options.value = res.data.list.map(item => ({
        label: item.<<.Relation.DisplayField>>,
        value: item.id // 假设关联表主键是 id
      }))
    }
    <<else>>
    // 未配置数据源，请自行实现
    // const res = await getOptions('<<.Name>>')
    // <<.Name>>Options.value = res.data
    <<end>>
  } catch (error) {
    console.error('Failed to load <<.Name>> options:', error)
  }
<<end>>
<<- end>>
}

const loadData = async () => {
  if (!props.editId) return
  
  try {
    const res = await get<<.ModelName>>Detail(props.editId)
    if (res.data && res.data.<<.ModuleName>>) {
      const data = res.data.<<.ModuleName>>
      form.value = {
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
        <<.Name>>: data.<<.Name>>,
<<- end>>
<<- end>>
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
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
        <<.Name>>: null,
<<- end>>
<<- end>>
      }
    }
  }
})
</script>

<style scoped>

</style>