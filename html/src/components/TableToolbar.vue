<template>
  <div class="table-toolbar">
    <div class="toolbar-left">
      <slot name="left"></slot>
    </div>
    <div class="toolbar-right">
      <el-button-group>
        <!-- 刷新按钮 -->
        <el-button 
          v-if="showRefresh"
          :icon="RefreshIcon" 
          circle
          :class="{ 'is-refreshing': isRefreshing }"
          @click="handleRefresh"
          :title="$t('common.refresh')"
        />
        <!-- 全屏按钮 -->
        <el-button 
          v-if="showFullscreen"
          :icon="FullScreenIcon" 
          circle
          @click="handleFullscreen"
          :title="$t('common.fullscreen')"
        />
        <!-- 列设置按钮 -->
        <ColumnSettingDialog
          v-if="showColumnSettingBtn"
          v-model="showColumnSettingDialog"
          :visible-columns="visibleColumns"
          :all-columns="allColumns"
          :default-visible-columns="defaultVisibleColumns"
          :column-order="columnOrder"
          :fixed-columns="fixedColumns"
          @confirm="handleColumnSettingConfirm"
        />
      </el-button-group>
      <slot name="right"></slot>
    </div>
  </div>
</template>

<script setup>
import { ref, markRaw, onMounted, onBeforeUnmount } from 'vue'
import { Refresh, FullScreen } from '@element-plus/icons-vue'
import ColumnSettingDialog from './ColumnSettingDialog.vue'

const props = defineProps({
  // 是否显示刷新按钮
  showRefresh: {
    type: Boolean,
    default: true
  },
  // 是否显示全屏按钮
  showFullscreen: {
    type: Boolean,
    default: true
  },
  // 是否显示列设置按钮
  showColumnSettingBtn: {
    type: Boolean,
    default: true
  },
  // 刷新回调函数
  onRefresh: {
    type: Function,
    default: null
  },
  // 全屏目标元素选择器（如 '.online-admin-list'）或元素引用
  fullscreenTarget: {
    type: [String, Object],
    default: null
  },
  // 列设置相关 props
  visibleColumns: {
    type: Array,
    default: () => []
  },
  allColumns: {
    type: Array,
    default: () => []
  },
  defaultVisibleColumns: {
    type: Array,
    default: () => []
  },
  columnOrder: {
    type: Array,
    default: () => []
  },
  fixedColumns: {
    type: Object,
    default: () => ({})
  },
  // 列设置确认回调
  onColumnSettingConfirm: {
    type: Function,
    default: null
  }
})

const emit = defineEmits(['refresh', 'fullscreen-change'])

// 图标
const RefreshIcon = markRaw(Refresh)
const FullScreenIcon = markRaw(FullScreen)

// 列设置对话框显示状态
const showColumnSettingDialog = ref(false)

// 全屏状态
const isFullscreen = ref(false)
// 刷新动画状态
const isRefreshing = ref(false)

// 处理刷新
const handleRefresh = async () => {
  if (isRefreshing.value) return
  isRefreshing.value = true
  const startAt = Date.now()

  try {
    if (props.onRefresh) {
      await props.onRefresh()
    }
    emit('refresh')
  } finally {
    // 给动画一个最短展示时间，避免点击后几乎不可见
    const elapsed = Date.now() - startAt
    const minDuration = 450
    const remain = Math.max(0, minDuration - elapsed)
    window.setTimeout(() => {
      isRefreshing.value = false
    }, remain)
  }
}

// 处理全屏
const handleFullscreen = () => {
  let el = null
  
  if (props.fullscreenTarget) {
    if (typeof props.fullscreenTarget === 'string') {
      el = document.querySelector(props.fullscreenTarget)
    } else if (props.fullscreenTarget.value) {
      // 如果是 ref 对象
      el = props.fullscreenTarget.value
    } else {
      // 如果直接是元素
      el = props.fullscreenTarget
    }
  }
  
  if (!el) {
    // 如果没有指定目标，尝试查找最近的 .table-toolbar 的父容器
    const toolbar = document.querySelector('.table-toolbar')
    if (toolbar) {
      el = toolbar.closest('.el-card') || toolbar.closest('[class*="list"]') || toolbar.parentElement
    }
  }
  
  if (!el) return
  
  if (!isFullscreen.value) {
    if (el.requestFullscreen) {
      el.requestFullscreen()
    } else if (el.webkitRequestFullscreen) {
      el.webkitRequestFullscreen()
    } else if (el.mozRequestFullScreen) {
      el.mozRequestFullScreen()
    } else if (el.msRequestFullscreen) {
      el.msRequestFullscreen()
    }
    isFullscreen.value = true
  } else {
    if (document.exitFullscreen) {
      document.exitFullscreen()
    } else if (document.webkitExitFullscreen) {
      document.webkitExitFullscreen()
    } else if (document.mozCancelFullScreen) {
      document.mozCancelFullScreen()
    } else if (document.msExitFullscreen) {
      document.msExitFullscreen()
    }
    isFullscreen.value = false
  }
  
  emit('fullscreen-change', isFullscreen.value)
}

// 处理列设置确认
const handleColumnSettingConfirm = (result) => {
  if (props.onColumnSettingConfirm) {
    props.onColumnSettingConfirm(result)
  }
}

// 监听全屏状态变化
const handleFullscreenChange = () => {
  isFullscreen.value = !!(
    document.fullscreenElement ||
    document.webkitFullscreenElement ||
    document.mozFullScreenElement ||
    document.msFullscreenElement
  )
  emit('fullscreen-change', isFullscreen.value)
}

onMounted(() => {
  // 初始化全屏状态
  isFullscreen.value = !!(
    document.fullscreenElement ||
    document.webkitFullscreenElement ||
    document.mozFullScreenElement ||
    document.msFullscreenElement
  )
  
  // 监听全屏状态变化
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.addEventListener('mozfullscreenchange', handleFullscreenChange)
  document.addEventListener('MSFullscreenChange', handleFullscreenChange)
})

onBeforeUnmount(() => {
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  document.removeEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.removeEventListener('mozfullscreenchange', handleFullscreenChange)
  document.removeEventListener('MSFullscreenChange', handleFullscreenChange)
})
</script>

<style scoped>
.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
  padding: 4px 6px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--card-bg, #fff) 92%, transparent);
  border: 1px solid color-mix(in srgb, var(--border-color-light) 70%, transparent);
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.05);
}

.toolbar-left {
  flex: 1;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-right :deep(.el-button-group) {
  display: inline-flex;
  gap: 6px;
}

.toolbar-right :deep(.el-button-group > .el-button) {
  border-radius: 10px !important;
  border: 1px solid color-mix(in srgb, var(--border-color-light) 72%, transparent);
  background: color-mix(in srgb, var(--card-bg, #fff) 94%, transparent);
  transition: all 0.2s ease;
}

.toolbar-right :deep(.el-button-group > .el-button:hover) {
  color: var(--el-color-primary);
  border-color: color-mix(in srgb, var(--el-color-primary) 30%, transparent);
  background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
}

.toolbar-right :deep(.el-button-group > .el-button.is-refreshing .el-icon) {
  animation: toolbar-refresh-spin 0.8s linear infinite;
}

@keyframes toolbar-refresh-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>