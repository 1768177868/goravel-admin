<template>
  <div class="captcha-config">
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
      label-position="left"
    >
      <el-form-item :label="$t('config.captcha_enabled')" prop="captcha_enabled">
        <el-switch
          v-model="formData.captcha_enabled"
          :active-text="$t('common.enabled')"
          :inactive-text="$t('common.disabled')"
        />
        <span style="margin-left: 10px; color: #909399;">{{ $t('config.captcha_enabled_tip') }}</span>
      </el-form-item>

      <el-form-item :label="$t('config.captcha_expire')" prop="captcha_expire">
        <el-input-number
          v-model="formData.captcha_expire"
          :min="60"
          :max="600"
          :placeholder="$t('config.captcha_expire_placeholder')"
        />
        <span style="margin-left: 10px; color: #909399;">{{ $t('config.captcha_expire_unit') }}</span>
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
  captcha_enabled: false,
  captcha_expire: 120
})

const formRules = {
  captcha_expire: [
    { required: true, message: t('config.captcha_expire_required'), trigger: 'blur' }
  ]
}

const loadData = async () => {
  try {
    const res = await getConfigByGroup('captcha')
    if (res.data && res.data.configs) {
      const configs = res.data.configs
      configs.forEach(config => {
        const key = config.Key || config.key
        let value = config.Value || config.value || ''
        
        if (key === 'captcha_enabled') {
          value = value === '1' || value === 'true' || value === true
        } else if (key === 'captcha_expire') {
          value = value ? parseInt(value) : 120
        }
        
        if (formData.hasOwnProperty(key)) {
          formData[key] = value
        }
      })
    }
  } catch (error) {
    console.error('Load captcha config error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const configs = {
          captcha_enabled: formData.captcha_enabled ? '1' : '0',
          captcha_expire: formData.captcha_expire
        }

        await saveConfig('captcha', configs)
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
.captcha-config {
  padding: 20px 0;
}
</style>

