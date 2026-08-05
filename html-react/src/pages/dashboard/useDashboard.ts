import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  getCount,
  getMonthlySales,
  getRecentActivities,
  getUserAccessSource,
  getWeeklyUserActivity,
} from '@/api/dashboard'
import { useAppStore, THEME_COLORS } from '@/stores/app'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import { entityField } from '@/utils/normalize'

export interface DashboardCounts {
  order_count_in_year: number
  today_visits: number
  role_count: number
  menu_count: number
}

export interface WeeklyActivityPoint {
  date: string
  visits: number
  users: number
}

export interface SourcePoint {
  name: string
  value: number
}

export interface MonthlyOperationPoint {
  month: string
  count: number
}

export interface DashboardActivity {
  user: string
  action: string
  time: string
  status: string
  type: string
  avatarColor: string
}

function asRecord(value: unknown): Record<string, unknown> {
  return (value && typeof value === 'object' ? value : {}) as Record<string, unknown>
}

export function useDashboard() {
  const { t } = useTranslation()
  const showError = useUnhandledError()
  const darkMode = useAppStore((s) => s.darkMode)
  const themeColor = useAppStore((s) => s.themeColor)

  const primaryColor = useMemo(
    () => THEME_COLORS.find((c) => c.key === themeColor)?.color || THEME_COLORS[0].color,
    [themeColor],
  )

  const [loading, setLoading] = useState(false)
  const [counts, setCounts] = useState<DashboardCounts>({
    order_count_in_year: 0,
    today_visits: 0,
    role_count: 0,
    menu_count: 0,
  })
  const [weeklyActivity, setWeeklyActivity] = useState<WeeklyActivityPoint[]>([])
  const [accessSource, setAccessSource] = useState<SourcePoint[]>([])
  const [monthlyOperations, setMonthlyOperations] = useState<MonthlyOperationPoint[]>([])
  const [activities, setActivities] = useState<DashboardActivity[]>([])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [countRes, sourceRes, weeklyRes, monthlyRes, activitiesRes] = await Promise.all([
        getCount(),
        getUserAccessSource().catch(() => null),
        getWeeklyUserActivity().catch(() => null),
        getMonthlySales().catch(() => null),
        getRecentActivities().catch(() => null),
      ])

      const countData = asRecord(countRes.data)
      setCounts({
        order_count_in_year: Number(entityField(countData, 'order_count_in_year', 0) ?? 0),
        today_visits: Number(entityField(countData, 'today_visits', 0) ?? 0),
        role_count: Number(entityField(countData, 'role_count', 0) ?? 0),
        menu_count: Number(entityField(countData, 'menu_count', 0) ?? 0),
      })

      const sourceRaw = (sourceRes as { data?: unknown } | null)?.data
      if (Array.isArray(sourceRaw)) {
        setAccessSource(
          sourceRaw.map((item) => {
            const record = asRecord(item)
            return {
              name: String(entityField(record, 'name', '') ?? ''),
              value: Number(entityField(record, 'value', 0) ?? 0),
            }
          }),
        )
      }

      const weeklyRaw = (weeklyRes as { data?: unknown } | null)?.data
      if (Array.isArray(weeklyRaw)) {
        setWeeklyActivity(
          weeklyRaw.map((item) => {
            const record = asRecord(item)
            return {
              date: String(entityField(record, 'date', '') ?? ''),
              visits: Number(entityField(record, 'visits', 0) ?? 0),
              users: Number(entityField(record, 'users', 0) ?? 0),
            }
          }),
        )
      }

      const monthlyRaw = (monthlyRes as { data?: unknown } | null)?.data
      if (Array.isArray(monthlyRaw)) {
        setMonthlyOperations(
          monthlyRaw.map((item) => {
            const record = asRecord(item)
            return {
              month: String(entityField(record, 'month', '') ?? ''),
              count: Number(entityField(record, 'count', 0) ?? 0),
            }
          }),
        )
      }

      const activitiesRaw = (activitiesRes as { data?: unknown } | null)?.data
      if (Array.isArray(activitiesRaw)) {
        setActivities(
          activitiesRaw.map((item) => {
            const record = asRecord(item)
            return {
              user: String(entityField(record, 'user', t('notification.system')) ?? ''),
              action: String(entityField(record, 'action', '') ?? ''),
              time: String(entityField(record, 'time', '') ?? ''),
              status: String(entityField(record, 'status', t('common.success')) ?? ''),
              type: String(entityField(record, 'type', 'success') ?? 'success'),
              avatarColor: String(entityField(record, 'avatarColor', primaryColor) ?? primaryColor),
            }
          }),
        )
      }
    } catch (error) {
      showError(error, t('common.query_failed'))
    } finally {
      setLoading(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- stable deps only
  }, [showError, t])

  useEffect(() => {
    void load()
  }, [load])

  return {
    loading,
    counts,
    weeklyActivity,
    accessSource,
    monthlyOperations,
    activities,
    darkMode,
    primaryColor,
    refresh: load,
  }
}
