import { computed } from 'vue'
import { useUserStore } from '../store/user'

/**
 * 权限相关的 composable
 * 提供权限检查和按钮显示控制的功能
 */
export function usePermission() {
  const userStore = useUserStore()

  /**
   * 检查是否有权限
   * @param {string} permission - 权限标识
   * @returns {boolean}
   */
  const hasPermission = (permission) => {
    return userStore.hasPermission(permission)
  }

  /**
   * 检查是否应该显示按钮（考虑权限和配置）
   * 如果用户有权限，总是返回 true
   * 如果用户没有权限，根据配置 ADMIN_SHOW_BUTTONS_WITHOUT_PERMISSION 决定是否显示
   * 
   * @param {string} permission - 权限标识
   * @returns {boolean}
   * 
   * @example
   * // 在模板中使用
   * <el-button v-if="shouldShowButton('admin.store')" @click="handleAdd">添加</el-button>
   * 
   * // 在脚本中使用
   * const { shouldShowButton } = usePermission()
   * if (shouldShowButton('admin.update')) {
   *   // 显示编辑按钮
   * }
   */
  const shouldShowButton = (permission) => {
    return userStore.shouldShowButton(permission)
  }

  /**
   * 检查按钮是否应该被禁用（没有权限时禁用）
   * @param {string} permission - 权限标识
   * @returns {boolean}
   */
  const isButtonDisabled = (permission) => {
    return !userStore.hasPermission(permission)
  }

  /**
   * 获取按钮的显示和禁用状态（推荐使用，一个方法搞定）
   * @param {string} permission - 权限标识
   * @returns {{ show: boolean, disabled: boolean }}
   * 
   * @example
   * // 方式一：同时控制显示和禁用（根据配置决定是否隐藏）
   * <el-button 
   *   v-if="getButtonState('admin.store').show" 
   *   :disabled="getButtonState('admin.store').disabled"
   *   @click="handleAdd"
   * >
   *   添加
   * </el-button>
   * 
   * // 方式二：只控制禁用（按钮始终显示，无权限时禁用）
   * <el-button 
   *   :disabled="getButtonState('admin.store').disabled"
   *   @click="handleAdd"
   * >
   *   添加
   * </el-button>
   */
  const getButtonState = (permission) => {
    const hasPerm = userStore.hasPermission(permission)
    const show = userStore.shouldShowButton(permission)
    const disabled = !hasPerm
    
    return { show, disabled }
  }

  return {
    hasPermission,
    shouldShowButton,
    isButtonDisabled,
    getButtonState
  }
}

