import { useListPage } from './useListPage'

/**
 * useListPage 的语义化别名，保持与历史页面兼容。
 * 新页面建议优先使用 useTable，老页面可继续使用 useListPage。
 */
export function useTable(options = {}) {
  return useListPage(options)
}

