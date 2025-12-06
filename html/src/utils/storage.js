import logger from './logger'

/**
 * localStorage 工具类
 * 提供安全的存储操作，包含错误处理和配额管理
 */
export class Storage {
  /**
   * 设置存储项
   * @param {string} key - 键名
   * @param {any} value - 值
   * @returns {boolean} 是否成功
   */
  static setItem(key, value) {
    try {
      const serialized = JSON.stringify(value)
      localStorage.setItem(key, serialized)
      return true
    } catch (error) {
      if (error.name === 'QuotaExceededError') {
        // 存储空间不足，尝试清理
        logger.warn('Storage quota exceeded, attempting to clear old data')
        const cleared = this.clearOldData()
        
        if (cleared) {
          try {
            localStorage.setItem(key, JSON.stringify(value))
            logger.info('Storage item saved after clearing old data')
            return true
          } catch (e) {
            logger.error('Failed to save after clearing:', e)
            return false
          }
        }
        
        logger.error('Failed to clear enough space for storage')
        return false
      }
      
      logger.error('Storage setItem error:', error)
      return false
    }
  }

  /**
   * 获取存储项
   * @param {string} key - 键名
   * @param {any} defaultValue - 默认值
   * @returns {any} 存储的值或默认值
   */
  static getItem(key, defaultValue = null) {
    try {
      const item = localStorage.getItem(key)
      if (item === null) {
        return defaultValue
      }
      
      // 尝试解析 JSON
      try {
        return JSON.parse(item)
      } catch (parseError) {
        // 如果解析失败，可能是旧代码直接存储的字符串
        // 检查是否是简单的字符串值
        if (item.startsWith('"') && item.endsWith('"')) {
          // 是 JSON 字符串格式，但解析失败，返回默认值
          logger.warn(`Storage getItem: Invalid JSON for key "${key}", using default value`)
          return defaultValue
        }
        // 直接存储的字符串，直接返回
        return item
      }
    } catch (error) {
      logger.error('Storage getItem error:', error)
      return defaultValue
    }
  }

  /**
   * 删除存储项
   * @param {string} key - 键名
   * @returns {boolean} 是否成功
   */
  static removeItem(key) {
    try {
      localStorage.removeItem(key)
      return true
    } catch (error) {
      logger.error('Storage removeItem error:', error)
      return false
    }
  }

  /**
   * 清空所有存储
   * @returns {boolean} 是否成功
   */
  static clear() {
    try {
      localStorage.clear()
      return true
    } catch (error) {
      logger.error('Storage clear error:', error)
      return false
    }
  }

  /**
   * 清理旧数据以释放空间
   * @param {string[]} keysToKeep - 需要保留的键名列表
   * @returns {boolean} 是否成功清理
   */
  static clearOldData(keysToKeep = ['token', 'adminInfo']) {
    try {
      const keysToRemove = []
      
      // 找出需要删除的键
      Object.keys(localStorage).forEach(key => {
        if (!keysToKeep.includes(key) && !key.startsWith('vxe-')) {
          keysToRemove.push(key)
        }
      })

      // 删除旧数据
      keysToRemove.forEach(key => {
        try {
          localStorage.removeItem(key)
        } catch (e) {
          logger.warn(`Failed to remove key: ${key}`, e)
        }
      })

      logger.info(`Cleared ${keysToRemove.length} old storage items`)
      return true
    } catch (error) {
      logger.error('Failed to clear old data:', error)
      return false
    }
  }

  /**
   * 检查存储是否可用
   * @returns {boolean} 是否可用
   */
  static isAvailable() {
    try {
      const test = '__storage_test__'
      localStorage.setItem(test, test)
      localStorage.removeItem(test)
      return true
    } catch {
      return false
    }
  }

  /**
   * 获取存储使用情况（估算）
   * @returns {Object} 存储使用情况
   */
  static getUsage() {
    try {
      let total = 0
      for (let key in localStorage) {
        if (localStorage.hasOwnProperty(key)) {
          total += localStorage[key].length + key.length
        }
      }
      return {
        used: total,
        available: 5 * 1024 * 1024 - total, // 假设 5MB 限制
        percentage: (total / (5 * 1024 * 1024)) * 100
      }
    } catch (error) {
      logger.error('Failed to get storage usage:', error)
      return null
    }
  }
}

export default Storage

