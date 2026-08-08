import { useEffect, useState } from 'react'

/** Breakpoints aligned with Vue `useResponsive`. */
export const breakpoints = {
  xs: 480,
  sm: 768,
  md: 992,
  lg: 1200,
  xl: 1920,
} as const

function readSize() {
  const width = typeof window !== 'undefined' ? window.innerWidth : breakpoints.md
  const height = typeof window !== 'undefined' ? window.innerHeight : 0
  return {
    windowWidth: width,
    windowHeight: height,
    isMobile: width < breakpoints.sm,
    isTablet: width >= breakpoints.sm && width < breakpoints.md,
    isDesktop: width >= breakpoints.md,
    isXs: width < breakpoints.xs,
    isSm: width >= breakpoints.xs && width < breakpoints.sm,
    isMd: width >= breakpoints.sm && width < breakpoints.md,
    isLg: width >= breakpoints.md && width < breakpoints.lg,
    isXl: width >= breakpoints.lg,
  }
}

export function useResponsive() {
  const [size, setSize] = useState(readSize)

  useEffect(() => {
    let resizeTimer: ReturnType<typeof setTimeout> | null = null
    const handleResize = () => {
      if (resizeTimer) clearTimeout(resizeTimer)
      resizeTimer = setTimeout(() => {
        setSize(readSize())
      }, 100)
    }

    window.addEventListener('resize', handleResize)
    setSize(readSize())

    return () => {
      window.removeEventListener('resize', handleResize)
      if (resizeTimer) clearTimeout(resizeTimer)
    }
  }, [])

  return {
    ...size,
    breakpoints,
  }
}
