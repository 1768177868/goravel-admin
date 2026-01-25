import { ref, onMounted, onUnmounted } from 'vue'

/**
 * 响应式断点配置
 */
export const breakpoints = {
  xs: 480,   // 手机竖屏
  sm: 768,   // 平板竖屏/手机横屏
  md: 992,   // 平板横屏
  lg: 1200,  // 小桌面
  xl: 1920   // 大桌面
}

/**
 * 响应式工具 Composable
 * 提供设备类型检测和窗口大小监听
 */
export function useResponsive() {
  const windowWidth = ref(window.innerWidth)
  const windowHeight = ref(window.innerHeight)

  const isMobile = ref(window.innerWidth < breakpoints.sm)
  const isTablet = ref(window.innerWidth >= breakpoints.sm && window.innerWidth < breakpoints.md)
  const isDesktop = ref(window.innerWidth >= breakpoints.md)
  
  const isXs = ref(window.innerWidth < breakpoints.xs)
  const isSm = ref(window.innerWidth >= breakpoints.xs && window.innerWidth < breakpoints.sm)
  const isMd = ref(window.innerWidth >= breakpoints.sm && window.innerWidth < breakpoints.md)
  const isLg = ref(window.innerWidth >= breakpoints.md && window.innerWidth < breakpoints.lg)
  const isXl = ref(window.innerWidth >= breakpoints.lg)

  const updateSize = () => {
    const width = window.innerWidth
    const height = window.innerHeight
    
    windowWidth.value = width
    windowHeight.value = height
    
    isMobile.value = width < breakpoints.sm
    isTablet.value = width >= breakpoints.sm && width < breakpoints.md
    isDesktop.value = width >= breakpoints.md
    
    isXs.value = width < breakpoints.xs
    isSm.value = width >= breakpoints.xs && width < breakpoints.sm
    isMd.value = width >= breakpoints.sm && width < breakpoints.md
    isLg.value = width >= breakpoints.md && width < breakpoints.lg
    isXl.value = width >= breakpoints.lg
  }

  // 防抖处理
  let resizeTimer = null
  const handleResize = () => {
    if (resizeTimer) {
      clearTimeout(resizeTimer)
    }
    resizeTimer = setTimeout(() => {
      updateSize()
    }, 100)
  }

  onMounted(() => {
    window.addEventListener('resize', handleResize)
    updateSize()
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
    if (resizeTimer) {
      clearTimeout(resizeTimer)
    }
  })

  return {
    windowWidth,
    windowHeight,
    isMobile,
    isTablet,
    isDesktop,
    isXs,
    isSm,
    isMd,
    isLg,
    isXl,
    breakpoints
  }
}

/**
 * 检查是否匹配断点
 * @param {string} breakpoint - 断点名称 (xs, sm, md, lg, xl)
 * @param {string} direction - 方向 ('up' | 'down' | 'only')
 */
export function useBreakpoint(breakpoint, direction = 'up') {
  const { windowWidth, breakpoints: bp } = useResponsive()
  
  const checkBreakpoint = () => {
    const width = windowWidth.value
    const bpValue = bp[breakpoint]
    
    if (!bpValue) {
      console.warn(`Unknown breakpoint: ${breakpoint}`)
      return false
    }
    
    switch (direction) {
      case 'up':
        return width >= bpValue
      case 'down':
        return width < bpValue
      case 'only':
        // 获取当前断点的范围
        const bpKeys = Object.keys(bp).sort((a, b) => bp[a] - bp[b])
        const currentIndex = bpKeys.indexOf(breakpoint)
        if (currentIndex === -1) return false
        
        const min = bp[bpKeys[currentIndex]]
        const max = currentIndex < bpKeys.length - 1 ? bp[bpKeys[currentIndex + 1]] : Infinity
        
        return width >= min && width < max
      default:
        return false
    }
  }
  
  return checkBreakpoint()
}
