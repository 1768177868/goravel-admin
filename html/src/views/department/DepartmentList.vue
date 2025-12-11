<template>
  <div class="department-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('department.title') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('department.store').disabled"
            @click="handleAdd"
          >
            <el-icon><Plus /></el-icon>
            {{ $t('department.add_department') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ name: '', status: '' }"
        i18n-prefix="department"
        @search="handleSearch"
        @reset="handleReset"
      />

      <vxe-table
        :data="tableData"
        :loading="loading"
        border
        :column-config="{ resizable: true }"
        :tree-config="hasSearch ? false : { childrenField: 'children', expandAll: false, indent: 20 }"
        height="600"
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
            :fixed="column.fixed"
            :formatter="column.formatter"
            :tree-node="column.treeNode"
          >
            <template v-if="column.slot === 'status'" #default="{ row }">
              <el-tag :type="(row.Status ?? row.status ?? 1) === 1 ? 'success' : 'danger'">
                {{ (row.Status ?? row.status ?? 1) === 1 ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
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
          </vxe-column>
        </template>
      </vxe-table>
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
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import DepartmentForm from './DepartmentForm.vue'
import { usePermission } from '../../composables/usePermission'
import {
  getDepartmentList,
  deleteDepartment
} from '../../api/department'

const { t } = useI18n()
const { getButtonState } = usePermission()
const loading = ref(false)
const dialogVisible = ref(false)
const editId = ref(null)

const tableData = ref([])
const hasSearch = ref(false) // 标记是否有搜索条件

const searchForm = reactive({
  name: '',
  status: ''
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

// 表格列配置（使用 vxe-table columns）
// 树形结构不需要排序功能
const tableColumns = computed(() => [
  {
    field: 'name',
    title: t('department.name'),
    treeNode: true,
    formatter: ({ row }) => row.Name || row.name || '-'
  },
  {
    field: 'remark',
    title: t('common.description'),
    formatter: ({ row }) => row.Remark || row.remark || row.description || '-'
  },
  {
    field: 'sort',
    title: t('common.sort'),
    width: 80,
    formatter: ({ row }) => row.Sort !== undefined ? row.Sort : (row.sort !== undefined ? row.sort : 0)
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 80,
    slot: 'status'
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    formatter: ({ row }) => row.created_at || row.CreatedAt || '-'
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation'
  }
])

const departmentOptions = computed(() => {
  const flatten = (departments, parentId = 0) => {
    const result = []
    departments.forEach(dept => {
      const deptParentId = dept.parent_id !== undefined ? dept.parent_id : (dept.ParentID !== undefined ? dept.ParentID : 0)
      if (deptParentId === parentId) {
        result.push({
          id: dept.id,
          name: dept.name || dept.Name || ''
        })
        const children = flatten(departments, dept.id)
        result.push(...children)
      }
    })
    return result
  }
  return flatten(tableData.value)
})

// 转换后端数据格式为前端格式
const transformDepartmentData = (dept) => {
  const children = dept.Children || dept.children
  let transformedChildren = []
  
  if (children && Array.isArray(children) && children.length > 0) {
    transformedChildren = children.map(child => transformDepartmentData(child))
  }
  
  const result = {
    id: dept.id,
    parent_id: dept.ParentID !== undefined ? dept.ParentID : (dept.parent_id !== undefined ? dept.parent_id : 0),
    name: dept.Name || dept.name || '',
    remark: dept.Remark || dept.remark || dept.description || '',
    description: dept.Remark || dept.remark || dept.description || '', // 兼容字段
    status: dept.Status !== undefined ? dept.Status : (dept.status !== undefined ? dept.status : 1),
    sort: dept.Sort !== undefined ? dept.Sort : (dept.sort !== undefined ? dept.sort : 0),
    created_at: dept.created_at || dept.CreatedAt || ''
  }
  
  if (transformedChildren.length > 0) {
    result.children = transformedChildren
  }
  
  return result
}

const loadData = async () => {
  loading.value = true
  try {
    const params = {}
    // 检查是否有搜索条件
    if (searchForm.name || searchForm.status) {
      hasSearch.value = true
      if (searchForm.name && searchForm.name.trim()) {
        params.name = searchForm.name.trim()
      }
      if (searchForm.status) {
        params.status = searchForm.status
      }
    } else {
      hasSearch.value = false
    }
    
    const res = await getDepartmentList(params)
    
    if (res.data && res.data.list) {
      const transformed = res.data.list.map(dept => transformDepartmentData(dept))
      tableData.value = transformed
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
  Object.keys(searchForm).forEach(key => {
    searchForm[key] = ''
  })
  loadData()
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
    await ElMessageBox.confirm(t('department.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteDepartment(row.id)
    ElMessage.success(t('department.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.department-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

