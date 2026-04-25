<template>
  <el-popover
    v-model:visible="visible"
    placement="bottom-end"
    :width="400"
    trigger="click"
    popper-class="menu-search-popover"
    :teleported="true"
  >
    <template #reference>
      <el-button 
        type="text" 
        class="header-btn topbar-icon-btn" 
        :title="$t('header.menu_search') || '搜索菜单'"
      >
        <el-icon class="header-icon-fixed topbar-icon"><Search /></el-icon>
      </el-button>
    </template>
      
      <div class="menu-search-content">
        <el-input
          v-model="searchKeyword"
          :placeholder="$t('header.menu_search_placeholder') || '搜索菜单...'"
          clearable
          @input="handleSearch"
          @keydown.enter="handleEnter"
          @keydown.down.prevent="handleArrowDown"
          @keydown.up.prevent="handleArrowUp"
          ref="inputRef"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        
        <div class="menu-search-results" v-if="filteredMenus.length > 0">
          <div
            v-for="(menu, index) in filteredMenus"
            :key="menu.id || index"
            class="menu-search-item"
            :class="{ 'is-active': index === selectedIndex }"
            @click="handleMenuClick(menu)"
            @mouseenter="selectedIndex = index"
          >
            <el-icon v-if="getIcon(menu.icon)" class="menu-item-icon">
              <component :is="getIcon(menu.icon)" />
            </el-icon>
            <span class="menu-item-title">{{ getMenuTitle(menu) }}</span>
            <span v-if="menu.path" class="menu-item-path">{{ menu.path }}</span>
          </div>
        </div>
        
        <div v-else-if="searchKeyword" class="menu-search-empty">
          {{ $t('header.menu_search_no_results') || '未找到匹配的菜单' }}
        </div>
        
        <div v-else class="menu-search-empty">
          {{ $t('header.menu_search_hint') || '输入关键词搜索菜单' }}
        </div>
      </div>
    </el-popover>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Search } from '@element-plus/icons-vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { getMenuTitle as getMenuTitleUtil } from '../utils/menuTranslation'

const router = useRouter()
const { t, te } = useI18n()

const props = defineProps({
  menus: {
    type: Array,
    default: () => []
  }
})

const visible = ref(false)
const searchKeyword = ref('')
const selectedIndex = ref(-1)
const inputRef = ref(null)

// 扁平化菜单树，用于搜索
const flattenMenus = (menus, result = []) => {
  menus.forEach(menu => {
    // 获取菜单类型（1:目录 2:菜单 3:按钮）
    const menuType = menu.type !== undefined ? menu.type : (menu.Type !== undefined ? menu.Type : 1)
    
    // 排除目录类型（type === 1）和按钮类型（type === 3）
    // 只添加菜单类型（type === 2）且有路径的菜单项
    if (menuType === 2 && menu.path && menu.path.trim() !== '') {
      result.push(menu)
    }
    
    // 递归处理子菜单
    if (menu.children && menu.children.length > 0) {
      flattenMenus(menu.children, result)
    }
  })
  return result
}

// 所有可搜索的菜单
const allMenus = computed(() => {
  return flattenMenus(props.menus || [])
})

// 过滤后的菜单
const filteredMenus = computed(() => {
  if (!searchKeyword.value || searchKeyword.value.trim() === '') {
    return []
  }
  
  const keyword = searchKeyword.value.toLowerCase().trim()
  return allMenus.value.filter(menu => {
    const title = getMenuTitle(menu).toLowerCase()
    const path = (menu.path || '').toLowerCase()
    const slug = ((menu.slug || menu.Slug) || '').toLowerCase()
    
    return title.includes(keyword) || 
           path.includes(keyword) || 
           slug.includes(keyword)
  }).slice(0, 10) // 最多显示10个结果
})

// 获取菜单标题
const getMenuTitle = (menu) => {
  return getMenuTitleUtil(t, te, menu)
}

// 获取图标
const normalizeIconName = (iconName) => {
  if (!iconName) return ''
  const trimmed = iconName.trim()
  if (!trimmed) return ''
  if (ElementPlusIconsVue[trimmed]) return trimmed
  const pascalCase = trimmed.charAt(0).toUpperCase() + trimmed.slice(1)
  if (ElementPlusIconsVue[pascalCase]) return pascalCase
  return ''
}

const getIcon = (iconName) => {
  const normalized = normalizeIconName(iconName)
  return normalized ? ElementPlusIconsVue[normalized] : null
}

// 处理搜索
const handleSearch = () => {
  selectedIndex.value = -1
}

// 处理菜单点击
const handleMenuClick = (menu) => {
  // 检查菜单类型和路径
  const menuType = menu.type !== undefined ? menu.type : (menu.Type !== undefined ? menu.Type : 1)
  
  // 如果是目录类型（type === 1）或按钮类型（type === 3），不处理
  if (menuType === 1 || menuType === 3) {
    return
  }
  
  // 如果没有路径，不处理
  if (!menu.path || menu.path.trim() === '') {
    return
  }
  
  const linkType = menu.link_type !== undefined ? menu.link_type : (menu.LinkType !== undefined ? menu.LinkType : 1)
  const openType = menu.open_type !== undefined ? menu.open_type : (menu.OpenType !== undefined ? menu.OpenType : 1)
  
  // 外部链接处理
  if (linkType === 2) {
    // iframe 嵌套显示
    if (openType === 1) {
      const title = getMenuTitle(menu)
      const iframePath = `/iframe?url=${encodeURIComponent(menu.path)}&title=${encodeURIComponent(title)}`
      router.push(iframePath)
    } 
    // 新窗口打开
    else if (openType === 2) {
      window.open(menu.path, '_blank')
    }
  } 
  // 内部页面路由
  else {
    router.push(menu.path)
  }
  
  // 关闭搜索框
  visible.value = false
  searchKeyword.value = ''
}

// 处理回车键
const handleEnter = () => {
  if (selectedIndex.value >= 0 && filteredMenus.value[selectedIndex.value]) {
    handleMenuClick(filteredMenus.value[selectedIndex.value])
  } else if (filteredMenus.value.length > 0) {
    handleMenuClick(filteredMenus.value[0])
  }
}

// 处理下箭头
const handleArrowDown = () => {
  if (selectedIndex.value < filteredMenus.value.length - 1) {
    selectedIndex.value++
  } else {
    selectedIndex.value = 0
  }
}

// 处理上箭头
const handleArrowUp = () => {
  if (selectedIndex.value > 0) {
    selectedIndex.value--
  } else {
    selectedIndex.value = filteredMenus.value.length - 1
  }
}

// 监听弹窗打开，自动聚焦输入框
watch(visible, (newVal) => {
  if (newVal) {
    nextTick(() => {
      if (inputRef.value && inputRef.value.focus) {
        inputRef.value.focus()
      }
    })
  } else {
    searchKeyword.value = ''
    selectedIndex.value = -1
  }
})
</script>

<style scoped>
.menu-search-content {
  padding: 8px;
}

.header-icon-fixed {
  display: flex;
  align-items: center;
  justify-content: center;
}

.menu-search-results {
  margin-top: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.menu-search-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  margin-bottom: 4px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
  gap: 8px;
}

.menu-search-item:hover,
.menu-search-item.is-active {
  background-color: #f5f7fa;
}

.menu-item-icon {
  flex-shrink: 0;
  font-size: 16px;
  color: #606266;
}

.menu-item-title {
  flex: 1;
  font-size: 14px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.menu-item-path {
  flex-shrink: 0;
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}

.menu-search-empty {
  margin-top: 12px;
  padding: 20px;
  text-align: center;
  color: #909399;
  font-size: 14px;
}
</style>

<style>
.menu-search-popover {
  padding: 0 !important;
}

.menu-search-popover .el-popover__content {
  padding: 0 !important;
}
</style>
