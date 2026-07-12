/**
 * Extract list/tree array from common admin API response shapes.
 */
export function extractListFromResponse(res) {
  const data = res?.data
  if (!data) {
    return []
  }
  if (Array.isArray(data)) {
    return data
  }
  if (Array.isArray(data.list)) {
    return data.list
  }
  if (Array.isArray(data.data)) {
    return data.data
  }
  if (Array.isArray(data.menus)) {
    return data.menus
  }
  return []
}
