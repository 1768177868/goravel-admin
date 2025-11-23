<template>
  <div class="menu-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu_management.title') }}</span>
          <div class="header-actions">
            <el-button @click="handleToggleExpand">
              <el-icon><component :is="isExpanded ? 'Fold' : 'Expand'" /></el-icon>
              {{ isExpanded ? $t('menu_management.collapse_all') : $t('menu_management.expand_all') }}
            </el-button>
            <el-button type="primary" @click="handleAdd">
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
        <el-table-column prop="path" :label="$t('menu_management.path')" min-width="200" />
        <el-table-column prop="icon" :label="$t('menu_management.icon')" width="100" />
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
            <el-button type="primary" link @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
            <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item :label="$t('menu_management.parent_menu')">
          <el-select v-model="formData.parent_id" :placeholder="$t('form.select_parent') + $t('menu_management.parent_menu')" clearable>
            <el-option :label="$t('menu_management.top_menu')" :value="0" />
            <el-option
              v-for="menu in menuOptions"
              :key="menu.id"
              :label="menu.name"
              :value="menu.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('menu_management.name')" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item :label="$t('menu_management.path')" prop="path">
          <el-input v-model="formData.path" />
        </el-form-item>
        <el-form-item :label="$t('menu_management.icon')">
          <el-input v-model="formData.icon" />
        </el-form-item>
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :label="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('common.sort')">
          <el-input-number v-model="formData.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Fold, Expand, Plus } from '@element-plus/icons-vue'
import { getMenuList, getMenuDetail, createMenu, updateMenu, deleteMenu } from '../../api/menu'

const { t } = useI18n()
const formRef = ref(null)
const tableRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('menu_management.edit_menu') : t('menu_management.add_menu'))
const isExpanded = ref(false)

const tableData = ref([])

const formData = reactive({
  id: null,
  parent_id: 0,
  name: '',
  path: '',
  icon: '',
  status: 1,
  sort: 0
})

const formRules = computed(() => ({
  name: [{ required: true, message: t('menu_management.name_required'), trigger: 'blur' }],
  path: [{ required: true, message: t('menu_management.path_required'), trigger: 'blur' }]
}))

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
    path: menu.Path || menu.path || '',
    icon: menu.Icon || menu.icon || '',
    status: menu.Status !== undefined ? menu.Status : (menu.status !== undefined ? menu.status : 1),
    sort: menu.Sort !== undefined ? menu.Sort : (menu.sort !== undefined ? menu.sort : 0),
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
      // 转换数据格式
      const transformed = menus.map(menu => transformMenuData(menu))
      console.log('Transformed menu data:', transformed)
      tableData.value = transformed
    }
  } catch (error) {
    console.error('Load menu list error:', error)
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  Object.assign(formData, {
    id: null,
    parent_id: 0,
    name: '',
    path: '',
    icon: '',
    status: 1,
    sort: 0
  })
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  try {
    const res = await getMenuDetail(row.id)
    if (res.data && res.data.menu) {
      const menu = res.data.menu
      Object.assign(formData, {
        id: menu.id,
        parent_id: menu.parent_id || 0,
        name: menu.name,
        path: menu.path,
        icon: menu.icon || '',
        status: menu.status,
        sort: menu.sort || 0
      })
      dialogVisible.value = true
    }
  } catch (error) {
    console.error('Load menu detail error:', error)
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const data = { ...formData }
        if (data.parent_id === 0) {
          data.parent_id = null
        }
        if (formData.id) {
          await updateMenu(formData.id, data)
          ElMessage.success(t('menu_management.update_success'))
        } else {
          await createMenu(data)
          ElMessage.success(t('menu_management.create_success'))
        }
        dialogVisible.value = false
        loadData()
      } catch (error) {
        console.error('Submit error:', error)
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
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
</style>

