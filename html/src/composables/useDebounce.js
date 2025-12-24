import { ref, watch } from 'vue'

/**
 * 防抖 composable
 * @param {Function} fn - 需要防抖的函数
 * @param {Number} delay - 延迟时间（毫秒），默认 300ms
 * @param {Boolean} immediate - 是否立即执行，默认 false
 * @returns {Function} 防抖后的函数
 */
export function useDebounce(fn, delay = 300, immediate = false) {
  let timer = null
  let isInvoked = false

  const debouncedFn = function (...args) {
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
  }

  // 取消防抖
  debouncedFn.cancel = () => {
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
 * @param {any} source - 源值（ref 或 computed）
 * @param {Number} delay - 延迟时间（毫秒），默认 300ms
 * @returns {Ref} 防抖后的响应式值
 */
export function useDebouncedRef(source, delay = 300) {
  const debouncedValue = ref(source.value)

  watch(
    source,
    (newValue) => {
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

