<template>
  <vxe-pager
    v-model:current-page="currentPage"
    v-model:page-size="pageSize"
    :total="total"
    :page-sizes="pageSizes"
    @page-change="handlePageChange"
  />
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    required: true,
    default: () => ({
      page: 1,
      pageSize: 10,
      total: 0
    })
  },
  pageSizes: {
    type: Array,
    default: () => [10, 20, 50, 100]
  }
})

const emit = defineEmits(['update:modelValue', 'page-change'])

const currentPage = computed({
  get: () => props.modelValue.page,
  set: (value) => {
    emit('update:modelValue', {
      ...props.modelValue,
      page: value
    })
  }
})

const pageSize = computed({
  get: () => props.modelValue.pageSize,
  set: (value) => {
    emit('update:modelValue', {
      ...props.modelValue,
      pageSize: value
    })
  }
})

const total = computed(() => props.modelValue.total)

const handlePageChange = ({ currentPage, pageSize }) => {
  emit('update:modelValue', {
    ...props.modelValue,
    page: currentPage,
    pageSize: pageSize
  })
  emit('page-change', { currentPage, pageSize })
}
</script>

