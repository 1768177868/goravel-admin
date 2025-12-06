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
            :sortable="column.sortable"
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

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item :label="$t('department.parent_department')">
          <el-select v-model="formData.parent_id" :placeholder="$t('form.select_parent') + $t('department.parent_department')" clearable>
            <el-option :label="$t('department.top_department')" :value="0" />
            <el-option
              v-for="dept in departmentOptions"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('department.name')" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item :label="$t('common.description')">
          <el-input v-model="formData.description" type="textarea" />
        </el-form-item>
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :label="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('common.sort')">
          <el-input-number v-model="formData.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import { usePermission } from '../../composables/usePermission'
import {
  getDepartmentList,
  getDepartmentDetail,
  createDepartment,
  updateDepartment,
  deleteDepartment
} from '../../api/department'

const { t } = useI18n()
const { getButtonState } = usePermission()
const formRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('department.edit_department') : t('department.add_department'))

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
const tableColumns = computed(() => [
  {
    field: 'name',
    title: t('department.name'),
    treeNode: true,
    sortable: true,
    formatter: ({ row }) => row.Name || row.name || '-'
  },
  {
    field: 'remark',
    title: t('common.description'),
    sortable: false,
    formatter: ({ row }) => row.Remark || row.remark || row.description || '-'
  },
  {
    field: 'sort',
    title: t('common.sort'),
    width: 80,
    sortable: true,
    formatter: ({ row }) => row.Sort !== undefined ? row.Sort : (row.sort !== undefined ? row.sort : 0)
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 80,
    sortable: true,
    slot: 'status'
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    sortable: true,
    formatter: ({ row }) => row.created_at || row.CreatedAt || '-'
  },
  {
    title: t('table.operation'),
    width: 150,
    fixed: 'right',
    slot: 'operation',
    sortable: false
  }
])

const formData = reactive({
  id: null,
  parent_id: 0,
  name: '',
  description: '',
  status: 1,
  sort: 0
})

const formRules = computed(() => ({
  name: [{ required: true, message: t('department.name_required'), trigger: 'blur' }]
}))

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
  Object.assign(formData, {
    id: null,
    parent_id: 0,
    name: '',
    description: '',
    status: 1,
    sort: 0
  })
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  try {
    const res = await getDepartmentDetail(row.id)
    if (res.data && res.data.department) {
      const dept = res.data.department
      // 后端返回的是 PascalCase 字段，需要正确映射
      Object.assign(formData, {
        id: dept.id,
        parent_id: dept.ParentID !== undefined ? dept.ParentID : (dept.parent_id || 0),
        name: dept.Name || dept.name || '',
        description: dept.Remark || dept.remark || dept.description || '',
        status: dept.Status !== undefined ? dept.Status : (dept.status !== undefined ? dept.status : 1),
        sort: dept.Sort !== undefined ? dept.Sort : (dept.sort !== undefined ? dept.sort : 0)
      })
      dialogVisible.value = true
    }
  } catch (error) {
    console.error('Load department detail error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        // 转换前端字段名为后端期望的字段名
        const data = {
          name: formData.name,
          remark: formData.description, // description 映射到 remark
          status: formData.status,
          sort: formData.sort,
          parent_id: formData.parent_id === 0 ? null : formData.parent_id
        }
        
        if (formData.id) {
          await updateDepartment(formData.id, data)
          ElMessage.success(t('department.update_success'))
        } else {
          await createDepartment(data)
          ElMessage.success(t('department.create_success'))
        }
        dialogVisible.value = false
        loadData()
      } catch (error) {
        console.error('Submit error:', error)
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
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

