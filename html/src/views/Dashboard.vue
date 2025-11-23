<template>
  <div class="dashboard">
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6" v-for="stat in stats" :key="stat.title">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" :style="{ backgroundColor: stat.color }">
              <el-icon :size="30"><component :is="stat.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stat.value }}</div>
              <div class="stat-title">{{ stat.title }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="charts-row">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>用户访问来源</span>
          </template>
          <div ref="accessSourceChart" style="height: 300px;"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>每周用户活动</span>
          </template>
          <div ref="weeklyActivityChart" style="height: 300px;"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { getCount, getUserAccessSource, getWeeklyUserActivity } from '../api/dashboard'

const stats = ref([
  { title: '总用户数', value: 0, icon: 'User', color: '#409EFF' },
  { title: '今日访问', value: 0, icon: 'View', color: '#67C23A' },
  { title: '总订单数', value: 0, icon: 'ShoppingCart', color: '#E6A23C' },
  { title: '总收入', value: 0, icon: 'Money', color: '#F56C6C' }
])

const accessSourceChart = ref(null)
const weeklyActivityChart = ref(null)

const loadDashboardData = async () => {
  try {
    // 加载统计数据
    const countRes = await getCount()
    if (countRes.data) {
      const data = countRes.data
      stats.value[0].value = data.total_users || 0
      stats.value[1].value = data.today_visits || 0
      stats.value[2].value = data.total_orders || 0
      stats.value[3].value = data.total_revenue || 0
    }

    // 加载访问来源图表
    const sourceRes = await getUserAccessSource()
    if (sourceRes.data && accessSourceChart.value) {
      const chart = echarts.init(accessSourceChart.value)
      chart.setOption({
        tooltip: {
          trigger: 'item'
        },
        series: [{
          type: 'pie',
          data: sourceRes.data
        }]
      })
    }

    // 加载每周活动图表
    const activityRes = await getWeeklyUserActivity()
    if (activityRes.data && weeklyActivityChart.value) {
      const chart = echarts.init(weeklyActivityChart.value)
      chart.setOption({
        tooltip: {
          trigger: 'axis'
        },
        xAxis: {
          type: 'category',
          data: activityRes.data.map(item => item.date)
        },
        yAxis: {
          type: 'value'
        },
        series: [{
          data: activityRes.data.map(item => item.count),
          type: 'line'
        }]
      })
    }
  } catch (error) {
    console.error('Load dashboard data error:', error)
  }
}

onMounted(() => {
  loadDashboardData()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  cursor: pointer;
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin-right: 15px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 5px;
}

.stat-title {
  font-size: 14px;
  color: #909399;
}

.charts-row {
  margin-top: 20px;
}
</style>

