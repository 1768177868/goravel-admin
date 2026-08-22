<template>
  <div class="customer-service-config">
    <el-form
      ref="formRef"
      :model="formData"
      label-width="140px"
      label-position="left"
    >
      <el-form-item :label="$t('config.cs_enabled')" prop="cs_enabled">
        <el-switch
          v-model="formData.cs_enabled"
          active-value="1"
          inactive-value="0"
          :active-text="$t('common.enabled')"
          :inactive-text="$t('common.disabled')"
        />
      </el-form-item>

      <el-form-item :label="$t('config.cs_work_time')" prop="cs_work_time">
        <el-input
          v-model="formData.cs_work_time"
          :placeholder="$t('config.cs_work_time_placeholder')"
          style="max-width: 480px"
        />
      </el-form-item>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.cs_phone')" prop="cs_phone">
            <el-input v-model="formData.cs_phone" :placeholder="$t('config.cs_phone_placeholder')" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('config.cs_email')" prop="cs_email">
            <el-input v-model="formData.cs_email" :placeholder="$t('config.cs_email_placeholder')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.cs_wechat')" prop="cs_wechat">
            <el-input v-model="formData.cs_wechat" :placeholder="$t('config.cs_wechat_placeholder')" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('config.cs_qq')" prop="cs_qq">
            <el-input v-model="formData.cs_qq" :placeholder="$t('config.cs_qq_placeholder')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item :label="$t('config.cs_wechat_qr')" prop="cs_wechat_qr">
        <AttachmentImageField
          v-model="formData.cs_wechat_qr"
          :placeholder="$t('config.cs_wechat_qr_placeholder')"
        />
      </el-form-item>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.cs_telegram')" prop="cs_telegram">
            <el-input v-model="formData.cs_telegram" :placeholder="$t('config.cs_telegram_placeholder')" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('config.cs_whatsapp')" prop="cs_whatsapp">
            <el-input v-model="formData.cs_whatsapp" :placeholder="$t('config.cs_whatsapp_placeholder')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item :label="$t('config.cs_online_url')" prop="cs_online_url">
        <el-input
          v-model="formData.cs_online_url"
          :placeholder="$t('config.cs_online_url_placeholder')"
          style="max-width: 640px"
        />
      </el-form-item>

      <el-form-item :label="$t('config.cs_custom_link')" prop="cs_custom_link">
        <el-input
          v-model="formData.cs_custom_link"
          :placeholder="$t('config.cs_custom_link_placeholder')"
          style="max-width: 640px"
        />
      </el-form-item>

      <el-form-item :label="$t('config.cs_remark')" prop="cs_remark">
        <el-input
          v-model="formData.cs_remark"
          type="textarea"
          :rows="3"
          :placeholder="$t('config.cs_remark_placeholder')"
          style="max-width: 640px"
        />
      </el-form-item>

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
import { forOwn } from 'lodash-es'
import { getConfigByGroup, saveConfig } from '../../../api/config'
import { usePermission } from '../../../composables/usePermission'
import AttachmentImageField from '../../../components/AttachmentImageField.vue'

const { t } = useI18n()
const { getButtonState } = usePermission()
const formRef = ref(null)
const submitting = ref(false)

const formData = reactive({
  cs_enabled: '1',
  cs_work_time: '',
  cs_phone: '',
  cs_email: '',
  cs_wechat: '',
  cs_wechat_qr: '',
  cs_qq: '',
  cs_telegram: '',
  cs_whatsapp: '',
  cs_online_url: '',
  cs_custom_link: '',
  cs_remark: ''
})

const loadData = async () => {
  try {
    const res = await getConfigByGroup('customer_service')
    if (res.data && res.data.configs) {
      res.data.configs.forEach(config => {
        const key = config.Key || config.key
        const value = config.Value || config.value || ''
        if (Object.prototype.hasOwnProperty.call(formData, key)) {
          formData[key] = value
        }
      })
    }
  } catch (error) {
    console.error('Load customer service config error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const configs = {}
        forOwn(formData, (value, key) => {
          configs[key] = value
        })

        await saveConfig('customer_service', configs)
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
.customer-service-config {
  padding: 20px 0;
}
</style>
