<template>
  <div class="monitor-page">
    <el-row :gutter="20">
      <!-- CPU信息 -->
      <el-col :span="12">
        <el-card class="monitor-card cpu-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <el-icon class="card-icon cpu-icon"><Cpu /></el-icon>
                <span>{{ $t('monitor.cpu') }}</span>
              </div>
              <el-button size="small" :icon="Refresh" circle @click="loadData" />
            </div>
          </template>
          <div class="monitor-content">
            <div class="monitor-item usage-item">
              <div class="usage-header">
                <span class="label">{{ $t('monitor.cpu_usage') }}</span>
                <span class="percent-value">{{ formatPercent(systemInfo.cpu?.percent || 0) }}</span>
              </div>
              <el-progress
                :percentage="formatPercentForProgress(systemInfo.cpu?.percent || 0)"
                :color="getProgressColor(systemInfo.cpu?.percent || 0)"
                :stroke-width="24"
                :show-text="false"
                class="usage-progress"
              />
            </div>
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.cpu_model') }}</span>
                <span class="info-value">{{ systemInfo.cpu?.model || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.cpu_cores') }}</span>
                <span class="info-value highlight">{{ systemInfo.cpu?.cores || 0 }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 内存信息 -->
      <el-col :span="12">
        <el-card class="monitor-card memory-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <el-icon class="card-icon memory-icon"><DataBoard /></el-icon>
                <span>{{ $t('monitor.memory') }}</span>
              </div>
              <el-button size="small" :icon="Refresh" circle @click="loadData" />
            </div>
          </template>
          <div class="monitor-content">
            <div class="monitor-item usage-item">
              <div class="usage-header">
                <span class="label">{{ $t('monitor.memory_usage') }}</span>
                <span class="percent-value">{{ formatPercent(systemInfo.memory?.percent || 0) }}</span>
              </div>
              <el-progress
                :percentage="formatPercentForProgress(systemInfo.memory?.percent || 0)"
                :color="getProgressColor(systemInfo.memory?.percent || 0)"
                :stroke-width="24"
                :show-text="false"
                class="usage-progress"
              />
            </div>
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.memory_total') }}</span>
                <span class="info-value">{{ formatBytes(systemInfo.memory?.total || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.memory_used') }}</span>
                <span class="info-value highlight">{{ formatBytes(systemInfo.memory?.used || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.memory_available') }}</span>
                <span class="info-value">{{ formatBytes(systemInfo.memory?.available || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.memory_free') }}</span>
                <span class="info-value">{{ formatBytes(systemInfo.memory?.free || 0) }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <!-- 磁盘信息 -->
      <el-col :span="12">
        <el-card class="monitor-card disk-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <el-icon class="card-icon disk-icon"><FolderOpened /></el-icon>
                <span>{{ $t('monitor.disk') }}</span>
              </div>
              <el-button size="small" :icon="Refresh" circle @click="loadData" />
            </div>
          </template>
          <div class="monitor-content">
            <div class="monitor-item usage-item">
              <div class="usage-header">
                <span class="label">{{ $t('monitor.disk_usage') }}</span>
                <span class="percent-value">{{ formatPercent(systemInfo.disk?.percent || 0) }}</span>
              </div>
              <el-progress
                :percentage="formatPercentForProgress(systemInfo.disk?.percent || 0)"
                :color="getProgressColor(systemInfo.disk?.percent || 0)"
                :stroke-width="24"
                :show-text="false"
                class="usage-progress"
              />
            </div>
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.disk_total') }}</span>
                <span class="info-value">{{ formatBytes(systemInfo.disk?.total || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.disk_used') }}</span>
                <span class="info-value highlight">{{ formatBytes(systemInfo.disk?.used || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.disk_free') }}</span>
                <span class="info-value">{{ formatBytes(systemInfo.disk?.free || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.disk_fstype') }}</span>
                <span class="info-value">{{ systemInfo.disk?.fstype || '-' }}</span>
              </div>
              <div class="info-item full-width">
                <span class="info-label">{{ $t('monitor.disk_path') }}</span>
                <span class="info-value">{{ systemInfo.disk?.path || '-' }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 网络信息汇总 -->
      <el-col :span="12">
        <el-card class="monitor-card network-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <el-icon class="card-icon network-icon"><Connection /></el-icon>
                <span>{{ $t('monitor.network') }}</span>
              </div>
              <el-button size="small" :icon="Refresh" circle @click="loadData" />
            </div>
          </template>
          <div class="monitor-content">
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.net_bytes_sent') }}</span>
                <span class="info-value highlight">{{ formatBytes(systemInfo.net?.bytes_sent || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.net_bytes_recv') }}</span>
                <span class="info-value highlight">{{ formatBytes(systemInfo.net?.bytes_recv || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.net_packets_sent') }}</span>
                <span class="info-value">{{ formatNumber(systemInfo.net?.packets_sent || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.net_packets_recv') }}</span>
                <span class="info-value">{{ formatNumber(systemInfo.net?.packets_recv || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.net_errin') }}</span>
                <span class="info-value error">{{ formatNumber(systemInfo.net?.errin || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.net_errout') }}</span>
                <span class="info-value error">{{ formatNumber(systemInfo.net?.errout || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.net_dropin') }}</span>
                <span class="info-value error">{{ formatNumber(systemInfo.net?.dropin || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.net_dropout') }}</span>
                <span class="info-value error">{{ formatNumber(systemInfo.net?.dropout || 0) }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 网卡流量详情 -->
    <el-row v-if="systemInfo.net?.interfaces && systemInfo.net.interfaces.length > 0" :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card class="monitor-card interfaces-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <el-icon class="card-icon interfaces-icon"><Monitor /></el-icon>
                <span>{{ $t('monitor.network_interfaces') }}</span>
              </div>
              <el-button size="small" :icon="Refresh" circle @click="loadData" />
            </div>
          </template>
          <div class="monitor-content">
            <el-table 
              :data="systemInfo.net.interfaces" 
              border 
              style="width: 100%"
              class="interfaces-table"
              stripe
            >
              <el-table-column :label="$t('monitor.interface_name')" prop="name" width="150" />
              <el-table-column :label="$t('monitor.interface_bytes_sent')" width="150">
                <template #default="{ row }">
                  {{ formatBytes(row.bytes_sent || 0) }}
                </template>
              </el-table-column>
              <el-table-column :label="$t('monitor.interface_bytes_recv')" width="150">
                <template #default="{ row }">
                  {{ formatBytes(row.bytes_recv || 0) }}
                </template>
              </el-table-column>
              <el-table-column :label="$t('monitor.interface_packets_sent')" width="150">
                <template #default="{ row }">
                  {{ formatNumber(row.packets_sent || 0) }}
                </template>
              </el-table-column>
              <el-table-column :label="$t('monitor.interface_packets_recv')" width="150">
                <template #default="{ row }">
                  {{ formatNumber(row.packets_recv || 0) }}
                </template>
              </el-table-column>
              <el-table-column :label="$t('monitor.interface_errin')" width="120">
                <template #default="{ row }">
                  {{ formatNumber(row.errin || 0) }}
                </template>
              </el-table-column>
              <el-table-column :label="$t('monitor.interface_errout')" width="120">
                <template #default="{ row }">
                  {{ formatNumber(row.errout || 0) }}
                </template>
              </el-table-column>
              <el-table-column :label="$t('monitor.interface_dropin')" width="120">
                <template #default="{ row }">
                  {{ formatNumber(row.dropin || 0) }}
                </template>
              </el-table-column>
              <el-table-column :label="$t('monitor.interface_dropout')" width="120">
                <template #default="{ row }">
                  {{ formatNumber(row.dropout || 0) }}
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 只在Linux系统上显示负载和文件描述符 -->
    <el-row v-if="isLinux" :gutter="20" style="margin-top: 20px">
      <!-- 负载信息 -->
      <el-col :span="12">
        <el-card class="monitor-card load-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <el-icon class="card-icon load-icon"><TrendCharts /></el-icon>
                <span>{{ $t('monitor.load') }}</span>
              </div>
              <el-button size="small" :icon="Refresh" circle @click="loadData" />
            </div>
          </template>
          <div class="monitor-content">
            <div class="load-display">
              <div class="load-value">
                <span class="load-number">{{ formatLoad(systemInfo.load?.load1 || 0) }}</span>
                <span class="load-percent">({{ formatPercent(systemInfo.load?.load1_percent || 0) }})</span>
              </div>
              <div class="load-label">{{ $t('monitor.load_current') }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 文件描述符信息 -->
      <el-col :span="12">
        <el-card class="monitor-card fd-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <el-icon class="card-icon fd-icon"><Document /></el-icon>
                <span>{{ $t('monitor.file_descriptors') }}</span>
              </div>
              <el-button size="small" :icon="Refresh" circle @click="loadData" />
            </div>
          </template>
          <div class="monitor-content">
            <div class="monitor-item usage-item">
              <div class="usage-header">
                <span class="label">{{ $t('monitor.fd_usage') }}</span>
                <span class="percent-value">{{ formatPercent(systemInfo.file_descriptors?.percent || 0) }}</span>
              </div>
              <el-progress
                :percentage="formatPercentForProgress(systemInfo.file_descriptors?.percent || 0)"
                :color="getProgressColor(systemInfo.file_descriptors?.percent || 0)"
                :stroke-width="24"
                :show-text="false"
                class="usage-progress"
              />
            </div>
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.fd_used') }}</span>
                <span class="info-value highlight">{{ formatNumber(systemInfo.file_descriptors?.used || 0) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">{{ $t('monitor.fd_free') }}</span>
                <span class="info-value">{{ formatNumber(systemInfo.file_descriptors?.free || 0) }}</span>
              </div>
              <div class="info-item full-width">
                <span class="info-label">{{ $t('monitor.fd_max') }}</span>
                <span class="info-value">{{ formatNumber(systemInfo.file_descriptors?.max || 0) }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { 
  Cpu, 
  DataBoard, 
  FolderOpened, 
  Connection, 
  Monitor, 
  TrendCharts, 
  Document, 
  Refresh 
} from '@element-plus/icons-vue'
import { getSystemInfo } from '../../api/monitor'

const { t } = useI18n()

const systemInfo = ref({
  os: 'linux',
  cpu: {},
  memory: {},
  disk: {},
  net: {},
  load: {},
  file_descriptors: {}
})

// 判断是否为Linux系统
const isLinux = computed(() => {
  return systemInfo.value.os === 'linux'
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

// 格式化百分比：根据值的大小决定保留的小数位数
// 优化显示：避免过多小数位
const formatPercent = (percent) => {
  if (percent === 0 || percent === null || percent === undefined) return '0%'
  
  // 如果百分比大于等于100，保留0位小数
  if (percent >= 100) {
    return Math.round(percent) + '%'
  }
  // 如果百分比大于等于1，保留1位小数
  if (percent >= 1) {
    return percent.toFixed(1) + '%'
  }
  // 如果百分比小于1，保留2位小数
  return percent.toFixed(2) + '%'
}

// 格式化百分比用于进度条（保留2位小数）
const formatPercentForProgress = (percent) => {
  return Math.round(percent * 100) / 100
}

// 格式化负载值
const formatLoad = (load) => {
  if (load === 0) return '0.00'
  return load.toFixed(2)
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

<style scoped lang="scss">
.monitor-page {
  padding: 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: calc(100vh - 60px);
}

.monitor-card {
  border-radius: 12px;
  border: none;
  transition: all 0.3s ease;
  overflow: hidden;
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 24px rgba(0, 0, 0, 0.15) !important;
  }

  :deep(.el-card__header) {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 16px 20px;
    border-bottom: none;
  }

  :deep(.el-card__body) {
    padding: 20px;
    background: white;
  }
}

.cpu-card :deep(.el-card__header) {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.memory-card :deep(.el-card__header) {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.disk-card :deep(.el-card__header) {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.network-card :deep(.el-card__header) {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
}

.interfaces-card :deep(.el-card__header) {
  background: linear-gradient(135deg, #30cfd0 0%, #330867 100%);
}

.load-card :deep(.el-card__header) {
  background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%);
}

.fd-card :deep(.el-card__header) {
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  font-size: 16px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-icon {
  font-size: 20px;
  opacity: 0.9;
}

.monitor-content {
  padding: 0;
}

.usage-item {
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 2px solid #f0f0f0;
}

.usage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  
  .label {
    font-weight: 600;
    color: #303133;
    font-size: 14px;
  }
  
  .percent-value {
    font-weight: 700;
    font-size: 18px;
    color: #409eff;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }
}

.usage-progress {
  :deep(.el-progress-bar__outer) {
    border-radius: 12px;
    overflow: hidden;
  }
  
  :deep(.el-progress-bar__inner) {
    border-radius: 12px;
    transition: width 0.6s ease;
  }
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  
  .full-width {
    grid-column: 1 / -1;
  }
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  background: linear-gradient(135deg, #f5f7fa 0%, #ffffff 100%);
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  transition: all 0.3s ease;
  
  &:hover {
    background: linear-gradient(135deg, #e8f4f8 0%, #f0f9ff 100%);
    border-color: #409eff;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
  }
}

.info-label {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: 15px;
  color: #303133;
  font-weight: 600;
  word-break: break-all;
  
  &.highlight {
    color: #409eff;
    font-size: 16px;
  }
  
  &.error {
    color: #f56c6c;
  }
}

.load-display {
  text-align: center;
  padding: 30px 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #ffffff 100%);
  border-radius: 12px;
  border: 2px solid #e4e7ed;
}

.load-value {
  margin-bottom: 12px;
}

.load-number {
  font-size: 48px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  display: inline-block;
  margin-right: 8px;
}

.load-percent {
  font-size: 20px;
  color: #909399;
  font-weight: 500;
}

.load-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.interfaces-table {
  :deep(.el-table__header) {
    th {
      background: linear-gradient(135deg, #f5f7fa 0%, #e8f4f8 100%);
      color: #303133;
      font-weight: 600;
      border-bottom: 2px solid #409eff;
    }
  }
  
  :deep(.el-table__row) {
    transition: all 0.3s ease;
    
    &:hover {
      background-color: #f0f9ff;
      transform: scale(1.01);
    }
  }
  
  :deep(.el-table__body) {
    tr {
      td {
        padding: 12px 0;
      }
    }
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .monitor-page {
    padding: 12px;
  }
  
  .monitor-card {
    margin-bottom: 16px;
  }
  
  .load-number {
    font-size: 36px;
  }
}
</style>

