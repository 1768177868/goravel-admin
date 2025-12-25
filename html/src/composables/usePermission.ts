import { useUserStore } from '../store/user'

/**
 * 按钮状态
 */
export interface ButtonState {
  /** 是否显示按钮 */
  show: boolean
  /** 是否禁用按钮 */
  disabled: boolean
}

/**
 * 权限 composable 返回值
 */
export interface UsePermissionReturn {
  /** 检查是否有权限 */
  hasPermission: (permission: string) => boolean
  /** 检查是否应该显示按钮 */
  shouldShowButton: (permission: string) => boolean
  /** 检查按钮是否应该被禁用 */
  isButtonDisabled: (permission: string) => boolean
  /** 获取按钮的显示和禁用状态 */
  getButtonState: (permission: string) => ButtonState
}

/**
 * 权限相关的 composable
 * 提供权限检查和按钮显示控制的功能
 *
 * @example
 * const { hasPermission, getButtonState } = usePermission()
 *
 * // 检查权限
 * if (hasPermission('admin.store')) {
 *   // 有添加权限
 * }
 *
 * // 获取按钮状态
 * const { show, disabled } = getButtonState('admin.update')
 */
export function usePermission(): UsePermissionReturn {
  const userStore = useUserStore()

  /**
   * 检查是否有权限
   * @param permission - 权限标识
   * @returns 是否有权限
   */
  const hasPermission = (permission: string): boolean => {
    return userStore.hasPermission(permission)
  }

  /**
   * 检查是否应该显示按钮（考虑权限和配置）
   * 如果用户有权限，总是返回 true
   * 如果用户没有权限，根据配置 ADMIN_SHOW_BUTTONS_WITHOUT_PERMISSION 决定是否显示
   *
   * @param permission - 权限标识
   * @returns 是否显示按钮
   *
   * @example
   * <el-button v-if="shouldShowButton('admin.store')" @click="handleAdd">添加</el-button>
   */
  const shouldShowButton = (permission: string): boolean => {
    return userStore.shouldShowButton(permission)
  }

  /**
   * 检查按钮是否应该被禁用（没有权限时禁用）
   * @param permission - 权限标识
   * @returns 是否禁用按钮
   */
  const isButtonDisabled = (permission: string): boolean => {
    return !userStore.hasPermission(permission)
  }

  /**
   * 获取按钮的显示和禁用状态（推荐使用，一个方法搞定）
   * @param permission - 权限标识
   * @returns 按钮状态对象
   *
   * @example
   * // 同时控制显示和禁用
   * <el-button
   *   v-if="getButtonState('admin.store').show"
   *   :disabled="getButtonState('admin.store').disabled"
   *   @click="handleAdd"
   * >
   *   添加
   * </el-button>
   */
  const getButtonState = (permission: string): ButtonState => {
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

