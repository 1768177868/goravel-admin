<template>
  <div class="markdown-editor-wrapper" :style="{ width: normalizedWidth }">
    <MdEditor
      v-model="markdownValue"
      :height="height"
      :placeholder="placeholder"
      :toolbars="toolbars"
      :theme="isDark ? 'dark' : 'light'"
      @onUploadImg="handleUploadImg"
      @onChange="handleChange"
    />
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import Storage from '../utils/storage'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../store/app'
import axios from 'axios'

const { locale } = useI18n()
const appStore = useAppStore()

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  width: {
    type: [String, Number],
    default: '100%'
  },
  height: {
    type: Number,
    default: 300
  },
  placeholder: {
    type: String,
    default: '请输入内容...'
  },
  toolbars: {
    type: Array,
    default: () => [
      'bold', 'underline', 'italic', '-',
      'title', 'strikeThrough', 'sub', 'sup',
      'quote', 'unorderedList', 'orderedList', 'task', '-',
      'codeRow', 'code', 'link', 'image', 'table', '-',
      'revoke', 'next', 'save',
      '=',
      'pageFullscreen', 'fullscreen', 'preview', 'catalog'
    ]
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const markdownValue = ref(props.modelValue)
const isDark = computed(() => appStore.darkMode)
const normalizedWidth = computed(() =>
  typeof props.width === 'number' ? `${props.width}px` : props.width
)

// 监听 props 变化
watch(() => props.modelValue, (newVal) => {
  if (newVal !== markdownValue.value) {
    markdownValue.value = newVal
  }
})

// 上传配置
const uploadAction = computed(() => {
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL
  const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
  if (apiBaseURL) {
    const base = apiBaseURL.replace(/\/+$/, '')
    const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
    return `${base}${prefix}/attachments/upload`
  }
  return `${apiPrefix}/attachments/upload`
})

const uploadHeaders = computed(() => {
  const token = Storage.getItem('token', '') || ''
  return {
    'Authorization': `Bearer ${typeof token === 'string' ? token.trim() : ''}`,
    'Accept-Language': locale.value === 'en-US' ? 'en-US' : 'zh-CN'
  }
})

// 处理图片上传
const handleUploadImg = async (files, callback) => {
  const formData = new FormData()
  const file = files[0]
  
  if (!file) {
    callback([])
    return
  }

  // 检查文件大小（5M）
  if (file.size > 5 * 1024 * 1024) {
    callback([])
    return
  }

  // 检查文件类型
  if (!file.type.startsWith('image/')) {
    callback([])
    return
  }

  formData.append('file', file)

  try {
    const response = await axios.post(uploadAction.value, formData, {
      headers: uploadHeaders.value,
      timeout: 10000
    })

    if (response.data.code === 200 && response.data.data) {
      const apiBaseURL = import.meta.env.VITE_API_BASE_URL
      const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
      
      // 优先使用 preview_url，如果没有则使用 file_url
      let url = response.data.data.preview_url || response.data.data.file_url
      
      // 如果不是完整的 URL，则需要拼接
      if (url && !url.startsWith('http')) {
        if (apiBaseURL) {
          const base = apiBaseURL.replace(/\/+$/, '')
          const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
          if (url.startsWith(prefix)) {
            url = `${base}${url}`
          } else {
            url = `${base}${prefix}${url.startsWith('/') ? '' : '/'}${url}`
          }
        } else {
          const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
          if (!url.startsWith(prefix)) {
            url = `${prefix}${url.startsWith('/') ? '' : '/'}${url}`
          }
        }
      }
      
      callback([url])
    } else {
      console.error('Upload error', response.data)
      callback([])
    }
  } catch (error) {
    console.error('Upload error', error)
    callback([])
  }
}

// 处理内容变化
const handleChange = (value) => {
  markdownValue.value = value
  emit('update:modelValue', value)
  emit('change', value)
}
</script>

<style scoped>
.markdown-editor-wrapper {
  width: 100%;
  min-width: 0;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}

.markdown-editor-wrapper :deep(.md-editor) {
  width: 100%;
  min-width: 0;
}

/* 暗黑模式样式 */
html.dark .markdown-editor-wrapper {
  border-color: var(--el-border-color);
}

/* 确保 md-editor-v3 使用 Element Plus 的颜色变量 */
html.dark :deep(.md-editor) {
  background-color: var(--el-bg-color) !important;
  color: var(--el-text-color-regular) !important;
}

html.dark :deep(.md-editor .md-editor-toolbar) {
  background-color: var(--el-bg-color) !important;
  border-color: var(--el-border-color) !important;
}

html.dark :deep(.md-editor .md-editor-toolbar button) {
  color: var(--el-text-color-regular) !important;
}

html.dark :deep(.md-editor .md-editor-toolbar button:hover) {
  background-color: var(--el-fill-color-light) !important;
}

html.dark :deep(.md-editor .md-editor-textarea) {
  background-color: var(--el-bg-color) !important;
  color: var(--el-text-color-regular) !important;
}

html.dark :deep(.md-editor .md-editor-textarea::placeholder) {
  color: var(--el-text-color-placeholder) !important;
}

html.dark :deep(.md-editor .md-editor-preview) {
  background-color: var(--el-bg-color) !important;
  color: var(--el-text-color-regular) !important;
}
</style>
