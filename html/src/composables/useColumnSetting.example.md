# useColumnSetting 使用示例

这是一个通用的列设置 composable，可以在任何列表页面中使用。

## 基本用法

```vue
<template>
  <div>
    <!-- 在表格头部添加列设置按钮 -->
    <el-button type="info" @click="showColumnSetting = true">
      <el-icon><Setting /></el-icon>
      {{ $t('common.column_setting') }}
    </el-button>

    <!-- 表格 -->
    <vxe-table :columns="tableColumns" :data="tableData">
      <!-- ... -->
    </vxe-table>

    <!-- 列设置对话框 -->
    <el-dialog
      v-model="showColumnSetting"
      :title="$t('common.column_setting')"
      width="500px"
    >
      <el-checkbox-group v-model="visibleColumns">
        <el-checkbox
          v-for="column in allColumns"
          :key="column.key"
          :label="column.key"
          :disabled="column.required"
        >
          {{ column.title }}
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="showColumnSetting = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSaveColumnSetting">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Setting } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { useColumnSetting } from '@/composables/useColumnSetting'

const { t } = useI18n()

// 1. 定义所有可配置的列（用于对话框显示）
const allColumnsConfig = computed(() => [
  { key: 'username', title: t('admin.username'), required: false },
  { key: 'nickname', title: t('admin.nickname'), required: false },
  { key: 'email', title: t('admin.email'), required: false },
  { key: 'phone', title: t('admin.phone'), required: false },
  { key: 'status', title: t('common.status'), required: false },
  { key: 'created_at', title: t('common.created_at'), required: false }
])

// 2. 使用列设置 composable
const {
  showColumnSetting,
  visibleColumns,
  allColumns,
  handleSaveColumnSetting,
  getVisibleColumns
} = useColumnSetting({
  storageKey: 'admin_list_column_setting', // localStorage 键名，每个页面应该不同
  allColumns: allColumnsConfig,
  defaultVisibleColumns: ['username', 'nickname', 'email', 'status'], // 默认显示的列
  alwaysVisibleKeys: ['checkbox', 'operation'] // 始终显示的列
})

// 3. 定义所有列的完整配置（包括始终显示的列）
const allTableColumns = computed(() => [
  { type: 'checkbox', width: 50, fixed: 'left', key: 'checkbox' },
  { field: 'username', title: t('admin.username'), width: 120, sortable: true, key: 'username' },
  { field: 'nickname', title: t('admin.nickname'), width: 120, sortable: true, key: 'nickname' },
  { field: 'email', title: t('admin.email'), width: 180, sortable: true, key: 'email' },
  { field: 'phone', title: t('admin.phone'), width: 150, sortable: true, key: 'phone' },
  { field: 'status', title: t('common.status'), width: 100, sortable: true, key: 'status', slot: 'status' },
  { field: 'created_at', title: t('common.created_at'), width: 180, sortable: true, key: 'created_at' },
  { slot: 'operation', title: t('common.operation'), width: 150, fixed: 'right', key: 'operation' }
])

// 4. 根据 visibleColumns 过滤显示的列
const tableColumns = getVisibleColumns(allTableColumns)
</script>
```

## 参数说明

### useColumnSetting 参数

- `storageKey` (string, 必需): localStorage 存储键名，每个页面应该使用不同的键名
- `allColumns` (ComputedRef<Array>, 必需): 所有可配置列的配置数组
  - 每个列需要包含 `key` (唯一标识) 和 `title` (显示标题)
  - 可选 `required` (boolean): 是否必需显示（禁用复选框）
- `defaultVisibleColumns` (Array<string>, 可选): 默认可见的列 key 数组
- `alwaysVisibleKeys` (Array<string>, 可选): 始终显示的列 key 数组，默认为 `['checkbox', 'operation']`

### 返回值

- `showColumnSetting` (Ref<boolean>): 控制列设置对话框显示
- `visibleColumns` (Ref<Array<string>>): 当前可见的列 key 数组
- `allColumns` (ComputedRef<Array>): 所有可配置的列（已过滤掉始终显示的列）
- `handleSaveColumnSetting` (Function): 保存列设置的方法
- `getVisibleColumns` (Function): 获取过滤后的列的方法，接收 `allTableColumns` 参数

## 注意事项

1. **storageKey 必须唯一**: 每个列表页面应该使用不同的 storageKey，例如：
   - `admin_list_column_setting`
   - `role_list_column_setting`
   - `blacklist_column_setting`

2. **列配置必须包含 key**: 所有列配置（包括始终显示的列）都必须包含 `key` 属性

3. **始终显示的列**: checkbox 和 operation 列默认始终显示，可以通过 `alwaysVisibleKeys` 参数自定义

4. **国际化**: 列的 `title` 应该使用 `t()` 函数进行国际化处理

