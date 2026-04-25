<template>
  <el-select
    class="timezone-switch"
    v-model="selectedTimezone"
    size="small"
    filterable
    allow-create
    default-first-option
    placement="bottom-end"
    :offset="8"
    popper-class="timezone-select-popper"
    :teleported="true"
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

// 预设覆盖每个整点时区（UTC-12 ~ UTC+14）
const presetTimezones = [
  { value: 'Etc/GMT+12', offset: '-12:00', label: 'UTC-12:00 (Etc/GMT+12)' },
  { value: 'Pacific/Pago_Pago', offset: '-11:00', label: 'UTC-11:00 (Pacific/Pago_Pago)' },
  { value: 'Pacific/Honolulu', offset: '-10:00', label: 'UTC-10:00 (Pacific/Honolulu)' },
  { value: 'America/Anchorage', offset: '-09:00', label: 'UTC-09:00 (America/Anchorage)' },
  { value: 'America/Los_Angeles', offset: '-08:00', label: 'UTC-08:00 (America/Los_Angeles)' },
  { value: 'America/Denver', offset: '-07:00', label: 'UTC-07:00 (America/Denver)' },
  { value: 'America/Chicago', offset: '-06:00', label: 'UTC-06:00 (America/Chicago)' },
  { value: 'America/New_York', offset: '-05:00', label: 'UTC-05:00 (America/New_York)' },
  { value: 'America/Halifax', offset: '-04:00', label: 'UTC-04:00 (America/Halifax)' },
  { value: 'America/Argentina/Buenos_Aires', offset: '-03:00', label: 'UTC-03:00 (America/Argentina/Buenos_Aires)' },
  { value: 'Etc/GMT+2', offset: '-02:00', label: 'UTC-02:00 (Etc/GMT+2)' },
  { value: 'Atlantic/Azores', offset: '-01:00', label: 'UTC-01:00 (Atlantic/Azores)' },
  { value: 'UTC', offset: '+00:00', label: 'UTC+00:00 (UTC)' },
  { value: 'Europe/Berlin', offset: '+01:00', label: 'UTC+01:00 (Europe/Berlin)' },
  { value: 'Europe/Athens', offset: '+02:00', label: 'UTC+02:00 (Europe/Athens)' },
  { value: 'Europe/Moscow', offset: '+03:00', label: 'UTC+03:00 (Europe/Moscow)' },
  { value: 'Asia/Dubai', offset: '+04:00', label: 'UTC+04:00 (Asia/Dubai)' },
  { value: 'Asia/Karachi', offset: '+05:00', label: 'UTC+05:00 (Asia/Karachi)' },
  { value: 'Asia/Dhaka', offset: '+06:00', label: 'UTC+06:00 (Asia/Dhaka)' },
  { value: 'Asia/Bangkok', offset: '+07:00', label: 'UTC+07:00 (Asia/Bangkok)' },
  { value: 'Asia/Shanghai', offset: '+08:00', label: 'UTC+08:00 (Asia/Shanghai)' },
  { value: 'Asia/Tokyo', offset: '+09:00', label: 'UTC+09:00 (Asia/Tokyo)' },
  { value: 'Australia/Sydney', offset: '+10:00', label: 'UTC+10:00 (Australia/Sydney)' },
  { value: 'Pacific/Noumea', offset: '+11:00', label: 'UTC+11:00 (Pacific/Noumea)' },
  { value: 'Pacific/Auckland', offset: '+12:00', label: 'UTC+12:00 (Pacific/Auckland)' },
  { value: 'Pacific/Enderbury', offset: '+13:00', label: 'UTC+13:00 (Pacific/Enderbury)' },
  { value: 'Pacific/Kiritimati', offset: '+14:00', label: 'UTC+14:00 (Pacific/Kiritimati)' }
]

const timezoneMap = reactive(new Map(presetTimezones.map(item => [item.value, item.label])))

const parseOffsetMinutes = (label) => {
  if (!label) return Number.POSITIVE_INFINITY
  const match = label.match(/^UTC([+-])(\d{2}):(\d{2})/)
  if (!match) return Number.POSITIVE_INFINITY
  const sign = match[1] === '-' ? -1 : 1
  const hours = Number.parseInt(match[2], 10)
  const minutes = Number.parseInt(match[3], 10)
  return sign * (hours * 60 + minutes)
}

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
    .sort((a, b) => {
      const offsetDiff = parseOffsetMinutes(a.label) - parseOffsetMinutes(b.label)
      if (offsetDiff !== 0) return offsetDiff
      return a.label.localeCompare(b.label)
    })
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
  width: min(240px, 100%);
  min-width: 0;
}

.timezone-switch :deep(.el-select__selection) {
  overflow: hidden;
}

.timezone-switch :deep(.el-select__selected-item) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.timezone-switch :deep(.el-input__wrapper) {
  border-radius: 10px;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--border-color-light) 75%, transparent) inset;
}

.timezone-switch :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--el-color-primary) 35%, transparent) inset;
}

.timezone-switch :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset, 0 0 0 3px color-mix(in srgb, var(--el-color-primary) 18%, transparent);
}
</style>

<style>
/* 时区下拉：与布局/账号等弹层统一质感 */
.el-popper.timezone-select-popper {
  border-radius: 12px !important;
  border: 1px solid color-mix(in srgb, var(--border-color-light) 70%, transparent) !important;
  background: color-mix(in srgb, var(--card-bg, #fff) 98%, transparent) !important;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.14), 0 2px 8px rgba(15, 23, 42, 0.08) !important;
  padding: 6px 0 !important;
  overflow: hidden;
  backdrop-filter: blur(8px);
  min-width: 280px !important;
  max-width: min(420px, 92vw);
}

.el-popper.timezone-select-popper .el-select-dropdown__wrap {
  max-height: min(380px, 52vh);
}

.el-popper.timezone-select-popper .el-select-dropdown__list {
  padding: 4px 6px;
}

.el-popper.timezone-select-popper .el-select-dropdown__item {
  border-radius: 8px;
  margin: 2px 0;
  padding: 0 12px;
  min-height: 36px;
  line-height: 36px;
  font-size: 13px;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.el-popper.timezone-select-popper .el-select-dropdown__item.is-hovering,
.el-popper.timezone-select-popper .el-select-dropdown__item:hover {
  background-color: color-mix(in srgb, var(--el-color-primary) 10%, transparent) !important;
}

.el-popper.timezone-select-popper .el-select-dropdown__item.is-selected {
  font-weight: 600;
  color: var(--el-color-primary) !important;
  background: color-mix(in srgb, var(--el-color-primary) 12%, transparent) !important;
}

.el-popper.timezone-select-popper .el-select-dropdown__empty {
  padding: 12px 16px;
  font-size: 13px;
}

.el-popper.timezone-select-popper .el-scrollbar__bar {
  z-index: 2;
}

html.dark .el-popper.timezone-select-popper {
  border-color: rgba(255, 255, 255, 0.12) !important;
  background: color-mix(in srgb, var(--card-bg, #1d1e1f) 96%, transparent) !important;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.5), 0 2px 8px rgba(0, 0, 0, 0.35) !important;
}

html.dark .el-popper.timezone-select-popper .el-select-dropdown__item.is-hovering,
html.dark .el-popper.timezone-select-popper .el-select-dropdown__item:hover {
  background: rgba(255, 255, 255, 0.08) !important;
}

html.dark .el-popper.timezone-select-popper .el-select-dropdown__item.is-selected {
  background: color-mix(in srgb, var(--el-color-primary) 22%, rgba(255, 255, 255, 0.06)) !important;
}
</style>


