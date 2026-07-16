<template>
  <div class="attachment-image-field">
    <el-input
      :model-value="modelValue"
      :placeholder="placeholder || $t('config.site_logo_placeholder')"
      clearable
      @update:model-value="emitValue"
    >
      <template #append>
        <el-button @click="openPicker">
          <el-icon><Picture /></el-icon>
          {{ $t('common.select_from_attachment') }}
        </el-button>
      </template>
    </el-input>

    <div v-if="previewUrl" class="logo-preview">
      <el-image
        :src="previewUrl"
        :preview-src-list="[previewUrl]"
        fit="contain"
        class="logo-preview-image"
        :preview-teleported="true"
      >
        <template #error>
          <div class="logo-preview-error">
            <el-icon><Picture /></el-icon>
          </div>
        </template>
      </el-image>
    </div>

    <el-dialog
      v-model="pickerVisible"
      :title="$t('common.select_from_attachment')"
      width="780px"
      append-to-body
      destroy-on-close
      @opened="handlePickerOpened"
    >
      <div class="picker-toolbar">
        <el-input
          v-model="keyword"
          clearable
          :placeholder="$t('attachment.filename')"
          style="width: 240px"
          @keyup.enter="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">{{ $t('common.search') }}</el-button>
        <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
      </div>

      <div v-loading="loading" class="picker-grid">
        <div
          v-for="item in list"
          :key="item.id"
          class="picker-item"
          :class="{ active: selectedId === item.id }"
          @click="selectedId = item.id"
          @dblclick="confirmSelect(item)"
        >
          <el-image :src="getThumbUrl(item)" fit="cover" class="picker-thumb">
            <template #error>
              <div class="picker-thumb-error">
                <el-icon><Picture /></el-icon>
              </div>
            </template>
          </el-image>
          <div class="picker-name" :title="item.display_name || item.filename">
            {{ item.display_name || item.filename }}
          </div>
        </div>
        <el-empty v-if="!loading && list.length === 0" :description="$t('common.no_data')" />
      </div>

      <div class="picker-pager">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>

      <template #footer>
        <el-button @click="pickerVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :disabled="!selectedItem" @click="confirmSelect(selectedItem)">
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Picture } from '@element-plus/icons-vue'
import { getAttachmentList } from '@/api/attachment'
import { transformAttachmentRow } from '@/views/attachment/attachment.config'
import { resolvePublicAssetUrl } from '@/utils/env'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  placeholder: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const pickerVisible = ref(false)
const loading = ref(false)
const list = ref([])
const keyword = ref('')
const page = ref(1)
const pageSize = ref(12)
const total = ref(0)
const selectedId = ref(null)

const selectedItem = computed(() => list.value.find((item) => item.id === selectedId.value) || null)

const previewUrl = computed(() => {
  const raw = String(props.modelValue || '').trim()
  if (!raw) return ''
  if (raw.startsWith('data:') || raw.startsWith('blob:')) return raw
  const publicUrl = resolvePublicAssetUrl(raw)
  if (publicUrl.startsWith('/')) return publicUrl
  if (/^(https?:)?\/\//i.test(raw)) return raw
  return publicUrl || raw
})

const toSubmitUrl = (url) => {
  const value = String(url || '').trim()
  if (!value) return ''
  if (value.startsWith('/')) return value
  if (value.startsWith('http')) {
    try {
      const parsed = new URL(value)
      if (parsed.pathname.includes('/api/admin/public/images/')) {
        return `${parsed.pathname}${parsed.search || ''}`
      }
      return value
    } catch {
      return value
    }
  }
  return value
}

const emitValue = (value) => {
  const finalValue = toSubmitUrl(value)
  emit('update:modelValue', finalValue)
  emit('change', finalValue)
}

const getThumbUrl = (item) => {
  const raw = item?.file_url || ''
  if (!raw) return ''
  const publicUrl = resolvePublicAssetUrl(raw)
  if (publicUrl.startsWith('/')) return publicUrl
  if (/^(https?:)?\/\//i.test(raw)) return raw
  return publicUrl || raw
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await getAttachmentList({
      page: page.value,
      page_size: pageSize.value,
      file_type: 'image',
      filename: keyword.value.trim() || undefined
    })
    const rows = res?.data?.list || res?.data?.data || []
    list.value = rows.map(transformAttachmentRow)
    total.value = Number(res?.data?.total || 0)
  } catch {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const openPicker = () => {
  selectedId.value = null
  pickerVisible.value = true
}

const handlePickerOpened = () => {
  page.value = 1
  loadList()
}

const handleSearch = () => {
  page.value = 1
  selectedId.value = null
  loadList()
}

const handleReset = () => {
  keyword.value = ''
  page.value = 1
  selectedId.value = null
  loadList()
}

const handlePageChange = (newPage) => {
  page.value = newPage
  selectedId.value = null
  loadList()
}

const confirmSelect = (item) => {
  if (!item) return
  const url = toSubmitUrl(item.file_url || `/api/admin/public/images/${item.id}`)
  emitValue(url)
  pickerVisible.value = false
}
</script>

<style scoped>
.attachment-image-field {
  width: 100%;
}

.logo-preview {
  margin-top: 12px;
}

.logo-preview-image {
  width: 120px;
  height: 120px;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  background: var(--el-fill-color-lighter);
}

.logo-preview-error {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--el-text-color-placeholder);
  font-size: 28px;
}

.picker-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.picker-grid {
  min-height: 280px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.picker-item {
  border: 2px solid transparent;
  border-radius: 6px;
  padding: 8px;
  cursor: pointer;
  background: var(--el-fill-color-lighter);
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.picker-item:hover {
  border-color: var(--el-color-primary-light-5);
}

.picker-item.active {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 1px var(--el-color-primary-light-7);
}

.picker-thumb {
  width: 100%;
  height: 110px;
  border-radius: 4px;
  display: block;
}

.picker-thumb-error {
  width: 100%;
  height: 110px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--el-text-color-placeholder);
  background: var(--el-fill-color);
  border-radius: 4px;
}

.picker-name {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.4;
  color: var(--el-text-color-regular);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.picker-pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

:deep(.el-empty) {
  grid-column: 1 / -1;
}
</style>
