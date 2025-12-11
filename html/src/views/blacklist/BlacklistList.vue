<template>
  <div class="blacklist-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('blacklist.title') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('blacklist.store').disabled"
            @click="handleAdd"
          >
            <el-icon><Plus /></el-icon>
            {{ $t('blacklist.add_blacklist') }}
          </el-button>
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ ip: '', status: '' }"
        i18n-prefix="blacklist"
        @search="handleSearch"
        @reset="handleReset"
      />

      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        :column-config="{ resizable: true }"
        height="600"
        :sort-config="{ multiple: false, trigger: 'default' }"
        @sort-change="handleSortChange"
      >
        <template v-for="column in tableColumns" :key="column.field || column.title || column.type">
          <vxe-column
            v-if="column.type === 'checkbox'"
            type="checkbox"
            :width="column.width"
            :fixed="column.fixed"
          />
          <vxe-column
            v-else
            :field="column.field"
            :title="column.title"
            :width="column.width"
            :sortable="column.sortable"
            :fixed="column.fixed"
            :formatter="column.formatter"
          >
            <template v-if="column.slot === 'status'" #default="{ row }">
              <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'danger' : 'info'">
                {{ (row.Status ?? row.status ?? 1) === 1 ? $t('blacklist.enabled') : $t('blacklist.disabled') }}
              </el-tag>
            </template>
            <template v-else-if="column.slot === 'ip'" #default="{ row }">
              <div style="word-break: break-all;">
                {{ formatIP(row.IP || row.ip || '') }}
              </div>
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <el-button 
                type="primary" 
                link 
                :disabled="getButtonState('blacklist.update').disabled"
                @click="handleEdit(row)"
              >
                {{ $t('common.edit') }}
              </el-button>
              <el-button 
                type="danger" 
                link 
                :disabled="getButtonState('blacklist.destroy').disabled"
                @click="handleDelete(row)"
              >
                {{ $t('common.delete') }}
              </el-button>
            </template>
          </vxe-column>
        </template>
      </vxe-table>

      <Pagination
        v-model="pagination"
        @page-change="handlePageChange"
      />
    </el-card>

    <BlacklistForm
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import BlacklistForm from './BlacklistForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import {
  getBlacklistList,
  deleteBlacklist
} from '../../api/blacklist'

const { t } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)
const dialogVisible = ref(false)
const editId = ref(null)

// 字段名映射
const fieldMapping = {
  'id': 'id',
  'ip': 'ip',
  'remark': 'remark',
  'status': 'status',
  'created_at': 'created_at'
}

// 使用列表页面 composable
const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handlePageChange,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getBlacklistList,
  initialSearchForm: {
    ip: '',
    status: ''
  },
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  }
})

// 表格列配置
const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },
  {
    field: 'ip',
    title: t('blacklist.ip'),
    sortable: true,
    slot: 'ip'
  },
  {
    field: 'remark',
    title: t('blacklist.remark'),
    sortable: false,
    formatter: ({ row }) => row.Remark || row.remark || '-'
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 100,
    sortable: true,
    slot: 'status'
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    width: 180,
    sortable: true
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation',
    sortable: false
  }
])

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'ip',
    label: t('blacklist.ip'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'select',
    width: '150px',
    advanced: false,
    options: [
      { label: t('blacklist.enabled'), value: '1' },
      { label: t('blacklist.disabled'), value: '0' }
    ]
  }
])

// 格式化IP显示
const formatIP = (ip) => {
  if (!ip) return '-'
  // 如果是IP范围，格式化显示
  if (ip.includes('-')) {
    const parts = ip.split('-')
    if (parts.length === 2) {
      return `${parts[0].trim()} ~ ${parts[1].trim()}`
    }
  }
  return ip
}

const handleAdd = () => {
  editId.value = null
  dialogVisible.value = true
}

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleFormSuccess = () => {
  loadData()
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(t('blacklist.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteBlacklist(row.id)
    ElMessage.success(t('blacklist.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

onMounted(() => {
  initDefaultSort()
  loadData()
})
</script>

<style scoped>
.blacklist-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

