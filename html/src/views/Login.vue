<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
      <h2>{{ $t('login.title') }}</h2>
      <LanguageSwitch class="login-language-switch" />
      </div>
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            :placeholder="$t('login.username')"
            size="large"
            prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            :placeholder="$t('login.password')"
            size="large"
            prefix-icon="Lock"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item v-if="captchaInfo.enabled" prop="captcha_answer">
          <div class="captcha-row">
            <img
              v-if="captchaInfo.image"
              :src="captchaInfo.image"
              class="captcha-image"
              :alt="$t('login.captcha_alt')"
              @click.prevent="fetchCaptcha"
            />
            <el-button
              class="captcha-refresh"
              type="primary"
              size="small"
              text
              @click.prevent="fetchCaptcha"
            >
              {{ $t('login.refresh_captcha') }}
            </el-button>
          </div>
          <el-input
            v-model="loginForm.captcha_answer"
            :placeholder="$t('login.captcha_placeholder')"
            size="large"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            class="login-button"
            :loading="loading"
            @click="handleLogin"
          >
            {{ $t('login.login') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { login, getLoginCaptcha } from '../api/auth'
import { useUserStore } from '../store/user'
import LanguageSwitch from '../components/LanguageSwitch.vue'

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

const loginFormRef = ref(null)
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: '',
  captcha_answer: ''
})

const captchaInfo = reactive({
  enabled: false,
  captcha_id: '',
  image: ''
})

const loginRules = computed(() => ({
  username: [
    { required: true, message: t('login.username_required'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: t('login.password_required'), trigger: 'blur' }
  ],
  captcha_answer: captchaInfo.enabled
    ? [{ required: true, message: t('login.captcha_required'), trigger: 'blur' }]
    : []
}))

const fetchCaptcha = async () => {
  try {
    const res = await getLoginCaptcha()
    const captcha = res.data?.captcha || {}
    captchaInfo.enabled = !!captcha.enabled
    captchaInfo.captcha_id = captcha.captcha_id || ''
    captchaInfo.image = captcha.captcha_image || ''
  } catch (error) {
    console.error('Fetch captcha error:', error)
    captchaInfo.enabled = false
    captchaInfo.captcha_id = ''
    captchaInfo.image = ''
  } finally {
    loginForm.captcha_answer = ''
    if (loginFormRef.value) {
      loginFormRef.value.clearValidate(['captcha_answer'])
    }
  }
}

onMounted(() => {
  fetchCaptcha()
})

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const payload = {
          username: loginForm.username,
          password: loginForm.password
        }
        if (captchaInfo.enabled) {
          payload.captcha_id = captchaInfo.captcha_id
          payload.captcha_answer = loginForm.captcha_answer
        }
        const res = await login(payload)
        console.log('Login response:', res)
        if (res.data && res.data.token) {
          const token = res.data.token
          console.log('Token received:', token.substring(0, 20) + '...')
          userStore.setToken(token)
          console.log('Token saved to localStorage:', localStorage.getItem('token')?.substring(0, 20) + '...')
          if (res.data.admin) {
            userStore.setAdminInfo(res.data.admin)
          }
          // 等待一下确保token已保存
          await new Promise(resolve => setTimeout(resolve, 100))
          await userStore.fetchUserInfo()
          ElMessage.success(t('login.login_success'))
          router.push('/')
        } else {
          console.error('No token in login response:', res)
          ElMessage.error(t('login.login_failed'))
        }
      } catch (error) {
        console.error('Login error:', error)
        if (error.response) {
          const message = error.response.data?.message || error.message
          ElMessage.error(message || t('login.login_failed'))
        } else {
          ElMessage.error(t('login.login_failed'))
        }
      } finally {
        loading.value = false
        if (captchaInfo.enabled) {
          await fetchCaptcha()
        }
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-box {
  width: 400px;
  padding: 40px;
  background: white;
  border-radius: 10px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
}

.login-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 30px;
}

.login-header h2 {
  color: #333;
  font-size: 24px;
  margin: 0;
}

.login-language-switch :deep(.language-switch) {
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
}

.login-form {
  margin-top: 20px;
}

.login-button {
  width: 100%;
}

.captcha-row {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
  margin-bottom: 8px;
}

.captcha-image {
  height: 48px;
  width: 150px;
  object-fit: cover;
  cursor: pointer;
  border-radius: 4px;
  border: 1px solid #ebeef5;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
}

.captcha-refresh {
  white-space: nowrap;
  padding: 0 8px;
}
</style>

