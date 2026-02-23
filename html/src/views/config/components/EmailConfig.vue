<template>
  <div class="email-config">
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
      label-position="left"
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.email_host')" prop="email_host">
            <el-input v-model="formData.email_host" :placeholder="$t('config.email_host_placeholder')" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('config.email_port')" prop="email_port">
            <el-input-number v-model="formData.email_port" :min="1" :max="65535" :placeholder="$t('config.email_port_placeholder')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.email_username')" prop="email_username">
            <el-input v-model="formData.email_username" :placeholder="$t('config.email_username_placeholder')" autocomplete="off" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('config.email_password')" prop="email_password">
            <el-input v-model="formData.email_password" type="password" show-password :placeholder="$t('config.email_password_placeholder')" autocomplete="new-password" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.email_from')" prop="email_from">
            <el-input v-model="formData.email_from" :placeholder="$t('config.email_from_placeholder')" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('config.email_from_name')" prop="email_from_name">
            <el-input v-model="formData.email_from_name" :placeholder="$t('config.email_from_name_placeholder')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('config.email_encryption')" prop="email_encryption">
            <el-select v-model="formData.email_encryption" :placeholder="$t('config.email_encryption_placeholder')">
              <el-option label="TLS" value="tls" />
              <el-option label="SSL" value="ssl" />
              <el-option label="None" value="" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('config.email_timeout')" prop="email_timeout">
            <el-input-number v-model="formData.email_timeout" :min="1" :max="300" :placeholder="$t('config.email_timeout_placeholder')" />
            <span style="margin-left: 10px; color: #909399;">{{ $t('config.email_timeout_unit') }}</span>
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item>
        <el-button type="primary" @click="handleSubmit" :loading="submitting" :disabled="getButtonState('config.save').disabled">
          {{ $t('common.save') }}
        </el-button>
        <el-button type="success" @click="handleTest" :loading="testing" :disabled="getButtonState('config.test_email').disabled">
          {{ $t('config.test_email') }}
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { forOwn } from 'lodash-es'
import { getConfigByGroup, saveConfig, testEmail } from '../../../api/config'
import { usePermission } from '../../../composables/usePermission'

const { t } = useI18n()
const { getButtonState } = usePermission()
const formRef = ref(null)
const submitting = ref(false)
const testing = ref(false)

const formData = reactive({
  email_host: '',
  email_port: 587,
  email_username: '',
  email_password: '',
  email_from: '',
  email_from_name: '',
  email_encryption: 'tls',
  email_timeout: 30
})

const formRules = {
  email_host: [
    { required: true, message: t('config.email_host_required'), trigger: 'blur' }
  ],
  email_port: [
    { required: true, message: t('config.email_port_required'), trigger: 'blur' }
  ],
  email_username: [
    { required: true, message: t('config.email_username_required'), trigger: 'blur' }
  ],
  email_from: [
    { required: true, message: t('config.email_from_required'), trigger: 'blur' },
    { type: 'email', message: t('config.email_from_invalid'), trigger: 'blur' }
  ]
}

const loadData = async () => {
  try {
    const res = await getConfigByGroup('email')
    if (res.data && res.data.configs) {
      const configs = res.data.configs
      // 将配置数组转换为表单对象
      configs.forEach(config => {
        const key = config.Key || config.key
        let value = config.Value || config.value || ''
        
        // 处理数字类型
        if (key === 'email_port' || key === 'email_timeout') {
          value = value ? parseInt(value) : (key === 'email_port' ? 587 : 30)
        }
        
        if (formData.hasOwnProperty(key)) {
          formData[key] = value
        }
      })
    }
  } catch (error) {
    console.error('Load email config error:', error)
    if (!error.__handled) {
      const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
      ElMessage.error(errorMessage)
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        // 将表单数据转换为配置对象
        // 如果密码为空，则不发送该字段，让后端保持原有值
        const configs = {}
        forOwn(formData, (value, key) => {
          // 如果是密码字段且为空，则跳过
          if (key === 'email_password' && !value) {
            return
          }
          configs[key] = value
        })

        await saveConfig('email', configs)
        ElMessage.success(t('config.update_success'))
        // 提交成功后重新加载数据，确保密码字段保持为空
        await loadData()
      } catch (error) {
        console.error('Submit error:', error)
        // 如果错误已经在响应拦截器中处理过，就不再重复显示
        if (!error.__handled) {
          const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
          ElMessage.error(errorMessage)
        }
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleTest = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      testing.value = true
      try {
        await testEmail({
          email_host: formData.email_host,
          email_port: formData.email_port,
          email_username: formData.email_username,
          email_password: formData.email_password,
          email_from: formData.email_from,
          email_from_name: formData.email_from_name,
          email_encryption: formData.email_encryption,
          email_timeout: formData.email_timeout
        })
        ElMessage.success(t('config.test_email_success'))
      } catch (error) {
        console.error('Test email error:', error)
        // 如果错误已经在响应拦截器中处理过，就不再重复显示
        if (!error.__handled) {
          const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
          ElMessage.error(errorMessage)
        }
      } finally {
        testing.value = false
      }
    }
  })
}

onMounted(() => {
  loadData()
  
  // 立即应用
  // applyInputStyles()
  
  // 监听主题变化
  const observer = new MutationObserver(() => {
    // applyInputStyles()
  })
  
  observer.observe(document.body, {
    attributes: true,
    attributeFilter: ['class']
  })
  
  // 组件卸载时清理
  onUnmounted(() => {
    observer.disconnect()
  })
})

// 暴露方法供父组件调用
defineExpose({
  loadData
})
</script>

<style scoped>
.email-config {
  padding: 20px 0;
}
</style>

