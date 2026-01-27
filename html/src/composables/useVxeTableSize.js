import { computed } from 'vue'
import { useAppStore } from '../store/app'

/**
 * VXE Table 大小管理 composable
 * 统一管理 vxe-table 的 size 属性，响应布局大小设置
 * 
 * @returns {Object} 包含 vxeSize 计算属性
 * 
 * @example
 * import { useVxeTableSize } from '@/composables/useVxeTableSize'
 * 
 * const { vxeSize } = useVxeTableSize()
 * 
 * // 在模板中使用
 * <vxe-table :size="vxeSize" ...>
 */
export function useVxeTableSize() {
  const appStore = useAppStore()

  // 映射布局大小到 vxe-table 的 size
  // vxe-table 支持: 'medium', 'small', 'mini'
  // 注意：vxe-table 没有 'large'，最大尺寸就是 'medium'
  const vxeSize = computed(() => {
    const size = appStore.layoutSize
    if (size === 'small') {
      return 'small'
    } else if (size === 'default') {
      return 'medium'
    } else if (size === 'large') {
      return 'medium' // vxe-table 没有 large，使用 medium 作为最大尺寸
    }
    return 'medium' // 默认值
  })

  return {
    vxeSize
  }
}
