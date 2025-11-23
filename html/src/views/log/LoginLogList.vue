<template>
  <div class="log-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>登录日志</span>
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
        <vxe-column field="ip" title="IP地址" width="150" />
        <vxe-column field="user_agent" title="用户代理" />
        <vxe-column field="status" title="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </vxe-column>
        <vxe-column field="message" title="消息" />
        <vxe-column field="created_at" title="登录时间" width="180" />
        <vxe-column title="操作" width="100" fixed="right">
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
        <el-descriptions-item label="IP地址">{{ logDetail.ip }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="logDetail.status === 1 ? 'success' : 'danger'">
            {{ logDetail.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="用户代理">{{ logDetail.user_agent }}</el-descriptions-item>
        <el-descriptions-item label="登录时间">{{ logDetail.created_at }}</el-descriptions-item>
        <el-descriptions-item label="消息" :span="2">{{ logDetail.message }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getLoginLogList,
  getLoginLogDetail,
  deleteLoginLog,
  cleanLoginLogs
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
    const res = await getLoginLogList(params)
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load login log list error:', error)
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
    const res = await getLoginLogDetail(row.id)
    if (res.data && res.data.login_log) {
      logDetail.value = res.data.login_log
      detailVisible.value = true
    }
  } catch (error) {
    console.error('Load login log detail error:', error)
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该日志吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteLoginLog(row.id)
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
    await ElMessageBox.confirm('确定要清空所有登录日志吗？此操作不可恢复！', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await cleanLoginLogs()
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
</style>

