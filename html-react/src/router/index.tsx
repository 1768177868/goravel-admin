import { Suspense, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
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
  const userInfoFetched = useUserStore((s) => s.userInfoFetched)
  const menus = useUserStore((s) => s.menus)
  const adminInfo = useUserStore((s) => s.adminInfo)
  const fetchUserInfo = useUserStore((s) => s.fetchUserInfo)
  const [ready, setReady] = useState(false)
  const bootstrapping = useRef(false)

  useEffect(() => {
    let cancelled = false

    async function bootstrap() {
      if (!token) {
        setReady(true)
        return
      }

      if (bootstrapping.current) return
      bootstrapping.current = true

      try {
        const menusEmpty = !menus || menus.length === 0
        if (!userInfoFetched || menusEmpty) {
          await fetchUserInfo(!userInfoFetched)
        }
      } catch (error) {
        logger.error('Auth bootstrap failed:', error)
      } finally {
        bootstrapping.current = false
        if (!cancelled) setReady(true)
      }
    }

    void bootstrap()
    return () => {
      cancelled = true
    }
  }, [token, userInfoFetched, menus, fetchUserInfo])

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  if (!ready && !adminInfo) {
    return <PageFallback />
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

export function AppRouter() {
  const menus = useUserStore((s) => s.menus)
  const dynamicChildren = useMemo(() => convertMenusToRoutes(menus), [menus])
  const router = useMemo(() => buildRouter(dynamicChildren), [dynamicChildren])

  return <RouterProvider router={router} />
}
