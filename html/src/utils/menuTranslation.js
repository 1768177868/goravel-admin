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

