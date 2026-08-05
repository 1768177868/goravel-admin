import { useCallback, useMemo } from 'react'
import { useUserStore } from '@/stores/user'

export interface ButtonState {
  show: boolean
  disabled: boolean
}

export function usePermission() {
  const hasPermission = useUserStore((s) => s.hasPermission)
  const shouldShowButton = useUserStore((s) => s.shouldShowButton)

  const isButtonDisabled = useCallback(
    (permission: string) => !hasPermission(permission),
    [hasPermission],
  )

  const getButtonState = useCallback(
    (permission: string): ButtonState => {
      const hasPerm = hasPermission(permission)
      return {
        show: shouldShowButton(permission),
        disabled: !hasPerm,
      }
    },
    [hasPermission, shouldShowButton],
  )

  return useMemo(
    () => ({
      hasPermission,
      shouldShowButton,
      isButtonDisabled,
      getButtonState,
    }),
    [hasPermission, shouldShowButton, isButtonDisabled, getButtonState],
  )
}
