<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <div class="not-found-icon">
        <el-icon :size="120" color="#909399">
          <DocumentDelete />
        </el-icon>
      </div>
      <h1 class="not-found-title">404</h1>
      <p class="not-found-message">{{ $t('notFound.message') || '页面不存在' }}</p>
      <div class="not-found-actions">
        <el-button type="primary" @click="goHome">
          {{ $t('notFound.backHome') || '返回首页' }}
        </el-button>
        <el-button @click="goBack">
          {{ $t('notFound.goBack') || '返回上一页' }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { DocumentDelete } from '@element-plus/icons-vue'

const router = useRouter()

const goHome = () => {
  router.push('/')
}

const goBack = () => {
  if (window.history.length > 1) {
    router.go(-1)
  } else {
    router.push('/')
  }
}
</script>

<style scoped>
.not-found-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
  padding: 20px;
}

.not-found-content {
  text-align: center;
}

.not-found-icon {
  margin-bottom: 20px;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-20px);
  }
}

.not-found-title {
  font-size: 72px;
  font-weight: bold;
  color: var(--el-color-primary);
  margin: 0 0 20px 0;
  line-height: 1;
}

.not-found-message {
  font-size: 18px;
  color: #909399;
  margin: 0 0 40px 0;
}

.not-found-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}
</style>

