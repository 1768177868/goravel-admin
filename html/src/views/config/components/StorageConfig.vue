<template>
  <div class="storage-config">
    <el-alert
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 20px;"
    >
      <template #title>
        <div>
          <p style="margin: 0 0 8px 0; font-weight: 500;">{{ $t('config.storage_config_title') }}</p>
          <ul style="margin: 0; padding-left: 20px;">
            <li>{{ $t('config.storage_config_desc_1') }}</li>
            <li>{{ $t('config.storage_config_desc_2') }}</li>
            <li>{{ $t('config.storage_config_desc_3') }}</li>
          </ul>
        </div>
      </template>
    </el-alert>

    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="150px"
      label-position="left"
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.file_disk')" prop="file_disk">
            <el-select v-model="formData.file_disk" :placeholder="$t('config.file_disk_placeholder')">
              <el-option label="local" value="local" />
              <el-option label="s3" value="s3" />
              <el-option label="oss" value="oss" />
              <el-option label="cos" value="cos" />
              <el-option label="qiniu" value="qiniu" />
              <el-option label="minio" value="minio" />
            </el-select>
            <div style="margin-top: 8px; color: #909399; font-size: 12px;">
              {{ $t('config.storage_config_default_tip') }}
            </div>
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item>
        <el-button type="primary" @click="handleSubmit" :loading="submitting" :disabled="getButtonState('config.save').disabled">
          {{ $t('common.save') }}
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getConfigByGroup, saveConfig } from '../../../api/config'
import { usePermission } from '../../../composables/usePermission'

const { t } = useI18n()
const { getButtonState } = usePermission()
const formRef = ref(null)
const submitting = ref(false)

const formData = reactive({
  file_disk: 'local'
})

const formRules = {
  file_disk: [
    { required: true, message: t('config.file_disk_required'), trigger: 'change' }
  ]
}

const loadData = async () => {
  try {
    const res = await getConfigByGroup('storage')
    if (res.data && res.data.configs) {
      const configs = res.data.configs
      configs.forEach(config => {
        const key = config.Key || config.key
        const value = config.Value || config.value || ''
        
        // 兼容旧的字段名：export_disk 和 storage_disk
        if (key === 'file_disk' || key === 'export_disk' || key === 'storage_disk') {
          formData.file_disk = value
        }
      })
    }
  } catch (error) {
    console.error('Load storage config error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        // 只保存驱动选择，不保存其他配置
        const configs = {
          file_disk: formData.file_disk
        }

        await saveConfig('storage', configs)
        ElMessage.success(t('config.update_success'))
      } catch (error) {
        console.error('Submit error:', error)
        ElMessage.error(t('config.update_failed'))
      } finally {
        submitting.value = false
      }
    }
  })
}

onMounted(() => {
  loadData()
})

defineExpose({
  loadData
})
</script>

<style scoped>
.storage-config {
  padding: 20px 0;
}

code {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}
</style>
