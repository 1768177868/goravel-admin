<template>
  <el-select
    class="timezone-switch"
    v-model="selectedTimezone"
    size="small"
    filterable
    allow-create
    default-first-option
    :placeholder="$t('header.timezone')"
  >
    <el-option
      v-for="option in timezoneOptions"
      :key="option.value"
      :label="option.label"
      :value="option.value"
    />
  </el-select>
</template>

<script setup>
import { computed, reactive } from 'vue'
import { useAppStore } from '../store/app'

const appStore = useAppStore()

const presetTimezones = [
  { value: 'Pacific/Honolulu', offset: '-10:00', label: 'UTC-10:00 (Pacific/Honolulu)' },
  { value: 'America/Anchorage', offset: '-09:00', label: 'UTC-09:00 (America/Anchorage)' },
  { value: 'America/Los_Angeles', offset: '-08:00', label: 'UTC-08:00 (America/Los_Angeles)' },
  { value: 'America/Denver', offset: '-07:00', label: 'UTC-07:00 (America/Denver)' },
  { value: 'America/Chicago', offset: '-06:00', label: 'UTC-06:00 (America/Chicago)' },
  { value: 'America/New_York', offset: '-05:00', label: 'UTC-05:00 (America/New_York)' },
  { value: 'America/Sao_Paulo', offset: '-03:00', label: 'UTC-03:00 (America/Sao_Paulo)' },
  { value: 'UTC', offset: '+00:00', label: 'UTC+00:00 (UTC)' },
  { value: 'Europe/Berlin', offset: '+01:00', label: 'UTC+01:00 (Europe/Berlin)' },
  { value: 'Europe/Moscow', offset: '+03:00', label: 'UTC+03:00 (Europe/Moscow)' },
  { value: 'Asia/Dubai', offset: '+04:00', label: 'UTC+04:00 (Asia/Dubai)' },
  { value: 'Asia/Kolkata', offset: '+05:30', label: 'UTC+05:30 (Asia/Kolkata)' },
  { value: 'Asia/Bangkok', offset: '+07:00', label: 'UTC+07:00 (Asia/Bangkok)' },
  { value: 'Asia/Shanghai', offset: '+08:00', label: 'UTC+08:00 (Asia/Shanghai)' },
  { value: 'Asia/Tokyo', offset: '+09:00', label: 'UTC+09:00 (Asia/Tokyo)' },
  { value: 'Australia/Sydney', offset: '+10:00', label: 'UTC+10:00 (Australia/Sydney)' },
  { value: 'Pacific/Auckland', offset: '+12:00', label: 'UTC+12:00 (Pacific/Auckland)' }
]

const timezoneMap = reactive(new Map(presetTimezones.map(item => [item.value, item.label])))

const formatOffsetLabel = (tz) => {
  if (!tz) {
    return ''
  }

  if (timezoneMap.has(tz)) {
    return timezoneMap.get(tz)
  }

  try {
    const formatter = new Intl.DateTimeFormat('en-US', {
      timeZone: tz,
      timeZoneName: 'short',
      hour: '2-digit',
      minute: '2-digit'
    })
    const parts = formatter.formatToParts(new Date())
    const tzName = parts.find(part => part.type === 'timeZoneName')?.value || ''
    const match = tzName.match(/([+-]\d{1,2})(?::(\d{2}))?/)
    if (match) {
      const sign = match[1].startsWith('-') ? '-' : '+'
      const hours = Math.abs(parseInt(match[1], 10)).toString().padStart(2, '0')
      const minutes = (match[2] || '00').padStart(2, '0')
      return `UTC${sign}${hours}:${minutes} (${tz})`
    }
  } catch {
    // ignore errors and fall back to raw name
  }

  return tz
}

const ensureTimezoneIncluded = (tz) => {
  if (!tz) {
    return
  }
  if (!timezoneMap.has(tz)) {
    timezoneMap.set(tz, formatOffsetLabel(tz))
  }
}

ensureTimezoneIncluded(appStore.timezone)

const timezoneOptions = computed(() => {
  return Array.from(timezoneMap.entries())
    .map(([value, label]) => ({ value, label }))
    .sort((a, b) => a.label.localeCompare(b.label))
})

const selectedTimezone = computed({
  get: () => appStore.timezone,
  set: (val) => {
    ensureTimezoneIncluded(val)
    appStore.setTimezone(val)
  }
})
</script>

<style scoped>
.timezone-switch {
  width: 210px;
}
</style>


