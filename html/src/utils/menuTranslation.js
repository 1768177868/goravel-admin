/**
 * 从路径提取可能的 slug
 * 例如：/admins -> admin, /online-users -> online-user, /operation-logs -> operation-log
 */
function extractSlugFromPath(path) {
  if (!path) return null
  
  // 移除查询参数和通配符
  let cleanPath = path.split('?')[0].replace(/\*/g, '').replace(/\/$/, '')
  // 移除路径参数（如 /admins/123 -> /admins）
  cleanPath = cleanPath.replace(/\/\d+(\/|$)/g, '/').replace(/\/$/, '')
  
  // 获取路径的最后一部分
  const parts = cleanPath.split('/').filter(p => p)
  if (parts.length > 0) {
    return parts[parts.length - 1]
  }
  
  return null
}

/**
 * 菜单翻译工具函数
 * 自动处理 slug 的各种格式（连字符、下划线）和变体（带/不带 _management 后缀）
 * 使用 te() 检查键是否存在，避免警告
 */
export function getMenuTranslation(t, te, slug) {
  if (!slug) return null
  
  // 尝试多种 slug 格式
  const slugVariants = [
    slug, // 原始 slug（如 online-user）
    slug.replace(/-/g, '_'), // 连字符转下划线（如 online_user）
    slug.replace(/_/g, '-') // 下划线转连字符（如 online-user）
  ]
  
  // 去重
  const uniqueVariants = [...new Set(slugVariants)]
  
  // 尝试每种格式
  for (const variant of uniqueVariants) {
    // 尝试简短键
    const slugKey = `menu.${variant}`
    // 使用 te() 检查键是否存在，避免警告
    if (typeof te === 'function' && te(slugKey)) {
      return t(slugKey)
    }
    
    // 尝试添加 _management 后缀
    const slugKeyWithSuffix = `menu.${variant}_management`
    if (typeof te === 'function' && te(slugKeyWithSuffix)) {
      return t(slugKeyWithSuffix)
    }
  }
  
  return null
}

/**
 * 获取菜单标题（完整版本，支持从路径提取 slug）
 * 优先使用 slug，如果没有则从路径提取，最后使用原始标题
 */
export function getMenuTitle(t, te, menu) {
  if (!menu || typeof menu !== 'object') {
    return ''
  }
  
  // 优先使用 slug 作为翻译键标识
  const slug = menu.Slug || menu.slug || ''
  if (slug) {
    const translated = getMenuTranslation(t, te, slug)
    if (translated) {
      return translated
    }
  }
  
  // 如果 slug 不存在或翻译失败，尝试从路径提取 slug
  const path = menu.path || menu.Path || ''
  if (path) {
    const extractedSlug = extractSlugFromPath(path)
    if (extractedSlug && extractedSlug !== slug) {
      const translated = getMenuTranslation(t, te, extractedSlug)
      if (translated) {
        return translated
      }
    }
  }
  
  // 最后使用原始标题
  return menu.Title || menu.title || ''
}

