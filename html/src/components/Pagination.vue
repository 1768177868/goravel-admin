<template>
  <div :class="['pagination-wrapper', `align-${align}`, { 'pagination-compact': compact }]" :style="wrapperStyle">
    <!-- 总数信息 -->
    <div v-if="showTotal" class="pagination-total">
      <span>{{ totalTextComputed }}</span>
    </div>
    
    <!-- 分页组件 -->
    <vxe-pager
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="pageSizes"
      :layouts="layouts"
      :border="border"
      :background="background"
      :pager-count="pagerCount"
      :disabled="disabled"
      :loading="loading"
      @page-change="handlePageChange"
      @page-size-change="handlePageSizeChange"
    />
    
    <!-- 快速跳转 -->
    <div v-if="showQuickJumper && totalPages > 0" class="pagination-jumper">
      <span>{{ jumpText }}</span>
      <el-input-number
        v-model="jumpPage"
        :min="1"
        :max="Math.max(1, totalPages)"
        :size="inputSize"
        :controls="false"
        style="width: 80px; margin: 0 8px"
        @keyup.enter="handleJump"
      />
      <el-button :size="buttonSize" @click="handleJump">{{ confirmText }}</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  // 分页数据模型
  modelValue: {
    type: Object,
    required: true,
    default: () => ({
      page: 1,
      pageSize: 10,
      total: 0
    })
  },
  // 每页显示条数选项
  pageSizes: {
    type: Array,
    default: () => [10, 20, 50, 100]
  },
  // 布局配置
  layouts: {
    type: Array,
    default: () => ['PrevJump', 'PrevPage', 'Number', 'NextPage', 'NextJump', 'Sizes', 'FullJump', 'Total']
  },
  // 是否显示边框
  border: {
    type: Boolean,
    default: false
  },
  // 是否为分页按钮添加背景色
  background: {
    type: Boolean,
    default: false
  },
  // 页码按钮的数量
  pagerCount: {
    type: Number,
    default: 7
  },
  // 是否禁用
  disabled: {
    type: Boolean,
    default: false
  },
  // 是否显示加载状态
  loading: {
    type: Boolean,
    default: false
  },
  // 是否显示总数
  showTotal: {
    type: Boolean,
    default: true
  },
  // 是否显示快速跳转
  showQuickJumper: {
    type: Boolean,
    default: false
  },
  // 是否紧凑模式
  compact: {
    type: Boolean,
    default: false
  },
  // 总数文本模板
  totalText: {
    type: String,
    default: ''
  },
  // 跳转文本
  jumpText: {
    type: String,
    default: '跳至'
  },
  // 确认文本
  confirmText: {
    type: String,
    default: '确定'
  },
  // 输入框尺寸
  inputSize: {
    type: String,
    default: 'small',
    validator: (value) => ['large', 'default', 'small'].includes(value)
  },
  // 按钮尺寸
  buttonSize: {
    type: String,
    default: 'small',
    validator: (value) => ['large', 'default', 'small'].includes(value)
  },
  // 自定义样式
  wrapperStyle: {
    type: Object,
    default: () => ({})
  },
  // 对齐方式
  align: {
    type: String,
    default: 'right',
    validator: (value) => ['left', 'center', 'right'].includes(value)
  }
})

const emit = defineEmits(['update:modelValue', 'page-change', 'page-size-change', 'change'])

const jumpPage = ref(1)

// 当前页
const currentPage = computed({
  get: () => props.modelValue.page || 1,
  set: (value) => {
    updateModelValue({ page: value })
  }
})

// 每页条数
const pageSize = computed({
  get: () => props.modelValue.pageSize || 10,
  set: (value) => {
    updateModelValue({ pageSize: value, page: 1 })
  }
})

// 总数
const total = computed(() => props.modelValue.total || 0)

// 总页数
const totalPages = computed(() => {
  if (total.value === 0 || pageSize.value === 0) return 0
  return Math.ceil(total.value / pageSize.value)
})

// 总数文本
const totalTextComputed = computed(() => {
  if (props.totalText) {
    return props.totalText
      .replace('{total}', total.value)
      .replace('{start}', ((currentPage.value - 1) * pageSize.value + 1))
      .replace('{end}', Math.min(currentPage.value * pageSize.value, total.value))
  }
  const start = total.value === 0 ? 0 : (currentPage.value - 1) * pageSize.value + 1
  const end = Math.min(currentPage.value * pageSize.value, total.value)
  return `共 ${total.value} 条，显示第 ${start}-${end} 条`
})

// 更新模型值
const updateModelValue = (updates) => {
  const newValue = {
    ...props.modelValue,
    ...updates
  }
  emit('update:modelValue', newValue)
  emit('change', newValue)
}

// 页码变化处理
const handlePageChange = ({ currentPage, pageSize }) => {
  updateModelValue({
    page: currentPage,
    pageSize: pageSize
  })
  emit('page-change', { currentPage, pageSize })
  jumpPage.value = currentPage
}

// 每页条数变化处理
const handlePageSizeChange = ({ pageSize }) => {
  updateModelValue({
    pageSize: pageSize,
    page: 1
  })
  emit('page-size-change', { pageSize })
  emit('page-change', { currentPage: 1, pageSize })
}

// 快速跳转
const handleJump = () => {
  const page = Number(jumpPage.value)
  if (page >= 1 && page <= totalPages.value && page !== currentPage.value) {
    currentPage.value = page
  } else {
    jumpPage.value = currentPage.value
  }
}

// 监听当前页变化，同步跳转输入框
watch(() => currentPage.value, (val) => {
  jumpPage.value = val
}, { immediate: true })

// 监听总页数变化，限制跳转输入框最大值
watch(totalPages, (val) => {
  if (jumpPage.value > val) {
    jumpPage.value = val || 1
  }
})
</script>

<style scoped lang="scss">
.pagination-wrapper {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 20px;
  padding: 16px 0;

  &.pagination-compact {
    margin-top: 15px;
    padding: 12px 0;
    gap: 12px;
  }

  .pagination-total {
    color: var(--text-color-regular, #606266);
    font-size: 14px;
    white-space: nowrap;
    transition: color 0.3s ease;
  }

  .pagination-jumper {
    display: flex;
    align-items: center;
    color: var(--text-color-regular, #606266);
    font-size: 14px;
    white-space: nowrap;
    transition: color 0.3s ease;
  }

  // 对齐方式
  &.align-left {
    justify-content: flex-start;
  }

  &.align-center {
    justify-content: center;
  }

  &.align-right {
    justify-content: flex-end;
  }

  // 响应式布局
  @media (max-width: 768px) {
    flex-direction: column;
    align-items: stretch;

    .pagination-total,
    .pagination-jumper {
      width: 100%;
      justify-content: center;
      margin-bottom: 8px;
    }

    :deep(.vxe-pager) {
      width: 100%;
      justify-content: center;
    }
  }
}
</style>
