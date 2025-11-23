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
              <el-button type="primary" link @click="handleOpenAvatarDialog">
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
              <template v-if="adminInfo.roles && adminInfo.roles.length > 0">
                <el-tag
                  v-for="role in adminInfo.roles"
                  :key="role.id"
                  style="margin-right: 5px;"
                >
                  {{ role.name }}
                </el-tag>
              </template>
              <span v-else>-</span>
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

    <!-- 头像选择对话框 -->
    <el-dialog
      v-model="showAvatarDialog"
      :title="$t('profile.change_avatar')"
      width="500px"
    >
      <div class="avatar-selector">
        <div class="avatar-grid">
          <div
            v-for="avatar in defaultAvatars"
            :key="avatar"
            class="avatar-item"
            :class="{ active: selectedAvatar === avatar }"
            @click="selectedAvatar = avatar"
          >
            <el-avatar :size="60" :src="avatar">
              <el-icon><User /></el-icon>
            </el-avatar>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showAvatarDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSaveAvatar" :loading="avatarSubmitting" :disabled="!selectedAvatar">
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
const selectedAvatar = ref('')
const avatarSubmitting = ref(false)

// 系统默认头像列表（使用UI Avatars服务生成）
const defaultAvatars = [
  'https://ui-avatars.com/api/?name=Admin&background=409EFF&color=fff&size=128',
  'https://ui-avatars.com/api/?name=User&background=67C23A&color=fff&size=128',
  'https://ui-avatars.com/api/?name=Manager&background=E6A23C&color=fff&size=128',
  'https://ui-avatars.com/api/?name=Admin&background=F56C6C&color=fff&size=128',
  'https://ui-avatars.com/api/?name=User&background=909399&color=fff&size=128',
  'https://ui-avatars.com/api/?name=Admin&background=409EFF&color=fff&size=128&bold=true',
  'https://ui-avatars.com/api/?name=User&background=67C23A&color=fff&size=128&bold=true',
  'https://ui-avatars.com/api/?name=Manager&background=E6A23C&color=fff&size=128&bold=true',
  'https://ui-avatars.com/api/?name=Admin&background=F56C6C&color=fff&size=128&bold=true',
  'https://ui-avatars.com/api/?name=User&background=909399&color=fff&size=128&bold=true',
  'https://ui-avatars.com/api/?name=Admin&background=409EFF&color=fff&size=128&rounded=true',
  'https://ui-avatars.com/api/?name=User&background=67C23A&color=fff&size=128&rounded=true',
  'https://ui-avatars.com/api/?name=Manager&background=E6A23C&color=fff&size=128&rounded=true',
  'https://ui-avatars.com/api/?name=Admin&background=F56C6C&color=fff&size=128&rounded=true',
  'https://ui-avatars.com/api/?name=User&background=909399&color=fff&size=128&rounded=true',
  'https://ui-avatars.com/api/?name=Admin&background=409EFF&color=fff&size=128&bold=true&rounded=true',
  'https://ui-avatars.com/api/?name=User&background=67C23A&color=fff&size=128&bold=true&rounded=true',
  'https://ui-avatars.com/api/?name=Manager&background=E6A23C&color=fff&size=128&bold=true&rounded=true'
]

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

const validateEmail = (rule, value, callback) => {
  if (value && value.trim() !== '') {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(value)) {
      callback(new Error(t('profile.email_invalid')))
    } else {
      callback()
    }
  } else {
    callback()
  }
}

const validatePhone = (rule, value, callback) => {
  if (value && value.trim() !== '') {
    const phoneRegex = /^1[3-9]\d{9}$/
    if (!phoneRegex.test(value)) {
      callback(new Error(t('profile.phone_invalid')))
    } else {
      callback()
    }
  } else {
    callback()
  }
}

const infoRules = {
  email: [
    { validator: validateEmail, trigger: 'blur' }
  ],
  phone: [
    { validator: validatePhone, trigger: 'blur' }
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
    console.log('Profile response:', res)
    if (res.data && res.data.admin) {
      const admin = res.data.admin
      console.log('Admin data:', admin)
      console.log('Department:', admin.Department || admin.department)
      console.log('Roles:', admin.Roles || admin.roles)
      
      // 处理部门数据
      let department = null
      if (admin.Department && (admin.Department.ID || admin.Department.id)) {
        department = {
          id: admin.Department.ID || admin.Department.id,
          name: admin.Department.Name || admin.Department.name || '-'
        }
      } else if (admin.department && admin.department.id) {
        department = {
          id: admin.department.id,
          name: admin.department.name || '-'
        }
      }
      
      // 处理角色数据（去重）
      const rolesArray = admin.Roles || admin.roles || []
      const roleMap = new Map()
      rolesArray.forEach(role => {
        const roleId = role.ID || role.id
        if (roleId && !roleMap.has(roleId)) {
          roleMap.set(roleId, {
            id: roleId,
            name: role.Name || role.name,
            slug: role.Slug || role.slug
          })
        }
      })
      const uniqueRoles = Array.from(roleMap.values())
      
      // 转换数据格式（PascalCase -> snake_case）
      const transformedAdmin = {
        id: admin.ID || admin.id,
        username: admin.Username || admin.username,
        nickname: admin.Nickname || admin.nickname,
        email: admin.Email || admin.email,
        phone: admin.Phone || admin.phone,
        avatar: admin.Avatar || admin.avatar,
        department_id: admin.DepartmentID || admin.department_id,
        department: department,
        roles: uniqueRoles
      }
      
      console.log('Transformed admin:', transformedAdmin)
      
      infoForm.nickname = transformedAdmin.nickname || ''
      infoForm.email = transformedAdmin.email || ''
      infoForm.phone = transformedAdmin.phone || ''
      userStore.setAdminInfo(transformedAdmin)
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
          const admin = res.data.admin
          
          // 处理部门数据
          let department = null
          if (admin.Department && (admin.Department.ID || admin.Department.id)) {
            department = {
              id: admin.Department.ID || admin.Department.id,
              name: admin.Department.Name || admin.Department.name || '-'
            }
          } else if (admin.department && admin.department.id) {
            department = {
              id: admin.department.id,
              name: admin.department.name || '-'
            }
          }
          
          // 处理角色数据（去重）
          const rolesArray = admin.Roles || admin.roles || []
          const roleMap = new Map()
          rolesArray.forEach(role => {
            const roleId = role.ID || role.id
            if (roleId && !roleMap.has(roleId)) {
              roleMap.set(roleId, {
                id: roleId,
                name: role.Name || role.name,
                slug: role.Slug || role.slug
              })
            }
          })
          const uniqueRoles = Array.from(roleMap.values())
          
          // 转换数据格式（PascalCase -> snake_case）
          const transformedAdmin = {
            id: admin.ID || admin.id,
            username: admin.Username || admin.username,
            nickname: admin.Nickname || admin.nickname,
            email: admin.Email || admin.email,
            phone: admin.Phone || admin.phone,
            avatar: admin.Avatar || admin.avatar,
            department_id: admin.DepartmentID || admin.department_id,
            department: department,
            roles: uniqueRoles
          }
          
          userStore.setAdminInfo(transformedAdmin)
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

const handleOpenAvatarDialog = () => {
  // 打开对话框时，如果已有头像，设置为选中状态
  selectedAvatar.value = adminInfo.value.avatar || ''
  showAvatarDialog.value = true
}

const handleSaveAvatar = async () => {
  if (!selectedAvatar.value) {
    ElMessage.warning(t('profile.please_select_avatar'))
    return
  }

  avatarSubmitting.value = true
  try {
    await updateProfile({ avatar: selectedAvatar.value })
    await loadProfile()
    ElMessage.success(t('profile.avatar_update_success'))
    showAvatarDialog.value = false
    selectedAvatar.value = ''
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

.avatar-selector {
  padding: 20px 0;
}

.avatar-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 15px;
  max-height: 400px;
  overflow-y: auto;
}

.avatar-item {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 8px;
  border: 2px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.avatar-item:hover {
  border-color: #409EFF;
  transform: scale(1.05);
}

.avatar-item.active {
  border-color: #409EFF;
  background-color: #ecf5ff;
}
</style>

