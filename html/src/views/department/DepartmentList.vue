<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.department') }}</span>
          <div class="header-actions">
            <el-button 
              v-if="!hasSearch"
              @click="handleToggleExpand"
            >
              <el-icon><component :is="isExpanded ? 'Fold' : 'Expand'" /></el-icon>
              {{ isExpanded ? $t('menu_management.collapse_all') : $t('menu_management.expand_all') }}
            </el-button>
            <el-button 
              type="primary" 
              :disabled="getButtonState('department.store').disabled"
              @click="handleAdd"
            >
              <el-icon><Plus /></el-icon>
              {{ $t('department.add_department') }}
            </el-button>
          </div>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="department"
        @search="handleSearch"
        @reset="handleReset"
      />

      <el-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="isExpanded"
        style="width: 100%"
        height="600"
      >
        <el-table-column type="index" width="60" :label="$t('table.seq')" />
        <el-table-column prop="name" :label="$t('department.name')" min-width="150" />
        <el-table-column prop="remark" :label="$t('common.description')" min-width="200">
          <template #default="{ row }">
            {{ row.remark || row.description || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="sort" :label="$t('common.sort')" width="100" />
        <el-table-column prop="status" :label="$t('table.status')" width="200">
          <template #default="{ row }">
            <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'success' : 'danger'">
              {{ (row.Status ?? row.status ?? 1) === 1 ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <!-- <el-table-column prop="created_at" :label="$t('table.created_at')" width="180" /> -->
        <el-table-column :label="$t('table.operation')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button 
              type="primary" 
              link 
              :disabled="getButtonState('department.update').disabled"
              @click="handleEdit(row)"
            >
              {{ $t('common.edit') }}
            </el-button>
            <el-button 
              type="danger" 
              link 
              :disabled="getButtonState('department.destroy').disabled"
              @click="handleDelete(row)"
            >
              {{ $t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <DepartmentForm
      v-model="dialogVisible"
      :edit-id="editId"
      :department-options="departmentOptions"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Fold, Expand } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import DepartmentForm from './DepartmentForm.vue'
import { buildSearchParams } from '../../utils/buildSearchParams'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import { flattenTree } from '../../utils/tree'
import {
  getDepartmentList,
  deleteDepartment
} from '../../api/department'

const { t } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)
const loading = ref(false)
const isExpanded = ref(false)

const {
  dialogVisible,
  editId,
  handleAdd,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: deleteDepartment
})

const tableData = ref([])

// 初始搜索表单
const initialSearchForm = {
  name: '',
  status: ''
}

const searchForm = reactive({
  name: '',
  status: ''
})

const hasSearch = computed(() => {
  return !!(searchForm.name || searchForm.status)
})

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'name',
    label: t('department.name'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'select',
    width: '120px',
    options: [
      { label: t('common.enabled'), value: '1' },
      { label: t('common.disabled'), value: '0' }
    ],
    advanced: false
  }
])

// 部门选项（保持树形结构，用于表单选择）
const departmentOptions = computed(() => {
  return tableData.value || []
})

const loadData = async () => {
  loading.value = true
  try {
    const params = buildSearchParams(searchForm)
    const res = await getDepartmentList(params)
    
    if (res.data && res.data.list) {
      // 后端已返回前端可直接使用的树形结构，无需转换
      tableData.value = res.data.list
    } else {
      tableData.value = []
    }
  } catch (error) {
    console.error('Load department list error:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadData()
}

const handleReset = () => {
  searchForm.name = ''
  searchForm.status = ''
  loadData()
}

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleFormSuccess = () => {
  loadData()
}

const handleDelete = (row) => handleDeleteCrud(row, loadData)

const handleRefresh = () => {
  loadData()
}

const handleToggleExpand = () => {
  isExpanded.value = !isExpanded.value
  
  if (tableRef.value) {
    // Element Plus 的 el-table 使用 toggleRowExpansion 方法
    // 递归处理所有节点
    const toggleNode = (rows) => {
      if (Array.isArray(rows)) {
        rows.forEach(row => {
          // 切换当前节点的展开状态
          tableRef.value.toggleRowExpansion(row, isExpanded.value)
          
          // 如果有子节点，递归处理
          if (row.children && row.children.length > 0) {
            toggleNode(row.children)
          }
        })
      }
    }
    
    // 处理所有顶级节点
    toggleNode(tableData.value)
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>

</style>
