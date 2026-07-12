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
import { useDashboard } from './dashboard/useDashboard.js'

const {
  lastUpdatedAt,
  selectedPeriod,
  selectedChannel,
  RefreshIcon,
  refreshing,
  handleRefresh,
  dashboardLoading,
  stats,
  formatNumber,
  ArrowUpIcon,
  ArrowDownIcon,
  periodText,
  kpiOverview,
  dashboardWarnings,
  visitTrendChart,
  accessSourceChart,
  channelFunnel,
  pendingTasks,
  deviceChart,
  regionChart,
  recentActivities,
  handleViewAllActivities,
  quickActions,
  handleQuickAction
} = useDashboard()
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

