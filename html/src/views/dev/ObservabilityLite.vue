<template>
  <div class="lite-observability-page">
    <el-card shadow="never">
      <template #header>
        <div class="header">
          <span>{{ $t('observability_lite.title') }}</span>
          <el-button size="small" :loading="loading" @click="loadData">{{ $t('common.refresh') }}</el-button>
        </div>
      </template>

      <el-alert
        :title="$t('observability_lite.tip')"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      />

      <el-descriptions :column="2" border>
        <el-descriptions-item :label="$t('observability_lite.timestamp')">{{ data.timestamp || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.app_env')">{{ data.app?.env || data.app_env || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.app_debug')">{{ formatBool(data.app?.debug ?? data.app_debug) }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.queue_connection')">{{ data.app?.queue_connection || data.queue_connection || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.cache_store')">{{ data.app?.cache_store || data.cache_store || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.ws_connections')">{{ data.websocket?.connections ?? data.ws?.connections ?? 0 }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.ws_online_admins')">{{ data.websocket?.online_admins ?? data.ws?.online_admins ?? 0 }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.goroutines')">{{ data.runtime?.goroutines ?? 0 }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.go_version')">{{ data.runtime?.go_version || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.num_cpu')">{{ data.runtime?.num_cpu ?? 0 }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.cpu_usage')">{{ formatPercent(data.system_snapshot?.cpu?.percent) }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.memory_usage')">{{ formatPercent(data.system_snapshot?.memory?.percent) }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.disk_usage')">{{ formatPercent(data.system_snapshot?.disk?.percent) }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.net_speed')">{{ formatSpeed(data.system_snapshot?.net?.speed_total_mbps) }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.system_os')">{{ data.system_snapshot?.system?.os || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('observability_lite.system_host')">{{ data.system_snapshot?.system?.hostname || '-' }}</el-descriptions-item>
      </el-descriptions>

      <el-collapse style="margin-top: 16px">
        <el-collapse-item name="raw" :title="$t('observability_lite.raw_snapshot')">
          <pre class="json-pre">{{ formattedSnapshot }}</pre>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getObservabilityLite } from '../../api/monitor'

const loading = ref(false)
const data = ref({})

const formatBool = (value) => (value ? 'true' : 'false')
const formatPercent = (value) => {
  if (typeof value !== 'number') return '-'
  return `${value.toFixed(2)}%`
}
const formatSpeed = (value) => {
  if (typeof value !== 'number') return '-'
  return `${value.toFixed(2)} Mbps`
}
const formattedSnapshot = computed(() => JSON.stringify(data.value.system_snapshot || {}, null, 2))

const loadData = async () => {
  loading.value = true
  try {
    const res = await getObservabilityLite()
    data.value = res.data || {}
  } catch (error) {
    if (!error.__handled) {
      ElMessage.error(error?.response?.data?.message || error.message || 'Load failed')
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.lite-observability-page {
  padding: 20px;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.json-pre {
  margin: 0;
  max-height: 420px;
  overflow: auto;
  padding: 12px;
  border-radius: 6px;
  background: var(--el-fill-color-light);
  font-size: 12px;
  line-height: 1.5;
}
</style>
