<template>
  <div class="department-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>部门列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            添加部门
          </el-button>
        </div>
      </template>

      <vxe-table
        :data="tableData"
        :loading="loading"
        border
        resizable
        tree-config
        height="600"
      >
        <vxe-column type="seq" width="60" title="序号" />
        <vxe-column field="name" title="部门名称" tree-node />
        <vxe-column field="description" title="描述" />
        <vxe-column field="sort" title="排序" width="80" />
        <vxe-column field="status" title="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="created_at" title="创建时间" />
        <vxe-column title="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
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
        <el-form-item label="父部门">
          <el-select v-model="formData.parent_id" placeholder="请选择父部门" clearable>
            <el-option label="顶级部门" :value="0" />
            <el-option
              v-for="dept in departmentOptions"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="formData.description" type="textarea" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="formData.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDepartmentList,
  getDepartmentDetail,
  createDepartment,
  updateDepartment,
  deleteDepartment
} from '../../api/department'

const formRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加部门')

const tableData = ref([])

const formData = reactive({
  id: null,
  parent_id: 0,
  name: '',
  description: '',
  status: 1,
  sort: 0
})

const formRules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }]
}

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
    const res = await getDepartmentList()
    if (res.data && res.data.list) {
      tableData.value = res.data.list
    }
  } catch (error) {
    console.error('Load department list error:', error)
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  dialogTitle.value = '添加部门'
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
  dialogTitle.value = '编辑部门'
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
          ElMessage.success('更新成功')
        } else {
          await createDepartment(data)
          ElMessage.success('创建成功')
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
    await ElMessageBox.confirm('确定要删除该部门吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteDepartment(row.id)
    ElMessage.success('删除成功')
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

