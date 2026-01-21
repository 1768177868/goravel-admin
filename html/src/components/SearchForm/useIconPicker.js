import { ref, computed } from 'vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

export function useIconPicker() {
  // 图标组件映射
  const iconComponents = ElementPlusIconsVue
  // 图标选择器显隐
  const iconPickerVisible = ref(false)
  // 图标搜索关键词
  const iconSearch = ref('')
  // 过滤后的图标列表（首字母大写的Element Plus图标）
  const iconList = Object.keys(ElementPlusIconsVue)
    .filter(name => /^[A-Z]/.test(name))
    .sort()

  // 标准化图标名称（兼容大小写、空值）
  const normalizeIconName = (iconName) => {
    if (!iconName) return ''
    const trimmed = iconName.trim()
    if (!trimmed) return ''
    // 直接匹配
    if (iconComponents[trimmed]) return trimmed
    // 首字母大写匹配
    const pascalCase = trimmed.charAt(0).toUpperCase() + trimmed.slice(1)
    if (iconComponents[pascalCase]) return pascalCase
    return ''
  }

  // 获取图标组件（标准化后）
  const getIconComponent = (iconName) => {
    const normalized = normalizeIconName(iconName)
    return normalized ? iconComponents[normalized] : null
  }

  // 过滤图标列表（根据搜索关键词）
  const filteredIcons = computed(() => {
    const keyword = iconSearch.value.trim().toLowerCase()
    if (!keyword) return iconList
    return iconList.filter(name => name.toLowerCase().includes(keyword))
  })

  // 选择图标
  const selectIcon = (icon, model, prop) => {
    model[prop] = icon
    iconPickerVisible.value = false
  }

  // 清空图标
  const clearIcon = (model, prop) => {
    model[prop] = ''
  }

  return {
    iconComponents,
    iconPickerVisible,
    iconSearch,
    iconList,
    normalizeIconName,
    getIconComponent,
    filteredIcons,
    selectIcon,
    clearIcon
  }
}