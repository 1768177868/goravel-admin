export function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  const value = bytes / Math.pow(k, i)

  if (value >= 100) return `${Math.round(value)} ${sizes[i]}`
  if (value >= 10) return `${Math.round(value * 10) / 10} ${sizes[i]}`
  return `${Math.round(value * 100) / 100} ${sizes[i]}`
}

export function formatPercent(percent: number | null | undefined): string {
  if (percent === 0 || percent == null) return '0%'
  if (percent >= 100) return `${Math.round(percent)}%`
  if (percent >= 1) return `${percent.toFixed(1)}%`
  return `${percent.toFixed(2)}%`
}

export function formatPercentForProgress(percent: number): number {
  return Math.round(percent * 100) / 100
}

export function formatUptime(seconds: number): string {
  if (!seconds || seconds <= 0) return '0s'

  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  const parts: string[] = []
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  if (minutes > 0) parts.push(`${minutes}m`)
  if (secs > 0 || parts.length === 0) parts.push(`${secs}s`)
  return parts.join(' ')
}

export function getProgressColor(percentage: number): string {
  if (percentage < 50) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

export function formatNumber(num: number | null | undefined, decimals = 0): string {
  if (num == null) return '0'
  const value = Number(num)

  if (decimals > 0) {
    if (value < 1) return value.toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    if (value >= 100) return Math.round(value).toLocaleString()
    if (value >= 10) return (Math.round(value * 10) / 10).toFixed(1).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    return value.toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  }

  return value.toLocaleString()
}

export function formatLoad(load: number): string {
  if (load === 0) return '0'
  const value = Number(load)
  return value >= 10 ? value.toFixed(1) : value.toFixed(2)
}

export function formatDuration(milliseconds: number): string {
  if (!milliseconds || milliseconds <= 0) return '0ms'
  if (milliseconds < 1000) return `${Math.round(milliseconds)}ms`

  const seconds = milliseconds / 1000
  if (seconds < 60) return `${Math.round(seconds * 10) / 10}s`

  const minutes = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${minutes}m${secs}s`
}
