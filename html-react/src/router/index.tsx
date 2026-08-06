import { Suspense, useEffect, useMemo, useState, type ReactNode } from 'react'
import {
  Navigate,
  Outlet,
  RouterProvider,
  createBrowserRouter,
  useLocation,
  useNavigate,
} from 'react-router-dom'
import { Spin } from 'antd'
import { useUserStore } from '@/stores/user'
import { setNavigator } from '@/utils/navigation'
import { convertMenusToRoutes } from './dynamicRoutes'
import { lazyLoad } from './lazyLoad'
import logger from '@/utils/logger'

const LoginPage = lazyLoad(() => import('../pages/Login'))
const MainLayout = lazyLoad(() => import('../layouts/MainLayout'))
const DashboardPage = lazyLoad(() => import('../pages/Dashboard'))
const ProfilePage = lazyLoad(() => import('../pages/profile/Profile'))
const NotFoundPage = lazyLoad(() => import('../pages/NotFound'))

function PageFallback() {
  return (
    <div className="page-fallback">
      <Spin size="large" />
    </div>
  )
}

function withSuspense(element: ReactNode) {
  return <Suspense fallback={<PageFallback />}>{element}</Suspense>
}

/** Registers navigate for axios 401 redirects. */
function NavigatorBridge() {
  const navigate = useNavigate()
  useEffect(() => {
    setNavigator((to, options) => {
      navigate(to, { replace: options?.replace })
    })
  }, [navigate])
  return null
}

function AuthGuard() {
  const location = useLocation()
  const token = useUserStore((s) => s.token)

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return (
    <>
      <NavigatorBridge />
      <Outlet />
    </>
  )
}

function GuestGuard() {
  const token = useUserStore((s) => s.token)
  if (token) return <Navigate to="/dashboard" replace />
  return (
    <>
      <NavigatorBridge />
      <Outlet />
    </>
  )
}

function buildRouter(dynamicChildren: ReturnType<typeof convertMenusToRoutes>) {
  return createBrowserRouter([
    {
      path: '/login',
      element: <GuestGuard />,
      children: [
        {
          index: true,
          element: withSuspense(<LoginPage />),
          handle: { requiresAuth: false },
        },
      ],
    },
    {
      path: '/',
      element: <AuthGuard />,
      children: [
        {
          element: withSuspense(<MainLayout />),
          children: [
            { index: true, element: <Navigate to="/dashboard" replace /> },
            {
              path: 'dashboard',
              element: withSuspense(<DashboardPage />),
              handle: { titleKey: 'menu.dashboard' },
            },
            {
              path: 'profile',
              element: withSuspense(<ProfilePage />),
              handle: { titleKey: 'menu.profile' },
            },
            ...dynamicChildren.map((route) => ({
              ...route,
              element: withSuspense(route.element),
            })),
            {
              path: '*',
              element: withSuspense(<NotFoundPage />),
              handle: { titleKey: 'notFound.title' },
            },
          ],
        },
      ],
    },
  ])
}

/**
 * Bootstrap menus before mounting the data router, so hard-refresh never hits
 * catch-all with empty dynamic routes, and AuthGuard is not remounted mid-fetch.
 */
export function AppRouter() {
  const token = useUserStore((s) => s.token)
  const menus = useUserStore((s) => s.menus)
  const fetchUserInfo = useUserStore((s) => s.fetchUserInfo)
  const [bootstrapped, setBootstrapped] = useState(() => !token)

  useEffect(() => {
    let cancelled = false

    async function bootstrap() {
      if (!token) {
        if (!cancelled) setBootstrapped(true)
        return
      }

      if (!cancelled) setBootstrapped(false)

      try {
        const state = useUserStore.getState()
        const menusEmpty = !state.menus || state.menus.length === 0
        if (!state.userInfoFetched || menusEmpty) {
          await fetchUserInfo(!state.userInfoFetched)
        }
      } catch (error) {
        logger.error('Auth bootstrap failed:', error)
      } finally {
        if (!cancelled) setBootstrapped(true)
      }
    }

    void bootstrap()
    return () => {
      cancelled = true
    }
  }, [token, fetchUserInfo])

  const dynamicChildren = useMemo(() => convertMenusToRoutes(menus), [menus])
  const router = useMemo(() => buildRouter(dynamicChildren), [dynamicChildren])

  if (!bootstrapped) {
    return <PageFallback />
  }

  return <RouterProvider router={router} />
}
