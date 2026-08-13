import type { ReactNode } from 'react'
import { groupBy } from 'lodash-es'
import { mapTree } from '@/utils/tree'

export interface TreeNodeData {
  key: string
  title: ReactNode
  children?: TreeNodeData[]
  isMenu?: boolean
  isPermission?: boolean
  rawId?: number
  method?: string
  path?: string
  slug?: string
}

type MenuLike = Record<string, unknown>
type PermLike = Record<string, unknown>

export interface RolePermissionTranslate {
  menuTitle: (menu: MenuLike) => string
  permissionTitle: (perm: PermLike) => string
  otherLabel: string
}

function fallbackMenuTitle(menu: MenuLike): string {
  return String(menu.Title || menu.title || menu.Name || menu.name || menu.Slug || menu.slug || '')
}

function fallbackPermissionTitle(perm: PermLike): string {
  return String(
    perm.Description || perm.description || perm.Name || perm.name || perm.Slug || perm.slug || '',
  )
}

function transformMenuToTree(menus: MenuLike[], translate: RolePermissionTranslate): TreeNodeData[] {
  if (!Array.isArray(menus)) return []

  return menus.map((node) => {
    const id = Number(node.id ?? node.ID)
    const children = (node.children || node.Children || []) as MenuLike[]
    const slug = String(node.Slug || node.slug || '')
    return {
      key: `menu_${id}`,
      title: translate.menuTitle(node),
      rawId: id,
      isMenu: true,
      isPermission: false,
      slug,
      path: String(node.Path || node.path || ''),
      children: children.length ? transformMenuToTree(children, translate) : undefined,
    }
  })
}

function attachPermissionsToMenus(
  menuTree: TreeNodeData[],
  permissions: PermLike[],
  translate: RolePermissionTranslate,
): TreeNodeData[] {
  const permissionMap = groupBy(permissions, (perm) => Number(perm.MenuID || perm.menu_id || 0))

  const walk = (nodes: TreeNodeData[]): TreeNodeData[] =>
    nodes.map((node) => {
      const result: TreeNodeData = {
        ...node,
        children: node.children ? walk(node.children) : undefined,
      }

      if (result.isMenu && result.rawId != null) {
        const matched = permissionMap[result.rawId] || []
        if (matched.length > 0) {
          const permChildren: TreeNodeData[] = matched.map((perm) => {
            const id = Number(perm.id || perm.ID)
            const method = String(perm.Method || perm.method || '')
            const slug = String(perm.Slug || perm.slug || '')
            const title = translate.permissionTitle(perm)
            return {
              key: `perm_${id}`,
              title: method ? `[${method}] ${title}` : title,
              rawId: id,
              isMenu: false,
              isPermission: true,
              method,
              slug,
              path: String(perm.Path || perm.path || ''),
            }
          })
          permChildren.sort((a, b) => {
            const order: Record<string, number> = { GET: 1, POST: 2, PUT: 3, PATCH: 4, DELETE: 5 }
            return (order[a.method || ''] || 99) - (order[b.method || ''] || 99)
          })
          const menuChildren = (result.children || []).filter((c) => c.isMenu)
          result.children = [...menuChildren, ...permChildren]
        }
      }
      return result
    })

  return walk(menuTree)
}

export function buildMenuPermissionTree(
  menus: MenuLike[],
  permissions: PermLike[],
  translate: RolePermissionTranslate | string = '其他权限',
): TreeNodeData[] {
  const resolved: RolePermissionTranslate =
    typeof translate === 'string'
      ? {
          menuTitle: fallbackMenuTitle,
          permissionTitle: fallbackPermissionTitle,
          otherLabel: translate,
        }
      : translate

  const tree = attachPermissionsToMenus(transformMenuToTree(menus, resolved), permissions, resolved)

  const matched = new Set<number>()
  const collect = (nodes: TreeNodeData[]) => {
    nodes.forEach((n) => {
      if (n.isPermission && n.rawId != null) matched.add(n.rawId)
      if (n.children) collect(n.children)
    })
  }
  collect(tree)

  const unmatched = permissions.filter((p) => !matched.has(Number(p.id || p.ID)))
  if (unmatched.length > 0) {
    tree.push({
      key: 'other_permissions',
      title: resolved.otherLabel,
      isMenu: true,
      children: unmatched.map((perm) => {
        const id = Number(perm.id || perm.ID)
        const method = String(perm.Method || perm.method || '')
        const slug = String(perm.Slug || perm.slug || '')
        const title = resolved.permissionTitle(perm)
        return {
          key: `perm_${id}`,
          title: method ? `[${method}] ${title}` : title,
          rawId: id,
          isMenu: false,
          isPermission: true,
          method,
          slug,
          path: String(perm.Path || perm.path || ''),
        }
      }),
    })
  }

  return tree
}

export function buildCheckedKeysFromIds(
  tree: TreeNodeData[],
  menuIds: number[],
  permissionIds: number[],
): string[] {
  const permissionIdSet = new Set(permissionIds)
  const menuNodeMap = new Map<number, TreeNodeData>()
  const permissionNodeMap = new Map<number, TreeNodeData>()

  const collect = (nodes: TreeNodeData[]) => {
    nodes.forEach((node) => {
      if (node.rawId != null) {
        if (node.isMenu) menuNodeMap.set(node.rawId, node)
        if (node.isPermission) permissionNodeMap.set(node.rawId, node)
      }
      if (node.children) collect(node.children)
    })
  }
  collect(tree)

  const keys = new Set<string>()
  permissionIds.forEach((id) => {
    const node = permissionNodeMap.get(id)
    if (node?.key) keys.add(node.key)
  })

  menuIds.forEach((menuId) => {
    const menuNode = menuNodeMap.get(menuId)
    if (!menuNode) return
    const directPerms = (menuNode.children || []).filter((c) => c.isPermission && c.rawId != null)
    if (directPerms.length === 0) {
      keys.add(menuNode.key)
      return
    }
    const selectedCount = directPerms.filter((p) => permissionIdSet.has(p.rawId!)).length
    if (selectedCount === directPerms.length) keys.add(menuNode.key)
  })

  return Array.from(keys)
}

export function collectIdsFromCheckedKeys(
  tree: TreeNodeData[],
  checkedKeys: Array<string | number>,
): { menu_ids: number[]; permission_ids: number[] } {
  const checked = new Set(checkedKeys.map(String))
  const menuIds: number[] = []
  const permissionIds: number[] = []

  const walk = (nodes: TreeNodeData[]) => {
    nodes.forEach((node) => {
      const key = String(node.key)
      const isChecked = checked.has(key)

      if (node.isPermission && isChecked && node.rawId != null) {
        permissionIds.push(node.rawId)
      }

      if (node.isMenu && key !== 'other_permissions' && node.rawId != null) {
        const hasDirectPermission = (node.children || []).some(
          (child) => child.isPermission && checked.has(String(child.key)),
        )
        if (isChecked || hasDirectPermission) menuIds.push(node.rawId)
      }

      if (node.children) walk(node.children)
    })
  }

  walk(tree)
  return {
    menu_ids: [...new Set(menuIds)],
    permission_ids: [...new Set(permissionIds)],
  }
}

export function flattenPermissionList(data: unknown): PermLike[] {
  if (Array.isArray(data)) return data as PermLike[]
  if (data && typeof data === 'object') {
    const obj = data as { list?: unknown[]; data?: unknown[] }
    if (Array.isArray(obj.list)) return obj.list as PermLike[]
    if (Array.isArray(obj.data)) return obj.data as PermLike[]
  }
  return []
}

export function flattenMenuList(data: unknown): MenuLike[] {
  if (Array.isArray(data)) return data as MenuLike[]
  if (data && typeof data === 'object') {
    const obj = data as { menus?: unknown[]; list?: unknown[]; data?: unknown[] }
    if (Array.isArray(obj.menus)) return obj.menus as MenuLike[]
    if (Array.isArray(obj.list)) return obj.list as MenuLike[]
    if (Array.isArray(obj.data)) return obj.data as MenuLike[]
  }
  return []
}

// keep mapTree available for future reuse without unused import noise in build
void mapTree
