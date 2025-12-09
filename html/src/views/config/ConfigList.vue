<template>
  <div class="config-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('config.title') }}</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane :label="$t('config.website_config')" name="website">
          <WebsiteConfig ref="websiteConfigRef" />
        </el-tab-pane>
        <el-tab-pane :label="$t('config.email_config')" name="email">
          <EmailConfig ref="emailConfigRef" />
        </el-tab-pane>
        <el-tab-pane :label="$t('config.captcha_config')" name="captcha">
          <CaptchaConfig ref="captchaConfigRef" />
        </el-tab-pane>
        <el-tab-pane :label="$t('config.storage_config')" name="storage">
          <StorageConfig ref="storageConfigRef" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import WebsiteConfig from './components/WebsiteConfig.vue'
import EmailConfig from './components/EmailConfig.vue'
import CaptchaConfig from './components/CaptchaConfig.vue'
import StorageConfig from './components/StorageConfig.vue'

const activeTab = ref('website')
const websiteConfigRef = ref(null)
const emailConfigRef = ref(null)
const captchaConfigRef = ref(null)
const storageConfigRef = ref(null)

const handleTabChange = (tabName) => {
  // 切换tab时可以重新加载数据
  if (tabName === 'website' && websiteConfigRef.value) {
    websiteConfigRef.value.loadData()
  } else if (tabName === 'email' && emailConfigRef.value) {
    emailConfigRef.value.loadData()
  } else if (tabName === 'captcha' && captchaConfigRef.value) {
    captchaConfigRef.value.loadData()
  } else if (tabName === 'storage' && storageConfigRef.value) {
    storageConfigRef.value.loadData()
  }
}
</script>

<style scoped>
.config-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

<style>
/* 配置管理页面夜间模式适配 */
.dark-mode .config-list {
  background: var(--card-bg) !important;
}

/* 确保配置管理页面内的所有输入框都有正确的背景色 */
.dark-mode .config-list :deep(.el-input__wrapper) {
  background-color: #252526 !important;
  border-color: #3d3e40 !important;
}

.dark-mode .config-list :deep(.el-input__wrapper .el-input__inner) {
  background-color: #252526 !important;
  color: #e5eaf3 !important;
}

.dark-mode .config-list :deep(.el-input[type="password"] .el-input__wrapper),
.dark-mode .config-list :deep(.el-input[type="password"] .el-input__wrapper .el-input__inner) {
  background-color: #252526 !important;
  color: #e5eaf3 !important;
}

.dark-mode .config-list :deep(.el-input.show-password .el-input__wrapper),
.dark-mode .config-list :deep(.el-input.show-password .el-input__wrapper .el-input__inner) {
  background-color: #252526 !important;
  color: #e5eaf3 !important;
}

.dark-mode .config-list :deep(.el-input.show-password .el-input__suffix) {
  background-color: transparent !important;
}

.dark-mode .config-list :deep(.el-input.show-password .el-input__suffix .el-input__icon) {
  background-color: transparent !important;
  color: #909399 !important;
}

.dark-mode .config-list :deep(.el-input-number .el-input__wrapper) {
  background-color: #252526 !important;
  border-color: #3d3e40 !important;
}

.dark-mode .config-list :deep(.el-input-number .el-input__wrapper .el-input__inner) {
  background-color: #252526 !important;
  color: #e5eaf3 !important;
}

.dark-mode .config-list :deep(.el-select .el-input__wrapper) {
  background-color: #252526 !important;
  border-color: #3d3e40 !important;
}

.dark-mode .config-list :deep(.el-select .el-input__wrapper .el-input__inner) {
  background-color: #252526 !important;
  color: #e5eaf3 !important;
}

.dark-mode .config-list :deep(.el-input__wrapper::before),
.dark-mode .config-list :deep(.el-input__wrapper::after) {
  background-color: transparent !important;
}

.dark-mode .config-list :deep(.el-input__wrapper *) {
  background-color: transparent !important;
}

.dark-mode .config-list :deep(.el-input__wrapper .el-input__inner) {
  background-color: #252526 !important;
}
</style>

