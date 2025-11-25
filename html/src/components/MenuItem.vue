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

export default defineComponent({
  name: 'MenuItem',
  props: {
    menu: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const { t } = useI18n()
    
    // 路径到翻译键的映射
    const pathToTranslationKey = {
      '/system': 'menu.system_management',
      '/system/admin': 'menu.admin_management',
      '/system/role': 'menu.role_management',
      '/admins': 'menu.admin_management',
      '/roles': 'menu.role_management',
      '/permissions': 'menu.permission_management',
      '/menus': 'menu.menu_management',
      '/departments': 'menu.department_management',
      '/dictionaries': 'menu.dictionary_management',
      '/logs': 'menu.log_management',
      '/operation-logs': 'menu.operation_log',
      '/login-logs': 'menu.login_log',
      '/system-logs': 'menu.system_log',
      '/monitor': 'menu.service_monitor',
      '/profile': 'menu.profile'
    }
    
    // 标题到翻译键的映射（根据后端返回的中文标题）
    const titleToTranslationKey = {
      '系统管理': 'menu.system_management',
      '管理员管理': 'menu.admin_management',
      '角色管理': 'menu.role_management',
      '权限管理': 'menu.permission_management',
      '菜单管理': 'menu.menu_management',
      '部门管理': 'menu.department_management',
      '字典管理': 'menu.dictionary_management',
      '日志管理': 'menu.log_management',
      '操作日志': 'menu.operation_log',
      '登录日志': 'menu.login_log',
      '系统日志': 'menu.system_log',
      '个人中心': 'menu.profile'
    }

    // 获取菜单标题（优先使用翻译键，如果没有则使用原始标题）
    const getMenuTitle = (menu) => {
      const pathKey = pathToTranslationKey[menu.path]
      if (pathKey) {
        return t(pathKey)
      }
      
      const titleKey = titleToTranslationKey[menu.title]
      if (titleKey) {
        return t(titleKey)
      }
      
      return menu.title
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

