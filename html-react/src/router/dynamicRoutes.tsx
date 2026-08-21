import type { ComponentType, LazyExoticComponent, ReactElement } from 'react'
import type { RouteObject } from 'react-router-dom'
import { lazyLoad } from './lazyLoad'
import { flattenTree } from '@/utils/tree'
import logger from '@/utils/logger'
import { menuTitleKeyFromSlug, normalizeMenuSlug } from '@/utils/menuTitle'
import type { MenuNode } from '@/types'

/**
 * Vite glob of page modules. Menu `component` values map like Vue:
 * - `admin/AdminList` -> `../pages/admin/AdminList.tsx`
 */
const pageModules = import.meta.glob('../pages/**/*.{tsx,jsx}')

/** Reuse the same React.lazy instance across router rebuilds (menus refresh / HMR). */
const lazyElementCache = new Map<string, ReactElement>()

export interface AppRouteMeta {
  titleKey?: string
  menuId?: string | number
  menuSlug?: string
  noCache?: boolean
  requiresAuth?: boolean
  externalUrl?: string
}

export type AppRouteObject = RouteObject & {
  handle?: AppRouteMeta
  children?: AppRouteObject[]
}

function resolvePageModule(
  component: string,
): { key: string; importFn: () => Promise<{ default: ComponentType<object> }> } | null {
  if (!component || component === 'Layout') return null

  const modulePath = component.replace(/^(\.\.\/)?(pages|views)\//, '').replace(/\.(tsx|jsx|vue)$/, '')

  const possiblePaths = [
    `../pages/${modulePath}.tsx`,
    `../pages/${modulePath}.jsx`,
    `../pages/${modulePath}/index.tsx`,
    `../pages/${modulePath}/index.jsx`,
  ]

  for (const path of possiblePaths) {
    const mod = pageModules[path]
    if (mod) {
      return {
        key: path,
        importFn: mod as () => Promise<{ default: ComponentType<object> }>,
      }
    }
  }

  logger.warn(`Component not found: ${component} (tried: ${possiblePaths.join(', ')})`)
  logger.debug('Available modules:', Object.keys(pageModules))
  return null
}

function createLazyElement(
  cacheKey: string,
  importFn: () => Promise<{ default: ComponentType<object> }>,
) {
  const cached = lazyElementCache.get(cacheKey)
  if (cached) return cached

  const LazyComp: LazyExoticComponent<ComponentType<object>> = lazyLoad(importFn)
  const element = <LazyComp />
  lazyElementCache.set(cacheKey, element)
  return element
}

/**
 * Convert backend menus into React Router child routes under MainLayout.
 */
export function convertMenusToRoutes(menus: MenuNode[] | null | undefined): AppRouteObject[] {
  if (!menus || !Array.isArray(menus)) return []

  const withChildren = menus.map(normalizeMenuChildren)
  const flatMenus = flattenTree(withChildren, 'children') as MenuNode[]
  const routes: AppRouteObject[] = []
  const processedPaths = new Set<string>()

  flatMenus.forEach((menu) => {
    const type = menu.Type ?? menu.type ?? 1
    const status = menu.Status ?? menu.status ?? 1
    const linkType = menu.LinkType ?? menu.link_type ?? 1
    const component = menu.Component || menu.component || ''

    if (status !== 1 || type === 3) return
    if (type === 1 && !component) return

    const path = menu.Path || menu.path || ''
    if (!path || path === '/') return
    if (processedPaths.has(path)) return
    processedPaths.add(path)

    const routePath = path.startsWith('/') ? path.slice(1) : path
    const routeName = routePath
      .split('/')
      .filter(Boolean)
      .map((part) =>
        part
          .split('-')
          .map((p) => p.charAt(0).toUpperCase() + p.slice(1))
          .join(''),
      )
      .join('')

    const slug = menu.Slug || menu.slug || routePath
    const cleanSlug = normalizeMenuSlug(slug.startsWith('/') ? slug.slice(1) : slug)
    const noCache = menu.no_cache === 1 || menu.NoCache === 1

    const handle: AppRouteMeta = {
      titleKey: menuTitleKeyFromSlug(cleanSlug),
      menuId: menu.id || menu.ID,
      menuSlug: cleanSlug,
      noCache: !!noCache,
    }

    if (linkType === 2) {
      routes.push({
        path: routePath,
        id: routeName,
        handle: { ...handle, externalUrl: path },
        element: createLazyElement('iframe/IframeView', () => import('../pages/iframe/IframeView')),
      })
      return
    }

    const pageImport = resolvePageModule(component)
    if (!pageImport) {
      if (component === 'Layout') return
      logger.warn(`Using Placeholder for missing component: ${component} (${path})`)
      routes.push({
        path: routePath,
        id: routeName,
        handle,
        element: createLazyElement('Placeholder', () => import('../pages/Placeholder')),
      })
      return
    }

    routes.push({
      path: routePath,
      id: routeName,
      handle,
      element: createLazyElement(pageImport.key, pageImport.importFn),
    })
  })

  return routes
}

function normalizeMenuChildren(menu: MenuNode): MenuNode {
  const children = menu.children || menu.Children || []
  return {
    ...menu,
    children: Array.isArray(children) ? children.map(normalizeMenuChildren) : [],
  }
}
