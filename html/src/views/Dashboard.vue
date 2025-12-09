<template>
  <div class="dashboard">
    <!-- 页面头部：刷新按钮 -->
    <div class="dashboard-header">
      <h2>{{ $t('menu.dashboard') }}</h2>
      <el-button 
        :icon="RefreshIcon" 
        :loading="refreshing"
        :disabled="refreshing"
        @click="handleRefresh"
      >
        {{ $t('tabs.refresh') || '刷新' }}
      </el-button>
    </div>
    
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
      
      <!-- 月度操作统计 -->
      <el-col :xs="24" :sm="24" :md="12" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>月度操作统计</span>
              <el-tag type="info" size="small">近12个月</el-tag>
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
              <el-button type="primary" text size="small" @click="handleViewAllActivities">查看全部</el-button>
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
import { ref, onMounted, nextTick, onBeforeUnmount, markRaw, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { useAppStore } from '../store/app'
import { 
  getCount, 
  getUserAccessSource, 
  getWeeklyUserActivity, 
  getMonthlySales,
  getRecentActivities
} from '../api/dashboard'
import logger from '../utils/logger'
import ErrorHandler from '../utils/errorHandler'
import { useI18n } from 'vue-i18n'
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
  Setting,
  Refresh
} from '@element-plus/icons-vue'

const router = useRouter()
const appStore = useAppStore()
const { t, te, tm } = useI18n()

// 获取当前主题
const isDark = computed(() => appStore.theme === 'dark')
const textColor = computed(() => isDark.value ? '#e5eaf3' : '#303133')
const secondaryTextColor = computed(() => isDark.value ? '#a3a6ad' : '#909399')

// 使用 markRaw 标记图标组件，避免被 Vue 做成响应式对象
const UserFilledIcon = markRaw(UserFilled)
const ViewIcon = markRaw(View)
const KeyIcon = markRaw(Key)
const MenuIcon = markRaw(Menu)
const ArrowUpIcon = markRaw(ArrowUp)
const ArrowDownIcon = markRaw(ArrowDown)
const PlusIcon = markRaw(Plus)
const SettingIcon = markRaw(Setting)
const RefreshIcon = markRaw(Refresh)

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

// 刷新状态
const refreshing = ref(false)

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
  if (!sourceData || !Array.isArray(sourceData)) {
    // 如果没有数据，使用默认数据
    deviceData.value = [
      { value: 0, name: '桌面端' },
      { value: 0, name: '移动端' },
      { value: 0, name: '平板端' },
      { value: 0, name: '其他' }
    ]
  } else {
    // 使用真实的访问来源数据
    deviceData.value = sourceData.map(item => ({
      value: item.value || item.Value || 0,
      name: item.name || item.Name || ''
    }))
  }
  
  if (deviceChartInstance) {
    deviceChartInstance.setOption({
      series: [{
        data: deviceData.value
      }]
    })
  }
}

// 更新月度操作统计图表（替换地区分布）
const updateRegionChart = (monthlyData) => {
  if (!monthlyData || !Array.isArray(monthlyData)) return
  
  regionData.value = monthlyData.map(item => ({
    name: item.month || item.Month || '',
    value: item.count || item.Count || 0
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

// 手动刷新 Dashboard 数据
const handleRefresh = async () => {
  if (refreshing.value) return
  
  refreshing.value = true
  try {
    await loadDashboardData()
  } finally {
    refreshing.value = false
  }
}

// 加载 Dashboard 数据
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
      // 同时更新设备分布图表
      updateDeviceChart(sourceRes.data)
    }
    
    // 加载每周活动
    const weeklyRes = await getWeeklyUserActivity()
    if (weeklyRes.data) {
      updateVisitTrendChart(weeklyRes.data)
    }
    
    // 加载月度操作统计（替换销售额）
    const salesRes = await getMonthlySales()
    if (salesRes.data) {
      updateRegionChart(salesRes.data)
    }
    
    // 加载最近活动
    const activitiesRes = await getRecentActivities()
    if (activitiesRes.data) {
      updateRecentActivities(activitiesRes.data)
    }
  } catch (error) {
    logger.error('Failed to load dashboard data:', error)
    ErrorHandler.handle(error, { showNotification: true })
  }
}

// 最近活动
const recentActivities = ref([])

// 复数转单数转换函数
const pluralToSingular = (word) => {
  if (!word || word.length <= 1) return word
  
  // 特殊映射
  const specialCases = {
    'dictionaries': 'dictionary',
    'notifications': 'notification',
    'departments': 'department',
    'admins': 'admin',
    'roles': 'role',
    'permissions': 'permission',
    'menus': 'menu',
    'operation_logs': 'operation_log',
    'login_logs': 'login_log',
    'system_logs': 'system_log'
  }
  
  if (specialCases[word]) {
    return specialCases[word]
  }
  
  // 以 -s 结尾的单词，去掉 s
  if (word.endsWith('s')) {
    // 特殊情况：-ies 结尾的单词（如 dictionaries -> dictionary）
    if (word.endsWith('ies') && word.length > 3) {
      return word.slice(0, -3) + 'y'
    }
    // 特殊情况：-es 结尾的单词（如 permissions -> permission）
    if (word.endsWith('es') && word.length > 2) {
      const beforeEs = word.slice(0, -2)
      // 如果去掉 es 后以 ch, sh, x, s, z 结尾，保留 e
      const lastChar = beforeEs[beforeEs.length - 1]
      if (['c', 's', 'x', 'z'].includes(lastChar)) {
        return beforeEs
      }
      return beforeEs
    }
    return word.slice(0, -1)
  }
  
  // 默认返回原值
  return word
}

// 获取操作标题的翻译
const getOperationTitle = (titleKey) => {
  if (!titleKey) return '-'
  
  // 先尝试将复数形式转换为单数形式
  let slug = titleKey
  if (slug.includes('.')) {
    const parts = slug.split('.')
    if (parts.length >= 2) {
      const module = pluralToSingular(parts[0])
      slug = module + '.' + parts.slice(1).join('.')
    }
  } else {
    slug = pluralToSingular(slug)
  }
  
  // 作为权限标识翻译：permission.dictionary.update 这种形式
  const slugKey = `permission.${slug}`
  
  // 使用 te 检测路径是否存在（兼容嵌套路径）
  if (typeof te === 'function' && te(slugKey)) {
    return t(slugKey)
  }
  
  // 直接从 permission 命名空间对象里取（兼容平铺的 "dictionary.update" 键）
  const messages = typeof tm === 'function' ? tm('permission') : null
  if (messages && Object.prototype.hasOwnProperty.call(messages, slug)) {
    const value = messages[slug]
    if (typeof value === 'string') {
      return value
    }
  }
  
  // 如果转换后的 slug 和原始 titleKey 不同，再尝试用原始值查找一次（兼容旧数据）
  if (slug !== titleKey) {
    const originalSlugKey = `permission.${titleKey}`
    if (typeof te === 'function' && te(originalSlugKey)) {
      return t(originalSlugKey)
    }
    const originalMessages = typeof tm === 'function' ? tm('permission') : null
    if (originalMessages && Object.prototype.hasOwnProperty.call(originalMessages, titleKey)) {
      const value = originalMessages[titleKey]
      if (typeof value === 'string') {
        return value
      }
    }
  }
  
  // 找不到翻译就原样返回
  return titleKey
}

// 更新最近活动
const updateRecentActivities = (activities) => {
  if (activities && Array.isArray(activities)) {
    recentActivities.value = activities.map(item => ({
      user: item.user || '未知用户',
      action: getOperationTitle(item.action || ''), // 翻译操作标题
      time: item.time || '',
      status: item.status || '成功',
      type: item.type || 'success',
      avatarColor: item.avatarColor || '#409EFF'
    }))
  }
}

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
  
  // 如果没有数据，使用默认空数据
  if (!visitTrendData.value.dates || visitTrendData.value.dates.length === 0) {
    const now = new Date()
    visitTrendData.value = {
      dates: Array.from({ length: 7 }, (_, i) => {
        const date = new Date(now)
        date.setDate(date.getDate() - (6 - i))
        return date.toISOString().split('T')[0]
      }),
      visits: Array(7).fill(0),
      users: Array(7).fill(0)
    }
  }
  
  visitTrendChartInstance.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      },
      textStyle: {
        color: textColor.value
      }
    },
    legend: {
      data: ['访问量', '用户数'],
      textStyle: {
        color: textColor.value
      }
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
      data: visitTrendData.value.dates,
      axisLabel: {
        color: textColor.value
      },
      axisLine: {
        lineStyle: {
          color: isDark.value ? '#3d3e40' : '#dcdfe6'
        }
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: textColor.value
      },
      axisLine: {
        lineStyle: {
          color: isDark.value ? '#3d3e40' : '#dcdfe6'
        }
      },
      splitLine: {
        lineStyle: {
          color: isDark.value ? '#3d3e40' : '#ebeef5'
        }
      }
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
  
  // 如果没有数据，使用默认空数据
  if (!accessSourceData.value || accessSourceData.value.length === 0) {
    accessSourceData.value = [
      { value: 0, name: '桌面端' },
      { value: 0, name: '移动端' },
      { value: 0, name: '平板端' },
      { value: 0, name: '其他' }
    ]
  }
  
  accessSourceChartInstance.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)',
      textStyle: {
        color: textColor.value
      }
    },
    legend: {
      orient: 'vertical',
      left: 'left',
      top: 'middle',
      textStyle: {
        color: textColor.value
      }
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
          formatter: '{b}: {d}%',
          color: textColor.value
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 16,
            fontWeight: 'bold',
            color: textColor.value
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
  
  // 如果没有数据，使用默认空数据
  if (!deviceData.value || deviceData.value.length === 0) {
    deviceData.value = [
      { value: 0, name: '桌面端' },
      { value: 0, name: '移动端' },
      { value: 0, name: '平板端' },
      { value: 0, name: '其他' }
    ]
  }
  
  deviceChartInstance.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c}% ({d}%)',
      textStyle: {
        color: textColor.value
      }
    },
    legend: {
      bottom: '5%',
      left: 'center',
      textStyle: {
        color: textColor.value
      }
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

// 初始化月度操作统计图表（替换地区分布）
const initRegionChart = () => {
  if (!regionChart.value) return
  
  regionChartInstance = echarts.init(regionChart.value)
  
  // 如果没有数据，使用默认空数据
  if (!regionData.value || regionData.value.length === 0) {
    const now = new Date()
    regionData.value = Array.from({ length: 12 }, (_, i) => {
      const date = new Date(now)
      date.setMonth(date.getMonth() - (11 - i))
      return {
        name: `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`,
        value: 0
      }
    })
  }
  
  regionChartInstance.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      },
      textStyle: {
        color: textColor.value
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      axisLabel: {
        color: textColor.value
      },
      axisLine: {
        lineStyle: {
          color: isDark.value ? '#3d3e40' : '#dcdfe6'
        }
      },
      splitLine: {
        lineStyle: {
          color: isDark.value ? '#3d3e40' : '#ebeef5'
        }
      }
    },
    yAxis: {
      type: 'category',
      data: regionData.value.map(item => item.name),
      axisLabel: {
        color: textColor.value
      },
      axisLine: {
        lineStyle: {
          color: isDark.value ? '#3d3e40' : '#dcdfe6'
        }
      }
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

// 查看全部活动
const handleViewAllActivities = () => {
  router.push('/operation-logs')
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

// 监听主题变化，重新初始化图表以应用新的文字颜色
watch(isDark, () => {
  if (visitTrendChartInstance) {
    initVisitTrendChart()
  }
  if (accessSourceChartInstance) {
    initAccessSourceChart()
  }
  if (deviceChartInstance) {
    initDeviceChart()
  }
  if (regionChartInstance) {
    initRegionChart()
  }
})

onMounted(() => {
  initCharts()
  // 初始加载数据
  loadDashboardData()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  visitTrendChartInstance?.dispose()
  accessSourceChartInstance?.dispose()
  deviceChartInstance?.dispose()
  regionChartInstance?.dispose()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 0 4px;
}

.dashboard-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
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

<style>
/* 仪表盘页面夜间模式适配 - 需要非 scoped 样式来覆盖组件内部样式 */
.dark-mode .dashboard-header h2 {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-value {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-title {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-trend {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-trend span {
  color: var(--text-color-primary) !important;
}

.dark-mode .trend-up {
  color: #67C23A !important;
}

.dark-mode .trend-down {
  color: #F56C6C !important;
}

.dark-mode .card-header {
  color: var(--text-color-primary) !important;
}

.dark-mode .card-header span {
  color: var(--text-color-primary) !important;
}

.dark-mode .card-header * {
  color: var(--text-color-primary) !important;
}

/* 确保 el-card__header 中的文字也是白色 */
.dark-mode .el-card__header {
  color: var(--text-color-primary) !important;
}

.dark-mode .el-card__header * {
  color: var(--text-color-primary) !important;
}

.dark-mode .el-card__header .card-header {
  color: var(--text-color-primary) !important;
}

.dark-mode .el-card__header .card-header span {
  color: var(--text-color-primary) !important;
}

.dark-mode .el-card__header .card-header * {
  color: var(--text-color-primary) !important;
}

/* 确保所有卡片头部内的文字都是白色 */
.dark-mode .dashboard .el-card__header {
  color: var(--text-color-primary) !important;
}

.dark-mode .dashboard .el-card__header * {
  color: var(--text-color-primary) !important;
}

.dark-mode .dashboard .el-card__header span {
  color: var(--text-color-primary) !important;
}

.dark-mode .activity-text {
  color: var(--text-color-primary) !important;
}

.dark-mode .activity-user {
  color: var(--text-color-primary) !important;
}

.dark-mode .activity-time {
  color: var(--text-color-regular) !important;
}

.dark-mode .activity-content {
  color: var(--text-color-primary) !important;
}

.dark-mode .activity-content * {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-info {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-info * {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-info .stat-title {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-info .stat-trend {
  color: var(--text-color-primary) !important;
}

/* 确保表格中的文字也是白色 */
.dark-mode .el-table {
  color: var(--text-color-primary) !important;
}

.dark-mode .el-table * {
  color: var(--text-color-primary) !important;
}

.dark-mode .el-table td {
  color: var(--text-color-primary) !important;
}

.dark-mode .el-table td * {
  color: var(--text-color-primary) !important;
}

/* 确保所有卡片内的文字都是白色 */
.dark-mode .el-card {
  color: var(--text-color-primary) !important;
}

.dark-mode .el-card * {
  color: var(--text-color-primary) !important;
}

/* 确保整个仪表盘页面的所有文字都是白色 */
.dark-mode .dashboard {
  color: var(--text-color-primary) !important;
}

.dark-mode .dashboard * {
  color: var(--text-color-primary) !important;
}

/* 确保快速操作按钮文字也是白色 */
.dark-mode .quick-actions {
  color: var(--text-color-primary) !important;
}

.dark-mode .quick-actions * {
  color: var(--text-color-primary) !important;
}

/* 确保统计卡片内的所有文字都是白色 */
.dark-mode .stat-card * {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-content * {
  color: var(--text-color-primary) !important;
}

.dark-mode .stat-info * {
  color: var(--text-color-primary) !important;
}
</style>