type NavigateFn = (to: string, options?: { replace?: boolean }) => void

let navigateRef: NavigateFn | null = null

/** Register React Router navigate so axios interceptors can redirect without circular imports. */
export function setNavigator(navigate: NavigateFn) {
  navigateRef = navigate
}

export function navigateTo(to: string, options?: { replace?: boolean }) {
  if (navigateRef) {
    navigateRef(to, options)
    return
  }
  if (options?.replace) {
    window.location.replace(to)
  } else {
    window.location.href = to
  }
}

export function getCurrentPath(): string {
  return window.location.pathname
}
