<template>
  <div class="storage-config">
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
      label-position="left"
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.storage_disk')" prop="storage_disk">
            <el-select v-model="formData.storage_disk" :placeholder="$t('config.storage_disk_placeholder')">
              <el-option label="local" value="local" />
              <el-option label="public" value="public" />
              <el-option label="s3" value="s3" />
              <el-option label="oss" value="oss" />
              <el-option label="cos" value="cos" />
              <el-option label="minio" value="minio" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('config.export_disk')" prop="export_disk">
            <el-select v-model="formData.export_disk" :placeholder="$t('config.export_disk_placeholder')">
              <el-option label="local" value="local" />
              <el-option label="public" value="public" />
              <el-option label="s3" value="s3" />
              <el-option label="oss" value="oss" />
              <el-option label="cos" value="cos" />
              <el-option label="minio" value="minio" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.export_path')" prop="export_path">
            <el-input v-model="formData.export_path" :placeholder="$t('config.export_path_placeholder')" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('config.export_format')" prop="export_format">
            <el-select v-model="formData.export_format" :placeholder="$t('config.export_format_placeholder')">
              <el-option label="CSV" value="csv" />
              <el-option label="Excel" value="xlsx" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item :label="$t('config.export_url_prefix')" prop="export_url_prefix">
        <el-input v-model="formData.export_url_prefix" :placeholder="$t('config.export_url_prefix_placeholder')" />
        <span style="margin-left: 10px; color: #909399;">{{ $t('config.export_url_prefix_tip') }}</span>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
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

const { t } = useI18n()
const formRef = ref(null)
const submitting = ref(false)

const formData = reactive({
  storage_disk: 'local',
  export_disk: 'public',
  export_path: 'exports',
  export_format: 'csv',
  export_url_prefix: ''
})

const formRules = {
  storage_disk: [
    { required: true, message: t('config.storage_disk_required'), trigger: 'change' }
  ],
  export_disk: [
    { required: true, message: t('config.export_disk_required'), trigger: 'change' }
  ],
  export_path: [
    { required: true, message: t('config.export_path_required'), trigger: 'blur' }
  ],
  export_format: [
    { required: true, message: t('config.export_format_required'), trigger: 'change' }
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
        
        if (formData.hasOwnProperty(key)) {
          formData[key] = value
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
        const configs = {}
        Object.keys(formData).forEach(key => {
          configs[key] = formData[key]
        })

        await saveConfig('storage', configs)
        ElMessage.success(t('config.update_success'))
      } catch (error) {
        console.error('Submit error:', error)
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
</style>

