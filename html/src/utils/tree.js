/**
 * 树形数据处理工具函数
 */

/**
 * 扁平化树形结构
 * @param {Array} tree - 树形数组
 * @param {string} childrenKey - 子节点字段名，默认为 'children'
 * @param {Function} transform - 可选的转换函数，用于转换每个节点
 * @returns {Array} 扁平化后的数组
 * 
 * @example
 * const tree = [
 *   { id: 1, name: 'A', children: [{ id: 2, name: 'B' }] }
 * ]
 * flattenTree(tree) // [{ id: 1, name: 'A' }, { id: 2, name: 'B' }]
 * 
 * @example
 * // 使用转换函数
 * flattenTree(tree, 'children', node => ({ id: node.id, label: node.name }))
 */
export function flattenTree(tree, childrenKey = 'children', transform = null) {
  if (!Array.isArray(tree)) {
    return []
  }

  const result = []
  
  tree.forEach(node => {
    // 如果有转换函数，先转换节点
    const processedNode = transform ? transform(node) : node
    if (processedNode !== null && processedNode !== undefined) {
      result.push(processedNode)
    }
    
    // 递归处理子节点
    const children = node[childrenKey]
    if (Array.isArray(children) && children.length > 0) {
      result.push(...flattenTree(children, childrenKey, transform))
    }
  })
  
  return result
}

/**
 * 过滤树形结构
 * @param {Array} tree - 树形数组
 * @param {Function} predicate - 过滤条件函数，返回 true 的节点会被保留
 * @param {string} childrenKey - 子节点字段名，默认为 'children'
 * @returns {Array} 过滤后的树形结构
 * 
 * @example
 * const tree = [
 *   { id: 1, status: 1, children: [{ id: 2, status: 0 }] }
 * ]
 * filterTree(tree, node => node.status === 1)
 * // [{ id: 1, status: 1, children: [] }]
 */
export function filterTree(tree, predicate, childrenKey = 'children') {
  if (!Array.isArray(tree)) {
    return []
  }

  return tree
    .filter(predicate)
    .map(node => {
      const children = node[childrenKey]
      if (Array.isArray(children) && children.length > 0) {
        return {
          ...node,
          [childrenKey]: filterTree(children, predicate, childrenKey)
        }
      }
      return { ...node, [childrenKey]: [] }
    })
}

/**
 * 映射树形结构（转换每个节点）
 * @param {Array} tree - 树形数组
 * @param {Function} mapper - 映射函数，用于转换每个节点
 * @param {string} childrenKey - 子节点字段名，默认为 'children'
 * @returns {Array} 转换后的树形结构
 * 
 * @example
 * const tree = [{ id: 1, name: 'A', children: [{ id: 2, name: 'B' }] }]
 * mapTree(tree, node => ({ value: node.id, label: node.name }))
 * // [{ value: 1, label: 'A', children: [{ value: 2, label: 'B' }] }]
 */
export function mapTree(tree, mapper, childrenKey = 'children') {
  if (!Array.isArray(tree)) {
    return []
  }

  return tree.map(node => {
    const mappedNode = mapper(node)
    const originalChildren = node[childrenKey]
    const mappedChildren = mappedNode[childrenKey]
    
    // 如果 mapper 已经设置了 children，需要判断：
    // 1. 如果 mappedChildren 是数组，说明 mapper 可能修改了 children（添加了权限等）
    //    此时需要递归处理 mappedChildren，但要保留 mapper 的修改
    // 2. 如果 mappedChildren 是 undefined，说明 mapper 没有设置 children
    //    此时需要递归处理原节点的 children
    
    if (mappedChildren !== undefined) {
      // mapper 已经设置了 children，递归处理这些 children（可能是子菜单+权限的混合）
      if (Array.isArray(mappedChildren) && mappedChildren.length > 0) {
        mappedNode[childrenKey] = mapTree(mappedChildren, mapper, childrenKey)
      } else {
        // mappedChildren 是空数组或其他值，保持原样
        mappedNode[childrenKey] = mappedChildren
      }
    } else if (Array.isArray(originalChildren) && originalChildren.length > 0) {
      // mapper 没有设置 children，递归处理原节点的 children
      mappedNode[childrenKey] = mapTree(originalChildren, mapper, childrenKey)
    } else {
      // 没有 children，设置为 undefined
      mappedNode[childrenKey] = undefined
    }
    
    return mappedNode
  })
}

/**
 * 查找树形结构中的节点
 * @param {Array} tree - 树形数组
 * @param {Function} predicate - 查找条件函数
 * @param {string} childrenKey - 子节点字段名，默认为 'children'
 * @returns {Object|null} 找到的节点，未找到返回 null
 * 
 * @example
 * findInTree(tree, node => node.id === 2)
 */
export function findInTree(tree, predicate, childrenKey = 'children') {
  if (!Array.isArray(tree)) {
    return null
  }

  for (const node of tree) {
    if (predicate(node)) {
      return node
    }
    
    const children = node[childrenKey]
    if (Array.isArray(children) && children.length > 0) {
      const found = findInTree(children, predicate, childrenKey)
      if (found) {
        return found
      }
    }
  }
  
  return null
}

/**
 * 对树形结构进行排序
 * @param {Array} tree - 树形数组
 * @param {Function} compareFn - 比较函数，与 Array.sort 的 compareFn 相同
 * @param {string} childrenKey - 子节点字段名，默认为 'children'
 * @returns {Array} 排序后的树形结构
 * 
 * @example
 * sortTree(tree, (a, b) => a.sort - b.sort)
 */
export function sortTree(tree, compareFn, childrenKey = 'children') {
  if (!Array.isArray(tree)) {
    return []
  }

  return tree
    .map(node => {
      const children = node[childrenKey]
      if (Array.isArray(children) && children.length > 0) {
        return {
          ...node,
          [childrenKey]: sortTree(children, compareFn, childrenKey)
        }
      }
      return { ...node }
    })
    .sort(compareFn)
}

/**
 * 过滤并排序树形结构（常用组合操作）
 * @param {Array} tree - 树形数组
 * @param {Function} predicate - 过滤条件函数
 * @param {Function} compareFn - 排序比较函数
 * @param {string} childrenKey - 子节点字段名，默认为 'children'
 * @returns {Array} 过滤并排序后的树形结构
 * 
 * @example
 * filterAndSortTree(tree, node => node.status === 1, (a, b) => a.sort - b.sort)
 */
export function filterAndSortTree(tree, predicate, compareFn, childrenKey = 'children') {
  return sortTree(filterTree(tree, predicate, childrenKey), compareFn, childrenKey)
}

/**
 * 排除指定节点及其所有子节点
 * @param {Array} tree - 树形数组
 * @param {number|string|Function} targetIdOrPredicate - 要排除的节点ID，或查找节点的条件函数
 * @param {string} idKey - 节点ID字段名，默认为 'id'
 * @param {string} childrenKey - 子节点字段名，默认为 'children'
 * @returns {Array} 排除指定节点及其子节点后的树形结构
 * 
 * @example
 * // 排除ID为5的节点及其所有子节点
 * excludeNodeAndChildren(tree, 5)
 * 
 * @example
 * // 使用条件函数查找节点
 * excludeNodeAndChildren(tree, node => node.slug === 'admin')
 */
export function excludeNodeAndChildren(tree, targetIdOrPredicate, idKey = 'id', childrenKey = 'children') {
  if (!Array.isArray(tree)) {
    return []
  }

  // 收集所有需要排除的节点ID（目标节点及其所有子节点）
  const excludeIds = new Set()
  
  // 查找目标节点并收集其ID
  let targetId = null
  if (typeof targetIdOrPredicate === 'function') {
    // 使用条件函数查找节点
    const targetNode = findInTree(tree, targetIdOrPredicate, childrenKey)
    if (targetNode) {
      targetId = targetNode[idKey]
    }
  } else {
    // 直接使用ID
    targetId = targetIdOrPredicate
  }
  
  if (targetId === null || targetId === undefined) {
    // 未找到目标节点，返回原树
    return tree
  }
  
  // 添加目标节点ID
  excludeIds.add(targetId)
  
  // 递归收集所有子节点ID
  const collectChildrenIds = (tree, targetId) => {
    tree.forEach(node => {
      const nodeId = node[idKey]
      if (nodeId === targetId) {
        // 找到目标节点，收集其所有子节点
        const children = node[childrenKey]
        if (Array.isArray(children) && children.length > 0) {
          children.forEach(child => {
            const childId = child[idKey]
            excludeIds.add(childId)
            collectChildrenIds([child], childId)
          })
        }
      } else {
        // 继续递归查找
        const children = node[childrenKey]
        if (Array.isArray(children) && children.length > 0) {
          collectChildrenIds(children, targetId)
        }
      }
    })
  }
  
  collectChildrenIds(tree, targetId)
  
  // 使用 filterTree 过滤掉需要排除的节点
  return filterTree(tree, node => !excludeIds.has(node[idKey]), childrenKey)
}
