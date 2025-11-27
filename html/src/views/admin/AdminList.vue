<template>
  <div class="admin-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('admin.title') }}</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            {{ $t('admin.add_admin') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ username: '', status: '' }"
        i18n-prefix="admin"
        @search="handleSearch"
        @reset="handleReset"
      >
        <template #extra-buttons>
          <el-button type="success" @click="handleExport">{{ $t('common.export') }}</el-button>
        </template>
      </SearchForm>

      <!-- vxe-table -->
      <vxe-table
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        border
        :column-config="{ resizable: true }"
        height="600"
        :sort-config="{ multiple: true, trigger: 'default' }"
        @sort-change="handleSortChange"
      >
        <template v-for="column in tableColumns" :key="column.field || column.type">
          <vxe-column
            v-if="column.type !== 'operation'"
            :field="column.field"
            :title="column.title"
            :width="column.width"
            :sortable="column.sortable"
            :fixed="column.fixed"
          >
            <template #default="{ row }">
              <!-- 文本类型 -->
              <template v-if="!column.type || column.type === 'text'">
                {{ getFieldValue(row, column.field, column.formatter) || '-' }}
              </template>
              <!-- 标签类型 -->
              <template v-else-if="column.type === 'tag'">
                <el-tag :type="getTagType(row, column)">
                  {{ getTagText(row, column) }}
                </el-tag>
              </template>
              <!-- 自定义格式化 -->
              <template v-else-if="column.type === 'custom' && column.formatter">
                {{ column.formatter(row) }}
              </template>
              <!-- 自定义插槽 - department -->
              <template v-else-if="column.type === 'custom' && column.slotName === 'department'">
                {{ (row.Department || row.department)?.Name || (row.Department || row.department)?.name || '-' }}
              </template>
              <!-- 自定义插槽 - roles -->
              <template v-else-if="column.type === 'custom' && column.slotName === 'roles'">
                <template v-if="(row.Roles || row.roles) && (row.Roles || row.roles).length > 0">
                  <el-tag 
                    v-for="role in getUniqueRoles(row.Roles || row.roles)" 
                    :key="role.id || role.ID" 
                    style="margin-right: 5px;"
                  >
                    {{ role.Name || role.name }}
                  </el-tag>
                </template>
                <span v-else>-</span>
              </template>
              <!-- 其他自定义插槽 -->
              <template v-else-if="column.type === 'custom' && column.slotName">
                <slot :name="column.slotName" :row="row" :column="column" />
              </template>
            </template>
          </vxe-column>
          <!-- 操作列 -->
          <vxe-column
            v-else
            :title="column.title"
            :width="column.width"
            :fixed="column.fixed"
          >
            <template #default="{ row }">
              <slot name="operation" :row="row">
                <el-button type="primary" link @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
                <el-button type="warning" link @click="handleResetPassword(row)">{{ $t('admin.reset_password') }}</el-button>
                <el-button type="info" link @click="handleKickOut(row)">{{ $t('admin.kick_out') }}</el-button>
                <el-button 
                  v-if="!isProtectedAdmin(row.id)"
                  type="danger" 
                  link 
                  @click="handleDelete(row)"
                >
                  {{ $t('common.delete') }}
                </el-button>
              </slot>
            </template>
          </vxe-column>
        </template>
      </vxe-table>

      <!-- 分页 -->
      <Pagination
        v-model="pagination"
        @page-change="handlePageChange"
      />
    </el-card>

    <!-- 添加/编辑对话框 -->
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
        <el-form-item :label="$t('table.username')" prop="username">
          <el-input v-model="formData.username" :disabled="!!formData.id" />
        </el-form-item>
        <el-form-item :label="$t('common.password')" prop="password" v-if="!formData.id">
          <el-input v-model="formData.password" type="password" />
        </el-form-item>
        <el-form-item :label="$t('table.nickname')" prop="nickname">
          <el-input v-model="formData.nickname" />
        </el-form-item>
        <el-form-item :label="$t('table.email')" prop="email">
          <el-input v-model="formData.email" />
        </el-form-item>
        <el-form-item :label="$t('table.phone')" prop="phone">
          <el-input v-model="formData.phone" />
        </el-form-item>
        <el-form-item :label="$t('table.department')" prop="department_id">
          <el-popover
            placement="bottom-start"
            :width="300"
            trigger="click"
            v-model="departmentSelectVisible"
          >
            <template #reference>
              <el-input
                :model-value="getDepartmentName(formData.department_id)"
                :placeholder="$t('form.select_department')"
                readonly
                @click="departmentSelectVisible = !departmentSelectVisible"
                style="cursor: pointer"
              >
                <template #suffix>
                  <el-icon class="el-input__icon">
                    <ArrowDown />
                  </el-icon>
                </template>
              </el-input>
            </template>
            <el-tree
              :data="departmentTree"
              :props="{ label: 'name', children: 'children' }"
              node-key="id"
              :default-expand-all="false"
              :expand-on-click-node="false"
              :highlight-current="true"
              @node-click="handleDepartmentSelect"
              style="max-height: 300px; overflow-y: auto;"
            >
              <template #default="{ node, data }">
                <span class="custom-tree-node" style="flex: 1; display: flex; align-items: center; justify-content: space-between; font-size: 14px; padding-right: 8px;">
                  <span>{{ node.label }}</span>
                </span>
              </template>
            </el-tree>
          </el-popover>
        </el-form-item>
        <el-form-item :label="$t('table.roles')" prop="role_ids">
          <el-select v-model="formData.role_ids" multiple :placeholder="$t('form.select_role')" style="width: 100%">
            <el-option
              v-for="role in roles"
              :key="role.id || role.ID"
              :label="role.Name || role.name"
              :value="role.id || role.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :label="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
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
import { ArrowDown } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import { useTableSort } from '../../composables/useTableSort'
import {
  getAdminList,
  createAdmin,
  updateAdmin,
  deleteAdmin,
  exportAdmin,
  resetPassword,
  kickOutUser
} from '../../api/admin'
import { getDepartmentList } from '../../api/department'
import { getRoleList } from '../../api/role'

const { t } = useI18n()
const tableRef = ref(null)
const formRef = ref(null)
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('admin.edit_admin') : t('admin.add_admin'))

const searchForm = reactive({
  username: '',
  status: ''
})

// 表格列配置
const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true,
    type: 'text'
  },
  {
    field: 'username',
    title: t('table.username'),
    sortable: true,
    type: 'text'
  },
  {
    field: 'nickname',
    title: t('table.nickname'),
    sortable: true,
    type: 'text'
  },
  {
    field: 'email',
    title: t('table.email'),
    sortable: true,
    type: 'text'
  },
  {
    field: 'phone',
    title: t('table.phone'),
    sortable: true,
    type: 'text'
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 80,
    sortable: true,
    type: 'tag',
    tagConfig: {
      value: (row) => row.status,
      type: (val) => val === 1 ? 'success' : 'danger',
      text: (val) => val === 1 ? t('common.enabled') : t('common.disabled')
    }
  },
  {
    field: 'department',
    title: t('table.department'),
    type: 'custom',
    slotName: 'department'
  },
  {
    field: 'roles',
    title: t('table.roles'),
    type: 'custom',
    slotName: 'roles'
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    sortable: true,
    type: 'text'
  },
  {
    type: 'operation',
    title: t('table.operation'),
    width: 250,
    fixed: 'right'
  }
])

// 获取字段值（支持 PascalCase 和 snake_case，以及格式化函数）
const getFieldValue = (row, field, formatter) => {
  if (formatter && typeof formatter === 'function') {
    return formatter(row)
  }
  if (!field) return ''
  const pascalField = field.charAt(0).toUpperCase() + field.slice(1)
  return row[pascalField] !== undefined ? row[pascalField] : (row[field] !== undefined ? row[field] : '')
}

// 获取标签类型
const getTagType = (row, column) => {
  if (column.tagConfig && column.tagConfig.type) {
    const value = column.tagConfig.value(row)
    return column.tagConfig.type(value)
  }
  return 'info'
}

// 获取标签文本
const getTagText = (row, column) => {
  if (column.tagConfig && column.tagConfig.text) {
    const value = column.tagConfig.value(row)
    return column.tagConfig.text(value)
  }
  return getFieldValue(row, column.field) || '-'
}

// 搜索表单字段配置
const searchFields = computed(() => [
  {
    prop: 'username',
    label: t('table.username'),
    type: 'input',
    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('table.status'),
    type: 'select',
    width: '150px',
    options: [
      { label: t('common.enabled'), value: '1' },
      { label: t('common.disabled'), value: '0' }
    ],
    advanced: false
  }
])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const tableData = ref([])
const departments = ref([])
const departmentTree = ref([])
const roles = ref([])
const departmentSelectVisible = ref(false)
const protectedAdminIds = ref([1, 2])

const formData = reactive({
  id: null,
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  department_id: null,
  role_ids: [],
  status: 1
})

const formRules = computed(() => ({
  username: [{ required: true, message: t('admin.username_required'), trigger: 'blur' }],
  password: [{ required: true, message: t('admin.password_required'), trigger: 'blur' }]
}))

// 字段名映射：前端字段名 -> 数据库字段名
const fieldMapping = {
  'id': 'id',
  'username': 'username',
  'nickname': 'nickname',
  'email': 'email',
  'phone': 'phone',
  'status': 'status',
  'created_at': 'created_at'
}

// 使用排序 composable
const { buildOrderBy, handleSortChange, resetSort, initDefaultSort } = useTableSort({
  tableRef,
  fieldMapping,
  defaultSort: 'id:desc',
  onSortChange: () => {
    pagination.page = 1
    loadData()
  }
})

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      order_by: buildOrderBy()
    }
    // 只添加有值的搜索条件
    if (searchForm.username && searchForm.username.trim()) {
      params.username = searchForm.username.trim()
    }
    if (searchForm.status) {
      params.status = searchForm.status
    }
    
    console.log('Admin search params:', params)
    const res = await getAdminList(params)
    console.log('Admin list response:', res)
    
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('Load admin list error:', error)
  } finally {
    loading.value = false
  }
}

// 转换部门数据为树形结构（支持后端返回的树形结构和扁平结构）
const transformDepartmentToTree = (depts) => {
  if (!depts || depts.length === 0) {
    return []
  }
  
  // 检查是否已经是树形结构（有 Children 字段，即使是空数组也算）
  const firstDept = depts[0]
  const hasChildrenField = firstDept.Children !== undefined || firstDept.children !== undefined
  
  if (hasChildrenField) {
    // 已经是树形结构，只需要转换字段名
    const convertNode = (node) => {
      const children = node.Children || node.children
      const result = {
        id: node.id,
        name: node.Name || node.name || '',
      }
      if (children && Array.isArray(children) && children.length > 0) {
        result.children = children.map(child => convertNode(child))
      }
      return result
    }
    return depts.map(dept => convertNode(dept))
  }
  
  // 扁平结构，需要构建树形结构
  const buildTree = (items, parentId = 0) => {
    const result = []
    items.forEach(item => {
      // 处理 parent_id，支持 null、0 和 undefined
      let itemParentId = 0
      if (item.ParentID !== undefined && item.ParentID !== null) {
        itemParentId = item.ParentID
      } else if (item.parent_id !== undefined && item.parent_id !== null) {
        itemParentId = item.parent_id
      }
      
      // 将 parentId 也转换为数字进行比较
      const compareParentId = parentId === null ? 0 : parentId
      
      if (itemParentId === compareParentId) {
        const node = {
          id: item.id,
          name: item.Name || item.name || '',
          children: buildTree(items, item.id)
        }
        if (node.children.length === 0) {
          delete node.children
        }
        result.push(node)
      }
    })
    return result
  }
  
  return buildTree(depts)
}

const loadDepartments = async () => {
  try {
    const res = await getDepartmentList()
    if (res.data && res.data.list) {
      const depts = res.data.list
      console.log('Loaded departments (raw):', JSON.stringify(depts, null, 2))
      // 保存原始列表（用于扁平化选择）
      departments.value = depts
      // 转换为树形结构
      departmentTree.value = transformDepartmentToTree(depts)
      console.log('Transformed department tree:', JSON.stringify(departmentTree.value, null, 2))
    }
  } catch (error) {
    console.error('Load departments error:', error)
  }
}

const loadRoles = async () => {
  try {
    const res = await getRoleList()
    if (res.data && res.data.list) {
      // 支持 PascalCase 和 snake_case
      roles.value = res.data.list.map(role => ({
        id: role.id || role.ID,
        name: role.Name || role.name || '',
        slug: role.Slug || role.slug || ''
      }))
    }
  } catch (error) {
    console.error('Load roles error:', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.username = ''
  searchForm.status = ''
  resetSort()
  handleSearch()
}

const handlePageChange = ({ currentPage, pageSize }) => {
  pagination.page = currentPage
  pagination.pageSize = pageSize
  loadData()
}

const handleAdd = () => {
  Object.assign(formData, {
    id: null,
    username: '',
    password: '',
    nickname: '',
    email: '',
    phone: '',
    department_id: null,
    role_ids: [],
    status: 1
  })
  dialogVisible.value = true
}

// 获取去重后的角色列表
const getUniqueRoles = (roles) => {
  if (!roles || !Array.isArray(roles)) return []
  const seen = new Set()
  const unique = []
  for (const role of roles) {
    const roleId = role.id || role.ID
    if (roleId && !seen.has(roleId)) {
      seen.add(roleId)
      unique.push(role)
    } else if (roleId) {
      // console.warn('Duplicate role found:', roleId, role)
    }
  }
  if (roles.length !== unique.length) {
    // console.warn(`Roles deduplicated: ${roles.length} -> ${unique.length}`, roles, unique)
  }
  return unique
}

// 获取部门名称
const getDepartmentName = (departmentId) => {
  if (!departmentId) return ''
  const findDept = (depts, id) => {
    for (const dept of depts) {
      if (dept.id === id) {
        return dept.name
      }
      if (dept.children && dept.children.length > 0) {
        const found = findDept(dept.children, id)
        if (found) return found
      }
    }
    return ''
  }
  return findDept(departmentTree.value, departmentId) || ''
}

// 处理部门选择
const handleDepartmentSelect = (data, node) => {
  console.log('Department selected:', data, node)
  if (data && data.id) {
    formData.department_id = data.id
    departmentSelectVisible.value = false
  }
}

const handleEdit = async (row) => {
  // 处理字段映射，支持 PascalCase 和 snake_case
  const adminDepartment = row.Department || row.department
  const adminRoles = row.Roles || row.roles
  
  // 去重角色ID
  const uniqueRoleIds = adminRoles ? [...new Set(adminRoles.map(r => r.id || r.ID).filter(id => id))] : []
  
  Object.assign(formData, {
    id: row.id,
    username: row.Username || row.username || '',
    password: '',
    nickname: row.Nickname || row.nickname || '',
    email: row.Email || row.email || '',
    phone: row.Phone || row.phone || '',
    department_id: row.DepartmentID !== undefined ? row.DepartmentID : (row.department_id !== undefined ? row.department_id : null),
    role_ids: uniqueRoleIds,
    status: row.Status !== undefined ? row.Status : (row.status !== undefined ? row.status : 1)
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const data = { ...formData }
        
        // 确保用户名不为空且去除首尾空格
        if (data.username) {
          data.username = data.username.trim()
        }
        
        console.log('Submit admin data:', data)
        
        if (formData.id) {
          // 编辑时，如果没有修改密码，不传 password
          if (!data.password) {
            delete data.password
          }
          await updateAdmin(formData.id, data)
          ElMessage.success(t('admin.update_success'))
        } else {
          await createAdmin(data)
          ElMessage.success(t('admin.create_success'))
        }
        dialogVisible.value = false
        loadData()
      } catch (error) {
        console.error('Submit error:', error)
        // 显示更详细的错误信息
        if (error.response && error.response.data && error.response.data.message) {
          ElMessage.error(error.response.data.message)
        }
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const isProtectedAdmin = (adminId) => {
  return protectedAdminIds.value.includes(adminId)
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(t('admin.delete_confirm'), t('form.tip'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    await deleteAdmin(row.id)
    ElMessage.success(t('admin.delete_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
      // 显示后端返回的错误信息
      if (error.response && error.response.data && error.response.data.message) {
        const errorMsg = error.response.data.message
        if (errorMsg === 'admin_protected_cannot_delete') {
          ElMessage.error(t('admin.protected_cannot_delete'))
        } else if (errorMsg === 'admin_cannot_delete_self') {
          ElMessage.error(t('admin.cannot_delete_self'))
        } else {
          ElMessage.error(error.response.data.message || t('admin.delete_failed'))
        }
      } else {
        ElMessage.error(t('admin.delete_failed'))
      }
    }
  }
}

const handleResetPassword = async (row) => {
  try {
    const { value: password } = await ElMessageBox.prompt(t('admin.new_password'), t('admin.reset_password'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      inputType: 'password'
    })
    await resetPassword(row.id, { password })
    ElMessage.success(t('admin.reset_password_success'))
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Reset password error:', error)
    }
  }
}

const handleKickOut = async (row) => {
  try {
    await ElMessageBox.confirm(
      t('admin.kick_out_confirm', { username: row.username || row.Username }),
      t('form.tip'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    await kickOutUser(row.id)
    ElMessage.success(t('admin.kick_out_success'))
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Kick out error:', error)
    }
  }
}

const handleExport = async () => {
  try {
    const res = await exportAdmin(searchForm)
    if (res.data && res.data.file_url) {
      window.open(res.data.file_url, '_blank')
      ElMessage.success(t('admin.export_success'))
    }
  } catch (error) {
    console.error('Export error:', error)
  }
}

onMounted(() => {
  initDefaultSort()
  loadData()
  loadDepartments()
  loadRoles()
})
</script>

<style scoped>
.admin-list {
  background: white;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

</style>

