<template>
  <div class="monitor-page">
    <el-row :gutter="20">
      <!-- CPU信息 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>{{ $t('monitor.cpu') }}</span>
              <el-button size="small" @click="loadData">
                {{ $t('tabs.refresh') }}
              </el-button>
            </div>
          </template>
          <div class="monitor-content">
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.cpu_usage') }}:</span>
              <el-progress
                :percentage="systemInfo.cpu?.percent || 0"
                :color="getProgressColor(systemInfo.cpu?.percent || 0)"
                :stroke-width="20"
              />
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.cpu_model') }}:</span>
              <span class="value">{{ systemInfo.cpu?.model || '-' }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.cpu_cores') }}:</span>
              <span class="value">{{ systemInfo.cpu?.cores || 0 }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 内存信息 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>{{ $t('monitor.memory') }}</span>
              <el-button size="small" @click="loadData">
                {{ $t('tabs.refresh') }}
              </el-button>
            </div>
          </template>
          <div class="monitor-content">
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.memory_usage') }}:</span>
              <el-progress
                :percentage="systemInfo.memory?.percent || 0"
                :color="getProgressColor(systemInfo.memory?.percent || 0)"
                :stroke-width="20"
              />
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.memory_total') }}:</span>
              <span class="value">{{ formatBytes(systemInfo.memory?.total || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.memory_used') }}:</span>
              <span class="value">{{ formatBytes(systemInfo.memory?.used || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.memory_available') }}:</span>
              <span class="value">{{ formatBytes(systemInfo.memory?.available || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.memory_free') }}:</span>
              <span class="value">{{ formatBytes(systemInfo.memory?.free || 0) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <!-- 磁盘信息 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>{{ $t('monitor.disk') }}</span>
              <el-button size="small" @click="loadData">
                {{ $t('tabs.refresh') }}
              </el-button>
            </div>
          </template>
          <div class="monitor-content">
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.disk_usage') }}:</span>
              <el-progress
                :percentage="systemInfo.disk?.percent || 0"
                :color="getProgressColor(systemInfo.disk?.percent || 0)"
                :stroke-width="20"
              />
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.disk_total') }}:</span>
              <span class="value">{{ formatBytes(systemInfo.disk?.total || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.disk_used') }}:</span>
              <span class="value">{{ formatBytes(systemInfo.disk?.used || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.disk_free') }}:</span>
              <span class="value">{{ formatBytes(systemInfo.disk?.free || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.disk_fstype') }}:</span>
              <span class="value">{{ systemInfo.disk?.fstype || '-' }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.disk_path') }}:</span>
              <span class="value">{{ systemInfo.disk?.path || '-' }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 网络信息 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>{{ $t('monitor.network') }}</span>
              <el-button size="small" @click="loadData">
                {{ $t('tabs.refresh') }}
              </el-button>
            </div>
          </template>
          <div class="monitor-content">
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.net_bytes_sent') }}:</span>
              <span class="value">{{ formatBytes(systemInfo.net?.bytes_sent || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.net_bytes_recv') }}:</span>
              <span class="value">{{ formatBytes(systemInfo.net?.bytes_recv || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.net_packets_sent') }}:</span>
              <span class="value">{{ formatNumber(systemInfo.net?.packets_sent || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.net_packets_recv') }}:</span>
              <span class="value">{{ formatNumber(systemInfo.net?.packets_recv || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.net_errin') }}:</span>
              <span class="value">{{ formatNumber(systemInfo.net?.errin || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.net_errout') }}:</span>
              <span class="value">{{ formatNumber(systemInfo.net?.errout || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.net_dropin') }}:</span>
              <span class="value">{{ formatNumber(systemInfo.net?.dropin || 0) }}</span>
            </div>
            <div class="monitor-item">
              <span class="label">{{ $t('monitor.net_dropout') }}:</span>
              <span class="value">{{ formatNumber(systemInfo.net?.dropout || 0) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getSystemInfo } from '../../api/monitor'

const { t } = useI18n()

const systemInfo = ref({
  cpu: {},
  memory: {},
  disk: {},
  net: {}
})

const loading = ref(false)
let refreshTimer = null

const loadData = async () => {
  loading.value = true
  try {
    const { data } = await getSystemInfo()
    systemInfo.value = data || {}
  } catch (error) {
    console.error('Load system info error:', error)
    ElMessage.error(t('error.default'))
  } finally {
    loading.value = false
  }
}

const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

const formatNumber = (num) => {
  return num.toLocaleString()
}

const getProgressColor = (percentage) => {
  if (percentage < 50) {
    return '#67c23a'
  } else if (percentage < 80) {
    return '#e6a23c'
  } else {
    return '#f56c6c'
  }
}

onMounted(() => {
  loadData()
  // 每30秒自动刷新
  refreshTimer = setInterval(() => {
    loadData()
  }, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.monitor-page {
  padding: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
}

.monitor-content {
  padding: 10px 0;
}

.monitor-item {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
  gap: 12px;
}

.monitor-item:last-child {
  margin-bottom: 0;
}

.monitor-item .label {
  min-width: 120px;
  font-weight: 500;
  color: #606266;
}

.monitor-item .value {
  flex: 1;
  color: #303133;
  word-break: break-all;
}

.monitor-item .el-progress {
  flex: 1;
}
</style>

