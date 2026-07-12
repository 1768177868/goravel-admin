<template>
  <el-dialog
    :model-value="lockDialogVisible"
    :title="$t('header.lock_screen')"
    width="420px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    append-to-body
    @update:model-value="$emit('update:lockDialogVisible', $event)"
  >
    <el-input
      :model-value="pendingLockPassword"
      type="password"
      show-password
      :name="lockDialogInputName"
      autocomplete="new-password"
      autocorrect="off"
      spellcheck="false"
      :placeholder="$t('header.lock_password_placeholder')"
      @update:model-value="$emit('update:pendingLockPassword', $event)"
      @keyup.enter="$emit('confirm-lock')"
    />
    <div v-if="lockDialogError" class="lock-screen-error">{{ lockDialogError }}</div>
    <template #footer>
      <el-button @click="$emit('update:lockDialogVisible', false)">{{ $t('common.cancel') }}</el-button>
      <el-button type="primary" @click="$emit('confirm-lock')">{{ $t('common.confirm') }}</el-button>
    </template>
  </el-dialog>

  <div v-if="isScreenLocked" class="lock-screen-overlay">
    <div class="lock-screen-card">
      <div class="lock-screen-avatar-wrap">
        <el-avatar
          v-if="userStore.adminInfo?.avatar"
          :size="68"
          :src="userStore.adminInfo.avatar"
        />
        <el-avatar v-else :size="68">
          <el-icon><User /></el-icon>
        </el-avatar>
      </div>
      <div class="lock-screen-title">{{ $t('header.lock_screen_title') }}</div>
      <div class="lock-screen-user">{{ userStore.adminInfo?.nickname || userStore.adminInfo?.username }}</div>
      <el-input
        :model-value="unlockPassword"
        type="password"
        show-password
        class="lock-screen-input"
        autocomplete="new-password"
        :name="lockInputName"
        autocorrect="off"
        spellcheck="false"
        :placeholder="$t('header.lock_password_placeholder')"
        @update:model-value="$emit('update:unlockPassword', $event)"
        @input="$emit('unlock-input')"
        @keyup.enter="$emit('unlock')"
      />
      <div v-if="unlockError" class="lock-screen-error">{{ unlockError }}</div>
      <div class="lock-screen-actions">
        <el-button type="primary" @click="$emit('unlock')">{{ $t('header.unlock') }}</el-button>
        <el-button @click="$emit('go-login')">{{ $t('header.back_to_login') }}</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { User } from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'

defineProps({
  isScreenLocked: { type: Boolean, default: false },
  lockDialogVisible: { type: Boolean, default: false },
  pendingLockPassword: { type: String, default: '' },
  lockDialogError: { type: String, default: '' },
  unlockPassword: { type: String, default: '' },
  unlockError: { type: String, default: '' },
  lockInputName: { type: String, default: '' },
  lockDialogInputName: { type: String, default: '' }
})

defineEmits([
  'update:lockDialogVisible',
  'update:pendingLockPassword',
  'update:unlockPassword',
  'confirm-lock',
  'unlock-input',
  'unlock',
  'go-login'
])

const userStore = useUserStore()
</script>
