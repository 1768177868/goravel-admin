import { describe, it, expect, beforeEach, vi } from 'vitest'
import { Storage } from '../storage'

// Mock localStorage
const localStorageMock = (() => {
  let store = {}
  return {
    getItem: vi.fn((key) => store[key] || null),
    setItem: vi.fn((key, value) => {
      store[key] = value
    }),
    removeItem: vi.fn((key) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    }),
    get length() {
      return Object.keys(store).length
    },
    key: vi.fn((index) => Object.keys(store)[index] || null)
  }
})()

// Mock logger
vi.mock('../logger', () => ({
  default: {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn()
  }
}))

Object.defineProperty(global, 'localStorage', { value: localStorageMock })

describe('Storage', () => {
  beforeEach(() => {
    localStorageMock.clear()
    vi.clearAllMocks()
  })

  describe('setItem', () => {
    it('应该成功存储字符串', () => {
      const result = Storage.setItem('key1', 'value1')
      expect(result).toBe(true)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('key1', '"value1"')
    })

    it('应该成功存储对象', () => {
      const obj = { name: 'test', value: 123 }
      const result = Storage.setItem('key2', obj)
      expect(result).toBe(true)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('key2', JSON.stringify(obj))
    })

    it('应该成功存储数组', () => {
      const arr = [1, 2, 3]
      const result = Storage.setItem('key3', arr)
      expect(result).toBe(true)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('key3', JSON.stringify(arr))
    })

    it('应该成功存储数字', () => {
      const result = Storage.setItem('key4', 123)
      expect(result).toBe(true)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('key4', '123')
    })

    it('应该成功存储布尔值', () => {
      const result = Storage.setItem('key5', true)
      expect(result).toBe(true)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('key5', 'true')
    })
  })

  describe('getItem', () => {
    it('应该成功获取字符串', () => {
      localStorageMock.getItem.mockReturnValueOnce('"value1"')
      const result = Storage.getItem('key1')
      expect(result).toBe('value1')
    })

    it('应该成功获取对象', () => {
      const obj = { name: 'test', value: 123 }
      localStorageMock.getItem.mockReturnValueOnce(JSON.stringify(obj))
      const result = Storage.getItem('key2')
      expect(result).toEqual(obj)
    })

    it('应该返回默认值当键不存在时', () => {
      localStorageMock.getItem.mockReturnValueOnce(null)
      const result = Storage.getItem('nonexistent', 'default')
      expect(result).toBe('default')
    })

    it('应该返回 null 作为默认默认值', () => {
      localStorageMock.getItem.mockReturnValueOnce(null)
      const result = Storage.getItem('nonexistent')
      expect(result).toBe(null)
    })

    it('应该处理非 JSON 字符串', () => {
      localStorageMock.getItem.mockReturnValueOnce('plain-string')
      const result = Storage.getItem('key')
      expect(result).toBe('plain-string')
    })
  })

  describe('removeItem', () => {
    it('应该成功删除项', () => {
      const result = Storage.removeItem('key1')
      expect(result).toBe(true)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('key1')
    })
  })

  describe('clear', () => {
    it('应该成功清空存储', () => {
      const result = Storage.clear()
      expect(result).toBe(true)
      expect(localStorageMock.clear).toHaveBeenCalled()
    })
  })

  describe('isAvailable', () => {
    it('应该返回 true 当 localStorage 可用', () => {
      const result = Storage.isAvailable()
      expect(result).toBe(true)
    })
  })
})

