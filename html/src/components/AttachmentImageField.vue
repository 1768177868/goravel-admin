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

    <div v-if="displayPreviewUrl" class="logo-preview">
      <el-image
        :src="displayPreviewUrl"
        :preview-src-list="[displayPreviewUrl]"
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
      @closed="handlePickerClosed"
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
          <el-image
            v-if="getImageUrl(item)"
            :src="getImageUrl(item)"
            fit="cover"
            class="picker-thumb"
            @load="handleImageLoad(item)"
            @error="handleImageError(item)"
          >
            <template #error>
              <div class="picker-thumb-error">
                <el-icon><Picture /></el-icon>
              </div>
            </template>
          </el-image>
          <div v-else class="picker-thumb-error">
            <el-icon v-if="getImageLoadingState(item) === 'loading'" class="is-loading"><Loading /></el-icon>
            <el-icon v-else><Picture /></el-icon>
          </div>
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
import { ref, computed, watch, nextTick, onBeforeUnmount } from 'vue'
import { Picture, Loading } from '@element-plus/icons-vue'
import axios from 'axios'
import { getAttachmentList } from '@/api/attachment'
import { transformAttachmentRow } from '@/views/attachment/attachment.config'
import { useAttachmentImagePreview } from '@/composables/useAttachmentImagePreview'
import { resolvePublicAssetUrl, getApiPrefix } from '@/utils/env'
import Storage from '@/utils/storage'

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
const displayPreviewUrl = ref('')
let previewBlobUrl = ''

const {
  loadImageAsBlob,
  getImageUrl,
  getImageLoadingState,
  handleImageLoad,
  handleImageError
} = useAttachmentImagePreview()

const selectedItem = computed(() => list.value.find((item) => item.id === selectedId.value) || null)

const PUBLIC_IMAGE_RE = /\/api\/admin\/public\/images\/(\d+)/

const toSubmitUrl = (url) => {
  const value = String(url || '').trim()
  if (!value) return ''
  if (value.startsWith('/')) return value
  if (value.startsWith('http')) {
    try {
      const parsed = new URL(value)
      if (PUBLIC_IMAGE_RE.test(parsed.pathname)) {
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

const buildFetchUrl = (raw) => {
  const value = String(raw || '').trim()
  if (!value) return ''
  if (value.startsWith('data:') || value.startsWith('blob:')) return value
  if (/^https?:\/\//i.test(value)) return value

  const publicPath = resolvePublicAssetUrl(value)
  const path = publicPath.startsWith('/') ? publicPath : value
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL
  if (apiBaseURL) {
    return `${String(apiBaseURL).replace(/\/+$/, '')}${path.startsWith('/') ? path : `/${path}`}`
  }
  const prefix = getApiPrefix().startsWith('/') ? getApiPrefix() : `/${getApiPrefix()}`
  if (path.startsWith(prefix) || path.startsWith('/')) return path
  return `${prefix}/${path}`
}

const revokePreviewBlob = () => {
  if (previewBlobUrl) {
    URL.revokeObjectURL(previewBlobUrl)
    previewBlobUrl = ''
  }
}

const loadPreview = async (raw) => {
  revokePreviewBlob()
  displayPreviewUrl.value = ''

  const value = String(raw || '').trim()
  if (!value) return

  if (value.startsWith('data:') || value.startsWith('blob:')) {
    displayPreviewUrl.value = value
    return
  }

  const isPublicImage = PUBLIC_IMAGE_RE.test(value) || PUBLIC_IMAGE_RE.test(resolvePublicAssetUrl(value) || '')

  // 非公开图的外链直接展示
  if (/^https?:\/\//i.test(value) && !isPublicImage) {
    displayPreviewUrl.value = value
    return
  }

  const fetchUrl = buildFetchUrl(value)
  if (!fetchUrl) return

  // 与附件列表一致：鉴权拉取为 blob，避免开发环境跨域/代理导致不显示
  try {
    const token = Storage.getItem('token', '') || ''
    const response = await axios.get(fetchUrl, {
      responseType: 'blob',
      headers: {
        Authorization: `Bearer ${typeof token === 'string' ? token.trim() : ''}`
      }
    })
    previewBlobUrl = URL.createObjectURL(new Blob([response.data]))
    displayPreviewUrl.value = previewBlobUrl
  } catch {
    displayPreviewUrl.value = resolvePublicAssetUrl(value) || fetchUrl
  }
}

watch(
  () => props.modelValue,
  (val) => {
    loadPreview(val)
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  revokePreviewBlob()
})

const ensureFileUrl = (item) => {
  if (!item) return ''
  if (item.file_url) return item.file_url
  if (item.id) return `/api/admin/public/images/${item.id}`
  return ''
}

/** 配置类长期引用：统一存公开预览路径，避免 S3 临时签名链过期 */
const toStablePublicUrl = (item) => {
  if (!item?.id) return ''
  return `/api/admin/public/images/${item.id}`
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
    list.value = rows.map((row) => {
      const item = transformAttachmentRow(row)
      item.file_url = ensureFileUrl(item)
      return item
    })
    total.value = Number(res?.data?.total || 0)
    await nextTick()
    list.value.forEach((item) => {
      if (item.file_type === 'image') {
        loadImageAsBlob(item)
      }
    })
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

const handlePickerClosed = () => {
  list.value = []
  selectedId.value = null
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
  const url = toSubmitUrl(toStablePublicUrl(item))
  emitValue(url)
  pickerVisible.value = false
}
</script>

<style scoped>
.attachment-image-field {
  width: 100%;
  max-width: 480px;
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
  font-size: 28px;
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
