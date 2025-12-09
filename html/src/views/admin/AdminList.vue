<template>
  <div class="admin-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('admin.title') }}</span>
          <el-button 
            type="primary" 
            :disabled="getButtonState('admin.store').disabled"
            @click="handleAdd"
          >
            <el-icon><PlusIcon /></el-icon>
            {{ $t('admin.add_admin') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索表单 -->
      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="{ username: '', status: '', role_id: '', department_id: '', is_2fa_bound: '' }"
        i18n-prefix="admin"
        @search="handleSearch"
        @reset="handleReset"
      >
        <template #extra-buttons>
          <el-button 
            type="success" 
            :disabled="getButtonState('admin.export').disabled"
            @click="handleExport"
          >
            {{ $t('common.export') }}
          </el-button>
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
        :sort-config="{ multiple: false, trigger: 'default' }"
        @sort-change="handleSortChange"
      >
        <template v-for="column in tableColumns" :key="column.field || column.title || column.type">
          <vxe-column
            v-if="column.type === 'checkbox'"
            type="checkbox"
            :width="column.width"
            :fixed="column.fixed"
          />
          <vxe-column
            v-else
            :field="column.field"
            :title="column.title"
            :width="column.width"
            :sortable="column.sortable"
            :fixed="column.fixed"
            :formatter="column.formatter"
            :tree-node="column.treeNode"
          >
            <template v-if="column.slot === 'status'" #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                {{ row.status === 1 ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
            </template>
            <template v-else-if="column.slot === 'department'" #default="{ row }">
              {{ getDepartmentDisplayName(row.Department || row.department) }}
            </template>
            <template v-else-if="column.slot === 'roles'" #default="{ row }">
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
            <template v-else-if="column.slot === 'is_2fa_bound'" #default="{ row }">
              {{ (row.is_2fa_bound || row.Is2FABound) ? $t('admin.google_auth_bound') : $t('admin.google_auth_not_bound') }}
            </template>
            <template v-else-if="column.slot === 'operation'" #default="{ row }">
              <TableActionButtons
                :row="row"
                :primary-actions="getPrimaryActions(row)"
                :more-actions="getMoreActions(row)"
                :get-button-state="getButtonState"
                @action="handleAction"
              />
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
                    <ArrowDownIcon />
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
          <el-select 
            v-model="formData.role_ids" 
            multiple 
            :placeholder="$t('form.select_role')" 
            :disabled="isDefaultAdmin"
            style="width: 100%"
          >
            <el-option
              v-for="role in roles"
              :key="role.id || role.ID"
              :label="role.Name || role.name"
              :value="role.id || role.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('table.status')" prop="status">
          <el-radio-group v-model="formData.status" :disabled="isDefaultAdmin">
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
import { ref, reactive, onMounted, computed, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowDown } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { getStatusOptions } from '../../utils/fieldOptions'
import {
  getAdminList,
  createAdmin,
  updateAdmin,
  deleteAdmin,
  exportAdmin,
  resetPassword,
  kickOutUser,
  unbindAdminGoogleAuth
} from '../../api/admin'
import { getOptions } from '../../api/option'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

// 使用 markRaw 标记图标组件，避免被 Vue 做成响应式对象
const PlusIcon = markRaw(Plus)
const ArrowDownIcon = markRaw(ArrowDown)

// 权限控制
const { getButtonState } = usePermission()

const { t } = useI18n()
const router = useRouter()
const tableRef = ref(null)
const formRef = ref(null)
const submitting = ref(false)
const dialogVisible = ref(false)
const dialogTitle = computed(() => formData.id ? t('admin.edit_admin') : t('admin.add_admin'))

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

// 使用列表页面 composable
const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handlePageChange,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getAdminList,
  initialSearchForm: {
    username: '',
    status: '',
    role_id: '',
    department_id: '',
    is_2fa_bound: ''
  },
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  }
})

// 表格列配置（使用 vxe-table columns）
const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },
  {
    field: 'username',
    title: t('table.username'),
    sortable: true
  },
  {
    field: 'nickname',
    title: t('table.nickname'),
    sortable: true
  },
  {
    field: 'email',
    title: t('table.email'),
    sortable: true
  },
  {
    field: 'phone',
    title: t('table.phone'),
    sortable: true
  },
  {
    field: 'status',
    title: t('table.status'),
    width: 80,
    sortable: true,
    slot: 'status'
  },
  {
    field: 'is_2fa_bound',
    title: t('admin.google_auth_status'),
    width: 120,
    sortable: false,
    slot: 'is_2fa_bound'
  },
  {
    field: 'department',
    title: t('table.department'),
    slot: 'department',
    sortable: false
  },
  {
    field: 'roles',
    title: t('table.roles'),
    slot: 'roles',
    sortable: false
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    sortable: true
  },
  {
    title: t('table.operation'),
    width: 220,
    fixed: 'right',
    slot: 'operation'
  }
])

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
    options: getStatusOptions(t),
    advanced: false
  },
  {
    prop: 'is_2fa_bound',
    label: t('admin.google_auth_status'),
    type: 'select',
    width: '150px',
    options: [
      { label: t('admin.google_auth_bound'), value: '1' },
      { label: t('admin.google_auth_not_bound'), value: '0' }
    ],
    advanced: false
  },
  {
    prop: 'role_id',
    label: t('role.title'),
    type: 'select',
    width: '150px',
    filterable: true,
    apiUrl: '/options?type=role', 
    advanced: false
  },
  {
    prop: 'department_id',
    label: t('table.department'),
    type: 'tree-select',
    width: '200px',
    filterable: true,
    apiUrl: '/options?type=department', 
    treeProps: { label: 'name', children: 'children', value: 'id' },
    advanced: false
  }
])

const departmentTree = ref([])
const roles = ref([])
const departmentSelectVisible = ref(false)
const protectedAdminIds = ref([1, 2])

// 判断是否是超级管理员（通过 is_super_admin 字段判断，不依赖用户名）
const isDefaultAdmin = computed(() => {
  return formData.is_super_admin === true && formData.id !== null
})

const formData = reactive({
  id: null,
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  department_id: null,
  role_ids: [],
  status: 1,
  is_super_admin: false // 是否是超级管理员
})

const formRules = computed(() => ({
  username: [{ required: true, message: t('admin.username_required'), trigger: 'blur' }],
  password: [{ required: true, message: t('admin.password_required'), trigger: 'blur' }]
}))

const loadDepartments = async () => {
  try {
    const res = await getOptions('department')
    if (res.data && res.data.options) {
      departmentTree.value = res.data.options
    }
  } catch (error) {
    logger.error('Load departments error:', error)
    ErrorHandler.handle(error, { silent: true })
  }
}

const loadRoles = async () => {
  try {
    const res = await getOptions('role')
    if (res.data && res.data.options) {
      roles.value = res.data.options.map(option => ({
        id: parseInt(option.value),
        ID: parseInt(option.value),
        name: option.label,
        Name: option.label
      }))
    }
  } catch (error) {
    logger.error('Load roles error:', error)
    ErrorHandler.handle(error, { silent: true })
  }
}

// handleSearch, handleReset, handlePageChange 已由 useListPage 提供

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
    status: 1,
    is_super_admin: false // 重置超级管理员标识
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
    }
  }
  return unique
}

const getDepartmentDisplayName = (department) => {
  if (!department) return '-'
  return department.Name || department.name || '-'
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
const handleDepartmentSelect = (data) => {
  if (data && data.id) {
    formData.department_id = data.id
    departmentSelectVisible.value = false
  }
}

const handleEdit = async (row) => {
  // 处理字段映射，支持 PascalCase 和 snake_case
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
    status: row.Status !== undefined ? row.Status : (row.status !== undefined ? row.status : 1),
    is_super_admin: row.is_super_admin === true || row.IsSuperAdmin === true // 保存超级管理员标识
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
        
        if (data.username) {
          data.username = data.username.trim()
        }
        
        if (formData.id) {
          if (!data.password) {
            delete data.password
          }
          // 如果是默认 admin 用户，不发送 role_ids 字段（即使禁用了也可能包含原值）
          if (isDefaultAdmin.value) {
            delete data.role_ids
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
        logger.error('Submit error:', error)
        // 如果错误已经在响应拦截器中处理过，就不再重复显示
        if (!error.__handled) {
          // 显示更详细的错误信息
          if (error.response && error.response.data && error.response.data.message) {
            ElMessage.error(error.response.data.message)
          } else if (error.message) {
            ElMessage.error(error.message)
          }
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
      logger.error('Delete error:', error)
      // 如果错误已经在响应拦截器中处理过，就不再重复显示
      if (!error.__handled) {
        const errorMessage = error.response?.data?.message || error.message || t('common.operation_failed')
        ElMessage.error(errorMessage)
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
      logger.error('Reset password error:', error)
      ErrorHandler.handle(error, { silent: true })
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
      logger.error('Kick out error:', error)
      ErrorHandler.handle(error, { silent: true })
    }
  }
}

// 获取主要操作按钮配置
const getPrimaryActions = (row) => {
  return [
    {
      key: 'edit',
      label: t('common.edit'),
      type: 'primary',
      permission: 'admin.update',
      handler: handleEdit
    },
    {
      key: 'delete',
      label: t('common.delete'),
      type: 'danger',
      permission: 'admin.destroy',
      show: () => !isProtectedAdmin(row.id),
      handler: handleDelete
    }
  ]
}

// 获取更多操作按钮配置
const getMoreActions = (row) => {
  return [
    {
      key: 'resetPassword',
      command: 'resetPassword',
      label: t('admin.reset_password'),
      permission: 'admin.password',
      handler: handleResetPassword
    },
    {
      key: 'kickOut',
      command: 'kickOut',
      label: t('admin.kick_out'),
      permission: 'admin.kick_out',
      divided: true, // 在踢出和重置密码之间添加分割线
      handler: handleKickOut
    },
    {
      key: 'unbindGoogleAuth',
      command: 'unbindGoogleAuth',
      label: t('admin.unbind_google_auth'),
      permission: 'admin.unbind_google_auth',
      show: () => (row.is_2fa_bound || row.Is2FABound) && !isProtectedAdmin(row.id),
      divided: true, // 在解绑谷歌验证码和踢出之间添加分割线
      handler: handleUnbindGoogleAuth
    }
  ]
}

// 处理操作事件
const handleAction = (command, row) => {
  switch (command) {
    case 'edit':
      handleEdit(row)
      break
    case 'delete':
      handleDelete(row)
      break
    case 'resetPassword':
      handleResetPassword(row)
      break
    case 'kickOut':
      handleKickOut(row)
      break
    case 'unbindGoogleAuth':
      handleUnbindGoogleAuth(row)
      break
  }
}

const handleUnbindGoogleAuth = async (row) => {
  try {
    const { value: code } = await ElMessageBox.prompt(
      t('admin.unbind_google_auth_confirm', { username: row.username || row.Username }),
      t('admin.unbind_google_auth'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        inputPlaceholder: t('profile.enter_6_digit_code'),
        inputType: 'text',
        inputPattern: /^\d{6}$/,
        inputErrorMessage: t('profile.google_code_format')
      }
    )
    
    await unbindAdminGoogleAuth(row.id, { code })
    ElMessage.success(t('admin.unbind_google_auth_success'))
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      logger.error('Unbind google auth error:', error)
      if (!error?.__handled) {
        const errorMessage = error.response?.data?.message || error.translatedMessage || error.message || t('common.operation_failed')
        ElMessage.error(errorMessage)
      }
    }
  }
}

const handleExport = async () => {
  try {
    await exportAdmin(searchForm)
    // 不再直接触发下载，导出记录会写入导出管理列表，由用户在导出管理中查看和下载
    ElMessage.success(t('admin.export_success'))
    // 导出完成后跳转到导出管理列表
    router.push('/exports')
  } catch (error) {
    logger.error('Export error:', error)
    ErrorHandler.handle(error, { silent: true })
  }
}

onMounted(async () => {
  try {
    initDefaultSort()
    await Promise.all([
      loadData(),
      loadDepartments(),
      loadRoles()
    ])
  } catch (error) {
    logger.error('AdminList onMounted error:', error)
    ErrorHandler.handle(error)
  }
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

