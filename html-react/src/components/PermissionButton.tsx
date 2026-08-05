import { Button, type ButtonProps } from 'antd'
import { usePermission } from '@/hooks/usePermission'

interface PermissionButtonProps extends ButtonProps {
  permission: string
  /** When true, hide completely if no permission; otherwise disable. Default: use getButtonState. */
  hideWhenDenied?: boolean
}

export default function PermissionButton({
  permission,
  hideWhenDenied,
  children,
  disabled,
  ...rest
}: PermissionButtonProps) {
  const { getButtonState } = usePermission()
  const state = getButtonState(permission)

  if (!state.show || (hideWhenDenied && state.disabled)) {
    return null
  }

  return (
    <Button {...rest} disabled={disabled || state.disabled}>
      {children}
    </Button>
  )
}
