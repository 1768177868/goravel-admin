/**
 * 日期时间工具函数
 */

/**
 * 格式化日期时间为 YYYY-MM-DD HH:mm:ss
 * @param {Date} date - 日期对象
 * @returns {string} 格式化后的日期时间字符串
 */
export function formatDateTime(date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

/**
 * 获取 N 天前的日期时间（用于默认开始时间）
 * @param {number} days - 天数，默认 7 天
 * @param {boolean} setToStartOfDay - 是否设置为当天的 00:00:00，默认 true
 * @returns {string} 格式化后的日期时间字符串
 */
export function getDaysAgo(days = 7, setToStartOfDay = true) {
  const date = new Date()
  date.setDate(date.getDate() - days)
  if (setToStartOfDay) {
    date.setHours(0, 0, 0, 0) // 设置为当天的00:00:00
  }
  return formatDateTime(date)
}

/**
 * 获取 N 个月前的日期时间（用于默认开始时间）
 * @param {number} months - 月数，默认 1 个月
 * @param {boolean} setToStartOfDay - 是否设置为当天的 00:00:00，默认 true
 * @returns {string} 格式化后的日期时间字符串
 */
export function getMonthsAgo(months = 1, setToStartOfDay = true) {
  const date = new Date()
  date.setMonth(date.getMonth() - months)
  if (setToStartOfDay) {
    date.setHours(0, 0, 0, 0) // 设置为当天的00:00:00
  }
  return formatDateTime(date)
}

/**
 * 获取 N 年前的日期时间（用于默认开始时间）
 * @param {number} years - 年数，默认 1 年
 * @param {boolean} setToStartOfDay - 是否设置为当天的 00:00:00，默认 true
 * @returns {string} 格式化后的日期时间字符串
 */
export function getYearsAgo(years = 1, setToStartOfDay = true) {
  const date = new Date()
  date.setFullYear(date.getFullYear() - years)
  if (setToStartOfDay) {
    date.setHours(0, 0, 0, 0) // 设置为当天的00:00:00
  }
  return formatDateTime(date)
}

/**
 * 获取指定时间单位前的日期时间（通用方法）
 * @param {Object} options - 配置选项
 * @param {number} options.days - 天数
 * @param {number} options.months - 月数
 * @param {number} options.years - 年数
 * @param {boolean} options.setToStartOfDay - 是否设置为当天的 00:00:00，默认 true
 * @returns {string} 格式化后的日期时间字符串
 * @example
 * getTimeAgo({ days: 7 }) // 7天前
 * getTimeAgo({ months: 1 }) // 1个月前
 * getTimeAgo({ years: 1 }) // 1年前
 * getTimeAgo({ days: 7, months: 1 }) // 1个月零7天前
 */
export function getTimeAgo({ days = 0, months = 0, years = 0, setToStartOfDay = true } = {}) {
  const date = new Date()
  
  if (years > 0) {
    date.setFullYear(date.getFullYear() - years)
  }
  if (months > 0) {
    date.setMonth(date.getMonth() - months)
  }
  if (days > 0) {
    date.setDate(date.getDate() - days)
  }
  
  if (setToStartOfDay) {
    date.setHours(0, 0, 0, 0) // 设置为当天的00:00:00
  }
  
  return formatDateTime(date)
}

/**
 * 获取7天前的日期时间（便捷方法）
 * @returns {string} 格式化后的日期时间字符串
 */
export function getSevenDaysAgo() {
  return getDaysAgo(7, true)
}

/**
 * 获取1个月前的日期时间（便捷方法）
 * @returns {string} 格式化后的日期时间字符串
 */
export function getOneMonthAgo() {
  return getMonthsAgo(1, true)
}

/**
 * 获取3个月前的日期时间（便捷方法）
 * @returns {string} 格式化后的日期时间字符串
 */
export function getThreeMonthsAgo() {
  return getMonthsAgo(3, true)
}

