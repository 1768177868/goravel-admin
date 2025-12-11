<template>
  <div class="menu-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu_management.title') }}</span>
          <div class="header-actions">
            <el-button @click="handleRefresh">
              <el-icon><Refresh /></el-icon>
              {{ $t('tabs.refresh') }}
            </el-button>
            <el-button @click="handleToggleExpand">
              <el-icon><component :is="isExpanded ? 'Fold' : 'Expand'" /></el-icon>
              {{ isExpanded ? $t('menu_management.collapse_all') : $t('menu_management.expand_all') }}
            </el-button>
            <el-button 
              type="primary" 
              :disabled="getButtonState('menu.store').disabled"
              @click="handleAdd"
            >
              <el-icon><Plus /></el-icon>
              {{ $t('menu_management.add_menu') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="isExpanded"
        style="width: 100%"
        height="600"
      >
        <el-table-column type="index" width="60" :label="$t('table.seq')" />
        <el-table-column prop="name" :label="$t('menu_management.name')" min-width="200" />
        <el-table-column prop="slug" :label="$t('menu_management.slug')" min-width="150" />
        <el-table-column prop="path" :label="$t('menu_management.path')" min-width="200" />
        <el-table-column prop="link_type" :label="$t('menu_management.link_type')" width="120">
          <template #default="{ row }">
            <el-tag :type="row.link_type === 1 ? 'primary' : 'success'">
              {{ row.link_type === 1 ? $t('menu_management.link_type_internal') : $t('menu_management.link_type_external') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="open_type" :label="$t('menu_management.open_type')" width="140">
          <template #default="{ row }">
            <span v-if="row.link_type === 2">
              <el-tag :type="row.open_type === 1 ? 'info' : 'warning'">
                {{ row.open_type === 1 ? $t('menu_management.open_type_iframe') : $t('menu_management.open_type_new_window') }}
              </el-tag>
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="icon" :label="$t('menu_management.icon')" width="140">
          <template #default="{ row }">
            <span v-if="getIconComponent(row.icon)" class="menu-icon-preview">
              <el-icon><component :is="getIconComponent(row.icon)" /></el-icon>
              <span class="menu-icon-name">{{ normalizeIconName(row.icon) }}</span>
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" :label="$t('common.sort')" width="80" />
        <el-table-column prop="status" :label="$t('table.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" :label="$t('table.created_at')" width="180" />
        <el-table-column :label="$t('table.operation')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button 
              type="primary" 
              link 
              :disabled="getButtonState('menu.update').disabled"
              @click="handleEdit(row)"
            >
              {{ $t('common.edit') }}
            </el-button>
            <el-button 
              type="danger" 
              link 
              :disabled="getButtonState('menu.destroy').disabled"
              @click="handleDelete(row)"
            >
              {{ $t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <MenuForm
      v-model="dialogVisible"
      :edit-id="editId"
      :menu-options="menuOptions"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Fold, Expand, Plus, Refresh } from '@element-plus/icons-vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import MenuForm from './MenuForm.vue'
import { getMenuList, deleteMenu } from '../../api/menu'
import { usePermission } from '../../composables/usePermission'

const { t } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)
const loading = ref(false)
const dialogVisible = ref(false)
const editId = ref(null)
const isExpanded = ref(false)

const tableData = ref([])

const iconComponents = ElementPlusIconsVue

const normalizeIconName = (iconName) => {
  if (!iconName) {
    return ''
  }
  const trimmed = iconName.trim()
  if (!trimmed) {
    return ''
  }
  if (iconComponents[trimmed]) {
    return trimmed
  }
  const pascalCase = trimmed.charAt(0).toUpperCase() + trimmed.slice(1)
  if (iconComponents[pascalCase]) {
    return pascalCase
  }
  return ''
}

const getIconComponent = (iconName) => {
  const normalized = normalizeIconName(iconName)
  return normalized ? iconComponents[normalized] : null
}

// 扁平化菜单选项（递归处理树形结构）
const menuOptions = computed(() => {
  const flatten = (menus, parentId = 0) => {
    const result = []
    menus.forEach(menu => {
      if (menu.parent_id === parentId) {
        result.push(menu)
        // 递归处理子菜单
        if (menu.children && menu.children.length > 0) {
          const children = flatten(menu.children, menu.id)
          result.push(...children)
        }
      }
    })
    return result
  }
  return flatten(tableData.value)
})

// 转换后端数据格式为前端格式
const transformMenuData = (menu) => {
  // 处理 children，确保递归转换
  const children = menu.Children || menu.children
  let transformedChildren = []
  
  if (children && Array.isArray(children) && children.length > 0) {
    transformedChildren = children.map(child => transformMenuData(child))
  }
  
  const result = {
    id: menu.id,
    parent_id: menu.ParentID || menu.parent_id || 0,
    name: menu.Title || menu.name || '',
    slug: menu.Slug || menu.slug || '',
    path: menu.Path || menu.path || '',
    icon: menu.Icon || menu.icon || '',
    status: menu.Status !== undefined ? menu.Status : (menu.status !== undefined ? menu.status : 1),
    sort: menu.Sort !== undefined ? menu.Sort : (menu.sort !== undefined ? menu.sort : 0),
    link_type: menu.LinkType !== undefined ? menu.LinkType : (menu.link_type !== undefined ? menu.link_type : 1),
    open_type: menu.OpenType !== undefined ? menu.OpenType : (menu.open_type !== undefined ? menu.open_type : 1),
    created_at: menu.created_at || '',
    updated_at: menu.updated_at || ''
  }
  
  // 只有当有子节点时才添加 children 字段
  if (transformedChildren.length > 0) {
    result.children = transformedChildren
  }
  
  return result
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await getMenuList()
    if (res.data) {
      // 后端返回的是 menus 数组，不是 list
      const menus = res.data.menus || res.data.list || []
      const transformed = menus.map(menu => transformMenuData(menu))
      tableData.value = transformed
    }
  } catch (error) {
    console.error('Load menu list error:', error)
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  editId.value = null
  dialogVisible.value = true
}

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleFormSuccess = () => {
  loadData()
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(t('menu_management.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteMenu(row.id)
    ElMessage.success(t('menu_management.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
    }
  }
}

const handleRefresh = () => {
  loadData()
}

const handleToggleExpand = () => {
  isExpanded.value = !isExpanded.value
  
  if (tableRef.value) {
    // Element Plus 的 el-table 使用 toggleRowExpansion 方法
    // 递归处理所有节点
    const toggleNode = (rows) => {
      if (Array.isArray(rows)) {
        rows.forEach(row => {
          // 切换当前节点的展开状态
          tableRef.value.toggleRowExpansion(row, isExpanded.value)
          
          // 如果有子节点，递归处理
          if (row.children && row.children.length > 0) {
            toggleNode(row.children)
          }
        })
      }
    }
    
    // 处理所有顶级节点
    toggleNode(tableData.value)
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.menu-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.menu-icon-preview {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.menu-icon-name {
  font-size: 12px;
  color: #666;
}
</style>

