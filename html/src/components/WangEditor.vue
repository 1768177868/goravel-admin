<template>
  <div style="border: 1px solid #ccc">
    <Toolbar
      style="border-bottom: 1px solid #ccc"
      :editor="editorRef"
      :defaultConfig="toolbarConfig"
      :mode="mode"
    />
    <Editor
      :style="{ height: height + 'px', overflowY: 'hidden' }"
      v-model="valueHtml"
      :defaultConfig="editorConfig"
      :mode="mode"
      @onCreated="handleCreated"
      @onChange="handleChange"
    />
  </div>
</template>

<script setup>
import '@wangeditor/editor/dist/css/style.css' // 引入 css
import { onBeforeUnmount, ref, shallowRef, onMounted, watch, computed } from 'vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import Storage from '../utils/storage'
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  mode: {
    type: String,
    default: 'default' // 'default' or 'simple'
  },
  height: {
    type: Number,
    default: 300
  },
  placeholder: {
    type: String,
    default: '请输入内容...'
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

// 编辑器实例，必须用 shallowRef
const editorRef = shallowRef()

// 内容 HTML
const valueHtml = ref('')

// 初始化内容
onMounted(() => {
    valueHtml.value = props.modelValue
})

// 监听 props 变化 (外部修改 v-model)
watch(() => props.modelValue, (newVal) => {
    if (newVal !== valueHtml.value) {
        valueHtml.value = newVal
    }
})

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

const toolbarConfig = {
    excludeKeys: [
        'group-video' // 排除视频相关菜单
    ]
}
const editorConfig = { 
    placeholder: props.placeholder,
    MENU_CONF: {
        uploadImage: {
            server: uploadAction.value,
            fieldName: 'file',
            headers: uploadHeaders.value,
            maxFileSize: 5 * 1024 * 1024, // 5M
            maxNumberOfFiles: 10,
            allowedFileTypes: ['image/*'],
            metaWithUrl: false,
            withCredentials: false,
            timeout: 10 * 1000, // 10 秒
            customInsert(res, insertFn) {
                // res 即服务端的返回结果
                // 从 res 中找到 url alt href ，然后插入图片
                if (res.code === 200 && res.data) {
                    const apiBaseURL = import.meta.env.VITE_API_BASE_URL
                    const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'
                    
                    // 优先使用 preview_url，如果没有则使用 file_url
                    let url = res.data.preview_url || res.data.file_url
                    
                    // 如果不是完整的 URL，则需要拼接
                    if (url && !url.startsWith('http')) {
                        if (apiBaseURL) {
                            const base = apiBaseURL.replace(/\/+$/, '')
                            // 如果 url 已经包含 apiPrefix (例如 /api/admin/attachments/...), 则只拼接 base
                            // 否则拼接 base + apiPrefix
                            const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
                            if (url.startsWith(prefix)) {
                                url = `${base}${url}`
                            } else {
                                url = `${base}${prefix}${url.startsWith('/') ? '' : '/'}${url}`
                            }
                        } else {
                            // 如果没有 base URL，假设是相对路径
                            const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
                             if (!url.startsWith(prefix)) {
                                url = `${prefix}${url.startsWith('/') ? '' : '/'}${url}`
                            }
                        }
                    }
                    
                    const alt = res.data.filename
                    const href = url // 图片链接点击跳转
                    
                    // 如果 URL 需要认证，且不是 http 开头（说明是相对路径，且需要鉴权），则尝试 Blob
                    // 但由于我们已经公开了 preview 接口，理论上不需要 Blob
                    // 但如果用户坚持要像附件列表那样（可能附件列表逻辑是为了处理非公开接口），我们可以保留
                    // 不过，最简单的方式是直接插入 URL，因为我们已经在后端公开了接口
                    insertFn(url, alt, href)
                } else {
                    console.error('Upload error', res)
                }
            }
        }
    }
}

// 监听上传配置变化（如 token 更新）
watch([uploadAction, uploadHeaders], () => {
    if (editorConfig.MENU_CONF.uploadImage) {
        editorConfig.MENU_CONF.uploadImage.server = uploadAction.value
        editorConfig.MENU_CONF.uploadImage.headers = uploadHeaders.value
    }
})

// 组件销毁时，也及时销毁编辑器
onBeforeUnmount(() => {
    const editor = editorRef.value
    if (editor == null) return
    editor.destroy()
})

const handleCreated = (editor) => {
    editorRef.value = editor // 记录 editor 实例，重要！
}

const handleChange = (editor) => {
    emit('update:modelValue', editor.getHtml())
    emit('change', editor.getHtml())
}
</script>
