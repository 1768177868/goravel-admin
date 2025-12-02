<template>
  <div class="login-container">
    <div class="login-background">
      <div class="bg-shape bg-shape-1"></div>
      <div class="bg-shape bg-shape-2"></div>
      <div class="bg-shape bg-shape-3"></div>
    </div>
    <div class="login-box">
      <div class="login-header">
        <div class="login-logo">
          <div class="logo-icon">
            <el-icon :size="32"><Lock /></el-icon>
          </div>
          <h2>{{ $t('login.title') }}</h2>
        </div>
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
            class="login-input"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            :placeholder="$t('login.password')"
            size="large"
            prefix-icon="Lock"
            class="login-input"
            show-password
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
            class="login-input"
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
import { Lock } from '@element-plus/icons-vue'
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
        if (res.data && res.data.token) {
          const token = res.data.token
          // 登录时清除旧的数据，确保获取最新的数据
          userStore.menus = []
          userStore.adminInfo = null
          userStore.permissions = []
          localStorage.removeItem('adminInfo')
          
          userStore.setToken(token)
          // 注意：登录接口返回的 admin 信息可能不完整，所以先不设置
          // 等待 fetchUserInfo() 获取完整的管理员信息（包括权限和菜单）
          // 等待一下确保token已保存
          await new Promise(resolve => setTimeout(resolve, 100))
          await userStore.fetchUserInfo()
          ElMessage.success(t('login.login_success'))
          router.push('/')
        } else {
          throw new Error(t('login.login_failed'))
        }
      } catch (error) {
        if (error?.__handled) {
          // 已在 axios 拦截器中提示
        } else if (error.response) {
          const errorMessage = error.response.data?.message || ''
          // 检查是否是账号被禁用
          if (errorMessage === 'account_disabled') {
            ElMessage.error(t('login.account_disabled'))
          } else {
            // 尝试翻译错误消息，如果翻译不存在则使用原始消息
            const translatedMessage = t(errorMessage) !== errorMessage ? t(errorMessage) : errorMessage
            ElMessage.error(translatedMessage || t('login.login_failed'))
          }
        } else if (error.message) {
          ElMessage.error(error.message)
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
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  overflow: hidden;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%);
  background-size: 400% 400%;
  animation: gradientShift 15s ease infinite;
}

@keyframes gradientShift {
  0% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0% 50%;
  }
}

.login-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  z-index: 0;
}

.bg-shape {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  animation: float 20s infinite ease-in-out;
}

.bg-shape-1 {
  width: 300px;
  height: 300px;
  top: -100px;
  left: -100px;
  animation-delay: 0s;
}

.bg-shape-2 {
  width: 200px;
  height: 200px;
  bottom: -50px;
  right: -50px;
  animation-delay: 5s;
}

.bg-shape-3 {
  width: 150px;
  height: 150px;
  top: 50%;
  right: 10%;
  animation-delay: 10s;
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) rotate(0deg);
  }
  33% {
    transform: translate(30px, -30px) rotate(120deg);
  }
  66% {
    transform: translate(-20px, 20px) rotate(240deg);
  }
}

.login-box {
  position: relative;
  z-index: 1;
  width: 420px;
  padding: 48px 40px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3),
              0 0 0 1px rgba(255, 255, 255, 0.2) inset;
  animation: slideUp 0.6s ease-out;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.login-box:hover {
  transform: translateY(-5px);
  box-shadow: 0 25px 70px rgba(0, 0, 0, 0.35),
              0 0 0 1px rgba(255, 255, 255, 0.3) inset;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.login-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 36px;
}

.login-logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: white;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
  }
  50% {
    transform: scale(1.05);
    box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
  }
}

.login-header h2 {
  color: #2c3e50;
  font-size: 28px;
  font-weight: 600;
  margin: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.login-language-switch :deep(.language-switch) {
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  background: rgba(255, 255, 255, 0.8);
  transition: all 0.3s ease;
}

.login-language-switch :deep(.language-switch):hover {
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.1);
}

.login-form {
  margin-top: 8px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 24px;
}

.login-input :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  background: rgba(255, 255, 255, 0.9);
}

.login-input :deep(.el-input__wrapper):hover {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
}

.login-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.25);
  border-color: #667eea;
}

.login-input :deep(.el-input__inner) {
  font-size: 15px;
  padding: 0 12px;
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
  transition: all 0.3s ease;
  margin-top: 8px;
}

.login-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.5);
}

.login-button:active {
  transform: translateY(0);
  box-shadow: 0 2px 10px rgba(102, 126, 234, 0.3);
}

.captcha-row {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}

.captcha-image {
  height: 48px;
  width: 170px;
  object-fit: cover;
  cursor: pointer;
  border-radius: 8px;
  border: 1px solid #ebeef5;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.captcha-image:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.captcha-refresh {
  white-space: nowrap;
  padding: 0 8px;
  transition: all 0.3s ease;
}

.captcha-refresh:hover {
  transform: rotate(180deg);
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-box {
    width: 90%;
    padding: 36px 28px;
    margin: 20px;
  }

  .login-header h2 {
    font-size: 24px;
  }

  .logo-icon {
    width: 40px;
    height: 40px;
  }
}
</style>

