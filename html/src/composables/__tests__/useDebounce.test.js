import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref, nextTick } from 'vue'
import { useDebounce, useDebouncedRef } from '../useDebounce'

describe('useDebounce', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('基本防抖功能', () => {
    it('应该在延迟后调用函数', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn, 100)

      debouncedFn()
      expect(fn).not.toHaveBeenCalled()

      vi.advanceTimersByTime(100)
      expect(fn).toHaveBeenCalledTimes(1)
    })

    it('应该只在最后一次调用后执行', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn, 100)

      debouncedFn('a')
      debouncedFn('b')
      debouncedFn('c')

      vi.advanceTimersByTime(100)
      expect(fn).toHaveBeenCalledTimes(1)
      expect(fn).toHaveBeenCalledWith('c')
    })

    it('应该在连续调用时重置定时器', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn, 100)

      debouncedFn()
      vi.advanceTimersByTime(50)
      debouncedFn()
      vi.advanceTimersByTime(50)
      debouncedFn()
      vi.advanceTimersByTime(50)

      expect(fn).not.toHaveBeenCalled()

      vi.advanceTimersByTime(50)
      expect(fn).toHaveBeenCalledTimes(1)
    })

    it('应该使用默认延迟 300ms', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn)

      debouncedFn()
      vi.advanceTimersByTime(299)
      expect(fn).not.toHaveBeenCalled()

      vi.advanceTimersByTime(1)
      expect(fn).toHaveBeenCalledTimes(1)
    })
  })

  describe('立即执行模式', () => {
    it('应该在首次调用时立即执行', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn, 100, true)

      debouncedFn()
      expect(fn).toHaveBeenCalledTimes(1)
    })

    it('应该在立即执行后的延迟期间不再执行', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn, 100, true)

      debouncedFn()
      debouncedFn()
      debouncedFn()

      expect(fn).toHaveBeenCalledTimes(1)
    })

    it('应该在延迟结束后重置状态', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn, 100, true)

      debouncedFn()
      expect(fn).toHaveBeenCalledTimes(1)

      vi.advanceTimersByTime(100)

      debouncedFn()
      expect(fn).toHaveBeenCalledTimes(2)
    })
  })

  describe('取消功能', () => {
    it('应该能够取消待执行的函数', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn, 100)

      debouncedFn()
      debouncedFn.cancel()
      vi.advanceTimersByTime(100)

      expect(fn).not.toHaveBeenCalled()
    })

    it('多次取消应该是安全的', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn, 100)

      debouncedFn()
      debouncedFn.cancel()
      debouncedFn.cancel()
      debouncedFn.cancel()

      vi.advanceTimersByTime(100)
      expect(fn).not.toHaveBeenCalled()
    })
  })

  describe('参数传递', () => {
    it('应该正确传递参数', () => {
      const fn = vi.fn()
      const debouncedFn = useDebounce(fn, 100)

      debouncedFn('arg1', 'arg2', { key: 'value' })
      vi.advanceTimersByTime(100)

      expect(fn).toHaveBeenCalledWith('arg1', 'arg2', { key: 'value' })
    })
  })
})

describe('useDebouncedRef', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('应该创建带有初始值的防抖 ref', async () => {
    const source = ref('initial')
    const debounced = useDebouncedRef(source, 100)

    expect(debounced.value).toBe('initial')
  })

  it('应该在延迟后更新值', async () => {
    const source = ref('initial')
    const debounced = useDebouncedRef(source, 100)

    source.value = 'updated'
    await nextTick()

    // 还未到延迟时间
    expect(debounced.value).toBe('initial')

    vi.advanceTimersByTime(100)
    await nextTick()

    expect(debounced.value).toBe('updated')
  })
})

