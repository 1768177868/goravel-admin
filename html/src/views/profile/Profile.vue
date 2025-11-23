<template>
  <div class="profile-container">
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card class="profile-card">
          <template #header>
            <div class="card-header">
              <span>{{ $t('profile.basic_info') }}</span>
            </div>
          </template>
          <div class="avatar-section">
            <el-avatar :size="100" :src="adminInfo.avatar" class="avatar">
              <el-icon><User /></el-icon>
            </el-avatar>
            <div class="avatar-actions">
              <el-button type="primary" link @click="showAvatarDialog = true">
                {{ $t('profile.change_avatar') }}
              </el-button>
            </div>
          </div>
          <el-descriptions :column="1" border class="info-descriptions">
            <el-descriptions-item :label="$t('profile.username')">
              {{ adminInfo.username }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('profile.nickname')">
              {{ adminInfo.nickname || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('profile.email')">
              {{ adminInfo.email || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('profile.phone')">
              {{ adminInfo.phone || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('profile.department')">
              {{ adminInfo.department?.name || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('profile.roles')">
              <el-tag
                v-for="role in adminInfo.roles"
                :key="role.id"
                style="margin-right: 5px;"
              >
                {{ role.name }}
              </el-tag>
              <span v-if="!adminInfo.roles || adminInfo.roles.length === 0">-</span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <el-col :span="16">
        <el-card>
          <el-tabs v-model="activeTab">
            <el-tab-pane :label="$t('profile.edit_info')" name="info">
              <el-form
                ref="infoFormRef"
                :model="infoForm"
                :rules="infoRules"
                label-width="120px"
                style="max-width: 600px;"
              >
                <el-form-item :label="$t('profile.nickname')" prop="nickname">
                  <el-input v-model="infoForm.nickname" />
                </el-form-item>
                <el-form-item :label="$t('profile.email')" prop="email">
                  <el-input v-model="infoForm.email" type="email" />
                </el-form-item>
                <el-form-item :label="$t('profile.phone')" prop="phone">
                  <el-input v-model="infoForm.phone" />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="handleUpdateInfo" :loading="infoSubmitting">
                    {{ $t('common.save') }}
                  </el-button>
                  <el-button @click="handleResetInfo">{{ $t('common.reset') }}</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <el-tab-pane :label="$t('profile.change_password')" name="password">
              <el-form
                ref="passwordFormRef"
                :model="passwordForm"
                :rules="passwordRules"
                label-width="120px"
                style="max-width: 600px;"
              >
                <el-form-item :label="$t('profile.old_password')" prop="old_password">
                  <el-input
                    v-model="passwordForm.old_password"
                    type="password"
                    show-password
                  />
                </el-form-item>
                <el-form-item :label="$t('profile.new_password')" prop="new_password">
                  <el-input
                    v-model="passwordForm.new_password"
                    type="password"
                    show-password
                  />
                </el-form-item>
                <el-form-item :label="$t('profile.confirm_password')" prop="confirm_password">
                  <el-input
                    v-model="passwordForm.confirm_password"
                    type="password"
                    show-password
                  />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="handleUpdatePassword" :loading="passwordSubmitting">
                    {{ $t('common.save') }}
                  </el-button>
                  <el-button @click="handleResetPassword">{{ $t('common.reset') }}</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>
    </el-row>

    <!-- 头像上传对话框 -->
    <el-dialog
      v-model="showAvatarDialog"
      :title="$t('profile.change_avatar')"
      width="400px"
    >
      <el-upload
        class="avatar-uploader"
        action="#"
        :show-file-list="false"
        :before-upload="beforeAvatarUpload"
        :http-request="handleAvatarUpload"
      >
        <img v-if="avatarUrl" :src="avatarUrl" class="avatar-preview" />
        <el-icon v-else class="avatar-uploader-icon"><Plus /></el-icon>
      </el-upload>
      <template #footer>
        <el-button @click="showAvatarDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSaveAvatar" :loading="avatarSubmitting">
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { User, Plus } from '@element-plus/icons-vue'
import { getProfile, updateProfile, updatePassword } from '../../api/profile'
import { useUserStore } from '../../store/user'

const { t } = useI18n()
const userStore = useUserStore()

const activeTab = ref('info')
const infoFormRef = ref(null)
const passwordFormRef = ref(null)
const infoSubmitting = ref(false)
const passwordSubmitting = ref(false)
const showAvatarDialog = ref(false)
const avatarUrl = ref('')
const avatarSubmitting = ref(false)
const avatarFile = ref(null)

const adminInfo = computed(() => userStore.adminInfo || {})

const infoForm = reactive({
  nickname: '',
  email: '',
  phone: ''
})

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const validateConfirmPassword = (rule, value, callback) => {
  if (value !== passwordForm.new_password) {
    callback(new Error(t('profile.password_not_match')))
  } else {
    callback()
  }
}

const infoRules = {
  email: [
    { type: 'email', message: t('profile.email_invalid'), trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: t('profile.phone_invalid'), trigger: 'blur' }
  ]
}

const passwordRules = {
  old_password: [
    { required: true, message: t('profile.old_password_required'), trigger: 'blur' }
  ],
  new_password: [
    { required: true, message: t('profile.new_password_required'), trigger: 'blur' },
    { min: 6, message: t('profile.password_length_error'), trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: t('profile.confirm_password_required'), trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const loadProfile = async () => {
  try {
    const res = await getProfile()
    if (res.data && res.data.admin) {
      const admin = res.data.admin
      infoForm.nickname = admin.nickname || ''
      infoForm.email = admin.email || ''
      infoForm.phone = admin.phone || ''
      userStore.setAdminInfo(admin)
    }
  } catch (error) {
    console.error('Load profile error:', error)
  }
}

const handleUpdateInfo = async () => {
  if (!infoFormRef.value) return

  await infoFormRef.value.validate(async (valid) => {
    if (valid) {
      infoSubmitting.value = true
      try {
        const res = await updateProfile(infoForm)
        if (res.data && res.data.admin) {
          userStore.setAdminInfo(res.data.admin)
          ElMessage.success(t('profile.update_success'))
        }
      } catch (error) {
        console.error('Update info error:', error)
      } finally {
        infoSubmitting.value = false
      }
    }
  })
}

const handleResetInfo = () => {
  loadProfile()
  infoFormRef.value?.resetFields()
}

const handleUpdatePassword = async () => {
  if (!passwordFormRef.value) return

  await passwordFormRef.value.validate(async (valid) => {
    if (valid) {
      passwordSubmitting.value = true
      try {
        await updatePassword(passwordForm)
        ElMessage.success(t('profile.password_update_success'))
        handleResetPassword()
      } catch (error) {
        console.error('Update password error:', error)
      } finally {
        passwordSubmitting.value = false
      }
    }
  })
}

const handleResetPassword = () => {
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
  passwordFormRef.value?.resetFields()
}

const beforeAvatarUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error(t('profile.avatar_format_error'))
    return false
  }
  if (!isLt2M) {
    ElMessage.error(t('profile.avatar_size_error'))
    return false
  }

  // 预览图片
  const reader = new FileReader()
  reader.onload = (e) => {
    avatarUrl.value = e.target.result
  }
  reader.readAsDataURL(file)
  avatarFile.value = file

  return false // 阻止自动上传
}

const handleAvatarUpload = () => {
  // 这里可以实现实际上传逻辑
  // 暂时只保存到本地预览
}

const handleSaveAvatar = async () => {
  if (!avatarFile.value) {
    ElMessage.warning(t('profile.please_select_avatar'))
    return
  }

  avatarSubmitting.value = true
  try {
    // 这里应该实际上传头像到服务器
    // 暂时使用 base64 或 URL
    // 假设上传后返回 URL
    const avatar = avatarUrl.value // 实际应该是服务器返回的 URL

    await updateProfile({ avatar })
    await loadProfile()
    ElMessage.success(t('profile.avatar_update_success'))
    showAvatarDialog.value = false
    avatarUrl.value = ''
    avatarFile.value = null
  } catch (error) {
    console.error('Update avatar error:', error)
  } finally {
    avatarSubmitting.value = false
  }
}

onMounted(() => {
  loadProfile()
})
</script>

<style scoped>
.profile-container {
  padding: 0;
}

.profile-card {
  height: 100%;
}

.card-header {
  font-weight: 500;
  font-size: 16px;
}

.avatar-section {
  text-align: center;
  margin-bottom: 20px;
}

.avatar {
  margin-bottom: 10px;
}

.avatar-actions {
  margin-top: 10px;
}

.info-descriptions {
  margin-top: 20px;
}

.avatar-uploader {
  display: flex;
  justify-content: center;
}

.avatar-uploader :deep(.el-upload) {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
}

.avatar-uploader :deep(.el-upload:hover) {
  border-color: #409EFF;
}

.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 178px;
  height: 178px;
  line-height: 178px;
  text-align: center;
}

.avatar-preview {
  width: 178px;
  height: 178px;
  display: block;
  object-fit: cover;
}
</style>

