export function formatDateTime(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

export function getDaysAgo(days = 7, setToStartOfDay = true): string {
  const date = new Date()
  date.setDate(date.getDate() - days)
  if (setToStartOfDay) {
    date.setHours(0, 0, 0, 0)
  }
  return formatDateTime(date)
}

export function getSevenDaysAgo(): string {
  return getDaysAgo(7, true)
}
