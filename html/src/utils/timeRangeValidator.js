/**
 * 时间范围验证工具
 */

/**
 * 验证时间范围是否超过指定月数
 * @param {string|Date} startTime - 开始时间（字符串格式：YYYY-MM-DD HH:mm:ss 或 Date 对象）
 * @param {string|Date} endTime - 结束时间（字符串格式：YYYY-MM-DD HH:mm:ss 或 Date 对象）
 * @param {number} maxMonths - 最大允许的月数，默认 3 个月
 * @returns {{ valid: boolean, error?: string }} 验证结果
 */
export function validateTimeRange(startTime, endTime, maxMonths = 3) {
  if (!startTime || !endTime) {
    return { valid: true } // 如果时间未填写，不验证
  }

  // 转换为 Date 对象
  const start = typeof startTime === 'string' ? new Date(startTime) : startTime
  const end = typeof endTime === 'string' ? new Date(endTime) : endTime

  // 检查日期是否有效
  if (isNaN(start.getTime()) || isNaN(end.getTime())) {
    return { valid: false, error: '时间格式无效' }
  }

  // 检查开始时间是否晚于结束时间
  if (start > end) {
    return { valid: false, error: '开始时间不能晚于结束时间' }
  }

  // 计算时间差（月数）
  const monthsDiff = (end.getFullYear() - start.getFullYear()) * 12 + (end.getMonth() - start.getMonth())
  
  // 如果结束时间的日期大于开始时间的日期，可能需要额外考虑
  if (end.getDate() > start.getDate()) {
    // 这里简化处理，如果日期差较大，可能跨月
    // 但主要看月份差
  }

  // 检查是否超过最大月数
  if (monthsDiff >= maxMonths) {
    return { valid: false, error: `查询时间范围不能超过${maxMonths}个月` }
  }

  return { valid: true }
}

/**
 * 默认最大时间范围（月数）- 用于订单查询
 */
export const DEFAULT_MAX_TIME_RANGE_MONTHS = 3

/**
 * 订单查询的最大时间范围（月数）
 */
export const ORDER_MAX_TIME_RANGE_MONTHS = 3

/**
 * 操作日志查询的最大时间范围（月数）
 */
export const OPERATION_LOG_MAX_TIME_RANGE_MONTHS = 3

