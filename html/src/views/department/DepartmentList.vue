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
        <vxe-column type="seq" width="60" :title="$t('table.seq')" />
        <vxe-column field="name" :title="$t('department.name')" tree-node />
        <vxe-column field="description" :title="$t('common.description')" />
        <vxe-column field="sort" :title="$t('common.sort')" width="80" />
        <vxe-column field="status" :title="$t('table.status')" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? $t('common.enabled') : $t('common.disabled') }}
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
      if (dept.parent_id === parentId) {
        result.push(dept)
        const children = flatten(departments, dept.id)
        result.push(...children)
      }
    })
    return result
  }
  return flatten(tableData.value)
})

const loadData = async () => {
  loading.value = true
  try {
    const params = {}
    // 检查是否有搜索条件
    if (searchForm.name || searchForm.status) {
      hasSearch.value = true
      if (searchForm.name) params.name = searchForm.name
      if (searchForm.status) params.status = searchForm.status
    } else {
      hasSearch.value = false
    }
    
    const res = await getDepartmentList(params)
    if (res.data && res.data.list) {
      tableData.value = res.data.list
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
      Object.assign(formData, {
        id: dept.id,
        parent_id: dept.parent_id || 0,
        name: dept.name,
        description: dept.description || '',
        status: dept.status,
        sort: dept.sort || 0
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
        const data = { ...formData }
        if (data.parent_id === 0) {
          data.parent_id = null
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

