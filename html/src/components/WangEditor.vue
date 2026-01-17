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
import { onBeforeUnmount, ref, shallowRef, onMounted, watch } from 'vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'

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

const toolbarConfig = {}
const editorConfig = { 
    placeholder: props.placeholder,
    MENU_CONF: {}
}

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
