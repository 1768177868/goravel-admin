/**
 * Exclude a node and all its descendants from a tree (for parent TreeSelect).
 */
export function excludeNodeAndChildren<T extends object>(
  tree: T[] | null | undefined,
  excludeId: string | number,
  idKey: string = 'id',
  childrenKey: string = 'children',
): T[] {
  if (!Array.isArray(tree)) return []

  return tree
    .filter((node) => String((node as Record<string, unknown>)[idKey]) !== String(excludeId))
    .map((node) => {
      const children = (node as Record<string, unknown>)[childrenKey] as T[] | undefined
      if (!Array.isArray(children) || children.length === 0) return { ...node }
      return {
        ...node,
        [childrenKey]: excludeNodeAndChildren(children, excludeId, idKey, childrenKey),
      }
    }) as T[]
}

export function flattenTree<T extends object>(
  tree: T[] | null | undefined,
  childrenKey: string = 'children',
  transform: ((node: T) => unknown) | null = null,
): unknown[] {
  if (!Array.isArray(tree)) return []

  const result: unknown[] = []

  tree.forEach((node) => {
    const processed = transform ? transform(node) : node
    if (processed !== null && processed !== undefined) {
      result.push(processed)
    }
    const children = (node as Record<string, unknown>)[childrenKey] as T[] | undefined
    if (Array.isArray(children) && children.length > 0) {
      result.push(...flattenTree(children, childrenKey, transform))
    }
  })

  return result
}

export function filterTree<T extends object>(
  tree: T[] | null | undefined,
  predicate: (node: T) => boolean,
  childrenKey: string = 'children',
): T[] {
  if (!Array.isArray(tree)) return []

  return tree.filter(predicate).map((node) => {
    const children = (node as Record<string, unknown>)[childrenKey] as T[] | undefined
    if (Array.isArray(children) && children.length > 0) {
      return {
        ...node,
        [childrenKey]: filterTree(children, predicate, childrenKey),
      }
    }
    return { ...node, [childrenKey]: [] }
  }) as T[]
}

export function mapTree<T extends object, R extends object>(
  tree: T[] | null | undefined,
  mapper: (node: T) => R,
  childrenKey: string = 'children',
): R[] {
  if (!Array.isArray(tree)) return []

  return tree.map((node) => {
    const mappedNode = mapper(node)
    const originalChildren = (node as Record<string, unknown>)[childrenKey] as T[] | undefined
    const mappedChildren = (mappedNode as Record<string, unknown>)[childrenKey] as R[] | undefined

    if (mappedChildren !== undefined) {
      if (Array.isArray(mappedChildren) && mappedChildren.length > 0) {
        ;(mappedNode as Record<string, unknown>)[childrenKey] = mapTree(
          mappedChildren as unknown as T[],
          mapper,
          childrenKey,
        )
      } else {
        ;(mappedNode as Record<string, unknown>)[childrenKey] = mappedChildren
      }
    } else if (Array.isArray(originalChildren) && originalChildren.length > 0) {
      ;(mappedNode as Record<string, unknown>)[childrenKey] = mapTree(originalChildren, mapper, childrenKey)
    } else {
      ;(mappedNode as Record<string, unknown>)[childrenKey] = undefined
    }

    return mappedNode
  })
}

export function findInTree<T extends object>(
  tree: T[] | null | undefined,
  predicate: (node: T) => boolean,
  childrenKey: string = 'children',
): T | null {
  if (!Array.isArray(tree)) return null

  for (const node of tree) {
    if (predicate(node)) return node
    const children = (node as Record<string, unknown>)[childrenKey] as T[] | undefined
    if (Array.isArray(children) && children.length > 0) {
      const found = findInTree(children, predicate, childrenKey)
      if (found) return found
    }
  }
  return null
}

export function sortTree<T extends object>(
  tree: T[] | null | undefined,
  compareFn: (a: T, b: T) => number,
  childrenKey: string = 'children',
): T[] {
  if (!Array.isArray(tree)) return []

  return tree
    .map((node) => {
      const children = (node as Record<string, unknown>)[childrenKey] as T[] | undefined
      if (Array.isArray(children) && children.length > 0) {
        return {
          ...node,
          [childrenKey]: sortTree(children, compareFn, childrenKey),
        }
      }
      return { ...node }
    })
    .sort(compareFn) as T[]
}
