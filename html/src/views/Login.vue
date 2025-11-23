<template>
  <div class="login-container">
    <div class="language-switch-wrapper">
      <LanguageSwitch />
    </div>
    <div class="login-box">
      <div class="login-header">
        <h2>{{ $t('login.title') }}</h2>
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
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { login } from '../api/auth'
import { useUserStore } from '../store/user'
import LanguageSwitch from '../components/LanguageSwitch.vue'

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

const loginFormRef = ref(null)
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules = computed(() => ({
  username: [
    { required: true, message: t('login.username_required'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: t('login.password_required'), trigger: 'blur' }
  ]
}))

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const res = await login(loginForm)
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
  text-align: center;
  margin-bottom: 30px;
}

.login-header h2 {
  color: #333;
  font-size: 24px;
}

.login-form {
  margin-top: 20px;
}

.login-button {
  width: 100%;
}

.language-switch-wrapper {
  position: absolute;
  top: 20px;
  right: 20px;
}
</style>

