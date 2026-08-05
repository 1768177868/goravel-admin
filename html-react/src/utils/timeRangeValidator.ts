export interface TimeRangeValidation {
  valid: boolean
  error?: string
  errorKey?: string
  errorParams?: Record<string, unknown>
}

export function validateTimeRange(
  startTime: string | Date | undefined | null,
  endTime: string | Date | undefined | null,
  maxMonths = 3,
): TimeRangeValidation {
  if (!startTime || !endTime) {
    return { valid: true }
  }

  const start = typeof startTime === 'string' ? new Date(startTime) : startTime
  const end = typeof endTime === 'string' ? new Date(endTime) : endTime

  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
    return { valid: false, error: 'Invalid date' }
  }

  if (start > end) {
    return { valid: false, errorKey: 'start_time_after_end_time' }
  }

  const monthsDiff = (end.getFullYear() - start.getFullYear()) * 12 + (end.getMonth() - start.getMonth())
  if (monthsDiff >= maxMonths) {
    return {
      valid: false,
      errorKey: 'time_range_exceeded',
      errorParams: { months: maxMonths },
    }
  }

  return { valid: true }
}

export const OPERATION_LOG_MAX_TIME_RANGE_MONTHS = 3
