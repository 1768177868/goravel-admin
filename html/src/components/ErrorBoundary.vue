<template>
  <div v-if="hasError" class="error-boundary">
    <el-result
      icon="error"
      :title="title"
      :sub-title="subTitle"
    >
      <template #extra>
        <el-button type="primary" @click="handleReset">重试</el-button>
        <el-button @click="handleReload">刷新页面</el-button>
      </template>
    </el-result>
    
    <!-- 开发环境显示错误详情 -->
    <el-collapse v-if="isDev && error" class="error-details">
      <el-collapse-item title="错误详情（开发环境）" name="details">
        <pre class="error-stack">{{ errorStack }}</pre>
      </el-collapse-item>
    </el-collapse>
  </div>
  <slot v-else />
</template>

<script setup>
import { ref, computed, onErrorCaptured } from 'vue'
import { useI18n } from 'vue-i18n'
import logger from '../utils/logger'
import { isDev } from '../utils/env'

const props = defineProps({
  title: {
    type: String,
    default: '出现错误'
  },
  subTitle: {
    type: String,
    default: '页面加载时出现错误，请刷新页面或联系技术支持'
  },
  onError: {
    type: Function,
    default: null
  }
})

const emit = defineEmits(['error', 'reset'])

const hasError = ref(false)
const error = ref(null)
const errorStack = ref('')

const { t } = useI18n()

// 捕获子组件错误
onErrorCaptured((err, instance, info) => {
  hasError.value = true
  error.value = err
  
  // 构建错误堆栈信息
  errorStack.value = [
    `错误信息: ${err.message}`,
    `错误堆栈: ${err.stack || 'N/A'}`,
    `组件信息: ${info || 'N/A'}`,
    `实例: ${instance?.$?.type?.name || 'Unknown'}`
  ].join('\n')

  // 记录错误
  logger.error('ErrorBoundary caught error:', {
    error: err,
    instance,
    info,
    stack: errorStack.value
  })

  // 触发错误回调
  if (props.onError) {
    props.onError(err, instance, info)
  }

  emit('error', err, instance, info)

  // 阻止错误继续传播
  return false
})

/**
 * 重置错误状态
 */
const handleReset = () => {
  hasError.value = false
  error.value = null
  errorStack.value = ''
  emit('reset')
}

/**
 * 重新加载页面
 */
const handleReload = () => {
  window.location.reload()
}
</script>

<style scoped>
.error-boundary {
  padding: 20px;
  min-height: 400px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.error-details {
  margin-top: 20px;
  width: 100%;
  max-width: 800px;
}

.error-stack {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 400px;
  overflow-y: auto;
}
</style>

