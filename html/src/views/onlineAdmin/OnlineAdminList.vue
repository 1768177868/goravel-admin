<template>
  <div class="online-admin-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.online_admin') }}</span>
          <el-button
            type="danger"
            :disabled="selectedRows.length === 0 || getButtonState('admin.kick_out').disabled"
            @click="handleBatchKickOut"
          >
            <el-icon><Delete /></el-icon>
            {{ t('online_admin.batch_kick_out') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="online_admin"
        @search="handleSearch"
        @reset="handleReset"
      />

      <!-- 表格 -->
      <vxe-grid
        ref="gridRef"
        :data="tableData"
        :columns="columns"
        :loading="loading"
        height="600"
        border
        :toolbar-config="toolbarConfig"
        :column-config="{ resizable: true }"
        :sort-config="{ multiple: false, trigger: 'default' }"
        @sort-change="handleSortChange"
        @checkbox-change="handleCheckboxChange"
        @checkbox-all="handleCheckboxAll"
      />

      <!-- 分页 -->
      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { ElMessage, ElMessageBox, ElAvatar, ElButton } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

import SearchForm from '@/components/SearchForm.vue'
import Pagination from '@/components/Pagination.vue'
import { usePermission } from '@/composables/usePermission'
import { useListPage } from '@/composables/useListPage'
import {
  getOnlineAdminList,
  kickOutOnlineAdmin,
  batchKickOutOnlineAdmins
} from '@/api/onlineAdmin'

/* ================= 基础 ================= */

const { t } = useI18n()
const { getButtonState } = usePermission()

const gridRef = ref(null)
const selectedRows = ref([])

/* ================= 列表逻辑 ================= */

const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getOnlineAdminList,
  initialSearchForm: {
    username: '',
    ip: '',
    browser: '',
    os: ''
  },
  fieldMapping: {
    last_active: 'last_used_at'
  },
  defaultSort: 'last_used_at:desc',
  tableRef: computed(() => gridRef.value)
})

/* ================= 搜索 ================= */

const initialSearchForm = {
  username: '',
  ip: '',
  browser: '',
  os: ''
}

const searchFields = computed(() => [
  { prop: 'username', label: t('online_admin.username'), type: 'input', width: '200px' },
  { prop: 'ip', label: t('online_admin.ip'), type: 'input', width: '200px' },
  { prop: 'browser', label: t('online_admin.browser'), type: 'input', width: '200px' },
  { prop: 'os', label: t('online_admin.os'), type: 'input', width: '200px' }
])

/* ================= 表格列 ================= */

const columns = computed(() => [
  { type: 'checkbox', width: 50, fixed: 'left' },

  { field: 'username', title: t('online_admin.username'), width: 120 },
  { field: 'nickname', title: t('online_admin.nickname'), width: 120 },

  {
    field: 'avatar',
    title: t('online_admin.avatar'),
    width: 80,
    slots: {
      default: ({ row }) =>
        h(
          ElAvatar,
          { size: 32, src: row.avatar },
          () =>
            row.nickname?.charAt(0) ||
            row.username?.charAt(0) ||
            'U'
        )
    }
  },

  { field: 'browser', title: t('online_admin.browser'), width: 150 },
  { field: 'ip', title: t('online_admin.ip'), width: 150 },
  { field: 'os', title: t('online_admin.os'), width: 150 },
  { field: 'session_id', title: t('online_admin.session_id'), width: 200 },

  {
    field: 'last_active',
    title: t('online_admin.last_active'),
    width: 180,
    sortable: true,
    slots: {
      default: ({ row }) => formatTime(row.last_active)
    }
  },

  {
    title: t('common.operation'),
    width: 120,
    fixed: 'right',
    slots: {
      default: ({ row }) =>
        h(
          ElButton,
          {
            type: 'danger',
            link: true,
            disabled: getButtonState('admin.kick_out').disabled,
            onClick: () => handleKickOut(row)
          },
          () => t('online_admin.kick_out')
        )
    }
  }
])

/* ================= 工具栏 ================= */

const toolbarConfig = computed(() => ({
  custom: true,
  refresh: {
    query: loadData
  },
  zoom: true
}))

/* ================= 勾选 ================= */

const handleCheckboxChange = ({ records }) => {
  selectedRows.value = records || []
}

const handleCheckboxAll = ({ records }) => {
  selectedRows.value = records || []
}

/* ================= 操作 ================= */

const handleKickOut = async (row) => {
  try {
    await ElMessageBox.confirm(
      t('online_admin.kick_out_confirm', { username: row.username }),
      t('common.confirm'),
      { type: 'warning' }
    )
    await kickOutOnlineAdmin(row.id)
    ElMessage.success(t('online_admin.kick_out_success'))
    loadData()
  } catch {}
}

const handleBatchKickOut = async () => {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      t('online_admin.batch_kick_out_confirm', {
        count: selectedRows.value.length
      }),
      t('common.confirm'),
      { type: 'warning' }
    )
    await batchKickOutOnlineAdmins(selectedRows.value.map(i => i.id))
    ElMessage.success(t('online_admin.batch_kick_out_success'))
    selectedRows.value = []
    loadData()
  } catch {}
}

/* ================= 工具 ================= */

const formatTime = (time) => {
  if (!time) return '-'
  const d = new Date(time)
  if (isNaN(d.getTime())) return time
  return d.toLocaleString().replace(/\//g, '-')
}

onMounted(() => {
  initDefaultSort()
  loadData()
})
</script>

<style scoped>
.online-admin-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
