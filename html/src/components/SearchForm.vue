<template>
  <el-form :model="model" :inline="true" class="search-form">
    <slot :expanded="expanded" />
    <el-form-item v-if="showExpand">
      <el-button @click="toggleExpand">
        <el-icon><component :is="expanded ? 'ArrowUp' : 'ArrowDown'" /></el-icon>
        {{ expanded ? $t('log.collapse') : $t('log.expand') }}
      </el-button>
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="handleSearch">
        <el-icon><Search /></el-icon>
        {{ $t('log.search') }}
      </el-button>
      <el-button @click="handleReset">
        <el-icon><Refresh /></el-icon>
        {{ $t('log.reset') }}
      </el-button>
      <slot name="extra-buttons" />
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref } from 'vue'
import { Search, Refresh, ArrowUp, ArrowDown } from '@element-plus/icons-vue'

const props = defineProps({
  model: {
    type: Object,
    required: true
  },
  showExpand: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['search', 'reset', 'expand-change'])

const expanded = ref(false)

const toggleExpand = () => {
  expanded.value = !expanded.value
  emit('expand-change', expanded.value)
}

const handleSearch = () => {
  emit('search')
}

const handleReset = () => {
  emit('reset')
}

defineExpose({
  expanded
})
</script>

<style scoped>
.search-form {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
}
</style>

