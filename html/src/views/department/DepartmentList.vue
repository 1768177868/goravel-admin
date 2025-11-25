<template>
  <div class="department-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('department.title') }}</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            {{ $t('department.add_department') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <el-form :model="searchForm" :inline="true" class="search-form">
        <el-form-item :label="$t('department.name')">
          <el-input
            v-model="searchForm.name"
            :placeholder="$t('form.please_enter') + $t('department.name')"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item :label="$t('table.status')">
          <el-select
            v-model="searchForm.status"
            :placeholder="$t('form.please_select') + $t('table.status')"
            clearable
            style="width: 120px"
          >
            <el-option :label="$t('common.enabled')" value="1" />
            <el-option :label="$t('common.disabled')" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            {{ $t('log.search') }}
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            {{ $t('log.reset') }}
          </el-button>
        </el-form-item>
      </el-form>

      <vxe-table
        :data="tableData"
        :loading="loading"
        border
        resizable
        :tree-config="hasSearch ? false : { children: 'children', expandAll: false, indent: 20 }"
        height="600"
      >
        <vxe-column field="name" :title="$t('department.name')" tree-node>
          <template #default="{ row }">
            {{ row.Name || row.name || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="remark" :title="$t('common.description')">
          <template #default="{ row }">
            {{ row.Remark || row.remark || row.description || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="sort" :title="$t('common.sort')" width="80">
          <template #default="{ row }">
            {{ row.Sort !== undefined ? row.Sort : (row.sort !== undefined ? row.sort : 0) }}
          </template>
        </vxe-column>
        <vxe-column field="status" :title="$t('table.status')" width="80">
          <template #default="{ row }">
            <el-tag :type="(row.Status !== undefined ? row.Status : (row.status !== undefined ? row.status : 1)) === 1 ? 'success' : 'danger'">
              {{ (row.Status !== undefined ? row.Status : (row.status !== undefined ? row.status : 1)) === 1 ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="created_at" :title="$t('table.created_at')" />
        <vxe-column :title="$t('table.operation')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
            <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </vxe-column>
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
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import {
  getDepartmentList,
  getDepartmentDetail,
  createDepartment,
  updateDepartment,
  deleteDepartment
} from '../../api/department'

const { t } = useI18n()
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
    
    console.log('Department search params:', params)
    const res = await getDepartmentList(params)
    console.log('Department list response:', res)
    
    if (res.data && res.data.list) {
      // 转换数据格式，支持 PascalCase 和 snake_case
      const transformed = res.data.list.map(dept => transformDepartmentData(dept))
      console.log('Transformed department data:', transformed)
      tableData.value = transformed
    } else {
      console.warn('No department data in response:', res)
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

