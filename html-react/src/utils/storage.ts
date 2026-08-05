import logger from './logger'

/**
 * Safe localStorage wrapper with JSON serialization and quota recovery.
 * Mirrors html/src/utils/storage.js behavior (shared key names with Vue app).
 */
export class Storage {
  static setItem(key: string, value: unknown): boolean {
    try {
      localStorage.setItem(key, JSON.stringify(value))
      return true
    } catch (error) {
      const err = error as DOMException
      if (err?.name === 'QuotaExceededError') {
        logger.warn('Storage quota exceeded, attempting to clear old data')
        if (this.clearOldData()) {
          try {
            localStorage.setItem(key, JSON.stringify(value))
            return true
          } catch (e) {
            logger.error('Failed to save after clearing:', e)
            return false
          }
        }
        return false
      }
      logger.error('Storage setItem error:', error)
      return false
    }
  }

  static getItem<T>(key: string, defaultValue: T | null = null): T | null {
    try {
      const item = localStorage.getItem(key)
      if (item === null) return defaultValue

      try {
        return JSON.parse(item) as T
      } catch {
        if (item.startsWith('"') && item.endsWith('"')) {
          logger.warn(`Storage getItem: Invalid JSON for key "${key}", using default value`)
          return defaultValue
        }
        return item as unknown as T
      }
    } catch (error) {
      logger.error('Storage getItem error:', error)
      return defaultValue
    }
  }

  static removeItem(key: string): boolean {
    try {
      localStorage.removeItem(key)
      return true
    } catch (error) {
      logger.error('Storage removeItem error:', error)
      return false
    }
  }

  static clear(): boolean {
    try {
      localStorage.clear()
      return true
    } catch (error) {
      logger.error('Storage clear error:', error)
      return false
    }
  }

  static clearOldData(keysToKeep: string[] = ['token', 'adminInfo']): boolean {
    try {
      const keysToRemove: string[] = []
      Object.keys(localStorage).forEach((key) => {
        if (!keysToKeep.includes(key) && !key.startsWith('vxe-')) {
          keysToRemove.push(key)
        }
      })
      keysToRemove.forEach((key) => {
        try {
          localStorage.removeItem(key)
        } catch (e) {
          logger.warn(`Failed to remove key: ${key}`, e)
        }
      })
      return true
    } catch (error) {
      logger.error('Failed to clear old data:', error)
      return false
    }
  }

  static isAvailable(): boolean {
    try {
      const test = '__storage_test__'
      localStorage.setItem(test, test)
      localStorage.removeItem(test)
      return true
    } catch {
      return false
    }
  }
}

export default Storage
