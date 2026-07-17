<template>
  <el-dialog
    :model-value="modelValue"
    :title="$t('attachment.category_manage')"
    width="720px"
    append-to-body
    destroy-on-close
    @update:model-value="$emit('update:modelValue', $event)"
    @opened="loadData"
  >
    <div class="category-toolbar">
      <el-button type="primary" :disabled="getButtonState('attachment_category.store').disabled" @click="openForm()">
        {{ $t('attachment.category_add') }}
      </el-button>
    </div>

    <el-table v-loading="loading" :data="list" border size="small">
      <el-table-column prop="id" :label="$t('table.id')" width="70" />
      <el-table-column prop="name" :label="$t('attachment.category_name')" min-width="140" />
      <el-table-column prop="sort" :label="$t('common.sort')" width="80" />
      <el-table-column :label="$t('common.status')" width="90">
        <template #default="{ row }">
          <el-tag :type="Number(row.status) === 1 ? 'success' : 'info'" size="small">
            {{ Number(row.status) === 1 ? $t('common.enabled') : $t('common.disabled') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('table.operation')" width="160" fixed="right">
        <template #default="{ row }">
          <el-button
            link
            type="primary"
            :disabled="getButtonState('attachment_category.update').disabled"
            @click="openForm(row)"
          >
            {{ $t('common.edit') }}
          </el-button>
          <el-button
            link
            type="danger"
            :disabled="Number(row.is_system) === 1 || getButtonState('attachment_category.destroy').disabled"
            @click="handleDelete(row)"
          >
            {{ $t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      v-model="formVisible"
      :title="formData.id ? $t('attachment.category_edit') : $t('attachment.category_add')"
      width="420px"
      append-to-body
      destroy-on-close
    >
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item :label="$t('attachment.category_name')" prop="name">
          <el-input v-model="formData.name" maxlength="50" />
        </el-form-item>
        <el-form-item :label="$t('common.sort')" prop="sort">
          <el-input-number v-model="formData.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item :label="$t('common.status')" prop="status">
          <el-switch
            v-model="formData.status"
            :active-value="1"
            :inactive-value="0"
            :disabled="Number(formData.is_system) === 1"
          />
        </el-form-item>
        <el-form-item :label="$t('common.description')" prop="remark">
          <el-input v-model="formData.remark" type="textarea" :rows="2" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { usePermission } from '@/composables/usePermission'
import {
  getAttachmentCategoryList,
  createAttachmentCategory,
  updateAttachmentCategory,
  deleteAttachmentCategory
} from '@/api/attachmentCategory'

defineProps({
  modelValue: { type: Boolean, default: false }
})

const emit = defineEmits(['update:modelValue', 'changed'])

const { t } = useI18n()
const { getButtonState } = usePermission()

const loading = ref(false)
const list = ref([])
const formVisible = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const formData = reactive({
  id: null,
  name: '',
  sort: 0,
  status: 1,
  remark: '',
  is_system: 0
})

const formRules = {
  name: [{ required: true, message: t('attachment.category_name_required'), trigger: 'blur' }]
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await getAttachmentCategoryList({ page: 1, page_size: 100 })
    list.value = res?.data?.list || []
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

const openForm = (row = null) => {
  formData.id = row?.id || null
  formData.name = row?.name || ''
  formData.sort = Number(row?.sort || 0)
  formData.status = row ? Number(row.status ?? 1) : 1
  formData.remark = row?.remark || ''
  formData.is_system = Number(row?.is_system || 0)
  formVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload = {
        name: formData.name,
        sort: formData.sort,
        status: formData.status,
        remark: formData.remark
      }
      if (formData.id) {
        await updateAttachmentCategory(formData.id, payload)
      } else {
        await createAttachmentCategory(payload)
      }
      ElMessage.success(t('common.save_success'))
      formVisible.value = false
      await loadData()
      emit('changed')
    } catch (error) {
      console.error(error)
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (row) => {
  if (Number(row.is_system) === 1) {
    ElMessage.warning(t('attachment.category_system_cannot_delete'))
    return
  }
  try {
    await ElMessageBox.confirm(t('attachment.category_delete_confirm'), t('common.confirm'), { type: 'warning' })
    await deleteAttachmentCategory(row.id)
    ElMessage.success(t('common.delete_success'))
    await loadData()
    emit('changed')
  } catch (error) {
    if (error !== 'cancel') console.error(error)
  }
}
</script>

<style scoped>
.category-toolbar {
  margin-bottom: 12px;
}
</style>
