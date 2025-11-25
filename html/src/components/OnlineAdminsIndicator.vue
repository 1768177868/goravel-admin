<template>
  <el-tooltip
    :content="$t('header.online_admins_tooltip', { minutes: windowMinutes })"
    placement="bottom"
  >
    <div class="online-indicator" @click="handleRefresh">
      <el-icon><UserFilled /></el-icon>
      <span class="count">{{ count }}</span>
    </div>
  </el-tooltip>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { UserFilled } from '@element-plus/icons-vue'
import { fetchOnlineAdmins } from '../api/system'

const { t } = useI18n()
const count = ref(0)
const windowMinutes = ref(5)
const loading = ref(false)
let timer = null

const load = async () => {
  if (loading.value) {
    return
  }
  loading.value = true
  try {
    const { data } = await fetchOnlineAdmins()
    count.value = data.count ?? 0
    windowMinutes.value = data.window_minutes ?? 5
  } catch (error) {
    console.error('Fetch online admins failed:', error)
    ElMessage.error(t('error.network'))
  } finally {
    loading.value = false
  }
}

const startTimer = () => {
  timer = setInterval(load, 30000)
}

const stopTimer = () => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

const handleRefresh = () => {
  load()
}

onMounted(() => {
  load()
  startTimer()
})

onBeforeUnmount(() => {
  stopTimer()
})
</script>

<style scoped>
.online-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid #ebeef5;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.2s;
  color: #606266;
}

.online-indicator:hover {
  border-color: #409eff;
  color: #409eff;
  background: #ecf5ff;
}

.count {
  font-weight: 600;
  min-width: 20px;
  text-align: center;
}
</style>

