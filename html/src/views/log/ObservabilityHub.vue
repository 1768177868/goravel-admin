<template>
  <div class="list-page observability-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.observability') }}</span>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <el-tab-pane :label="$t('observability.trace_tab')" name="trace">
          <div class="search-row">
            <el-input v-model="traceQuery.trace_id" :placeholder="$t('observability.trace_id_placeholder')" clearable />
            <el-button type="primary" :loading="traceLoading" @click="loadTrace">{{ $t('common.search') }}</el-button>
          </div>

          <el-empty v-if="!traceLoading && !traceData.trace_id" :description="$t('observability.trace_empty')" />
          <div v-else-if="traceData.trace_id">
            <el-descriptions border :column="2">
              <el-descriptions-item :label="$t('log.trace_id')">{{ traceData.trace_id }}</el-descriptions-item>
              <el-descriptions-item :label="$t('observability.trace_request')">{{ traceData.request?.method }} {{ traceData.request?.path }}</el-descriptions-item>
            </el-descriptions>

            <el-divider>{{ $t('observability.trace_operations') }}</el-divider>
            <el-table :data="traceData.operations || []" size="small" border>
              <el-table-column prop="created_at" :label="$t('log.time')" width="180" />
              <el-table-column prop="method" :label="$t('log.method')" width="90" />
              <el-table-column prop="path" :label="$t('log.path')" min-width="260" />
              <el-table-column prop="duration" :label="$t('observability.duration_ms')" width="120" />
              <el-table-column prop="title" :label="$t('log.title')" min-width="180" />
            </el-table>

            <el-divider>{{ $t('observability.trace_exceptions') }}</el-divider>
            <el-table :data="traceData.exceptions || []" size="small" border>
              <el-table-column prop="created_at" :label="$t('log.time')" width="180" />
              <el-table-column prop="level" :label="$t('log.level')" width="100" />
              <el-table-column prop="module" :label="$t('log.module')" width="140" />
              <el-table-column prop="message" :label="$t('log.message')" min-width="320" />
            </el-table>
          </div>
        </el-tab-pane>

        <el-tab-pane :label="$t('observability.slow_sql_tab')" name="slowSql">
          <div class="search-row">
            <el-input-number v-model="slowSqlQuery.hours" :min="1" :max="168" />
            <el-input-number v-model="slowSqlQuery.min_duration_ms" :min="1" :max="10000" />
            <el-input-number v-model="slowSqlQuery.limit" :min="1" :max="100" />
            <el-button type="primary" :loading="slowSqlLoading" @click="loadSlowSql">{{ $t('common.search') }}</el-button>
          </div>

          <el-table :data="slowSqlData" size="small" border>
            <el-table-column prop="max_duration_ms" :label="$t('observability.max_ms')" width="120" />
            <el-table-column prop="avg_duration_ms" :label="$t('observability.avg_ms')" width="120" />
            <el-table-column prop="count" :label="$t('common.count')" width="90" />
            <el-table-column prop="normalized_sql" :label="$t('observability.normalized_sql')" min-width="420">
              <template #default="{ row }">
                <code>{{ row.normalized_sql }}</code>
              </template>
            </el-table-column>
            <el-table-column prop="last_seen_at" :label="$t('observability.last_seen')" width="180" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="$t('observability.audit_tab')" name="audit">
          <div class="search-row">
            <el-input v-model="auditQuery.trace_id" :placeholder="$t('observability.trace_id_placeholder')" clearable />
            <el-input v-model="auditQuery.keyword" :placeholder="$t('observability.keyword_placeholder')" clearable />
            <el-button type="primary" :loading="auditLoading" @click="loadAudit">{{ $t('common.search') }}</el-button>
          </div>

          <el-table :data="auditData.list" size="small" border>
            <el-table-column prop="time" :label="$t('log.time')" width="180" />
            <el-table-column prop="type" :label="$t('observability.event_type')" width="110" />
            <el-table-column prop="trace_id" :label="$t('log.trace_id')" width="220" />
            <el-table-column prop="module" :label="$t('log.module')" width="130" />
            <el-table-column prop="message" :label="$t('log.message')" min-width="340" />
          </el-table>

          <div class="pager">
            <el-pagination
              v-model:current-page="auditQuery.page"
              v-model:page-size="auditQuery.page_size"
              :total="auditData.total"
              layout="total, prev, pager, next"
              @current-change="loadAudit"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getAuditTimeline, getSlowSqlTop, getTraceAggregate } from '../../api/observability'

const activeTab = ref('trace')

const traceLoading = ref(false)
const traceQuery = reactive({ trace_id: '' })
const traceData = reactive({})

const slowSqlLoading = ref(false)
const slowSqlQuery = reactive({ hours: 24, min_duration_ms: 200, limit: 20 })
const slowSqlData = ref([])

const auditLoading = ref(false)
const auditQuery = reactive({ trace_id: '', keyword: '', page: 1, page_size: 20 })
const auditData = reactive({ list: [], total: 0 })

const loadTrace = async () => {
  if (!traceQuery.trace_id) {
    ElMessage.warning('trace_id is required')
    return
  }
  traceLoading.value = true
  try {
    const res = await getTraceAggregate(traceQuery)
    Object.keys(traceData).forEach(key => delete traceData[key])
    Object.assign(traceData, res.data || {})
  } finally {
    traceLoading.value = false
  }
}

const loadSlowSql = async () => {
  slowSqlLoading.value = true
  try {
    const res = await getSlowSqlTop(slowSqlQuery)
    slowSqlData.value = res.data?.list || []
  } finally {
    slowSqlLoading.value = false
  }
}

const loadAudit = async () => {
  auditLoading.value = true
  try {
    const res = await getAuditTimeline(auditQuery)
    auditData.list = res.data?.list || []
    auditData.total = res.data?.total || 0
  } finally {
    auditLoading.value = false
  }
}

onMounted(() => {
  loadSlowSql()
  loadAudit()
})
</script>

<style scoped>
.search-row {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
