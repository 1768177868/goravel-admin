# 前端开发指南

本文档提供 Goravel Admin 前端项目的开发指南。

## 目录

- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [开发环境](#开发环境)
- [核心概念](#核心概念)
- [组件开发](#组件开发)
- [Composables 使用](#composables-使用)
- [状态管理](#状态管理)
- [国际化](#国际化)
- [测试](#测试)

---

## 技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.4+ | 渐进式 JavaScript 框架 |
| Element Plus | 2.4+ | Vue 3 UI 组件库 |
| VXE-Table | 4.7+ | 高性能表格组件 |
| Pinia | 2.1+ | Vue 状态管理 |
| Vue Router | 4.2+ | 路由管理 |
| Axios | 1.6+ | HTTP 客户端 |
| vue-i18n | 9.8+ | 国际化 |
| TypeScript | 5.3+ | 类型支持（可选） |
| Vitest | 1.6+ | 单元测试 |

---

## 项目结构

```
html/
├── public/                  # 静态资源
├── src/
│   ├── api/                # API 请求模块
│   │   ├── admin.js        # 管理员 API
│   │   ├── role.js         # 角色 API
│   │   ├── menu.js         # 菜单 API
│   │   └── ...
│   ├── components/         # 通用组件
│   │   ├── SearchForm.vue  # 搜索表单
│   │   ├── Pagination.vue  # 分页组件
│   │   ├── ErrorBoundary.vue
│   │   └── ColumnSettingDialog.vue
│   ├── composables/        # 可复用逻辑
│   │   ├── useCrud.ts      # CRUD 操作 (TypeScript)
│   │   ├── useDebounce.ts  # 防抖 (TypeScript)
│   │   ├── useTableSort.ts # 表格排序 (TypeScript)
│   │   ├── usePermission.ts# 权限检查 (TypeScript)
│   │   ├── useListPage.js  # 列表页面
│   │   ├── useColumnSetting.js
│   │   └── index.ts        # 导出入口
│   ├── i18n/               # 国际化
│   │   ├── index.js        # 配置
│   │   └── locales/
│   │       ├── zh-CN.json  # 中文
│   │       └── en-US.json  # 英文
│   ├── layouts/            # 布局组件
│   │   └── MainLayout.vue
│   ├── router/             # 路由配置
│   │   └── index.js
│   ├── store/              # 状态管理
│   │   ├── user.js         # 用户状态
│   │   ├── app.js          # 应用状态
│   │   └── tabs.js         # 标签页状态
│   ├── types/              # TypeScript 类型
│   │   ├── index.d.ts      # 实体类型
│   │   └── composables.d.ts
│   ├── utils/              # 工具函数
│   │   ├── request.js      # Axios 封装
│   │   ├── storage.js      # 存储工具
│   │   ├── validation.js   # 验证器
│   │   └── logger.js       # 日志工具
│   ├── views/              # 页面组件
│   │   ├── admin/          # 管理员模块
│   │   ├── role/           # 角色模块
│   │   ├── menu/           # 菜单模块
│   │   └── ...
│   ├── App.vue             # 根组件
│   ├── main.js             # 入口文件
│   └── style.css           # 全局样式
├── package.json
├── vite.config.js          # Vite 配置
├── vitest.config.js        # 测试配置
└── tsconfig.json           # TypeScript 配置
```

---

## 开发环境

### 安装依赖

```bash
npm install
```

### 开发服务器

```bash
npm run dev
```

### 生产构建

```bash
npm run build
```

### 类型检查

```bash
npm run type-check
```

### 运行测试

```bash
# 交互式模式
npm test

# 单次运行
npm run test:run

# 覆盖率报告
npm run test:coverage
```

### 环境配置

创建 `.env` 文件：

```env
VITE_API_BASE_URL=http://localhost:3000
VITE_API_PREFIX=/api/admin
```

---

## 核心概念

### 1. 列表页面模式

所有列表页面遵循统一模式：

```vue
<template>
  <el-card>
    <!-- 搜索表单 -->
    <SearchForm :fields="searchFields" v-model="searchForm" @search="handleSearch" />
    
    <!-- 工具栏 -->
    <div class="table-toolbar">
      <el-button @click="handleAdd">添加</el-button>
    </div>
    
    <!-- 表格 -->
    <vxe-table :data="tableData" :loading="loading">
      <vxe-column field="id" title="ID" />
      <!-- ... -->
    </vxe-table>
    
    <!-- 分页 -->
    <Pagination :model-value="pagination" @page-change="handlePageChange" />
    
    <!-- 表单弹窗 -->
    <XxxForm v-model="dialogVisible" :edit-id="editId" @success="handleFormSuccess" />
  </el-card>
</template>

<script setup>
import { useListPage } from '@/composables/useListPage'
import { useCrud } from '@/composables/useCrud'

// 列表逻辑
const { tableData, loading, pagination, searchForm, loadData, handleSearch, handlePageChange } = useListPage({
  fetchApi: getList,
  // ...
})

// CRUD 逻辑
const { dialogVisible, editId, handleAdd, handleEdit, handleDelete, handleFormSuccess } = useCrud({
  deleteApi: deleteItem,
})
</script>
```

### 2. 表单弹窗模式

```vue
<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑' : '添加'">
    <el-form ref="formRef" :model="form" :rules="rules">
      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" />
      </el-form-item>
    </el-form>
    
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
const props = defineProps({
  editId: [Number, String]
})

const emit = defineEmits(['update:modelValue', 'success'])

const isEdit = computed(() => !!props.editId)

const handleSubmit = async () => {
  await formRef.value.validate()
  
  if (isEdit.value) {
    await updateApi(props.editId, form.value)
  } else {
    await createApi(form.value)
  }
  
  emit('success')
}
</script>
```

---

## Composables 使用

### useCrud (TypeScript)

CRUD 操作的统一封装：

```typescript
import { useCrud } from '@/composables/useCrud'

// 完整用法
const {
  dialogVisible,      // 对话框显示状态
  editId,             // 编辑ID
  handleAdd,          // 打开添加弹窗
  handleEdit,         // 打开编辑弹窗
  handleClose,        // 关闭弹窗
  handleFormSuccess,  // 表单提交成功
  handleDelete,       // 删除单条
  handleBatchDelete   // 批量删除
} = useCrud({
  deleteApi: deleteItem,
  batchDeleteApi: batchDeleteItems,
  deleteConfirmKey: 'common.delete_confirm',
  deleteSuccessKey: 'common.delete_success',
  onDeleteSuccess: () => loadData()
})

// 按需使用
// 只需要添加/编辑
const { dialogVisible, editId, handleAdd, handleEdit } = useCrud()

// 只需要删除
const { handleDelete } = useCrud({ deleteApi })
```

### useListPage

列表页面逻辑封装：

```javascript
import { useListPage } from '@/composables/useListPage'

const {
  tableData,      // 表格数据
  loading,        // 加载状态
  pagination,     // 分页信息
  searchForm,     // 搜索表单
  loadData,       // 加载数据
  handleSearch,   // 搜索
  handleReset,    // 重置
  handlePageChange // 分页变化
} = useListPage({
  fetchApi: getList,
  searchFields: ['name', 'status'],
  defaultSearchValues: { status: 1 },
  immediate: true
})
```

### useTableSort (TypeScript)

表格排序逻辑：

```typescript
import { useTableSort } from '@/composables/useTableSort'

const {
  sortConfig,       // 排序配置
  buildOrderBy,     // 构建排序参数
  handleSortChange, // 处理排序变化
  initDefaultSort   // 初始化默认排序
} = useTableSort({
  tableRef,
  defaultSort: 'id:desc',
  fieldMapping: { created_at: 'created_at' },
  onSortChange: () => loadData()
})
```

### usePermission (TypeScript)

权限检查：

```typescript
import { usePermission } from '@/composables/usePermission'

const { hasPermission, shouldShowButton, getButtonState } = usePermission()

// 检查权限
if (hasPermission('admin.store')) {
  // 有权限
}

// 按钮状态
const { show, disabled } = getButtonState('admin.update')
```

---

## 状态管理

### userStore

```javascript
import { useUserStore } from '@/store/user'

const userStore = useUserStore()

// 用户信息
userStore.userInfo

// 权限列表
userStore.permissions

// 菜单列表
userStore.menus

// 检查权限
userStore.hasPermission('admin.store')

// 登录
await userStore.login(username, password)

// 登出
await userStore.logout()
```

### appStore

```javascript
import { useAppStore } from '@/store/app'

const appStore = useAppStore()

// 侧边栏状态
appStore.sidebarCollapsed
appStore.toggleSidebar()

// 全屏
appStore.isFullscreen
appStore.toggleFullscreen()

// 语言
appStore.language
appStore.setLanguage('en-US')
```

---

## 国际化

### 添加翻译

```json
// i18n/locales/zh-CN.json
{
  "module": {
    "title": "模块标题",
    "add": "添加",
    "edit": "编辑"
  }
}
```

### 使用翻译

```vue
<template>
  <!-- 模板中使用 -->
  <span>{{ $t('module.title') }}</span>
  
  <!-- 带参数 -->
  <span>{{ $t('common.total', { count: 10 }) }}</span>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

// 脚本中使用
const title = t('module.title')
</script>
```

---

## 测试

### 目录结构

```
src/
├── composables/
│   ├── useCrud.ts
│   └── __tests__/
│       └── useCrud.test.js
├── utils/
│   ├── validation.js
│   └── __tests__/
│       └── validation.test.js
```

### 编写测试

```javascript
// src/utils/__tests__/validation.test.js
import { describe, it, expect } from 'vitest'
import { validators } from '../validation'

describe('validators', () => {
  describe('required', () => {
    it('应该拒绝空字符串', () => {
      expect(validators.required('')).not.toBe(true)
    })

    it('应该接受有效字符串', () => {
      expect(validators.required('hello')).toBe(true)
    })
  })
})
```

### Mock 示例

```javascript
import { vi } from 'vitest'

// Mock API
vi.mock('@/api/admin', () => ({
  getAdminList: vi.fn().mockResolvedValue({ data: { list: [] } })
}))

// Mock Element Plus
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn().mockResolvedValue(true)
  }
}))
```

---

## 最佳实践

### 1. 组件命名

```
- PascalCase: UserList.vue, UserForm.vue
- 目录按模块组织: views/user/, views/role/
```

### 2. 类型安全

```typescript
// 使用 TypeScript composables
import { useCrud } from '@/composables/useCrud'

// 或使用 JSDoc 类型
/** @type {import('@/types').Admin} */
const admin = {}
```

### 3. 性能优化

```vue
<script setup>
// 使用 shallowRef 减少响应式开销
import { shallowRef } from 'vue'
const tableData = shallowRef([])

// 使用 v-memo 缓存列表项
<template v-for="item in list" :key="item.id" v-memo="[item.id, item.status]">
```

### 4. 错误处理

```javascript
// 使用 ErrorBoundary 组件
<ErrorBoundary>
  <YourComponent />
</ErrorBoundary>

// API 错误统一处理（已在 request.js 中配置）
```

---

## 常见问题

### Q: 如何添加新页面？

1. 在 `views/` 下创建页面组件
2. 在 `api/` 下添加 API 请求
3. 在 `router/index.js` 添加路由
4. 在后台菜单管理中添加菜单

### Q: 如何添加新的 Composable？

1. 在 `composables/` 下创建文件
2. 使用 TypeScript 编写（推荐）
3. 在 `composables/index.ts` 中导出
4. 添加对应的测试文件

### Q: 如何自定义主题？

修改 `src/style.css` 中的 CSS 变量：

```css
:root {
  --el-color-primary: #409eff;
  --el-color-success: #67c23a;
}
```

---

## 参考资源

- [Vue 3 文档](https://vuejs.org)
- [Element Plus 文档](https://element-plus.org)
- [VXE-Table 文档](https://vxetable.cn)
- [Pinia 文档](https://pinia.vuejs.org)
- [vue-i18n 文档](https://vue-i18n.intlify.dev)
- [Vitest 文档](https://vitest.dev)

