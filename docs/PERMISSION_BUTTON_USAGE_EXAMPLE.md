# 权限按钮显示功能使用示例

## 示例：在权限列表页面中使用

### 修改前（不控制显示）

```vue
<template>
  <template v-else-if="column.slot === 'operation'" #default="{ row }">
    <el-button type="primary" link @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
    <el-button type="danger" link @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
  </template>
</template>
```

### 修改后（根据配置和权限控制显示）

#### 方式一：使用 getButtonState（推荐，一个方法搞定）

```vue
<template>
  <template v-else-if="column.slot === 'operation'" #default="{ row }">
    <el-button 
      v-if="getButtonState('permission.update').show" 
      type="primary" 
      link 
      :disabled="getButtonState('permission.update').disabled"
      @click="handleEdit(row)"
    >
      {{ $t('common.edit') }}
    </el-button>
    <el-button 
      v-if="getButtonState('permission.destroy').show" 
      type="danger" 
      link 
      :disabled="getButtonState('permission.destroy').disabled"
      @click="handleDelete(row)"
    >
      {{ $t('common.delete') }}
    </el-button>
  </template>
</template>

<script setup>
import { usePermission } from '@/composables/usePermission'

const { getButtonState } = usePermission()
</script>
```

#### 方式二：使用两个方法（兼容旧代码）

```vue
<template>
  <template v-else-if="column.slot === 'operation'" #default="{ row }">
    <el-button 
      v-if="shouldShowButton('permission.update')" 
      type="primary" 
      link 
      :disabled="isButtonDisabled('permission.update')"
      @click="handleEdit(row)"
    >
      {{ $t('common.edit') }}
    </el-button>
    <el-button 
      v-if="shouldShowButton('permission.destroy')" 
      type="danger" 
      link 
      :disabled="isButtonDisabled('permission.destroy')"
      @click="handleDelete(row)"
    >
      {{ $t('common.delete') }}
    </el-button>
  </template>
</template>

<script setup>
import { usePermission } from '@/composables/usePermission'

const { shouldShowButton, isButtonDisabled } = usePermission()
</script>
```

## 示例：在管理员列表页面中使用

### 添加按钮

```vue
<template>
  <div class="card-header">
    <span>{{ $t('admin.title') }}</span>
    <el-button 
      v-if="getButtonState('admin.store').show" 
      type="primary" 
      :disabled="getButtonState('admin.store').disabled"
      @click="handleAdd"
    >
      <el-icon><PlusIcon /></el-icon>
      {{ $t('admin.add_admin') }}
    </el-button>
  </div>
</template>

<script setup>
import { usePermission } from '@/composables/usePermission'

const { getButtonState } = usePermission()
</script>
```

### 操作列按钮

```vue
<template>
  <template v-else-if="column.slot === 'operation'" #default="{ row }">
    <el-button 
      v-if="getButtonState('admin.update').show" 
      type="primary" 
      link 
      :disabled="getButtonState('admin.update').disabled"
      @click="handleEdit(row)"
    >
      {{ $t('common.edit') }}
    </el-button>
    <el-button 
      v-if="getButtonState('admin.password').show" 
      type="warning" 
      link 
      :disabled="getButtonState('admin.password').disabled"
      @click="handleResetPassword(row)"
    >
      {{ $t('admin.reset_password') }}
    </el-button>
    <el-button 
      v-if="getButtonState('admin.destroy').show" 
      type="danger" 
      link 
      :disabled="getButtonState('admin.destroy').disabled"
      @click="handleDelete(row)"
    >
      {{ $t('common.delete') }}
    </el-button>
  </template>
</template>

<script setup>
import { usePermission } from '@/composables/usePermission'

const { getButtonState } = usePermission()
</script>
```

## 配置效果对比

### 配置：`ADMIN_SHOW_BUTTONS_WITHOUT_PERMISSION=false`（默认）

**用户有权限时：**
- ✅ 按钮显示且可用

**用户无权限时：**
- ❌ 按钮完全隐藏（不显示）

### 配置：`ADMIN_SHOW_BUTTONS_WITHOUT_PERMISSION=true`

**用户有权限时：**
- ✅ 按钮显示且可用

**用户无权限时：**
- ⚠️ 按钮显示但禁用（灰色，不可点击）

## 完整示例：权限管理页面

```vue
<template>
  <div class="permission-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('permission.title') }}</span>
          <el-button 
            v-if="getButtonState('permission.store').show" 
            type="primary" 
            :disabled="getButtonState('permission.store').disabled"
            @click="handleAdd"
          >
            <el-icon><Plus /></el-icon>
            {{ $t('permission.add_permission') }}
          </el-button>
        </div>
      </template>

      <vxe-table>
        <!-- ... 其他列 ... -->
        <template v-else-if="column.slot === 'operation'" #default="{ row }">
          <el-button 
            v-if="getButtonState('permission.update').show" 
            type="primary" 
            link 
            :disabled="getButtonState('permission.update').disabled"
            @click="handleEdit(row)"
          >
            {{ $t('common.edit') }}
          </el-button>
          <el-button 
            v-if="getButtonState('permission.destroy').show" 
            type="danger" 
            link 
            :disabled="getButtonState('permission.destroy').disabled"
            @click="handleDelete(row)"
          >
            {{ $t('common.delete') }}
          </el-button>
        </template>
      </vxe-table>
    </el-card>
  </div>
</template>

<script setup>
import { usePermission } from '@/composables/usePermission'
import { Plus } from '@element-plus/icons-vue'

const { getButtonState } = usePermission()

// ... 其他代码 ...
</script>
```

## 注意事项

1. **权限标识必须正确**：确保使用的权限标识与后端定义的权限 slug 一致
2. **按钮禁用状态**：使用 `:disabled` 属性来禁用按钮，而不是隐藏
3. **配置优先级**：如果用户有权限，按钮总是显示（不受配置影响）
4. **性能考虑**：`shouldShowButton` 和 `isButtonDisabled` 都是响应式的，可以安全地在模板中使用

