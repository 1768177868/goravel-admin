<template>
  <div class="login-container">
    <div class="login-background">
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
        <el-form-item v-if="needGoogleCode" prop="google_code">
          <el-input
            v-model="loginForm.google_code"
            :placeholder="$t('login.google_code_placeholder')"
            size="large"
            class="login-input"
            maxlength="6"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item v-if="captchaInfo.shouldShow && !needGoogleCode" prop="captcha_answer">
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
              <el-icon class="refresh-icon"><Refresh /></el-icon>
              <span>{{ $t('login.refresh_captcha') }}</span>
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
import { Lock, Refresh } from '@element-plus/icons-vue'
import { login, getLoginCaptcha } from '../api/auth'
import { useUserStore } from '../store/user'
import LanguageSwitch from '../components/LanguageSwitch.vue'
import { ERROR_CODES } from '../utils/request'
import Storage from '../utils/storage'

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

const loginFormRef = ref(null)
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: '',
  captcha_answer: '',
  google_code: ''
})

const captchaInfo = reactive({
  enabled: false,
  captcha_id: '',
  image: '',
  shouldShow: false // 是否应该显示图形验证码（需要先验证账号密码后才能确定）
})

const needGoogleCode = ref(false)

const loginRules = computed(() => ({
  username: [
    { required: true, message: t('login.username_required'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: t('login.password_required'), trigger: 'blur' }
  ],
  google_code: needGoogleCode.value
    ? [
        { required: true, message: t('login.google_code_required'), trigger: 'blur' },
        { pattern: /^\d{6}$/, message: t('login.google_code_format'), trigger: 'blur' }
      ]
    : [],
  captcha_answer: captchaInfo.shouldShow && !needGoogleCode.value
    ? [{ required: true, message: t('login.captcha_required'), trigger: 'blur' }]
    : []
}))

// 获取图形验证码配置（不自动获取图片）
const checkCaptchaEnabled = async () => {
  try {
    const res = await getLoginCaptcha()
    const captcha = res.data?.captcha || {}
    captchaInfo.enabled = !!captcha.enabled
    // 不自动显示图形验证码，需要先验证账号密码
    captchaInfo.shouldShow = false
  } catch (error) {
    console.error('Check captcha enabled error:', error)
    captchaInfo.enabled = false
    captchaInfo.shouldShow = false
  }
}

// 获取图形验证码图片（当需要显示时才获取）
const fetchCaptcha = async () => {
  try {
    const res = await getLoginCaptcha()
    const captcha = res.data?.captcha || {}
    captchaInfo.enabled = !!captcha.enabled
    captchaInfo.captcha_id = captcha.captcha_id || ''
    captchaInfo.image = captcha.captcha_image || ''
    captchaInfo.shouldShow = true
  } catch (error) {
    console.error('Fetch captcha error:', error)
    captchaInfo.enabled = false
    captchaInfo.captcha_id = ''
    captchaInfo.image = ''
    captchaInfo.shouldShow = false
  } finally {
    loginForm.captcha_answer = ''
    if (loginFormRef.value) {
      loginFormRef.value.clearValidate(['captcha_answer'])
    }
  }
}

onMounted(() => {
  // 只检查图形验证码是否启用，不自动获取图片
  checkCaptchaEnabled()
})

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  // 先验证账号密码（不包含图形验证码）
  // 如果绑定了 2FA，后端会返回 google_code_required
  // 如果没有绑定 2FA 且图形验证码开启，后端会返回需要图形验证码的错误
  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const payload = {
          username: loginForm.username,
          password: loginForm.password
        }
        
        // 如果已经需要谷歌验证码，添加谷歌验证码
        if (needGoogleCode.value) {
          payload.google_code = loginForm.google_code
        }
        // 如果图形验证码应该显示，添加图形验证码
        else if (captchaInfo.shouldShow) {
          payload.captcha_id = captchaInfo.captcha_id
          payload.captcha_answer = loginForm.captcha_answer
        }
        // 否则，先只提交账号密码，让后端判断是否需要图形验证码或谷歌验证码
        
        const res = await login(payload)
        if (res.data && res.data.token) {
          const token = res.data.token
          // 登录时清除旧的数据，确保获取最新的数据
          userStore.menus = []
          userStore.adminInfo = null
          userStore.permissions = []
          Storage.removeItem('adminInfo')
          
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
          return
        }
        
        // 使用错误码判断，简化逻辑
        // 优先从 error.errorCode 获取，如果没有则从 response.data.error_code 获取
        const errorCode = error.errorCode || error.response?.data?.error_code || ''
        const message = error.translatedMessage || error.message || error.response?.data?.message || ''
        
        // 根据错误码处理 UI 状态
        if (errorCode === ERROR_CODES.GOOGLE_CODE_REQUIRED) {
          // 绑定了 2FA，需要谷歌验证码，隐藏图形验证码
          needGoogleCode.value = true
          captchaInfo.shouldShow = false
          loginForm.google_code = ''
          loginForm.captcha_answer = ''
          if (loginFormRef.value) {
            loginFormRef.value.clearValidate(['captcha_answer'])
          }
          ElMessage.warning(message)
          return
        }
        
        if (errorCode === ERROR_CODES.GOOGLE_CODE_INVALID) {
          ElMessage.error(message)
          loginForm.google_code = ''
          return
        }
        
        if (errorCode === ERROR_CODES.ACCOUNT_DISABLED) {
          ElMessage.error(message)
          return
        }
        
        // 检查是否是验证码相关的错误
        if (errorCode === ERROR_CODES.CAPTCHA_INVALID || errorCode === ERROR_CODES.CAPTCHA_REQUIRED) {
          // 验证码错误，如果图形验证码开启且还没有显示，则显示图形验证码
          if (captchaInfo.enabled && !captchaInfo.shouldShow && !needGoogleCode.value) {
            await fetchCaptcha()
          }
          ElMessage.error(message)
          return
        }
        
        // 其他错误（可能是密码错误等）
        // 如果图形验证码开启且还没有显示，则显示图形验证码
        if (captchaInfo.enabled && !captchaInfo.shouldShow && !needGoogleCode.value) {
          await fetchCaptcha()
        }
        
        // 显示错误消息
        ElMessage.error(message)
      } finally {
        loading.value = false
        // 如果图形验证码已显示且不需要谷歌验证码，刷新图形验证码
        if (captchaInfo.shouldShow && !needGoogleCode.value) {
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
  background: linear-gradient(135deg, #409EFF 0%, #66b1ff 50%, #79bbff 100%);
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
  background: rgba(255, 255, 255, 0.15);
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
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(64, 158, 255, 0.12),
              0 2px 8px rgba(0, 0, 0, 0.08);
  animation: slideUp 0.6s ease-out;
  transition: all 0.3s ease;
}

.login-box:hover {
  box-shadow: 0 12px 40px rgba(64, 158, 255, 0.16),
              0 4px 12px rgba(0, 0, 0, 0.1);
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
  margin-bottom: 40px;
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
  background: #409EFF;
  border-radius: 10px;
  color: white;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
  transition: all 0.3s ease;
}

.logo-icon:hover {
  background: #66b1ff;
  box-shadow: 0 6px 16px rgba(64, 158, 255, 0.4);
  transform: translateY(-2px);
}

.login-header h2 {
  color: #303133;
  font-size: 28px;
  font-weight: 600;
  margin: 0;
  letter-spacing: -0.5px;
}

.login-language-switch :deep(.language-switch) {
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid #dcdfe6;
  background: #ffffff;
  transition: all 0.3s ease;
}

.login-language-switch :deep(.language-switch):hover {
  border-color: #409EFF;
  background: #ecf5ff;
  color: #409EFF;
}

.login-form {
  margin-top: 8px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 24px;
}

.login-form :deep(.el-form-item:last-child) {
  margin-bottom: 0;
}

.login-input :deep(.el-input__wrapper) {
  border-radius: 6px;
  box-shadow: 0 0 0 1px #dcdfe6 inset;
  transition: all 0.3s ease;
  background: #ffffff;
  padding: 0 12px;
}

.login-input :deep(.el-input__wrapper):hover {
  box-shadow: 0 0 0 1px #c0c4cc inset;
}

.login-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #409EFF inset;
}

.login-input :deep(.el-input__inner) {
  font-size: 14px;
  color: #606266;
  height: 40px;
  line-height: 40px;
}

.login-input :deep(.el-input__inner::placeholder) {
  color: #c0c4cc;
}

.login-button {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 6px;
  background: #409EFF;
  border: none;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.3);
  transition: all 0.3s ease;
  margin-top: 8px;
}

.login-button:hover {
  background: #66b1ff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.4);
}

.login-button:active {
  background: #3a8ee6;
  box-shadow: 0 2px 6px rgba(64, 158, 255, 0.3);
}

.captcha-row {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}

.captcha-image {
  height: 40px;
  width: 170px;
  object-fit: cover;
  cursor: pointer;
  border-radius: 6px;
  border: 1px solid #dcdfe6;
  box-shadow: 0 0 0 1px #dcdfe6;
  transition: all 0.3s ease;
}

.captcha-image:hover {
  border-color: #409EFF;
  box-shadow: 0 0 0 1px #409EFF;
}

.captcha-refresh {
  white-space: nowrap;
  padding: 0 8px;
  transition: all 0.3s ease;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.captcha-refresh:hover {
  color: #409EFF;
}

.captcha-refresh .refresh-icon {
  font-size: 16px;
  transition: transform 0.3s ease;
}

.captcha-refresh:hover .refresh-icon {
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

