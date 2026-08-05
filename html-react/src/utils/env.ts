/** Resolve API base URL from Vite env (same contract as Vue frontend). */
export function getApiBaseURL(): string {
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL
  const apiPrefix = import.meta.env.VITE_API_PREFIX || '/api/admin'

  if (apiBaseURL) {
    const base = apiBaseURL.replace(/\/+$/, '')
    const prefix = apiPrefix.startsWith('/') ? apiPrefix : `/${apiPrefix}`
    return `${base}${prefix}`
  }

  return apiPrefix
}

export function getWsBaseURL(): string {
  const apiBaseURL = import.meta.env.VITE_API_BASE_URL || `${window.location.protocol}//${window.location.host}`
  return apiBaseURL.replace(/^http/, 'ws').replace(/\/+$/, '')
}
