import { ref, watch, type Ref, type WatchSource } from 'vue'

/**
 * 防抖函数接口
 */
export interface DebouncedFunction<T extends (...args: any[]) => any> {
  (...args: Parameters<T>): void
  /** 取消待执行的防抖函数 */
  cancel: () => void
}

/**
 * 防抖 composable
 * @param fn - 需要防抖的函数
 * @param delay - 延迟时间（毫秒），默认 300ms
 * @param immediate - 是否立即执行，默认 false
 * @returns 防抖后的函数
 *
 * @example
 * const debouncedSearch = useDebounce((keyword: string) => {
 *   console.log('Searching:', keyword)
 * }, 300)
 *
 * // 使用
 * debouncedSearch('hello')
 *
 * // 取消
 * debouncedSearch.cancel()
 */
export function useDebounce<T extends (...args: any[]) => any>(
  fn: T,
  delay: number = 300,
  immediate: boolean = false
): DebouncedFunction<T> {
  let timer: ReturnType<typeof setTimeout> | null = null
  let isInvoked = false

  const debouncedFn = function (this: any, ...args: Parameters<T>): void {
    const context = this

    // 清除之前的定时器
    if (timer) {
      clearTimeout(timer)
    }

    // 如果立即执行且还未执行过
    if (immediate && !isInvoked) {
      fn.apply(context, args)
      isInvoked = true
    }

    // 设置新的定时器
    timer = setTimeout(() => {
      if (!immediate) {
        fn.apply(context, args)
      }
      isInvoked = false
      timer = null
    }, delay)
  } as DebouncedFunction<T>

  // 取消防抖
  debouncedFn.cancel = (): void => {
    if (timer) {
      clearTimeout(timer)
      timer = null
      isInvoked = false
    }
  }

  return debouncedFn
}

/**
 * 防抖响应式值
 * @param source - 源值（ref 或 computed）
 * @param delay - 延迟时间（毫秒），默认 300ms
 * @returns 防抖后的响应式值
 *
 * @example
 * const searchKeyword = ref('')
 * const debouncedKeyword = useDebouncedRef(searchKeyword, 500)
 *
 * // 当 searchKeyword 变化时，debouncedKeyword 会延迟 500ms 更新
 */
export function useDebouncedRef<T>(source: Ref<T>, delay: number = 300): Ref<T> {
  const debouncedValue = ref(source.value) as Ref<T>

  watch(
    source as WatchSource<T>,
    (newValue: T) => {
      const timer = setTimeout(() => {
        debouncedValue.value = newValue
      }, delay)

      return () => {
        clearTimeout(timer)
      }
    },
    { immediate: true }
  )

  return debouncedValue
}

