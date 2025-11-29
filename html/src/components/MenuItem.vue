<template>
  <el-sub-menu v-if="menu.children && menu.children.length > 0" :index="menu.path || `menu-${menu.id}`">
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
  <el-menu-item v-else :index="menu.path" :disabled="menu.status === 0">
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
    
    // 获取菜单标题（使用工具函数，自动从 slug 或路径提取翻译）
    const getMenuTitle = (menu) => {
      return getMenuTitleUtil(t, te, menu)
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
      getMenuTitle
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

