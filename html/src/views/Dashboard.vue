<template>
  <div class="dashboard">
    <!-- 页面头部：刷新按钮 -->
    <div class="dashboard-header">
      <div>
        <h2>{{ $t('menu.dashboard') }}</h2>
        <div class="dashboard-subtitle">
          数据口径：已支付订单 / 独立访客（UV） · 最近更新 {{ lastUpdatedAt }}
        </div>
      </div>
      <div class="header-actions">
        <el-select v-model="selectedPeriod" size="small" class="header-select">
          <el-option label="今日" value="today" />
          <el-option label="近7天" value="week" />
          <el-option label="近30天" value="month" />
        </el-select>
        <el-select v-model="selectedChannel" size="small" class="header-select">
          <el-option label="全渠道" value="all" />
          <el-option label="自然流量" value="organic" />
          <el-option label="广告投放" value="ads" />
          <el-option label="私域流量" value="private" />
        </el-select>
        <el-button 
          :icon="RefreshIcon" 
          :loading="refreshing"
          :disabled="refreshing"
          @click="handleRefresh"
        >
          {{ $t('tabs.refresh') || '刷新' }}
        </el-button>
      </div>
    </div>
    
    <el-skeleton v-if="dashboardLoading" :rows="5" animated class="dashboard-skeleton" />

    <!-- 统计卡片 -->
    <el-row v-else :gutter="20" class="stats-row">
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

    <el-row :gutter="20" class="overview-row">
      <el-col :xs="24" :sm="24" :md="16" :lg="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>经营概览</span>
              <el-tag size="small" type="info">{{ periodText }}</el-tag>
            </div>
          </template>
          <div v-if="kpiOverview.length > 0" class="kpi-grid">
            <div class="kpi-item" v-for="kpi in kpiOverview" :key="kpi.label">
              <div class="kpi-label">{{ kpi.label }}</div>
              <div class="kpi-value">{{ kpi.value }}</div>
              <div class="kpi-foot">
                <span>{{ kpi.subtitle }}</span>
                <span :class="kpi.change >= 0 ? 'trend-up' : 'trend-down'">
                  {{ kpi.change >= 0 ? '+' : '' }}{{ kpi.change }}%
                </span>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无经营数据" :image-size="74" />
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="24" :md="8" :lg="8">
        <el-card shadow="hover" class="warn-card">
          <template #header>
            <div class="card-header">
              <span>运营预警</span>
              <el-tag size="small" type="danger">需关注</el-tag>
            </div>
          </template>
          <div v-if="dashboardWarnings.length > 0" class="warning-list">
            <div class="warning-item" v-for="warning in dashboardWarnings" :key="warning.title">
              <div class="warning-title">
                <el-tag :type="warning.type" size="small">{{ warning.level }}</el-tag>
                <span>{{ warning.title }}</span>
              </div>
              <div class="warning-desc">{{ warning.desc }}</div>
            </div>
          </div>
          <el-empty v-else description="暂无预警信息" :image-size="74" />
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

    <el-row :gutter="20" class="overview-row">
      <el-col :xs="24" :sm="24" :md="12" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>渠道转化漏斗</span>
              <el-tag size="small">{{ periodText }}</el-tag>
            </div>
          </template>
          <div v-if="channelFunnel.length > 0" class="funnel-list">
            <div class="funnel-row" v-for="item in channelFunnel" :key="item.stage">
              <div class="funnel-header">
                <span>{{ item.stage }}</span>
                <span>{{ formatNumber(item.value) }} / {{ item.rate }}%</span>
              </div>
              <el-progress
                :percentage="item.rate"
                :stroke-width="10"
                :color="item.color"
                :show-text="false"
              />
            </div>
          </div>
          <el-empty v-else description="暂无漏斗数据" :image-size="74" />
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="24" :md="12" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>今日待办</span>
              <el-tag size="small" type="warning">{{ pendingTasks.length }}项</el-tag>
            </div>
          </template>
          <div v-if="pendingTasks.length > 0" class="todo-list">
            <div class="todo-item" v-for="task in pendingTasks" :key="task.title">
              <div class="todo-main">
                <div class="todo-title">{{ task.title }}</div>
                <div class="todo-meta">{{ task.owner }} · 截止 {{ task.deadline }}</div>
              </div>
              <el-tag :type="task.type" size="small">{{ task.priority }}</el-tag>
            </div>
          </div>
          <el-empty v-else description="暂无待办事项" :image-size="74" />
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
          <el-table v-if="recentActivities.length > 0" :data="recentActivities" style="width: 100%" :show-header="false">
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
          <el-empty v-else description="暂无最近活动" :image-size="74" />
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
import { useAppStore, THEME_COLORS } from '../store/app'
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
import { createOperationTitleTranslator } from '../utils/operationTitle'
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

const hexToRgba = (hex, alpha = 1) => {
  const m = hex.match(/^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i)
  if (!m) return `rgba(64, 158, 255, ${alpha})`
  const r = parseInt(m[1], 16)
  const g = parseInt(m[2], 16)
  const b = parseInt(m[3], 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}
const { t, te, tm } = useI18n()
const translateOperationTitle = createOperationTitleTranslator({ t, te, tm })

// 获取当前主题
const isDark = computed(() => appStore.darkMode)
const textColor = computed(() => isDark.value ? '#e5eaf3' : '#303133')
const secondaryTextColor = computed(() => isDark.value ? '#a3a6ad' : '#909399')
const tooltipTextColor = computed(() => isDark.value ? '#f5f7fa' : '#303133')
const tooltipBackgroundColor = computed(() => isDark.value ? 'rgba(22, 24, 29, 0.95)' : '#ffffff')
const tooltipBorderColor = computed(() => isDark.value ? '#4c4d4f' : '#dcdfe6')
const primaryColor = computed(() => {
  const preset = THEME_COLORS.find((t) => t.key === appStore.themeColor) || THEME_COLORS[0]
  return preset.color
})

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

// 统计数据（第一项颜色随主题色变化）
const statsData = ref([
  {
    title: '最近一年订单总数',
    value: 0,
    icon: markRaw(ShoppingCart),
    colorKey: 'primary',
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
const stats = computed(() =>
  statsData.value.map((s) => ({
    ...s,
    color: s.colorKey === 'primary' ? primaryColor.value : s.color
  }))
)

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
const dashboardLoading = ref(true)
const selectedPeriod = ref('week')
const selectedChannel = ref('all')
const lastUpdatedAt = ref('')
const kpiOverview = ref([])
const dashboardWarnings = ref([])
const channelFunnel = ref([])
const pendingTasks = ref([])
const periodText = computed(() => {
  if (selectedPeriod.value === 'today') return '今日'
  if (selectedPeriod.value === 'month') return '近30天'
  return '近7天'
})

const updateLastUpdatedAt = () => {
  lastUpdatedAt.value = new Date().toLocaleString('zh-CN', { hour12: false })
}

const buildMockDashboardMeta = () => {
  const scale = selectedPeriod.value === 'today' ? 0.35 : selectedPeriod.value === 'month' ? 4.2 : 1
  const channelFactorMap = {
    all: 1,
    organic: 0.46,
    ads: 0.34,
    private: 0.2
  }
  const factor = channelFactorMap[selectedChannel.value] || 1
  const exposure = Math.round(42650 * scale * factor)
  const click = Math.round(exposure * 0.276)
  const addCart = Math.round(click * 0.397)
  const order = Math.round(addCart * 0.325)
  const pay = Math.round(order * 0.846)

  kpiOverview.value = [
    { label: '支付订单', value: formatNumber(pay), subtitle: `${periodText.value}转化订单数`, change: 8.6 },
    { label: 'GMV(元)', value: formatNumber(pay * 305), subtitle: `${periodText.value}交易总额`, change: 12.4 },
    { label: '客单价(元)', value: `${Math.round(305 * (1 + (factor - 1) * 0.2))}`, subtitle: '平均每单金额', change: -2.1 },
    { label: '退款率', value: `${(1.26 + (1 - factor) * 0.3).toFixed(2)}%`, subtitle: '较上周期', change: -0.4 }
  ]

  dashboardWarnings.value = [
    { level: '高', title: '支付成功率波动', desc: `${periodText.value}支付成功率 92.4%，较基线低 1.8%。`, type: 'danger' },
    { level: '中', title: '库存预警 SKU 12 个', desc: '建议优先补货近7天转化率前20的商品。', type: 'warning' },
    { level: '低', title: '待审核内容 8 条', desc: '内容审核队列未超时，建议在晚间批量处理。', type: 'info' }
  ]

  channelFunnel.value = [
    { stage: '曝光', value: exposure, rate: 100, color: '#409EFF' },
    { stage: '点击', value: click, rate: Number(((click / exposure) * 100).toFixed(1)), color: '#67C23A' },
    { stage: '加购', value: addCart, rate: Number(((addCart / exposure) * 100).toFixed(1)), color: '#E6A23C' },
    { stage: '下单', value: order, rate: Number(((order / exposure) * 100).toFixed(1)), color: '#F56C6C' },
    { stage: '支付', value: pay, rate: Number(((pay / exposure) * 100).toFixed(1)), color: '#9B59B6' }
  ]

  pendingTasks.value = [
    { title: '处理支付异常订单', owner: '财务组', deadline: '20:30', priority: '紧急', type: 'danger' },
    { title: '审核高优促销活动', owner: '运营组', deadline: '21:00', priority: '高', type: 'warning' },
    { title: '复盘昨日转化漏斗', owner: '增长组', deadline: '22:00', priority: '中', type: 'info' },
    { title: '更新首页 Banner 素材', owner: '设计组', deadline: '23:00', priority: '低', type: 'success' }
  ]
  updateLastUpdatedAt()
}

// 更新统计数据
const updateStats = (countData) => {
  if (!countData) return
  
  stats.value[0].value = countData.order_count_in_year || 0
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
  dashboardLoading.value = true
  try {
    await loadDashboardData()
    buildMockDashboardMeta()
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
  } finally {
    if (kpiOverview.value.length === 0) {
      buildMockDashboardMeta()
    }
    dashboardLoading.value = false
  }
}

// 最近活动
const recentActivities = ref([])

// 更新最近活动
const updateRecentActivities = (activities) => {
  if (activities && Array.isArray(activities)) {
    recentActivities.value = activities.map(item => ({
      user: item.user || '未知用户',
      action: translateOperationTitle(item.action || ''),
      time: item.time || '',
      status: item.status || '成功',
      type: item.type || 'success',
      avatarColor: item.avatarColor || primaryColor.value
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

const getTooltipStyle = () => ({
  backgroundColor: tooltipBackgroundColor.value,
  borderColor: tooltipBorderColor.value,
  borderWidth: 1,
  textStyle: {
    color: tooltipTextColor.value
  }
})

// 初始化访问趋势图表
const initVisitTrendChart = () => {
  if (!visitTrendChart.value) return
  
  // 使用暗黑主题（如果启用）
  visitTrendChartInstance = echarts.init(visitTrendChart.value, isDark.value ? 'dark' : null)
  
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
      ...getTooltipStyle()
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
          color: primaryColor.value
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: hexToRgba(primaryColor.value, 0.3) },
              { offset: 1, color: hexToRgba(primaryColor.value, 0.1) }
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
  
  accessSourceChartInstance = echarts.init(accessSourceChart.value, isDark.value ? 'dark' : null)
  
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
      ...getTooltipStyle()
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
  
  deviceChartInstance = echarts.init(deviceChart.value, isDark.value ? 'dark' : null)
  
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
      ...getTooltipStyle()
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
  
  regionChartInstance = echarts.init(regionChart.value, isDark.value ? 'dark' : null)
  
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
      ...getTooltipStyle()
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
  // 暗黑模式切换时，重新初始化所有图表
  if (visitTrendChartInstance) {
    visitTrendChartInstance.dispose()
    initVisitTrendChart()
  }
  if (accessSourceChartInstance) {
    accessSourceChartInstance.dispose()
    initAccessSourceChart()
  }
  if (deviceChartInstance) {
    deviceChartInstance.dispose()
    initDeviceChart()
  }
  if (regionChartInstance) {
    regionChartInstance.dispose()
    initRegionChart()
  }
})
watch([selectedPeriod, selectedChannel], () => {
  buildMockDashboardMeta()
})

onMounted(() => {
  buildMockDashboardMeta()
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
  margin-bottom: var(--space-md);
  padding: 0 4px;
  gap: 12px;
}

.dashboard-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color-primary, #303133);
}

.dashboard-subtitle {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-color-secondary, #909399);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-select {
  width: 128px;
}

.stats-row {
  margin-bottom: var(--space-md);
}

.dashboard-skeleton {
  margin-bottom: var(--space-md);
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
  color: #fff;
  margin-right: var(--space-md);
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--text-color-primary, #303133);
  margin-bottom: 6px;
  line-height: 1.2;
}

.stat-title {
  font-size: 14px;
  color: var(--text-color-secondary, #909399);
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
  color: var(--el-color-success);
}

.trend-down {
  color: var(--el-color-danger);
}

.charts-row {
  margin-bottom: var(--space-md);
}

.overview-row {
  margin-bottom: var(--space-md);
}

.bottom-row {
  margin-top: var(--space-md);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

:deep(.el-card__header) {
  padding: 18px 20px;
  border-bottom: 1px solid var(--border-color-lighter);
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
  color: var(--text-color-regular);
}

.activity-user {
  font-weight: 600;
  color: var(--text-color-primary);
  margin-right: 6px;
}

.activity-time {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.quick-action-btn {
  width: 100%;
  justify-content: flex-start;
  height: 44px;
  font-size: 14px;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.kpi-item {
  border: 1px solid var(--border-color-lighter);
  border-radius: 10px;
  padding: 14px;
  background: var(--el-bg-color-page);
}

.kpi-label {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.kpi-value {
  margin-top: 8px;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-color-primary);
}

.kpi-foot {
  margin-top: 8px;
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.warn-card {
  height: 100%;
}

.warning-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.warning-item {
  padding: 10px 12px;
  border: 1px solid var(--border-color-lighter);
  border-radius: 8px;
  background: var(--el-fill-color-lighter);
}

.warning-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.warning-desc {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.funnel-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.funnel-header {
  margin-bottom: 8px;
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--text-color-regular);
}

.todo-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.todo-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border: 1px solid var(--border-color-lighter);
  border-radius: 8px;
  background: var(--el-bg-color-page);
}

.todo-main {
  min-width: 0;
}

.todo-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.todo-meta {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-color-secondary);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .dashboard-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .header-select {
    width: calc(50% - 6px);
  }

  .stat-value {
    font-size: 24px;
  }
  
  .stat-icon {
    width: 56px;
    height: 56px;
    margin-right: 12px;
  }

  .kpi-grid {
    grid-template-columns: 1fr;
  }
}
</style>

