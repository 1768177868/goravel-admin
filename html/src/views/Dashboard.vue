<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="12" :md="6" :lg="6" v-for="stat in stats" :key="stat.title">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" :style="{ backgroundColor: stat.color }">
              <el-icon :size="28"><component :is="stat.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatNumber(stat.value) }}</div>
              <div class="stat-title">{{ stat.title }}</div>
              <div class="stat-trend" v-if="stat.trend">
                <el-icon :class="stat.trend > 0 ? 'trend-up' : 'trend-down'">
                  <ArrowUpIcon v-if="stat.trend > 0" />
                  <ArrowDownIcon v-else />
                </el-icon>
                <span>{{ Math.abs(stat.trend) }}%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <!-- 访问趋势 -->
      <el-col :xs="24" :sm="24" :md="12" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>访问趋势</span>
              <el-tag type="success" size="small">近7天</el-tag>
            </div>
          </template>
          <div ref="visitTrendChart" style="height: 320px;"></div>
        </el-card>
      </el-col>
      
      <!-- 用户访问来源 -->
      <el-col :xs="24" :sm="24" :md="12" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>用户访问来源</span>
            </div>
          </template>
          <div ref="accessSourceChart" style="height: 320px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="charts-row">
      <!-- 设备分布 -->
      <el-col :xs="24" :sm="24" :md="12" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>设备分布</span>
            </div>
          </template>
          <div ref="deviceChart" style="height: 320px;"></div>
        </el-card>
      </el-col>
      
      <!-- 地区分布 -->
      <el-col :xs="24" :sm="24" :md="12" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>地区分布</span>
            </div>
          </template>
          <div ref="regionChart" style="height: 320px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近活动和快速操作 -->
    <el-row :gutter="20" class="bottom-row">
      <el-col :xs="24" :sm="24" :md="16" :lg="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>最近活动</span>
              <el-button type="primary" text size="small">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentActivities" style="width: 100%" :show-header="false">
            <el-table-column width="50">
              <template #default="{ row }">
                <el-avatar :size="36" :style="{ backgroundColor: row.avatarColor }">
                  {{ row.user.charAt(0) }}
                </el-avatar>
              </template>
            </el-table-column>
            <el-table-column>
              <template #default="{ row }">
                <div class="activity-content">
                  <div class="activity-text">
                    <span class="activity-user">{{ row.user }}</span>
                    <span>{{ row.action }}</span>
                  </div>
                  <div class="activity-time">{{ row.time }}</div>
                </div>
              </template>
            </el-table-column>
            <el-table-column width="80" align="right">
              <template #default="{ row }">
                <el-tag :type="row.type" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="24" :md="8" :lg="8">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>快速操作</span>
            </div>
          </template>
          <div class="quick-actions">
            <el-button 
              v-for="action in quickActions" 
              :key="action.name"
              :type="action.type"
              :icon="action.icon"
              class="quick-action-btn"
              @click="handleQuickAction(action)"
            >
              {{ action.name }}
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, onBeforeUnmount, onUnmounted, markRaw } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { 
  getCount, 
  getUserAccessSource, 
  getWeeklyUserActivity, 
  getMonthlySales,
  createDashboardSSE 
} from '../api/dashboard'
import { createSSEConnection, closeSSEConnection } from '../utils/sse'
import {
  User,
  View,
  ShoppingCart,
  Money,
  UserFilled,
  Key,
  Menu,
  Document,
  ArrowUp,
  ArrowDown,
  Plus,
  Edit,
  Delete,
  Setting
} from '@element-plus/icons-vue'

const router = useRouter()

// 使用 markRaw 标记图标组件，避免被 Vue 做成响应式对象
const UserFilledIcon = markRaw(UserFilled)
const ViewIcon = markRaw(View)
const KeyIcon = markRaw(Key)
const MenuIcon = markRaw(Menu)
const ArrowUpIcon = markRaw(ArrowUp)
const ArrowDownIcon = markRaw(ArrowDown)
const PlusIcon = markRaw(Plus)
const SettingIcon = markRaw(Setting)

// 统计数据
const stats = ref([
  { 
    title: '管理员总数', 
    value: 0, 
    icon: UserFilledIcon, 
    color: '#409EFF',
    trend: 0
  },
  { 
    title: '今日访问', 
    value: 0, 
    icon: ViewIcon, 
    color: '#67C23A',
    trend: 0
  },
  { 
    title: '角色数量', 
    value: 0, 
    icon: KeyIcon, 
    color: '#E6A23C',
    trend: 0
  },
  { 
    title: '菜单数量', 
    value: 0, 
    icon: MenuIcon, 
    color: '#F56C6C',
    trend: 0
  }
])

// 图表引用
const visitTrendChart = ref(null)
const accessSourceChart = ref(null)
const deviceChart = ref(null)
const regionChart = ref(null)

// 图表实例
let visitTrendChartInstance = null
let accessSourceChartInstance = null
let deviceChartInstance = null
let regionChartInstance = null

// 访问趋势数据
const visitTrendData = ref({
  dates: [],
  visits: [],
  users: []
})

// 访问来源数据
const accessSourceData = ref([])

// 设备分布数据（从访问来源数据中提取）
const deviceData = ref([])

// 地区分布数据（暂时使用空数据，后续可以从后端获取）
const regionData = ref([])

// Dashboard SSE 连接
let dashboardEventSource = null

// 更新统计数据
const updateStats = (countData) => {
  if (!countData) return
  
  stats.value[0].value = countData.admin_count || 0
  stats.value[1].value = countData.today_visits || 0
  stats.value[2].value = countData.role_count || 0
  stats.value[3].value = countData.menu_count || 0
}

// 更新访问趋势图表
const updateVisitTrendChart = (weeklyData) => {
  if (!weeklyData || !Array.isArray(weeklyData) || weeklyData.length === 0) return
  
  visitTrendData.value = {
    dates: weeklyData.map(item => item.date || item.Date || ''),
    visits: weeklyData.map(item => item.visits || item.Visits || 0),
    users: weeklyData.map(item => item.users || item.Users || 0)
  }
  
  if (visitTrendChartInstance) {
    visitTrendChartInstance.setOption({
      xAxis: {
        data: visitTrendData.value.dates
      },
      series: [
        {
          data: visitTrendData.value.visits
        },
        {
          data: visitTrendData.value.users
        }
      ]
    })
  }
}

// 更新访问来源图表
const updateAccessSourceChart = (sourceData) => {
  if (!sourceData || !Array.isArray(sourceData) || sourceData.length === 0) return
  
  accessSourceData.value = sourceData.map(item => ({
    value: item.value || item.Value || 0,
    name: item.name || item.Name || ''
  }))
  
  if (accessSourceChartInstance) {
    accessSourceChartInstance.setOption({
      series: [{
        data: accessSourceData.value
      }]
    })
  }
}

// 更新设备分布图表（从访问来源数据中提取设备相关数据）
const updateDeviceChart = (sourceData) => {
  if (!sourceData || !Array.isArray(sourceData)) return
  
  // 假设访问来源数据中包含设备信息，或者可以从其他接口获取
  // 这里暂时使用默认数据，实际应该从后端获取
  deviceData.value = [
    { value: 45, name: '桌面端' },
    { value: 35, name: '移动端' },
    { value: 20, name: '平板端' }
  ]
  
  if (deviceChartInstance) {
    deviceChartInstance.setOption({
      series: [{
        data: deviceData.value
      }]
    })
  }
}

// 更新地区分布图表
const updateRegionChart = (regionDataArray) => {
  if (!regionDataArray || !Array.isArray(regionDataArray)) return
  
  regionData.value = regionDataArray.map(item => ({
    name: item.name || item.Name || '',
    value: item.value || item.Value || 0
  }))
  
  if (regionChartInstance) {
    regionChartInstance.setOption({
      yAxis: {
        data: regionData.value.map(item => item.name)
      },
      series: [{
        data: regionData.value.map(item => item.value)
      }]
    })
  }
}

// 启动 SSE 实时更新
const startDashboardSSE = () => {
  try {
    const url = createDashboardSSE({ interval: 5 })
    dashboardEventSource = createSSEConnection(url, {
      onMessage: (data) => {
        if (data.type === 'dashboard_data') {
          const dashboardData = data.data || {}
          
          // 更新统计数据
          if (dashboardData.count) {
            updateStats(dashboardData.count)
          }
          
          // 更新访问来源
          if (dashboardData.user_access_source) {
            updateAccessSourceChart(dashboardData.user_access_source)
          }
          
          // 更新每周活动
          if (dashboardData.weekly_user_activity) {
            updateVisitTrendChart(dashboardData.weekly_user_activity)
          }
          
          // 更新每月销售（如果有）
          if (dashboardData.monthly_sales) {
            // 可以在这里更新销售相关的图表
          }
        }
      },
      onError: (error) => {
        console.error('Dashboard SSE error:', error)
        // SSE 连接失败时，降级到定时刷新
        if (dashboardEventSource && dashboardEventSource.readyState === EventSource.CLOSED) {
          closeSSEConnection(dashboardEventSource)
          dashboardEventSource = null
          // 可以启动定时刷新作为降级方案
          loadDashboardData()
        }
      },
      onOpen: () => {
        console.log('Dashboard SSE connected')
      }
    })
  } catch (error) {
    console.error('Failed to start Dashboard SSE:', error)
    // 降级到普通 API 调用
    loadDashboardData()
  }
}

// 加载 Dashboard 数据（降级方案）
const loadDashboardData = async () => {
  try {
    // 加载统计数据
    const countRes = await getCount()
    if (countRes.data) {
      updateStats(countRes.data)
    }
    
    // 加载访问来源
    const sourceRes = await getUserAccessSource()
    if (sourceRes.data) {
      updateAccessSourceChart(sourceRes.data)
    }
    
    // 加载每周活动
    const weeklyRes = await getWeeklyUserActivity()
    if (weeklyRes.data) {
      updateVisitTrendChart(weeklyRes.data)
    }
    
    // 加载每月销售（如果有）
    // const salesRes = await getMonthlySales()
    // if (salesRes.data) {
    //   updateSalesChart(salesRes.data)
    // }
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
  }
}

// 最近活动 - 写死的默认数据
const recentActivities = ref([
  {
    user: '张三',
    action: '创建了新角色',
    time: '2分钟前',
    status: '成功',
    type: 'success',
    avatarColor: '#409EFF'
  },
  {
    user: '李四',
    action: '修改了菜单权限',
    time: '15分钟前',
    status: '成功',
    type: 'success',
    avatarColor: '#67C23A'
  },
  {
    user: '王五',
    action: '删除了管理员',
    time: '1小时前',
    status: '成功',
    type: 'success',
    avatarColor: '#E6A23C'
  },
  {
    user: '赵六',
    action: '更新了系统配置',
    time: '2小时前',
    status: '成功',
    type: 'success',
    avatarColor: '#F56C6C'
  },
  {
    user: '钱七',
    action: '导出了用户数据',
    time: '3小时前',
    status: '成功',
    type: 'success',
    avatarColor: '#909399'
  }
])

// 快速操作
const quickActions = [
  { name: '添加管理员', type: 'primary', icon: PlusIcon, path: '/admins' },
  { name: '创建角色', type: 'success', icon: PlusIcon, path: '/roles' },
  { name: '管理菜单', type: 'warning', icon: MenuIcon, path: '/menus' },
  { name: '系统设置', type: 'info', icon: SettingIcon, path: '/configs' }
]

// 格式化数字
const formatNumber = (num) => {
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toLocaleString()
}

// 初始化访问趋势图表
const initVisitTrendChart = () => {
  if (!visitTrendChart.value) return
  
  visitTrendChartInstance = echarts.init(visitTrendChart.value)
  visitTrendChartInstance.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      }
    },
    legend: {
      data: ['访问量', '用户数']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: visitTrendData.value.dates
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '访问量',
        type: 'line',
        smooth: true,
        data: visitTrendData.value.visits,
        itemStyle: {
          color: '#409EFF'
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
              { offset: 1, color: 'rgba(64, 158, 255, 0.1)' }
            ]
          }
        }
      },
      {
        name: '用户数',
        type: 'line',
        smooth: true,
        data: visitTrendData.value.users,
        itemStyle: {
          color: '#67C23A'
        }
      }
    ]
  })
}

// 初始化访问来源图表
const initAccessSourceChart = () => {
  if (!accessSourceChart.value) return
  
  accessSourceChartInstance = echarts.init(accessSourceChart.value)
  accessSourceChartInstance.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left',
      top: 'middle'
    },
    series: [
      {
        name: '访问来源',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: true,
          formatter: '{b}: {d}%'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 16,
            fontWeight: 'bold'
          }
        },
        data: accessSourceData.value
      }
    ]
  })
}

// 初始化设备分布图表
const initDeviceChart = () => {
  if (!deviceChart.value) return
  
  deviceChartInstance = echarts.init(deviceChart.value)
  deviceChartInstance.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c}% ({d}%)'
    },
    legend: {
      bottom: '5%',
      left: 'center'
    },
    series: [
      {
        name: '设备分布',
        type: 'pie',
        radius: '60%',
        center: ['50%', '45%'],
        data: deviceData.value,
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  })
}

// 初始化地区分布图表
const initRegionChart = () => {
  if (!regionChart.value) return
  
  regionChartInstance = echarts.init(regionChart.value)
  regionChartInstance.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'value'
    },
    yAxis: {
      type: 'category',
      data: regionData.value.map(item => item.name)
    },
    series: [
      {
        name: '访问量',
        type: 'bar',
        data: regionData.value.map(item => item.value),
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
            { offset: 0, color: '#83bff6' },
            { offset: 0.5, color: '#188df0' },
            { offset: 1, color: '#188df0' }
          ])
        },
        emphasis: {
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
              { offset: 0, color: '#2378f7' },
              { offset: 0.7, color: '#2378f7' },
              { offset: 1, color: '#83bff6' }
            ])
          }
        }
      }
    ]
  })
}

// 处理窗口大小变化
const handleResize = () => {
  visitTrendChartInstance?.resize()
  accessSourceChartInstance?.resize()
  deviceChartInstance?.resize()
  regionChartInstance?.resize()
}

// 快速操作处理
const handleQuickAction = (action) => {
  if (action.path) {
    router.push(action.path)
  }
}

// 初始化所有图表
const initCharts = async () => {
  await nextTick()
  initVisitTrendChart()
  initAccessSourceChart()
  initDeviceChart()
  initRegionChart()
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)
}

onMounted(() => {
  initCharts()
  // 优先使用 SSE 实时更新
  startDashboardSSE()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  visitTrendChartInstance?.dispose()
  accessSourceChartInstance?.dispose()
  deviceChartInstance?.dispose()
  regionChartInstance?.dispose()
})

onUnmounted(() => {
  // 关闭 SSE 连接
  if (dashboardEventSource) {
    closeSSEConnection(dashboardEventSource)
    dashboardEventSource = null
  }
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
  transition: all 0.3s ease;
  border: none;
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin-right: 16px;
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
  line-height: 1.2;
}

.stat-title {
  font-size: 14px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-trend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  margin-top: 4px;
}

.trend-up {
  color: #67C23A;
}

.trend-down {
  color: #F56C6C;
}

.charts-row {
  margin-bottom: 20px;
}

.bottom-row {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

:deep(.el-card__header) {
  padding: 18px 20px;
  border-bottom: 1px solid #EBEEF5;
}

:deep(.el-card__body) {
  padding: 20px;
}

.activity-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.activity-text {
  font-size: 14px;
  color: #606266;
}

.activity-user {
  font-weight: 600;
  color: #303133;
  margin-right: 6px;
}

.activity-time {
  font-size: 12px;
  color: #909399;
}

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.quick-action-btn {
  width: 100%;
  justify-content: flex-start;
  height: 44px;
  font-size: 14px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stat-value {
    font-size: 24px;
  }
  
  .stat-icon {
    width: 56px;
    height: 56px;
    margin-right: 12px;
  }
}
</style>
