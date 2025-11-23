<template>
  <div class="dictionary-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>字典列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            添加字典
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="类型">
          <el-input v-model="searchForm.type" placeholder="请输入类型" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <vxe-table
        :data="tableData"
        :loading="loading"
        border
        resizable
        height="600"
      >
        <vxe-column type="seq" width="60" title="序号" />
        <vxe-column field="id" title="ID" width="80" />
        <vxe-column field="type" title="类型" />
        <vxe-column field="label" title="标签" />
        <vxe-column field="value" title="值" />
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

      <vxe-pager
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        @page-change="handlePageChange"
      />
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
        <el-form-item label="类型" prop="type">
          <el-input v-model="formData.type" />
        </el-form-item>
        <el-form-item label="标签" prop="label">
          <el-input v-model="formData.label" />
        </el-form-item>
        <el-form-item label="值" prop="value">
          <el-input v-model="formData.value" />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDictionaryList,
  getDictionaryDetail,
  createDictionary,
  updateDictionary,
  deleteDictionary
} from '../../api/dictionary'

const formRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加字典')

const searchForm = reactive({
  type: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])

const formData = reactive({
  id: null,
  type: '',
  label: '',
  value: '',
  status: 1,
  sort: 0
})

const formRules = {
  type: [{ required: true, message: '请输入类型', trigger: 'blur' }],
  label: [{ required: true, message: '请输入标签', trigger: 'blur' }],
  value: [{ required: true, message: '请输入值', trigger: 'blur' }]
}

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...searchForm
    }
    const res = await getDictionaryList(params)
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load dictionary list error:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.type = ''
  handleSearch()
}

const handlePageChange = ({ currentPage, pageSize }) => {
  pagination.page = currentPage
  pagination.pageSize = pageSize
  loadData()
}

const handleAdd = () => {
  dialogTitle.value = '添加字典'
  Object.assign(formData, {
    id: null,
    type: '',
    label: '',
    value: '',
    status: 1,
    sort: 0
  })
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  dialogTitle.value = '编辑字典'
  try {
    const res = await getDictionaryDetail(row.id)
    if (res.data && res.data.dictionary) {
      const dict = res.data.dictionary
      Object.assign(formData, {
        id: dict.id,
        type: dict.type,
        label: dict.label,
        value: dict.value,
        status: dict.status,
        sort: dict.sort || 0
      })
      dialogVisible.value = true
    }
  } catch (error) {
    console.error('Load dictionary detail error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (formData.id) {
          await updateDictionary(formData.id, formData)
          ElMessage.success('更新成功')
        } else {
          await createDictionary(formData)
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
    await ElMessageBox.confirm('确定要删除该字典吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteDictionary(row.id)
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
.dictionary-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 20px;
}
</style>

