<template>
  <div class="log-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>操作日志</span>
          <el-button type="danger" @click="handleClean">
            <el-icon><Delete /></el-icon>
            清空日志
          </el-button>
        </div>
      </template>

      <vxe-table
        :data="tableData"
        :loading="loading"
        border
        resizable
        height="600"
      >
        <vxe-column type="seq" width="60" title="序号" />
        <vxe-column field="id" title="ID" width="80" />
        <vxe-column field="admin.username" title="管理员" />
        <vxe-column field="method" title="请求方法" width="100" />
        <vxe-column field="path" title="请求路径" />
        <vxe-column field="ip" title="IP地址" width="150" />
        <vxe-column field="status_code" title="状态码" width="100" />
        <vxe-column field="created_at" title="操作时间" width="180" />
        <vxe-column title="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleView(row)">查看</el-button>
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

    <el-dialog v-model="detailVisible" title="日志详情" width="800px">
      <el-descriptions :column="2" border v-if="logDetail">
        <el-descriptions-item label="ID">{{ logDetail.id }}</el-descriptions-item>
        <el-descriptions-item label="管理员">{{ logDetail.admin?.username }}</el-descriptions-item>
        <el-descriptions-item label="请求方法">{{ logDetail.method }}</el-descriptions-item>
        <el-descriptions-item label="请求路径">{{ logDetail.path }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ logDetail.ip }}</el-descriptions-item>
        <el-descriptions-item label="状态码">{{ logDetail.status_code }}</el-descriptions-item>
        <el-descriptions-item label="操作时间" :span="2">{{ logDetail.created_at }}</el-descriptions-item>
        <el-descriptions-item label="请求参数" :span="2">
          <pre>{{ JSON.stringify(logDetail.params, null, 2) }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getOperationLogList,
  getOperationLogDetail,
  deleteOperationLog,
  cleanOperationLogs
} from '../../api/log'

const loading = ref(false)
const detailVisible = ref(false)
const logDetail = ref(null)

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    const res = await getOperationLogList(params)
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load operation log list error:', error)
  } finally {
    loading.value = false
  }
}

const handlePageChange = ({ currentPage, pageSize }) => {
  pagination.page = currentPage
  pagination.pageSize = pageSize
  loadData()
}

const handleView = async (row) => {
  try {
    const res = await getOperationLogDetail(row.id)
    if (res.data && res.data.operation_log) {
      logDetail.value = res.data.operation_log
      detailVisible.value = true
    }
  } catch (error) {
    console.error('Load operation log detail error:', error)
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该日志吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteOperationLog(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

const handleClean = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有操作日志吗？此操作不可恢复！', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await cleanOperationLogs()
    ElMessage.success('清空成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Clean error:', error)
    }
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.log-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

pre {
  margin: 0;
  padding: 10px;
  background: #f5f5f5;
  border-radius: 4px;
  overflow-x: auto;
}
</style>

