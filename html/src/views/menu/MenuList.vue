<template>
  <div class="list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.menu') }}</span>
          <div class="header-actions">
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

      <!-- 表格工具栏 -->
      <TableToolbar
        :on-refresh="handleRefresh"
        fullscreen-target=".list-page"
        :show-column-setting-btn="false"
        :visible-columns="visibleColumns"
        :all-columns="allTableColumns"
        :default-visible-columns="defaultVisibleColumns"
        :column-order="columnOrder"
        :fixed-columns="fixedColumns"
        :on-column-setting-confirm="handleColumnSettingConfirm"
      />

      <div class="list-table-scroll">
      <el-table
        ref="tableRef"
        :key="`table-${tableColumns.length}-${JSON.stringify(columnOrder)}`"
        :data="tableData"
        :loading="loading"
        border
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="isExpanded"
        style="width: 100%"
        height="600"
      >
        <el-table-column 
          v-for="column in tableColumns" 
          :key="column.key || column.prop || column.type"
          :type="column.type"
          :prop="column.prop"
          :label="column.label"
          :width="column.width"
          :min-width="column.minWidth"
          :fixed="column.fixed"
        >
          <template v-if="column.slot" #default="{ row }">
            <template v-if="column.key === 'type'">
              <el-tag :type="row.type === 1 ? 'info' : (row.type === 2 ? 'primary' : 'warning')">
                {{ row.type === 1 ? $t('menu.type_directory') : (row.type === 2 ? $t('menu.type_menu') : $t('menu.type_button')) }}
              </el-tag>
            </template>
            <template v-else-if="column.key === 'link_type'">
              <el-tag :type="row.link_type === 1 ? 'primary' : 'success'">
                {{ row.link_type === 1 ? $t('menu_management.link_type_internal') : $t('menu_management.link_type_external') }}
              </el-tag>
            </template>
            <template v-else-if="column.key === 'open_type'">
              <span v-if="row.link_type === 2">
                <el-tag :type="row.open_type === 1 ? 'info' : 'warning'">
                  {{ row.open_type === 1 ? $t('menu_management.open_type_iframe') : $t('menu_management.open_type_new_window') }}
                </el-tag>
              </span>
              <span v-else>-</span>
            </template>
            <template v-else-if="column.key === 'no_cache'">
              <el-tooltip
                v-if="row.no_cache === 1"
                :content="$t('menu_management.no_cache_no')"
                placement="top"
              >
                <el-tag type="warning">
                  {{ $t('common.no') }}
                </el-tag>
              </el-tooltip>
              <el-tag v-else type="success">
                {{ $t('common.yes') }}
              </el-tag>
            </template>
            <template v-else-if="column.key === 'icon'">
              <span v-if="getIconComponent(row.icon)" class="menu-icon-preview">
                <el-icon><component :is="getIconComponent(row.icon)" /></el-icon>
                <span class="menu-icon-name">{{ normalizeIconName(row.icon) }}</span>
              </span>
              <span v-else>-</span>
            </template>
            <template v-else-if="column.key === 'status'">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                {{ row.status === 1 ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
            </template>
            <template v-else-if="column.key === 'is_hidden'">
              <el-tag :type="row.is_hidden === 0 ? 'success' : 'info'">
                {{ row.is_hidden === 0 ? $t('menu_management.is_hidden_show') : $t('menu_management.is_hidden_hide') }}
              </el-tag>
            </template>
            <template v-else-if="column.key === 'operation'">
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
          </template>
        </el-table-column>
      </el-table>
      </div>
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
import TableToolbar from '../../components/TableToolbar.vue'
import { getMenuList, deleteMenu } from '../../api/menu'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'
import { useColumnSetting } from '../../composables/useColumnSetting'
import { useElTableColumns } from '../../composables/useElTableColumns'
import { mapTree } from '../../utils/tree'

const { t } = useI18n()
const { getButtonState } = usePermission()
const tableRef = ref(null)
const loading = ref(false)
const isExpanded = ref(false)

// 使用 CRUD composable
const {
  dialogVisible,
  editId,
  handleAdd,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: deleteMenu
})

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

// 树形菜单选项（保持树形结构，用于 el-tree-select）
const menuOptions = computed(() => {
  return mapTree(tableData.value, menu => ({
    value: menu.id,
    label: menu.name
  }))
})

// 表格列配置
const allTableColumns = computed(() => [
  { type: 'index', width: 60, title: t('table.seq'), key: 'index' },
  { field: 'name', title: t('menu_management.name'), minWidth: 200, key: 'name' },
  { field: 'slug', title: t('menu_management.slug'), minWidth: 150, key: 'slug' },
  { field: 'path', title: t('menu_management.path'), minWidth: 200, key: 'path' },
  { field: 'type', title: t('table.type'), width: 100, slot: 'type', key: 'type' },
  { field: 'link_type', title: t('menu_management.link_type'), width: 120, slot: 'link_type', key: 'link_type' },
  { field: 'open_type', title: t('menu_management.open_type'), width: 140, slot: 'open_type', key: 'open_type' },
  { field: 'no_cache', title: t('menu_management.no_cache'), width: 100, slot: 'no_cache', key: 'no_cache' },
  { field: 'icon', title: t('menu_management.icon'), width: 140, slot: 'icon', key: 'icon' },
  { field: 'sort', title: t('common.sort'), width: 80, key: 'sort' },
  { field: 'status', title: t('table.status'), width: 100, slot: 'status', key: 'status' },
  { field: 'is_hidden', title: t('menu_management.is_hidden'), width: 100, slot: 'is_hidden', key: 'is_hidden' },
  { field: 'created_at', title: t('table.created_at'), width: 180, key: 'created_at' },
  { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', key: 'operation' }
])

// 使用列设置 composable
const {
  tableColumns: tableColumnsConfig,
  visibleColumns,
  allColumns,
  defaultVisibleColumns,
  columnOrder,
  fixedColumns,
  handleColumnSettingConfirm
} = useColumnSetting('menu', allTableColumns)

// 将 VxeTable 格式的列配置转换为 el-table 格式
const tableColumns = useElTableColumns(tableColumnsConfig, visibleColumns, columnOrder, fixedColumns)

const loadData = async () => {
  loading.value = true
  try {
    const res = await getMenuList()
    if (res.data) {
      // 后端已返回前端可直接使用的树形结构，无需转换
      const menus = res.data.menus || res.data.list || []
      tableData.value = menus
    }
  } catch (error) {
    console.error('Load menu list error:', error)
  } finally {
    loading.value = false
  }
}

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleFormSuccess = () => {
  loadData()
}

const handleDelete = (row) => handleDeleteCrud(row, loadData)

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


.menu-icon-preview {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.menu-icon-name {
  font-size: 12px;
  color: var(--text-color-regular);
}
</style>

