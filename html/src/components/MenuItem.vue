<template>
  <el-sub-menu v-if="menu.children && menu.children.length > 0" :index="getMenuIndex(menu)">
    <template #title>
      <el-icon v-if="getIcon(menu.icon)" class="menu-icon">
        <component :is="getIcon(menu.icon)" />
      </el-icon>
      <el-tooltip
        :content="getMenuTitle(menu)"
        placement="right"
        effect="dark"
        :show-after="300"
      >
        <span class="menu-title">{{ getMenuTitle(menu) }}</span>
      </el-tooltip>
    </template>
    <MenuItem
      v-for="child in menu.children"
      :key="child.id"
      :menu="child"
    />
  </el-sub-menu>
  <el-menu-item 
    v-else 
    :index="getMenuIndex(menu)" 
    :disabled="menu.status === 0"
    @click="handleMenuClick(menu, $event)"
  >
    <el-icon v-if="getIcon(menu.icon)" class="menu-icon">
      <component :is="getIcon(menu.icon)" />
    </el-icon>
    <template #title>
      <el-tooltip
        :content="getMenuTitle(menu)"
        placement="right"
        effect="dark"
        :show-after="300"
      >
        <span class="menu-title">{{ getMenuTitle(menu) }}</span>
      </el-tooltip>
    </template>
  </el-menu-item>
</template>

<script>
import { defineComponent } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useTabsStore } from '../store/tabs'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { getMenuTitle as getMenuTitleUtil } from '../utils/menuTranslation'

export default defineComponent({
  name: 'MenuItem',
  props: {
    menu: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const { t, te } = useI18n()
    const router = useRouter()
    const tabsStore = useTabsStore()
    
    // 获取菜单标题（使用工具函数，自动从 slug 或路径提取翻译）
    const getMenuTitle = (menu) => {
      return getMenuTitleUtil(t, te, menu)
    }
    
    // 获取菜单项的 index（用于 el-menu-item）
    // 对于外部链接，使用唯一标识符而不是 URL，避免路由导航问题
    const getMenuIndex = (menu) => {
      const linkType = menu.link_type !== undefined ? menu.link_type : 1
      // 外部链接使用唯一标识符
      if (linkType === 2) {
        return `external-${menu.id || menu.path}`
      }
      // 内部页面使用路径
      return menu.path || `menu-${menu.id}`
    }
    
    // 处理菜单点击
    const handleMenuClick = (menu, event) => {
      if (!menu.path) {
        return
      }

      // 安全地阻止默认行为（如果事件对象存在）
      if (event && typeof event.preventDefault === 'function') {
        event.preventDefault()
      }
      if (event && typeof event.stopPropagation === 'function') {
        event.stopPropagation()
      }

      const linkType = menu.link_type !== undefined ? menu.link_type : 1
      const openType = menu.open_type !== undefined ? menu.open_type : 1

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
    }
    
    const normalizeIconName = (iconName) => {
      if (!iconName) {
        return ''
      }
      const trimmed = iconName.trim()
      if (!trimmed) {
        return ''
      }
      if (ElementPlusIconsVue[trimmed]) {
        return trimmed
      }
      const pascalCase = trimmed.charAt(0).toUpperCase() + trimmed.slice(1)
      if (ElementPlusIconsVue[pascalCase]) {
        return pascalCase
      }
      return ''
    }

    const getIcon = (iconName) => {
      const normalized = normalizeIconName(iconName)
      return normalized ? ElementPlusIconsVue[normalized] : null
    }

    return {
      getIcon,
      getMenuTitle,
      getMenuIndex,
      handleMenuClick
    }
  }
})
</script>

<style scoped>
.menu-title {
  display: inline-block;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  max-width: 100%;
}

.menu-icon {
  flex-shrink: 0;
  margin-right: 8px;
}
</style>

