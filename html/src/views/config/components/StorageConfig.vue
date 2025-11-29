<template>
  <div class="storage-config">
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="150px"
      label-position="left"
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.storage_disk')" prop="storage_disk">
            <el-select v-model="formData.storage_disk" :placeholder="$t('config.storage_disk_placeholder')">
              <el-option label="local" value="local" />
              <el-option label="s3" value="s3" />
              <el-option label="oss" value="oss" />
              <el-option label="cos" value="cos" />
              <el-option label="qiniu" value="qiniu" />
              <el-option label="minio" value="minio" />
            </el-select>
          </el-form-item>
        </el-col>
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
          </el-form-item>
        </el-col>
      </el-row>

      <!-- S3 配置项 -->
      <template v-if="formData.storage_disk === 's3' || formData.file_disk === 's3'">
        <el-divider content-position="left">S3 配置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.s3_key')" prop="s3_key">
              <el-input v-model="formData.s3_key" :placeholder="$t('config.s3_key_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.s3_secret')" prop="s3_secret">
              <el-input v-model="formData.s3_secret" :placeholder="$t('config.s3_secret_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.s3_region')" prop="s3_region">
              <el-input v-model="formData.s3_region" :placeholder="$t('config.s3_region_placeholder')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.s3_bucket')" prop="s3_bucket">
              <el-input v-model="formData.s3_bucket" :placeholder="$t('config.s3_bucket_placeholder')" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.s3_url')" prop="s3_url">
              <el-input v-model="formData.s3_url" :placeholder="$t('config.s3_url_placeholder')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.s3_use_path_style')" prop="s3_use_path_style">
              <el-switch v-model="formData.s3_use_path_style" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">{{ $t('config.s3_use_path_style_tip') }}</span>
            </el-form-item>
          </el-col>
        </el-row>
      </template>

      <!-- OSS 配置项 -->
      <template v-if="formData.storage_disk === 'oss' || formData.file_disk === 'oss'">
        <el-divider content-position="left">阿里云 OSS 配置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.oss_key')" prop="oss_key">
              <el-input v-model="formData.oss_key" :placeholder="$t('config.oss_key_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.oss_secret')" prop="oss_secret">
              <el-input v-model="formData.oss_secret" :placeholder="$t('config.oss_secret_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.oss_bucket')" prop="oss_bucket">
              <el-input v-model="formData.oss_bucket" :placeholder="$t('config.oss_bucket_placeholder')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.oss_endpoint')" prop="oss_endpoint">
              <el-input v-model="formData.oss_endpoint" :placeholder="$t('config.oss_endpoint_placeholder')" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.oss_url')" prop="oss_url">
              <el-input v-model="formData.oss_url" :placeholder="$t('config.oss_url_placeholder')" />
            </el-form-item>
          </el-col>
        </el-row>
      </template>

      <!-- COS 配置项 -->
      <template v-if="formData.storage_disk === 'cos' || formData.file_disk === 'cos'">
        <el-divider content-position="left">腾讯云 COS 配置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.cos_key')" prop="cos_key">
              <el-input v-model="formData.cos_key" :placeholder="$t('config.cos_key_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.cos_secret')" prop="cos_secret">
              <el-input v-model="formData.cos_secret" :placeholder="$t('config.cos_secret_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.cos_bucket')" prop="cos_bucket">
              <el-input v-model="formData.cos_bucket" :placeholder="$t('config.cos_bucket_placeholder')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.cos_region')" prop="cos_region">
              <el-input v-model="formData.cos_region" :placeholder="$t('config.cos_region_placeholder')" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.cos_url')" prop="cos_url">
              <el-input v-model="formData.cos_url" :placeholder="$t('config.cos_url_placeholder')" />
            </el-form-item>
          </el-col>
        </el-row>
      </template>

      <!-- 七牛云 配置项 -->
      <template v-if="formData.storage_disk === 'qiniu' || formData.file_disk === 'qiniu'">
        <el-divider content-position="left">七牛云 配置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.qiniu_key')" prop="qiniu_key">
              <el-input v-model="formData.qiniu_key" :placeholder="$t('config.qiniu_key_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.qiniu_secret')" prop="qiniu_secret">
              <el-input v-model="formData.qiniu_secret" :placeholder="$t('config.qiniu_secret_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.qiniu_bucket')" prop="qiniu_bucket">
              <el-input v-model="formData.qiniu_bucket" :placeholder="$t('config.qiniu_bucket_placeholder')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.qiniu_domain')" prop="qiniu_domain">
              <el-input v-model="formData.qiniu_domain" :placeholder="$t('config.qiniu_domain_placeholder')" />
            </el-form-item>
          </el-col>
        </el-row>
      </template>

      <!-- MinIO 配置项 -->
      <template v-if="formData.storage_disk === 'minio' || formData.file_disk === 'minio'">
        <el-divider content-position="left">MinIO 配置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.minio_key')" prop="minio_key">
              <el-input v-model="formData.minio_key" :placeholder="$t('config.minio_key_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.minio_secret')" prop="minio_secret">
              <el-input v-model="formData.minio_secret" :placeholder="$t('config.minio_secret_placeholder')" type="password" show-password />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.minio_bucket')" prop="minio_bucket">
              <el-input v-model="formData.minio_bucket" :placeholder="$t('config.minio_bucket_placeholder')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.minio_region')" prop="minio_region">
              <el-input v-model="formData.minio_region" :placeholder="$t('config.minio_region_placeholder')" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.minio_endpoint')" prop="minio_endpoint">
              <el-input v-model="formData.minio_endpoint" :placeholder="$t('config.minio_endpoint_placeholder')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('config.minio_url')" prop="minio_url">
              <el-input v-model="formData.minio_url" :placeholder="$t('config.minio_url_placeholder')" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('config.minio_ssl')" prop="minio_ssl">
              <el-switch v-model="formData.minio_ssl" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">{{ $t('config.minio_ssl_tip') }}</span>
            </el-form-item>
          </el-col>
        </el-row>
      </template>

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

// 定义所有密钥字段（涉密信息）
const secretFields = [
  's3_key',
  's3_secret',
  'oss_key',
  'oss_secret',
  'cos_key',
  'cos_secret',
  'qiniu_key',
  'qiniu_secret',
  'minio_key',
  'minio_secret'
]

// 记录原始密钥字段是否已设置（用于判断是否跳过更新）
const originalSecretFields = reactive({})

const formData = reactive({
  storage_disk: 'local',
  file_disk: 'local',
  // S3 配置
  s3_key: '',
  s3_secret: '',
  s3_region: '',
  s3_bucket: '',
  s3_url: '',
  s3_use_path_style: false,
  // OSS 配置
  oss_key: '',
  oss_secret: '',
  oss_bucket: '',
  oss_endpoint: '',
  oss_url: '',
  // COS 配置
  cos_key: '',
  cos_secret: '',
  cos_bucket: '',
  cos_region: '',
  cos_url: '',
  // 七牛云配置
  qiniu_key: '',
  qiniu_secret: '',
  qiniu_bucket: '',
  qiniu_domain: '',
  // MinIO 配置
  minio_key: '',
  minio_secret: '',
  minio_bucket: '',
  minio_region: '',
  minio_endpoint: '',
  minio_url: '',
  minio_ssl: false
})

const formRules = {
  storage_disk: [
    { required: true, message: t('config.storage_disk_required'), trigger: 'change' }
  ],
  file_disk: [
    { required: true, message: t('config.file_disk_required'), trigger: 'change' }
  ]
}

const loadData = async () => {
  try {
    // 重置原始密钥字段记录
    secretFields.forEach(field => {
      originalSecretFields[field] = false
    })

    const res = await getConfigByGroup('storage')
    if (res.data && res.data.configs) {
      const configs = res.data.configs
      configs.forEach(config => {
        const key = config.Key || config.key
        const value = config.Value || config.value || ''
        
        // 兼容旧的 export_disk 字段名
        if (key === 'export_disk') {
          formData.file_disk = value
        } else if (formData.hasOwnProperty(key)) {
          // 处理布尔值
          if (key === 's3_use_path_style' || key === 'minio_ssl') {
            formData[key] = value === 'true' || value === true || value === '1' || value === 1
          } else if (secretFields.includes(key)) {
            // 对于密钥字段，如果数据库有值，记录为已设置但不显示真实值
            if (value && value.trim() !== '') {
              originalSecretFields[key] = true
              formData[key] = '' // 不显示真实值
            } else {
              originalSecretFields[key] = false
              formData[key] = ''
            }
          } else {
            formData[key] = value
          }
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
          // 对于密钥字段，如果已设置过但传的是空值，就不处理（不更新）
          if (secretFields.includes(key)) {
            if (formData[key] && formData[key].trim() !== '') {
              // 有新值，更新
              configs[key] = formData[key]
            } else if (originalSecretFields[key]) {
              // 已设置过但传的是空值，跳过不更新
              // 不添加到 configs 中，这样后端就不会更新这个字段
            } else {
              // 从未设置过且为空，可以设置为空
              configs[key] = ''
            }
          } else {
            // 非密钥字段，正常处理
            configs[key] = formData[key]
          }
        })

        await saveConfig('storage', configs)
        ElMessage.success(t('config.update_success'))
        // 提交成功后重新加载数据，更新原始密钥字段记录
        await loadData()
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
